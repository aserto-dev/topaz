package cmd

type templateParams struct {
	LocalPolicyImage string
	PolicyName       string
	Resource         string
	Authorization    string
	EdgeDirectory    bool
	SeedMetadata     bool
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
`

const templatePreamble = `---
logging:
  prod: true
  log_level: info

directory:
  db_path: ${TOPAZ_DIR}/db/directory.db
  seed_metadata: {{ .SeedMetadata }}
  
api:
  authorizer:
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
`
