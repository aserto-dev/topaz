package access

import data.access.req

# METADATA
# title: SubjectSearch
# description:
# custom:
#   category: AuthZEN
#   url: https://openid.net/specs/authorization-api-1_0.html#name-subject-search-api
# entrypoint: true
subject_search(
	subject, # REQUIRED. The subject (or principal) of type Subject, MUST contain a type, and SHOULD omit the id.
	action,
	resource,
	context,
	page,
) := az.subject_search({
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
