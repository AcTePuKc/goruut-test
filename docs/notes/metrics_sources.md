# Metric sources inventory

This note lists metric-like outputs present in the repo and their likely producers.

## Summary tables
### Static output files (checked into repo)
- `success_forward.txt` / `success_reverse.txt`
  - **Format**: `[success rate] <percent> % with <errors> errors <successes> successes for <lang>`
  - **Likely producer**: `cmd/backtest/main.go` (prints the exact format).
  - **Observed in code**: success rate print near end of `cmd/backtest/main.go`.

### Script outputs (stdout)
- `cmd/backtest/main.go`
  - **Prints**: success rate line for a language.
  - **Inputs**: `dicts/<lang>/missing*.tsv` or `missing.all.zlib` via dict loader.
  - **Purpose**: full lexicon evaluation; can run in “testing” mode to track best weights.

- `cmd/phondephontest/main.go`
  - **Prints**: `[success rate WER]` and `[success rate CER]` to stdout.
  - **Inputs**: language name and test data.

### Coverage outputs
- `cmd/backtest/coverage.sh`
  - **Prints**: `Coverage forward: <percent>%` and `Coverage reverse: <percent>%`.
  - **Inputs**: line counts of `dicts/<lang>/lexicon.tsv`, `clean.tsv`, `clean_reverse.tsv`.

- `doc/README.md`
  - Mentions `coverage_*.txt` files, but **no generator** found in this repo. Likely produced in an external workflow or older scripts.

## Per-epoch / time-series metrics
No files or code paths in this repo explicitly write per-epoch metrics. The external trainer invoked by `cmd/analysis2/train_language.sh` likely produces epoch logs, but those binaries live outside this repo.

## Log formats / regex hints
- Success rate (backtest): `^\[success rate\] (\d+) % with (\d+) errors (\d+) successes for (.+)$`
- WER/CER (phondephontest): `^\[success rate (WER|CER)\] (\d+) % (\d+) for (.+)$`

## Gaps / follow-ups
- If a graph of “Success Rate vs Epoch” is required, the epoch data must be captured from:
  - External training binary logs, or
  - New logging in this repo (not currently present).
