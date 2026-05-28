package main

import rego.v1

# ── FROM coverage ────────────────────────────────────────────────────

test_from_tag_only_denied if {
	some msg in deny with input as [
		{"Cmd": "from", "Stage": 0, "Value": ["alpine:3.18"]},
	]
	contains(msg, "is not pinned by a supported digest")
}

test_from_sha256_digest_pinned_not_denied if {
	msgs := deny with input as [
		{"Cmd": "from", "Stage": 0, "Value": ["alpine:3.18@sha256:abc"]},
	]
	every msg in msgs {
		not contains(msg, "is not pinned by a supported digest")
	}
}

test_from_sha512_digest_pinned_not_denied if {
	msgs := deny with input as [
		{"Cmd": "from", "Stage": 0, "Value": ["alpine:3.18@sha512:abc"]},
	]
	every msg in msgs {
		not contains(msg, "is not pinned by a supported digest")
	}
}

test_from_blake3_digest_pinned_not_denied if {
	msgs := deny with input as [
		{"Cmd": "from", "Stage": 0, "Value": ["alpine:3.18@blake3:abc"]},
	]
	every msg in msgs {
		not contains(msg, "is not pinned by a supported digest")
	}
}

test_from_unsupported_digest_denied if {
	some msg in deny with input as [
		{"Cmd": "from", "Stage": 0, "Value": ["alpine:3.18@md5:abc"]},
	]
	contains(msg, "is not pinned by a supported digest")
}

test_from_scratch_not_denied if {
	msgs := deny with input as [
		{"Cmd": "from", "Stage": 0, "Value": ["scratch"]},
	]
	every msg in msgs {
		not contains(msg, "is not pinned by a supported digest")
	}
}

test_from_stage_alias_reference_not_denied if {
	msgs := deny with input as [
		{"Cmd": "from", "Stage": 0, "Value": ["alpine:3.18@sha256:abc", "as", "builder"]},
		{"Cmd": "from", "Stage": 1, "Value": ["builder"]},
	]
	every msg in msgs {
		not contains(msg, "is not pinned by a supported digest")
	}
}

test_multistage_unpinned_builder_denied if {
	some msg in deny with input as [
		{"Cmd": "from", "Stage": 0, "Value": ["golang:1.22", "as", "builder"]},
		{"Cmd": "from", "Stage": 1, "Value": ["scratch"]},
	]
	contains(msg, "is not pinned by a supported digest")
}

# ── COPY --from=<image> coverage ─────────────────────────────────────

test_copy_from_tag_only_denied if {
	some msg in deny with input as [
		{"Cmd": "copy", "Stage": 0, "Flags": ["--from=alpine:3.18"], "Value": ["/foo", "/foo"]},
	]
	contains(msg, "COPY --from=\"alpine:3.18\" is not pinned by a supported digest")
}

test_copy_from_latest_denied if {
	some msg in deny with input as [
		{"Cmd": "copy", "Stage": 0, "Flags": ["--from=alpine:latest"], "Value": ["/foo", "/foo"]},
	]
	contains(msg, "COPY --from=\"alpine:latest\" is not pinned by a supported digest")
}

test_copy_from_untagged_denied if {
	some msg in deny with input as [
		{"Cmd": "copy", "Stage": 0, "Flags": ["--from=alpine"], "Value": ["/foo", "/foo"]},
	]
	contains(msg, "COPY --from=\"alpine\" is not pinned by a supported digest")
}

test_copy_from_sha256_digest_pinned_not_denied if {
	msgs := deny with input as [
		{"Cmd": "copy", "Stage": 0, "Flags": ["--from=alpine:3.18@sha256:abc"], "Value": ["/foo", "/foo"]},
	]
	every msg in msgs {
		not contains(msg, "is not pinned by a supported digest")
	}
}

test_copy_from_sha512_digest_pinned_not_denied if {
	msgs := deny with input as [
		{"Cmd": "copy", "Stage": 0, "Flags": ["--from=alpine:3.18@sha512:abc"], "Value": ["/foo", "/foo"]},
	]
	every msg in msgs {
		not contains(msg, "is not pinned by a supported digest")
	}
}

test_copy_from_blake3_digest_pinned_not_denied if {
	msgs := deny with input as [
		{"Cmd": "copy", "Stage": 0, "Flags": ["--from=alpine:3.18@blake3:abc"], "Value": ["/foo", "/foo"]},
	]
	every msg in msgs {
		not contains(msg, "is not pinned by a supported digest")
	}
}

test_copy_from_digest_only_not_denied if {
	msgs := deny with input as [
		{"Cmd": "copy", "Stage": 0, "Flags": ["--from=alpine@sha256:abc"], "Value": ["/foo", "/foo"]},
	]
	every msg in msgs {
		not contains(msg, "is not pinned by a supported digest")
	}
}

test_copy_from_unsupported_digest_denied if {
	some msg in deny with input as [
		{"Cmd": "copy", "Stage": 0, "Flags": ["--from=alpine@md5:abc"], "Value": ["/foo", "/foo"]},
	]
	contains(msg, "is not pinned by a supported digest")
}

test_copy_from_stage_alias_not_denied if {
	msgs := deny with input as [
		{"Cmd": "from", "Stage": 0, "Value": ["alpine:3.18@sha256:abc", "as", "builder"]},
		{"Cmd": "from", "Stage": 1, "Value": ["scratch"]},
		{"Cmd": "copy", "Stage": 1, "Flags": ["--from=builder"], "Value": ["/app", "/app"]},
	]
	every msg in msgs {
		not contains(msg, "is not pinned by a supported digest")
	}
}

test_plain_copy_without_from_flag_not_denied if {
	msgs := deny with input as [
		{"Cmd": "copy", "Stage": 0, "Flags": [], "Value": ["./app", "/app"]},
	]
	every msg in msgs {
		not contains(msg, "is not pinned by a supported digest")
	}
}
