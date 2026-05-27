package main

import rego.v1

# Positive cases — rule fires

test_curl_pipe_bash_denied if {
	some msg in deny with input as [
		{"Cmd": "run", "Stage": 0, "Value": ["curl https://get.docker.com | bash"]},
	]
	contains(msg, "pipes remote content directly into a shell")
}

test_curl_pipe_sh_denied if {
	some msg in deny with input as [
		{"Cmd": "run", "Stage": 0, "Value": ["curl -fsSL https://example.com/install.sh | sh"]},
	]
	contains(msg, "pipes remote content directly into a shell")
}

test_wget_pipe_bash_denied if {
	some msg in deny with input as [
		{"Cmd": "run", "Stage": 0, "Value": ["wget -qO- https://nodesource.com/setup_20.x | bash"]},
	]
	contains(msg, "pipes remote content directly into a shell")
}

test_curl_pipe_zsh_denied if {
	some msg in deny with input as [
		{"Cmd": "run", "Stage": 0, "Value": ["curl https://example.com/install.zsh | zsh"]},
	]
	contains(msg, "pipes remote content directly into a shell")
}

test_curl_pipe_bash_with_args_denied if {
	some msg in deny with input as [
		{"Cmd": "run", "Stage": 0, "Value": ["curl https://rustup.rs -sSf | sh -s -- -y"]},
	]
	contains(msg, "pipes remote content directly into a shell")
}

test_curl_pipe_bash_no_space_denied if {
	some msg in deny with input as [
		{"Cmd": "run", "Stage": 0, "Value": ["curl https://example.com/x |bash"]},
	]
	contains(msg, "pipes remote content directly into a shell")
}

test_sudo_curl_pipe_bash_denied if {
	some msg in deny with input as [
		{"Cmd": "run", "Stage": 0, "Value": ["sudo curl https://example.com/x | bash"]},
	]
	contains(msg, "pipes remote content directly into a shell")
}

test_curl_pipe_in_later_stage_reports_stage if {
	some msg in deny with input as [
		{"Cmd": "from", "Stage": 0, "Value": ["alpine@sha256:abc"]},
		{"Cmd": "from", "Stage": 1, "Value": ["alpine@sha256:def"]},
		{"Cmd": "run", "Stage": 1, "Value": ["curl https://example.com/install | sh"]},
	]
	contains(msg, "RUN in stage 1")
}

# Negative cases — rule does not fire

test_curl_to_file_not_denied if {
	msgs := deny with input as [
		{"Cmd": "run", "Stage": 0, "Value": ["curl -fsSL https://example.com/file -o /tmp/file"]},
	]
	every msg in msgs {
		not contains(msg, "pipes remote content directly into a shell")
	}
}

test_curl_pipe_jq_not_denied if {
	msgs := deny with input as [
		{"Cmd": "run", "Stage": 0, "Value": ["curl https://api.example.com/data | jq '.field'"]},
	]
	every msg in msgs {
		not contains(msg, "pipes remote content directly into a shell")
	}
}

test_curl_pipe_tee_not_denied if {
	msgs := deny with input as [
		{"Cmd": "run", "Stage": 0, "Value": ["curl https://example.com/x | tee /tmp/x"]},
	]
	every msg in msgs {
		not contains(msg, "pipes remote content directly into a shell")
	}
}

test_word_boundary_avoids_false_positive_on_shellcheck if {
	msgs := deny with input as [
		{"Cmd": "run", "Stage": 0, "Value": ["curl https://example.com/file | shellcheck"]},
	]
	every msg in msgs {
		not contains(msg, "pipes remote content directly into a shell")
	}
}

test_word_boundary_avoids_false_positive_on_mycurl if {
	msgs := deny with input as [
		{"Cmd": "run", "Stage": 0, "Value": ["mycurlwrapper --verify | bash"]},
	]
	every msg in msgs {
		not contains(msg, "pipes remote content directly into a shell")
	}
}

test_no_run_instructions_not_denied if {
	msgs := deny with input as [
		{"Cmd": "from", "Stage": 0, "Value": ["scratch"]},
		{"Cmd": "entrypoint", "Stage": 0, "Value": ["/app"]},
	]
	every msg in msgs {
		not contains(msg, "pipes remote content directly into a shell")
	}
}

test_plain_curl_no_pipe_not_denied if {
	msgs := deny with input as [
		{"Cmd": "run", "Stage": 0, "Value": ["curl --version"]},
	]
	every msg in msgs {
		not contains(msg, "pipes remote content directly into a shell")
	}
}
