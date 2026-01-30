# Epoch metrics discovery & capture plan

## Training workflow entrypoints (in-repo)
The repo’s training-related entrypoints are shell scripts that wrap Go tools and an **external** trainer binary:

- `cmd/analysis2/study_language.sh` → runs the `analysis2` binary to seed or update the language model JSON. It feeds `dicts/<lang>/lexicon.tsv` through the analyzer and writes back the JSON if `-save` is set.【F:cmd/analysis2/study_language.sh†L1-L16】【F:cmd/analysis2/main.go†L83-L118】
- `cmd/analysis2/clean_language.sh`/`clean_languages.sh` are preprocessing steps around `analysis2` (not emitting epoch metrics).【F:cmd/analysis2/clean_language.sh†L1-L16】
- `cmd/analysis2/train_language.sh` invokes the external trainer binary `../../../classifier/cmd/train_phonemizer_ulevel/train_phonemizer_ulevel` with lang dir + model output path. This is where per-epoch logs would originate (outside this repo).【F:cmd/analysis2/train_language.sh†L1-L38】
- `cmd/analysis2/train_languages.sh` just loops over languages and calls `train_language.sh` for each language; it does not add logging.【F:cmd/analysis2/train_languages.sh†L1-L20】
- `cmd/backtest/main.go` provides a **success-rate** line (not per-epoch). It evaluates the lexicon and prints a single aggregate success rate per run.【F:cmd/backtest/main.go†L334-L347】

## Does goruut emit per-epoch metrics?
No. Within this repo, there are **no per-epoch metrics** in Go code or scripts. The only explicit metrics lines are:

- Backtest success rate: `"[success rate] <percent> % with <errors> errors <successes> successes for <lang>"` from `cmd/backtest/main.go` (single aggregate line).【F:cmd/backtest/main.go†L334-L347】
- WER/CER success rate lines in `cmd/phondephontest/main.go`, also aggregate per run (not per-epoch).【F:cmd/phondephontest/main.go†L156-L161】

Per-epoch logs are expected to come from external trainers in `../../../classifier/...` (not included in this repo).【F:cmd/analysis2/train_language.sh†L31-L35】

## Example log lines to parse
Since the external trainer logs are not visible in this repo, below are **typical** patterns that the new parsers can handle (based on common training logs, not assumptions about the trainer):

1) **Epoch with train/eval tokens** (explicit epoch, key/value pairs):
```
Epoch 3 train loss=0.25 acc=0.90 eval loss=0.50 acc=0.80
```
2) **Key/value with split prefix** (epoch optional):
```
epoch=5 train_loss=0.12 eval_loss=0.22
```
3) **No explicit epoch** (auto-increment by line):
```
train loss=1.2 eval loss=1.0
train loss=0.9 eval loss=0.7
```

## Recommended log output “contract”
To reliably produce epoch graphs, the external trainer should emit **one line per epoch**, with:
- An epoch number (e.g., `Epoch 3` or `epoch=3`),
- Split tokens `train` and `eval`,
- Key/value pairs for each metric (e.g., `loss=...`, `acc=...`).

Recommended format:
```
Epoch <N> train loss=<float> acc=<float> eval loss=<float> acc=<float> [lang=<language>]
```

This format is accepted by the new parsers and avoids ambiguity about split association.

## External trainer capture plan (if per-epoch logs live outside this repo)
1) **Capture stdout/stderr** from the training entrypoint via a wrapper script:
   - `scripts/capture_training_log.sh` (bash)
   - `scripts/capture_training_log.ps1` (PowerShell)

2) **Parse the log into CSV** with:
   - `scripts/extract_epoch_metrics.sh` (bash)
   - `scripts/extract_epoch_metrics.ps1` (PowerShell)
   - Optional Go CLI: `cmd/metrics_extract` for cross-platform use.

3) **Graphing** can be done outside this repo using the CSV (e.g., in Python, gnuplot, or Excel).

## Usage examples
Capture and extract (bash):
```
scripts/capture_training_log.sh --lang afrikaans --log /tmp/afrikaans.log
scripts/extract_epoch_metrics.sh --log /tmp/afrikaans.log --out /tmp/afrikaans.csv --language afrikaans
```

Capture and extract (PowerShell):
```
.\scripts\capture_training_log.ps1 -Lang afrikaans -Log C:\temp\afrikaans.log
.\scripts\extract_epoch_metrics.ps1 -Log C:\temp\afrikaans.log -Out C:\temp\afrikaans.csv -Language afrikaans
```

Go CLI (cross-platform):
```
go run ./cmd/metrics_extract --log /tmp/afrikaans.log --out /tmp/afrikaans.csv --language afrikaans
```
