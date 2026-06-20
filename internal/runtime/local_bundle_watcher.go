package runtime

import (
	"context"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/open-policy-agent/opa/v1/ast"
	"github.com/open-policy-agent/opa/v1/bundle"
	"github.com/open-policy-agent/opa/v1/loader"
	"github.com/open-policy-agent/opa/v1/metrics"
	"github.com/open-policy-agent/opa/v1/storage"
	"github.com/open-policy-agent/opa/v1/version"
	"github.com/pkg/errors"
)

func (r *Runtime) onReloadLogger(d time.Duration, err error) {
	r.Logger.Warn().
		Dur("duration", d).
		Err(err).
		Msg("Processed file watch event.")
}

func (r *Runtime) startWatcher(ctx context.Context, paths []string, onReload func(time.Duration, error)) error {
	watcher, err := r.getWatcher(paths)
	if err != nil {
		return err
	}

	go r.readWatcher(ctx, watcher, paths, onReload)

	return nil
}

func (r *Runtime) getWatcher(rootPaths []string) (*fsnotify.Watcher, error) {
	watchPaths, err := getWatchPaths(rootPaths)
	if err != nil {
		return nil, err
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	for _, path := range watchPaths {
		r.Logger.Debug().Str("path", path).Msg("watching path")

		if err := watcher.Add(path); err != nil {
			return nil, err
		}
	}

	if r.Config.LocalBundles.LocalPolicyImage != "" {
		if err := watcher.Add(filepath.Join(r.Config.LocalBundles.FileStoreRoot, "policies-root", "index.json")); err != nil {
			return nil, err
		}
	}

	return watcher, nil
}

func getWatchPaths(rootPaths []string) ([]string, error) {
	paths := []string{}

	for _, path := range rootPaths {
		_, path = loader.SplitPrefix(path)

		result, err := loader.Paths(path, true)
		if err != nil {
			return nil, err
		}

		paths = append(paths, loader.Dirs(result)...)
	}

	return paths, nil
}

func (r *Runtime) readWatcher(ctx context.Context, watcher *fsnotify.Watcher, paths []string, onReload func(time.Duration, error)) {
	for {
		evt := <-watcher.Events
		removalMask := (fsnotify.Remove | fsnotify.Rename)

		mask := (fsnotify.Create | fsnotify.Write | removalMask)
		if (evt.Op & mask) != 0 {
			r.Logger.Debug().Str("event", evt.String()).Msg("registered file event")

			t0 := time.Now()
			removed := ""

			if (evt.Op & removalMask) != 0 {
				removed = evt.Name
			}

			err := r.processWatcherUpdate(ctx, paths, removed)
			onReload(time.Since(t0), err)
		}
	}
}

func (r *Runtime) processWatcherUpdate(ctx context.Context, paths []string, removed string) error {
	if r.Config.LocalBundles.LocalPolicyImage != "" {
		if err := r.deactivate(ctx); err != nil {
			return err
		}
	}

	loadedBundles, err := r.loadPaths(paths)
	if err != nil {
		return err
	}

	if removed != "" {
		r.Logger.Debug().Msgf("Removed event name value: %v", removed)
	}

	return storage.Txn(ctx, r.storage, storage.WriteParams, func(txn storage.Transaction) error {
		_, err = insertAndCompile(ctx, &insertAndCompileOptions{
			Store:     r.storage,
			Txn:       txn,
			Bundles:   loadedBundles,
			MaxErrors: -1,
		})
		if err != nil {
			return err
		}

		return nil
	})
}

const (
	bundleRootOffset = 3
	rootindexOffset  = 2
)

func (r *Runtime) deactivate(ctx context.Context) error {
	err := storage.Txn(ctx, r.storage, storage.WriteParams, func(txn storage.Transaction) error {
		deactivateMap := make(map[string]struct{})

		policies, err := r.storage.ListPolicies(ctx, txn)
		if err != nil {
			return err
		}

		if len(policies) == 0 {
			return nil
		}

		path := strings.Split(policies[0], "/")
		rootIndex := len(path) - bundleRootOffset // default bundle root.

		// bundle root detection for build images.
		for i := range path {
			if path[i] == "sha256" {
				rootIndex = i + rootindexOffset
				break
			}
		}

		root := strings.Join(path[:rootIndex], "/")
		deactivateMap[root] = struct{}{}

		return bundle.Deactivate(&bundle.DeactivateOpts{
			Ctx:         ctx,
			Store:       r.storage,
			Txn:         txn,
			BundleNames: deactivateMap,
		})
	})

	return err
}

// insertAndCompileOptions contains input for the operation.
type insertAndCompileOptions struct {
	Store     storage.Store
	Txn       storage.Transaction
	Files     loader.Result
	Bundles   map[string]*bundle.Bundle
	MaxErrors int
}

// insertAndCompileResult contains the output of the operation.
type insertAndCompileResult struct {
	Compiler *ast.Compiler
	Metrics  metrics.Metrics
}

// insertAndCompile writes data and policy into the store and returns a compiler for the
// store contents.
func insertAndCompile(ctx context.Context, opts *insertAndCompileOptions) (*insertAndCompileResult, error) {
	if len(opts.Files.Documents) > 0 {
		if err := opts.Store.Write(ctx, opts.Txn, storage.AddOp, storage.Path{}, opts.Files.Documents); err != nil {
			return nil, errors.Wrap(err, "storage error")
		}
	}

	policies := make(map[string]*ast.Module, len(opts.Files.Modules))

	for id, parsed := range opts.Files.Modules {
		policies[id] = parsed.Parsed
	}

	compiler := ast.NewCompiler().SetErrorLimit(opts.MaxErrors).WithPathConflictsCheck(storage.NonEmpty(ctx, opts.Store, opts.Txn))
	m := metrics.New()

	activation := &bundle.ActivateOpts{
		Ctx:          ctx,
		Store:        opts.Store,
		Txn:          opts.Txn,
		Compiler:     compiler,
		Metrics:      m,
		Bundles:      opts.Bundles,
		ExtraModules: policies,
	}

	err := bundle.Activate(activation)
	if err != nil {
		return nil, err
	}

	// Policies in bundles will have already been added to the store, but
	// modules loaded outside of bundles will need to be added manually.
	for id, parsed := range opts.Files.Modules {
		if err := opts.Store.UpsertPolicy(ctx, opts.Txn, id, parsed.Raw); err != nil {
			return nil, errors.Wrap(err, "storage error")
		}
	}

	// Set the version in the store last to prevent data files from overwriting.
	if err := writeVersion(ctx, opts.Store, opts.Txn); err != nil {
		return nil, errors.Wrap(err, "storage error")
	}

	return &insertAndCompileResult{Compiler: compiler, Metrics: m}, nil
}

// writeVersion writes the build version information into storage. This makes the
// version information available to the REPL and the HTTP server.
func writeVersion(ctx context.Context, store storage.Store, txn storage.Transaction) error {
	versionPath := storage.MustParsePath("/system/version")

	if err := storage.MakeDir(ctx, store, txn, versionPath); err != nil {
		return err
	}

	if err := store.Write(ctx, txn, storage.AddOp, versionPath, map[string]any{
		"version":         version.Version,
		"build_commit":    version.Vcs,
		"build_timestamp": version.Timestamp,
		"build_hostname":  version.Hostname,
	}); err != nil {
		return errors.Wrap(err, "failed to write version information to storage")
	}

	return nil
}
