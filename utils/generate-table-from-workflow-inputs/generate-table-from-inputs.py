#!/usr/bin/env python3
"""
This script generates a markdown table from a GitHub Actions reusable workflow inputs.
You can use this generated markdown table in your doc to show the available inputs for your workflow.

Usage:
    python markdown-table-from-workflow-inputs.py <path-to-workflow-yaml-file>
"""

import sys
import yaml
from pathlib import Path
from pprint import pprint


def get_longest_length(arr_string: list[str]) -> int:
    return max([len(s) for s in arr_string])


def main(workflow_file_path: Path):
    with open(workflow_file, "r") as f:
        workflow = yaml.safe_load(f)

    try:
        inputs = workflow[True]["workflow_call"]["inputs"]
    except KeyError:
        print("Could not find the inputs in the workflow file. Please ensure the workflow is a reusable one and has inputs.")
        sys.exit(1)

    # adding 2 for the backticks
    name_column_padding = get_longest_length(list(inputs.keys())) + 2
    description_column_padding = get_longest_length([value.get("description", "") for value in inputs.values()])
    type_column_padding = get_longest_length([str(value.get("type", "")) for value in inputs.values()])

    print(f"| {'Name'.ljust(name_column_padding)} | {'Type'.ljust(type_column_padding)} | {'Description'.ljust(description_column_padding)} |")
    print(f"| {'-' * name_column_padding} | {'-' * type_column_padding} | {'-' * description_column_padding} |")

    for name, value in inputs.items():
        description = value.get("description", "")
        the_type = str(value.get("type", ""))
        formatted_name = f'`{name}`'.ljust(name_column_padding)
        formatted_type = the_type.ljust(type_column_padding)
        formatted_description = description.ljust(description_column_padding)
        print(f"| {formatted_name} | {formatted_type} | {formatted_description} |")

if __name__ == "__main__":
    if len(sys.argv) != 2:
        print(__doc__)
        sys.exit(1)
    workflow_file = sys.argv[1]
    main(Path(workflow_file))
