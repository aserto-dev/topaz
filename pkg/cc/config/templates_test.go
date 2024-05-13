package config_test

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"text/template"

	"github.com/Masterminds/sprig"
	"github.com/pkg/errors"
)

const recursionMaxNums = 1000

// 'include' needs to be defined in the scope of a 'tpl' template as
// well as regular file-loaded templates.
func includeFun(t *template.Template, includedNames map[string]int) func(string, interface{}) (string, error) {
	return func(name string, data interface{}) (string, error) {
		var buf strings.Builder
		if v, ok := includedNames[name]; ok {
			if v > recursionMaxNums {
				return "", errors.Wrapf(fmt.Errorf("unable to execute template"), "rendering template has a nested reference name: %s", name)
			}
			includedNames[name]++
		} else {
			includedNames[name] = 1
		}
		err := t.ExecuteTemplate(&buf, name, data)
		includedNames[name]--
		return buf.String(), err
	}
}

func TestTemplate(t *testing.T) {
	tmpl := template.New("CONFIG_TEMPLATE")

	var funcMap template.FuncMap = map[string]interface{}{}

	// copied from: https://github.com/helm/helm/blob/8648ccf5d35d682dcd5f7a9c2082f0aaf071e817/pkg/engine/engine.go#L147-L154
	// funcMap["include"] = func(name string, data interface{}) (string, error) {
	// 	buf := bytes.NewBuffer(nil)
	// 	if err := tmpl.ExecuteTemplate(buf, name, data); err != nil {
	// 		log.Fatal(err)
	// 		// return "", err
	// 	}
	// 	return buf.String(), nil
	// }
	includedNames := make(map[string]int)
	funcMap["include"] = includeFun(tmpl, includedNames)

	tmpl = tmpl.Funcs(sprig.TxtFuncMap()).Funcs(funcMap)

	tmpl = template.Must(tmpl.Parse(configTemplate))

	tmpl = template.Must(tmpl.Parse(configHeaderTemplate))
	tmpl = template.Must(tmpl.Parse(loggingTemplate))
	tmpl = template.Must(tmpl.Parse(debugTemplate))
	tmpl = template.Must(tmpl.Parse(directoryTemplate))
	tmpl = template.Must(tmpl.Parse(remoteDirectoryTemplate))
	tmpl = template.Must(tmpl.Parse(jwtTemplate))
	tmpl = template.Must(tmpl.Parse(apiTemplate))
	tmpl = template.Must(tmpl.Parse(healthTemplate))
	tmpl = template.Must(tmpl.Parse(metricsTemplate))
	tmpl = template.Must(tmpl.Parse(servicesTemplate))
	tmpl = template.Must(tmpl.Parse(serviceTemplate))
	tmpl = template.Must(tmpl.Parse(needsTemplate))
	tmpl = template.Must(tmpl.Parse(grpcTemplate))
	tmpl = template.Must(tmpl.Parse(grpcCertsTemplate))
	tmpl = template.Must(tmpl.Parse(gatewayTemplate))
	tmpl = template.Must(tmpl.Parse(gatewayCertsTemplate))
	tmpl = template.Must(tmpl.Parse(corsAllowedHeaders))
	tmpl = template.Must(tmpl.Parse(corsAllowedMethods))
	tmpl = template.Must(tmpl.Parse(corsAllowedOrigins))
	tmpl = template.Must(tmpl.Parse(controllerTemplate))
	tmpl = template.Must(tmpl.Parse(decisionLoggerTemplate))
	tmpl = template.Must(tmpl.Parse(opaTemplate))

	t.Logf(tmpl.DefinedTemplates())

	w, _ := os.Create("./templates_test.yaml")
	t.Cleanup(func() { w.Close() })

	tmpl.Execute(w, map[string]interface{}{
		"Version":      2,
		"services":     []string{"console", "model", "reader", "writer"},
		"name":         "authorizer",
		"grpc_host":    "0.0.0.0",
		"grpc_port":    "8282",
		"gateway_host": "localhost",
		"gateway_port": "8383",
		"needs":        []string{"model", "reader"},
	})
}

const configTemplate = `{{define "CONFIG_TEMPLATE"}}
{{- include "CONFIG_HEADER" . | indent 0 }}
{{- include "LOGGING" . | indent 0 }}
{{- include "DEBUG" . | indent 0 }}
{{- include "DIRECTORY" . | indent 0 }}
{{- include "REMOTE_DIRECTORY" . | indent 0 }}
{{- include "JWT" . | indent 0 }}
{{- include "API" . | indent 0 }}
{{- include "HEALTH" . | indent 2 }}
{{- include "METRICS" . | indent 2 }}
{{- include "SERVICES" . | indent 2 }}
{{- include "CONTROLLER" . | indent 0 }}
{{- include "DECISION_LOGGER" . | indent 0 }}
{{- include "OPA" . | indent 0 }}
{{end}}`

const configHeaderTemplate = `{{define "CONFIG_HEADER"}}
# yaml-language-server: $schema=https://topaz.sh/schema/config.json
---
# config schema version
version: {{ .Version }}
{{end}}`

const loggingTemplate = `{{define "LOGGING"}}
logging:
  prod: true
  log_level: info
  grpc_log_level: info
{{end}}`

const debugTemplate = `{{define "DEBUG"}}
debug:
  enabled: false
  listen_address: ""
  shutdown_timeout: 0
{{end}}`

const directoryTemplate = `{{define "DIRECTORY"}}
directory:
  db_path: '${TOPAZ_DB_DIR}/{{ .PolicyName }}.db'
  request_timeout: 5s # set as default, 5 secs.
{{end}}`

const remoteDirectoryTemplate = `{{define "REMOTE_DIRECTORY"}}
remote_directory:
  address: "0.0.0.0:9292" # set as default, it should be the same as the reader as we resolve the identity from the local directory service.
  tenant_id: ""
  api_key: ""
  token: ""
  client_cert_path: ""
  client_key_path: ""
  ca_cert_path: ""
  timeout_in_seconds: 5
  insecure: false
  headers:
{{end}}`

const jwtTemplate = `{{define "JWT"}}
# default jwt validation configuration
jwt:
  acceptable_time_skew_seconds: 5 # set as default, 5 secs
{{end}}`

const apiTemplate = `{{define "API"}}
api:
{{end}}`

const healthTemplate = `{{define "HEALTH"}}
health:
  listen_address: "0.0.0.0:9494"
  certs:
{{end}}`

const metricsTemplate = `{{define "METRICS"}}
metrics:
  listen_address: "0.0.0.0:9696"
  certs:
  zpages: true
{{end}}`

const servicesTemplate = `{{define "SERVICES"}}
services:
{{ range $service := .services }}
  {{- $service | indent 2 }}:
  {{- include "SERVICE_TEMPLATE" . | indent 4 }}
{{ end }}{{end}}`

const serviceTemplate = `{{define "SERVICE_TEMPLATE"}}
needs:
{{- include "NEEDS" . | indent 2 }}
grpc:
{{- include "GRPC" . | indent 2 }}
gateway:
{{- include "GATEWAY" . | indent 2 }}
{{end}}`

const needsTemplate = `{{define "NEEDS"}}
{{end}}`

const grpcTemplate = `{{define "GRPC"}}
listen_address: "0.0.0.0:9292"
fqdn: ""
connection_timeout_seconds: 2
certs:
{{- include "GRPC_CERTS" . | indent 2 }}
{{end}}`

const grpcCertsTemplate = `{{define "GRPC_CERTS"}}
tls_key_path: '${TOPAZ_CERTS_DIR}/grpc.key'
tls_cert_path: '${TOPAZ_CERTS_DIR}/grpc.crt'
tls_ca_cert_path: '${TOPAZ_CERTS_DIR}/grpc-ca.crt'
{{end}}`

const gatewayTemplate = `{{define "GATEWAY"}}
listen_address: "0.0.0.0:9393"
fqdn: ""
certs:
{{- include "GATEWAY_CERTS" . | indent 2 }}
{{- include "CORS_ALLOWED_ORIGINS" . | indent 0}}
{{- include "CORS_ALLOWED_HEADERS" . | indent 0}}
{{- include "CORS_ALLOWED_METHODS" . | indent 0}}
http: false
read_timeout: 2s
read_header_timeout: 2s
write_timeout: 2s
idle_timeout: 30s
{{end}}`

const gatewayCertsTemplate = `{{define "GATEWAY_CERTS" }}
tls_key_path: '${TOPAZ_CERTS_DIR}/gateway.key'
tls_cert_path: '${TOPAZ_CERTS_DIR}/gateway.crt'
tls_ca_cert_path: '${TOPAZ_CERTS_DIR}/gateway-ca.crt'
{{end}}`

// const serviceTemplate = `{{define "SERVICE_TEMPLATE"}}
//   needs:
//   {{ range $svc_name := .needs -}}
//   - {{ $svc_name }}
//   {{ end -}}
//   grpc:
//     connection_timeout_seconds: 2
//     listen_address: "{{.grpc_host}}:{{.grpc_port}}"
//     certs:
//       tls_key_path: '${TOPAZ_CERTS_DIR}/grpc.key'
//       tls_cert_path: '${TOPAZ_CERTS_DIR}/grpc.crt'
//       tls_ca_cert_path: '${TOPAZ_CERTS_DIR}/grpc-ca.crt'
//   gateway:
//     listen_address: "{{.gateway_host}}:{{.gateway_port}}"
//     {{- template "CORS_ALLOWED_HEADERS" -}}
//     {{- template "CORS_ALLOWED_METHODS" -}}
//     {{- template "CORS_ALLOWED_ORIGINS" }}
//   certs:
//     tls_key_path: '${TOPAZ_CERTS_DIR}/gateway.key'
//     tls_cert_path: '${TOPAZ_CERTS_DIR}/gateway.crt'
//     tls_ca_cert_path: '${TOPAZ_CERTS_DIR}/gateway-ca.crt'
//   http: false
//   read_timeout: 2s
//   read_header_timeout: 2s
//   write_timeout: 2s
//   idle_timeout: 30s
// {{end}}`

const corsAllowedHeaders = `{{define "CORS_ALLOWED_HEADERS"}}
allowed_headers:
- "Authorization"
- "Content-Type"
- "If-Match"
- "If-None-Match"
- "Depth"
{{end}}`

const corsAllowedMethods = `{{define "CORS_ALLOWED_METHODS"}}
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
{{end}}`

const corsAllowedOrigins = `{{define "CORS_ALLOWED_ORIGINS"}}
allowed_origins:
- http://localhost
- http://localhost:*
- https://localhost
- https://localhost:*
- https://0.0.0.0:*
- https://*.aserto.com
- https://*aserto-console.netlify.app
{{end}}`

const controllerTemplate = `{{ define "CONTROLLER"}}
controller:
  enabled: true
  server:
    address: {{ .ControlPlane.Address }}
    client_cert_path: {{ .ControlPlane.ClientCertPath }}
    client_key_path: {{ .ControlPlane.ClientKeyPath }}
{{end}}`

const decisionLoggerTemplate = `{{ define "DECISION_LOGGER"}}
decision_logger:
  type: self
  config:
    store_directory: {{ .DecisionLogger.StorePath }}
    scribe:
      address: {{ .DecisionLogger.EMSAddress }}
      client_cert_path: {{ .DecisionLogger.ClientCertPath }}
      client_key_path: {{ .DecisionLogger.ClientKeyPath }}
      ack_wait_seconds: 30
      headers:
        Aserto-Tenant-Id: {{ .TenantID }}
    shipper:
      publish_timeout_seconds: 2
{{end}}`

const opaTemplate = `{{define "OPA"}}
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
{{end}}`

const opaDiscoveryTemplate = `{{define "OPA_DISCOVERY"}}
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
{{end}}`

const opaLocalPolicyTemplate = `{{define "OPA_LOCAL"}}
opa:
  instance_id: "-"
  graceful_shutdown_period_seconds: 2
  # max_plugin_wait_time_seconds: 30 set as default
  local_bundles:
    local_policy_image: {{ .LocalPolicyImage }}
    watch: true
    skip_verification: true
{{end}}`
