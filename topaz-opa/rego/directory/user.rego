package directory

# METADATA
# description: User
# entrypoint: true
ds_user(id) := result if {
	is_string(id)
	object("user", id, false)
}
