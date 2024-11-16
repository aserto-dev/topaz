package directory

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path"

	dse3 "github.com/aserto-dev/go-directory/aserto/directory/exporter/v3"
	"github.com/aserto-dev/topaz/pkg/cli/js"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func (c *Client) Backup(ctx context.Context, file string) error {
	stream, err := c.Exporter.Export(ctx, &dse3.ExportRequest{
		Options:   uint32(dse3.Option_OPTION_DATA),
		StartFrom: &timestamppb.Timestamp{},
	})
	if err != nil {
		return err
	}

	tmpDir, err := os.MkdirTemp("", "*")
	if err != nil {
		return err
	}
	defer func() {
		_ = os.RemoveAll(tmpDir)
	}()

	dirPath := path.Join(tmpDir, "backup")
	if err := os.MkdirAll(dirPath, 0o700); err != nil {
		return err
	}

	if err := c.createBackupFiles(stream, dirPath); err != nil {
		return err
	}

	tf, err := os.Create(file)
	if err != nil {
		return nil
	}
	defer func() {
		tf.Close()
	}()

	gw, err := gzip.NewWriterLevel(tf, gzip.BestCompression)
	if err != nil {
		return nil
	}
	defer func() {
		gw.Close()
	}()

	tw := tar.NewWriter(gw)
	defer func() {
		tw.Close()
	}()

	_ = addToArchive(tw, path.Join(dirPath, ObjectsFileName))
	_ = addToArchive(tw, path.Join(dirPath, RelationsFileName))

	return nil
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

func (c *Client) createBackupFiles(stream dse3.Exporter_ExportClient, dirPath string) error {
	objects, err := js.NewWriter(path.Join(dirPath, ObjectsFileName), ObjectsStr)
	if err != nil {
		return err
	}
	defer objects.Close()

	relations, err := js.NewWriter(path.Join(dirPath, RelationsFileName), RelationsStr)
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

		switch m := msg.Msg.(type) {
		case *dse3.ExportResponse_Object:
			err = objects.Write(m.Object)
			objectsCounter.Incr().Print(os.Stdout)

		case *dse3.ExportResponse_Relation:
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
