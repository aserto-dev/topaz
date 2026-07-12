# Aserto Edge plugin

The `aserto_edge` plugin synchronizes the local directory with the data from the remote directory, configured by the `aserto_edge` plugin.

```
opa:
  config:

    # plugins section
    plugins:
      aserto_edge:
        enabled: false 
        addr: ""                    # gRPC directory service address.
        apikey: ""                  # directory API key.
        timeout: 5                  # gRPC connection timeout in seconds.
        sync_interval: 1            # sync run interval in minutes.
        insecure: true              # when using TLS connections, skip verification of the server certificate. 
        page_size: 100              # deprecated: no longer used.
        client_cert_path: ""        # when using mTLS connections, ClientCertPath is the path of the client's certificate file.
        client_key_path: ""         # when using mTLS connections, ClientKeyPath is the path of the client's private key file.
        no_tls: false               # disable TLS and use a plaintext connection.
        no_proxy: false             # bypasses any configured HTTP proxy.
        headers:                    # additional headers to include in requests to the service.
```