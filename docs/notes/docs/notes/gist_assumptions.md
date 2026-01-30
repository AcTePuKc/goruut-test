# Gist Assumptions

## Assumed dictionary formats

* `dicts/<lang>/missing.tsv` and `dicts/<lang>/missing_reverse.tsv` are tab-delimited files with **two columns** (word, IPA). Example from `dicts/swedish/missing.tsv`:
  * `bb<TAB>bˌeːbˈeː`

* `repo/dict_phonemizer_repo.go` also accepts **three-column** rows (word, IPA, tags JSON) and consumes `missing.all.zlib` (zlib-compressed TSV) in addition to the plain TSVs.

* Tags are stored as JSON arrays in the third TSV column (e.g., `"[\"dict\",\"learn\"]"`) and are merged when duplicate word pairs are encountered.

## Assumed entrypoints

* `cmd/analysis2/main.go` calls `load(filename string, top int)` and expects two- or three-column TSV input. This is the primary ingest path for `lexicon.tsv` and related training files.

* `cmd/analysis2/*.sh` scripts are assumed to be the orchestrators for training and preprocessing (e.g., `learn_language.sh`, `study_language.sh`, `clean_language.sh`). They reference `lexicon.tsv`, `learn.tsv`, and `learn_reverse.tsv` as inputs for both training and analysis.

* `cmd/backtest/main.go` uses `lexicon.tsv` for evaluation and writes `learn.tsv`, `learn_reverse.tsv`, and `missing.all.zlib` for downstream use.

## Unknowns and validation steps

* `lexicon.tsv` is referenced by multiple scripts/commands but **does not appear in the repo**; confirm its presence and schema in the deployed environment or build pipeline.

* There is no canonical in-repo schema doc for the TSV tag column. Validate with a sample `missing.all.zlib` by decoding it and inspecting rows for third-column JSON tags.

* Validate that `missing_reverse.tsv` swaps word/IPA columns as expected for reverse mode by comparing `cmd/analysis2` reverse runs with the runtime loader behavior in `DictPhonemizerRepository`.

* Confirm whether `learn.tsv`/`learn_reverse.tsv` share the same tag conventions as `missing.tsv` so that a gist selector can preserve or merge tags consistently.

* Ensure any new gist selector CLI can process both plain TSV and zlib-compressed TSV (`missing.all.zlib`) and that its output integrates with existing scripts without needing to modify `analysis2` defaults.
