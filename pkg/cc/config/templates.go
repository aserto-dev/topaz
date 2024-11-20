package config

type templateParams struct {
	Version           int
	PolicyName        string
	Resource          string
	Authorization     string
	LocalPolicy       bool
	EdgeDirectory     bool
	SeedMetadata      bool
	EnableDirectoryV2 bool

	TenantID     string
	DiscoveryURL string
	TenantKey    string
	ControlPlane struct {
		Enabled        bool
		Address        string
		ClientCertPath string
		ClientKeyPath  string
	}
	DecisionLogging bool
	DecisionLogger  struct {
		EMSAddress     string
		StorePath      string
		ClientCertPath string
		ClientKeyPath  string
	}
}

const LocalImageTemplate = templatePreamble + `
opa:
  instance_id: "-"
  graceful_shutdown_period_seconds: 2
  # max_plugin_wait_time_seconds: 30 set as default
  local_bundles:
    local_policy_image: {{ .Resource }}
    watch: true
    skip_verification: true
`

const Template = templatePreamble + `
opa:
  instance_id: "-"
  graceful_shutdown_period_seconds: 2
  # max_plugin_wait_time_seconds: 30 set as default
  local_bundles:
    paths: []
    skip_verification: true
  config:
    services:
      ghcr:
        url: https://ghcr.io
        type: "oci"
        response_header_timeout_seconds: 5
    bundles:
      {{ .PolicyName }}:
        service: ghcr
        resource: "{{ .Resource }}"
        persist: false
        config:
          polling:
            min_delay_seconds: 60
            max_delay_seconds: 120
`

const EdgeTemplate = templatePreamble + `
opa:
  instance_id: {{ .TenantID }}
  graceful_shutdown_period_seconds: 2
  local_bundles:
    paths: []
    skip_verification: true
  config:
    services:
      aserto-discovery:
        url: {{ .DiscoveryURL }}
        credentials:
          bearer:
            token: {{ .TenantKey }}
            scheme: "basic"
        headers:
          Aserto-Tenant-Id: {{ .TenantID }}
    discovery:
      service: aserto-discovery
      resource: {{ .PolicyName }}/{{ .PolicyName }}/opa
{{ if .ControlPlane.Enabled }}
controller:
  enabled: true
  server:
    address: {{ .ControlPlane.Address }}
    client_cert_path: '{{ .ControlPlane.ClientCertPath }}'
    client_key_path: '{{ .ControlPlane.ClientKeyPath }}'
{{ else }}
controller:
  enabled: false
{{ end }}
{{ if .DecisionLogging }}
decision_logger:
  type: self
  config:
    store_directory: {{ .DecisionLogger.StorePath }}
    scribe:
      address: {{ .DecisionLogger.EMSAddress }}
      client_cert_path: '{{ .DecisionLogger.ClientCertPath }}'
      client_key_path: '{{ .DecisionLogger.ClientKeyPath }}'
      ack_wait_seconds: 30
      headers:
        Aserto-Tenant-Id: {{ .TenantID }}
    shipper:
      publish_timeout_seconds: 2
{{ end }}
`

const templatePreamble = `# yaml-language-server: $schema=https://topaz.sh/schema/config.json
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
  db_path: '${TOPAZ_DB_DIR}/{{ .PolicyName }}.db'
  request_timeout: 5s # set as default, 5 secs.

# remote directory is used to resolve the identity for the authorizer.
remote_directory:
  address: "0.0.0.0:9292" # set as default, it should be the same as the reader as we resolve the identity from the local directory service.
  tenant_id: ""
  api_key: ""
  token: ""
  client_cert_path: ""
  client_key_path: ""
  ca_cert_path: ""
  insecure: true
  no_tls: false
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
        - https://*.aserto.com
        - https://*aserto-console.netlify.app
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
        - https://*.aserto.com
        - https://*aserto-console.netlify.app
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
        - https://*.aserto.com
        - https://*aserto-console.netlify.app
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
        - https://*.aserto.com
        - https://*aserto-console.netlify.app
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
        - https://*.aserto.com
        - https://*aserto-console.netlify.app
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
