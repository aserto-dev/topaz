package directory

# METADATA
# description: Object
# entrypoint: true
ds_object(
	object_type,
	object_id,
	with_relations,
) := ds.object({
	"object_type": object_type,
	"object_id": object_id,
	"with_relations": with_relations,
})

# ds_object(object_type, object_id, with_relations) if := _ {
#   not is_string(object_type)
#   trace(sprintf("object_type: a must be string, got %v", [object_type]))
#   false
# }
# ds_object(object_type, object_id, with_relations) if := _ {
#   not is_string(object_id)
#   trace(sprintf("id: a must be string, got %v", [object_id]))
#   false
# }
# ds_object(object_type, object_id, with_relations) if := _ {
#   not is_boolean(with_relations)
#   trace(sprintf("id: a must be boolean, got %v", [with_relations]))
#   false
# }
