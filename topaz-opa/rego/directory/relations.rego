package directory

# METADATA
# description: Relations
# entrypoint: true
ds_relations(
	object_type,
	object_id,
	relation,
	subject_type,
	subject_id,
	subject_relation,
	with_objects,
	with_empty_subject_relation,
) := ds.relations({
	"object_type": object_type,
	"object_id": object_id,
	"relation": relation,
	"subject_type": subject_type,
	"subject_id": subject_id,
	"subject_relation": subject_relation,
	"with_objects": with_objects,
	"with_empty_subject_relation": with_empty_subject_relation,
})
