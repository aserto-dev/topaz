package engine_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	api "github.com/aserto-dev/go-grpc/aserto/api/v1"
	"github.com/aserto-dev/go-grpc/aserto/authorizer/directory/v1"
	"github.com/aserto-dev/go-lib/ids"
	"github.com/aserto-dev/go-utils/cerr"
	"github.com/aserto-dev/topaz/pkg/cc/config"
	atesting "github.com/aserto-dev/topaz/pkg/testing"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

func setupDirectory(t *testing.T) (*atesting.EngineHarness, func()) {
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
		t.Fail()
	}

	harness := atesting.SetupOffline(t, func(cfg *config.Config) {
		cfg.Directory.Path = filepath.Join(tmpDir, "test.db")
		t.Logf("dbpath %s", cfg.Directory.Path)
	})

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

	return harness, func() {
		harness.Cleanup()
		t.Logf("cleanup db")
		dbpath := harness.Engine.Configuration.Directory.Path
		t.Logf("delete scratch db [%s]", dbpath)
		os.Remove(dbpath)
	}

}

func TestGetRolesNoUserError(t *testing.T) {
	harness, teardownSuite := setupDirectory(t)
	defer teardownSuite()
	assert := require.New(t)

	client := harness.CreateGRPCDirectoryClient().Directory

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	uid, err := ids.NewAccountID()
	assert.Nil(err)

	_, err = client.GetApplRoles(ctx, &directory.GetApplRolesRequest{
		Id:   uid,
		Name: "test",
	})
	assert.Equal(cerr.ErrUserNotFound, cerr.UnwrapAsertoError(err))
}
