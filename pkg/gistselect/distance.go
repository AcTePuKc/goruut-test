package gistselect

import "math"

// DefaultDistance uses normalized Levenshtein distance on "Word|IPA" strings.
func DefaultDistance(a, b Entry) float64 {
	return normalizedLevenshtein(a.Word+"|"+a.IPA, b.Word+"|"+b.IPA)
}

// WordLevenshtein uses normalized Levenshtein distance on words only.
func WordLevenshtein(a, b Entry) float64 {
	return normalizedLevenshtein(a.Word, b.Word)
}

// IPALevenshtein uses normalized Levenshtein distance on IPA only.
func IPALevenshtein(a, b Entry) float64 {
	return normalizedLevenshtein(a.IPA, b.IPA)
}

// JointLevenshtein uses normalized Levenshtein distance on "Word|IPA" strings.
func JointLevenshtein(a, b Entry) float64 {
	return normalizedLevenshtein(a.Word+"|"+a.IPA, b.Word+"|"+b.IPA)
}

func normalizedLevenshtein(left, right string) float64 {
	if left == right {
		return 0
	}
	maxLen := max(len([]rune(left)), len([]rune(right)))
	if maxLen == 0 {
		return 0
	}
	distance := levenshtein([]rune(left), []rune(right))
	return float64(distance) / float64(maxLen)
}

func levenshtein(a, b []rune) int {
	if len(a) == 0 {
		return len(b)
	}
	if len(b) == 0 {
		return len(a)
	}
	if len(a) < len(b) {
		return levenshtein(b, a)
	}

	previous := make([]int, len(b)+1)
	current := make([]int, len(b)+1)
	for j := 0; j <= len(b); j++ {
		previous[j] = j
	}

	for i := 1; i <= len(a); i++ {
		current[0] = i
		for j := 1; j <= len(b); j++ {
			cost := 0
			if a[i-1] != b[j-1] {
				cost = 1
			}
			current[j] = minInt(
				current[j-1]+1,
				previous[j]+1,
				previous[j-1]+cost,
			)
		}
		copy(previous, current)
	}
	return previous[len(b)]
}

func minInt(values ...int) int {
	min := math.MaxInt
	for _, v := range values {
		if v < min {
			min = v
		}
	}
	return min
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
