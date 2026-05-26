package main

import rego.v1

test_chmod_777_denied if {
	some msg in deny with input as [
		{"Cmd": "run", "Stage": 0, "Value": ["chmod 777 /tmp"]},
	]
	contains(msg, "world-writable chmod")
}

test_chmod_0777_denied if {
	some msg in deny with input as [
		{"Cmd": "run", "Stage": 0, "Value": ["chmod 0777 /tmp"]},
	]
	contains(msg, "world-writable chmod")
}

test_chmod_recursive_777_denied if {
	some msg in deny with input as [
		{"Cmd": "run", "Stage": 0, "Value": ["chmod -R 777 /var/cache"]},
	]
	contains(msg, "world-writable chmod")
}

test_chmod_a_rwx_denied if {
	some msg in deny with input as [
		{"Cmd": "run", "Stage": 0, "Value": ["chmod a+rwx /opt/app"]},
	]
	contains(msg, "world-writable chmod")
}

test_chmod_ugo_rwx_denied if {
	some msg in deny with input as [
		{"Cmd": "run", "Stage": 0, "Value": ["chmod ugo+rwx /opt/app"]},
	]
	contains(msg, "world-writable chmod")
}

test_chmod_777_in_later_stage_reports_stage if {
	some msg in deny with input as [
		{"Cmd": "from", "Stage": 0, "Value": ["alpine@sha256:abc"]},
		{"Cmd": "from", "Stage": 1, "Value": ["alpine@sha256:def"]},
		{"Cmd": "run", "Stage": 1, "Value": ["chmod 777 /data"]},
	]
	msg == "RUN in stage 1 uses world-writable chmod (777 or equivalent)"
}

test_chmod_755_not_denied if {
	msgs := deny with input as [
		{"Cmd": "run", "Stage": 0, "Value": ["chmod 755 /usr/local/bin/app"]},
	]
	every msg in msgs {
		not contains(msg, "world-writable chmod")
	}
}

test_chmod_644_not_denied if {
	msgs := deny with input as [
		{"Cmd": "run", "Stage": 0, "Value": ["chmod 644 /etc/config"]},
	]
	every msg in msgs {
		not contains(msg, "world-writable chmod")
	}
}

test_chmod_u_rwx_not_denied if {
	msgs := deny with input as [
		{"Cmd": "run", "Stage": 0, "Value": ["chmod u+rwx /opt/app"]},
	]
	every msg in msgs {
		not contains(msg, "world-writable chmod")
	}
}

test_no_run_instructions_not_denied if {
	msgs := deny with input as [
		{"Cmd": "from", "Stage": 0, "Value": ["scratch"]},
		{"Cmd": "entrypoint", "Stage": 0, "Value": ["/app"]},
	]
	every msg in msgs {
		not contains(msg, "world-writable chmod")
	}
}
