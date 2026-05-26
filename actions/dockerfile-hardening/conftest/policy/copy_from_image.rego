package main

import rego.v1

copy_from_refs contains entry if {
	some instr in input
	instr.Cmd == "copy"
	some flag in instr.Flags
	startswith(flag, "--from=")
	entry := {"ref": replace(flag, "--from=", "")}
}

deny contains msg if {
	some entry in copy_from_refs
	not is_stage_alias(entry.ref)
	endswith(entry.ref, ":latest")
	msg := sprintf("COPY --from=%q uses :latest tag", [entry.ref])
}

deny contains msg if {
	some entry in copy_from_refs
	not is_stage_alias(entry.ref)
	not contains(entry.ref, ":")
	not contains(entry.ref, "@sha256:")
	msg := sprintf("COPY --from=%q has no tag (implicit :latest)", [entry.ref])
}
