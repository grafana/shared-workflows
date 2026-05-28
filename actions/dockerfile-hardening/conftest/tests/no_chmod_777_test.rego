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

# ── symbolic "other-write" coverage (added by regex tightening) ──────

test_chmod_o_w_denied if {
	some msg in deny with input as [
		{"Cmd": "run", "Stage": 0, "Value": ["chmod o+w /tmp"]},
	]
	contains(msg, "world-writable chmod")
}

test_chmod_og_w_denied if {
	some msg in deny with input as [
		{"Cmd": "run", "Stage": 0, "Value": ["chmod og+w /opt/app"]},
	]
	contains(msg, "world-writable chmod")
}

test_chmod_recursive_o_w_denied if {
	some msg in deny with input as [
		{"Cmd": "run", "Stage": 0, "Value": ["chmod -R o+w /var/cache"]},
	]
	contains(msg, "world-writable chmod")
}

# Negative cases — symbolic forms that do NOT grant world-write

test_chmod_g_w_not_denied if {
	msgs := deny with input as [
		{"Cmd": "run", "Stage": 0, "Value": ["chmod g+w /opt/shared"]},
	]
	every msg in msgs {
		not contains(msg, "world-writable chmod")
	}
}

test_chmod_ug_w_not_denied if {
	msgs := deny with input as [
		{"Cmd": "run", "Stage": 0, "Value": ["chmod ug+w /opt/shared"]},
	]
	every msg in msgs {
		not contains(msg, "world-writable chmod")
	}
}

test_chmod_o_r_only_not_denied if {
	msgs := deny with input as [
		{"Cmd": "run", "Stage": 0, "Value": ["chmod o+r /etc/public.conf"]},
	]
	every msg in msgs {
		not contains(msg, "world-writable chmod")
	}
}

test_chmod_o_minus_w_not_denied if {
	msgs := deny with input as [
		{"Cmd": "run", "Stage": 0, "Value": ["chmod o-w /etc/sensitive"]},
	]
	every msg in msgs {
		not contains(msg, "world-writable chmod")
	}
}

# ── set-exactly (=) form coverage ────────────────────────────────────

test_chmod_a_eq_rwx_denied if {
	some msg in deny with input as [
		{"Cmd": "run", "Stage": 0, "Value": ["chmod a=rwx /tmp"]},
	]
	contains(msg, "world-writable chmod")
}

test_chmod_ugo_eq_rwx_denied if {
	some msg in deny with input as [
		{"Cmd": "run", "Stage": 0, "Value": ["chmod ugo=rwx /tmp"]},
	]
	contains(msg, "world-writable chmod")
}

test_chmod_o_eq_w_denied if {
	some msg in deny with input as [
		{"Cmd": "run", "Stage": 0, "Value": ["chmod o=w /tmp"]},
	]
	contains(msg, "world-writable chmod")
}

test_chmod_og_eq_w_denied if {
	some msg in deny with input as [
		{"Cmd": "run", "Stage": 0, "Value": ["chmod og=w /opt/app"]},
	]
	contains(msg, "world-writable chmod")
}
