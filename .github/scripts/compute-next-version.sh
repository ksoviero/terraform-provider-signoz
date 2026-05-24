#!/usr/bin/env bash
# Compute the next semver tag from commit subjects since the latest v* tag.
# Bump keywords must appear at the start of the subject (case-insensitive).
# See AGENTS.md "Commit messages and releases".
set -euo pipefail

bump_rank() {
  case "$1" in
    major) echo 3 ;;
    minor) echo 2 ;;
    patch) echo 1 ;;
    *) echo 0 ;;
  esac
}

max_bump() {
  local a="$1" b="$2"
  if [[ $(bump_rank "$a") -ge $(bump_rank "$b") ]]; then
    echo "$a"
  else
    echo "$b"
  fi
}

# Optional Conventional Commits scope: fix(ci): msg -> fix: msg
normalize_subject() {
  printf '%s' "$1" | sed -E 's/^([a-z][a-z0-9_-]*)\([^)]+\):/\1:/'
}

# Subject line -> patch | minor | major | none
commit_bump() {
  local subject
  subject="$(normalize_subject "$1")"
  local lower="${subject,,}"

  # Merge commits carry no release intent; rely on merged branch commits in the range.
  if [[ "$lower" =~ ^merge[[:space:]] ]]; then
    echo "none"
    return
  fi

  if [[ "$lower" =~ ^(major): ]]; then
    echo "major"
  elif [[ "$lower" =~ ^(minor|feat): ]]; then
    echo "minor"
  elif [[ "$lower" =~ ^(patch|bug|fix): ]]; then
    echo "patch"
  elif [[ "$lower" =~ ^(skip|chore|docs|test|ci|refactor|style|build): ]]; then
    echo "none"
  else
    echo "none"
  fi
}

latest_tag="$(git tag -l 'v*' --sort=-v:refname | head -n1 || true)"

if [[ -n "$latest_tag" ]]; then
  current="${latest_tag#v}"
  range="${latest_tag}..HEAD"
else
  current="0.0.0"
  range="HEAD"
fi

if ! [[ "$current" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
  echo "error: latest tag ${latest_tag:-<none>} is not semver vMAJOR.MINOR.PATCH" >&2
  exit 1
fi

mapfile -t subjects < <(git log --format=%s "$range" 2>/dev/null || true)

if [[ ${#subjects[@]} -eq 0 ]]; then
  echo "skip=no commits since ${latest_tag:-<initial>}"
  exit 0
fi

chosen="none"
for subject in "${subjects[@]}"; do
  level="$(commit_bump "$subject")"
  chosen="$(max_bump "$chosen" "$level")"
done

if [[ "$chosen" == "none" ]]; then
  echo "skip=no release keyword in commits since ${latest_tag:-<initial>}"
  exit 0
fi

IFS=. read -r maj min pat <<<"$current"
case "$chosen" in
  major)
    maj=$((maj + 1))
    min=0
    pat=0
    ;;
  minor)
    min=$((min + 1))
    pat=0
    ;;
  patch)
    pat=$((pat + 1))
    ;;
esac

next="v${maj}.${min}.${pat}"
echo "next=${next}"
echo "bump=${chosen}"
echo "previous=${latest_tag:-v0.0.0}"
