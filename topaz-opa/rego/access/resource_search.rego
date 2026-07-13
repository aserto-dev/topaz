package access

import data.access.req

# METADATA
# title: ResourceSearch
# description:
# custom:
#   category: AuthZEN
#   url: https://openid.net/specs/authorization-api-1_0.html#name-resource-search-api
# entrypoint: true
resource_search(
	subject, # REQUIRED. The subject (or principal) of type Subject.
	action, # REQUIRED. The action (or verb) of type Action.
	resource, # REQUIRED. The resource of type Resource, Resource MUST contain a type, and SHOULD omit the id.
	context, # OPTIONAL. Contextual data about the request.
	page, # OPTIONAL. A page object for paginated requests.
) := az.resource_search({
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
