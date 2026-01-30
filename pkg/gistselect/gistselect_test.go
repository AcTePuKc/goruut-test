package gistselect

import "testing"

type countUtility struct{}

type countState struct {
	value float64
}

func (countUtility) Init(cfg Config, candidates []Candidate) Utility {
	return &countState{}
}

func (state *countState) Gain(candidate Candidate) float64 {
	return 1
}

func (state *countState) Add(candidate Candidate) {
	state.value += 1
}

func (state *countState) Value() float64 {
	return state.value
}

func TestSelectCoverageUtility(t *testing.T) {
	entries := []Entry{
		{Word: "ab", IPA: ""},
		{Word: "bc", IPA: ""},
	}
	cfg := Config{
		K:          2,
		Thresholds: []float64{0},
		NGramMin:   1,
		NGramMax:   1,
		Lambda:     0,
	}
	result := Select(entries, cfg)
	if len(result.Entries) != 2 {
		t.Fatalf("expected 2 entries selected, got %d", len(result.Entries))
	}
	if result.Utility != 3 {
		t.Fatalf("expected utility 3, got %.2f", result.Utility)
	}
}

func TestSelectDeterministicOrdering(t *testing.T) {
	entries := []Entry{
		{Word: "aa", IPA: "", Extra: "first"},
		{Word: "aa", IPA: "", Extra: "second"},
	}
	cfg := Config{
		K:          1,
		Thresholds: []float64{0},
		NGramMin:   1,
		NGramMax:   1,
		Lambda:     0,
	}
	result := Select(entries, cfg)
	if result.Entries[0].Extra != "first" {
		t.Fatalf("expected stable tie-break to select first entry")
	}

	cfg.Seed = 42
	first := Select(entries, cfg)
	second := Select(entries, cfg)
	if len(first.Entries) != len(second.Entries) || first.Entries[0] != second.Entries[0] {
		t.Fatalf("expected deterministic selection with seed")
	}
}

func TestThresholdConstraint(t *testing.T) {
	entries := []Entry{
		{Word: "a"},
		{Word: "b"},
		{Word: "c"},
	}
	distance := func(a, b Entry) float64 {
		if a.Word == b.Word {
			return 0
		}
		pair := a.Word + b.Word
		switch pair {
		case "ab", "ba":
			return 0.2
		case "ac", "ca", "bc", "cb":
			return 0.9
		default:
			return 1
		}
	}
	cfg := Config{
		K:          2,
		Thresholds: []float64{0.5},
		Lambda:     0,
		Distance:   distance,
		Utility:    countUtility{},
	}
	result := Select(entries, cfg)
	if len(result.Entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(result.Entries))
	}
	if result.Entries[0].Word != "a" || result.Entries[1].Word != "c" {
		t.Fatalf("expected entries [a c], got [%s %s]", result.Entries[0].Word, result.Entries[1].Word)
	}
}
