package directory

import rego.v1

# person_sum(p) if = r {
#     is_object(p)
#     p.name
#     is_string(p.name)
#     is_number(p.age)
#     r := example.process(p.name, p.age)
# }
