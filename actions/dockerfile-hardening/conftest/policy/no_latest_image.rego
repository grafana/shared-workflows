package main

import rego.v1

deny contains msg if {
	some instr in from_instructions
	ref := strip_alias(image_ref(instr))
	endswith(ref, ":latest")
	msg := sprintf("FROM %q uses :latest tag", [ref])
}

deny contains msg if {
	some instr in from_instructions
	ref := strip_alias(image_ref(instr))
	ref != "scratch"
	not contains(ref, ":")
	not contains(ref, "@sha256:")
	not is_stage_alias(ref)
	msg := sprintf("FROM %q has no tag (implicit :latest)", [ref])
}
