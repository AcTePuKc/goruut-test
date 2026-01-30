# GIST insertion plan for graph-producing workflow

## Goal
Add an **opt-in** preprocessing stage that applies `cmd/gistselect` to training lexicons used in the training/eval workflow that ultimately produces success metrics. The default behavior must remain unchanged.

## Recommended insertion points (opt-in)
### 1) Wrapper script before analysis/training (lowest risk)
Add a wrapper that:
1. Runs `cmd/gistselect` on `dicts/<lang>/lexicon.tsv` (or `learn.tsv` when applicable).
2. Writes a temporary `lexicon.gist.tsv`.
3. Calls existing `cmd/analysis2/*.sh` scripts with the GIST-ed TSV path.

This keeps all existing scripts untouched and only changes input paths when explicitly requested.

### 2) Optional flags in analysis/training entrypoints (already present)
`cmd/analysis2/main.go` now supports:
- `--gist_enable`
- `--gist_select_k`
- `--gist_lambda`

These flags allow pre-filtering during `analysis2` runs without changing defaults. For graph workflows that already use `analysis2`, this is the simplest opt-in path.

## Suggested experimental defaults
- Start with `k = 1000–5000` for medium lexicons.
- `lambda = 0.5–1.0` to balance coverage and diversity.
- Utility: joint word + IPA n-grams (default in `cmd/gistselect`).
- Distance: joint word + IPA Levenshtein.

## Example usage
### Wrapper script (preferred for training pipeline)
```
./scripts/graph_runner.sh --lang english --gist --gist-k 2000 --gist-lambda 0.8
```
### Direct analysis2 usage
```
./cmd/analysis2/analysis2 --lang dicts/english/language.json --srcfile dicts/english/lexicon.tsv \
  --gist_enable --gist_select_k 2000 --gist_lambda 0.8
```

## Notes
- `cmd/analysis2/train_language.sh` invokes an external trainer (`../../../classifier/...`). GIST should only adjust the **input TSV**, leaving training parameters unchanged.
- For graphing, capture logs from the external trainer and/or `cmd/backtest` runs after training.
