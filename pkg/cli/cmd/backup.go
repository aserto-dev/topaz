package cmd

import (
	"context"
	"os"
	"path"

	asertogoClient "github.com/aserto-dev/aserto-go/client"
	"github.com/aserto-dev/go-edge-ds/pkg/client"
	"github.com/aserto-dev/topaz/pkg/cli/cc"
	"github.com/aserto-dev/topaz/pkg/cli/dockerx"
	"github.com/fatih/color"
)

type BackupCmd struct {
	File string `arg:""  default:"backup.tar.gz" help:"absolute file path to make backup to"`
}

var defaultFileName = "backup.tar.gz"

func (cmd BackupCmd) Run(c *cc.CommonCtx) error {
	if running, err := dockerx.IsRunning(dockerx.Topaz); !running || err != nil {
		if err != nil {
			return err
		}
		color.Yellow("!!! topaz is not running")
		return nil
	}

	conn, err := asertogoClient.NewConnection(context.Background(), asertogoClient.WithInsecure(true), asertogoClient.WithAddr("localhost:9292"))
	if err != nil {
		return err
	}

	dirClient, err := client.New(conn)
	if err != nil {
		return err
	}

	if cmd.File == defaultFileName {
		currentDir, err := os.Getwd()
		if err != nil {
			return err
		}
		cmd.File = path.Join(currentDir, defaultFileName)
	}

	color.Green(">>> starting backup...")
	return dirClient.Backup(c.Context, cmd.File)
}
