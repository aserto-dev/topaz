package cmd

type templateParams struct {
	PolicyName    string
	Resource      string
	Authorization string
	EdgeDirectory bool
	SeedMetadata  bool
}

const configTemplate = templatePreamble + `
opa:
  instance_id: "-"
  graceful_shutdown_period_seconds: 2
  local_bundles:
    paths: []
    skip_verification: true
  config:
    services:
      opcr:
        url: https://opcr.io/v2/
        type: "oci"
        response_header_timeout_seconds: 5
        credentials:
          bearer:
            token: "iDog"
            scheme: "basic"
    bundles:
      {{ .PolicyName }}:
        service: opcr
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

directory_service:
  edge:
    db_path: /db/directory.db
    seed_metadata: {{ .SeedMetadata }}
    {{if .EdgeDirectory}}
  remote:
    address: "0.0.0.0:9292"
    insecure: true
    {{end}}
api:
  grpc:
    connection_timeout_seconds: 2
    listen_address: "0.0.0.0:8282"
    certs:
      tls_key_path: "/certs/grpc.key"
      tls_cert_path: "/certs/grpc.crt"
      tls_ca_cert_path: "/certs/grpc-ca.crt"
  gateway:
    listen_address: "0.0.0.0:8383"
    allowed_origins:
    - https://*.aserto.com
    - https://*aserto-console.netlify.app
    certs:
      tls_key_path: "/certs/gateway.key"
      tls_cert_path: "/certs/gateway.crt"
      tls_ca_cert_path: "/certs/gateway-ca.crt"
  health:
    listen_address: "0.0.0.0:8484"
`
