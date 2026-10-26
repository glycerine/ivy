package abc

import (
	"slices"
	"testing"
)

func TestPdrCoreReductionAliasesAndInitialization(t *testing.T) {
	for _, tc := range []struct {
		name     string
		next     []SatLit
		state    []PdrLit
		initial  []LatchInit
		conflict []SatLit
		want     []PdrLit
	}{
		{"aliases", []SatLit{2, 2, 3}, []PdrLit{0, 2, 5}, []LatchInit{0, 0, 0}, []SatLit{3}, []PdrLit{0}},
		{"complement", []SatLit{2, 3}, []PdrLit{1, 2}, []LatchInit{1, 0}, []SatLit{2}, []PdrLit{1}},
		{"restore zero", []SatLit{2, 4}, []PdrLit{1, 2}, []LatchInit{0, 0}, []SatLit{2}, []PdrLit{1, 2}},
		{"restore one", []SatLit{2, 4}, []PdrLit{1, 2}, []LatchInit{1, 1}, []SatLit{5}, []PdrLit{1, 2}},
		{"empty conflict", []SatLit{2, 4}, []PdrLit{1, 2}, []LatchInit{1, 2}, nil, []PdrLit{1}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := pdrEngine{next: tc.next, model: &AigerModel{initial: tc.initial}}
			cube, err := NewPdrCube(tc.state, nil)
			if err != nil {
				t.Fatal(err)
			}
			got, err := p.reduceClause(cube, tc.conflict)
			if err != nil || !slices.Equal(got.state, tc.want) {
				t.Fatal(got, err)
			}
			overlaps, err := got.IntersectsInitial(tc.initial)
			if err != nil || overlaps {
				t.Fatal("lost reset blocker", err)
			}
			if !cube.Contains(got) {
				t.Fatal("reduction added literals")
			}
		})
	}
	p := pdrEngine{next: []SatLit{2}, model: &AigerModel{initial: []LatchInit{0}}}
	cube, _ := NewPdrCube([]PdrLit{0}, nil)
	if _, err := p.reduceClause(cube, []SatLit{5}); err == nil {
		t.Fatal("unmapped core accepted")
	}
	p.options.SAT.Stop = func() bool { return true }
	if _, err := p.reduceClause(cube, []SatLit{3}); err != errPdrStopped {
		t.Fatal("core reduction ignored cancellation", err)
	}
	p.options.SAT.Stop = nil
	initial, _ := NewPdrCube([]PdrLit{1}, nil)
	if _, err := p.reduceClause(initial, nil); err == nil {
		t.Fatal("initial cube admitted for blocking")
	}
}
