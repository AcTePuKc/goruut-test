#!/usr/bin/env bash
set -euo pipefail

# Tiny runner stub for the training/eval workflow.
# NOTE: This repo calls external trainer binaries located at ../../../classifier.
# If those binaries are missing, this script will fail. See TODO below.

usage() {
  echo "Usage: $0 --lang <language> [--gist] [--gist-k <k>] [--gist-lambda <lambda>] [--log <path>]"
}

lang=""
log_path="./graph_run.log"
gist_enable=false
gist_k=0
gist_lambda=1.0

while [[ $# -gt 0 ]]; do
  case "$1" in
    --lang)
      lang="$2"; shift 2;;
    --log)
      log_path="$2"; shift 2;;
    --gist)
      gist_enable=true; shift;;
    --gist-k)
      gist_k="$2"; shift 2;;
    --gist-lambda)
      gist_lambda="$2"; shift 2;;
    -h|--help)
      usage; exit 0;;
    *)
      echo "Unknown arg: $1"; usage; exit 1;;
  esac
done

if [[ -z "$lang" ]]; then
  echo "Missing --lang"; usage; exit 1
fi

# TODO: Ensure the external trainer exists at ../../../classifier/cmd/train_phonemizer_ulevel/train_phonemizer_ulevel
# TODO: Provide a Windows-friendly equivalent or use WSL to execute bash scripts.

{
  echo "[graph-runner] language=$lang"
  echo "[graph-runner] gist_enable=$gist_enable gist_k=$gist_k gist_lambda=$gist_lambda"
  echo "[graph-runner] start: $(date -u)"

  pushd cmd/analysis2 >/dev/null
  if [[ "$gist_enable" == "true" ]]; then
    ./analysis2 --lang "../../dicts/$lang/language.json" --srcfile "../../dicts/$lang/lexicon.tsv" \
      --gist_enable --gist_select_k "$gist_k" --gist_lambda "$gist_lambda"
  else
    ./study_language.sh "$lang"
    ./clean_language.sh "$lang"
  fi
  ./train_language.sh "$lang"
  popd >/dev/null

  # Optional: backtest to produce success-rate output lines.
  # TODO: adjust path/flags as needed for your environment.
  # pushd cmd/backtest >/dev/null
  # go run . --langname "$lang" --testing
  # popd >/dev/null

  echo "[graph-runner] end: $(date -u)"
} | tee "$log_path"
