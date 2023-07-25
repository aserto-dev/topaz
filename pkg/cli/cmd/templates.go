package cmd

type templateParams struct {
	LocalPolicyImage  string
	PolicyName        string
	Resource          string
	Authorization     string
	EdgeDirectory     bool
	EdgeAuthorzier    bool
	SeedMetadata      bool
	EdgeCertFile      string
	EdgeKeyFile       string
	RelayAddress      string
	EMSAddress        string
	LogStoreDirectory string
	TenantID          string
	DiscoveryURL      string
	DiscoveryKey      string
}

const localImageTemplate = templatePreamble + `
opa:
  instance_id: "-"
  graceful_shutdown_period_seconds: 2
  local_bundles:
    local_policy_image: {{ .LocalPolicyImage }}
    watch: true
    skip_verification: true
`
const configTemplate = templatePreamble + `
{{ if .EdgeAuthorzier }}
opa:
  instance_id: "{{ .TenantID }}"
  graceful_shutdown_period_seconds: 2
  config:
    services:
      aserto-discovery:
        url: {{.DiscoveryURL}}
        credentials:
          bearer:
            token: "{{ .DiscoveryKey }}"
            scheme: "basic"
        headers:
          Aserto-Tenant-Id: {{ .TenantID }}
    discovery:
      service: aserto-discovery
      resource: {{.PolicyName}}/{{.PolicyName}}/opa
{{ else }}
opa:
  instance_id: "-"
  graceful_shutdown_period_seconds: 2
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
{{ end }}
`

const templatePreamble = `---
logging:
  prod: true
  log_level: info

directory:
  db_path: ${TOPAZ_DIR}/db/directory.db
  seed_metadata: {{ .SeedMetadata }}

# remote directory is used to resolve the identity for the authorizer.
remote_directory:
  address: "0.0.0.0:9292" # same as the reader as we resolve the identity from the local directory service.
  insecure: true
  
api:
  reader:
    grpc:
      listen_address: "0.0.0.0:9292"
      certs:
        tls_key_path: "${TOPAZ_DIR}/certs/grpc.key"
        tls_cert_path: "${TOPAZ_DIR}/certs/grpc.crt"
        tls_ca_cert_path: "${TOPAZ_DIR}/certs/grpc-ca.crt"
    gateway:
      listen_address: "0.0.0.0:9393"
      allowed_origins:
      - https://*.aserto.com
      - https://*aserto-console.netlify.app
      certs:
        tls_key_path: "${TOPAZ_DIR}/certs/gateway.key"
        tls_cert_path: "${TOPAZ_DIR}/certs/gateway.crt"
        tls_ca_cert_path: "${TOPAZ_DIR}/certs/gateway-ca.crt"
    health:
      listen_address: "0.0.0.0:9494"
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
      - https://*.aserto.com
      - https://*aserto-console.netlify.app
      certs:
        tls_key_path: "${TOPAZ_DIR}/certs/gateway.key"
        tls_cert_path: "${TOPAZ_DIR}/certs/gateway.crt"
        tls_ca_cert_path: "${TOPAZ_DIR}/certs/gateway-ca.crt"
    health:
      listen_address: "0.0.0.0:9494"
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
      - https://*.aserto.com
      - https://*aserto-console.netlify.app
      certs:
        tls_key_path: "${TOPAZ_DIR}/certs/gateway.key"
        tls_cert_path: "${TOPAZ_DIR}/certs/gateway.crt"
        tls_ca_cert_path: "${TOPAZ_DIR}/certs/gateway-ca.crt"
    health:
      listen_address: "0.0.0.0:9494"
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
      - https://*.aserto.com
      - https://*aserto-console.netlify.app
      certs:
        tls_key_path: "${TOPAZ_DIR}/certs/gateway.key"
        tls_cert_path: "${TOPAZ_DIR}/certs/gateway.crt"
        tls_ca_cert_path: "${TOPAZ_DIR}/certs/gateway-ca.crt"
    health:
      listen_address: "0.0.0.0:9494"
  
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
      - https://*.aserto.com
      - https://*aserto-console.netlify.app
      certs:
        tls_key_path: "${TOPAZ_DIR}/certs/gateway.key"
        tls_cert_path: "${TOPAZ_DIR}/certs/gateway.crt"
        tls_ca_cert_path: "${TOPAZ_DIR}/certs/gateway-ca.crt"
    health:
      listen_address: "0.0.0.0:8484"

{{if .EdgeAuthorzier }}
decision_logger:
  type: self
  config:
    store_directory: {{.LogStoreDirectory}}
    scribe:
      address: {{.EMSAddress}}
      client_cert_path: {{.EdgeCertFile}}
      client_key_path: {{.EdgeKeyFile}}
      ack_wait_seconds: 30
    shipper:
      publish_timeout_seconds: 2

controller:
  enabled: true
  server:
    address: {{.RelayAddress}}
    client_cert_path: {{.EdgeCertFile}}
    client_key_path: {{.EdgeKeyFile}}
{{ end }}  
`
