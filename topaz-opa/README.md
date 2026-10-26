# topaz-opa 

Topaz OPA is [OPA](https://www.openpolicyagent.org/) extended with the `topaz` plugin to connect to the Topaz Directory Service, and enable the Topaz built-ins to enable rego policies to interact with the Topaz Directory.

To enable the `topaz` plugin, add the following plugin configuration fragment to your OPA configuration:

```
plugins:
    topaz:
      enabled: true
      connection:
        address: localhost:9292
        token: ""
        api_key: ""
        client_cert_path: ""
        client_key_path: ""
        ca_cert_path: ""
        insecure: true
        no_tls: false
        no_proxy: false
        headers:
          "key1": "val1"
          "key2": "val2"
      request_timeout: 10s
```

and start `topaz-opa` with the configuration.



## Enable the plugin


## Topaz Directory connection


## Complete example configuration

An example of acomplete OPA config which include the 'topaz' plugin section looks like this:

```
services:
  ghcr-registry:
    url: https://ghcr.io
    type: oci
    credentials:
      bearer:
        scheme: "Bearer"
        token: "${GIT_TOKEN}"

bundles:
  authz:
    service: ghcr-registry
    resource: ghcr.io/aserto-policies/policy-rebac:2.5.1
    persist: false
    polling:
      min_delay_seconds: 30
      max_delay_seconds: 120

plugins:
    topaz:
      enabled: true
      connection:
        address: :9292
        token: ""
        api_key: ""
        client_cert_path: ""
        client_key_path: ""
        ca_cert_path: ""
        insecure: true
        no_tls: false
        no_proxy: false
        headers:
          # "header": "value"
      request_timeout: 5s
```

## topaz-cli command line interface

The command line interface of `topaz-opa` exactly matches OPA, only the entrypoint is `topaz-opa` instead of `opa`.

```
topaz-opa

Extended Open Policy Agent (OPA) with Topaz support.

Usage:
  topaz-opa [command]

Available Commands:
  bench        Benchmark a Rego query
  build        Build an topaz-opa bundle
  capabilities Print the capabilities of topaz-opa
  check        Check Rego source files
  completion   Generate the autocompletion script for the specified shell
  deps         Analyze Rego query dependencies
  eval         Evaluate a Rego query
  exec         Execute against input files
  fmt          Format Rego source files
  help         Help about any command
  inspect      Inspect topaz-opa bundle(s)
  parse        Parse Rego source file
  run          Start topaz-opa in interactive or server mode
  sign         Generate an topaz-opa bundle signature
  test         Execute Rego test cases
  version      Print the version of topaz-opa

Flags:
  -h, --help   help for topaz-opa

Use "topaz-opa [command] --help" for more information about a command.
```

