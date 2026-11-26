package goivy

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInvariantUsingPatterns(t *testing.T) {
	for _, tc := range []struct {
		fixture, filter, wantError string
	}{
		{"usingpat1", "", ""},
		{"usingpat2", "", "failed checks: 1"},
		{"usingpat3", "", "failed checks: 2"},
		{"usingpat4", "", "failed checks: 1"},
		{"usingpat5", "", ""},
		{"patdef1", "", "redefining x"},
		{"patdef2", "", ""},
		{"patdef3", "", ""},
		{"derived4", "", ""},
		{"derived5", "", ""},
		{"derived5", "m.c", ""},
	} {
		name := tc.fixture
		if tc.filter != "" {
			name += "/" + tc.filter
		}
		t.Run(name, func(t *testing.T) {
			filename := filepath.Join("test_vectors", "invariant_patterns", tc.fixture+".ivy")
			source, err := os.ReadFile(filename)
			if err != nil {
				t.Fatal(err)
			}
			cfg := NewConfig()
			if err := ApplyIvyCheckParams(cfg, map[string]string{"check": tc.filter}); err != nil {
				t.Fatal(err)
			}
			err = StartSourceWithConfig(filename, string(source), cfg)
			if tc.wantError == "" {
				if err != nil {
					t.Fatal(err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tc.wantError) {
				t.Fatalf("got %v, want an error containing %q", err, tc.wantError)
			}
		})
	}
}
