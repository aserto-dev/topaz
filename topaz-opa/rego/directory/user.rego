package directory

import rego.v1

# METADATA
# title: ""
# description: User
# entrypoint: true
# authors:
# - gert@topaz.sh
ds_user(id) := result if {
	is_string(id)

	result := ds.object({"object_type": "user", "object_id": id})
}
