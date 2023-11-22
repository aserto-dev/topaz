package todoApp.PUT.todos.__id

import input.user
import input.resource
import future.keywords.in
import data.todoApp.common.is_member_of

default allowed = false

allowed {
  is_member_of(user, "editor")
  user.key == resource.ownerID
}

allowed {
  is_member_of(user, "evil_genius")
}
