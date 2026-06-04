#!/usr/bin/env bash
# Thin wrapper: maps friendly /news arguments to news-fetch flags.
# Usage: run.sh [category] [window]
#   category: ai | security | hacker | tech | all   (default: all)
#   window:   a Go duration like 24h, 12h, 48h        (default: 24h)
# Examples:
#   run.sh              -> all, 24h
#   run.sh ai           -> ai, 24h
#   run.sh security 12h -> security, 12h
set -euo pipefail

DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BIN="${DIR}/bin/news-fetch"
FEEDS="${DIR}/feeds.json"

category="all"
window="24h"

for arg in "$@"; do
  case "$arg" in
    ai|security|hacker|tech|all) category="$arg" ;;
    *h|*m) window="$arg" ;;            # duration like 24h, 90m
    *) ;;                               # ignore anything else
  esac
done

# Self-healing: build the binary from bundled src/ on first run (cross-platform).
# Avoids depending on plugin PostInstall hooks, which aren't available on all
# Claude Code versions.
if [[ ! -x "$BIN" ]]; then
  if [[ -f "${DIR}/src/main.go" ]] && command -v go >/dev/null 2>&1; then
    echo "news-fetch: first run, building binary with go..." >&2
    mkdir -p "${DIR}/bin"
    ( cd "${DIR}/src" && go build -o "${BIN}" . ) >&2
  else
    echo "error: news-fetch binary not found at $BIN." >&2
    echo "       install Go (https://go.dev/dl/) then run: cd ${DIR}/src && go build -o ../bin/news-fetch ." >&2
    exit 1
  fi
fi

exec "$BIN" -feeds "$FEEDS" -category "$category" -since "$window" -max 25
