package access.req

subject := {
	type,
	id,
	properties,
}

action := {
	name,
	properties,
}

resource := {
	type,
	id,
	properties,
}

properties := {}

context := {}

page := {
	size,
	next_token,
}

evaluation := {
	subject,
	action,
	resource,
	context,
}
