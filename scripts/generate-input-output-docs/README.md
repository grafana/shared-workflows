# generate-input-output-docs

Keeps the input and output tables in this repo's action READMEs and reusable
workflow docs in sync with the YAML that declares them.

Before this existed, every table was maintained by hand, and most of them had
drifted: inputs missing entirely, outputs undocumented, and at least one default
value that no longer matched the action. See [#137] and [#1466].

[#137]: https://github.com/grafana/shared-workflows/issues/137
[#1466]: https://github.com/grafana/shared-workflows/issues/1466

## Usage

`generate`, `check` and `print` need the pinned prettier, so install dependencies
once first (see [Prettier](#prettier)):

```sh
bun install
```

All commands run from this directory:

```sh
cd scripts/generate-input-output-docs
```

Regenerate every doc in the repo:

```sh
go run . generate -root-dir ../../
```

Check for drift without writing anything:

```sh
go run . check -root-dir ../../
```

Check that a reusable workflow still forwards the inputs and outputs of the
composite action it wraps:

```sh
go run . parity -root-dir ../../
```

Print the tables for a single file, without touching any doc:

```sh
go run . print -file ../../actions/aws-auth/action.yaml
```

[`check-action-docs.yaml`](../../.github/workflows/check-action-docs.yaml) runs
`generate` and then fails on `git diff`, runs `parity`, and runs the tests.
`check` is the local equivalent of that first job for when you would rather not
have your working tree written to.

## How docs are updated

Generated tables live between HTML comment markers:

```markdown
## Inputs

<!-- BEGIN_INPUTS -->

| Name  | Type   | Required | Default | Description |
| ----- | ------ | -------- | ------- | ----------- |
| `foo` | String | No       | `bar`   | Does a foo. |

<!-- END_INPUTS -->

Hand-written prose here is left alone.
```

Only the text between the markers is replaced, so notes, footnotes and
`markdownlint` pragmas around a table survive regeneration. `BEGIN_OUTPUTS` /
`END_OUTPUTS` work the same way.

If a doc has a `## Inputs` heading but no markers, the generator inserts them and
drops the table that directly followed. If it has neither, a new section is
appended. Anything that is not a plain table — a bullet list, say — is left in
place for a human to remove deliberately, rather than being silently deleted.

## What gets discovered

- Composite actions at `actions/<name>/action.yml` or `action.yaml`, documented
  in `actions/<name>/README.md`.
- Reusable workflows in `.github/workflows` that declare an `on.workflow_call`
  trigger, documented in a sibling `.md` file of the same name. Ordinary
  workflows are skipped.

## Columns

Inputs render as Name, Type, Required, Default and Description. `Required` and
`Default` come straight from the YAML rather than from prose, which is the whole
point: a default documented in a sentence can go stale, a generated one cannot.

Outputs render as Name and Description only. Neither composite actions nor
reusable workflows give an output a type, a default or a required flag, so those
columns would carry nothing but filler.

### Type inference

Reusable workflow inputs declare a `type`, and it is used verbatim — GitHub
spells these lowercase (`string`, `boolean`, `number`).

Composite action inputs have no type field in GitHub's schema, so the type is
inferred from the default and rendered capitalized (`String`, `Boolean`) to match
the convention the READMEs already used. **A boolean input with no default is
reported as `String`**, because nothing in the YAML distinguishes it from a
string. Give such an input an explicit `default: false` if the type matters to
readers.

## Prettier

The generator shells out to prettier and fails if it cannot find it. This is not
cosmetic: prettier runs as a pre-commit hook over every markdown file in the
repo, and it rewrites table _content_ as well as padding — collapsing a
` ```x``` ` code fence to `` `x` ``, for instance. If the generator emitted
markdown that prettier disagreed with, the hook and the generator would take
turns rewriting each other and the CI drift check would never pass.

### You need the pinned version

**Run `bun install` before `generate`.** The generator requires the exact prettier
version pinned in `package.json` and refuses anything else:

```
error: prettier 3.9.6 is required (pinned in package.json) but only found
prettier (3.8.3); run `bun install`, or pass -prettier to override
```

That is deliberate, not fussiness. Prettier's markdown output changes between
releases — 3.8.3 renders an asterisk in a table cell as `\*.out` and 3.9.6 renders
it as `*.out`. Generating with a different version than CI uses produces a
one-byte difference, and the drift check then fails on a PR whose docs were just
regenerated. Requiring the pin makes local output identical to CI by
construction.

Candidates are tried in order — `node_modules/.bin/prettier`, `bunx prettier`,
`npx prettier`, then a global `prettier` — and each is run to read its actual
version, not merely looked up on `PATH`. The launcher existing says nothing about
prettier being reachable through it: a machine with Node but no `node_modules`
has `npx`, and `npx --no-install prettier` fails there.

`-prettier "<command>"` overrides resolution entirely and skips the version
check, for when you want to point at something specific.

## Parity rules

`parity.yaml` declares reusable workflows that are expected to expose the same
inputs and outputs as a composite action they wrap — currently
`docker-build-push-multiarch` and `docker-build-push-image`, which was the
original reason for writing this tool.

Names that genuinely belong to only one side are listed in the rule with a
reason. Listing a name that no longer exists is itself reported as drift, so the
exceptions cannot quietly rot.

## Tests

```sh
cd scripts/generate-input-output-docs
go test ./...
```

Three tests need the pinned prettier and skip themselves without it, so `go test`
works before `bun install`. Two assert that generated markdown survives prettier
unchanged — the tool's most important invariant — and the third asserts that the
resolved prettier really is the pinned version. CI installs dependencies before
running the suite so those three are never silently skipped there.
