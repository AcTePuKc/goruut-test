#!/usr/bin/env bash
set -euo pipefail

# Tiny metrics extractor stub.
# Parses success-rate lines and emits CSV.
# TODO: Extend to parse per-epoch metrics from external trainer logs.

usage() {
  echo "Usage: $0 --log <path> --out <csv_path>"
}

log_path=""
out_path=""

while [[ $# -gt 0 ]]; do
  case "$1" in
    --log)
      log_path="$2"; shift 2;;
    --out)
      out_path="$2"; shift 2;;
    -h|--help)
      usage; exit 0;;
    *)
      echo "Unknown arg: $1"; usage; exit 1;;
  esac
done

if [[ -z "$log_path" || -z "$out_path" ]]; then
  usage; exit 1
fi

{
  echo "epoch,metric,value,language"
  grep -E "^\[success rate\]" "$log_path" | while read -r line; do
    # Example: [success rate] 98 % with 310 errors 9895 successes for afrikaans
    metric="success_rate"
    value=$(echo "$line" | awk '{print $3}')
    language=$(echo "$line" | awk '{print $11}')
    echo "NA,$metric,$value,$language"
  done
} > "$out_path"

# TODO: If your trainer logs include epoch numbers, map them into the CSV "epoch" column.
