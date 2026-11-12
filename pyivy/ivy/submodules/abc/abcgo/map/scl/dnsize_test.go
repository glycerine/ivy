package scl

import (
	"io"
	"testing"
)

type downsizeNetwork struct{ mapping NetworkMapping }

func (n *downsizeNetwork) SclName() string     { return "downsize" }
func (n *downsizeNetwork) SclObjectIDs() []int { return []int{0, 1, 2, 3} }
func (n *downsizeNetwork) SclInputs() []int    { return []int{0} }
func (n *downsizeNetwork) SclOutputs() []int   { return []int{3} }
func (n *downsizeNetwork) SclFanins(id int) []int {
	if id == 0 {
		return nil
	}
	return []int{id - 1}
}
func (n *downsizeNetwork) SclFanouts(id int) []int {
	if id == 3 {
		return nil
	}
	return []int{id + 1}
}
func (n *downsizeNetwork) SclIsNode(id int) bool       { return id == 1 || id == 2 }
func (n *downsizeNetwork) SclIsBarrier(id int) bool    { return false }
func (n *downsizeNetwork) SclMapping() *NetworkMapping { return &n.mapping }
func downsizeFixture(t *testing.T) (*SC_Lib, *downsizeNetwork) {
	t.Helper()
	lib := &SC_Lib{Name: "downsize", UnitTime: 12, UnitCapMultiplier: 1, UnitCapExponent: 15}
	for i := 0; i < 2; i++ {
		name := "small"
		if i == 1 {
			name = "large"
		}
		cell := &SC_Cell{Name: name, ID: i, Area: float32(i + 1), Inputs: 1, Outputs: 1}
		input := &SC_Pin{Name: "A", Direction: ScDirInput, RiseCap: float32(i + 1), FallCap: float32(i + 1)}
		out := &SC_Pin{Name: "Y", Direction: ScDirOutput, Function: "A", Truth: []uint64{0xaaaaaaaaaaaaaaaa}}
		timing := &SC_Timing{RelatedPin: "A", Sense: ScTSPos}
		for _, s := range []*SC_Surface{&timing.CellRise, &timing.CellFall, &timing.RiseTrans, &timing.FallTrans} {
			s.Index0 = []float32{0}
			s.Index1 = []float32{0}
			s.Data = [][]float32{{1}}
		}
		out.Timings = []*SC_Timings{{Name: "A", Timings: []*SC_Timing{timing}}}
		cell.Pins = []*SC_Pin{input, out}
		lib.Cells = append(lib.Cells, cell)
	}
	initializeIntegerTables(lib)
	Abc_SclHashCells(lib)
	Abc_SclLinkCells(lib)
	mioLib, e := Abc_SclDeriveGenlibSimple(lib)
	if e != nil {
		t.Fatal(e)
	}
	return lib, &downsizeNetwork{mapping: NetworkMapping{Lib: lib, Mio: mioLib, GateIDs: []int32{-1, 1, 1, -1}}}
}
func TestDownsizeAreaAndDelay(t *testing.T) {
	lib, net := downsizeFixture(t)
	p, e := Abc_SclManStart(lib, net, 0, 1, 0, 0)
	if e != nil {
		t.Fatal(e)
	}
	defer Abc_SclManFree(p)
	p.Output = io.Discard
	initial := p.MaxDelay
	if e = DownsizeManager(p, &SC_SizePars{Iterations: 2, Notches: 4, DelayGap: 1, UseDepartures: true}); e != nil {
		t.Fatal(e)
	}
	if net.mapping.GateIDs[1] != 0 || net.mapping.GateIDs[2] != 0 || p.SumArea != 2 || p.MaxDelay > initial {
		t.Fatal(net.mapping.GateIDs, p.SumArea, p.MaxDelay, initial)
	}
}
func TestDownsizeWindowAndOverlap(t *testing.T) {
	_, net := downsizeFixture(t)
	nodes, evals, e := Abc_SclFindWindow(net, 1)
	if e != nil {
		t.Fatal(e)
	}
	if len(nodes) != 3 || len(evals) != 1 || evals[0] != 2 {
		t.Fatal(nodes, evals)
	}
	if !Abc_SclCheckOverlap(map[int]bool{1: true}, nodes) || Abc_SclCheckOverlap(map[int]bool{3: true}, nodes) {
		t.Fatal("overlap")
	}
	if _, _, e = Abc_SclFindWindow(net, 0); e == nil {
		t.Fatal("accepted input pivot")
	}
}

func TestDownsizeRejectsNoBoundaryAndRestoresTrial(t *testing.T) {
	lib, net := downsizeFixture(t)
	p, e := Abc_SclManStart(lib, net, 0, 1, 0, 0)
	if e != nil {
		t.Fatal(e)
	}
	defer Abc_SclManFree(p)
	p.Output = io.Discard
	nodes, evals, e := Abc_SclFindWindow(net, 1)
	if e != nil {
		t.Fatal(e)
	}
	before := p.Times[2]
	changed, e := Abc_SclCheckImprovement(p, 1, nodes, evals, 0, 1)
	if e != nil {
		t.Fatal(e)
	}
	if changed || net.mapping.GateIDs[1] != 1 || p.Times[2] != before {
		t.Fatal("rejected trial changed state", changed, net.mapping.GateIDs, p.Times[2], before)
	}
}
