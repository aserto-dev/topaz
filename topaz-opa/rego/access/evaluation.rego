package access

# METADATA
# description: Evaluation
# entrypoint: true
evaluation(
	subject,
	action,
	resource,
	context,
) := az.evaluation({
	"subject": {
		"type": subject.type,
		"id": subject_id,
		"properties": subject.properties,
	},
	"action": {
		"name": action.name,
		"properties": action.properties,
	},
	"resource": {
		"type": resource.type,
		"id": resource.id,
		"properties": resource.properties,
	},
	"context": context,
})
