#!/usr/bin/env bash
# Write GitHub release notes markdown for commits between two v* tags.
# Usage: generate-release-notes.sh <new-tag> [previous-tag]
# If previous-tag is omitted, resolves the tag before <new-tag>.
set -euo pipefail

NEW_TAG="${1:?new tag required (e.g. v0.1.1)}"
PREV_TAG="${2:-}"

if [[ -z "$PREV_TAG" ]]; then
  PREV_TAG="$(git describe --tags --abbrev=0 "${NEW_TAG}^" 2>/dev/null || true)"
fi

if [[ -n "$PREV_TAG" ]]; then
  RANGE="${PREV_TAG}..${NEW_TAG}"
  COMPARE_URL="https://github.com/${GITHUB_REPOSITORY:-ksoviero/terraform-provider-signoz}/compare/${PREV_TAG}...${NEW_TAG}"
else
  RANGE="${NEW_TAG}"
  COMPARE_URL=""
fi

# Classify commit subject (same prefixes as compute-next-version.sh / AGENTS.md).
commit_group() {
  local subject="$1"
  local lower="${subject,,}"
  if [[ "$lower" =~ ^merge[[:space:]] ]]; then
    echo "skip"
  elif [[ "$lower" =~ ^(major): ]]; then
    echo "major"
  elif [[ "$lower" =~ ^(minor|feat): ]]; then
    echo "minor"
  elif [[ "$lower" =~ ^(patch|bug|fix): ]]; then
    echo "patch"
  elif [[ "$lower" =~ ^(chore|docs|test|ci|refactor|style|build|skip): ]]; then
    echo "skip"
  else
    echo "other"
  fi
}

major_lines=()
minor_lines=()
patch_lines=()
other_lines=()

while IFS= read -r subject; do
  [[ -z "$subject" ]] && continue
  case "$(commit_group "$subject")" in
    major) major_lines+=("$subject") ;;
    minor) minor_lines+=("$subject") ;;
    patch) patch_lines+=("$subject") ;;
    other) other_lines+=("$subject") ;;
  esac
done < <(git log --format=%s "$RANGE" 2>/dev/null || true)

print_section() {
  local title="$1"
  shift
  if [[ $# -eq 0 ]]; then
    return
  fi
  printf '## %s\n\n' "$title"
  for line in "$@"; do
    printf -- '- %s\n' "$line"
  done
  printf '\n'
}

printf '# %s\n\n' "$NEW_TAG"
cat <<'EOF'
Terraform provider for [SigNoz](https://signoz.io/). Install from the [Terraform Registry](https://registry.terraform.io/providers/ksoviero/signoz/latest).

Signed provider binaries and `terraform-registry-manifest.json` are attached below for the HashiCorp Registry.

EOF

if [[ -n "$COMPARE_URL" ]]; then
  printf '**Full changelog:** [%s...%s](%s)\n\n' "$PREV_TAG" "$NEW_TAG" "$COMPARE_URL"
fi

if [[ ${#major_lines[@]} -gt 0 ]]; then
  print_section "Breaking changes" "${major_lines[@]}"
fi
if [[ ${#minor_lines[@]} -gt 0 ]]; then
  print_section "Features" "${minor_lines[@]}"
fi
if [[ ${#patch_lines[@]} -gt 0 ]]; then
  print_section "Bug fixes" "${patch_lines[@]}"
fi
if [[ ${#other_lines[@]} -gt 0 ]]; then
  print_section "Other changes" "${other_lines[@]}"
fi

if [[ ${#major_lines[@]} -eq 0 && ${#minor_lines[@]} -eq 0 && ${#patch_lines[@]} -eq 0 && ${#other_lines[@]} -eq 0 ]]; then
  printf '_No categorized commits in range %s._\n' "$RANGE"
fi
