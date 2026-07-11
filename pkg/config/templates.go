package config

type templateParams struct {
	Version           int    // must be 2
	ConfigName        string //
	PolicyRegistry    string //
	PolicyName        string //
	Resource          string //
	Authorization     string //
	LocalPolicy       bool   //
	EdgeDirectory     bool   // OBSOLETE
	SeedMetadata      bool   // OBSOLETE
	EnableDirectoryV2 bool   // OBSOLETE
	RegistryService   string //
	RegistryImage     string //
	RegistryTag       string //
}

const LocalImageTemplate string = templatePreamble + opaLocalPolicyImage + topazFileDecisionLoggerPlugin + asertoEdgePlugin

const RemoteImageTemplate string = templatePreamble + opaRemotePolicyImage + topazFileDecisionLoggerPlugin + asertoEdgePlugin

const templatePreamble string = `# yaml-language-server: $schema=https://topaz.sh/schema/config.json
---
# config schema version
version: {{ .Version }}

# logger settings.
logging:
  prod: true
  log_level: info
  grpc_log_level: info

# edge directory configuration.
directory:
  db_path: '${TOPAZ_DB_DIR}/{{ .ConfigName }}.db'
  request_timeout: 5s # set as default, 5 secs.

# remote directory is used to resolve the identity for the authorizer.
remote_directory:
  address: "0.0.0.0:9292" # set as default, it should be the same as the reader as we resolve the identity from the local directory service.
  insecure: true
  no_tls: false
  no_proxy: false
  api_key: ""
  token: ""
  client_cert_path: ""
  client_key_path: ""
  ca_cert_path: ""
  headers:

# default jwt validation configuration
jwt:
  acceptable_time_skew_seconds: 5 # set as default, 5 secs

# authentication configuration
auth:
  keys:
    # - "<API key>"
    # - "<Password>"
  options:
    default:
      enable_api_key: false
      enable_anonymous: true
    overrides:
      paths:
        - /aserto.authorizer.v2.Authorizer/Info
        - /grpc.reflection.v1.ServerReflection/ServerReflectionInfo
        - /grpc.reflection.v1alpha.ServerReflection/ServerReflectionInfo
      override:
        enable_api_key: false
        enable_anonymous: true

api:
  health:
    listen_address: "0.0.0.0:9494"

  metrics:
    listen_address: "0.0.0.0:9696"
    zpages: true

  services:
    console:
      grpc:
        listen_address: "0.0.0.0:8081"
        fqdn: ""
        certs:
          tls_key_path: '${TOPAZ_CERTS_DIR}/grpc.key'
          tls_cert_path: '${TOPAZ_CERTS_DIR}/grpc.crt'
          tls_ca_cert_path: '${TOPAZ_CERTS_DIR}/grpc-ca.crt'
      gateway:
        listen_address: "0.0.0.0:8080"
        fqdn: ""
        allowed_headers:
        - "Authorization"
        - "Content-Type"
        - "If-Match"
        - "If-None-Match"
        - "Depth"
        allowed_methods:
        - "GET"
        - "POST"
        - "HEAD"
        - "DELETE"
        - "PUT"
        - "PATCH"
        - "PROFIND"
        - "MKCOL"
        - "COPY"
        - "MOVE"
        allowed_origins:
        - http://localhost
        - http://localhost:*
        - https://localhost
        - https://localhost:*
        - https://0.0.0.0:*
        certs:
          tls_key_path: '${TOPAZ_CERTS_DIR}/gateway.key'
          tls_cert_path: '${TOPAZ_CERTS_DIR}/gateway.crt'
          tls_ca_cert_path: '${TOPAZ_CERTS_DIR}/gateway-ca.crt'
        http: false
        read_timeout: 2s
        read_header_timeout: 2s
        write_timeout: 2s
        idle_timeout: 30s

    model:
      grpc:
        listen_address: "0.0.0.0:9292"
        fqdn: ""
        certs:
          tls_key_path: '${TOPAZ_CERTS_DIR}/grpc.key'
          tls_cert_path: '${TOPAZ_CERTS_DIR}/grpc.crt'
          tls_ca_cert_path: '${TOPAZ_CERTS_DIR}/grpc-ca.crt'
      gateway:
        listen_address: "0.0.0.0:9393"
        fqdn: ""
        allowed_headers:
        - "Authorization"
        - "Content-Type"
        - "If-Match"
        - "If-None-Match"
        - "Depth"
        allowed_methods:
        - "GET"
        - "POST"
        - "HEAD"
        - "DELETE"
        - "PUT"
        - "PATCH"
        - "PROFIND"
        - "MKCOL"
        - "COPY"
        - "MOVE"
        allowed_origins:
        - http://localhost
        - http://localhost:*
        - https://localhost
        - https://localhost:*
        certs:
          tls_key_path: '${TOPAZ_CERTS_DIR}/gateway.key'
          tls_cert_path: '${TOPAZ_CERTS_DIR}/gateway.crt'
          tls_ca_cert_path: '${TOPAZ_CERTS_DIR}/gateway-ca.crt'
        http: false
        read_timeout: 2s
        read_header_timeout: 2s
        write_timeout: 2s
        idle_timeout: 30s

    reader:
      needs:
        - model
      grpc:
        listen_address: "0.0.0.0:9292"
        fqdn: ""
        certs:
          tls_key_path: '${TOPAZ_CERTS_DIR}/grpc.key'
          tls_cert_path: '${TOPAZ_CERTS_DIR}/grpc.crt'
          tls_ca_cert_path: '${TOPAZ_CERTS_DIR}/grpc-ca.crt'
      gateway:
        listen_address: "0.0.0.0:9393"
        fqdn: ""
        allowed_headers:
        - "Authorization"
        - "Content-Type"
        - "If-Match"
        - "If-None-Match"
        - "Depth"
        allowed_methods:
        - "GET"
        - "POST"
        - "HEAD"
        - "DELETE"
        - "PUT"
        - "PATCH"
        - "PROFIND"
        - "MKCOL"
        - "COPY"
        - "MOVE"
        allowed_origins:
        - http://localhost
        - http://localhost:*
        - https://localhost
        - https://localhost:*
        - https://0.0.0.0:*
        certs:
          tls_key_path: '${TOPAZ_CERTS_DIR}/gateway.key'
          tls_cert_path: '${TOPAZ_CERTS_DIR}/gateway.crt'
          tls_ca_cert_path: '${TOPAZ_CERTS_DIR}/gateway-ca.crt'
        http: false
        read_timeout: 2s # default 2 seconds
        read_header_timeout: 2s
        write_timeout: 2s
        idle_timeout: 30s # default 30 seconds

    writer:
      needs:
        - model
      grpc:
        listen_address: "0.0.0.0:9292"
        fqdn: ""
        certs:
          tls_key_path: '${TOPAZ_CERTS_DIR}/grpc.key'
          tls_cert_path: '${TOPAZ_CERTS_DIR}/grpc.crt'
          tls_ca_cert_path: '${TOPAZ_CERTS_DIR}/grpc-ca.crt'
      gateway:
        listen_address: "0.0.0.0:9393"
        fqdn: ""
        allowed_headers:
        - "Authorization"
        - "Content-Type"
        - "If-Match"
        - "If-None-Match"
        - "Depth"
        allowed_methods:
        - "GET"
        - "POST"
        - "HEAD"
        - "DELETE"
        - "PUT"
        - "PATCH"
        - "PROFIND"
        - "MKCOL"
        - "COPY"
        - "MOVE"
        allowed_origins:
        - http://localhost
        - http://localhost:*
        - https://localhost
        - https://localhost:*
        certs:
          tls_key_path: '${TOPAZ_CERTS_DIR}/gateway.key'
          tls_cert_path: '${TOPAZ_CERTS_DIR}/gateway.crt'
          tls_ca_cert_path: '${TOPAZ_CERTS_DIR}/gateway-ca.crt'
        http: false
        read_timeout: 2s
        read_header_timeout: 2s
        write_timeout: 2s
        idle_timeout: 30s

    exporter:
      grpc:
        listen_address: "0.0.0.0:9292"
        fqdn: ""
        certs:
          tls_key_path: '${TOPAZ_CERTS_DIR}/grpc.key'
          tls_cert_path: '${TOPAZ_CERTS_DIR}/grpc.crt'
          tls_ca_cert_path: '${TOPAZ_CERTS_DIR}/grpc-ca.crt'

    importer:
      needs:
        - model
      grpc:
        listen_address: "0.0.0.0:9292"
        fqdn: ""
        certs:
          tls_key_path: '${TOPAZ_CERTS_DIR}/grpc.key'
          tls_cert_path: '${TOPAZ_CERTS_DIR}/grpc.crt'
          tls_ca_cert_path: '${TOPAZ_CERTS_DIR}/grpc-ca.crt'

    authorizer:
      needs:
        - reader
      grpc:
        connection_timeout_seconds: 2
        listen_address: "0.0.0.0:8282"
        fqdn: ""
        certs:
          tls_key_path: '${TOPAZ_CERTS_DIR}/grpc.key'
          tls_cert_path: '${TOPAZ_CERTS_DIR}/grpc.crt'
          tls_ca_cert_path: '${TOPAZ_CERTS_DIR}/grpc-ca.crt'
      gateway:
        listen_address: "0.0.0.0:8383"
        fqdn: ""
        allowed_headers:
        - "Authorization"
        - "Content-Type"
        - "If-Match"
        - "If-None-Match"
        - "Depth"
        allowed_methods:
        - "GET"
        - "POST"
        - "HEAD"
        - "DELETE"
        - "PUT"
        - "PATCH"
        - "PROFIND"
        - "MKCOL"
        - "COPY"
        - "MOVE"
        allowed_origins:
        - http://localhost
        - http://localhost:*
        - https://localhost
        - https://localhost:*
        - https://0.0.0.0:*
        certs:
          tls_key_path: '${TOPAZ_CERTS_DIR}/gateway.key'
          tls_cert_path: '${TOPAZ_CERTS_DIR}/gateway.crt'
          tls_ca_cert_path: '${TOPAZ_CERTS_DIR}/gateway-ca.crt'
        http: false
        read_timeout: 2s
        read_header_timeout: 2s
        write_timeout: 2s
        idle_timeout: 30s
`

const opaLocalPolicyImage string = `
opa:
  instance_id: "-"
  graceful_shutdown_period_seconds: 2
  # max_plugin_wait_time_seconds: 30 set as default
  local_bundles:
    local_policy_image: {{ .Resource }}
    watch: true
    skip_verification: true
  config:
    decision_logs:
      console: false
    plugins:
`

const opaRemotePolicyImage string = `
opa:
  instance_id: "-"
  graceful_shutdown_period_seconds: 2
  # max_plugin_wait_time_seconds: 30 set as default
  local_bundles:
    paths: []
    skip_verification: true
  config:
    services:
      policy-registry:
        url: "{{ .PolicyRegistry }}"
        type: "oci"
        response_header_timeout_seconds: 15
    bundles:
      {{ .PolicyName }}:
        service: policy-registry
        resource: "{{ .Resource }}"
        persist: false
        config:
          polling:
            min_delay_seconds: 60
            max_delay_seconds: 120
    decision_logs:
      console: false
    plugins:
`

const topazFileDecisionLoggerPlugin string = `
      # topaz file decision logger plugin configuration
      topaz_file_decision_logger:
        enabled: false
        logger:
          filename: '${TOPAZ_DECISIONS_DIR}/{{ .ConfigName }}.json'
          max_size: 100
          max_age: 0
          max_backups: 0
          local_time: false
          compress: false
        policy_info:
          policy_name: '{{ .PolicyName }}'
          registry_service: '{{ .RegistryService }}'
          registry_image: '{{ .RegistryImage }}'
          registry_tag: '{{ .RegistryTag }}'
          digest: ''
`

const asertoEdgePlugin string = `
      # aserto edge directory sync plugin configuration
      aserto_edge:
        enabled: false 
        addr: ""                    # gRPC directory service address.
        apikey: ""                  # directory API key.
        timeout: 5                  # gRPC connection timeout in seconds.
        sync_interval: 1            # sync run interval in minutes.
        insecure: true              # when using TLS connections, skip verification of the server certificate. 
        page_size: 0                # deprecated: no longer used.
        client_cert_path: ""        # when using mTLS connections, ClientCertPath is the path of the client's certificate file.
        client_key_path: ""         # when using mTLS connections, ClientKeyPath is the path of the client's private key file.
        no_tls: false               # disable TLS and use a plaintext connection.
        no_proxy: false             # bypasses any configured HTTP proxy.
        headers:                    # additional headers to include in requests to the service.
`
