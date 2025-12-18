package tests_test

import (
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
	"testing"
	"time"

	dsc3 "github.com/aserto-dev/go-directory/aserto/directory/common/v3"
	dsi3 "github.com/aserto-dev/go-directory/aserto/directory/importer/v3"
	dsr3 "github.com/aserto-dev/go-directory/aserto/directory/reader/v3"
	dsw3 "github.com/aserto-dev/go-directory/aserto/directory/writer/v3"
	"github.com/aserto-dev/topaz/internal/pkg/eds/pkg/directory"
	"github.com/aserto-dev/topaz/internal/pkg/eds/pkg/server"
	"github.com/aserto-dev/topaz/internal/pkg/fs"
	"github.com/pkg/errors"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

type TestCase struct {
	Name   string
	Req    proto.Message
	Checks func(*testing.T, proto.Message, error) func(proto.Message)
}

var (
	client *server.TestEdgeClient
	closer func()
)

func TestMain(m *testing.M) {
	zerolog.SetGlobalLevel(zerolog.WarnLevel)

	logger := zerolog.New(io.Discard)

	ctx, cancel := context.WithCancel(context.Background())

	dirPath := os.TempDir()
	if err := fs.EnsureDirPath(dirPath, fs.FileModeOwnerRWX); err != nil {
		panic(err)
	}

	dbPath := path.Join(dirPath, "edge-ds", "test-eds.db")
	os.Remove(dbPath)
	fmt.Println(dbPath)

	cfg := directory.Config{
		DBPath:         dbPath,
		RequestTimeout: time.Second * 2,
		Seed:           true,
		EnableV2:       true,
	}

	client, closer = server.NewTestEdgeServer(ctx, &logger, &cfg)

	exitVal := m.Run()

	closer()
	cancel()

	os.Exit(exitVal)
}

func importFile(stream dsi3.Importer_ImportClient, file string) error {
	r, err := os.Open(file)
	if err != nil {
		return errors.Wrapf(err, "failed to open file: [%s]", file)
	}
	defer r.Close()

	reader, err := NewReader(r)
	if err != nil || reader == nil {
		fmt.Fprintf(os.Stderr, "Skipping file [%s]: [%s]\n", file, err.Error())
		return nil
	}
	defer reader.Close()

	objectType := reader.GetObjectType()
	switch objectType {
	case ObjectsStr:
		if err := loadObjects(stream, reader); err != nil {
			return err
		}

	case RelationsStr:
		if err := loadRelations(stream, reader); err != nil {
			return err
		}

	default:
		fmt.Fprintf(os.Stderr, "skipping file [%s] with object type [%s]\n", file, objectType)
	}

	return nil
}

func loadObjects(stream dsi3.Importer_ImportClient, objects *Reader) error {
	defer objects.Close()

	var m dsc3.Object

	for {
		err := objects.Read(&m)
		if errors.Is(err, io.EOF) {
			break
		}

		if err != nil {
			if strings.Contains(err.Error(), "unknown field") {
				continue
			}

			return err
		}

		if err := stream.Send(&dsi3.ImportRequest{
			OpCode: dsi3.Opcode_OPCODE_SET,
			Msg: &dsi3.ImportRequest_Object{
				Object: &m,
			},
		}); err != nil {
			return err
		}
	}

	return nil
}

func loadRelations(stream dsi3.Importer_ImportClient, relations *Reader) error {
	defer relations.Close()

	var m dsc3.Relation

	for {
		err := relations.Read(&m)
		if errors.Is(err, io.EOF) {
			break
		}

		if err != nil {
			if strings.Contains(err.Error(), "unknown field") {
				continue
			}

			return err
		}

		if err := stream.Send(&dsi3.ImportRequest{
			OpCode: dsi3.Opcode_OPCODE_SET,
			Msg: &dsi3.ImportRequest_Relation{
				Relation: &m,
			},
		}); err != nil {
			return err
		}
	}

	return nil
}

func testInit() (*server.TestEdgeClient, func()) {
	return client, func() {}
}

func testRunner(t *testing.T, tcs []*TestCase) {
	client, cleanup := testInit()
	t.Cleanup(cleanup)

	ctx := t.Context()

	manifest, err := os.ReadFile("./manifest_v3_test.yaml")
	require.NoError(t, err)

	require.NoError(t, deleteManifest(client))
	require.NoError(t, setManifest(client, manifest))

	var apply func(proto.Message)

	for _, tc := range tcs {
		t.Run(tc.Name, func(t *testing.T) {
			if apply != nil {
				apply(tc.Req)
			}

			runTestCase(ctx, t, tc)
		})
	}
}

func runTestCase(ctx context.Context, t *testing.T, tc *TestCase) func(proto.Message) {
	switch req := tc.Req.(type) {
	// V3
	///////////////////////////////////////////////////////////////
	case *dsr3.GetObjectRequest:
		resp, err := client.V3.Reader.GetObject(ctx, req)
		return tc.Checks(t, resp, err)

	case *dsw3.SetObjectRequest:
		resp, err := client.V3.Writer.SetObject(ctx, req)
		return tc.Checks(t, resp, err)

	case *dsw3.DeleteObjectRequest:
		resp, err := client.V3.Writer.DeleteObject(ctx, req)
		return tc.Checks(t, resp, err)

	case *dsr3.GetRelationRequest:
		resp, err := client.V3.Reader.GetRelation(ctx, req)
		return tc.Checks(t, resp, err)

	case *dsw3.SetRelationRequest:
		resp, err := client.V3.Writer.SetRelation(ctx, req)
		return tc.Checks(t, resp, err)

	case *dsw3.DeleteRelationRequest:
		resp, err := client.V3.Writer.DeleteRelation(ctx, req)
		return tc.Checks(t, resp, err)

	case *dsr3.GetRelationsRequest:
		resp, err := client.V3.Reader.GetRelations(ctx, req)
		return tc.Checks(t, resp, err)
	}

	return func(proto.Message) {}
}
