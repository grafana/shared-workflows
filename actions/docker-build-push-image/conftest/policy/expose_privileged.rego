package main

import rego.v1

deny contains msg if {
	some instr in input
	instr.Cmd == "expose"
	some value in instr.Value
	port_str := split(value, "/")[0]
	port := to_number(port_str)
	port > 0
	port < 1024
	msg := sprintf("EXPOSE %q uses privileged port (<1024)", [value])
}
