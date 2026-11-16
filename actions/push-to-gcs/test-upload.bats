#!/usr/bin/env bats

export BATS_TEST_TIMEOUT=30

setup() {
  script="$(awk '
    /id: upload-files-gcloud/ { step = 1 }
    step && /^      run: \|/ { body = 1; next }
    body && /^    - / { exit }
    body { sub(/^        /, ""); print }
  ' "$BATS_TEST_DIRNAME/action.yaml")"
  workspace="$BATS_TEST_TMPDIR/workspace"
  mkdir -p "$workspace/dist/nested" "$workspace/dist/.hidden"
  printf artifact > "$workspace/dist/package.tar.gz"
  printf checksum > "$workspace/dist/nested/package.sha256"
  printf hidden > "$workspace/dist/.hidden/file.txt"
  printf artifact > "$workspace/dist/package with spaces.tar.gz"
  export -f gcloud
}

gcloud() {
  local source="${*: -2:1}" target="${*: -1}" IFS=
  local -a sources peers
  shopt -s nullglob dotglob
  # Match the CLI's wildcard expansion, including for quoted arguments.
  # shellcheck disable=SC2206
  sources=($source)
  ((${#sources[@]} == 1)) || return 64
  printf '%s\n' "$@" >> "$COMMAND_LOG"
  printf '%s\n' "${sources[0]}" >> "$SOURCE_LOG"
  if [[ "$target" == */ || "$target" == "${DIRECTORY_DESTINATION:-}" ]]; then
    target="${target%/}/$(basename -- "${sources[0]}")"
  fi
  printf '%s\n' "$target" >> "$DESTINATION_LOG"
  if [[ -n "${BARRIER_DIR:-}" ]]; then
    : > "$BARRIER_DIR/$BASHPID"
    while :; do
      peers=("$BARRIER_DIR"/*)
      ((${#peers[@]} > 1)) && break
      sleep 0.01
    done
  fi
  if [[ -n "${FAIL_BASENAME:-}" && "${sources[0]##*/}" != "$FAIL_BASENAME" ]]; then
    return 0
  fi
  return "${FAIL_UPLOAD:-0}"
}

run_upload() {
  local name="$1"
  shift
  env GITHUB_WORKSPACE="$workspace" GITHUB_OUTPUT="$BATS_TEST_TMPDIR/$name.output" RUNNER_TEMP="$BATS_TEST_TMPDIR" \
    SOURCE_PATH=dist SOURCE_GLOB= DESTINATION=bucket/releases/ \
    INCLUDE_PARENT=false GZIP=false PREDEFINED_ACL=projectPrivate \
    COMMAND_LOG="$BATS_TEST_TMPDIR/$name.commands" DESTINATION_LOG="$BATS_TEST_TMPDIR/$name.destinations" SOURCE_LOG="$BATS_TEST_TMPDIR/$name.sources" \
    "$@" bash -e -o pipefail -c "$script"
}

@test "directory uploads retain nested and hidden paths and export object names" {
  run_upload directory
  [[ "$(LC_ALL=C sort "$BATS_TEST_TMPDIR/directory.destinations")" == $'gs://bucket/releases/.hidden/file.txt\ngs://bucket/releases/nested/package.sha256\ngs://bucket/releases/package with spaces.tar.gz\ngs://bucket/releases/package.tar.gz' ]]
  [[ "$(< "$BATS_TEST_TMPDIR/directory.output")" == *'releases/.hidden/file.txt,releases/nested/package.sha256,releases/package with spaces.tar.gz,releases/package.tar.gz'* ]]
  [[ "$(< "$BATS_TEST_TMPDIR/directory.commands")" == *'--predefined-acl=projectPrivate'* ]]
  [[ "$(< "$BATS_TEST_TMPDIR/directory.commands")" != *'--gzip'* ]]
}

@test "parent includes the source directory and accepts trailing slashes" {
  run_upload parent INCLUDE_PARENT=true SOURCE_PATH=dist/
  [[ "$(< "$BATS_TEST_TMPDIR/parent.destinations")" == *'gs://bucket/releases/dist/nested/package.sha256'* ]]
}

@test "single file uploads omit the parent directory for either parent setting" {
  for parent in true false; do
    run_upload "file-$parent" SOURCE_PATH=dist/package.tar.gz "INCLUDE_PARENT=$parent"
    [[ "$(< "$BATS_TEST_TMPDIR/file-$parent.destinations")" == 'gs://bucket/releases/package.tar.gz' ]]
  done
}

@test "relative globs filter files and retain directory paths" {
  run_upload glob SOURCE_GLOB='nested/*.sha256' INCLUDE_PARENT=true
  [[ "$(< "$BATS_TEST_TMPDIR/glob.destinations")" == 'gs://bucket/releases/dist/nested/package.sha256' ]]
}

@test "extglobs select matching files including hidden directories" {
  run_upload extglob SOURCE_GLOB='**/@(package.sha256|file.txt)'
  [[ "$(LC_ALL=C sort "$BATS_TEST_TMPDIR/extglob.destinations")" == $'gs://bucket/releases/.hidden/file.txt\ngs://bucket/releases/nested/package.sha256' ]]
}

@test "workspace source and default bucket path do not add dot components" {
  run_upload root SOURCE_PATH=. INCLUDE_PARENT=true DESTINATION=bucket/.
  [[ "$(< "$BATS_TEST_TMPDIR/root.destinations")" == *'gs://bucket/dist/package.tar.gz'* ]]
  [[ "$(< "$BATS_TEST_TMPDIR/root.output")" == *'dist/package.tar.gz'* ]]
}

@test "absolute source paths preserve filenames with spaces" {
  run_upload absolute "SOURCE_PATH=$workspace/dist" SOURCE_GLOB='*.tar.gz'
  [[ "$(< "$BATS_TEST_TMPDIR/absolute.destinations")" == *'gs://bucket/releases/package with spaces.tar.gz'* ]]
}

@test "gzip is enabled on request and blank ACLs are omitted" {
  run_upload gzip SOURCE_GLOB=package.tar.gz GZIP=true PREDEFINED_ACL=' '
  [[ "$(< "$BATS_TEST_TMPDIR/gzip.commands")" == *'--gzip-local-all'* ]]
  [[ "$(< "$BATS_TEST_TMPDIR/gzip.commands")" != *'--predefined-acl'* ]]
}

@test "shell syntax in filenames is passed literally" {
  filename="package \$(command).tar.gz"
  printf artifact > "$workspace/dist/$filename"
  run_upload literal "SOURCE_PATH=dist/$filename"
  [[ "$(< "$BATS_TEST_TMPDIR/literal.destinations")" == "gs://bucket/releases/$filename" ]]
}

@test "unmatched globs warn without invoking gcloud" {
  run run_upload empty SOURCE_GLOB='*.missing'
  [ "$status" -eq 0 ]
  [[ ! -f "$BATS_TEST_TMPDIR/empty.commands" ]]
  [[ "$output" == *'::warning::No files matched'* ]]
}

@test "gcloud failures stop the upload and preserve its exit status" {
  run run_upload failure FAIL_UPLOAD=23
  [ "$status" -eq 23 ]
  [[ ! -f "$BATS_TEST_TMPDIR/failure.output" ]]
  [ "$(wc -l < "$BATS_TEST_TMPDIR/failure.destinations")" -le 4 ]
}

@test "missing source paths fail before invoking gcloud" {
  run run_upload missing SOURCE_PATH=missing
  [ "$status" -ne 0 ]
  [[ ! -f "$BATS_TEST_TMPDIR/missing.commands" && ! -f "$BATS_TEST_TMPDIR/missing.output" ]]
}

@test "a glob combined with a file source fails before invoking gcloud" {
  run run_upload file-glob SOURCE_PATH=dist/package.tar.gz SOURCE_GLOB='*'
  [ "$status" -ne 0 ]
  [[ "$output" == *'The glob input cannot be used with a file path'* ]]
  [[ ! -f "$BATS_TEST_TMPDIR/file-glob.commands" && ! -f "$BATS_TEST_TMPDIR/file-glob.output" ]]
}

@test "brace patterns fail before invoking gcloud" {
  run run_upload brace SOURCE_GLOB='**/*.{gz,sha256}'
  [ "$status" -ne 0 ]
  [[ "$output" == *'Brace expansion is unsupported'* ]]
  [[ ! -f "$BATS_TEST_TMPDIR/brace.commands" && ! -f "$BATS_TEST_TMPDIR/brace.output" ]]
}

@test "absolute globs produce relative object paths and retain nonrecursive matching" {
  printf nested > "$workspace/dist/nested/extra.tar.gz"
  run_upload absolute-glob "SOURCE_GLOB=$workspace/dist/*.tar.gz"
  [[ "$(LC_ALL=C sort "$BATS_TEST_TMPDIR/absolute-glob.destinations")" == $'gs://bucket/releases/package with spaces.tar.gz\ngs://bucket/releases/package.tar.gz' ]]
  [[ "$(< "$BATS_TEST_TMPDIR/absolute-glob.output")" != *"$workspace"* ]]
}

@test "recursive globs follow nested directory symlinks" {
  mkdir -p "$workspace/shared/deep"
  printf shared > "$workspace/shared/deep/asset.js"
  ln -s ../shared "$workspace/dist/vendor"
  run_upload symlink SOURCE_GLOB='**/*.js'
  [[ "$(< "$BATS_TEST_TMPDIR/symlink.destinations")" == 'gs://bucket/releases/vendor/deep/asset.js' ]]
  [[ "$(< "$BATS_TEST_TMPDIR/symlink.output")" == *'releases/vendor/deep/asset.js'* ]]
}

@test "symlink cycles terminate without repeated uploads" {
  mkdir -p "$workspace/shared/deep"
  printf shared > "$workspace/shared/deep/asset.js"
  ln -s ../shared "$workspace/dist/vendor"
  ln -s .. "$workspace/dist/nested/back"
  run run_upload cycle
  if [ "$status" -eq 0 ]; then
    [ "$(wc -l < "$BATS_TEST_TMPDIR/cycle.destinations")" -eq 5 ]
    [[ "$(< "$BATS_TEST_TMPDIR/cycle.destinations")" == *"vendor/deep/asset.js"* ]]
    [[ "$(< "$BATS_TEST_TMPDIR/cycle.output")" != *"nested/back/"* ]]
  else
    [ "$status" -eq 1 ]
    [[ "$output" == *loop* || "$output" == *cycle* ]]
    [[ ! -f "$BATS_TEST_TMPDIR/cycle.commands" && ! -f "$BATS_TEST_TMPDIR/cycle.output" ]]
  fi
}

@test "existing object-name prefixes do not change the uploaded object name" {
  run_upload prefix SOURCE_PATH=dist/package.tar.gz DIRECTORY_DESTINATION=gs://bucket/releases/package.tar.gz
  [[ "$(< "$BATS_TEST_TMPDIR/prefix.destinations")" == 'gs://bucket/releases/package.tar.gz' ]]
  [[ "$(< "$BATS_TEST_TMPDIR/prefix.commands")" == *$'\ngs://bucket/releases/' ]]
}

@test "wildcard characters in selected filenames are passed literally to gcloud" {
  mkdir -p "$workspace/app/[slug]" "$workspace/app/s"
  printf intended > "$workspace/app/[slug]/page.js"
  printf decoy > "$workspace/app/s/page.js"
  for filename in 'bundle*.js' 'bundle?.js' 'closing].js'; do
    printf artifact > "$workspace/app/$filename"
  done
  run_upload wildcard SOURCE_PATH=app
  [[ "$(< "$BATS_TEST_TMPDIR/wildcard.sources")" == *'./[slug]/page.js'* ]]
  [[ "$(< "$BATS_TEST_TMPDIR/wildcard.sources")" == *'./bundle*.js'* ]]
  [[ "$(< "$BATS_TEST_TMPDIR/wildcard.sources")" == *'./bundle?.js'* ]]
  [[ "$(< "$BATS_TEST_TMPDIR/wildcard.destinations")" == *'gs://bucket/releases/[slug]/page.js'* ]]
}

@test "independent files upload concurrently" {
  mkdir "$BATS_TEST_TMPDIR/barrier"
  run_upload parallel "BARRIER_DIR=$BATS_TEST_TMPDIR/barrier"
  [ "$(wc -l < "$BATS_TEST_TMPDIR/parallel.destinations")" -eq 4 ]
}

@test "a failed parallel transfer prevents a successful uploaded output" {
  run run_upload parallel-failure FAIL_UPLOAD=23 FAIL_BASENAME=package.tar.gz
  [ "$status" -eq 23 ]
  [[ ! -f "$BATS_TEST_TMPDIR/parallel-failure.output" ]]
}
