package directory

# METADATA
# description: Graph
# entrypoint: true
ds_graph(
	object_type,
	object_id,
	relation,
	subject_type,
	subject_id,
	subject_relation,
	with_explain,
	with_trace,
) := ds.graph({
	"object_type": object_type,
	"object_id": object_id,
	"relation": relation,
	"subject_type": subject_type,
	"subject_id": subject_id,
	"subject_relation": subject_relation,
	"explain": false,
	"trace": false,
})
