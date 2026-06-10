package main

import rego.v1

root_identifiers := {"0", "root"}

is_root_user(value) if {
	uid_part := split(value, ":")[0]
	uid_part in root_identifiers
}

deny contains msg if {
	count(user_indices) == 0
	msg := "Dockerfile has no USER instruction — container would run as root"
}

deny contains msg if {
	instr := final_user
	some value in instr.Value
	is_root_user(value)
	msg := sprintf("final USER %q is root — containers must run as a non-root user", [value])
}
