package main

import rego.v1

test_copy_dot_in_final_stage_warned if {
	some msg in warn with input as [
		{"Cmd": "from", "Stage": 0, "Value": ["scratch"]},
		{"Cmd": "copy", "Stage": 0, "Flags": [], "Value": [".", "/app"]},
	]
	contains(msg, "ship source tree to runtime image")
}

test_copy_dot_in_builder_stage_not_warned if {
	msgs := warn with input as [
		{"Cmd": "from", "Stage": 0, "Value": ["golang:1.22@sha256:abc", "as", "builder"]},
		{"Cmd": "copy", "Stage": 0, "Flags": [], "Value": [".", "/src"]},
		{"Cmd": "from", "Stage": 1, "Value": ["scratch"]},
		{"Cmd": "copy", "Stage": 1, "Flags": ["--from=builder"], "Value": ["/src/app", "/app"]},
	]
	every msg in msgs {
		not contains(msg, "ship source tree to runtime image")
	}
}

test_copy_specific_file_in_final_stage_not_warned if {
	msgs := warn with input as [
		{"Cmd": "from", "Stage": 0, "Value": ["scratch"]},
		{"Cmd": "copy", "Stage": 0, "Flags": [], "Value": ["./app", "/app"]},
	]
	every msg in msgs {
		not contains(msg, "ship source tree to runtime image")
	}
}

test_copy_from_other_stage_with_dot_not_warned if {
	msgs := warn with input as [
		{"Cmd": "from", "Stage": 0, "Value": ["alpine@sha256:abc", "as", "builder"]},
		{"Cmd": "from", "Stage": 1, "Value": ["scratch"]},
		{"Cmd": "copy", "Stage": 1, "Flags": ["--from=builder"], "Value": [".", "/app"]},
	]
	every msg in msgs {
		not contains(msg, "ship source tree to runtime image")
	}
}

test_single_stage_copy_dot_warned if {
	some msg in warn with input as [
		{"Cmd": "from", "Stage": 0, "Value": ["alpine@sha256:abc"]},
		{"Cmd": "copy", "Stage": 0, "Flags": [], "Value": [".", "/app"]},
	]
	contains(msg, "ship source tree to runtime image")
}

test_no_copy_at_all_not_warned if {
	msgs := warn with input as [
		{"Cmd": "from", "Stage": 0, "Value": ["scratch"]},
		{"Cmd": "entrypoint", "Stage": 0, "Value": ["/app"]},
	]
	every msg in msgs {
		not contains(msg, "ship source tree to runtime image")
	}
}
