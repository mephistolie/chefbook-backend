#!/usr/bin/env bash

set -euo pipefail

repository_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$repository_root"

service_modules=(
  api-gateway
  services/auth
  services/mail
  services/encryption
  services/profile
  services/recipe
  services/shopping-list
  services/subscription
  services/tag
  services/template
  services/user
)

legacy_pattern='log\.(Auto(Trace|Debug|Info|Warn|Error|Fatal)|Trace|Debug|Info|Warn|Error|Fatal)f?\('
legacy_declaration_pattern='func (Auto(Trace|Debug|Info|Warn|Error|Fatal)|Trace|Debug|Info|Warn|Error|Fatal)f?\b'
stdlib_pattern='log\.(Print|Printf|Println|Fatal|Fatalf|Fatalln|Panic|Panicf|Panicln)\('

if rg -n "$legacy_pattern" "${service_modules[@]}" common --glob '*.go'; then
  echo "structured logging check failed: legacy logging call found" >&2
  exit 1
fi

if rg -n "$legacy_declaration_pattern" common/log --glob '*.go'; then
  echo "structured logging check failed: legacy logging API found" >&2
  exit 1
fi

if rg -n "$stdlib_pattern" "${service_modules[@]}" common --glob '*.go'; then
  echo "structured logging check failed: standard-library logging call found" >&2
  exit 1
fi

if rg -n 'log\.Event\{' "${service_modules[@]}" \
  --glob '*.go' \
  --glob '!**/internal/logging/**'; then
  echo "structured logging check failed: construct events in internal/logging only" >&2
  exit 1
fi

echo "structured logging check passed"
