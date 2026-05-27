package main

import rego.v1

from_instructions := [instr |
	some instr in input
	instr.Cmd == "from"
]

from_indices := [i |
	some i, instr in input
	instr.Cmd == "from"
]

final_from := from_instructions[count(from_instructions) - 1] if {
	count(from_instructions) > 0
}

last_from_index := max(from_indices) if {
	count(from_indices) > 0
}

stage_aliases contains alias if {
	some instr in from_instructions
	count(instr.Value) >= 3
	lower(instr.Value[1]) == "as"
	alias := instr.Value[2]
}

is_stage_alias(ref) if {
	ref in stage_aliases
}

image_ref(instr) := instr.Value[0]

strip_alias(ref) := split(ref, " ")[0]

user_indices := [i |
	some i, instr in input
	instr.Cmd == "user"
	i > last_from_index
]

final_user := input[i] if {
	count(user_indices) > 0
	i := max(user_indices)
}
