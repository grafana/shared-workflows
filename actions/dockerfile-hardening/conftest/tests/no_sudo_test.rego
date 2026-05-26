package main

import rego.v1

test_plain_sudo_denied if {
	some msg in deny with input as [
		{"Cmd": "run", "Stage": 0, "Value": ["sudo whoami"]},
	]
	contains(msg, "invokes sudo")
}

test_sudo_then_semicolon_denied if {
	some msg in deny with input as [
		{"Cmd": "run", "Stage": 0, "Value": ["sudo;ls"]},
	]
	contains(msg, "invokes sudo")
}

test_sudo_backticked_denied if {
	some msg in deny with input as [
		{"Cmd": "run", "Stage": 0, "Value": ["echo `sudo whoami`"]},
	]
	contains(msg, "invokes sudo")
}

test_sudo_pipe_denied if {
	some msg in deny with input as [
		{"Cmd": "run", "Stage": 0, "Value": ["sudo|tee log"]},
	]
	contains(msg, "invokes sudo")
}

test_sudo_in_later_stage_reports_stage if {
	some msg in deny with input as [
		{"Cmd": "from", "Stage": 0, "Value": ["alpine@sha256:abc"]},
		{"Cmd": "from", "Stage": 1, "Value": ["alpine@sha256:def"]},
		{"Cmd": "run", "Stage": 1, "Value": ["sudo apk add curl"]},
	]
	msg == "RUN in stage 1 invokes sudo"
}

test_pseudo_random_not_denied if {
	msgs := deny with input as [
		{"Cmd": "run", "Stage": 0, "Value": ["echo pseudo-random && ls /usr/sudoers.d"]},
	]
	every msg in msgs {
		not contains(msg, "invokes sudo")
	}
}

test_sudoers_path_not_denied if {
	msgs := deny with input as [
		{"Cmd": "run", "Stage": 0, "Value": ["cat /etc/sudoers"]},
	]
	every msg in msgs {
		not contains(msg, "invokes sudo")
	}
}

test_no_run_instructions_not_denied if {
	msgs := deny with input as [
		{"Cmd": "from", "Stage": 0, "Value": ["scratch"]},
		{"Cmd": "entrypoint", "Stage": 0, "Value": ["/app"]},
	]
	every msg in msgs {
		not contains(msg, "invokes sudo")
	}
}
