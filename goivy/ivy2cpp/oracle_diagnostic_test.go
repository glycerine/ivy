package ivy2cpp

import (
	"strings"
	"testing"
)

func TestOracleGeneratorFallbackLabelsNormalizeOnlyActionLocationDifference(t *testing.T) {
	const py = `bool ext__step_gen::generate(probe &obj) {
 __ivy_out << "assumption_unsatisfied(\"step: action generator precondition cannot be satisfied\")" << std::endl;
 std::cerr << "step: error: action generator precondition cannot be satisfied\n";
 return false;
}`
	located := strings.ReplaceAll(py, "step:", "probe.ivy: line 10:")
	tests := []struct {
		name, goSource, pySource string
		equal                    bool
	}{
		{"action location fallback", located, py, true},
		{"matching source locations", located, located, true},
		{"different precondition locations", strings.ReplaceAll(located, "line 10", "line 11"), located, false},
		{"wrong action name", located, strings.ReplaceAll(py, "step:", "other:"), false},
		{"changed behavior", strings.ReplaceAll(located, "return false", "return true"), py, false},
		{"changed diagnostic", strings.ReplaceAll(located, "precondition cannot", "precondition must not"), py, false},
		{"missing diagnostic", strings.ReplaceAll(located, `std::cerr << "probe.ivy: line 10: error: action generator precondition cannot be satisfied\n";`, ""), py, false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			normalized := normalizeOracleGeneratorFallbackLabels(tc.goSource, tc.pySource)
			if got := CompareCPPTokens("Go", normalized, "Python", tc.pySource).Equal; got != tc.equal {
				t.Fatalf("comparison equality = %v, want %v\nnormalized:\n%s", got, tc.equal, normalized)
			}
		})
	}
}
