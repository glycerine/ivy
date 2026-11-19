package goivy

import "testing"

// Python IvyDomainSetup (ivy_compiler.py) raises the definition checks as
// IvyError(ldf, "Variable {} occurs free on right-hand side of
// definition"): located at the labeled definition. goivy raised them at
// the compiled definition, which has no location (test/fundef1.ivy).
func TestAddDefinitionCheckErrorIsLocatedAtLdf(t *testing.T) {
	ts := isolateMkSort("t")
	fs, err := NewFunctionSort(ts, ts)
	if err != nil {
		t.Fatal(err)
	}
	x, _ := NewVariable("X", ts)
	y, _ := NewVariable("Y", ts)
	def := NewDefinition(MustApply(NewConst("f", fs), x), y)
	cfg := NewAstConfig()
	ldf := cfg.NewLabeledFormula(cfg.NewAtom("def1"), def)
	ldf.SetLineno(Location{Filename: "f.ivy", Line: 5})
	d := NewDomainSetup(newTestCompiler())
	err = d.AddDefinition(ldf, nil)
	want := "f.ivy: line 5: error: Variable Y:t occurs free on right-hand side of definition"
	if err == nil || err.Error() != want {
		t.Fatalf("got %v\nwant %s", err, want)
	}
}
