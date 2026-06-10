package main

import rego.v1

has_from_flag(instr) if {
	some flag in instr.Flags
	startswith(flag, "--from=")
}

warn contains msg if {
	some i, instr in input
	instr.Cmd == "copy"
	i > last_from_index
	not has_from_flag(instr)
	instr.Value[0] == "."
	msg := "COPY . in final stage may ship source tree to runtime image — confirm the build context only contains intended artifacts"
}
