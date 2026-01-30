package gistselect

import (
	"math"
	"math/rand"
	"sort"
)

const (
	defaultEpsilon           = 1e-6
	defaultMaxThresholdPairs = 10000
	defaultNGramMin          = 2
	defaultNGramMax          = 3
	defaultWordWeight        = 1.0
	defaultIPAWeight         = 1.0
)

// Select runs the GIST algorithm and returns the best-scoring subset.
func Select(entries []Entry, cfg Config) Result {
	cfg = cfg.withDefaults(len(entries))
	if len(entries) == 0 || cfg.K == 0 {
		return Result{}
	}

	candidates := makeCandidates(entries, cfg)
	thresholds := cfg.Thresholds
	if len(thresholds) == 0 {
		thresholds = deriveThresholds(candidates, cfg)
	}
	if len(thresholds) == 0 {
		thresholds = []float64{0}
	}

	best := Result{Score: math.Inf(-1)}
	for _, threshold := range thresholds {
		selected, utility := greedyIndependentSet(candidates, cfg, threshold)
		minDist := minDistance(selected, cfg.Distance)
		score := utility + cfg.Lambda*minDist
		if isBetterScore(score, utility, threshold, best) {
			best = Result{
				Entries:     entriesFromCandidates(selected),
				Utility:     utility,
				MinDistance: minDist,
				Threshold:   threshold,
				Score:       score,
			}
		}
	}

	return best
}

func (cfg Config) withDefaults(entryCount int) Config {
	if cfg.K <= 0 {
		cfg.K = entryCount
	}
	if cfg.Distance == nil {
		cfg.Distance = DefaultDistance
	}
	if cfg.Utility == nil {
		cfg.Utility = CoverageUtility{}
	}
	if cfg.Epsilon <= 0 {
		cfg.Epsilon = defaultEpsilon
	}
	if cfg.MaxThresholdPairs <= 0 {
		cfg.MaxThresholdPairs = defaultMaxThresholdPairs
	}
	if cfg.NGramMin <= 0 {
		cfg.NGramMin = defaultNGramMin
	}
	if cfg.NGramMax < cfg.NGramMin {
		cfg.NGramMax = cfg.NGramMin
	}
	if cfg.WordFeatureWeight == 0 && cfg.IPAFeatureWeight == 0 {
		cfg.WordFeatureWeight = defaultWordWeight
		cfg.IPAFeatureWeight = defaultIPAWeight
	} else {
		if cfg.WordFeatureWeight < 0 {
			cfg.WordFeatureWeight = defaultWordWeight
		}
		if cfg.IPAFeatureWeight < 0 {
			cfg.IPAFeatureWeight = defaultIPAWeight
		}
	}
	return cfg
}

func makeCandidates(entries []Entry, cfg Config) []Candidate {
	candidates := make([]Candidate, len(entries))
	rng := rand.New(rand.NewSource(cfg.Seed))
	useSeed := cfg.Seed != 0
	for i, entry := range entries {
		tieBreak := int64(i)
		if useSeed {
			tieBreak = rng.Int63()
		}
		candidates[i] = Candidate{
			Entry:    entry,
			Index:    i,
			TieBreak: tieBreak,
		}
	}

	if _, ok := cfg.Utility.(CoverageUtility); ok {
		for i := range candidates {
			candidates[i].features = extractFeatures(candidates[i], cfg)
		}
	}

	return candidates
}

func deriveThresholds(candidates []Candidate, cfg Config) []float64 {
	if len(candidates) < 2 {
		return []float64{0}
	}
	maxPairs := cfg.MaxThresholdPairs
	distances := make([]float64, 0, maxPairs+1)
	distances = append(distances, 0)
	pairs := 0
	for i := 0; i < len(candidates); i++ {
		for j := i + 1; j < len(candidates); j++ {
			d := cfg.Distance(candidates[i].Entry, candidates[j].Entry)
			distances = append(distances, d)
			pairs++
			if pairs >= maxPairs {
				break
			}
		}
		if pairs >= maxPairs {
			break
		}
	}

	sort.Float64s(distances)
	unique := make([]float64, 0, len(distances))
	for _, d := range distances {
		if len(unique) == 0 || math.Abs(d-unique[len(unique)-1]) > cfg.Epsilon {
			unique = append(unique, d)
		}
	}
	return unique
}

func greedyIndependentSet(candidates []Candidate, cfg Config, threshold float64) ([]Candidate, float64) {
	utility := cfg.Utility.Init(cfg, candidates)
	selected := make([]Candidate, 0, cfg.K)
	used := make([]bool, len(candidates))

	for len(selected) < cfg.K {
		bestIdx := -1
		bestGain := 0.0
		for i, candidate := range candidates {
			if used[i] {
				continue
			}
			if !meetsThreshold(candidate, selected, cfg.Distance, threshold) {
				continue
			}
			gain := utility.Gain(candidate)
			if bestIdx == -1 || gain > bestGain+cfg.Epsilon || (math.Abs(gain-bestGain) <= cfg.Epsilon && tieBreak(candidate, candidates[bestIdx])) {
				bestIdx = i
				bestGain = gain
			}
		}

		if bestIdx == -1 || bestGain <= 0 {
			break
		}
		selected = append(selected, candidates[bestIdx])
		used[bestIdx] = true
		utility.Add(candidates[bestIdx])
	}

	return selected, utility.Value()
}

func meetsThreshold(candidate Candidate, selected []Candidate, distance DistanceFunc, threshold float64) bool {
	if threshold <= 0 || len(selected) == 0 {
		return true
	}
	for _, existing := range selected {
		if distance(candidate.Entry, existing.Entry) < threshold {
			return false
		}
	}
	return true
}

func minDistance(selected []Candidate, distance DistanceFunc) float64 {
	if len(selected) < 2 {
		return 0
	}
	min := math.Inf(1)
	for i := 0; i < len(selected); i++ {
		for j := i + 1; j < len(selected); j++ {
			d := distance(selected[i].Entry, selected[j].Entry)
			if d < min {
				min = d
			}
		}
	}
	if math.IsInf(min, 1) {
		return 0
	}
	return min
}

func entriesFromCandidates(selected []Candidate) []Entry {
	entries := make([]Entry, len(selected))
	for i, candidate := range selected {
		entries[i] = candidate.Entry
	}
	return entries
}

func tieBreak(a, b Candidate) bool {
	if a.TieBreak == b.TieBreak {
		return a.Index < b.Index
	}
	return a.TieBreak < b.TieBreak
}

func isBetterScore(score, utility, threshold float64, best Result) bool {
	if score > best.Score+defaultEpsilon {
		return true
	}
	if math.Abs(score-best.Score) <= defaultEpsilon {
		if utility > best.Utility+defaultEpsilon {
			return true
		}
		if math.Abs(utility-best.Utility) <= defaultEpsilon {
			return threshold < best.Threshold
		}
	}
	return false
}
