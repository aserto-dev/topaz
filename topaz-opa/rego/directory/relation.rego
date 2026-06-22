package directory

# METADATA
# description: Relation
# entrypoint: true
ds_relation(
	object_type,
	object_id,
	relation,
	subject_type,
	subject_id,
	subject_relation,
	with_objects,
) := ds.relation({
	"object_type": object_type,
	"object_id": object_id,
	"relation": relation,
	"subject_type": subject_type,
	"subject_id": subject_id,
	"subject_relation": subject_relation,
	"with_objects": with_objects,
})
