package main

import rego.v1

# Every image reference in a Dockerfile must be pinned to a digest.
# Covers both FROM and COPY --from=<image>. Excludes scratch (for FROM)
# and stage aliases (for both).

deny contains msg if {
	some instr in from_instructions
	ref := image_ref(instr)
	ref != "scratch"
	not is_stage_alias(ref)
	not contains(ref, "@sha256:")
	msg := sprintf("FROM %q is not pinned by @sha256: digest", [ref])
}

deny contains msg if {
	some instr in input
	instr.Cmd == "copy"
	some flag in instr.Flags
	startswith(flag, "--from=")
	ref := replace(flag, "--from=", "")
	not is_stage_alias(ref)
	not contains(ref, "@sha256:")
	msg := sprintf("COPY --from=%q is not pinned by @sha256: digest", [ref])
}
