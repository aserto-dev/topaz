# topaz plugin

Plugin configuration structure

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
      enable_directory_builtins: true
      enable_access_builtins: true
```

Plugin configuration default values:

```
    topaz:
      enabled: false
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
      request_timeout: 5s
      enable_directory_builtins: false
      enable_access_builtins: false
```
