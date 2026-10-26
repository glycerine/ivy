package abc

import (
	"slices"
	"testing"
)

func TestPdrPriorityOrderingAndUpdates(t *testing.T) {
	cube, _ := NewPdrCube([]PdrLit{0, 3, 4, 7}, nil)
	p := pdrEngine{flopPriority: []uint64{2, 2, 1, 0}}
	order, err := p.orderCube(cube)
	if err != nil || !slices.Equal(order, []PdrLit{7, 4, 3, 0}) {
		t.Fatal("C selection tie behavior", order, err)
	}
	if !slices.Equal(cube.state, []PdrLit{0, 3, 4, 7}) {
		t.Fatal("cube mutated")
	}
	if err := p.bumpCubePriority(cube); err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(p.flopPriority, []uint64{3, 3, 2, 1}) {
		t.Fatal(p.flopPriority)
	}
	p.flopPriority = []uint64{^uint64(0), 8, 6, 2}
	if err := p.bumpCubePriority(cube); err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(p.flopPriority, []uint64{1 << 63, 5, 4, 2}) {
		t.Fatal("priority overflow", p.flopPriority)
	}
	invalid, _ := NewPdrCube([]PdrLit{8}, nil)
	if _, err := p.orderCube(invalid); err == nil {
		t.Fatal("invalid register")
	}
	if err := p.bumpCubePriority(invalid); err == nil {
		t.Fatal("invalid update")
	}
	p.options.SAT.Stop = func() bool { return true }
	if _, err := p.orderCube(cube); err != errPdrStopped {
		t.Fatal("sort cancellation", err)
	}
}

func TestPdrPriorityAssumptionOrder(t *testing.T) {
	cube, _ := NewPdrCube([]PdrLit{0, 3, 4, 7}, nil)
	p := pdrEngine{flopPriority: []uint64{0, 0, 1, 2}, current: []SatLit{2, 4, 6, 8}, next: []SatLit{10, 12, 14, 16}, options: PdrOptions{OrderAssumptions: true}}
	got, err := p.queryCubeLiterals(cube, true)
	if err != nil || !slices.Equal(got, []SatLit{17, 14, 13, 10}) {
		t.Fatal(got, err)
	}
	current, err := p.queryCubeLiterals(cube, false)
	if err != nil || !slices.Equal(current, []SatLit{2, 5, 6, 9}) {
		t.Fatal(current, err)
	}
	if !slices.Equal(cube.state, []PdrLit{0, 3, 4, 7}) {
		t.Fatal("assumption ordering mutated cube")
	}
	p.options.OrderAssumptions = false
	got, err = p.queryCubeLiterals(cube, true)
	if err != nil || !slices.Equal(got, []SatLit{10, 13, 14, 17}) {
		t.Fatal(got, err)
	}
}
func TestPdrOrderedAssumptionsReachability(t *testing.T) {
	for i := 0; i < 256; i++ {
		pdrVerifyTruthOptions(t, bmcTruthCase{bad: uint8(i * 53), next0: uint8(i * 73), next1: uint8(i * 131), constraint: uint8(i * 17), initial: [2]LatchInit{LatchInit(i % 3), LatchInit((i / 3) % 3)}}, PdrOptions{OrderAssumptions: true, RecycleVariables: 1, MaxFrames: 16, MaxQueries: 10000})
	}
}
