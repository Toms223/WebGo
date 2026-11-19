#!/bin/sh
dir=$(cd "$(dirname "$0")" && pwd)
exec node --stack-size=8192 --import "$dir/wasm_fake_loader.mjs" "$(go env GOROOT)/lib/wasm/wasm_exec_node.js" "$@"
