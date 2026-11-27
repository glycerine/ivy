package goivy

import (
	"slices"
	"testing"

	"github.com/glycerine/ivy/goivy/xtracer"
)

// Python find_references (ivy_isolate.py) collects lineno -> None in an
// insertion-ordered dict (it was a set): formulas first, then actions, in
// module order, keyed by the full lineno (file and line). goivy used a Go
// map[int]bool: random order, and two files' same line number collapsed.
func TestFindReferencesOrderAndFullLocation(t *testing.T) {
	m := mkModuleWithSig()
	s := isolateMkSort("T")
	p := NewConst("p", s)
	pred, err := NewFunctionSort(s, Boolean)
	if err != nil {
		t.Fatal(err)
	}
	f := MustApply(NewConst("f", pred), p)
	acfg := NewAstConfig()
	mkLF := func(file string, line int) *LabeledFormula {
		lf := acfg.NewLabeledFormula(acfg.NewAtom("l"), f)
		lf.Base.Loc = Location{Filename: file, Line: line}
		return lf
	}
	m.LabeledAxioms = []*LabeledFormula{mkLF("b.ivy", 7), mkLF("a.ivy", 7)}
	m.LabeledConjs = []*LabeledFormula{mkLF("a.ivy", 3)}

	refs := FindReferences(m, map[string]bool{"p": true}, NewInsMap[string, ActionsAction]())
	var got []string
	for _, loc := range refs.All() {
		got = append(got, loc.String())
	}
	want := []string{"b.ivy: line 7: ", "a.ivy: line 7: ", "a.ivy: line 3: "}
	if len(got) != len(want) {
		t.Fatalf("got %q, want %q", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %q, want %q", got, want)
		}
	}
}

func TestFindReferencesTracesOnlyActionSymbols(t *testing.T) {
	m := mkModuleWithSig()
	acfg := NewAstConfig()
	state := NewConst("state", Boolean)
	formulaLoc := Location{Filename: "example.ivy", Line: 3}
	lf := acfg.NewLabeledFormula(acfg.NewAtom("fact"), state)
	lf.Base.Loc = formulaLoc
	m.LabeledAxioms = []*LabeledFormula{lf}

	actionLoc := Location{Filename: "example.ivy", Line: 7}
	action := NewSequence(NewAssumeAction(NewConst("guard", Boolean)), NewAssumeAction(state))
	action.SetLineno(actionLoc)
	actions := NewInsMap[string, ActionsAction]()
	actions.Set("check", action)

	var got []Location
	trace := captureStdout(t, func() {
		for _, loc := range FindReferences(m, map[string]bool{"state": true}, actions).All() {
			got = append(got, loc)
		}
	})
	if want := []Location{formulaLoc, actionLoc}; !slices.Equal(got, want) {
		t.Errorf("locations = %v, want %v", got, want)
	}
	const wantTrace = "XTRACE: isolate.collectActionSymNames.add guard\n" +
		"XTRACE: isolate.collectActionSymNames.add state\n" +
		"XTRACE: isolate.FindReferences n_syms=1 n_refs=2\n"
	if xtracer.Enabled && trace != wantTrace {
		t.Errorf("trace:\n%s\nwant:\n%s", trace, wantTrace)
	}
}
