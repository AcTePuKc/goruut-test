package metricsextract

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
)

// Record captures one metric entry extracted from a log file.
type Record struct {
	Epoch    int
	Metric   string
	Value    float64
	Split    string
	Language string
}

// Options controls how metrics are parsed from log lines.
type Options struct {
	EpochRegex    *regexp.Regexp
	SplitRegex    *regexp.Regexp
	KVRegex       *regexp.Regexp
	LanguageRegex *regexp.Regexp
	Language      string
}

// DefaultOptions returns parser defaults for epoch/train/eval-style logs.
func DefaultOptions() Options {
	return Options{
		EpochRegex: regexp.MustCompile(`(?i)\bepoch\b\s*[:=]?\s*(\d+)`),
		SplitRegex: regexp.MustCompile(`(?i)\b(train|eval|validation|valid|dev|test)\b`),
		KVRegex:    regexp.MustCompile(`(?i)([a-z][a-z0-9_.-]*)\s*[:=]\s*([0-9]+(?:\.[0-9]+)?)%?`),
		LanguageRegex: regexp.MustCompile(
			`(?i)(?:\blang(?:uage)?\b\s*[:=]\s*([a-z0-9_-]+)|\bfor\s+([a-z0-9_-]+))`,
		),
	}
}

// Parser extracts epoch metrics from log lines.
type Parser struct {
	opts      Options
	autoEpoch int
}

// NewParser constructs a Parser with the provided options.
func NewParser(opts Options) *Parser {
	defaults := DefaultOptions()
	if opts.EpochRegex == nil {
		opts.EpochRegex = defaults.EpochRegex
	}
	if opts.SplitRegex == nil {
		opts.SplitRegex = defaults.SplitRegex
	}
	if opts.KVRegex == nil {
		opts.KVRegex = defaults.KVRegex
	}
	if opts.LanguageRegex == nil {
		opts.LanguageRegex = defaults.LanguageRegex
	}
	if opts.Language == "" {
		opts.Language = defaults.Language
	}
	return &Parser{opts: opts}
}

// Parse reads logs from reader and emits metric records.
func (p *Parser) Parse(reader io.Reader) ([]Record, error) {
	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 0, 64*1024), 2*1024*1024)
	var records []Record
	for scanner.Scan() {
		line := scanner.Text()
		lineRecords, err := p.parseLine(line)
		if err != nil {
			return nil, err
		}
		records = append(records, lineRecords...)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return records, nil
}

func (p *Parser) parseLine(line string) ([]Record, error) {
	epoch := -1
	if matches := p.opts.EpochRegex.FindStringSubmatch(line); len(matches) > 1 {
		parsed, err := strconv.Atoi(matches[1])
		if err != nil {
			return nil, fmt.Errorf("parse epoch %q: %w", matches[1], err)
		}
		epoch = parsed
		if parsed > p.autoEpoch {
			p.autoEpoch = parsed
		}
	}

	splitTokens := p.splitTokenPositions(line)
	kvMatches := p.opts.KVRegex.FindAllStringSubmatchIndex(line, -1)
	if len(kvMatches) == 0 {
		return nil, nil
	}

	language := p.opts.Language
	if language == "" {
		language = p.extractLanguage(line)
	}

	var results []Record
	lineEpochAssigned := epoch >= 0
	for _, match := range kvMatches {
		key := strings.ToLower(line[match[2]:match[3]])
		valueText := line[match[4]:match[5]]
		value, err := strconv.ParseFloat(valueText, 64)
		if err != nil {
			return nil, fmt.Errorf("parse value %q: %w", valueText, err)
		}

		split := splitFromKey(key)
		metric := key
		if split != "" {
			metric = trimSplitPrefix(metric, split)
		} else {
			split = splitFromPosition(match[0], splitTokens)
			if split != "" {
				split = normalizeSplit(split)
			}
		}
		if split == "" {
			continue
		}
		metric = strings.Trim(metric, "_.-")
		if metric == "" || metric == split {
			metric = "value"
		}

		if !lineEpochAssigned {
			p.autoEpoch++
			epoch = p.autoEpoch
			lineEpochAssigned = true
		}

		results = append(results, Record{
			Epoch:    epoch,
			Metric:   metric,
			Value:    value,
			Split:    split,
			Language: language,
		})
	}

	return results, nil
}

func (p *Parser) splitTokenPositions(line string) []splitToken {
	matches := p.opts.SplitRegex.FindAllStringIndex(line, -1)
	results := make([]splitToken, 0, len(matches))
	for _, match := range matches {
		token := strings.ToLower(line[match[0]:match[1]])
		results = append(results, splitToken{pos: match[0], token: normalizeSplit(token)})
	}
	return results
}

func (p *Parser) extractLanguage(line string) string {
	if p.opts.LanguageRegex == nil {
		return ""
	}
	matches := p.opts.LanguageRegex.FindStringSubmatch(line)
	if len(matches) == 0 {
		return ""
	}
	for i := 1; i < len(matches); i++ {
		if matches[i] != "" {
			return strings.ToLower(matches[i])
		}
	}
	return ""
}

type splitToken struct {
	pos   int
	token string
}

func splitFromPosition(pos int, tokens []splitToken) string {
	var selected string
	for _, token := range tokens {
		if token.pos <= pos {
			selected = token.token
		} else {
			break
		}
	}
	return selected
}

func splitFromKey(key string) string {
	for _, prefix := range []string{"train", "eval", "validation", "valid", "dev", "test"} {
		if key == prefix {
			return normalizeSplit(prefix)
		}
		if strings.HasPrefix(key, prefix+"_") || strings.HasPrefix(key, prefix+"-") || strings.HasPrefix(key, prefix+".") {
			return normalizeSplit(prefix)
		}
	}
	return ""
}

func trimSplitPrefix(key, split string) string {
	prefixes := []string{split + "_", split + "-", split + "."}
	for _, prefix := range prefixes {
		if strings.HasPrefix(key, prefix) {
			return strings.TrimPrefix(key, prefix)
		}
	}
	return key
}

func normalizeSplit(split string) string {
	switch split {
	case "validation", "valid", "dev", "test":
		return "eval"
	default:
		return split
	}
}
