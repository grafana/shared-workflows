package main

import rego.v1

is_golang(ref) if {
	startswith(ref, "golang:")
}

is_golang(ref) if {
	startswith(ref, "golang@")
}

uses_golang_stage if {
	some instr in from_instructions
	is_golang(image_ref(instr))
}

deny contains msg if {
	uses_golang_stage
	ref := image_ref(final_from)
	is_golang(ref)
	msg := sprintf("final FROM %q ships the Go toolchain — runtime stage must not use golang:* base", [ref])
}
