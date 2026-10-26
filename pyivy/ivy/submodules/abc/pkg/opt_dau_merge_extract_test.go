package abc

// Go-only structural tests for pure-subtree extraction during DSD merging.

import (
	"errors"
	"slices"
	"testing"
)

func checkDauMergeExtraction(t *testing.T, text string, shared int, wantReduced string, wantDefinitions []string) {
	t.Helper()
	formula := mustParseDauMerge(t, text)
	statuses, err := dauDSDMergeStatuses(formula, shared, nil)
	if err != nil {
		t.Fatal(err)
	}
	definitions := &dauDSDMergeDefinitions{shared: shared, maxVariables: 12, maxBytes: 2000}
	reduced, err := dauDSDMergeExtract(formula, statuses, formula.root, definitions, nil)
	if err != nil || reduced != wantReduced || !slices.Equal(definitions.values, wantDefinitions) {
		t.Fatal(text, reduced, definitions.values, err)
	}
}

func TestDauDSDMergeExtractStatusClasses(t *testing.T) {
	checkDauMergeExtraction(t, "(ab)", 2, "(ab)", nil)
	checkDauMergeExtraction(t, "(a(bc))", 1, "(ab)", []string{"(bc)"})
	checkDauMergeExtraction(t, "(a(bc)(de))", 1, "(ab)", []string{"((bc)(de))"})
	checkDauMergeExtraction(t, "[a![bc][de]]", 1, "[ab]", []string{"[![bc][de]]"})
	checkDauMergeExtraction(t, "<abc>", 1, "<abc>", []string{"b", "c"})
	checkDauMergeExtraction(t, "96{a(bc)d}", 1, "96{abc}", []string{"(bc)", "d"})
	checkDauMergeExtraction(t, "<abc>{d(ef)g}", 4, "<abc>{def}", []string{"(ef)", "g"})
}

func TestDauDSDMergeExtractLimitsAndCancellation(t *testing.T) {
	formula := mustParseDauMerge(t, "(a(bc))")
	statuses, err := dauDSDMergeStatuses(formula, 1, nil)
	if err != nil {
		t.Fatal(err)
	}
	definitions := &dauDSDMergeDefinitions{shared: 1, maxVariables: 1, maxBytes: 2000}
	if result, err := dauDSDMergeExtract(formula, statuses, formula.root, definitions, nil); result != "" || !errors.Is(err, ErrLimit) || definitions.values != nil {
		t.Fatal(result, definitions.values, err)
	}
	definitions = &dauDSDMergeDefinitions{shared: 1, maxVariables: 12, maxBytes: 3}
	if result, err := dauDSDMergeExtract(formula, statuses, formula.root, definitions, nil); result != "" || !errors.Is(err, ErrLimit) {
		t.Fatal(result, err)
	}
	definitions = &dauDSDMergeDefinitions{shared: 1, maxVariables: 12, maxBytes: 2000}
	if result, err := dauDSDMergeExtract(formula, statuses, formula.root, definitions, func() bool { return true }); result != "" || !errors.Is(err, ErrCancelled) || definitions.values != nil {
		t.Fatal(result, definitions.values, err)
	}
}
