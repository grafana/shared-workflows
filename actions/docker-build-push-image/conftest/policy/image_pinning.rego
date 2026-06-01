package main

import rego.v1

# Every image reference in a Dockerfile must be pinned to a digest.
# Covers both FROM and COPY --from=<image>. Excludes scratch (for FROM)
# and stage aliases (for both). Accepted digest algorithms: sha256,
# sha512, blake3.

supported_digest_prefixes := ["@sha256:", "@sha512:", "@blake3:"]

has_supported_digest(ref) if {
	some prefix in supported_digest_prefixes
	contains(ref, prefix)
}

deny contains msg if {
	some instr in from_instructions
	ref := image_ref(instr)
	ref != "scratch"
	not is_stage_alias(ref)
	not has_supported_digest(ref)
	msg := sprintf("FROM %q is not pinned by a supported digest (@sha256:, @sha512:, @blake3:)", [ref])
}

deny contains msg if {
	some instr in input
	instr.Cmd == "copy"
	some flag in instr.Flags
	startswith(flag, "--from=")
	ref := replace(flag, "--from=", "")
	not is_stage_alias(ref)
	not has_supported_digest(ref)
	msg := sprintf("COPY --from=%q is not pinned by a supported digest (@sha256:, @sha512:, @blake3:)", [ref])
}
