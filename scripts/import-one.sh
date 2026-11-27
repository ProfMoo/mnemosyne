#!/usr/bin/env bash
set -euo pipefail

url="${1:?Usage: $0 ENDPOINT JSON_FILE}"
file="${2:?Usage: $0 ENDPOINT JSON_FILE}"

headers=(-H "Content-Type: application/json")
if [[ -n "${LISTENBRAINZ_TOKEN:-}" ]]; then
  headers+=(-H "Authorization: Token $LISTENBRAINZ_TOKEN")
fi

curl --fail-with-body --silent --show-error \
  "${headers[@]}" --data-binary "@$file" "$url"
