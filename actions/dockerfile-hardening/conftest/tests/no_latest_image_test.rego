package main

import rego.v1

test_explicit_latest_tag_denied if {
	some msg in deny with input as [
		{"Cmd": "from", "Stage": 0, "Value": ["alpine:latest"]},
	]
	contains(msg, "uses :latest tag")
}

test_implicit_no_tag_denied if {
	some msg in deny with input as [
		{"Cmd": "from", "Stage": 0, "Value": ["alpine"]},
	]
	contains(msg, "has no tag (implicit :latest)")
}

test_versioned_tag_not_denied if {
	msgs := deny with input as [
		{"Cmd": "from", "Stage": 0, "Value": ["alpine:3.18@sha256:abc"]},
	]
	every msg in msgs {
		not contains(msg, ":latest")
	}
}

test_scratch_not_denied if {
	msgs := deny with input as [
		{"Cmd": "from", "Stage": 0, "Value": ["scratch"]},
	]
	every msg in msgs {
		not contains(msg, ":latest")
	}
}

test_digest_only_not_denied if {
	msgs := deny with input as [
		{"Cmd": "from", "Stage": 0, "Value": ["alpine@sha256:abc"]},
	]
	every msg in msgs {
		not contains(msg, ":latest")
	}
}

test_stage_alias_reference_not_denied if {
	msgs := deny with input as [
		{"Cmd": "from", "Stage": 0, "Value": ["alpine:3.18@sha256:abc", "as", "builder"]},
		{"Cmd": "from", "Stage": 1, "Value": ["builder"]},
	]
	every msg in msgs {
		not contains(msg, ":latest")
	}
}

test_latest_in_aliased_from_denied if {
	some msg in deny with input as [
		{"Cmd": "from", "Stage": 0, "Value": ["alpine:latest", "as", "builder"]},
	]
	contains(msg, "uses :latest tag")
}
