package main

import rego.v1

test_copy_from_latest_denied if {
	some msg in deny with input as [
		{"Cmd": "copy", "Stage": 0, "Flags": ["--from=alpine:latest"], "Value": ["/foo", "/foo"]},
	]
	contains(msg, "uses :latest tag")
}

test_copy_from_untagged_denied if {
	some msg in deny with input as [
		{"Cmd": "copy", "Stage": 0, "Flags": ["--from=alpine"], "Value": ["/foo", "/foo"]},
	]
	contains(msg, "has no tag (implicit :latest)")
}

test_copy_from_pinned_not_denied if {
	msgs := deny with input as [
		{"Cmd": "copy", "Stage": 0, "Flags": ["--from=alpine:3.18@sha256:abc"], "Value": ["/foo", "/foo"]},
	]
	every msg in msgs {
		not contains(msg, "COPY --from=")
	}
}

test_copy_from_digest_only_not_denied if {
	msgs := deny with input as [
		{"Cmd": "copy", "Stage": 0, "Flags": ["--from=alpine@sha256:abc"], "Value": ["/foo", "/foo"]},
	]
	every msg in msgs {
		not contains(msg, "COPY --from=")
	}
}

test_copy_from_stage_alias_not_denied if {
	msgs := deny with input as [
		{"Cmd": "from", "Stage": 0, "Value": ["alpine:3.18@sha256:abc", "as", "builder"]},
		{"Cmd": "from", "Stage": 1, "Value": ["scratch"]},
		{"Cmd": "copy", "Stage": 1, "Flags": ["--from=builder"], "Value": ["/app", "/app"]},
	]
	every msg in msgs {
		not contains(msg, "COPY --from=")
	}
}

test_plain_copy_without_from_flag_not_denied if {
	msgs := deny with input as [
		{"Cmd": "copy", "Stage": 0, "Flags": [], "Value": ["./app", "/app"]},
	]
	every msg in msgs {
		not contains(msg, "COPY --from=")
	}
}
