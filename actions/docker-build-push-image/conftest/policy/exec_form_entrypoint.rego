package main

import rego.v1

deny contains msg if {
	some instr in input
	instr.Cmd in {"entrypoint", "cmd"}
	not instr.JSON
	msg := sprintf("%s uses shell form — must be exec form (JSON array) so the binary becomes PID 1 and receives signals", [upper(instr.Cmd)])
}
