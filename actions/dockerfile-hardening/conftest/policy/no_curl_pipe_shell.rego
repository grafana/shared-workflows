package main

import rego.v1

# Matches `curl` or `wget` whose output is piped directly into a shell
# (bash/sh/zsh). Catches the classic remote-code-execution anti-pattern
# `curl ... | bash` where remote content executes at build time with no
# digest verification.
deny contains msg if {
	some instr in input
	instr.Cmd == "run"
	regex.match(`\b(curl|wget)\b[^|]*\|\s*(bash|sh|zsh)\b`, instr.Value[0])
	msg := sprintf("RUN in stage %d pipes remote content directly into a shell (curl|bash or equivalent) — remote code execution with no verification", [instr.Stage])
}
