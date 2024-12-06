# Topaz configuration

The main configuration for Topaz can be divided in 3 main sections:

1. Common configuration
2. Auth configuration - optional
3. Decision logger configuration - optional
4. Controller configuration - optional

## Topaz configuration environment variables

>
> The topaz service configuration uses the [spf13/viper](https://github.com/spf13/viper) library, which allows all configuration parameters to be passed to the `topazd` service as environment variables with the `TOPAZ_` prefix.
>

If you use topaz CLI to generate your configuration file by default it will add the TOPAZ_DIR environment variable to the path configurations. By default this is empty and considered an NOP addition, but it can easily allow you to specify the desired value to the run/start topaz CLI command with the `-e` flag.

By default if you run/start the topaz container using the topaz CLI the following environment variables will be set in your topaz container:

- `TOPAZ_CERTS_DIR` - default $HOME/.config/topaz/certs - the directory where topaz will load/generate the certs
- `TOPAZ_CFG_DIR` - default $HOME/.config/topaz/cfg - the directory from where topaz will load the configuration file
- `TOPAZ_DB_DIR` - default $HOME/.config/topaz/db - the directory where topaz will store the edge directory DB

- `TOPAZ_CERTS_DIR` - default $HOME/.config/topaz/certs - the directory where topaz will load/generate the certs
- `TOPAZ_CFG_DIR` - default $HOME/.config/topaz/cfg - the directory from where topaz will load the configuration file
- `TOPAZ_DB_DIR` - default $HOME/.config/topaz/db - the directory where topaz will store the edge directory DB

Both run and start topaz CLI commands allow passing optional environment variables to your running container using the -e flag. This will allow you to use any desired environment variable in your configuration file as long as you pass it to the container.

## 1. Common configuration

### a. Version


The configuration version accepted by the version of topaz - current compatible version: 2

### b. Logging


The [logging mechanism](https://github.com/aserto-dev/logger) for topaz is based on [zerolog](https://github.com/rs/zerolog) and has the following available settings:

 - *prod* - boolean - if set to false the entire log output will be written using a zerolog ConsoleWriter, setting this to true will write the errors to the stderr output and other logs to the stdout
 - *log_level* - string - this value is parsed by zerolog to match desired logging level (default: info), available levels: trace, debug, info, warn, error, fatal and panic
 - *grpc_log_level* - string - same as above available values, however this is specific for the logged grpc messages, having the default value set to warn

```
logging:
  prod: true
  log_level: debug
```

### c. API


The API section is defines the configuration for the health, metrics and services that topaz is able to spin up.

#### Health:


The health configuration allows topaz to spin up a health server.


- *listen_address* - string - allows the health service to spin up on the configured port (default: "0.0.0.0:9494")
- *certs* - aserto.TLSConfig - based on [aserto-dev/go-aserto](https://github.com/aserto-dev/go-aserto) package allows setting the paths of your certificate files. By default the certificates are not configured.

#### Metrics:


The metrics configuration allows topaz to spin up a metric server.


- *listen_address* - string - allows the metric service to spin up on the configured port (default: "0.0.0.0:9696")
- *certs* - certs.TLSConfig - based on [aserto-dev/go-aserto](https://github.com/aserto-dev/go-aserto) package allows setting the paths of your certificate files. By default the certificates are not configured.

#### Services APIs:


#### 1. grpc


The grpc section allows configuring the listen address, the connection timeout and the certificates.

- *fqdn* - string - optional value to set the fully qualified domain name for the service endpoint
- *listen_address* - string - allows the topaz GRPC server to spin up on the requested port (default: "0.0.0.0:8282")
- *connection_timeout_seconds* - uint32 - sets the timeout for a [connection establishment](https://pkg.go.dev/google.golang.org/grpc#ConnectionTimeout) (default: 120)
- *certs* - aserto.TLSConfig - based on [aserto-dev/go-aserto](https://github.com/aserto-dev/go-aserto) package allows setting the paths of your certificate files. If you do not have your certificates in the specified paths, topaz will generate self-signed certificates for you. By default topaz will generate the certificates in ` ~/.config/topaz/certs/` path


Example:


```
grpc:
  listen_address: "localhost:8282"
  connection_timeout_seconds: 160
  certs:
    tls_cert_path: "/app/grpc.crt"
    tls_key_path: "/app/grpc.key"
    tls_ca_cert_path: "/app/grpc-ca.crt"
```

#### 2. gateway


The gateway section allows configuring the [grpc gateway](https://github.com/grpc-ecosystem/grpc-gateway) for your topaz authorizer.

- *fqdn* - string - optional value to set the fully qualified domain name for the service endpoint
- *listen_address* - string - allows the topaz Gateway server to spin up on the requested port (default: "0.0.0.0:8383")
- *http* - boolean - when set to true it allows the gateway service to respond to plain http request (default: false)
- *certs* - aserto.TLSConfig - based on [aserto-dev/go-aserto](https://github.com/aserto-dev/go-aserto) package allows setting the paths of your certificate files. If you do not have your certificates in the specified paths, topaz will generate self-signed certificates for you. By default topaz will generate the certificates in ` ~/.config/topaz/certs/` path
- *allowed_origins* - []string - allows setting the paths for the [CORS handler](https://github.com/rs/cors)

Detailed information about the gateway http server timeout configuration is available [here](https://pkg.go.dev/net/http#Server)

- *read_timeout* - time.Duration - default value set to 2 * time.Second (default: 2000000000)
- *read_header_timeout* - time.Duration - default value set to 2 * time.Second (default: 2000000000)
- *write_timeout* - time.Duration - default value set to 2 * time.Second (default: 2000000000)
- *idle_timeout* - time.Duration - default is set to 30 * time.Second (default: 30000000000)

Example:


```
gateway:
  listen_address: "localhost:8383"
  http: false
  fqdn: "https://mygateway.demo.org"
  certs:
    tls_cert_path: "/app/gateway.crt"
    tls_key_path: "/app/gateway.key"
    tls_ca_cert_path: "/app/gateway-ca.crt"
  allowed_origins:
  - https://*.aserto.com
  - https://*aserto-console.netlify.app
  - https://*aserto-playground.netlify.app
```

#### 3. needs


The `needs` section allows adding a dependency between the services started. For example when using an edge directory with a reader service it is recommended to set the authorizer services to be dependent on the reader spin-up to be able to resolve the identity for the authorization calls.

Example:

```
api:
  reader:
    gateway:
      listen_address: localhost:9393
    grpc:
      listen_address: localhost:9595
  authorizer:
    needs: reader
    gateway:
      allowed_origins:
      - https://*.aserto.com
      - https://*aserto-console.netlify.app
      - https://*aserto-playground.netlify.app
    grpc:
      connection_timeout_seconds: 2
```

### c. Directory


The directory section allows setting the configuration for the topaz [local edge directory](https://github.com/aserto-dev/go-edge-ds).


 - *db_path* - string - path to the bbolt database file
 - *request_timeout* - time.Duration - request timeout setting for the bbolt database connection
 - *enable_v2* - boolean - enable directory version 2 services for backward compatibility (default: false)

### d. Remote

Topaz is able to communicate with a directory service based on the [pb-directory proto](https://github.com/aserto-dev/pb-directory) definitions. When the remote address is configured to localhost, topaz is able to spin-up a grpc [edge directory service](https://github.com/aserto-dev/go-edge-ds) based on [bbolt](https://pkg.go.dev/go.etcd.io/bbolt)

The remote address can also be configured to a service that implements the proto definitions (for example, the Postgres-based Aserto directory service). In this case, Topaz will NOT spin-up a local edge directory service, and instead send all directory requests to this remote service.


- *address* - string - address:port of the remote directory service
- *api_key* - string - API key for the directory
- *tenant_id* - string - the directory tenant ID

Example (using the hosted Aserto directory):


```
remote_directory:
  address: "directory.prod.aserto.com:8443"
  api_key: <Your Aserto Directory Access Key>
  tenant_id: <Your Aserto Tenant ID>
```

### e. OPA

The OPA configuration section represent the [runtime configuration](https://github.com/aserto-dev/runtime/blob/main/config.go). The main elements of the runtime configuration are:


- *local_bundles* - runtime.LocalBundlesConfig - allows the runtime to run with a local bundle (local path or local policy OCI image)
- *instance_id* - string - represent the unique identifier of the runtime
- *plugins_error_limit* - int - represents the maximum number of errors that an OPA plugin can trigger before killing the runtime
- *graceful_shutdown_period_seconds* - int - passed to the runtime plugin manager, this represents the allowed time to stop the running plugins gracefully
- *max_plugin_wait_time_seconds* - int -  passed to the runtime plugin manager, this represents the maximum wait time for a plugin to spin-up (default: 30)
- *flags* - runtime.Flags - currently only the boolean *enable_status_plugin* is available. When set to true the runtime status is affected by the OPA status plugin
- *config* - runtime.OPAConfig - the details of the [OPA configuration](https://www.openpolicyagent.org/docs/latest/configuration/)

For more details regarding the OPA configuration see [examples folder](/docs/examples/)

### f. JWT

The JWT section allows setting a custom *acceptable_time_skew_seconds* - int - this specifies the duration in which exp (Expiry) and nbf (Not Before) claims may differ by (default: 5).


## 2. Auth configuration (optional)

By default Topaz authentication configuration is disabled, however if you want to configure API key basic authentication this section of the configuration allows you to set this up.

The options section allows you to specify overrides for specific paths if you want to enable the api key authentication or/and the anonymous authentication for these.

Example:


```
auth:
  keys:
    - dc8a1524dec311eda1ff8bd042196110
  options:
    default:
      enable_api_key: true
      enable_anonymous: false
    overrides:
      paths:
        - /aserto.authorizer.v2.Authorizer/Info
      override:
        enable_anonymous: true
        enable_api_key: false

```

## 3. Decision logger configuration (optional)

By default Topaz does not initiate a decision logger, however if you need to keep a history of the decisions taken by the **IS** API call a rolling logger based on [lumberjack](https://github.com/natefinch/lumberjack) is available to keep decision logs in a file.

Example configuration:


```
decision_logger:
  type: "file"
  config:
    log_file_path: /tmp/mytopaz.log
    max_file_size_mb: 50
    max_file_count: 2
```

To use the decision logger the OPA configuration must contain the [configuration information](https://github.com/aserto-dev/topaz/blob/main/plugins/decision_log/plugin.go#L23) for the decision log plugin.

Example of the decision log plugin configuration:


```
opa:
 instance_id: "-"
 graceful_shutdown_period_seconds: 2
 local_bundles:
   paths: []
   skip_verification: true
 config:
   plugins:
     aserto_decision_log:
       enabled: true
       policy_info:
         registry_service: 'ghcr.io'
         registry_image: 'aserto-policies/policy-peoplefinder-rbac'
         digest: 'b36c9fac3c4f3a20e524ef4eca4ac3170e30281fe003b80a499591043299c898'
```

When deploying topaz as an [Aserto Edge Authorizer](https://docs.aserto.com/docs/edge-authorizers/overview) you can configure the decision logger to send the logs to the upstream Aserto policy instance. For configuration details see: https://docs.aserto.com/docs/edge-authorizers/decision-logs

## 4. Controller configuration (optional)

The controller allows an edge Topaz authorizer to connect to the Aserto Control Plane through a secure mTLS connection.  This way the edge authorizers can sync their running policy with an upstream policy instance and sync their local directory with a remote directory.

For more details on the security and management of edge authorizers see documentation available [here](https://docs.aserto.com/docs/edge-authorizers/security-and-management).

```
controller:
  enabled: true
  server:
    address: relay.prod.aserto.com:8443
    client_cert_path: <path to client certificate>
    client_key_path: <path to client key>
```

