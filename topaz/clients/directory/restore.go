package directory

import (
	"archive/tar"
	"bufio"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	dsc "github.com/aserto-dev/go-directory/aserto/directory/common/v3"
	dsi "github.com/aserto-dev/go-directory/aserto/directory/importer/v3"
	"github.com/aserto-dev/topaz/internal/fs"
	"github.com/aserto-dev/topaz/topaz/js"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"golang.org/x/sync/errgroup"
)

func (c *Client) Restore(ctx context.Context, file string) error {
	tf, err := os.Open(file)
	if err != nil {
		return err
	}
	defer tf.Close()

	gz, err := gzip.NewReader(tf)
	if err != nil {
		return err
	}
	defer gz.Close()

	tr := tar.NewReader(gz)

	g, iCtx := errgroup.WithContext(context.Background())

	stream, err := c.Importer.Import(iCtx)
	if err != nil {
		return err
	}

	g.Go(c.receiver(stream))

	g.Go(c.restoreHandler(stream, tr))

	return g.Wait()
}

func (c *Client) receiver(stream dsi.Importer_ImportClient) func() error {
	return func() error {
		objCounter := &dsi.ImportCounter{Type: objectsCounter}
		relCounter := &dsi.ImportCounter{Type: relationsCounter}

		defer func() {
			printCounter(os.Stderr, objCounter)
			printCounter(os.Stderr, relCounter)
		}()

		for {
			msg, err := stream.Recv()
			if errors.Is(err, io.EOF) {
				return nil
			}

			if err != nil {
				return err
			}

			switch m := msg.GetMsg().(type) {
			case *dsi.ImportResponse_Status:
				printStatus(os.Stderr, m.Status)

			case *dsi.ImportResponse_Counter:
				switch m.Counter.GetType() {
				case objectsCounter:
					objCounter = m.Counter
				case relationsCounter:
					relCounter = m.Counter
				}

			// handle obsolete message usage as the default.
			//nolint:staticcheck // SA1019
			default:
				msg.Object.Type = objectsCounter
				objCounter = msg.GetObject()
				msg.Relation.Type = relationsCounter
				relCounter = msg.GetRelation()
			}
		}
	}
}

func (c *Client) restoreHandler(stream dsi.Importer_ImportClient, tr *tar.Reader) func() error {
	return func() error {
		for {
			header, err := tr.Next()
			if errors.Is(err, io.EOF) {
				break
			}

			if err != nil {
				return err
			}

			if header == nil || header.Typeflag != tar.TypeReg {
				continue
			}

			r, err := js.NewReader(tr)
			if err != nil {
				return err
			}

			name := path.Clean(header.Name)
			switch name {
			case ObjectsFileName:
				if err := c.loadObjects(stream, r); err != nil {
					return err
				}

			case RelationsFileName:
				if err := c.loadRelations(stream, r); err != nil {
					return err
				}
			}
		}

		if err := stream.CloseSend(); err != nil {
			return err
		}

		return nil
	}
}

func (c *Client) loadObjects(stream dsi.Importer_ImportClient, objects *js.Reader) error {
	defer objects.Close()

	var m dsc.Object

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

		if err := stream.Send(&dsi.ImportRequest{
			OpCode: dsi.Opcode_OPCODE_SET,
			Msg: &dsi.ImportRequest_Object{
				Object: &m,
			},
		}); err != nil {
			return err
		}
	}

	return nil
}

func (c *Client) loadRelations(stream dsi.Importer_ImportClient, relations *js.Reader) error {
	defer relations.Close()

	var m dsc.Relation

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

		if err := stream.Send(&dsi.ImportRequest{
			OpCode: dsi.Opcode_OPCODE_SET,
			Msg: &dsi.ImportRequest_Relation{
				Relation: &m,
			},
		}); err != nil {
			return err
		}
	}

	return nil
}

const (
	maxManifestSize int64 = 4 * 1024 * 1024        // 4 MB
	maxJsonlSize    int64 = 4 * 1024 * 1024 * 1024 // 4 GB
)

func (c *Client) RestoreFromFile(ctx context.Context, r io.Reader) error {
	// step 0 -- create temp directory to export tarbal artifacts to.
	tmpDir, err := os.MkdirTemp("", "topaz")
	if err != nil {
		return err
	}

	defer func() {
		_ = os.RemoveAll(tmpDir)
	}()

	dirPath := filepath.Join(tmpDir, "restore")
	if err := os.MkdirAll(dirPath, fs.FileModeOwnerRWX); err != nil {
		return err
	}

	gz, err := gzip.NewReader(r)
	if err != nil {
		return err
	}
	defer gz.Close()

	tr := tar.NewReader(gz)

	if err := c.extractFiles(tr, dirPath); err != nil {
		return err
	}

	{
		r, err := os.Open(filepath.Join(dirPath, ManifestFile))
		if err != nil {
			return err
		}
		defer r.Close()

		if err := c.SetManifest(ctx, r); err != nil {
			return err
		}
	}

	{
		r, err := os.Open(filepath.Join(dirPath, ObjectsFile))
		if err != nil {
			return err
		}
		defer r.Close()

		if err := c.ImportFromFile(ctx, r); err != nil {
			return err
		}
	}

	{
		r, err := os.Open(filepath.Join(dirPath, RelationsFile))
		if err != nil {
			return err
		}
		defer r.Close()

		if err := c.ImportFromFile(ctx, r); err != nil {
			return err
		}
	}

	return nil
}

func (*Client) extractFiles(tr *tar.Reader, dirPath string) error {
	processedFiles := make(map[string]bool)

	for {
		header, err := tr.Next()
		if errors.Is(err, io.EOF) {
			break
		}

		if err != nil {
			return fmt.Errorf("failed reading tar header: %w", err)
		}

		// Explicitly reject symlinks and hardlinks
		if header.Typeflag == tar.TypeSymlink || header.Typeflag == tar.TypeLink {
			return status.Error(codes.Internal, "security violation: symbolic or hard links are strictly forbidden")
		}

		// skip anything that os not a regular file
		if header.Typeflag != tar.TypeReg {
			continue
		}
		// Path validation & Zip-Slip defense
		cleanedName := filepath.Clean(header.Name)
		if strings.HasPrefix(cleanedName, "..") || strings.HasPrefix(cleanedName, "/") {
			return status.Errorf(codes.Internal, "security violation: invalid file path traversal detected: %s", header.Name)
		}

		// Extract just the file name to handle files safely even if they are nested inside a root folder
		baseName := filepath.Base(cleanedName)

		var sizeLimit int64

		// Enforce max file size validation
		switch baseName {
		case ManifestFile:
			sizeLimit = maxManifestSize
		case ObjectsFile:
			fallthrough
		case RelationsFile:
			sizeLimit = maxJsonlSize
		default:
			return status.Errorf(codes.Internal, "unexpected file type found in tarball: %s", baseName)
		}

		// Prevent tracking/writing the same file target twice
		if processedFiles[baseName] {
			return status.Errorf(codes.Internal, "duplicate file entry detected: %s", baseName)
		}

		processedFiles[baseName] = true

		// Stream to Temp Disk Storage safely
		targetPath := filepath.Join(dirPath, baseName)
		if err := writeStreamToDisk(tr, targetPath, sizeLimit); err != nil {
			return status.Errorf(codes.Internal, "failed writing %s to disk: %s", baseName, err.Error())
		}
	}

	return nil
}

func writeStreamToDisk(tr *tar.Reader, destPath string, limit int64) error {
	outFile, err := os.OpenFile(destPath, os.O_CREATE|os.O_RDWR|os.O_TRUNC, fs.FileModeOwnerRW)
	if err != nil {
		return err
	}
	defer outFile.Close()

	bufferedWriter := bufio.NewWriter(outFile)
	defer bufferedWriter.Flush()

	// Enforce strict payload limitations.
	limitedReader := io.LimitReader(tr, limit)

	written, err := io.Copy(bufferedWriter, limitedReader)
	if err != nil {
		return err
	}

	// Check if the file went over your threshold boundaries.
	if written == limit {
		var singleByte [1]byte
		if n, _ := tr.Read(singleByte[:]); n > 0 {
			return status.Errorf(codes.Aborted, "file size exceeded allowed limit of %d bytes", limit)
		}
	}

	return nil
}
