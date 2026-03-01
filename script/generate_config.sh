#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

CONFIG_FILE="$ROOT_DIR/config/prod.yaml"
TEMPLATE_FILE="$ROOT_DIR/templates/keys.gen.go.tpl" 
OUTPUT_FILE="$ROOT_DIR/internal/config/keys.gen.go"

YQ_BIN="$ROOT_DIR/bin/yq"

if [ ! -x "$YQ_BIN" ]; then
  echo "ERROR: yq not found in $YQ_BIN"
  echo "Run 'make tools' first"
  exit 1
fi

TMP_JSON="$(mktemp)"

"$YQ_BIN" -o=json '.config' "$CONFIG_FILE" > "$TMP_JSON"

if [ ! -s "$TMP_JSON" ]; then
  echo "ERROR: extracted config is empty or config/prod.yaml has no .config section"
  rm -f "$TMP_JSON"
  exit 1
fi

go run "$ROOT_DIR/script/generate_config.go" \
  -config "$CONFIG_FILE" \
  -template "$TEMPLATE_FILE" \
  -output "$OUTPUT_FILE"

rm -f "$TMP_JSON"

echo "Config keys generated → $OUTPUT_FILE"