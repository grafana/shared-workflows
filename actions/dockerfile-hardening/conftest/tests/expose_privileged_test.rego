package main

import rego.v1

test_port_80_denied if {
	some msg in deny with input as [
		{"Cmd": "expose", "Stage": 0, "Value": ["80"]},
	]
	contains(msg, "uses privileged port")
}

test_port_443_denied if {
	some msg in deny with input as [
		{"Cmd": "expose", "Stage": 0, "Value": ["443"]},
	]
	contains(msg, "uses privileged port")
}

test_port_22_with_protocol_denied if {
	some msg in deny with input as [
		{"Cmd": "expose", "Stage": 0, "Value": ["22/tcp"]},
	]
	contains(msg, "uses privileged port")
}

test_port_1023_denied if {
	some msg in deny with input as [
		{"Cmd": "expose", "Stage": 0, "Value": ["1023"]},
	]
	contains(msg, "uses privileged port")
}

test_port_1024_not_denied if {
	msgs := deny with input as [
		{"Cmd": "expose", "Stage": 0, "Value": ["1024"]},
	]
	every msg in msgs {
		not contains(msg, "uses privileged port")
	}
}

test_port_8080_not_denied if {
	msgs := deny with input as [
		{"Cmd": "expose", "Stage": 0, "Value": ["8080"]},
	]
	every msg in msgs {
		not contains(msg, "uses privileged port")
	}
}

test_port_8443_with_protocol_not_denied if {
	msgs := deny with input as [
		{"Cmd": "expose", "Stage": 0, "Value": ["8443/tcp"]},
	]
	every msg in msgs {
		not contains(msg, "uses privileged port")
	}
}

test_multiple_ports_mixed_flags_first_offender if {
	some msg in deny with input as [
		{"Cmd": "expose", "Stage": 0, "Value": ["8080", "80"]},
	]
	contains(msg, "uses privileged port")
}
