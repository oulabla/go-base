#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

CONFIG_FILE="$ROOT_DIR/config/prod.yaml"
TEMPLATE_FILE="$ROOT_DIR/template/keys.gen.go.tpl"
OUTPUT_FILE="$ROOT_DIR/internal/config/keys.gen.go"

TMP_JSON="$(mktemp)"

echo "Using yq at: ${YQ_PATH}"

"${YQ_PATH}" -o=json '.config' "$CONFIG_FILE" > "$TMP_JSON"

if [ ! -s "$TMP_JSON" ]; then
  echo "ERROR: extracted config is empty"
  exit 1
fi

go run "$ROOT_DIR/script/generate_config.go" \
  -input "$TMP_JSON" \
  -template "$TEMPLATE_FILE" \
  -output "$OUTPUT_FILE"

rm "$TMP_JSON"

echo "Config generated successfully"
