package main

import (
	"os"
	"path/filepath"
	"testing"
)

func writeTempTSV(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "data.tsv")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write tsv: %v", err)
	}
	return path
}

func TestLoadWithoutGistFilter(t *testing.T) {
	path := writeTempTSV(t, "alpha\talfa\nbeta\tbeta\ngamma\tgamma\n")
	entries := load(path, -1, gistFilterConfig{})
	if got, want := len(entries), 3; got != want {
		t.Fatalf("expected %d entries, got %d", want, got)
	}
}

func TestLoadWithGistFilter(t *testing.T) {
	path := writeTempTSV(t, "alpha\talfa\nbeta\tbeta\ngamma\tgamma\n")
	entries := load(path, -1, gistFilterConfig{
		enabled: true,
		selectK: 2,
		lambda:  1.0,
	})
	if got, want := len(entries), 2; got != want {
		t.Fatalf("expected %d entries, got %d", want, got)
	}
}
