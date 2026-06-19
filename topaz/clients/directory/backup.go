package directory

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	dse "github.com/aserto-dev/go-directory/aserto/directory/exporter/v3"
	"github.com/aserto-dev/topaz/internal/fs"
	"github.com/aserto-dev/topaz/topaz/js"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func (c *Client) Backup(ctx context.Context, file string) error {
	stream, err := c.Exporter.Export(ctx, &dse.ExportRequest{
		Options:   uint32(dse.Option_OPTION_DATA),
		StartFrom: &timestamppb.Timestamp{},
	})
	if err != nil {
		return err
	}

	tmpDir, err := os.MkdirTemp("", "topaz")
	if err != nil {
		return err
	}

	defer func() {
		_ = os.RemoveAll(tmpDir)
	}()

	dirPath := filepath.Join(tmpDir, "backup")
	if err := os.MkdirAll(dirPath, fs.FileModeOwnerRWX); err != nil {
		return err
	}

	if err := c.createBackupFiles(stream, dirPath); err != nil {
		return err
	}

	tf, err := os.Create(file)
	if err != nil {
		return err
	}

	defer func() {
		_ = tf.Close()
	}()

	gw, err := gzip.NewWriterLevel(tf, gzip.BestCompression)
	if err != nil {
		return err
	}

	defer func() {
		_ = gw.Close()
	}()

	tw := tar.NewWriter(gw)

	defer func() {
		_ = tw.Close()
	}()

	if err := addToArchive(tw, filepath.Join(dirPath, ObjectsFileName)); err != nil {
		return err
	}

	if err := addToArchive(tw, filepath.Join(dirPath, RelationsFileName)); err != nil {
		return err
	}

	return nil
}

func (c *Client) createBackupFiles(stream dse.Exporter_ExportClient, dirPath string) error {
	objects, err := js.NewWriter(filepath.Join(dirPath, ObjectsFileName), ObjectsStr)
	if err != nil {
		return err
	}
	defer objects.Close()

	relations, err := js.NewWriter(filepath.Join(dirPath, RelationsFileName), RelationsStr)
	if err != nil {
		return err
	}
	defer relations.Close()

	ctr := &Counter{}
	objectsCounter := ctr.Objects()
	relationsCounter := ctr.Relations()

	for {
		msg, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			break
		}

		if err != nil {
			return err
		}

		switch m := msg.GetMsg().(type) {
		case *dse.ExportResponse_Object:
			err = objects.Write(m.Object)

			objectsCounter.Incr().Print(os.Stdout)

		case *dse.ExportResponse_Relation:
			err = relations.Write(m.Relation)

			relationsCounter.Incr().Print(os.Stdout)

		default:
			fmt.Fprintf(os.Stdout, "Unknown message type\n")
		}

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		}
	}

	ctr.Print(os.Stdout)

	return nil
}

const (
	ManifestFile  string = "manifest.yaml"
	ObjectsFile   string = "objects.jsonl"
	RelationsFile string = "relations.jsonl"
)

func (c *Client) BackupToFile(ctx context.Context, w io.Writer) error {
	// step 0 -- create temp directory folder to gather artifacts
	tmpDir, err := os.MkdirTemp("", "topaz")
	if err != nil {
		return err
	}

	defer func() {
		_ = os.RemoveAll(tmpDir)
	}()

	dirPath := filepath.Join(tmpDir, "backup")
	if err := os.MkdirAll(dirPath, fs.FileModeOwnerRWX); err != nil {
		return err
	}

	// step 1 - download manifest.yaml
	manifestFQN := filepath.Join(dirPath, ManifestFile)
	if err := c.getManifestFile(ctx, manifestFQN); err != nil {
		return err
	}

	// step 2 -- download objects.jsonl
	objectsFQN := filepath.Join(dirPath, ObjectsFile)
	if err := c.getObjectsFile(ctx, objectsFQN); err != nil {
		return err
	}

	// step 3 -- download relations.jsonl
	relationsFQN := filepath.Join(dirPath, RelationsFile)
	if err := c.getRelationsFile(ctx, relationsFQN); err != nil {
		return err
	}

	// step 4 -- create tarbal from tmp directory
	fqns := []string{
		manifestFQN,
		objectsFQN,
		relationsFQN,
	}

	if err := c.createTar(w, fqns); err != nil {
		return err
	}

	return nil
}

func (*Client) createTar(w io.Writer, fqns []string) error {
	gw, err := gzip.NewWriterLevel(w, gzip.BestCompression)
	if err != nil {
		return err
	}

	defer func() {
		_ = gw.Close()
	}()

	tw := tar.NewWriter(gw)

	defer func() {
		_ = tw.Close()
	}()

	for _, fqn := range fqns {
		if err := addToArchive(tw, fqn); err != nil {
			return err
		}
	}

	return nil
}

func (c *Client) getManifestFile(ctx context.Context, fqn string) error {
	manReader, err := c.GetManifest(ctx)
	if err != nil {
		return err
	}

	manWriter, err := os.Create(fqn)
	if err != nil {
		return err
	}
	defer manWriter.Close()

	if _, err := io.Copy(manWriter, manReader); err != nil {
		return err
	}

	return nil
}

func (c *Client) getObjectsFile(ctx context.Context, fqn string) error {
	objWriter, err := os.Create(fqn)
	if err != nil {
		return err
	}
	defer objWriter.Close()

	return c.ExportToFile(ctx, objWriter, uint32(dse.Option_OPTION_DATA_OBJECTS))
}

func (c *Client) getRelationsFile(ctx context.Context, fqn string) error {
	relWriter, err := os.Create(fqn)
	if err != nil {
		return err
	}
	defer relWriter.Close()

	return c.ExportToFile(ctx, relWriter, uint32(dse.Option_OPTION_DATA_RELATIONS))
}

func addToArchive(tw *tar.Writer, filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return err
	}

	header, err := tar.FileInfoHeader(info, info.Name())
	if err != nil {
		return err
	}

	if err := tw.WriteHeader(header); err != nil {
		return err
	}

	if _, err := io.Copy(tw, file); err != nil {
		return err
	}

	return nil
}
