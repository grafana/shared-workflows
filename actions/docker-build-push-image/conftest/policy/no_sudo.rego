package main

import rego.v1

# Regex matches `sudo` as a standalone shell command: bounded by
# start/end-of-string, whitespace, or shell metacharacters
# (; & | ( ) `).
deny contains msg if {
	some instr in input
	instr.Cmd == "run"
	cmd := instr.Value[0]
	regex.match("(^|[\\s;&|()`])sudo(\\s|$|[;&|()`])", cmd)
	msg := sprintf("RUN in stage %d invokes sudo", [instr.Stage])
}
