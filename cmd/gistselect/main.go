package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/neurlang/goruut/pkg/gistselect"
)

const defaultEpsilon = 1e-6

func main() {
	inPath := flag.String("in", "", "input TSV path (word<TAB>ipa[,<TAB>tags]) or '-' for stdin")
	outPath := flag.String("out", "", "output TSV path or '-' for stdout")
	k := flag.Int("k", 0, "number of entries to select (default: all)")
	lambda := flag.Float64("lambda", 1.0, "weight for minimum distance term")
	epsilon := flag.Float64("epsilon", defaultEpsilon, "tie-breaking epsilon for gains/distances")
	distance := flag.String("distance", "joint", "distance mode: word-levenshtein, ipa-levenshtein, joint")
	utility := flag.String("utility", "joint", "utility mode: word-trigram, ipa-ngram, joint")
	maxCandidates := flag.Int("max_candidates", 0, "limit number of candidates considered (0 = no limit)")
	seed := flag.Int64("seed", 0, "deterministic tie-break seed (0 = stable input order)")
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: gistselect --in INPUT --out OUTPUT [options]\n\nOptions:\n")
		flag.PrintDefaults()
		fmt.Fprintf(flag.CommandLine.Output(), "\nExamples:\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  gistselect --in dicts/en/lexicon.tsv --out dicts/en/lexicon.gist.tsv --k 200 --distance word-levenshtein --utility word-trigram\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  gistselect --in dicts/en/learn.tsv --out dicts/en/learn.gist.tsv --k 500 --lambda 0.5 --utility joint\n")
	}
	flag.Parse()

	if *inPath == "" || *outPath == "" {
		exitWithError(errors.New("both --in and --out are required"))
	}

	readStart := time.Now()
	entries, err := readTSV(*inPath)
	if err != nil {
		exitWithError(err)
	}
	if *maxCandidates > 0 && len(entries) > *maxCandidates {
		entries = entries[:*maxCandidates]
	}
	readDuration := time.Since(readStart)

	cfg := gistselect.Config{
		K:       *k,
		Lambda:  *lambda,
		Epsilon: *epsilon,
		Seed:    *seed,
	}

	cfg.Distance, err = parseDistance(*distance)
	if err != nil {
		exitWithError(err)
	}
	if err := applyUtility(&cfg, *utility); err != nil {
		exitWithError(err)
	}

	selectStart := time.Now()
	result := gistselect.Select(entries, cfg)
	selectDuration := time.Since(selectStart)

	writeStart := time.Now()
	if err := writeTSV(*outPath, result.Entries); err != nil {
		exitWithError(err)
	}
	writeDuration := time.Since(writeStart)

	totalDuration := readDuration + selectDuration + writeDuration
	fmt.Fprintf(os.Stderr,
		"gistselect summary: input=%d output=%d threshold=%.6f min_distance=%.6f utility=%.6f score=%.6f timings(read=%s select=%s write=%s total=%s)\n",
		len(entries), len(result.Entries), result.Threshold, result.MinDistance, result.Utility, result.Score,
		readDuration, selectDuration, writeDuration, totalDuration,
	)
}

func readTSV(path string) ([]gistselect.Entry, error) {
	reader, closeFn, err := openInput(path)
	if err != nil {
		return nil, err
	}
	if closeFn != nil {
		defer closeFn()
	}

	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 0, 64*1024), 10*1024*1024)
	var entries []gistselect.Entry
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}
		columns := strings.Split(line, "\t")
		if len(columns) < 2 || len(columns) > 3 {
			return nil, fmt.Errorf("invalid TSV row (expected 2 or 3 columns): %q", line)
		}
		entry := gistselect.Entry{
			Word: columns[0],
			IPA:  columns[1],
		}
		if len(columns) == 3 {
			entry.Extra = columns[2]
		}
		entries = append(entries, entry)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read input: %w", err)
	}
	return entries, nil
}

func writeTSV(path string, entries []gistselect.Entry) error {
	writer, closeFn, err := openOutput(path)
	if err != nil {
		return err
	}
	if closeFn != nil {
		defer closeFn()
	}
	buffered := bufio.NewWriter(writer)
	for _, entry := range entries {
		if entry.Extra != "" {
			if _, err := fmt.Fprintf(buffered, "%s\t%s\t%s\n", entry.Word, entry.IPA, entry.Extra); err != nil {
				return fmt.Errorf("write output: %w", err)
			}
			continue
		}
		if _, err := fmt.Fprintf(buffered, "%s\t%s\n", entry.Word, entry.IPA); err != nil {
			return fmt.Errorf("write output: %w", err)
		}
	}
	if err := buffered.Flush(); err != nil {
		return fmt.Errorf("flush output: %w", err)
	}
	return nil
}

func openInput(path string) (io.Reader, func() error, error) {
	if path == "-" {
		return os.Stdin, nil, nil
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, nil, fmt.Errorf("open input: %w", err)
	}
	return file, file.Close, nil
}

func openOutput(path string) (io.Writer, func() error, error) {
	if path == "-" {
		return os.Stdout, nil, nil
	}
	file, err := os.Create(path)
	if err != nil {
		return nil, nil, fmt.Errorf("open output: %w", err)
	}
	return file, file.Close, nil
}

func parseDistance(mode string) (gistselect.DistanceFunc, error) {
	switch normalizeMode(mode) {
	case "word", "wordlevenshtein":
		return gistselect.WordLevenshtein, nil
	case "ipa", "ipalevenshtein":
		return gistselect.IPALevenshtein, nil
	case "joint":
		return gistselect.JointLevenshtein, nil
	default:
		return nil, fmt.Errorf("unknown distance mode %q", mode)
	}
}

func applyUtility(cfg *gistselect.Config, mode string) error {
	cfg.Utility = gistselect.CoverageUtility{}
	switch normalizeMode(mode) {
	case "wordtrigram":
		cfg.NGramMin = 3
		cfg.NGramMax = 3
		cfg.WordFeatureWeight = 1
		cfg.IPAFeatureWeight = 0
	case "ipangram":
		cfg.NGramMin = 2
		cfg.NGramMax = 3
		cfg.WordFeatureWeight = 0
		cfg.IPAFeatureWeight = 1
	case "joint":
		cfg.NGramMin = 2
		cfg.NGramMax = 3
		cfg.WordFeatureWeight = 1
		cfg.IPAFeatureWeight = 1
	default:
		return fmt.Errorf("unknown utility mode %q", mode)
	}
	return nil
}

func normalizeMode(mode string) string {
	mode = strings.ToLower(mode)
	mode = strings.ReplaceAll(mode, "-", "")
	mode = strings.ReplaceAll(mode, "_", "")
	return mode
}

func exitWithError(err error) {
	fmt.Fprintln(os.Stderr, "gistselect:", err)
	os.Exit(1)
}
