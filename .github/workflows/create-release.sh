#!/bin/sh

set -eu

AUTH="Bearer ${GITHUB_TOKEN}"
URL="${GITHUB_API_URL}/repos/${GITHUB_REPOSITORY}/releases"

version=$(./cas version --short)
body=$(./cas version --changelog | awk '{printf "%s\\n", $0}' )

json=$(cat <<EOF
{
  "tag_name": "${version}",
  "name": "${version}",
  "body": "${body}",
  "draft": false
}
EOF
)


curl -X POST \
  --url "${URL}" \
  --header "Authorization: ${AUTH}" \
  --header "Accept: application/vnd.github+json" \
  -d "${json}"
