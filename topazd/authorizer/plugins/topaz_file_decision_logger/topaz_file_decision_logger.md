# Topaz File Decision Logger plugin

Plugin configuration structure

```
opa:
  config:

    # plugins section
    plugins:
      topaz_file_decision_logger:
        enabled: true
        logger:
          filename: '/tmp/decisions.json' # default "", uses <processname>-lumberjack.log in os.TempDir()
          max_size: 100                   # default 100,  100Mb max file size.
          max_age: 2                      # default 0, is not to remove old log files based on age.
          max_backups: 2                  # default 0, is to retain all old log files (though MaxAge may still cause them to get deleted).
          local_time: false               # default false, is to use UTC time.
          compress: false                 # default false, is not to perform compression.
        policy_info:
          policy_name: 'rebac'
          registry_service: 'ghcr.io'
          registry_image: 'aserto-policies/policy-rebac'
          registry_tag: 'latest'
          digest: ''
```

## Updating from the deprecated Aserto Decision Log plugin (aserto_decision_log)

Replace `plugins.aserto_decision_log` section with the `plugins.topaz_file_decision_logger` and remove the `decision_logger` section completely.

Old deprecated `aserto_decision_log` plugin configuration 
```
opa:
  config:

    # plugins section
    plugins:
      aserto_decision_log:
        enabled: true
        policy_info:
          policy_id: ""
          policy_name: "rebac"
          instance_label: ""
          registry_service: "ghcr.io"
          registry_image: "aserto-policies/policy-rebac"
          registry_tag: "latest"
          digest: ""

decision_logger:
  type: file
  config:
    log_file_path: "/tmp/decisions-0.33.12.json"
    max_file_size_mb: 100
    max_file_count: 2
```

## NOTE regarding the OPA decision logs section 

See https://www.openpolicyagent.org/docs/configuration#decision-logs 

The OPA `config.decision_log` section is not utilized by the current `topaz_file_decision_logger` implementation, as the plugin logs `api.Decisions` instead of `logs.EventV1` instances.

```
type Decision struct {
	Id            string                 // unique id, replay a decision starting with this, also useful to de-dup
	Timestamp     *timestamppb.Timestamp // UTC time when the decision was made
	Path          string                 // Policy path used in decision
	User          *DecisionUser          // info about user for whom the decision as made
	Policy        *DecisionPolicy        // info about policy used for the decision
	Outcomes      map[string]bool        // outcome of the decisions specified in the policy context
	Resource      *structpb.Struct       // the resource context used in a decision
	Annotations   map[string]string      // annotations that may be added to a decision
}
```

As such you do NOT need to add the the plugin name to `decisions_logs.plugin`:

```
    decision_logs:
      console: false
      plugin: topaz_file_decision_logger # !!! DO NOT ADD THIS !!!
```
