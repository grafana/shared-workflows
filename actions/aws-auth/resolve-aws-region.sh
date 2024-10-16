#!/bin/sh
# Pulled from catnekaise/cognito-idpool-auth/action.yml
# https://github.com/catnekaise/cognito-idpool-auth/blob/83ae9e159de469b3acd87ecb361d6b5957ee35ae/action.yml#L192-L227
value=""

if [ -n "${AWS_REGION}" ] && [ -n "${AWS_DEFAULT_REGION}" ]; then
  value="$AWS_REGION"
fi

readonly value

if [ -z "${value}" ]; then
  echo 'Unable to resolve what AWS region to use'
  exit 1
fi

# Some-effort validation of aws region
if echo "${value}" | grep -Eqv '^[a-z]{2}-[a-z]{4,9}-[0-9]$'; then
  echo 'Resolved value for AWS region is invalid'
  exit 1
fi

echo "value=${value}" >> "${GITHUB_OUTPUT}"
echo "AWS_REGION=${AWS_REGION}" >> "${GITHUB_ENV}"
echo "AWS_DEFAULT_REGION=${AWS_DEFAULT_REGION}" >> "${GITHUB_ENV}"
