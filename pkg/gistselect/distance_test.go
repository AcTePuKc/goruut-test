package gistselect

import "testing"

func TestDistanceModes(t *testing.T) {
	a := Entry{Word: "cat", IPA: "kæt"}
	b := Entry{Word: "bat", IPA: "kæt"}

	word := WordLevenshtein(a, b)
	if word <= 0 {
		t.Fatalf("expected word distance to be positive, got %.3f", word)
	}

	ipa := IPALevenshtein(a, b)
	if ipa != 0 {
		t.Fatalf("expected IPA distance 0, got %.3f", ipa)
	}

	joint := JointLevenshtein(a, b)
	if joint <= 0 {
		t.Fatalf("expected joint distance to be positive, got %.3f", joint)
	}
}
