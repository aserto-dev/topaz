# AuthZEN Evaluations Logger plugin

Plugin configuration structure

```
plugins:
    authzen_logger:
      enabled: true
      logger:
        filename: "/var/log/opa/authzen-evaluations.log"
        max_size: 100 # Megabytes before splitting file
        max_age: 14 # Retain historical log files for 14 days
        max_backups: 5 # Retain up to 5 historical log files
        local_time: false # LocalTime determines if the time used for formatting the timestamps in backup files is the computer's local time.
        compress: true # Compress rolled files using gzip

decision_logs:
  plugin: authzen_logger
```

Plugin default configuration values:

```
plugins:
    authzen_logger:
      enabled: false
      logger:
        filename: "" # default "" uses <processname>-lumberjack.log in os.TempDir()
        max_size: 0 # default 100 megabytes.
        max_age: 0 # default is not to remove old log files based on age.
        max_backups: 0 # default is to retain all old log files (though MaxAge may still cause them to get deleted.).
        local_time: false # default is to use UTC time.
        compress: false # default is not to perform compression.
```
