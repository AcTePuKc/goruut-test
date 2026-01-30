// Package gistselect implements a deterministic, standard-library-only
// GIST-style selector for compact, diverse lexicons.
//
// The selector maximizes f(S)=g(S)+λ*minDist(S) where g(S) is a monotone
// submodular coverage utility and minDist(S) is the minimum pairwise distance
// among selected entries. It sweeps over candidate distance thresholds and, for
// each threshold, builds a greedy independent set that maximizes marginal gain
// subject to minDist >= threshold. The best-scoring set is returned.
//
// Complexity:
//   - Threshold discovery: O(p) distance evaluations where p is capped by
//     Config.MaxThresholdPairs (default 10k).
//   - Each threshold sweep: O(k*n) marginal gain evaluations plus O(k*n)
//     distance checks for the independence constraint.
//
// Determinism:
//   - Input order is preserved via stable tie-breaking by index.
//   - If Config.Seed != 0, deterministic pseudo-random tie-breaking is used for
//     equal marginal gains.
package gistselect
