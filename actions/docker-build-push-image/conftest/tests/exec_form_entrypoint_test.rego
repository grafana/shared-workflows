package main

import rego.v1

test_entrypoint_shell_form_denied if {
	some msg in deny with input as [
		{"Cmd": "entrypoint", "Stage": 0, "JSON": false, "Value": ["/app/server"]},
	]
	contains(msg, "ENTRYPOINT uses shell form")
}

test_cmd_shell_form_denied if {
	some msg in deny with input as [
		{"Cmd": "cmd", "Stage": 0, "JSON": false, "Value": ["/app/server --port=8080"]},
	]
	contains(msg, "CMD uses shell form")
}

test_entrypoint_exec_form_not_denied if {
	msgs := deny with input as [
		{"Cmd": "entrypoint", "Stage": 0, "JSON": true, "Value": ["/app/server"]},
	]
	every msg in msgs {
		not contains(msg, "uses shell form")
	}
}

test_cmd_exec_form_not_denied if {
	msgs := deny with input as [
		{"Cmd": "cmd", "Stage": 0, "JSON": true, "Value": ["--port=8080"]},
	]
	every msg in msgs {
		not contains(msg, "uses shell form")
	}
}

test_entrypoint_exec_form_with_args_not_denied if {
	msgs := deny with input as [
		{"Cmd": "entrypoint", "Stage": 0, "JSON": true, "Value": ["/app/server", "--port=8080", "--config=/etc/app.yaml"]},
	]
	every msg in msgs {
		not contains(msg, "uses shell form")
	}
}

test_entrypoint_pointing_at_script_not_denied if {
	msgs := deny with input as [
		{"Cmd": "entrypoint", "Stage": 0, "JSON": true, "Value": ["/usr/local/bin/entrypoint.sh"]},
	]
	every msg in msgs {
		not contains(msg, "uses shell form")
	}
}

test_no_entrypoint_or_cmd_not_denied if {
	msgs := deny with input as [
		{"Cmd": "from", "Stage": 0, "Value": ["scratch"]},
		{"Cmd": "user", "Stage": 0, "Value": ["65534:65534"]},
	]
	every msg in msgs {
		not contains(msg, "uses shell form")
	}
}

test_both_exec_form_not_denied if {
	msgs := deny with input as [
		{"Cmd": "entrypoint", "Stage": 0, "JSON": true, "Value": ["/app/server"]},
		{"Cmd": "cmd", "Stage": 0, "JSON": true, "Value": ["--port=8080"]},
	]
	every msg in msgs {
		not contains(msg, "uses shell form")
	}
}

test_entrypoint_shell_form_cmd_exec_form_denied if {
	some msg in deny with input as [
		{"Cmd": "entrypoint", "Stage": 0, "JSON": false, "Value": ["/app/server"]},
		{"Cmd": "cmd", "Stage": 0, "JSON": true, "Value": ["--port=8080"]},
	]
	contains(msg, "ENTRYPOINT uses shell form")
}

test_entrypoint_exec_cmd_shell_denied if {
	some msg in deny with input as [
		{"Cmd": "entrypoint", "Stage": 0, "JSON": true, "Value": ["/app/server"]},
		{"Cmd": "cmd", "Stage": 0, "JSON": false, "Value": ["--port=8080"]},
	]
	contains(msg, "CMD uses shell form")
}
