#!/usr/bin/env bash
set -euo pipefail

usage() {
  echo "Usage: $0 --lang <language> --log <path> [--reverse] [--overwrite] [--] [trainer args...]"
}

lang=""
log_path=""
reverse_flag=""
extra_args=()

while [[ $# -gt 0 ]]; do
  case "$1" in
    --lang)
      lang="$2"; shift 2;;
    --log)
      log_path="$2"; shift 2;;
    --reverse)
      reverse_flag="--reverse"; shift;;
    --overwrite)
      extra_args+=("--overwrite"); shift;;
    --)
      shift; extra_args+=("$@"); break;;
    -h|--help)
      usage; exit 0;;
    *)
      extra_args+=("$1"); shift;;
  esac
done

if [[ -z "$lang" || -z "$log_path" ]]; then
  usage; exit 1
fi

if [[ ! -x "../../../classifier/cmd/train_phonemizer_ulevel/train_phonemizer_ulevel" ]]; then
  echo "External trainer not found at ../../../classifier/cmd/train_phonemizer_ulevel/train_phonemizer_ulevel" >&2
  echo "This repo relies on an external trainer. Provide it or update the path." >&2
  exit 1
fi

pushd cmd/analysis2 >/dev/null
{
  echo "[capture-training-log] language=$lang reverse=$reverse_flag"
  echo "[capture-training-log] start: $(date -u)"
  ./train_language${reverse_flag:+_reverse}.sh "$lang" "${extra_args[@]}"
  echo "[capture-training-log] end: $(date -u)"
} | tee "$log_path"
popd >/dev/null
