package config_test

import (
	"bytes"
	"testing"
	"time"

	"github.com/aserto-dev/go-aserto"
	"github.com/aserto-dev/logger"
	"github.com/aserto-dev/runtime"
	cfgutil "github.com/aserto-dev/topaz/pkg/config"
	"github.com/aserto-dev/topaz/pkg/config/v3"
	"github.com/aserto-dev/topaz/topazd/authentication"
	"github.com/aserto-dev/topaz/topazd/authorizer"
	"github.com/aserto-dev/topaz/topazd/authorizer/controller"
	"github.com/aserto-dev/topaz/topazd/authorizer/decisionlog/logger/file"
	"github.com/aserto-dev/topaz/topazd/debug"
	"github.com/aserto-dev/topaz/topazd/directory"
	"github.com/aserto-dev/topaz/topazd/health"
	"github.com/aserto-dev/topaz/topazd/metrics"
	"github.com/aserto-dev/topaz/topazd/servers"

	"github.com/open-policy-agent/opa/v1/download"
	bundleplugin "github.com/open-policy-agent/opa/v1/plugins/bundle"
	"github.com/open-policy-agent/opa/v1/plugins/discovery"
	"github.com/open-policy-agent/opa/v1/plugins/logs"
	"github.com/open-policy-agent/opa/v1/plugins/status"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var cfg = &config.Config{
	Version: 3,
	Logging: logger.Config{
		Prod:         false,
		LogLevel:     "info",
		GrpcLogLevel: "info",
	},
	Authentication: authentication.Config{
		Optional: cfgutil.Optional{
			Enabled: false,
		},
		Provider: "local",
		Local: authentication.LocalConfig{
			Keys: []string{
				"69ba614c64ed4be69485de73d062a00b",
				"##Ve@rySecret123!!",
			},
			Options: authentication.CallOptions{
				Default: authentication.Options{
					AllowAnonymous: false,
				},
				Overrides: []authentication.OptionOverrides{
					{
						Paths: []string{
							"/grpc.reflection.v1.ServerReflection/ServerReflectionInfo",
							"/grpc.reflection.v1alpha.ServerReflection/ServerReflectionInfo",
						},
						Override: authentication.Options{
							AllowAnonymous: true,
						},
					},
					{
						Paths: []string{
							"/aserto.authorizer.v2.Authorizer/Info",
						},
						Override: authentication.Options{
							AllowAnonymous: true,
						},
					},
				},
			},
		},
	},
	Debug: debug.Config{
		Optional: cfgutil.Optional{
			Enabled: false,
		},
		HTTPServer: servers.HTTPServer{
			Listener: cfgutil.Listener{
				ListenAddress: "localhost:6060",
			},
		},
	},
	Health: health.Config{
		Optional: cfgutil.Optional{
			Enabled: true,
		},
		GRPCServer: servers.GRPCServer{
			Listener: cfgutil.Listener{
				ListenAddress: "localhost:8484",
				Certs: aserto.TLSConfig{
					Key:  "${TOPAZ_CERTS_DIR}/grpc.key",
					Cert: "${TOPAZ_CERTS_DIR}/grpc.crt",
					CA:   "${TOPAZ_CERTS_DIR}/grpc-ca.crt",
				},
			},
		},
	},
	Metrics: metrics.Config{
		Optional: cfgutil.Optional{
			Enabled: true,
		},
		Listener: cfgutil.Listener{
			ListenAddress: "localhost:8686",
			Certs: aserto.TLSConfig{
				Key:  "${TOPAZ_CERTS_DIR}/gateway.key",
				Cert: "${TOPAZ_CERTS_DIR}/gateway.crt",
				CA:   "${TOPAZ_CERTS_DIR}/gateway-ca.crt",
			},
		},
	},
	Servers: servers.Config{
		"topaz": &servers.Server{
			GRPC: servers.GRPCServer{
				Listener: cfgutil.Listener{
					ListenAddress: "0.0.0.0:9292",
					Certs: aserto.TLSConfig{
						Key:  "${TOPAZ_CERTS_DIR}/grpc.key",
						Cert: "${TOPAZ_CERTS_DIR}/grpc.crt",
						CA:   "${TOPAZ_CERTS_DIR}/grpc-ca.crt",
					},
				},
				ConnectionTimeout: time.Second * 7,
				NoReflection:      false,
			},
			HTTP: servers.HTTPServer{
				Listener: cfgutil.Listener{
					ListenAddress: "0.0.0.0:9393",
					Certs: aserto.TLSConfig{
						Key:  "${TOPAZ_CERTS_DIR}/gateway.key",
						Cert: "${TOPAZ_CERTS_DIR}/gateway.crt",
						CA:   "${TOPAZ_CERTS_DIR}/gateway-ca.crt",
					},
				},
				HostedDomain:      "localhost:9393",
				AllowedOrigins:    servers.DefaultAllowedOrigins(false),
				AllowedHeaders:    servers.DefaultAllowedHeaders(),
				AllowedMethods:    servers.DefaultAllowedMethods(),
				ReadTimeout:       servers.DefaultReadTimeout,
				ReadHeaderTimeout: servers.DefaultReadHeaderTimeout,
				WriteTimeout:      servers.DefaultWriteTimeout,
				IdleTimeout:       servers.DefaultIdleTimeout,
			},
			Services: []servers.ServiceName{
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
			Provider: directory.RemoteDirectoryStorePlugin,
			Remote: directory.RemoteDirectoryStore{
				Config: aserto.Config{
					Address:  "directory.prod.aserto.com:8443",
					TenantID: "00000000-1111-2222-3333-444455556666",
					APIKey:   "101520",
					Headers: map[string]string{
						"aserto-account-id": "11111111-9999-8888-7777-666655554444",
					},
				},
			},
		},
	},
	Authorizer: authorizer.Config{
		OPA: authorizer.OPAConfig(runtime.Config{
			InstanceID:                    "-",
			GracefulShutdownPeriodSeconds: 2,
			MaxPluginWaitTimeSeconds:      30,
			LocalBundles: runtime.LocalBundlesConfig{
				LocalPolicyImage: "",
				FileStoreRoot:    "",
				Watch:            false,
				SkipVerification: true,
				Paths:            []string{},
				// Ignore:           []string{},
				// VerificationConfig: &bundle.VerificationConfig{
				// 	PublicKeys: map[string]*bundle.KeyConfig{},
				// 	KeyID:      "",
				// 	Scope:      "",
				// 	Exclude:    []string{},
				// },
			},
			Config: runtime.OPAConfig{
				Services: map[string]any{
					"registry": map[string]any{
						"url": "https://ghcr.io",
					},
					"type":                            "oci",
					"response_header_timeout_seconds": 5,
				},
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
				DefaultDecision:              Ptr(""),
				DefaultAuthorizationDecision: Ptr(""),
				PersistenceDirectory:         nil,
				// Labels:                       map[string]string{},
				// Plugins:                      map[string]any{},
				// Keys:                         map[string]*keys.Config{},
				// Caching:                      &cache.Config{},
			},
		}),
		DecisionLogger: authorizer.DecisionLoggerConfig{
			Provider: authorizer.FileDecisionLoggerPlugin,
			File: authorizer.FileDecisionLoggerConfig(file.Config{
				LogFilePath:   "/tmp/topaz/decisions.log",
				MaxFileSizeMB: 20,
				MaxFileCount:  3,
			}),
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
		Controller: authorizer.ControllerConfig(controller.Config{
			Optional: cfgutil.Optional{
				Enabled: true,
			},
			Server: aserto.Config{
				Address:        "relay.prod.aserto.com:8443",
				APIKey:         "0xdeadbeef",
				ClientCertPath: "${TOPAZ_DIR}/certs/grpc.crt",
				ClientKeyPath:  "${TOPAZ_DIR}/certs/grpc.key",
			},
		}),
		JWT: authorizer.JWTConfig{
			AcceptableTimeSkew: time.Second * 2,
		},
	},
}

func TestSerialize(t *testing.T) {
	t.Skip("Needs improved handling of default vaules in config.")

	serialized := &bytes.Buffer{}
	if err := cfg.Serialize(serialized); err != nil {
		require.NoError(t, err)
	}

	deserialized, err := config.NewConfig(serialized, config.WithNoEnvSubstitution)
	require.NoError(t, err)

	assert.Equal(t, cfg, deserialized)
}

func Ptr[T any](v T) *T {
	return &v
}
