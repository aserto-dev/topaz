package access

import data.access.req

# METADATA
# description: Evaluation
# entrypoint: true
evaluation(
	subject, # REQUIRED. The subject (or principal) of type Subject
	action, # REQUIRED. The action (or verb) of type Action.
	resource, # REQUIRED. The resource of type Resource.
	context, # OPTIONAL. The context (or environment) of type Context.
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
