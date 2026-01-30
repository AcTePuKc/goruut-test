package gistselect

// Entry represents a dictionary row with optional extra metadata.
type Entry struct {
	Word  string
	IPA   string
	Extra string
}

// DistanceFunc computes a distance in [0, +inf) where larger means more diverse.
type DistanceFunc func(a, b Entry) float64

// UtilityFunc creates a stateful utility tracker for computing marginal gains.
// Implementations should be monotone: adding more entries never decreases Value().
type UtilityFunc interface {
	Init(cfg Config, candidates []Candidate) Utility
}

// Utility tracks marginal gains for a particular selection.
type Utility interface {
	Gain(candidate Candidate) float64
	Add(candidate Candidate)
	Value() float64
}

// Config controls GIST selection behavior.
type Config struct {
	K                 int
	Lambda            float64
	Epsilon           float64
	Thresholds        []float64
	Seed              int64
	Distance          DistanceFunc
	Utility           UtilityFunc
	MaxThresholdPairs int
	NGramMin          int
	NGramMax          int
	WordFeatureWeight float64
	IPAFeatureWeight  float64
}

// Candidate is an entry with deterministic ordering metadata and cached features.
type Candidate struct {
	Entry
	Index    int
	TieBreak int64
	features map[string]float64
}

// Result contains the selected entries and score components.
type Result struct {
	Entries     []Entry
	Utility     float64
	MinDistance float64
	Threshold   float64
	Score       float64
}
