package goivy

import (
	"strings"
	"testing"
)

// Python property tactic (ivy_proof.py):
//
//	vmap = dict((x.name,x) for x in lu.variables_ast(fmla))
//	for x in vmap:
//	    if x not in used:
//	        raise IvyError(lhs,'{} must be a parameter of {}'.format(x,lhs.rep))
//
// variables_ast yields the free variables in order (X, then Z, for
// exists Y. P(X,Z,Y)), so with no parameters the error names X, as an
// IvyError located at lhs. goivy iterated a Go map and returned a
// ProofError without location.
func TestPropertyTacticReportsFirstMissingParameterAtLhs(t *testing.T) {
	for i := 0; i < 30; i++ {
		cfg := NewAstConfig()
		sSort := &UninterpretedSort{Name: "S"}
		sig := NewSig()
		sig.Sorts.Set("S", sSort)
		pSort, _ := NewFunctionSort(sSort, sSort, sSort, Boolean)
		sig.Symbols.Set("P", &SymbolEntry{Name: "P", Sort: pSort})
		mod := NewWithSig(sig)
		goal := mkLF(cfg.NewAtom("g"), True)
		xv, zv, yv := cfg.NewVariable("X", "S"), cfg.NewVariable("Z", "S"), cfg.NewVariable("Y", "S")
		fa := cfg.NewForall([]Node{xv}, cfg.NewForall([]Node{zv}, cfg.NewExists([]Node{yv}, cfg.NewAtom("P", xv, zv, yv))))
		propLF := cfg.NewLabeledFormula(cfg.NewAtom("prop"), fa)
		lhs := cfg.NewAtom("f")
		lhs.SetLineno(Location{Filename: "f.ivy", Line: 12})
		pc := mkPCWithModule(mod)
		_, err := pc.propertyTactic([]*LabeledFormula{goal}, mkPropertyTacticNode(cfg, propLF, lhs, nil))
		if err == nil || !strings.HasPrefix(err.Error(), "f.ivy: line 12: error: X must be a parameter of f") {
			t.Fatalf("run %d: got %v, want X first, at lhs", i, err)
		}
	}
}
