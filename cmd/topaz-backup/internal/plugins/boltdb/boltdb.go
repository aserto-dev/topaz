package boltdb

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aserto-dev/topaz/cmd/topaz-backup/internal/plugin"
	"github.com/pkg/errors"
	bolt "go.etcd.io/bbolt"
)

// BoltDB plugin.
const (
	PluginName string = "boltdb"
	PluginDesc string = "boltdb plugin"
)

var _ plugin.StorePlugin = &Plugin{}

type Plugin struct {
	DBFile    string `flag:"" help:"database file path" type:"existingfile" required:""`
	BackupDir string `flag:"" help:"backup directory path" type:"existingdir" required:""`
}

func (cmd *Plugin) Run(ctx context.Context) error {
	backupFileName := backupFileName(cmd)

	// open DB in read-only mode.
	roDB, err := bolt.Open(cmd.DBFile, plugin.ReadOnly, &bolt.Options{ReadOnly: true})
	if err != nil {
		return errors.Errorf("failed to open database file (%s): %v", cmd.DBFile, err)
	}
	defer roDB.Close()

	// create backup file.
	backup, err := os.Create(backupFileName)
	if err != nil {
		return errors.Errorf("failed to create backup file (%s): %v", backupFileName, err)
	}
	defer backup.Close()

	// backup RO database to backup file.
	if err := backupDB(ctx, roDB, backup); err != nil {
		return err
	}

	// flush backup file to disk.
	if err := backup.Sync(); err != nil {
		return errors.Errorf("failed to sync backup file: %v", err)
	}

	fmt.Printf("\n%s\n", backupFileName)

	return nil
}

func backupFileName(cmd *Plugin) string {
	ts := time.Now().Format("20060102T150405")

	_, file := filepath.Split(cmd.DBFile)
	ext := filepath.Ext(file)
	base := strings.TrimSuffix(file, ext)
	backupFile := filepath.Join(cmd.BackupDir, base+"-"+ts+ext)

	return backupFile
}

func backupDB(ctx context.Context, db *bolt.DB, backup *os.File) error {
	errCh := make(chan error, 1)

	go func() {
		err := db.View(func(tx *bolt.Tx) error {
			_, err := tx.WriteTo(backup)
			if err != nil {
				return errors.Errorf("backup failed: %v", err)
			}

			return nil
		})
		errCh <- err
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-errCh:
		return err
	}
}
