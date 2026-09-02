package mycars.GET

default allowed = false
default visible = false
default enabled = false

allowed {
    trace("allowed block")

    u = input.user
    u.attributes.properties.department == "Sales Engagement Management"
}

visible {
    trace("visible block")

	allowed
}

enabled {
    trace("enabled block")

	allowed
}
