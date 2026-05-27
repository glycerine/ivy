package goivy

import "testing"

func TestParseV16TopLevelInit(t *testing.T) {
	result, err := Parse("type t\nindividual x:t\ninit x = x", Version{1, 6}, WithFilename("init16.ivy"))
	if err != nil {
		t.Fatalf("Parse v1.6 init: %v", err)
	}
	if got := countDeclsOf[*InitDecl](result.Decls); got != 1 {
		t.Fatalf("InitDecl count = %d, want 1", got)
	}
}

func TestParseV16UnlabeledAxiomDoesNotSynthesizeLabel(t *testing.T) {
	result, err := Parse("type t\nrelation r(X:t)\naxiom forall X:t . r(X)", Version{1, 6}, WithFilename("axiom16.ivy"))
	if err != nil {
		t.Fatalf("Parse v1.6 axiom: %v", err)
	}
	lf := firstLabeledFormulaInDecl[*AxiomDecl](t, result.Decls)
	if lf.LabelName() != "" {
		t.Fatalf("v1.6 unlabeled axiom synthesized label %q, want empty", lf.LabelName())
	}
}

func TestParseV16UnlabeledPropertyDoesNotSynthesizeLabel(t *testing.T) {
	result, err := Parse("type t\nrelation r(X:t)\nproperty forall X:t . r(X)", Version{1, 6}, WithFilename("prop16.ivy"))
	if err != nil {
		t.Fatalf("Parse v1.6 property: %v", err)
	}
	lf := firstLabeledFormulaInDecl[*PropertyDecl](t, result.Decls)
	if lf.LabelName() != "" {
		t.Fatalf("v1.6 unlabeled property synthesized label %q, want empty", lf.LabelName())
	}
}

func TestParseV17UnlabeledAxiomStillSynthesizesLabel(t *testing.T) {
	result, err := Parse("type t\nrelation r(X:t)\naxiom forall X:t . r(X)", Version{1, 7}, WithFilename("axiom17.ivy"))
	if err != nil {
		t.Fatalf("Parse v1.7 axiom: %v", err)
	}
	lf := firstLabeledFormulaInDecl[*AxiomDecl](t, result.Decls)
	if lf.LabelName() == "" {
		t.Fatal("v1.7 unlabeled axiom should still synthesize a label")
	}
}

func firstLabeledFormulaInDecl[T Node](t *testing.T, decls []Node) *LabeledFormula {
	t.Helper()
	for _, decl := range decls {
		if typed, ok := decl.(T); ok {
			args := typed.Args()
			if len(args) == 0 {
				t.Fatalf("%T has no args", typed)
			}
			lf, ok := args[0].(*LabeledFormula)
			if !ok {
				t.Fatalf("%T arg[0] = %T, want *LabeledFormula", typed, args[0])
			}
			return lf
		}
	}
	t.Fatalf("no declaration of requested type found")
	return nil
}
