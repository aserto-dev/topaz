package directory

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"

	dsc "github.com/aserto-dev/go-directory/aserto/directory/common/v3"
	dsi "github.com/aserto-dev/go-directory/aserto/directory/importer/v3"
	"github.com/aserto-dev/topaz/topaz/jsonx"
	"golang.org/x/sync/errgroup"
)

func (c *Client) ImportFromFile(ctx context.Context, r io.Reader) error {
	errGrp, errGrpCtx := errgroup.WithContext(ctx)

	stream, err := c.Importer.Import(errGrpCtx)
	if err != nil {
		return err
	}

	errGrp.Go(c.recv(stream))

	errGrp.Go(c.sender(stream, r))

	return errGrp.Wait()
}

func (c *Client) recv(stream dsi.Importer_ImportClient) func() error {
	return func() error {
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
				_ = m.Status

			case *dsi.ImportResponse_Counter:
				_ = m.Counter

			default:
				_ = m
			}
		}
	}
}

func (c *Client) sender(stream dsi.Importer_ImportClient, r io.Reader) func() error {
	return func() error {
		if err := c.importFromReader(stream, r); err != nil {
			return err
		}

		if err := stream.CloseSend(); err != nil {
			return err
		}

		return nil
	}
}

func (c *Client) importFromReader(stream dsi.Importer_ImportClient, r io.Reader) error {
	scanner := bufio.NewScanner(r)

	for scanner.Scan() {
		obj := &dsc.Object{}
		rel := &dsc.Relation{}

		if err := jsonx.Unmarshal(scanner.Bytes(), obj); err == nil {
			if err := stream.Send(&dsi.ImportRequest{
				OpCode: dsi.Opcode_OPCODE_SET,
				Msg: &dsi.ImportRequest_Object{
					Object: obj,
				},
			}); err != nil {
				return err
			}

			continue
		}

		if err := jsonx.Unmarshal(scanner.Bytes(), rel); err == nil {
			if err := stream.Send(&dsi.ImportRequest{
				OpCode: dsi.Opcode_OPCODE_SET,
				Msg: &dsi.ImportRequest_Relation{
					Relation: rel,
				},
			}); err != nil {
				return err
			}

			continue
		}
	}

	return scanner.Err()
}

func ObjStr(obj *dsc.Object) string {
	return fmt.Sprintf("[%s:%s]",
		obj.GetType(),
		obj.GetId(),
	)
}

func RelStr(rel *dsc.Relation) string {
	return fmt.Sprintf("[%s:%s#%s@%s:%s#%s]",
		rel.GetObjectType(),
		rel.GetObjectId(),
		rel.GetRelation(),
		rel.GetSubjectType(),
		rel.GetSubjectId(),
		rel.GetSubjectRelation(),
	)
}
