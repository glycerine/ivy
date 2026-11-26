package abci

import (
	"strings"
	"testing"
)

// constr detects functional constraints with Ssw_ManFindDirectImplications
// (values implied at the root by asserting every property output of the
// unrolled frames) and keeps the inductive ones (Saig_ManFilterUsingInd).
// The expected text is native ABC's ("constr -v").
func TestConstrDetectionLikeNative(t *testing.T) {
	for _, tc := range []struct{ bench, want string }{
		{"s713", `Found 102 candidates.
Level 0. Cands  =102    
Filtered cands:  0:153  0:160  0:161  0:168  0:169  0:176  0:177  
s713.bench      : pi =    89  po =   102  and =     389  lev =  43
Found   7 constraints after filtering.
Level 0. Constr =7    
`},
		{"s1494", `Found   0 candidates.
Filtered cands:  
s1494.bench     : pi =    22  po =     0  and =       0  lev =   0
`},
		{"s27", `Found   1 candidates.
Level 0. Cands  =1    
Filtered cands:  
s27.bench       : pi =    11  po =     1  and =      15  lev =   9
Level 0. Constr =0    
`},
	} {
		out, _ := runScriptStatus(t, "read test_vectors/seq/"+tc.bench+".blif; strash; constr -v")
		var kept []string
		for _, line := range strings.SplitAfter(out, "\n") {
			if line != "" && line != "\n" && !strings.Contains(line, "Skipping line") {
				kept = append(kept, line)
			}
		}
		if got := strings.Join(kept, ""); got != tc.want {
			t.Errorf("constr -v on %s:\n got:\n%s\nwant:\n%s", tc.bench, got, tc.want)
		}
	}
}
