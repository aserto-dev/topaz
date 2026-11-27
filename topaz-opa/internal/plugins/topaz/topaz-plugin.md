# topaz plugin

Topaz plugin configuration:

```
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
