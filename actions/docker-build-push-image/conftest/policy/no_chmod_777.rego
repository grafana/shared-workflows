package main

import rego.v1

deny contains msg if {
	some instr in input
	instr.Cmd == "run"
	cmd := instr.Value[0]
	regex.match(`chmod\s+(-R\s+)?(0?777|a[+=]rwx|[ugoa]*o[ugoa]*[+=][rwxst]*w[rwxst]*)`, cmd)
	msg := sprintf("RUN in stage %d uses world-writable chmod (777 or equivalent)", [instr.Stage])
}
