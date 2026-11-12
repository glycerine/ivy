package psat

import (
	"os"
	"regexp"
	"strings"
	"testing"
)

// This is a source-contract audit, not an original-C differential test: this
// checkout supplies declarations and consumers, but no MiniSat-1.14p bodies.
func TestOriginalHeaderAPICoverage(t *testing.T) {
	header, err := os.ReadFile("../../../src/sat/psat/m114p.h")
	if os.IsNotExist(err) {
		header, err = os.ReadFile("test_vectors/m114p.h")
	}
	if err != nil {
		t.Fatal(err)
	}
	compat, err := os.ReadFile("compat.go")
	if err != nil {
		t.Fatal(err)
	}
	pattern := regexp.MustCompile(`extern\s+[^;]+?\b(M114p_Solver\w+)\s*\(`)
	names := pattern.FindAllSubmatch(header, -1)
	if len(names) != 15 {
		t.Fatal("header contract changed", len(names))
	}
	for _, name := range names {
		if !strings.Contains(string(compat), "func "+string(name[1])+"(") {
			t.Fatalf("missing original API %s", name[1])
		}
	}
	module, err := os.ReadFile("../../../src/sat/psat/module.make")
	if os.IsNotExist(err) {
		return
	}
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(string(module)) != "SRC +=" {
		t.Log("original source manifest changed; recheck availability of a native MiniSat oracle")
	}
}
