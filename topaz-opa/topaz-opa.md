# topaz-opa

`topaz-opa` is a branded OPA instance, meaning it contains **all** the OPA functionality you know and love, augmented with a `plugin` to connect to the Topaz Directory service, and the OpenID AuthZEN Access and Topaz Directory built-ins. Enabling the integration of AuthZEN and Topaz authorization decisions as well as Topaz Directory data within Rego.

## Getting started

### Install topaz-opa

	
MacOS:

* Homebrew: 

```
brew trust --cask aserto-dev/tap/topaz-opa
brew install --cask aserto-dev/tap/topaz-opa
```

* Download release from Github 

```
gh release download --repo https://github.com/aserto-dev/topaz 
```

Linux:

	* Homebrew 

```
brew trust --cask aserto-dev/tap/topaz-opa
brew install --cask aserto-dev/tap/topaz-opa
```

	* Download release from Github 

```
gh release download --repo https://github.com/aserto-dev/topaz 
```

Windows:
	
  * WinGet

```
winget install topaz.topaz-opa
```

	* Download release from GitHub

```
gh release download --repo https://github.com/aserto-dev/topaz 
```

### Using topaz-opa

The `topaz-opa` command line exposes the exact same options as OPA. Execute `topaz-opa` to inspect the available command options.

```
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
  -h, --help[=true|false]   help for topaz-opa

Use "topaz-opa [command] --help" for more information about a command.
```

The `config.yaml` file used by `topaz-opa`, builds upon a normal OPA [configuration](https://www.openpolicyagent.org/docs/configuration) file, augemented with the topaz plugin configuration.


### Minimal OPA + topaz plugin configuration example:

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
    resource: ghcr.io/gertd/policy-rebac:latest
    persist: true
    polling:
      min_delay_seconds: 30
      max_delay_seconds: 120

plugins:
    topaz:                     # plugin name
      enabled: true            # enable plugin, default false
      connection:              # topaz directory client connection configuration
        address: :9292         # connect to local topaz directory service, using the default gRPC port 9292
        token: ""              # use an OAuth2.0 token to authenticate with the service
        api_key: ""            # use an authorization API key to authenticate with the service
        client_cert_path: ""   # configure the client certificate of the mTLS connections, both client certificate and private key paths must be specified
        client_key_path: ""    # configure the client key of the mTLS connections, both client certificate and private key paths must be specified
        ca_cert_path: ""       # treat the specified certificate file as a trusted root CA
        insecure: true         # disables TLS verification
        no_tls: false          # disables transport security, uses a plain-text HTTP/2 connection
        no_proxy: false        # sets connection option to bypasses any configured HTTP proxy
        headers:               # adds additional (gRPC) header(s) to the client config instance
          # "header": "value". # optional header name-value pair(s)
      request_timeout: 5s      # built-in request timeout, default 5s
```

### Available built-ins

OpenID [AuthZEN](https://openid.net/specs/authorization-api-1_0.html) Access Service built-ins:

| built-in name                                                                                      |
| -------------------------------------------------------------------------------------------------- |
| [az.evaluation](https://openid.net/specs/authorization-api-1_0.html#name-access-evaluation-api)    |
| [az.evaluations](https://openid.net/specs/authorization-api-1_0.html#name-access-evaluations-api)  |
| [az.action_search](https://openid.net/specs/authorization-api-1_0.html#name-action-search-api)     |
| [az.resource_search](https://openid.net/specs/authorization-api-1_0.html#name-resource-search-api) | 
| [az.subject_seatch](https://openid.net/specs/authorization-api-1_0.html#name-subject-search-api)   |

Topaz Directory Service built-ins:

| built-in name                                                                                      |
| -------------------------------------------------------------------------------------------------- |
| [ds.check](https://www.topaz.sh/docs/directory/built-ins#dscheck)                                  |
| [ds.checks](https://www.topaz.sh/docs/directory/built-ins#dschecks)                                |
| [ds.graph](https://www.topaz.sh/docs/directory/built-ins#dsgraph)                                  |
| [ds.object](https://www.topaz.sh/docs/directory/built-ins#dsobject)                                |
| [ds.relation](https://www.topaz.sh/docs/directory/built-ins#dsrelation)                            |

## Built-in Examples

> **NOTE**: The following examples will be using the Google Drive template, which provides the directory manifest and data. You can install the template using: `topaz templates install gdrive`.



