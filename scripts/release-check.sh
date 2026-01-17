#!/bin/sh
set -eu

ROOT_DIR=$(cd "$(dirname "$0")/.." && pwd)
BIN="$ROOT_DIR/bin/git-ai-commit"
REPORT_BASE="$ROOT_DIR/tmp"
TIMESTAMP=$(date +%Y%m%d-%H%M%S)
REPORT_DIR="$REPORT_BASE/acceptance-$TIMESTAMP"
SUMMARY="$REPORT_DIR/summary.md"

ENGINES="claude gemini codex"
PRESETS="conventional default gitmoji karma"

mkdir -p "$REPORT_DIR"

if [ ! -x "$BIN" ]; then
  echo "Building bin/git-ai-commit..."
  (cd "$ROOT_DIR" && make build)
fi

printf "# Release acceptance check\n\n" > "$SUMMARY"
printf "Generated: %s\n\n" "$TIMESTAMP" >> "$SUMMARY"

for engine in $ENGINES; do
  for preset in $PRESETS; do
    case_dir="$REPORT_DIR/${engine}_${preset}"
    mkdir -p "$case_dir"

    tmp_repo=$(mktemp -d)
    tmp_cfg=$(mktemp -d)

    git -C "$tmp_repo" init >/dev/null
    git -C "$tmp_repo" config user.name "Release Check"
    git -C "$tmp_repo" config user.email "release-check@example.com"

    mkdir -p "$tmp_cfg/git-ai-commit"

    case "$engine" in
      claude)
        cat > "$tmp_cfg/git-ai-commit/config.toml" <<'EOC'
engine = "claude"

[engines.claude]
args = ["-p", "--model", "haiku"]
EOC
        ;;
      gemini)
        cat > "$tmp_cfg/git-ai-commit/config.toml" <<'EOC'
engine = "gemini"

[engines.gemini]
args = ["-m", "gemini-2.5-flash", "-p", "{{prompt}}"]
EOC
        ;;
      codex)
        cat > "$tmp_cfg/git-ai-commit/config.toml" <<'EOC'
engine = "codex"

[engines.codex]
args = ["exec", "--model", "gpt-5.1-codex-mini"]
EOC
        ;;
      *)
        echo "Unknown engine: $engine" > "$case_dir/error.txt"
        printf -- "- [ ] %s + %s: UNKNOWN ENGINE\n" "$engine" "$preset" >> "$SUMMARY"
        continue
        ;;
    esac

    printf "engine: %s\npreset: %s\n" "$engine" "$preset" > "$case_dir/metadata.txt"

    echo "change for $engine $preset" > "$tmp_repo/file.txt"
    git -C "$tmp_repo" add file.txt

    set +e
    (cd "$tmp_repo" && XDG_CONFIG_HOME="$tmp_cfg" "$BIN" --prompt-preset "$preset" > "$case_dir/stdout.txt" 2> "$case_dir/stderr.txt")
    status=$?
    set -e

    if [ $status -ne 0 ]; then
      printf -- "- [ ] %s + %s: FAILED (exit %s)\n" "$engine" "$preset" "$status" >> "$SUMMARY"
      git -C "$tmp_repo" status -sb > "$case_dir/git-status.txt"
      continue
    fi

    git -C "$tmp_repo" log -1 --pretty=%B > "$case_dir/commit-message.txt"
    git -C "$tmp_repo" show -1 --format= --no-color > "$case_dir/diff.txt"
    git -C "$tmp_repo" status -sb > "$case_dir/git-status.txt"

    subject=$(head -n 1 "$case_dir/commit-message.txt" | tr -d '\r')
    printf -- "- [ ] %s + %s: %s\n" "$engine" "$preset" "$subject" >> "$SUMMARY"
  done
done

printf "\nReview checklist saved to %s\n" "$SUMMARY"
