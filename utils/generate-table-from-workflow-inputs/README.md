# Generate Markdown Table from Workflow Inputs

This lightweight script can help generate a simple markdown table from a reusable GitHub workflow yaml.
Then, you can copy+paste this markdown table into a doc/readme.

## Prerequisites

- Python 3.x

## Setup

1. Set up virtual environment (using `venv`)

   ```bash
   python3 -m venv env
   source env/bin/activate
   ```

1. Install Dependencies

   ```bash
   pip install -r requirements.txt
   ```

## Usage

Run the command:

```bash
./generate-table-from-inputs.py <path-to-workflow-yaml-file>
```

Take the output and add to your markdown doc.
Output should look something like this:

```console
| Name                             | Type    | Description                                                                                                                                                            |
| -------------------------------- | ------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `namespace`                      | string  | The entity's namespace within EngHub (usually `default`)                                                                                                               |
| `kind`                           | string  | The kind of the entity in EngHub (usually `component`)                                                                                                                 |
| `name`                           | string  | The name of the entity in EngHub (usually matches the name of the repository)                                                                                          |
| `default-working-directory`      | string  | The working directory to use for doc generation. Useful for cases without an mkdocs.yml file at the project root.                                                      |
| `rewrite-relative-links`         | boolean | Execute rewrite-relative-links step to rewrite relative links in the docs to point to the correct location in the GitHub repository                                    |
| `rewrite-relative-links-dry-run` | boolean | Execute rewrite-relative-links step but only print the diff without modifying the files                                                                                |
| `publish`                        | boolean | Enable or disable publishing after building the docs                                                                                                                   |
| `checkout-submodules`            | string  | Checkout submodules in the repository. Options are `true` (checkout submodules), `false` (don't checkout submodules), or `recursive` (recursively checkout submodules) |
| `bucket`                         | string  | The name of the GCS bucket to which the docs will be published                                                                                                         |
| `workload-identity-provider`     | string  | The GCP workload identity provider to use for authentication                                                                                                           |
| `service-account`                | string  | The GCP service account to use for authentication                                                                                                                      |
```
