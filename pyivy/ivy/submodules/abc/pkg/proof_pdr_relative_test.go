package abc

import (
	"errors"
	"testing"
)

func TestPdrRelativeQueryActivationLifetime(t *testing.T) {
	m := bmcTruthModel(t, bmcTruthCase{next0: 204, next1: 240, constraint: 255})
	options, err := (PdrOptions{}).normalized()
	if err != nil {
		t.Fatal(err)
	}
	p, err := newPdrEngine(m, options)
	if err != nil {
		t.Fatal(err)
	}
	if err := p.newFrame(false); err != nil {
		t.Fatal(err)
	}
	cube, err := NewPdrCube([]PdrLit{0}, nil)
	if err != nil {
		t.Fatal(err)
	}
	plain, err := p.queryCube(0, cube, true)
	if err != nil || plain.Status != SatSatisfiable {
		t.Fatal(plain, err)
	}
	relative, err := p.queryRelativeCube(0, cube)
	if err != nil || relative.Status != SatUnsatisfiable {
		t.Fatal("self loop is relatively inductive", relative, err)
	}
	plain, err = p.queryCube(0, cube, true)
	if err != nil || plain.Status != SatSatisfiable {
		t.Fatal("activation clause leaked", plain, err)
	}
	calls := 0
	p.options.SAT.Stop = func() bool { calls++; return calls >= 2 }
	if _, err := p.queryRelativeCube(0, cube); err != errPdrStopped {
		t.Fatal("expected query cancellation", err)
	}
	p.options.SAT.Stop = nil
	plain, err = p.queryCube(0, cube, true)
	if err != nil || plain.Status != SatSatisfiable {
		t.Fatal("cancelled activation leaked", plain, err)
	}
	if err := p.frames[0].solver.Check(); err != nil {
		t.Fatal(err)
	}
}
func TestPdrRelativeVariablesCountAgainstFrameBudget(t *testing.T) {
	m := bmcTruthModel(t, bmcTruthCase{next0: 204, constraint: 255})
	options, _ := (PdrOptions{}).normalized()
	p, err := newPdrEngine(m, options)
	if err != nil {
		t.Fatal(err)
	}
	if err := p.newFrame(false); err != nil {
		t.Fatal(err)
	}
	p.options.MaxFrameVariables = p.frames[0].solver.VariableCount()
	cube, _ := NewPdrCube([]PdrLit{0}, nil)
	if _, err := p.queryRelativeCube(0, cube); err != ErrLimit {
		t.Fatal("activation bypassed global budget", err)
	}
	p.options.MaxFrameVariables = 2 * p.relation.variables
	if _, err := p.queryRelativeCube(0, cube); err != nil {
		t.Fatal(err)
	}
	if err := p.newFrame(false); !errors.Is(err, ErrLimit) {
		t.Fatal("new frame ignored activation variables", err)
	}
}
