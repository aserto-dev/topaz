module github.com/aserto-dev/topaz

go 1.20

// replace github.com/aserto-dev/go-directory => ../go-directory
// replace github.com/aserto-dev/go-edge-ds => ../go-edge-ds
replace github.com/aserto-dev/service-host => ../service-host

require (
	github.com/alecthomas/kong v0.8.1
	github.com/aserto-dev/aserto-grpc v0.2.2
	github.com/aserto-dev/aserto-management v0.9.2
	github.com/aserto-dev/certs v0.0.3
	github.com/aserto-dev/clui v0.8.3
	github.com/aserto-dev/errors v0.0.6
	github.com/aserto-dev/go-aserto v0.20.3
	github.com/aserto-dev/go-authorizer v0.20.3
	github.com/aserto-dev/go-directory v0.21.8-0.20231015213014-2bac9fa0ff01
	github.com/aserto-dev/go-directory-cli v0.20.14-0.20230913232053-927068a67595
	github.com/aserto-dev/go-edge-ds v0.21.10-0.20231024184532-1fd643274507
	github.com/aserto-dev/go-grpc v0.8.57
	github.com/aserto-dev/header v0.0.5
	github.com/aserto-dev/logger v0.0.4
	github.com/aserto-dev/openapi-authorizer v0.8.81
	github.com/aserto-dev/openapi-directory v0.21.6-0.20230802231916-c64b7b8884fd
	github.com/aserto-dev/runtime v0.56.0
	github.com/aserto-dev/self-decision-logger v0.0.4
	github.com/aserto-dev/service-host v0.0.5
	github.com/cli/browser v1.2.0
	github.com/fatih/color v1.15.0
	github.com/fullstorydev/grpcurl v1.8.8
	github.com/google/uuid v1.3.1
	github.com/google/wire v0.5.0
	github.com/grpc-ecosystem/go-grpc-middleware v1.4.0
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.18.0
	github.com/lestrrat-go/jwx v1.2.26
	github.com/magefile/mage v1.15.0
	github.com/mennanov/fmutils v0.2.1
	github.com/mitchellh/mapstructure v1.5.0
	github.com/onsi/ginkgo v1.16.5
	github.com/onsi/gomega v1.28.0
	github.com/open-policy-agent/opa v0.56.0
	github.com/pkg/errors v0.9.1
	github.com/rs/zerolog v1.31.0
	github.com/spf13/cobra v1.7.0
	github.com/spf13/viper v1.17.0
	github.com/stretchr/testify v1.8.4
	go.opencensus.io v0.24.0
	golang.org/x/sync v0.4.0
	google.golang.org/grpc v1.59.0
	google.golang.org/protobuf v1.31.0
	gopkg.in/natefinch/lumberjack.v2 v2.2.1
	sigs.k8s.io/controller-runtime v0.15.0
)

require (
	buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go v1.31.0-20231017183020-0de7443d03cf.2 // indirect
	github.com/AdaLogics/go-fuzz-headers v0.0.0-20230811130428-ced1acdcaa24 // indirect
	github.com/Masterminds/semver v1.5.0 // indirect
	github.com/OneOfOne/xxhash v1.2.8 // indirect
	github.com/agnivade/levenshtein v1.1.1 // indirect
	github.com/antlr/antlr4/runtime/Go/antlr/v4 v4.0.0-20230512164433-5d1fd1a340c9 // indirect
	github.com/aserto-dev/azm v0.0.6 // indirect
	github.com/aserto-dev/go-decision-logs v0.0.4 // indirect
	github.com/aserto-dev/go-http-metrics v0.10.1-20221024-1 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/bufbuild/protocompile v0.6.0 // indirect
	github.com/bufbuild/protovalidate-go v0.3.4 // indirect
	github.com/bytecodealliance/wasmtime-go/v3 v3.0.2 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/containerd/containerd v1.7.4 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/decred/dcrd/dcrec/secp256k1/v4 v4.2.0 // indirect
	github.com/fsnotify/fsnotify v1.6.0 // indirect
	github.com/go-ini/ini v1.67.0 // indirect
	github.com/go-logr/logr v1.2.4 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/gobwas/glob v0.2.3 // indirect
	github.com/goccy/go-json v0.10.2 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/google/cel-go v0.18.1 // indirect
	github.com/google/go-cmp v0.5.9 // indirect
	github.com/google/subcommands v1.2.0 // indirect
	github.com/gorilla/mux v1.8.0 // indirect
	github.com/grpc-ecosystem/go-grpc-middleware/providers/prometheus v1.0.0-rc.0 // indirect
	github.com/grpc-ecosystem/go-grpc-middleware/v2 v2.0.0-rc.5 // indirect
	github.com/grpc-ecosystem/go-grpc-prometheus v1.2.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway v1.16.0 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/jhump/protoreflect v1.15.2 // indirect
	github.com/klauspost/compress v1.17.0 // indirect
	github.com/kyokomi/emoji v2.2.4+incompatible // indirect
	github.com/lestrrat-go/backoff/v2 v2.0.8 // indirect
	github.com/lestrrat-go/blackmagic v1.0.1 // indirect
	github.com/lestrrat-go/httpcc v1.0.1 // indirect
	github.com/lestrrat-go/iter v1.0.2 // indirect
	github.com/lestrrat-go/option v1.0.1 // indirect
	github.com/magiconair/properties v1.8.7 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mattn/go-runewidth v0.0.15 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.4 // indirect
	github.com/minio/highwayhash v1.0.2 // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/hashstructure/v2 v2.0.2 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/moby/locker v1.0.1 // indirect
	github.com/nats-io/jwt/v2 v2.4.1 // indirect
	github.com/nats-io/nats-server/v2 v2.9.20 // indirect
	github.com/nats-io/nats.go v1.30.2 // indirect
	github.com/nats-io/nkeys v0.4.5 // indirect
	github.com/nats-io/nuid v1.0.1 // indirect
	github.com/nxadm/tail v1.4.8 // indirect
	github.com/olekukonko/tablewriter v0.0.5 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.1.0-rc4 // indirect
	github.com/pelletier/go-toml/v2 v2.1.0 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/prometheus/client_golang v1.16.0 // indirect
	github.com/prometheus/client_model v0.4.0 // indirect
	github.com/prometheus/common v0.44.0 // indirect
	github.com/prometheus/procfs v0.11.0 // indirect
	github.com/rcrowley/go-metrics v0.0.0-20201227073835-cf1acfcdf475 // indirect
	github.com/rivo/uniseg v0.4.4 // indirect
	github.com/rs/cors v1.9.0 // indirect
	github.com/sagikazarmark/locafero v0.3.0 // indirect
	github.com/sagikazarmark/slog-shim v0.1.0 // indirect
	github.com/samber/lo v1.38.1 // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	github.com/sourcegraph/conc v0.3.0 // indirect
	github.com/spf13/afero v1.10.0 // indirect
	github.com/spf13/cast v1.5.1 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/stoewer/go-strcase v1.3.0 // indirect
	github.com/subosito/gotenv v1.6.0 // indirect
	github.com/tchap/go-patricia/v2 v2.3.1 // indirect
	github.com/xeipuuv/gojsonpointer v0.0.0-20190905194746-02993c407bfb // indirect
	github.com/xeipuuv/gojsonreference v0.0.0-20180127040603-bd5ef7bd5415 // indirect
	github.com/yashtewari/glob-intersection v0.2.0 // indirect
	go.etcd.io/bbolt v1.3.7 // indirect
	go.opentelemetry.io/otel v1.16.0 // indirect
	go.opentelemetry.io/otel/metric v1.16.0 // indirect
	go.opentelemetry.io/otel/sdk v1.16.0 // indirect
	go.opentelemetry.io/otel/trace v1.16.0 // indirect
	go.uber.org/atomic v1.9.0 // indirect
	go.uber.org/multierr v1.9.0 // indirect
	golang.org/x/crypto v0.14.0 // indirect
	golang.org/x/exp v0.0.0-20231006140011-7918f672742d // indirect
	golang.org/x/mod v0.13.0 // indirect
	golang.org/x/net v0.17.0 // indirect
	golang.org/x/sys v0.13.0 // indirect
	golang.org/x/term v0.13.0 // indirect
	golang.org/x/text v0.13.0 // indirect
	golang.org/x/time v0.3.0 // indirect
	golang.org/x/tools v0.14.0 // indirect
	google.golang.org/genproto v0.0.0-20231016165738-49dd2c1f3d0b // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20231016165738-49dd2c1f3d0b // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20231016165738-49dd2c1f3d0b // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
	gopkg.in/tomb.v1 v1.0.0-20141024135613-dd632973f1e7 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	oras.land/oras-go/v2 v2.3.0 // indirect
	sigs.k8s.io/yaml v1.3.0 // indirect
)
