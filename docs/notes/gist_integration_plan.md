# Gist Integration Plan

## Dictionary file locations and formats

* **Runtime dictionary load paths**: `repo/dict_phonemizer_repo.go` loads dictionary records from `missing.tsv`, `missing_reverse.tsv` (reverse mode), and `missing.all.zlib` for each language directory. It reads tab-delimited rows with two or three columns, interpreting the first two columns as source/target and optional JSON tag list in the third column, then normalizes by stripping spaces. The loader uses `missing` + `_reverse` naming to decide which TSV to open and also accepts zlib-compressed `missing.all.zlib` input. This is the main runtime lookup surface for dictionary data. (`LoadLanguage` in `DictPhonemizerRepository`).

* **Training inputs referenced by scripts**: `cmd/analysis2/*.sh` point at `dicts/<lang>/lexicon.tsv`, `dicts/<lang>/learn.tsv`, and `dicts/<lang>/learn_reverse.tsv` for training and study workflows. The scripts provide the corpus to external trainer binaries (e.g., `train_phonemizer2`) and to local CLI tools. These files are TSV with word/IPA (and sometimes tags) pairs.

* **Backtest/regeneration artifacts**: `cmd/backtest/main.go` reads `dicts/<lang>/lexicon.tsv`, writes `learn.tsv`/`learn_reverse.tsv`, and can produce `missing.all.zlib` by recompressing `missing.all.tsv`. This is a generator path for `missing.all.zlib` artifacts.

## Training/data pipeline entrypoints

* **TSV ingestion in analysis2**: `cmd/analysis2/main.go` calls `load(*srcFile, 999999999)` to ingest TSV word↔IPA data, then cleans/normalizes tokens and uses the data to initialize or update `language.json`. This is the primary in-repo entrypoint for processing TSV dictionaries in the analysis pipeline.

* **Shell scripts that call external trainers**: `cmd/analysis2/learn_language.sh` and `cmd/analysis2/learn_language_reverse.sh` invoke `train_phonemizer2` with `--lexicontsv` and `--learntsv` arguments and run `backtest` in parallel. `cmd/analysis2/clean_language.sh` and `cmd/analysis2/study_language.sh` call the local `analysis2` binary, typically using `lexicon.tsv` as the source and writing `clean*.tsv` outputs that are then fed to `dicttomap`.

* **Backtest path**: `cmd/backtest/main.go` is used for generating/validating missing entries and training updates, including `missing.all.zlib` output.

## Minimal-intrusion hook points

* **Preprocessing step before analysis2**: add a preprocessing pass that rewrites a TSV to a temp file (e.g., filtered subset of `lexicon.tsv`) and feed it into `cmd/analysis2` via `--srcfile`. This avoids touching existing training or runtime logic.

* **Reduced dictionary generator CLI**: add a new CLI that reads `missing.tsv`/`missing_reverse.tsv`/`missing.all.zlib`, computes a minimal subset (gist), and writes out a new TSV (e.g., `missing.gist.tsv`). That file can then be swapped in by existing scripts or via a new flag without changing core logic.

* **Optional hook in backtest output**: since `cmd/backtest/main.go` already generates `learn.tsv` and `missing.all.zlib`, adding a post-step there (or a wrapper script) to run a gist selector on `missing.all.tsv` keeps the changes localized.

## Proposed design

* **New package**: `pkg/gistselect` (public) or `internal/gistselect` (private) to implement the gist selection algorithm. It should:
  * Load TSV (two/three columns), preserve tags, and return a subset based on coverage goals.
  * Accept optional settings for reverse mode, prefilter, and coverage thresholds.
  * Expose deterministic selection (sorted + stable tie-breaks).

* **New CLI**: `cmd/gistselect` to:
  * Read `--src` (TSV or zlib), emit `--dst` TSV.
  * Provide `--reverse` or `--source-col`/`--target-col` options for flexible input.
  * Include `--max-entries` and `--prefilter` flags for bounded outputs.
  * Print summary stats (coverage, dropped rows, runtime).

* **Integration points**:
  * A wrapper script alongside `cmd/analysis2/*.sh` that calls `cmd/gistselect` before running `analysis2` or `train_phonemizer2`.
  * Optional config to switch the dict loader to `missing.gist.tsv` if present (opt-in to keep behavior stable).

## Complexity notes

* **Baseline complexity**: a straightforward independence check for each candidate row against a set of `k` coverage features yields `O(n·k)` time and requires in-memory maps of covered features.

* **Memory**: store per-feature coverage maps (e.g., `map[string]uint32` for counts or `map[string]struct{}` for presence). Size is proportional to the number of unique grapheme/phoneme features in the training data.

* **Cheap prefilter**: optional cheap filter pass (e.g., keep only rows with rare graphemes/phones or rows above a length threshold) can reduce `n` before the `O(n·k)` selection loop while keeping deterministic output.
