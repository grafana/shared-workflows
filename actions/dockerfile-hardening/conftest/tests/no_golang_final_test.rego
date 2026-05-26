package main

import rego.v1

test_golang_final_after_golang_builder_denied if {
	some msg in deny with input as [
		{"Cmd": "from", "Stage": 0, "Value": ["golang:1.22@sha256:abc", "as", "builder"]},
		{"Cmd": "from", "Stage": 1, "Value": ["golang:1.22@sha256:abc"]},
	]
	contains(msg, "ships the Go toolchain")
}

test_golang_digest_only_final_denied if {
	some msg in deny with input as [
		{"Cmd": "from", "Stage": 0, "Value": ["golang:1.22@sha256:abc", "as", "builder"]},
		{"Cmd": "from", "Stage": 1, "Value": ["golang@sha256:abc"]},
	]
	contains(msg, "ships the Go toolchain")
}

test_distroless_final_after_golang_builder_not_denied if {
	msgs := deny with input as [
		{"Cmd": "from", "Stage": 0, "Value": ["golang:1.22@sha256:abc", "as", "builder"]},
		{"Cmd": "from", "Stage": 1, "Value": ["gcr.io/distroless/static@sha256:def"]},
	]
	every msg in msgs {
		not contains(msg, "ships the Go toolchain")
	}
}

test_scratch_final_after_golang_builder_not_denied if {
	msgs := deny with input as [
		{"Cmd": "from", "Stage": 0, "Value": ["golang:1.22@sha256:abc", "as", "builder"]},
		{"Cmd": "from", "Stage": 1, "Value": ["scratch"]},
	]
	every msg in msgs {
		not contains(msg, "ships the Go toolchain")
	}
}

test_no_golang_anywhere_not_denied if {
	msgs := deny with input as [
		{"Cmd": "from", "Stage": 0, "Value": ["alpine:3.18@sha256:abc"]},
	]
	every msg in msgs {
		not contains(msg, "ships the Go toolchain")
	}
}

test_golang_only_in_builder_with_non_golang_final_not_denied if {
	msgs := deny with input as [
		{"Cmd": "from", "Stage": 0, "Value": ["golang:1.22@sha256:abc", "as", "builder"]},
		{"Cmd": "from", "Stage": 1, "Value": ["cgr.dev/chainguard/static@sha256:def"]},
	]
	every msg in msgs {
		not contains(msg, "ships the Go toolchain")
	}
}
