package main

import rego.v1

remote_url_schemes := {"http://", "https://", "ftp://"}

is_remote_url(src) if {
	some scheme in remote_url_schemes
	startswith(src, scheme)
}

deny contains msg if {
	some instr in input
	instr.Cmd == "add"
	src := instr.Value[0]
	is_remote_url(src)
	msg := sprintf("ADD %q fetches from a remote URL — supply chain risk, use COPY with a verified local file or RUN curl with sha256 verification", [src])
}
