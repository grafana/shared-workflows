package main

import rego.v1

test_alpine_base_denied if {
	some msg in deny with input as [
		{"Cmd": "from", "Stage": 0, "Value": ["alpine:3.18@sha256:abc"]},
	]
	contains(msg, "must be either scratch or a distroless base")
}

test_scratch_base_not_denied if {
	msgs := deny with input as [
		{"Cmd": "from", "Stage": 0, "Value": ["scratch"]},
	]
	every msg in msgs {
		not contains(msg, "must be either scratch or a distroless")
	}
}

test_gcr_distroless_not_denied if {
	msgs := deny with input as [
		{"Cmd": "from", "Stage": 0, "Value": ["gcr.io/distroless/static-debian12@sha256:abc"]},
	]
	every msg in msgs {
		not contains(msg, "must be either scratch or a distroless")
	}
}

test_chainguard_static_not_denied if {
	msgs := deny with input as [
		{"Cmd": "from", "Stage": 0, "Value": ["cgr.dev/chainguard/static@sha256:abc"]},
	]
	every msg in msgs {
		not contains(msg, "must be either scratch or a distroless")
	}
}

test_chainguard_private_not_denied if {
	msgs := deny with input as [
		{"Cmd": "from", "Stage": 0, "Value": ["cgr.dev/chainguard-private/foo@sha256:abc"]},
	]
	every msg in msgs {
		not contains(msg, "must be either scratch or a distroless")
	}
}

test_path_distroless_not_denied if {
	msgs := deny with input as [
		{"Cmd": "from", "Stage": 0, "Value": ["example.com/myorg/distroless/base@sha256:abc"]},
	]
	every msg in msgs {
		not contains(msg, "must be either scratch or a distroless")
	}
}

test_only_final_stage_evaluated if {
	msgs := deny with input as [
		{"Cmd": "from", "Stage": 0, "Value": ["golang:1.22@sha256:abc"]},
		{"Cmd": "from", "Stage": 1, "Value": ["scratch"]},
	]
	every msg in msgs {
		not contains(msg, "must be either scratch or a distroless")
	}
}

test_non_distroless_final_after_distroless_builder_denied if {
	some msg in deny with input as [
		{"Cmd": "from", "Stage": 0, "Value": ["gcr.io/distroless/static@sha256:abc"]},
		{"Cmd": "from", "Stage": 1, "Value": ["alpine:3.18@sha256:def"]},
	]
	contains(msg, "must be either scratch or a distroless base")
}
