package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"os"
	"regexp"
	"sort"

	"github.com/neurlang/goruut/pkg/metricsextract"
)

type csvRow struct {
	Epoch    int
	Metric   string
	Value    float64
	Split    string
	Language string
}

func main() {
	logPath := flag.String("log", "", "path to trainer log file")
	outPath := flag.String("out", "", "path to output CSV")
	language := flag.String("language", "", "language name for rows missing language tags")
	epochRegex := flag.String("epoch-regex", "", "override epoch regex")
	splitRegex := flag.String("split-regex", "", "override split regex")
	kvRegex := flag.String("kv-regex", "", "override key/value regex")
	langRegex := flag.String("language-regex", "", "override language regex")
	flag.Parse()

	if *logPath == "" || *outPath == "" {
		fmt.Fprintln(os.Stderr, "Usage: metrics_extract --log <path> --out <path> [--language <lang>] [--epoch-regex <regex>] [--split-regex <regex>] [--kv-regex <regex>] [--language-regex <regex>]")
		os.Exit(1)
	}

	logFile, err := os.Open(*logPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "open log: %v\n", err)
		os.Exit(1)
	}
	defer logFile.Close()

	opts := metricsextract.DefaultOptions()
	opts.Language = *language
	if *epochRegex != "" {
		opts.EpochRegex, err = regexp.Compile(*epochRegex)
		if err != nil {
			fmt.Fprintf(os.Stderr, "compile epoch regex: %v\n", err)
			os.Exit(1)
		}
	}
	if *splitRegex != "" {
		opts.SplitRegex, err = regexp.Compile(*splitRegex)
		if err != nil {
			fmt.Fprintf(os.Stderr, "compile split regex: %v\n", err)
			os.Exit(1)
		}
	}
	if *kvRegex != "" {
		opts.KVRegex, err = regexp.Compile(*kvRegex)
		if err != nil {
			fmt.Fprintf(os.Stderr, "compile kv regex: %v\n", err)
			os.Exit(1)
		}
	}
	if *langRegex != "" {
		opts.LanguageRegex, err = regexp.Compile(*langRegex)
		if err != nil {
			fmt.Fprintf(os.Stderr, "compile language regex: %v\n", err)
			os.Exit(1)
		}
	}

	parser := metricsextract.NewParser(opts)
	records, err := parser.Parse(logFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "parse log: %v\n", err)
		os.Exit(1)
	}

	rows := make([]csvRow, 0, len(records))
	for _, record := range records {
		rows = append(rows, csvRow{
			Epoch:    record.Epoch,
			Metric:   record.Metric,
			Value:    record.Value,
			Split:    record.Split,
			Language: record.Language,
		})
	}
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].Epoch != rows[j].Epoch {
			return rows[i].Epoch < rows[j].Epoch
		}
		if rows[i].Split != rows[j].Split {
			return rows[i].Split < rows[j].Split
		}
		if rows[i].Metric != rows[j].Metric {
			return rows[i].Metric < rows[j].Metric
		}
		return rows[i].Value < rows[j].Value
	})

	outFile, err := os.Create(*outPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "create output: %v\n", err)
		os.Exit(1)
	}
	defer outFile.Close()

	writer := csv.NewWriter(outFile)
	if err := writer.Write([]string{"epoch", "metric", "value", "split", "language"}); err != nil {
		fmt.Fprintf(os.Stderr, "write csv header: %v\n", err)
		os.Exit(1)
	}
	for _, row := range rows {
		if err := writer.Write([]string{
			fmt.Sprintf("%d", row.Epoch),
			row.Metric,
			fmt.Sprintf("%g", row.Value),
			row.Split,
			row.Language,
		}); err != nil {
			fmt.Fprintf(os.Stderr, "write csv row: %v\n", err)
			os.Exit(1)
		}
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		fmt.Fprintf(os.Stderr, "flush csv: %v\n", err)
		os.Exit(1)
	}
}
