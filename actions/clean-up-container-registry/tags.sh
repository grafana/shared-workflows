#!/usr/bin/env bash

set -eu
set -o pipefail

readonly image_name=us.gcr.io/kubernetes-dev/infra-build
readonly exclude_tags="2024-06-17-*"
readonly tag_filter=
readonly keep_latest=5
readonly dry_run="true"

# List all image tags using crane
mapfile -t tags < <(crane ls "${image_name}")

# Convert exclude_tags to an array
mapfile -t exclude_array <<< "${exclude_tags}"

declare -A tags_to_consider_removing
declare -A tags_kept

# Filter tags by the glob pattern, then sort them
for tag in "${tags[@]}"; do
    # We want glob patching, that's the point!
    # shellcheck disable=SC2053
    if [[ -z "${tag_filter}" ]] || [[ ${tag} == ${tag_filter} ]]; then
        full_image_reference="${image_name}:${tag}"

        creation_date=$(crane config "${full_image_reference}" | jq -r '.created')

        for exclude in "${exclude_array[@]}"; do
        if [[ $tag == ${exclude} ]]; then
            tags_kept[${tag}]="${creation_date}"
            continue 2
        fi
        done
        tags_to_consider_removing[${tag}]="${creation_date}"
    fi
done

# Sort tags by creation date
mapfile -t tags_to_consider_removing < <(
    for tag in "${!tags_to_consider_removing[@]}"; do
        echo "${tags_to_consider_removing["${tag}"]} ${tag}"
    done | sort -k2 -r | awk '{print $1}'
)

# Sort tags to keep by creation date
mapfile -t tags_kept < <(
    for tag in "${!tags_kept[@]}"; do
        echo "${tags_kept["${tag}"]} ${tag}"
    done | sort -k2 -r | awk '{print $1}'
)

# Determine tags to keep and tags to remove
tags_to_keep=("${tags_kept[@]: -${keep_latest}}")
tags_to_remove=("${tags_to_consider_removing[@]:${keep_latest}}")

if [ "$dry_run" == "true" ]; then
  echo "Dry-run mode: the following tags would be removed:"
  for tag in "${tags_to_remove[@]}"; do
    echo "- $tag"
  done

  echo "...and the following tags would be kept:"
  for tag in "${tags_to_keep[@]}"; do
    echo "- $tag"
  done
else
  echo "Removing the following tags:"
  printf "%s\n" "${tags_to_remove[@]}"
  for tag in "${tags_to_remove[@]}"; do
    crane delete "${image_name}:${tag}"
  done
fi

# Output removed tags
echo "::set-output name=removed_tags::$(IFS=,; echo "${tags_to_remove[*]}")"
