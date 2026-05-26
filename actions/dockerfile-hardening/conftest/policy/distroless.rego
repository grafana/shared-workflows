package main

import rego.v1

distroless_prefixes := {
	"gcr.io/distroless/",
	"cgr.dev/chainguard/",
	"cgr.dev/chainguard-private/",
}

is_scratch(ref) if {
	ref == "scratch"
}

is_distroless(ref) if {
	some prefix in distroless_prefixes
	startswith(ref, prefix)
}

is_distroless(ref) if {
	contains(ref, "/distroless/")
}

deny contains msg if {
	ref := image_ref(final_from)
	not is_scratch(ref)
	not is_distroless(ref)
	msg := sprintf("final FROM image %q must be either scratch or a distroless base", [ref])
}
