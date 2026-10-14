package main

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/alecthomas/kong"
	"github.com/aserto-dev/go-directory/aserto/directory/common/v3"
	"github.com/aserto-dev/go-directory/aserto/directory/migrate/v33"
	"github.com/aserto-dev/topaz/topaz/js"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/structpb"
)

type CLI struct {
	Input  string `flag:"" short:"i" help:"input file"  type:"existingfile"`
	Output string `flag:"" short:"o" help:"output file" type:"path"`
	Format string `flag:"" short:"f" help:"output json format" default:"line" enum:"line,objects"`
}

var marshalLine = protojson.MarshalOptions{
	Multiline:         false,
	Indent:            "",
	AllowPartial:      true,
	UseProtoNames:     true,
	UseEnumNumbers:    false,
	EmitUnpopulated:   false,
	EmitDefaultValues: false,
}

var marshalObject = protojson.MarshalOptions{
	Multiline:         true,
	Indent:            "  ",
	AllowPartial:      true,
	UseProtoNames:     true,
	UseEnumNumbers:    false,
	EmitUnpopulated:   false,
	EmitDefaultValues: false,
}

func main() {
	cli := CLI{}

	kongCtx := kong.Parse(&cli)
	kongCtx.Run()
}

func (cmd *CLI) Run() error {
	reader, err := os.Open(cmd.Input)
	if err != nil {
		return err
	}
	defer reader.Close()

	jr, err := js.NewReader(reader)
	if err != nil {
		return err
	}
	defer jr.Close()

	w := os.Stdout
	if cmd.Output != "" {
		out, err := os.Create(cmd.Output)
		if err != nil {
			return err
		}
		w = out
	}
	defer w.Close()

	fmt.Fprintf(os.Stdout, "type %q\n", jr.GetObjectType())

	var o33 migrate.Object33

	for {
		err := jr.Read(&o33)
		if errors.Is(err, io.EOF) {
			break
		}

		if err != nil {
			return err
		}

		x := &common.Object{
			ObjectType: o33.GetType(),
			ObjectId:   o33.GetId(),
			Properties: o33.GetProperties(),
			UpdatedAt:  o33.GetUpdatedAt(),
			Etag:       o33.GetEtag(),
		}

		if dn := o33.GetDisplayName(); dn != "" {
			x.Properties.Fields["display_name"] = structpb.NewStringValue(dn)
		}

		buf, err := marshalLine.Marshal(x)

		w.Write(buf)
		w.WriteString("\n")

	}

	return nil
}
