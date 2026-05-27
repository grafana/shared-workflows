package main

import rego.v1

test_no_user_instruction_denied if {
	some msg in deny with input as [
		{"Cmd": "from", "Stage": 0, "Value": ["scratch"]},
	]
	contains(msg, "has no USER instruction")
}

test_user_zero_uid_denied if {
	some msg in deny with input as [
		{"Cmd": "from", "Stage": 0, "Value": ["scratch"]},
		{"Cmd": "user", "Stage": 0, "Value": ["0"]},
	]
	contains(msg, "is root")
}

test_user_zero_uid_gid_denied if {
	some msg in deny with input as [
		{"Cmd": "from", "Stage": 0, "Value": ["scratch"]},
		{"Cmd": "user", "Stage": 0, "Value": ["0:0"]},
	]
	contains(msg, "is root")
}

test_user_root_name_denied if {
	some msg in deny with input as [
		{"Cmd": "from", "Stage": 0, "Value": ["scratch"]},
		{"Cmd": "user", "Stage": 0, "Value": ["root"]},
	]
	contains(msg, "is root")
}

test_uid_with_zero_substring_not_treated_as_root if {
	msgs := deny with input as [
		{"Cmd": "from", "Stage": 0, "Value": ["scratch"]},
		{"Cmd": "user", "Stage": 0, "Value": ["10:10"]},
	]
	every msg in msgs {
		not contains(msg, "is root")
	}
}

test_later_root_after_compliant_user_denied if {
	some msg in deny with input as [
		{"Cmd": "from", "Stage": 0, "Value": ["scratch"]},
		{"Cmd": "user", "Stage": 0, "Value": ["65534:65534"]},
		{"Cmd": "user", "Stage": 0, "Value": ["root"]},
	]
	contains(msg, "is root")
}

test_compliant_user_after_root_not_denied if {
	msgs := deny with input as [
		{"Cmd": "from", "Stage": 0, "Value": ["scratch"]},
		{"Cmd": "user", "Stage": 0, "Value": ["root"]},
		{"Cmd": "user", "Stage": 0, "Value": ["65534:65534"]},
	]
	every msg in msgs {
		not contains(msg, "is root")
	}
}

test_user_only_in_builder_stage_denied if {
	some msg in deny with input as [
		{"Cmd": "from", "Stage": 0, "Value": ["alpine@sha256:abc", "as", "builder"]},
		{"Cmd": "user", "Stage": 0, "Value": ["1000:1000"]},
		{"Cmd": "from", "Stage": 1, "Value": ["scratch"]},
		{"Cmd": "copy", "Stage": 1, "Flags": ["--from=builder"], "Value": ["/app", "/app"]},
	]
	contains(msg, "has no USER instruction")
}

test_user_in_builder_and_final_stage_uses_final if {
	msgs := deny with input as [
		{"Cmd": "from", "Stage": 0, "Value": ["alpine@sha256:abc", "as", "builder"]},
		{"Cmd": "user", "Stage": 0, "Value": ["root"]},
		{"Cmd": "from", "Stage": 1, "Value": ["scratch"]},
		{"Cmd": "user", "Stage": 1, "Value": ["65534:65534"]},
		{"Cmd": "copy", "Stage": 1, "Flags": ["--from=builder"], "Value": ["/app", "/app"]},
	]
	every msg in msgs {
		not contains(msg, "is root")
	}
}
