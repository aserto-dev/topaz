package todoApp.POST.todos

import input.user
import future.keywords.in
import data.todoApp.common.is_member_of

default allowed = false

allowed {
  allowedGroups := { "editor", "admin" }
  some group in allowedGroups
  is_member_of(user, group)
}
