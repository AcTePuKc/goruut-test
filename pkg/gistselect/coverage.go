package gistselect

// CoverageUtility implements a monotone submodular set coverage objective.
// It treats character n-grams from Word and IPA as features and assigns weights
// based on Config's feature weight settings.
type CoverageUtility struct{}

func (CoverageUtility) Init(cfg Config, candidates []Candidate) Utility {
	state := &coverageState{
		covered: make(map[string]struct{}),
	}
	return state
}

type coverageState struct {
	covered map[string]struct{}
	value   float64
}

func (state *coverageState) Gain(candidate Candidate) float64 {
	gain := 0.0
	for feature, weight := range candidate.features {
		if _, ok := state.covered[feature]; !ok {
			gain += weight
		}
	}
	return gain
}

func (state *coverageState) Add(candidate Candidate) {
	for feature, weight := range candidate.features {
		if _, ok := state.covered[feature]; !ok {
			state.covered[feature] = struct{}{}
			state.value += weight
		}
	}
}

func (state *coverageState) Value() float64 {
	return state.value
}

func extractFeatures(candidate Candidate, cfg Config) map[string]float64 {
	features := make(map[string]float64)
	addNGrams(features, "w:", candidate.Word, cfg.NGramMin, cfg.NGramMax, cfg.WordFeatureWeight)
	addNGrams(features, "i:", candidate.IPA, cfg.NGramMin, cfg.NGramMax, cfg.IPAFeatureWeight)
	return features
}

func addNGrams(features map[string]float64, prefix, text string, minN, maxN int, weight float64) {
	runes := []rune(text)
	if len(runes) == 0 {
		return
	}
	if minN < 1 {
		minN = 1
	}
	if maxN < minN {
		maxN = minN
	}
	for n := minN; n <= maxN; n++ {
		if n > len(runes) {
			continue
		}
		for i := 0; i+n <= len(runes); i++ {
			feature := prefix + string(runes[i:i+n])
			if _, exists := features[feature]; !exists {
				features[feature] = weight
			}
		}
	}
}
