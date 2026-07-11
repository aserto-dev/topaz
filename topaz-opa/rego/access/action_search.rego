package access

import data.access.req

# METADATA
# title: ActionSearch
# description:
# custom:
#   category: AuthZEN
#   url: https://openid.net/specs/authorization-api-1_0.html#name-action-search-api
# entrypoint: true
action_search(
	subject, # REQUIRED. The subject (or principal) of type Subject.
	resource, # REQUIRED. The resource of type Resource.
	action, # OMITTED.  The action key is omitted from the Action Search request payload.
	context, # OPTIONAL. Contextual data about the request.
	page, # OPTIONAL. A page object for paginated requests.
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
