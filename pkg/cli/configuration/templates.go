package configuration

type TemplateParams struct {
	Version           int
	LocalPolicyImage  string
	PolicyName        string
	Resource          string
	Authorization     string
	EdgeDirectory     bool
	SeedMetadata      bool
	EnableDirectoryV2 bool
}

const LocalImageTemplate = templatePreamble + `
opa:
  instance_id: "-"
  graceful_shutdown_period_seconds: 2
  # max_plugin_wait_time_seconds: 30 set as default
  local_bundles:
    local_policy_image: {{ .LocalPolicyImage }}
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

const templatePreamble = `# yaml-language-server: $schema=https://topaz.sh/schema/config.json
---
# config schema version
version: {{ .Version }}

logging:
  prod: true
  log_level: info

directory:
  db_path: ${TOPAZ_DIR}/db/directory.db
  request_timeout: 5s # set as default, 5 secs.
  seed_metadata: false # deprecated in v3, with the introduction of manifest files, this always defaults to false.

# remote directory is used to resolve the identity for the authorizer.
remote_directory:
  address: "0.0.0.0:9292" # set as default, it should be the same as the reader as we resolve the identity from the local directory service.
  insecure: true

# default jwt validation configuration
jwt:
  acceptable_time_skew_seconds: 5 # set as default, 5 secs

api:
  health:
    listen_address: "0.0.0.0:9494"
  metrics:
    listen_address: "0.0.0.0:9696"
    zpages: true
  services:
    console:
      gateway:
        listen_address: "0.0.0.0:8080"
        allowed_origins:
        - http://localhost
        - http://localhost:*
        - https://localhost
        - https://localhost:*
        - https://0.0.0.0:*
        - https://*.aserto.com
        - https://*aserto-console.netlify.app
        certs:
          tls_key_path: "${TOPAZ_DIR}/certs/gateway.key"
          tls_cert_path: "${TOPAZ_DIR}/certs/gateway.crt"
          tls_ca_cert_path: "${TOPAZ_DIR}/certs/gateway-ca.crt"
      grpc:
        listen_address: "0.0.0.0:8081"
        certs:
          tls_key_path: "${TOPAZ_DIR}/certs/grpc.key"
          tls_cert_path: "${TOPAZ_DIR}/certs/grpc.crt"
          tls_ca_cert_path: "${TOPAZ_DIR}/certs/grpc-ca.crt"
    reader:
      grpc:
        listen_address: "0.0.0.0:9292"
        # if certs are not specified default certs will be generate with the format reader_grpc.*
        certs:
          tls_key_path: "${TOPAZ_DIR}/certs/grpc.key"
          tls_cert_path: "${TOPAZ_DIR}/certs/grpc.crt"
          tls_ca_cert_path: "${TOPAZ_DIR}/certs/grpc-ca.crt"
      gateway:
        listen_address: "0.0.0.0:9393"
        # if not specified, the allowed_origins includes localhost by default
        allowed_origins:
        - http://localhost
        - http://localhost:*
        - https://localhost
        - https://localhost:*
        - https://0.0.0.0:*
        - https://*.aserto.com
        - https://*aserto-console.netlify.app
        # if no certs are specified, the gateway will have the http flag enabled (http: true)
        certs:
          tls_key_path: "${TOPAZ_DIR}/certs/gateway.key"
          tls_cert_path: "${TOPAZ_DIR}/certs/gateway.crt"
          tls_ca_cert_path: "${TOPAZ_DIR}/certs/gateway-ca.crt"
        http: false
        read_timeout: 2s # default 2 seconds
        read_header_timeout: 2s
        write_timeout: 2s
        idle_timeout: 30s # default 30 seconds      
    writer:
      grpc:
        listen_address: "0.0.0.0:9292"
        certs:
          tls_key_path: "${TOPAZ_DIR}/certs/grpc.key"
          tls_cert_path: "${TOPAZ_DIR}/certs/grpc.crt"
          tls_ca_cert_path: "${TOPAZ_DIR}/certs/grpc-ca.crt"
      gateway:
        listen_address: "0.0.0.0:9393"
        allowed_origins:
        - http://localhost
        - http://localhost:*
        - https://localhost
        - https://localhost:*
        - https://*.aserto.com
        - https://*aserto-console.netlify.app
        certs:
          tls_key_path: "${TOPAZ_DIR}/certs/gateway.key"
          tls_cert_path: "${TOPAZ_DIR}/certs/gateway.crt"
          tls_ca_cert_path: "${TOPAZ_DIR}/certs/gateway-ca.crt"
        http: false
        read_timeout: 2s
        read_header_timeout: 2s
        write_timeout: 2s
        idle_timeout: 30s
    model:
      grpc:
        listen_address: "0.0.0.0:9292"
        certs:
          tls_key_path: "${TOPAZ_DIR}/certs/grpc.key"
          tls_cert_path: "${TOPAZ_DIR}/certs/grpc.crt"
          tls_ca_cert_path: "${TOPAZ_DIR}/certs/grpc-ca.crt"
      gateway:
        listen_address: "0.0.0.0:9393"
        allowed_origins:
        - http://localhost
        - http://localhost:*
        - https://localhost
        - https://localhost:*
        - https://*.aserto.com
        - https://*aserto-console.netlify.app
        certs:
          tls_key_path: "${TOPAZ_DIR}/certs/gateway.key"
          tls_cert_path: "${TOPAZ_DIR}/certs/gateway.crt"
          tls_ca_cert_path: "${TOPAZ_DIR}/certs/gateway-ca.crt"
        http: false
        read_timeout: 2s
        read_header_timeout: 2s
        write_timeout: 2s
        idle_timeout: 30s
    exporter:
      grpc:
        listen_address: "0.0.0.0:9292"
        certs:
          tls_key_path: "${TOPAZ_DIR}/certs/grpc.key"
          tls_cert_path: "${TOPAZ_DIR}/certs/grpc.crt"
          tls_ca_cert_path: "${TOPAZ_DIR}/certs/grpc-ca.crt"
      gateway:
        listen_address: "0.0.0.0:9393"
        allowed_origins:
        - http://localhost
        - http://localhost:*
        - https://localhost
        - https://localhost:*
        - https://*.aserto.com
        - https://*aserto-console.netlify.app
        certs:
          tls_key_path: "${TOPAZ_DIR}/certs/gateway.key"
          tls_cert_path: "${TOPAZ_DIR}/certs/gateway.crt"
          tls_ca_cert_path: "${TOPAZ_DIR}/certs/gateway-ca.crt"
        http: false
        read_timeout: 2s
        read_header_timeout: 2s
        write_timeout: 2s
        idle_timeout: 30s
    importer:
      grpc:
        listen_address: "0.0.0.0:9292"
        certs:
          tls_key_path: "${TOPAZ_DIR}/certs/grpc.key"
          tls_cert_path: "${TOPAZ_DIR}/certs/grpc.crt"
          tls_ca_cert_path: "${TOPAZ_DIR}/certs/grpc-ca.crt"
      gateway:
        listen_address: "0.0.0.0:9393"
        allowed_origins:
        - http://localhost
        - http://localhost:*
        - https://localhost
        - https://localhost:*
        - https://*.aserto.com
        - https://*aserto-console.netlify.app
        certs:
          tls_key_path: "${TOPAZ_DIR}/certs/gateway.key"
          tls_cert_path: "${TOPAZ_DIR}/certs/gateway.crt"
          tls_ca_cert_path: "${TOPAZ_DIR}/certs/gateway-ca.crt"
        http: false
        read_timeout: 2s
        read_header_timeout: 2s
        write_timeout: 2s
        idle_timeout: 30s
  
    authorizer:
      needs:
        - reader
      grpc:
        connection_timeout_seconds: 2
        listen_address: "0.0.0.0:8282"
        certs:
          tls_key_path: "${TOPAZ_DIR}/certs/grpc.key"
          tls_cert_path: "${TOPAZ_DIR}/certs/grpc.crt"
          tls_ca_cert_path: "${TOPAZ_DIR}/certs/grpc-ca.crt"
      gateway:
        listen_address: "0.0.0.0:8383"
        allowed_origins:
        - http://localhost
        - http://localhost:*
        - https://localhost
        - https://localhost:*
        - https://0.0.0.0:*
        - https://*.aserto.com
        - https://*aserto-console.netlify.app
        certs:
          tls_key_path: "${TOPAZ_DIR}/certs/gateway.key"
          tls_cert_path: "${TOPAZ_DIR}/certs/gateway.crt"
          tls_ca_cert_path: "${TOPAZ_DIR}/certs/gateway-ca.crt"
        http: false
        read_timeout: 2s
        read_header_timeout: 2s
        write_timeout: 2s
        idle_timeout: 30s
`
