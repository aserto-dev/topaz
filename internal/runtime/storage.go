package runtime

import (
	"context"

	"github.com/open-policy-agent/opa/v1/storage"
	"github.com/open-policy-agent/opa/v1/storage/inmem"
	"github.com/rs/zerolog"
)

// AsertoStore implements the OPA storage interface for the Aserto Runtime.
type AsertoStore struct {
	logger *zerolog.Logger
	cfg    *Config

	backend storage.Store
}

var (
	_ storage.Store   = (*AsertoStore)(nil)
	_ storage.Trigger = (*AsertoStore)(nil)
)

// NewAsertoStore creates a new AsertoStore.
func NewAsertoStore(logger *zerolog.Logger, cfg *Config) *AsertoStore {
	newLogger := logger.With().Str("source", "aserto-storage").Logger()

	return &AsertoStore{
		logger:  &newLogger,
		cfg:     cfg,
		backend: inmem.New(),
	}
}

// NewTransaction is called to create a new transaction in the store.
func (s *AsertoStore) NewTransaction(ctx context.Context, params ...storage.TransactionParams) (storage.Transaction, error) {
	s.logger.Trace().Msg("new-transaction")
	return s.backend.NewTransaction(ctx, params...)
}

// Read is called to fetch a document referred to by path.
func (s *AsertoStore) Read(ctx context.Context, txn storage.Transaction, path storage.Path) (any, error) {
	s.logger.Trace().Str("path", path.String()).Msg("read")
	return s.backend.Read(ctx, txn, path)
}

// Write is called to modify a document referred to by path.
func (s *AsertoStore) Write(ctx context.Context, txn storage.Transaction, op storage.PatchOp, path storage.Path, value any) error {
	s.logger.Trace().Str("path", path.String()).Msg("write")
	return s.backend.Write(ctx, txn, op, path, value)
}

// Commit is called to finish the transaction. If Commit returns an error, the
// transaction must be automatically aborted by the Store implementation.
func (s *AsertoStore) Commit(ctx context.Context, txn storage.Transaction) error {
	s.logger.Trace().Msg("commit")
	return s.backend.Commit(ctx, txn)
}

// Abort is called to cancel the transaction.
func (s *AsertoStore) Abort(ctx context.Context, txn storage.Transaction) {
	s.logger.Trace().Msg("abort")
	s.backend.Abort(ctx, txn)
}

// Register registers a trigger with the storage.
func (s *AsertoStore) Register(ctx context.Context, txn storage.Transaction, config storage.TriggerConfig) (storage.TriggerHandle, error) {
	s.logger.Trace().Msg("register")
	return s.backend.Register(ctx, txn, config)
}

// ListPolicies lists all policies.
func (s *AsertoStore) ListPolicies(ctx context.Context, txn storage.Transaction) ([]string, error) {
	s.logger.Trace().Msg("list-policies")
	return s.backend.ListPolicies(ctx, txn)
}

// GetPolicy gets a policy.
func (s *AsertoStore) GetPolicy(ctx context.Context, txn storage.Transaction, id string) ([]byte, error) {
	s.logger.Trace().Str("id", id).Msg("get-policy")
	return s.backend.GetPolicy(ctx, txn, id)
}

// UpsertPolicy creates a policy, or updates it if it already exists.
func (s *AsertoStore) UpsertPolicy(ctx context.Context, txn storage.Transaction, id string, bs []byte) error {
	s.logger.Trace().Str("id", id).Msg("upsert-policy")
	return s.backend.UpsertPolicy(ctx, txn, id, bs)
}

// DeletePolicy deletes a policy.
func (s *AsertoStore) DeletePolicy(ctx context.Context, txn storage.Transaction, id string) error {
	s.logger.Trace().Str("id", id).Msg("delete-policy")
	return s.backend.DeletePolicy(ctx, txn, id)
}

// Truncate must be called within a transaction.
func (s *AsertoStore) Truncate(ctx context.Context, txn storage.Transaction, params storage.TransactionParams, it storage.Iterator) error {
	s.logger.Trace().Interface("iterator", it).Msg("truncate")
	return s.backend.Truncate(ctx, txn, params, it)
}
