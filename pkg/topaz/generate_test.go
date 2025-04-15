package topaz_test

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/aserto-dev/go-aserto"
	"github.com/aserto-dev/logger"
	"github.com/aserto-dev/runtime"
	"github.com/aserto-dev/topaz/controller"
	"github.com/aserto-dev/topaz/decisionlog/logger/file"
	"github.com/aserto-dev/topaz/pkg/authentication"
	"github.com/aserto-dev/topaz/pkg/authorizer"
	"github.com/aserto-dev/topaz/pkg/config/handler"
	"github.com/aserto-dev/topaz/pkg/debug"
	"github.com/aserto-dev/topaz/pkg/directory"
	"github.com/aserto-dev/topaz/pkg/health"
	"github.com/aserto-dev/topaz/pkg/metrics"
	"github.com/aserto-dev/topaz/pkg/services"
	"github.com/aserto-dev/topaz/pkg/topaz"

	"github.com/open-policy-agent/opa/v1/download"
	"github.com/open-policy-agent/opa/v1/keys"
	bundleplugin "github.com/open-policy-agent/opa/v1/plugins/bundle"
	"github.com/open-policy-agent/opa/v1/plugins/discovery"
	"github.com/open-policy-agent/opa/v1/plugins/logs"
	"github.com/open-policy-agent/opa/v1/plugins/status"
	"github.com/open-policy-agent/opa/v1/topdown/cache"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

var cfg = &topaz.Config{
	Version: 3,
	Logging: logger.Config{
		Prod:         false,
		LogLevel:     "info",
		GrpcLogLevel: "info",
	},
	Authentication: authentication.Config{
		Enabled: false,
		Plugin:  "local",
		Settings: authentication.LocalSettings{
			Keys: []string{
				"69ba614c64ed4be69485de73d062a00b",
				"##Ve@rySecret123!!",
			},
			Options: authentication.CallOptions{
				Default: authentication.Options{
					EnableAPIKey:    true,
					EnableAnonymous: false,
				},
				Overrides: []authentication.OptionOverrides{
					{
						Paths: []string{
							"/grpc.reflection.v1.ServerReflection/ServerReflectionInfo",
							"/grpc.reflection.v1alpha.ServerReflection/ServerReflectionInfo",
						},
						Override: authentication.Options{
							EnableAPIKey:    false,
							EnableAnonymous: true,
						},
					},
					{
						Paths: []string{
							"/aserto.authorizer.v2.Authorizer/Info",
						},
						Override: authentication.Options{
							EnableAPIKey:    true,
							EnableAnonymous: true,
						},
					},
				},
			},
		},
	},
	Debug: debug.Config{
		Enabled:         false,
		ListenAddress:   "localhost:6060",
		ShutdownTimeout: time.Second * 5,
	},
	Health: health.Config{
		Enabled:       true,
		ListenAddress: "localhost:8484",
		Certificates: aserto.TLSConfig{
			Key:  "${TOPAZ_CERTS_DIR}/grpc.key",
			Cert: "${TOPAZ_CERTS_DIR}/grpc.crt",
			CA:   "${TOPAZ_CERTS_DIR}/grpc-ca.crt",
		},
	},
	Metrics: metrics.Config{
		Enabled:       true,
		ListenAddress: "localhost:8686",
		Certificates: aserto.TLSConfig{
			Key:  "${TOPAZ_CERTS_DIR}/gateway.key",
			Cert: "${TOPAZ_CERTS_DIR}/gateway.crt",
			CA:   "${TOPAZ_CERTS_DIR}/gateway-ca.crt",
		},
	},
	Services: services.Config{
		"topaz": &services.Service{
			DependsOn: []string{},
			GRPC: services.GRPCService{
				ListenAddress: "0.0.0.0:9292",
				FQDN:          "localhost:9292",
				Certs: aserto.TLSConfig{
					Key:  "${TOPAZ_CERTS_DIR}/grpc.key",
					Cert: "${TOPAZ_CERTS_DIR}/grpc.crt",
					CA:   "${TOPAZ_CERTS_DIR}/grpc-ca.crt",
				},
				ConnectionTimeout: time.Second * 7,
				DisableReflection: false,
			},
			Gateway: services.GatewayService{
				ListenAddress: "0.0.0.0:9393",
				FQDN:          "localhost:9393",
				Certs: aserto.TLSConfig{
					Key:  "${TOPAZ_CERTS_DIR}/gateway.key",
					Cert: "${TOPAZ_CERTS_DIR}/gateway.crt",
					CA:   "${TOPAZ_CERTS_DIR}/gateway-ca.crt",
				},
				AllowedOrigins:    services.DefaultAllowedOrigins(false),
				AllowedHeaders:    services.DefaultAllowedHeaders(),
				AllowedMethods:    services.DefaultAllowedMethods(),
				HTTP:              false,
				ReadTimeout:       services.DefaultReadTimeout,
				ReadHeaderTimeout: services.DefaultReadHeaderTimeout,
				WriteTimeout:      services.DefaultWriteTimeout,
				IdleTimeout:       services.DefaultIdleTimeout,
			},
			Includes: []string{
				"model",
				"reader",
				"writer",
				"importer",
				"exporter",
				"authorizer",
				"console",
			},
		},
	},
	Directory: directory.Config{
		ReadTimeout:  time.Second * 3,
		WriteTimeout: time.Second * 6,
		// Store: directory.Store{
		// 	Plugin: directory.BoltDBStorePlugin,
		// 	Settings: directory.BoltDBStoreMap(&directory.BoltDBStore{
		// 		Config: boltdb.Config{
		// 			DBPath:         "${TOPAZ_DB_DIR}/directory.db",
		// 		},
		// 	}),
		// },
		Store: directory.Store{
			PluginConfig: handler.PluginConfig{Plugin: directory.RemoteDirectoryStorePlugin},
			Remote: directory.RemoteDirectoryStore{
				Address:  "directory.prod.aserto.com:8443",
				TenantID: "00000000-1111-2222-3333-444455556666",
				APIKey:   "101520",
				Headers: map[string]string{
					"Aserto-Account-ID": "11111111-9999-8888-7777-666655554444",
				},
			},
		},
	},
	Authorizer: authorizer.Config{
		OPA: authorizer.OPAConfig{
			Config: runtime.Config{
				InstanceID:                    "-",
				GracefulShutdownPeriodSeconds: 2,
				MaxPluginWaitTimeSeconds:      30,
				LocalBundles:                  runtime.LocalBundlesConfig{
					// LocalPolicyImage: "",
					// FileStoreRoot:    "",
					// Paths:            []string{},
					// Ignore:           []string{},
					// Watch:            false,
					// SkipVerification: true,
					// VerificationConfig: &bundle.VerificationConfig{
					// 	PublicKeys: map[string]*bundle.KeyConfig{},
					// 	KeyID:      "",
					// 	Scope:      "",
					// 	Exclude:    []string{},
					// },
				},
				Config: runtime.OPAConfig{
					Services: map[string]interface{}{
						"registry": map[string]interface{}{
							"url": "https://ghcr.io",
						},
						"type":                            "oci",
						"response_header_timeout_seconds": 5,
					},
					Labels:    map[string]string{},
					Discovery: &discovery.Config{},
					Bundles: map[string]*bundleplugin.Source{
						"gdrive": {
							Service:  "registry",
							Resource: "ghcr.io/aserto-policies/policy-rebac:latest",
							Persist:  false,
							Config: download.Config{
								Polling: download.PollingConfig{
									MinDelaySeconds: Ptr[int64](60),
									MaxDelaySeconds: Ptr[int64](120),
								},
							},
						},
					},
					DecisionLogs:                 &logs.Config{},
					Status:                       &status.Config{},
					Plugins:                      map[string]interface{}{},
					Keys:                         map[string]*keys.Config{},
					DefaultDecision:              Ptr[string](""),
					DefaultAuthorizationDecision: Ptr[string](""),
					Caching:                      &cache.Config{},
					PersistenceDirectory:         nil,
				},
			},
		},
		DecisionLogger: authorizer.DecisionLoggerConfig{
			Plugin: authorizer.FileDecisionLoggerPlugin,
			Settings: authorizer.FileDecisionLoggerConfig{
				Config: file.Config{
					LogFilePath:   "/tmp/topaz/decisions.log",
					MaxFileSizeMB: 20,
					MaxFileCount:  3,
				},
			}.Map(),
			// Plugin: authorizer.SelfDecisionLoggerPlugin,
			// Settings: authorizer.SelfDecisionLoggerConfig{
			// 	Config: self.Config{
			// 		StoreDirectory: "${TOPAZ_DIR}/decisions",
			// 		Port:           1234,
			// 		Scribe: scribe.Config{
			// 			Config: aserto.Config{
			// 				Address:        "ems.prod.aserto.com:8443",
			// 				ClientCertPath: "${TOPAZ_DIR}/certs/sidecar.crt",
			// 				ClientKeyPath:  "${TOPAZ_DIR}/certs/sidecar.key",
			// 				Headers: map[string]string{
			// 					"Aserto-Tenant-Id": "55cf8ea9-30b2-4f9a-b0bb-021ca12170f3",
			// 				},
			// 			},
			// 			AckWaitSeconds:     30,
			// 			MaxInflightBatches: 10,
			// 		},
			// 		Shipper: shipper.Config{
			// 			MaxBytes:              0,
			// 			MaxBatchSize:          0,
			// 			PublishTimeoutSeconds: 2,
			// 			MaxInflightBatches:    0,
			// 			AckWaitSeconds:        30,
			// 			DeleteStreamOnDone:    true,
			// 			BackoffSeconds:        []int{10, 9, 8, 7, 6, 5},
			// 		},
			// 	},
			// }.Map(),
		},
		Controller: authorizer.ControllerConfig{
			Config: controller.Config{
				Enabled: true,
				Server: aserto.Config{
					Address:        "relay.prod.aserto.com:8443",
					APIKey:         "0xdeadbeef",
					ClientCertPath: "${TOPAZ_DIR}/certs/grpc.crt",
					ClientKeyPath:  "${TOPAZ_DIR}/certs/grpc.key",
				},
			},
		},
		JWT: authorizer.JWTConfig{
			AcceptableTimeSkew: time.Second * 2,
		},
	},
}

func TestGenerate(t *testing.T) {
	if err := cfg.Generate(os.Stderr); err != nil {
		require.NoError(t, err)
	}

	if false {
		buf, err := json.MarshalIndent(cfg, "", "  ")
		if err != nil {
			require.NoError(t, err)
		}

		var v map[string]interface{}

		dec := yaml.NewDecoder(bytes.NewReader(buf))
		if err := dec.Decode(&v); err != nil {
			require.NoError(t, err)
		}

		enc := yaml.NewEncoder(os.Stdout)
		enc.SetIndent(2)

		if err := enc.Encode(v); err != nil {
			require.NoError(t, err)
		}
	}
}

func Ptr[T any](v T) *T {
	return &v
}
