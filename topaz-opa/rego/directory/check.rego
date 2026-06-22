package directory

# METADATA
# description: Check
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
