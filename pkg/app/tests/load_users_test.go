package engine_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/aserto-dev/go-eds/pkg/pb"
	api "github.com/aserto-dev/go-grpc/aserto/api/v1"
	dir "github.com/aserto-dev/go-grpc/aserto/authorizer/directory/v1"
	"github.com/aserto-dev/go-lib/ids"
	"github.com/aserto-dev/topaz/pkg/cc/config"
	atesting "github.com/aserto-dev/topaz/pkg/testing"
	"google.golang.org/protobuf/proto"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

func TestTenant(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "tests")
	assert := require.New(t)
	assert.NoErrorf(err, "TempDir")

	harness := atesting.SetupOffline(t, func(cfg *config.Config) {
		cfg.Directory.Path = filepath.Join(tmpDir, "test.db")
		t.Logf("dbpath %s", cfg.Directory.Path)
	})
	defer harness.Cleanup()

	client := harness.CreateGRPCDirectoryClient().Directory

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tests := []struct {
		name string
		test func(*testing.T)
	}{
		{"TestCreateDeleteTenantWithID", CreateDeleteTenantWithID(ctx, client)},
		{"TestCreateDeleteTenantNoID", CreateDeleteTenantNoID(ctx, client)},
		{"TestListTenants", ListTenants(ctx, client)},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, testCase.test)
	}
}

func CreateDeleteTenantWithID(ctx context.Context, client dir.DirectoryClient) func(t *testing.T) {
	return func(t *testing.T) {
		tid, err := ids.NewTenantID()
		assert := require.New(t)
		assert.NoErrorf(err, "NewTenantID")

		resp, err := client.CreateTenant(ctx, &dir.CreateTenantRequest{Id: tid})
		assert.NoErrorf(err, "CreateTenant")
		assert.NotNil(resp)
		assert.NotNil(resp.Id)
		assert.NotEmpty(resp.Id)
		assert.Equal(tid, resp.Id)

		resp2, err := client.DeleteTenant(ctx, &dir.DeleteTenantRequest{Id: tid})

		assert.NoErrorf(err, "DeleteTenant")
		assert.NotNil(resp2)
	}
}

func CreateDeleteTenantNoID(ctx context.Context, client dir.DirectoryClient) func(t *testing.T) {
	return func(t *testing.T) {
		resp, err := client.CreateTenant(ctx, &dir.CreateTenantRequest{})
		assert := require.New(t)
		assert.NoErrorf(err, "CreateTenant")
		assert.NotNil(resp)
		assert.NotNil(resp.Id)
		assert.NotEmpty(resp.Id)

		resp2, err := client.DeleteTenant(ctx, &dir.DeleteTenantRequest{
			Id: resp.Id,
		})

		assert.NoErrorf(err, "DeleteTenant")
		assert.NotNil(resp2)
	}
}

func ListTenants(ctx context.Context, client dir.DirectoryClient) func(t *testing.T) {
	return func(t *testing.T) {
		resp, err := client.ListTenants(ctx, &dir.ListTenantsRequest{})
		assert := require.New(t)
		assert.NoErrorf(err, "ListTenants")
		assert.NotNil(resp)
		assert.NotNil(resp.Results)
		assert.NotEmpty(resp.Results)

		for _, tenantID := range resp.Results {
			err := ids.CheckTenantID(tenantID)
			assert.NoErrorf(err, "parse tenant id")
		}
	}
}

func TestLoadUsers(t *testing.T) {
	errc := make(chan error, 1)

	go func() {
		t.Logf("errc spew subscribed")
		for e := range errc {
			t.Logf("errc %s", e.Error())
			t.Fail()
		}
		t.Logf("errc spew closed")
	}()

	tmpDir, err := os.MkdirTemp("", "tests")
	if err != nil {
		errc <- errors.Wrapf(err, "tempdir")
		return
	}

	harness := atesting.SetupOffline(t, func(cfg *config.Config) {
		cfg.Directory.Path = filepath.Join(tmpDir, "test.db")
		t.Logf("dbpath %s", cfg.Directory.Path)
	})
	defer harness.Cleanup()

	defer func() {
		t.Logf("cleanup db")
		dbpath := harness.Engine.Configuration.Directory.Path
		t.Logf("delete scratch db [%s]", dbpath)
		os.Remove(dbpath)
	}()

	// produce stream of users
	// subscriber sends users to GRP dir.LoadUsers API
	// which upserts each user instance
	{
		s := make(chan *api.User, 1)
		done := make(chan bool, 1)

		go subscriber(harness, s, done, errc)

		producer(s, errc)
		close(s)

		<-done
	}

	close(errc)
}

// producer create a stream of api.User instances (272 instances)
func producer(s chan<- *api.User, errc chan<- error) {

	r, err := os.Open(filepath.Join(atesting.AssetsDir(), "users.json"))
	if err != nil {
		errc <- errors.Wrapf(err, "open %s", filepath.Join(atesting.AssetsDir(), "users.json"))
	}
	defer r.Close()

	dec := json.NewDecoder(r)

	if _, err = dec.Token(); err != nil {
		errc <- errors.Wrapf(err, "token open")
	}

	for dec.More() {
		u := api.User{}

		if err = pb.UnmarshalNext(dec, &u); err != nil {
			errc <- errors.Wrapf(err, "unmarshal next")
		}

		s <- &u
	}

	if _, err = dec.Token(); err != nil {
		errc <- errors.Wrapf(err, "token close")
	}
}

func subscriber(harness *atesting.EngineHarness, s <-chan *api.User, done chan<- bool, errc chan<- error) {
	client := harness.CreateGRPCDirectoryClient().Directory

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	stream, err := client.LoadUsers(ctx)
	if err != nil {
		errc <- errors.Wrapf(err, "client.LoadUsers")
	}

	sendCount := int32(0)
	errCount := int32(0)

	for user := range s {

		req := &dir.LoadUsersRequest{
			Data: &dir.LoadUsersRequest_User{
				User: user,
			},
		}

		if err = stream.Send(req); err != nil {
			errc <- errors.Wrapf(err, "send %s", user.Id)
			errCount++
		}
		sendCount++
	}

	res, err := stream.CloseAndRecv()
	if err != nil {
		errc <- errors.Wrapf(err, "stream.CloseAndRecv()")
	}

	if res != nil && res.Received != sendCount {
		errc <- fmt.Errorf("send != received %d - %d", sendCount, res.Received)
	}

	done <- true
}

// const acmecorpJSON string = "https://storage.googleapis.com/aserto-cli/assets/peoplefinder.json"
const acmecorpJSON string = "https://storage.googleapis.com/aserto-cli/assets/acmecorp-dex-users.json"

func downloadAcmeCorpJSON() (io.Reader, error) {
	buf := new(bytes.Buffer)
	resp, err := http.Get(acmecorpJSON)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	_, err = io.Copy(buf, resp.Body)
	return buf, err
}

func TestLoadUsersWithExt(t *testing.T) {
	assert := require.New(t)
	r, err := downloadAcmeCorpJSON()
	if err != nil {
		assert.FailNow("failed to download acmecorm json", err)
	}

	tmpDir, err := os.MkdirTemp("", "tests")
	if err != nil {
		assert.FailNow("failed to create temp dir", err)
	}

	harness := atesting.SetupOffline(t, func(cfg *config.Config) {
		cfg.Directory.Path = filepath.Join(tmpDir, "test.db")
		t.Logf("dbpath %s", cfg.Directory.Path)
	})
	defer harness.Cleanup()

	defer func() {
		t.Logf("cleanup db")
		dbpath := harness.Engine.Configuration.Directory.Path
		t.Logf("delete scratch db [%s]", dbpath)
		os.Remove(dbpath)
	}()

	client := harness.CreateGRPCDirectoryClient().Directory

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	stream, err := client.LoadUsers(ctx)
	if err != nil {
		assert.FailNow("failed to create load user stream client", err)
	}

	dec := json.NewDecoder(r)

	_, err = dec.Token()
	if err != nil {
		assert.FailNow("failed to get token", err)
	}

	fnUser := func(stream dir.Directory_LoadUsersClient, user *api.User) error {
		req := &dir.LoadUsersRequest{
			Data: &dir.LoadUsersRequest_User{
				User: user,
			},
		}
		return stream.Send(req)
	}

	fnUserExt := func(stream dir.Directory_LoadUsersClient, userExt *api.UserExt) error {
		req := &dir.LoadUsersRequest{
			Data: &dir.LoadUsersRequest_UserExt{
				UserExt: userExt,
			},
		}
		return stream.Send(req)
	}

	for dec.More() {
		user := api.User{}

		if err := pb.UnmarshalNext(dec, &user); err != nil {
			assert.FailNow("failed to unmarshal user", err)
		}

		clonedAttributes := proto.Clone(user.Attributes)
		user.Attributes = &api.AttrSet{}

		clonedApplications := make(map[string]*api.AttrSet)
		for k, v := range user.Applications {
			cloneAttrSet := proto.Clone(v)
			clonedApplications[k] = cloneAttrSet.(*api.AttrSet)
		}
		user.Applications = make(map[string]*api.AttrSet)

		var pid string

		for key, value := range user.Identities {
			if value.Kind == api.IdentityKind_IDENTITY_KIND_PID {
				pid = key
				break
			}
		}
		assert.NotEmpty(pid)

		userExt := api.UserExt{
			Id:           pid,
			Attributes:   clonedAttributes.(*api.AttrSet),
			Applications: clonedApplications,
		}

		if err := fnUser(stream, &user); err != nil {
			if err == io.EOF {
				break
			}
			assert.FailNow("failed to get user", err)
		}

		if err := fnUserExt(stream, &userExt); err != nil {
			if err == io.EOF {
				break
			}
			assert.FailNow("failed to get user extensions", err)
		}

	}
	resp, err := client.ListTenants(ctx, &dir.ListTenantsRequest{})

	assert.NoErrorf(err, "ListTenants")
	fmt.Println(resp.Results)

	res, err := stream.CloseAndRecv()
	if err != nil {
		assert.FailNow("failed to receive users", err)
	}
	assert.NoError(err)

	_, err = dec.Token()
	if err != nil {
		assert.FailNow("failed get token", err)
	}

	assert.Equal(int32(554), res.Received)
	assert.Equal(int32(277), res.Created)
	assert.Equal(int32(277), res.Updated)
	assert.Equal(int32(0), res.Errors)
}
