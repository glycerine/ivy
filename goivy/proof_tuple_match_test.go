package goivy

import (
	"strings"
	"testing"
)

// Python apply_match_alt on add_prem_match's ia.Tuple runs alpha_avoid and
// apply_match_alt_rec over the tuple, rebuilding (cloning) each premise
// LabeledFormula twice, and traces the tuple canon (test/gcd2.ivy).
func TestApplyMatchTupleClonesPremisesLikePython(t *testing.T) {
	cfg := NewAstConfig()
	p := NewConst("p", Boolean)
	q := NewConst("q", Boolean)
	prem := cfg.NewLabeledFormula(cfg.NewAtom("prem1"), p)
	var res []Node
	out := captureActionUpdateStdout(t, func() {
		res = applyMatchTuple(map[NodeKey]Expr{}, []Node{prem, q}, true)
	})
	if n := strings.Count(out, "ast.LF.clone PRESERVE"); n != 2 {
		t.Fatalf("LF clones %d, want 2:\n%s", n, out)
	}
	if !strings.Contains(out, "proof.ApplyMatchAlt ENTER nmatch=0 fmlaType=Tuple") ||
		!strings.Contains(out, "proof.ApplyMatchAlt EXIT HASH canon=(tuple elems:[(labeledFormula") {
		t.Fatalf("tuple traces missing:\n%s", out)
	}
	if len(res) != 2 {
		t.Fatalf("elems %d", len(res))
	}
	if _, ok := res[0].(*LabeledFormula); !ok {
		t.Fatalf("premise became %T", res[0])
	}
}

// Python goal_apply_to_prem looks the premise up in
// dict((x.name,idx) for idx,x in enumerate(prems)), so with a repeated
// premise name the last one is used (test/gcd2.ivy).
func TestGoalApplyToPremUsesLastDuplicate(t *testing.T) {
	cfg := NewAstConfig()
	p := NewConst("p", Boolean)
	q := NewConst("q", Boolean)
	first := cfg.NewLabeledFormula(cfg.NewAtom("prem"), p)
	last := cfg.NewLabeledFormula(cfg.NewAtom("prem"), q)
	goal := cfg.NewLabeledFormula(cfg.NewAtom("g"), cfg.NewSchemaBody(first, last, True))
	var got Node
	GoalApplyToPrem(cfg, goal, "prem", func(lf *LabeledFormula) *LabeledFormula {
		got = lf.Formula
		return lf
	})
	if got != q {
		t.Fatalf("applied to %v, want the last prem (q)", got)
	}
}
