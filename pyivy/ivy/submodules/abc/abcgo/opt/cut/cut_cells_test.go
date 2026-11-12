package cut

import (
	"strings"
	"testing"
)

func TestCellPrimitiveTruths(t *testing.T) {
	a, b, c := []uint32{0xaaaaaaaa}, []uint32{0xcccccccc}, []uint32{0xf0f0f0f0}
	expected := []uint32{0, 0xffffffff, 0xaaaaaaaa, 0x88888888, 0x77777777, 0x66666666, 0x80808080, 0x7f7f7f7f, 0xa8a8a8a8, 0x57575757, 0x96969696, 0x6a6a6a6a, 0x28282828, 0xd7d7d7d7, 0xe8e8e8e8, 0x81818181, 0x7e7e7e7e, 0xd8d8d8d8, 0x68686868, 0x97979797, 0xc2c2c2c2, 0x3d3d3d3d}
	for kind, want := range expected {
		out := []uint32{0}
		Cut_CellTruthElem(a, b, c, out, 3, int32(kind))
		if out[0] != want {
			t.Fatal(kind, out, want)
		}
	}
}
func TestCellCrossBarUsesOwnPhase(t *testing.T) {
	for phase := byte(0); phase < 3; phase++ {
		c := &Cut_Cell_t{NVars: 2, CrossBar0: 0, CrossBar1: 1, CrossBarPhase: phase, CanonPhase: 0}
		c.Truth[0] = 0xaaaaaaaa
		Cut_CellCrossBar(c)
		want := uint32(0xaaaaaaaa)
		if phase == 1 {
			want = ^want
		}
		if c.Truth[0] != want {
			t.Fatal(phase, c.Truth[0], want)
		}
	}
}
func TestCellLibraryPrecomputeLoadDumpLookup(t *testing.T) {
	p := PrecomputeCellLibrary([]int32{6, 10}, []int32{2, 3}, true)
	defer Cut_CManStop(p)
	if p.Total < 16 || p.Good == 0 || len(p.table) == 0 {
		t.Fatal(p.Total, p.Good)
	}
	for _, cell := range p.table {
		cell.Used = 1
		if Cut_CellTruthLookup(cell.Truth[:], cell.NVars) == 0 {
			t.Fatalf("stored cell not found after lookup normalization: n%d truth%x", cell.NVars, cell.Truth)
		}
	}
	var out strings.Builder
	if err := DumpCellLibrary(p, &out); err != nil {
		t.Fatal(err)
	}
	reloaded, err := LoadCellLibrary(strings.NewReader(out.String()))
	if err != nil {
		t.Fatal(err)
	}
	if len(reloaded.table) != len(p.table) {
		t.Fatal("load/dump changed library", len(reloaded.table), len(p.table))
	}
	for _, cell := range p.table {
		if Cut_CellTruthLookup(cell.Truth[:], cell.NVars) == 0 {
			t.Fatal("roundtrip lookup")
		}
	}
	Cut_CManStop(reloaded)
	if Cut_CellIsRunning() != 0 {
		t.Fatal("stop left global dangling")
	}
}
func TestCellLoaderFinalLineAndErrors(t *testing.T) {
	p, err := LoadCellLibrary(strings.NewReader("80808080"))
	if err != nil {
		t.Fatal(err)
	}
	if p.Total != 1 || p.Good != 1 {
		t.Fatal(p.Total, p.Good)
	}
	Cut_CManStop(p)
	if _, err := LoadCellLibrary(strings.NewReader("not hex\n")); err == nil {
		t.Fatal("invalid hex accepted")
	}
}
