package metricsextract

import (
	"strings"
	"testing"
)

func TestParserEpochTrainEvalLine(t *testing.T) {
	parser := NewParser(DefaultOptions())
	log := "Epoch 3 train loss=0.25 acc=0.9 eval loss=0.5 acc=0.8"
	records, err := parser.Parse(strings.NewReader(log))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(records) != 4 {
		t.Fatalf("expected 4 records, got %d", len(records))
	}
	for _, record := range records {
		if record.Epoch != 3 {
			t.Fatalf("expected epoch 3, got %d", record.Epoch)
		}
		if record.Split != "train" && record.Split != "eval" {
			t.Fatalf("unexpected split %q", record.Split)
		}
	}
}

func TestParserAutoEpochFallback(t *testing.T) {
	parser := NewParser(DefaultOptions())
	log := strings.Join([]string{
		"train loss=1.2 eval loss=1.0",
		"train loss=0.9 eval loss=0.7",
	}, "\n")
	records, err := parser.Parse(strings.NewReader(log))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(records) != 4 {
		t.Fatalf("expected 4 records, got %d", len(records))
	}
	epochByLine := map[int]int{}
	for _, record := range records {
		epochByLine[record.Epoch]++
	}
	if epochByLine[1] != 2 || epochByLine[2] != 2 {
		t.Fatalf("expected auto epochs 1 and 2 to have 2 records each, got %+v", epochByLine)
	}
}

func TestParserSplitPrefixMetrics(t *testing.T) {
	parser := NewParser(DefaultOptions())
	log := "epoch=5 train_acc=0.87 eval_acc=0.81"
	records, err := parser.Parse(strings.NewReader(log))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(records) != 2 {
		t.Fatalf("expected 2 records, got %d", len(records))
	}
	for _, record := range records {
		if record.Metric != "acc" {
			t.Fatalf("expected metric acc, got %q", record.Metric)
		}
	}
}
