package directory

import rego.v1

check_schema := {
	"type": "object",
	"properties": {
		"object_type": {"type": "string"},
		"object_id": {"type": "string"},
		"relation": {"type": "string"},
		"subject_type": {"type": "string"},
		"subject_id": {"type": "string"},
		"with_trace": {"type": "bool"},
	},
	"required": [
		"object_type",
		"object_id",
		"relation",
		"subject_type",
		"subject_id",
	],
}

# METADATA
# description: ds_check
# entrypoint: true
ds_check(
	object_type,
	object_id,
	relation,
	subject_type,
	subject_id,
	with_trace,
) := ds.check({
	"object_type": object_type,
	"object_id": object_id,
	"relation": relation,
	"subject_type": subject_type,
	"subject_id": subject_id,
	"trace": with_trace,
})
