#!/usr/bin/env bash
set -euo pipefail

usage() {
  echo "Usage: $0 --log <path> --out <csv_path> [--language <lang>] [--epoch-regex <regex>] [--split-regex <regex>] [--kv-regex <regex>] [--language-regex <regex>]"
}

log_path=""
out_path=""
language=""
epoch_regex='[Ee]poch[[:space:]]*[:=]?[[:space:]]*([0-9]+)'
split_regex='(train|eval|validation|valid|dev|test)'
kv_regex='([A-Za-z][A-Za-z0-9_.-]*)[[:space:]]*[:=][[:space:]]*([0-9]+(\\.[0-9]+)?)%?'
language_regex='(lang(uage)?[[:space:]]*[:=][[:space:]]*([A-Za-z0-9_-]+)|for[[:space:]]+([A-Za-z0-9_-]+))'

while [[ $# -gt 0 ]]; do
  case "$1" in
    --log)
      log_path="$2"; shift 2;;
    --out)
      out_path="$2"; shift 2;;
    --language)
      language="$2"; shift 2;;
    --epoch-regex)
      epoch_regex="$2"; shift 2;;
    --split-regex)
      split_regex="$2"; shift 2;;
    --kv-regex)
      kv_regex="$2"; shift 2;;
    --language-regex)
      language_regex="$2"; shift 2;;
    -h|--help)
      usage; exit 0;;
    *)
      echo "Unknown arg: $1"; usage; exit 1;;
  esac
done

if [[ -z "$log_path" || -z "$out_path" ]]; then
  usage; exit 1
fi

awk -v epoch_re="$epoch_regex" \
    -v split_re="$split_regex" \
    -v kv_re="$kv_regex" \
    -v lang_re="$language_regex" \
    -v lang_default="$language" \
    'BEGIN { print "epoch,metric,value,split,language"; auto_epoch=0 }
     function normsplit(s) {
       if (s == "validation" || s == "valid" || s == "dev" || s == "test") { return "eval" }
       return s
     }
     function split_from_key(k) {
       if (k ~ /^train([._-]|$)/) { return "train" }
       if (k ~ /^(eval|validation|valid|dev|test)([._-]|$)/) { return "eval" }
       return ""
     }
     function trim_split(k, split) {
       gsub("^" split "[._-]", "", k)
       return k
     }
     function split_from_pos(pos, n, i, best) {
       best=""
       for (i = 1; i <= n; i++) {
         if (split_pos[i] <= pos) { best = split_tok[i] } else { break }
       }
       return best
     }
     {
       line=$0
       epoch=-1
       if (match(line, epoch_re, em)) {
         epoch = em[1] + 0
         if (epoch > auto_epoch) { auto_epoch = epoch }
       }

       lang=lang_default
       if (lang == "" && match(line, lang_re, lm)) {
         if (lm[3] != "") { lang=tolower(lm[3]) }
         else if (lm[4] != "") { lang=tolower(lm[4]) }
       }

       n=0
       rest=line
       offset=0
       while (match(rest, split_re, sm)) {
         n++
         split_pos[n] = offset + RSTART - 1
         split_tok[n] = normsplit(tolower(sm[1]))
         offset += RSTART + RLENGTH - 1
         rest = substr(rest, RSTART + RLENGTH)
       }

       rest=line
       offset=0
       line_epoch_assigned = (epoch >= 0)
       while (match(rest, kv_re, km)) {
         key=tolower(km[1])
         value=km[2]
         pos=offset + RSTART - 1
         split = split_from_key(key)
         metric = key
         if (split != "") {
           metric = trim_split(metric, split)
         } else {
           split = split_from_pos(pos, n)
         }
         if (split == "") {
           offset += RSTART + RLENGTH - 1
           rest = substr(rest, RSTART + RLENGTH)
           continue
         }
         split = normsplit(split)
         gsub(/^[._-]+|[._-]+$/, "", metric)
         if (metric == "" || metric == split) { metric = "value" }
         if (!line_epoch_assigned) {
           auto_epoch++
           epoch = auto_epoch
           line_epoch_assigned = 1
         }
         printf "%d,%s,%s,%s,%s\n", epoch, metric, value, split, lang
         offset += RSTART + RLENGTH - 1
         rest = substr(rest, RSTART + RLENGTH)
       }
     }' "$log_path" > "$out_path"
