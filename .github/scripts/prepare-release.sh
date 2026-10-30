#!/usr/bin/env bash
# Prepare a GitHub release tag in the same workflow run that publishes it.
# Usage: prepare-release.sh <release-major|release-minor|release-patch> [vX.Y.Z]
set -euo pipefail

mode="${1:?release mode is required}"
override="${2:-}"
semver='^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$'
case "$mode" in
  release-major|release-minor|release-patch) ;;
  *) echo "Unsupported publish mode: $mode" >&2; exit 2 ;;
esac
if [[ -n "$override" && ! "$override" =~ $semver ]]; then
  echo "Invalid release_version_override: expected vMAJOR.MINOR.PATCH" >&2
  exit 2
fi

# Explicit dispatches must publish the current default branch, not arbitrary refs.
if [[ "${RELEASE_TAG_DRY_RUN:-false}" != 'true' ]]; then
  if [[ "${GITHUB_REF:-}" != 'refs/heads/main' ]]; then
    echo 'Manual releases must be dispatched from main' >&2
    exit 2
  fi
  git fetch --force --tags origin
  git fetch origin main
  if [[ "$(git rev-parse HEAD)" != "$(git rev-parse refs/remotes/origin/main)" ]]; then
    echo 'Checkout is not the latest origin/main; release aborted' >&2
    exit 2
  fi
fi

existing="$(git tag --points-at HEAD | grep -E "$semver" | sort -V | tail -n 1 || true)"
latest="$(git tag --list | grep -E "$semver" | sort -V | tail -n 1 || true)"

if [[ -n "$override" ]]; then
  version="$override"
elif [[ -n "$existing" ]]; then
  # Retry an interrupted release on the same commit without creating a bump.
  version="$existing"
else
  if [[ -z "$latest" ]]; then
    major=0; minor=0; patch=0
  else
    number="${latest#v}"
    IFS=. read -r major minor patch <<< "$number"
  fi
  case "$mode" in
    release-major) version="v$((major+1)).0.0" ;;
    release-minor) version="v${major}.$((minor+1)).0" ;;
    release-patch) version="v${major}.${minor}.$((patch+1))" ;;
  esac
fi

if git show-ref --tags --verify --quiet "refs/tags/$version"; then
  if [[ "$(git rev-list -n 1 "$version")" != "$(git rev-parse HEAD)" ]]; then
    echo "Tag $version exists on another commit; refusing to retag" >&2
    exit 1
  fi
else
  if [[ "${RELEASE_TAG_DRY_RUN:-false}" != 'true' ]]; then
    git config user.name 'github-actions[bot]'
    git config user.email '41898282+github-actions[bot]@users.noreply.github.com'
    git tag -a "$version" -m "Release $version"
    git push origin "refs/tags/$version"
  fi
fi

if [[ "${RELEASE_TAG_DRY_RUN:-false}" != 'true' ]]; then
  if gh release view "$version" --repo "${GITHUB_REPOSITORY:?}" >/dev/null 2>&1; then
    echo "Release $version already exists; refusing duplicate publication" >&2
    exit 1
  fi
  echo "tag=$version" >> "${GITHUB_OUTPUT:?}"
fi
printf 'Release version: %s\n' "$version"
