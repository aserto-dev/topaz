package datasync

import (
	"context"
	"strings"

	dse3 "github.com/aserto-dev/go-directory/aserto/directory/exporter/v3"
	"github.com/aserto-dev/topaz/internal/pkg/eds/pkg/bdb"

	cuckoo "github.com/panmari/cuckoofilter"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type SyncClient interface {
	Sync(ctx context.Context, conn *grpc.ClientConn, opts ...Option) error
}

type Client struct {
	logger *zerolog.Logger
	store  *bdb.BoltDB
}

var _ SyncClient = &Client{}

func New(logger *zerolog.Logger, store *bdb.BoltDB) *Client {
	return &Client{
		logger: logger,
		store:  store,
	}
}

func (c *Client) Sync(ctx context.Context, conn *grpc.ClientConn, opts ...Option) error {
	options := &Options{}
	for _, f := range opts {
		f(options)
	}

	c.logger.Debug().Str("mode", options.Mode.String()).Msg("sync")

	return newSync(c, options).Run(ctx, conn)
}

const (
	syncScheduler  string = "scheduler"
	syncOnDemand   string = "on-demand"
	syncRun        string = "sync-run"
	syncProducer   string = "producer"
	syncSubscriber string = "subscriber"
	syncDifference string = "difference"
	syncManifest   string = "manifest"
	syncStatus     string = "status"
	syncStarted    string = "started"
	syncStage      string = "stage"
	syncFinished   string = "finished"
	channelSize    int    = 10000
)

type Sync struct {
	*Client

	options    *Options
	exportChan chan *dse3.ExportResponse
	errChan    chan error
	tsChan     chan *timestamppb.Timestamp
	filter     *cuckoo.Filter
}

func newSync(c *Client, o *Options) *Sync {
	return &Sync{
		options:    o,
		exportChan: make(chan *dse3.ExportResponse, channelSize),
		errChan:    make(chan error, 1),
		tsChan:     make(chan *timestamppb.Timestamp, 1),
		Client:     c,
	}
}

func (s *Sync) Run(ctx context.Context, conn *grpc.ClientConn) error {
	s.logger.Info().Str("mode", s.options.Mode.String()).Msg(syncRun)

	if Has(s.options.Mode, Manifest) {
		if err := s.syncManifest(ctx, conn); err != nil {
			return err
		}
	}

	if Has(s.options.Mode, Full|Diff|Watermark) {
		if err := s.syncDirectory(ctx, conn); err != nil {
			return err
		}
	}

	return nil
}

type Option func(*Options)

type Options struct {
	Mode Mode
}

type Mode int32

const (
	Unknown   Mode = 0         // unknown mode (default)
	Manifest  Mode = 1 << iota // sync directory manifest
	Full                       // sync all state from source (INSERT/UPDATE)
	Diff                       // sync all state from source (INSERT/UPDATE) and compare against target (DELETE)
	Watermark                  // sync all changes since last watermark timestamp (INSERT/UPDATE)
)

func Set(b, flag Mode) Mode    { return b | flag }
func Clear(b, flag Mode) Mode  { return b &^ flag }
func Toggle(b, flag Mode) Mode { return b ^ flag }
func Has(b, flag Mode) bool    { return b&flag != 0 }

var modes = map[int32]string{
	int32(Manifest):  "MANIFEST",
	int32(Full):      "FULL",
	int32(Diff):      "DIFF",
	int32(Watermark): "WATERMARK",
}

func (m Mode) String() string {
	str := []string{}

	for k, v := range modes {
		if Has(m, Mode(k)) {
			str = append(str, v)
		}
	}

	return strings.Join(str, "|")
}

func (m Mode) RunMode() string {
	mode := Clear(m, Manifest)
	str := []string{}

	for k, v := range modes {
		if Has(mode, Mode(k)) {
			str = append(str, v)
		}
	}

	return strings.Join(str, "|")
}

func StrToMode(s string) Mode {
	for k, v := range modes {
		if v == s {
			return Mode(k)
		}
	}

	return Unknown
}

func WithMode(mode Mode) Option {
	return func(o *Options) {
		o.Mode = Set(o.Mode, mode)
	}
}
