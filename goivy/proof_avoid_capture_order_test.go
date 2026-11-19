package goivy

import (
	"strings"
	"testing"
)

// Python avoid_capture_problem / rename_problem (ivy_proof.py):
//
//	cmatch = dict((v,v.rename(rn)) for v in prob.freesyms
//	              if v.name in matchnames and v not in match)
//	prob.revmap.update((y,x) for x,y in match.items())
//
// prob.freesyms is now an ordered dict (was a set), so free symbols b then
// a are renamed in that order, revmap maps each renamed symbol back to its
// original in that order, and detect_nonce_symbols names the first
// original: 'Symbol {} in schema clashes with {} in goal.\nSuggest
// renaming or instantiating it.'. goivy used Go maps and stored nil
// originals.
func TestAvoidCaptureProblemOrderAndRevmap(t *testing.T) {
	for i := 0; i < 30; i++ {
		b, a := NewConst("b", Boolean), NewConst("a", Boolean)
		fs := NewInsMap[NodeKey, Expr]()
		fs.Set(Key(b), b)
		fs.Set(Key(a), a)
		prob := NewMatchProblem(nil, True, True, fs, nil)
		f := NewConst("f", Boolean)
		match := map[NodeKey]Expr{Key(f): &LogicAnd{Terms: []Expr{NewConst("a", Boolean), NewConst("b", Boolean)}}}

		AvoidCaptureProblem(NewAstConfig(), prob, match)

		var origs []string
		for _, o := range prob.RevMap.All() {
			c, ok := o.(*Const)
			if !ok {
				t.Fatalf("run %d: revmap value %#v, want the original symbol", i, o)
			}
			origs = append(origs, c.Name)
		}
		if len(origs) != 2 || origs[0] != "b" || origs[1] != "a" {
			t.Fatalf("run %d: revmap originals %v, want [b a]", i, origs)
		}
		err := DetectNonceSymbols(prob)
		if err == nil || !strings.Contains(err.Error(), "Symbol b in schema clashes with b in goal.\nSuggest renaming or instantiating it.") {
			t.Fatalf("run %d: got %v", i, err)
		}
	}
}
