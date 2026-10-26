package abc

import (
	"encoding/json"
	"errors"
	"os"
	"slices"
	"testing"
)

func cnfFrameGraph(t testing.TB, pattern, regs int) *AigManager {
	t.Helper()
	g := NewAigManager()
	pi, _ := g.CreateCI()
	lo := make([]AigLit, regs)
	for i := range lo {
		lo[i], _ = g.CreateCI()
	}
	g.CreateCO(0)
	for i := range lo {
		var next AigLit
		switch pattern {
		case 0:
			next = pi
		case 1:
			next = pi.Not()
		case 2:
			next = lo[(i+1)%regs]
		case 3:
			next = lo[i].Not()
		case 4:
			next, _ = g.And(pi, lo[i])
		case 5:
			next, _ = g.Xor(pi, lo[i])
		}
		g.CreateCO(next)
	}
	if e := g.SetRegisterCount(regs); e != nil {
		t.Fatal(e)
	}
	return g
}
func cnfFrameNext(pattern, regs, state int, input bool) int {
	next := 0
	for i := 0; i < regs; i++ {
		old := state>>uint(i)&1 != 0
		var v bool
		switch pattern {
		case 0:
			v = input
		case 1:
			v = !input
		case 2:
			v = state>>uint((i+1)%regs)&1 != 0
		case 3:
			v = !old
		case 4:
			v = input && old
		case 5:
			v = input != old
		}
		if v {
			next |= 1 << uint(i)
		}
	}
	return next
}
func checkCNFFrameTrajectories(t *testing.T, pattern, regs int, initialize bool, f *AigCNFFrames) {
	t.Helper()
	solver := NewSatSolver()
	ok, e := f.LoadInto(solver)
	if e != nil || !ok {
		t.Fatal(ok, e)
	}
	solver2, ok, e := f.NewSolver2(SatSolver2Options{})
	if e != nil || !ok {
		t.Fatal(ok, e)
	}
	for sequence := 0; sequence < 1<<uint(f.frames); sequence++ {
		for initial := 0; initial < 1<<uint(regs); initial++ {
			state := initial
			var assumptions []SatLit
			for frame := 0; frame < f.frames; frame++ {
				inputs, e := f.InputLiterals(frame)
				if e != nil {
					t.Fatal(e)
				}
				outputs, e := f.OutputLiterals(frame)
				if e != nil {
					t.Fatal(e)
				}
				input := sequence>>uint(frame)&1 != 0
				for i, l := range inputs {
					want := input
					if i > 0 {
						want = state>>uint(i-1)&1 != 0
					}
					if !want {
						l = l.Not()
					}
					assumptions = append(assumptions, l)
				}
				next := cnfFrameNext(pattern, regs, state, input)
				for i, l := range outputs {
					want := true
					if i > 0 {
						want = next>>uint(i-1)&1 != 0
					}
					if !want {
						l = l.Not()
					}
					assumptions = append(assumptions, l)
				}
				state = next
			}
			answer, e := solver.Solve(assumptions, SatLimits{})
			answer2, e2 := solver2.Solve(assumptions, SatLimits{})
			want := SatSatisfiable
			if initialize && initial != 0 {
				want = SatUnsatisfiable
			}
			if e != nil || e2 != nil || answer.Status != want || answer2.Status != want {
				t.Fatal(pattern, regs, f.frames, initialize, sequence, initial, answer, answer2, e, e2)
			}
			if want == SatSatisfiable && (f.frames > 1 || initialize) {
				index := (f.frames-1)*(2+2*regs) + 1
				assumptions[index] ^= 1
				answer, e = solver.Solve(assumptions, SatLimits{})
				answer2, e2 = solver2.Solve(assumptions, SatLimits{})
				if e != nil || e2 != nil || answer.Status != SatUnsatisfiable || answer2.Status != SatUnsatisfiable {
					t.Fatal("disconnected state", pattern, regs, f.frames, answer, answer2, e, e2)
				}
			}
		}
	}
}
func TestCNFFramesNative(t *testing.T) {
	data, e := os.ReadFile("test_vectors/oracle/cnf_frames.json")
	if e != nil {
		t.Fatal(e)
	}
	var rows []struct {
		Pattern, Registers, Frames, Initialize, Variables int
		Clauses                                           [][]SatLit
	}
	if e = json.Unmarshal(data, &rows); e != nil {
		t.Fatal(e)
	}
	if len(rows) != 96 {
		t.Fatal(len(rows))
	}
	for _, row := range rows {
		g := cnfFrameGraph(t, row.Pattern, row.Registers)
		c, e := DeriveAigCNF4(g, row.Registers, AigCNFDeriveOptions{})
		if e != nil {
			t.Fatal(e)
		}
		before := c.Clauses()
		f, e := c.UnrollFrames(g, row.Frames, row.Initialize != 0, CNFFrameOptions{})
		if e != nil || f.variables != row.Variables || !equalSatClauses(f.clauses, row.Clauses) {
			t.Fatal(row.Pattern, row.Registers, row.Frames, row.Initialize, f, row, e)
		}
		if !equalSatClauses(before, c.clauses) {
			t.Fatal("source changed")
		}
		legacy, ok, e := f.NewSolver(0)
		if e != nil || !ok || legacy.VariableCount() != row.Variables {
			t.Fatal("legacy loader", ok, e)
		}
		solver2, ok, e := f.NewSolver2(SatSolver2Options{})
		if e != nil || !ok || solver2.VariableCount() != row.Variables || solver2.ClauseCount() != len(row.Clauses) {
			t.Fatal("solver2 loader", ok, e)
		}
		checkCNFFrameTrajectories(t, row.Pattern, row.Registers, row.Initialize != 0, f)
	}
}
func TestCNFFramesLimitsOwnershipAndPhases(t *testing.T) {
	g := cnfFrameGraph(t, 3, 2)
	c, e := DeriveAigCNF4(g, 2, AigCNFDeriveOptions{})
	if e != nil {
		t.Fatal(e)
	}
	for _, o := range []CNFFrameOptions{{MaxFrames: 1}, {MaxObjects: 1}, {MaxVariables: 1}, {MaxClauses: 1}, {MaxLiterals: 1}, {Stop: func() bool { return true }}} {
		r, e := c.UnrollFrames(g, 3, true, o)
		if r != nil || !(errors.Is(e, ErrLimit) || errors.Is(e, ErrCancelled)) {
			t.Fatal(r, e)
		}
	}
	frame, e := c.UnrollFrames(g, 2, true, CNFFrameOptions{})
	if e != nil {
		t.Fatal(e)
	}
	solver2, _ := NewSatSolver2(SatSolver2Options{})
	if ok, e := frame.LoadIntoSolver2(solver2, 7); e != nil || !ok || solver2.ClauseCount() != len(frame.clauses) {
		t.Fatal(ok, e, solver2.ClauseCount(), len(frame.clauses))
	}
	for handle := 1; handle <= solver2.ClauseCount(); handle++ {
		clause, e := solver2.Clause(handle)
		if e != nil || clause.ID != 6+handle {
			t.Fatal(handle, clause, e)
		}
	}
	beforeClauses, beforeVariables := solver2.ClauseCount(), solver2.VariableCount()
	bad := &AigCNFFrames{variables: 1, frames: 1, stride: 1, clauses: [][]SatLit{{}}}
	if ok, e := bad.LoadIntoSolver2(solver2, 100); e != nil || ok || solver2.ClauseCount() != beforeClauses || solver2.VariableCount() != beforeVariables {
		t.Fatal("atomic contradiction", ok, e, solver2.ClauseCount(), solver2.VariableCount())
	}
	if _, _, e := frame.NewSolver(1); !errors.Is(e, ErrLimit) {
		t.Fatal(e)
	}
	for _, include := range []bool{false, true} {
		phase, e := c.TransformPolarity(g, include, CNFTransformOptions{})
		if e != nil {
			t.Fatal(e)
		}
		f, e := phase.UnrollFrames(g, 3, true, CNFFrameOptions{Base: 7})
		if e != nil {
			t.Fatal(e)
		}
		checkCNFFrameTrajectories(t, 3, 2, true, f)
		lits, _ := f.InputLiterals(0)
		lits[0] ^= 1
		again, _ := f.InputLiterals(0)
		if slices.Equal(lits, again) {
			t.Fatal("alias")
		}
		if _, e = f.OutputLiterals(3); e == nil {
			t.Fatal("invalid frame")
		}
	}
	asserted, e := DeriveAigCNF4(g, 0, AigCNFDeriveOptions{})
	if e != nil {
		t.Fatal(e)
	}
	if r, e := asserted.UnrollFrames(g, 2, false, CNFFrameOptions{}); r != nil || e != ErrDimension {
		t.Fatal(r, e)
	}
}
func FuzzCNFFrames(f *testing.F) {
	f.Add(uint8(5), uint8(2), uint8(3), true, uint8(0))
	f.Fuzz(func(t *testing.T, pattern, registers, frames uint8, init bool, base uint8) {
		p, r, n := int(pattern%6), 1+int(registers%2), 1+int(frames%4)
		g := cnfFrameGraph(t, p, r)
		c, e := DeriveAigCNF4(g, r, AigCNFDeriveOptions{})
		if e != nil {
			t.Fatal(e)
		}
		out, e := c.UnrollFrames(g, n, init, CNFFrameOptions{Base: SatVar(base % 32)})
		if e != nil {
			t.Fatal(e)
		}
		checkCNFFrameTrajectories(t, p, r, init, out)
	})
}
