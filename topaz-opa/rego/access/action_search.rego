package access

# METADATA
# description: ActionSearch
# entrypoint: true
action_search(
	subject,
	action,
	resource,
	context,
	page,
) := az.action_search({
	"subject": {
		"type": subject.type,
		"id": subject.id,
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
	"page": {
		"size": page.size,
		"next_token": page.next_token,
	},
})
