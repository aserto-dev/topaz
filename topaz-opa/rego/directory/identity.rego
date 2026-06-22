package directory

# METADATA
# description: Identity
# entrypoint: true
ds_identity(id) := result if {
	is_string(id)
	object("user", id, false)
}
