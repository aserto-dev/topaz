package cmd

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pkg/errors"
	bolt "go.etcd.io/bbolt"
)

const roFileMode os.FileMode = 0o400

type CLI struct {
	DBFile    string `flag:"" help:"database file path" type:"existingfile"`
	BackupDir string `flag:"" help:"backup directory path" type:"existingdir"`
}

func (cmd *CLI) Run(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	backupFileName := backupFileName(cmd)

	// open db in read-only mode
	db, err := bolt.Open(cmd.DBFile, roFileMode, &bolt.Options{ReadOnly: true})
	if err != nil {
		return errors.Errorf("failed to open database file (%s): %v", cmd.DBFile, err)
	}
	defer db.Close()

	// create backup file
	backup, err := os.Create(backupFileName)
	if err != nil {
		return errors.Errorf("failed to create backup file (%s): %v", backupFileName, err)
	}
	defer backup.Close()

	// backup RO database to backup file
	if err := backupDB(ctx, db, backup); err != nil {
		return err
	}

	// flush backup file to disk
	if err := backup.Sync(); err != nil {
		return errors.Errorf("failed to sync backup file: %v", err)
	}

	return nil
}

func backupFileName(cmd *CLI) string {
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
			return errors.Errorf("backup failed: %v", err)
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
