package goivy

import "testing"

func TestNormalProgramCanonEmptyExact(t *testing.T) {
	np := &NormalProgram{}

	got := string(np.Canon())
	want := "(normalProgram bindings:[] init:nil invars:[] asms:[] calls:[] postconds:(hash))"
	if got != want {
		t.Fatalf("NormalProgram.Canon() mismatch\n got: %s\nwant: %s", got, want)
	}
}

func TestNormalProgramCanonPopulatedExact(t *testing.T) {
	binding := &ActionTermBinding{
		Name: "act",
		Action: &ActionTerm{
			Stmt: NewSequence(),
		},
	}
	np := &NormalProgram{
		Bindings: []*ActionTermBinding{binding},
		Init:     NewSequence(),
		Invars: []*LabeledFormula{
			{Formula: True, ID: 7},
		},
		Asms: []*LabeledFormula{
			{Formula: True, ID: 8},
		},
		Calls: []string{"act", `call"q`},
		Postconds: map[string][]*LabeledFormula{
			"a": nil,
			"z": {{Formula: True, ID: 9}},
		},
	}

	got := string(np.Canon())
	lf7 := "(labeledFormula label:nil formula:(And terms:[]) id:7 temporal:nil explicit:false isDefinition:false assumed:false unprovable:false)"
	lf8 := "(labeledFormula label:nil formula:(And terms:[]) id:8 temporal:nil explicit:false isDefinition:false assumed:false unprovable:false)"
	lf9 := "(labeledFormula label:nil formula:(And terms:[]) id:9 temporal:nil explicit:false isDefinition:false assumed:false unprovable:false)"
	want := "(normalProgram bindings:[(actionTermBinding name:\"act\" action:(actionTerm inputs:[] outputs:[] labels:[] stmt:(sequence stmts:[])))] init:(sequence stmts:[]) invars:[" +
		lf7 + "] asms:[" + lf8 + "] calls:[\"act\" \"call\\\"q\"] postconds:(hash \"a\":[] \"z\":[" + lf9 + "]))"
	if got != want {
		t.Fatalf("NormalProgram.Canon() mismatch\n got: %s\nwant: %s", got, want)
	}
}

func TestTemporalModelsCanonWithNormalProgramExact(t *testing.T) {
	tm := &TemporalModels{
		Model: &NormalProgram{},
		Fmla:  True,
	}

	got := string(tm.Canon())
	want := "(temporalModels model:(normalProgram bindings:[] init:nil invars:[] asms:[] calls:[] postconds:(hash)) fmla:(And terms:[]))"
	if got != want {
		t.Fatalf("TemporalModels.Canon() mismatch\n got: %s\nwant: %s", got, want)
	}
}
