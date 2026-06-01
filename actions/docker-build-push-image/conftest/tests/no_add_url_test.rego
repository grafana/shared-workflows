package main

import rego.v1

test_add_https_url_denied if {
	some msg in deny with input as [
		{"Cmd": "add", "Stage": 0, "Flags": [], "Value": ["https://example.com/installer.tar.gz", "/app/installer.tar.gz"]},
	]
	contains(msg, "fetches from a remote URL")
}

test_add_http_url_denied if {
	some msg in deny with input as [
		{"Cmd": "add", "Stage": 0, "Flags": [], "Value": ["http://example.com/installer.tar.gz", "/app/installer.tar.gz"]},
	]
	contains(msg, "fetches from a remote URL")
}

test_add_ftp_url_denied if {
	some msg in deny with input as [
		{"Cmd": "add", "Stage": 0, "Flags": [], "Value": ["ftp://example.com/file.tar.gz", "/app/file.tar.gz"]},
	]
	contains(msg, "fetches from a remote URL")
}

test_add_with_chown_flag_url_denied if {
	some msg in deny with input as [
		{"Cmd": "add", "Stage": 0, "Flags": ["--chown=1000:1000"], "Value": ["https://example.com/file", "/app/file"]},
	]
	contains(msg, "fetches from a remote URL")
}

test_add_local_file_not_denied if {
	msgs := deny with input as [
		{"Cmd": "add", "Stage": 0, "Flags": [], "Value": ["./local-archive.tar.gz", "/app/"]},
	]
	every msg in msgs {
		not contains(msg, "fetches from a remote URL")
	}
}

test_add_local_tarball_not_denied if {
	msgs := deny with input as [
		{"Cmd": "add", "Stage": 0, "Flags": [], "Value": ["build/dist.tar.gz", "/opt/"]},
	]
	every msg in msgs {
		not contains(msg, "fetches from a remote URL")
	}
}

test_copy_from_url_not_denied if {
	msgs := deny with input as [
		{"Cmd": "copy", "Stage": 0, "Flags": [], "Value": ["./https-named-file", "/app/"]},
	]
	every msg in msgs {
		not contains(msg, "fetches from a remote URL")
	}
}

test_no_add_instructions_not_denied if {
	msgs := deny with input as [
		{"Cmd": "from", "Stage": 0, "Value": ["scratch"]},
		{"Cmd": "copy", "Stage": 0, "Flags": [], "Value": ["./app", "/app"]},
	]
	every msg in msgs {
		not contains(msg, "fetches from a remote URL")
	}
}

test_add_filename_containing_http_substring_not_denied if {
	msgs := deny with input as [
		{"Cmd": "add", "Stage": 0, "Flags": [], "Value": ["./https-handler.tar.gz", "/app/"]},
	]
	every msg in msgs {
		not contains(msg, "fetches from a remote URL")
	}
}
