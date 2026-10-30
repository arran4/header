#!/usr/bin/env bash
set -euo pipefail
script="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/prepare-release.sh"
base="$(mktemp -d)"
trap 'rm -rf "$base"' EXIT
cd "$base"
git init -q
git config user.name Test
git config user.email test@example.invalid
git commit --allow-empty -qm initial
expect() {
  local expected="$1" mode="$2" override="${3:-}" actual
  actual="$(RELEASE_TAG_DRY_RUN=true bash "$script" "$mode" "$override")"
  if [[ "$actual" != "Release version: $expected" ]]; then
    echo "Unexpected release version: $actual; expected $expected" >&2
    exit 1
  fi
}
expect v0.0.1 release-patch
expect v0.1.0 release-minor
expect v1.0.0 release-major
expect v3.2.7 release-patch v3.2.7
if RELEASE_TAG_DRY_RUN=true bash "$script" release-patch 'v1.2.3;bad' >/dev/null 2>&1; then
  echo 'Accepted invalid version override' >&2; exit 1
fi
git tag -a v0.1.9 -m tag
git commit --allow-empty -qm update
expect v0.1.10 release-patch
expect v0.2.0 release-minor
expect v1.0.0 release-major
git tag -a v0.1.10 -m tag
expect v0.1.10 release-patch
if RELEASE_TAG_DRY_RUN=true bash "$script" release-patch v0.1.9 >/dev/null 2>&1; then
  echo 'Accepted a tag override pointing to a different commit' >&2; exit 1
fi
printf 'release version tests passed\n'
