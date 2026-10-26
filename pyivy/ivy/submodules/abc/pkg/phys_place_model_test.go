package abc

// Tests for the Go port of src/phys/place placement model, density, I/O, and pads.

import (
	"bytes"
	"errors"
	"math"
	"reflect"
	"testing"
)

func placementTestDB(t testing.TB) *Placement {
	t.Helper()
	p, err := NewPlacement(PlacementLimits{})
	if err != nil {
		t.Fatal(err)
	}
	typeCore := PlacementCellType{Label: "core", Width: 2, Height: 3}
	for _, cell := range []PlacementCell{
		{ID: 0, Label: "a", Type: typeCore, X: 1, Y: 4},
		{ID: 2, Label: "b", Type: typeCore, X: 6, Y: -1},
		{ID: 5, Label: "c", Type: typeCore, X: 3, Y: 2},
	} {
		if err = p.AddCell(cell); err != nil {
			t.Fatal(err)
		}
	}
	if err = p.AddNet(PlacementNet{ID: 3, Terms: []int{0, 2, 5}, Weight: 1}); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestPlacementGeometryOwnershipAndLifecycle(t *testing.T) {
	p := placementTestDB(t)
	if got, want := p.CellIDs(), []int{0, 2, 5}; !reflect.DeepEqual(got, want) {
		t.Fatalf("cell IDs %v want %v", got, want)
	}
	if got, want := p.NetIDs(), []int{3}; !reflect.DeepEqual(got, want) {
		t.Fatalf("net IDs %v want %v", got, want)
	}
	area, err := p.CellArea(2)
	if err != nil || area != 6 {
		t.Fatalf("area %v %v", area, err)
	}
	box, err := p.NetBBox(3)
	if err != nil || box != (PlacementRect{X: 1, Y: -1, Width: 5, Height: 5}) {
		t.Fatalf("box %+v %v", box, err)
	}
	wire, err := p.TotalWirelength()
	if err != nil || wire != 10 {
		t.Fatalf("wire %v %v", wire, err)
	}
	if ids, err := p.SortedCellIDs(PlacementCellByY); err != nil || !reflect.DeepEqual(ids, []int{2, 5, 0}) {
		t.Fatalf("cell sort %v %v", ids, err)
	}
	if ids, err := p.SortedNetIDs(PlacementNetByRight); err != nil || !reflect.DeepEqual(ids, []int{3}) {
		t.Fatalf("net sort %v %v", ids, err)
	}

	owned, _ := p.Net(3)
	owned.Terms[0] = 5
	again, _ := p.Net(3)
	if again.Terms[0] != 0 {
		t.Fatal("Net exposed internal term storage")
	}
	clone, err := p.Clone()
	if err != nil {
		t.Fatal(err)
	}
	if err = clone.DeleteNet(3); err != nil {
		t.Fatal(err)
	}
	if len(p.NetIDs()) != 1 || len(clone.NetIDs()) != 0 {
		t.Fatal("clone aliases source")
	}
	if err = p.DeleteCell(2); !errors.Is(err, ErrPlacementReferenced) {
		t.Fatalf("delete referenced: %v", err)
	}
	if err = p.DeleteNet(3); err != nil {
		t.Fatal(err)
	}
	if err = p.DeleteCell(2); err != nil {
		t.Fatal(err)
	}
	if err = p.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestPlacementCheckedMutation(t *testing.T) {
	p, _ := NewPlacement(PlacementLimits{MaxCellID: 3, MaxNetID: 2, MaxTerms: 2})
	valid := PlacementCell{ID: 1, Type: PlacementCellType{Width: 1, Height: 1}}
	if err := p.AddCell(valid); err != nil {
		t.Fatal(err)
	}
	if err := p.AddCell(valid); !errors.Is(err, ErrPlacementDuplicate) {
		t.Fatalf("duplicate: %v", err)
	}
	invalid := valid
	invalid.ID = 4
	if err := p.AddCell(invalid); !errors.Is(err, ErrLimit) {
		t.Fatalf("cell limit: %v", err)
	}
	if err := p.AddNet(PlacementNet{ID: 0, Terms: []int{1, 2}}); !errors.Is(err, ErrPlacementMissing) {
		t.Fatalf("missing term: %v", err)
	}
	if len(p.NetIDs()) != 0 {
		t.Fatal("failed AddNet was not atomic")
	}
	if err := p.AddNet(PlacementNet{ID: 3, Terms: []int{1}}); !errors.Is(err, ErrLimit) {
		t.Fatalf("net limit: %v", err)
	}
	cancelled, _ := NewPlacement(PlacementLimits{Stop: func() bool { return true }})
	if err := cancelled.AddCell(valid); !errors.Is(err, ErrCancelled) || len(cancelled.CellIDs()) != 0 {
		t.Fatalf("cancelled add: %v", err)
	}

	empty, _ := NewPlacement(PlacementLimits{})
	if err := empty.AddCell(valid); err != nil {
		t.Fatal(err)
	}
	if err := empty.AddNet(PlacementNet{ID: 0}); err != nil {
		t.Fatal(err)
	}
	if _, err := empty.NetBBox(0); !errors.Is(err, ErrDimension) {
		t.Fatalf("empty bbox: %v", err)
	}
}

func TestPlacementBookshelfRoundTrip(t *testing.T) {
	p := placementTestDB(t)
	pad := PlacementCell{ID: 7, Label: "pad", Type: PlacementCellType{Label: "pad", Width: 4, Height: 1, Pad: true}, Fixed: true, X: 12.5, Y: -2.25}
	if err := p.AddCell(pad); err != nil {
		t.Fatal(err)
	}
	files, err := p.Bookshelf()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(files.Nodes, []byte("NumNodes : 4")) || !bytes.Contains(files.Nodes, []byte("NumTerminals : 1")) || !bytes.Contains(files.Placement, []byte("\\FIXED")) {
		t.Fatalf("unexpected output\n%s\n%s", files.Nodes, files.Placement)
	}
	if !bytes.Contains(files.Nodes, []byte("CELL0 2.000000 3.000000 \n")) ||
		!bytes.Contains(files.Nodes, []byte("CELL7 4.000000 1.000000  terminal\n")) ||
		!bytes.Contains(files.Placement, []byte("CELL0 1.000000 4.000000 : N \n")) {
		t.Fatalf("source spacing was not preserved\n%s\n%s", files.Nodes, files.Placement)
	}
	parsed, err := ParseBookshelf(files.Nodes, files.Placement, PlacementLimits{})
	if err != nil {
		t.Fatal(err)
	}
	for _, id := range p.CellIDs() {
		before, _ := p.Cell(id)
		after, ok := parsed.Cell(id)
		if !ok || before.Type.Width != after.Type.Width || before.Type.Height != after.Type.Height || before.Type.Pad != after.Type.Pad || before.X != after.X || before.Y != after.Y || before.Fixed != after.Fixed {
			t.Fatalf("cell %d before %+v after %+v", id, before, after)
		}
	}
	bad := append([]byte(nil), files.Nodes...)
	bad = bytes.Replace(bad, []byte("NumNodes : 4"), []byte("NumNodes : 5"), 1)
	if _, err = ParseBookshelf(bad, files.Placement, PlacementLimits{}); !errors.Is(err, ErrPlacementFormat) {
		t.Fatalf("bad count: %v", err)
	}
	if _, err = ParseBookshelf(files.Nodes, append(files.Placement, []byte("CELL0 0 0 : N\n")...), PlacementLimits{}); !errors.Is(err, ErrPlacementFormat) {
		t.Fatalf("duplicate placement: %v", err)
	}
	if _, err = ParseBookshelf(files.Nodes, files.Placement, PlacementLimits{MaxBytes: 8}); !errors.Is(err, ErrLimit) {
		t.Fatalf("byte limit: %v", err)
	}
}

func TestPlacementPreplace(t *testing.T) {
	p, _ := NewPlacement(PlacementLimits{})
	core := PlacementCellType{Label: "core", Width: 2, Height: 2}
	pad := PlacementCellType{Label: "pad", Width: 1, Height: 1, Pad: true}
	for id := 0; id < 4; id++ {
		if err := p.AddCell(PlacementCell{ID: id, Type: core, X: float64(id)}); err != nil {
			t.Fatal(err)
		}
	}
	for id := 4; id < 12; id++ {
		if err := p.AddCell(PlacementCell{ID: id, Type: pad}); err != nil {
			t.Fatal(err)
		}
	}
	if err := p.Preplace(0.5); err != nil {
		t.Fatal(err)
	}
	if p.RowHeight != 2 || p.CoreBounds.Width*p.CoreBounds.Height+1e-9 < 32 {
		t.Fatalf("core %+v row %v", p.CoreBounds, p.RowHeight)
	}
	if p.PadBounds.Width != p.CoreBounds.Width+2 || p.PadBounds.Height != p.CoreBounds.Height+2 {
		t.Fatalf("pad bounds %+v", p.PadBounds)
	}
	for id := 4; id < 12; id++ {
		cell, _ := p.Cell(id)
		onSide := false
		switch {
		case id < 6:
			onSide = cell.Y == p.PadBounds.Y+cell.Type.Height/2
		case id < 8:
			onSide = cell.Y == p.PadBounds.Y+p.PadBounds.Height-cell.Type.Height/2
		case id < 10:
			onSide = cell.X == p.PadBounds.X+p.PadBounds.Width-cell.Type.Width/2
		default:
			onSide = cell.X == p.PadBounds.X+cell.Type.Width/2
		}
		if !onSide {
			t.Fatalf("pad %d is not on ring: %+v", id, cell)
		}
	}
}

func TestPlacementDensityBoundsMovementAndAtomicity(t *testing.T) {
	p, _ := NewPlacement(PlacementLimits{})
	p.CoreBounds = PlacementRect{Width: 20, Height: 10}
	typeCore := PlacementCellType{Width: 2, Height: 2}
	for id, point := range [][2]float64{{1, 1}, {2, 2}, {3, 7}, {19, 8}, {18, 3}, {17, 6}} {
		if err := p.AddCell(PlacementCell{ID: id, Type: typeCore, X: point[0], Y: point[1]}); err != nil {
			t.Fatal(err)
		}
	}
	before := make([]PlacementCell, 6)
	areaBefore := 0.0
	for id := range before {
		before[id], _ = p.Cell(id)
		areaBefore += before[id].Type.Width * before[id].Type.Height
	}
	if err := p.SpreadDensityX(3, 2.5); err != nil {
		t.Fatal(err)
	}
	if err := p.SpreadDensityY(3, 1.5); err != nil {
		t.Fatal(err)
	}
	for id, old := range before {
		cell, _ := p.Cell(id)
		if math.Abs(cell.X-old.X) > 2.5+1e-9 || math.Abs(cell.Y-old.Y) > 1.5+1e-9 {
			t.Fatalf("movement %d old %+v new %+v", id, old, cell)
		}
		if cell.X-cell.Type.Width/2 < 0 || cell.X+cell.Type.Width/2 > 20 || cell.Y-cell.Type.Height/2 < 0 || cell.Y+cell.Type.Height/2 > 10 {
			t.Fatalf("cell outside core: %+v", cell)
		}
	}
	areaAfter := 0.0
	for id := range before {
		cell, _ := p.Cell(id)
		areaAfter += cell.Type.Width * cell.Type.Height
	}
	if areaAfter != areaBefore {
		t.Fatalf("density changed area: %v -> %v", areaBefore, areaAfter)
	}

	tooLarge, _ := NewPlacement(PlacementLimits{})
	tooLarge.CoreBounds = PlacementRect{Width: 1, Height: 1}
	_ = tooLarge.AddCell(PlacementCell{ID: 0, Type: PlacementCellType{Width: 2, Height: 1}, X: 0.5, Y: 0.5})
	beforeLarge, _ := tooLarge.Cell(0)
	if err := tooLarge.SpreadDensityX(1, 1); !errors.Is(err, ErrDimension) {
		t.Fatalf("large cell: %v", err)
	}
	afterLarge, _ := tooLarge.Cell(0)
	if beforeLarge != afterLarge {
		t.Fatal("failed density mutation was not atomic")
	}

	cancelled, _ := p.Clone()
	cancelled.Limits.Stop = func() bool { return true }
	beforeCancel, _ := cancelled.Cell(0)
	if err := cancelled.SpreadDensityX(2, 2); !errors.Is(err, ErrCancelled) {
		t.Fatalf("cancel: %v", err)
	}
	afterCancel, _ := cancelled.Cell(0)
	if beforeCancel != afterCancel {
		t.Fatal("cancelled density changed cells")
	}
}

func TestPlacementIndependentGeometryModel5000(t *testing.T) {
	state := uint64(0x9e3779b97f4a7c15)
	next := func() uint64 {
		state = state*6364136223846793005 + 1442695040888963407
		return state
	}
	for run := 0; run < 5000; run++ {
		p, _ := NewPlacement(PlacementLimits{})
		count := int(next()%7) + 1
		minX, minY := math.Inf(1), math.Inf(1)
		maxX, maxY := math.Inf(-1), math.Inf(-1)
		terms := make([]int, count)
		for id := 0; id < count; id++ {
			x := float64(int(next()%401)-200) / 8
			y := float64(int(next()%401)-200) / 8
			if err := p.AddCell(PlacementCell{ID: id, Type: PlacementCellType{Width: 1, Height: 1}, X: x, Y: y}); err != nil {
				t.Fatal(err)
			}
			terms[id] = id
			minX, minY = math.Min(minX, x), math.Min(minY, y)
			maxX, maxY = math.Max(maxX, x), math.Max(maxY, y)
		}
		if err := p.AddNet(PlacementNet{ID: 0, Terms: terms, Weight: 1}); err != nil {
			t.Fatal(err)
		}
		box, err := p.NetBBox(0)
		if err != nil || box != (PlacementRect{X: minX, Y: minY, Width: maxX - minX, Height: maxY - minY}) {
			t.Fatalf("run %d box %+v want [%v %v %v %v] err %v", run, box, minX, minY, maxX-minX, maxY-minY, err)
		}
		wire, _ := p.NetWirelength(0)
		if wire != maxX-minX+maxY-minY {
			t.Fatalf("run %d wire %v", run, wire)
		}
	}
}

func FuzzPlacementModel(f *testing.F) {
	f.Add([]byte{3, 1, 2, 9, 5, 4, 7, 8})
	f.Add([]byte{1, 0, 0})
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) == 0 {
			return
		}
		count := int(data[0]%8) + 1
		p, _ := NewPlacement(PlacementLimits{MaxCellID: 32, MaxNetID: 8, MaxTerms: 64})
		terms := make([]int, count)
		minX, minY := math.Inf(1), math.Inf(1)
		maxX, maxY := math.Inf(-1), math.Inf(-1)
		for id := 0; id < count; id++ {
			at := 1 + 2*id
			var xb, yb byte
			if at < len(data) {
				xb = data[at]
			}
			if at+1 < len(data) {
				yb = data[at+1]
			}
			x, y := float64(int(xb)-128)/4, float64(int(yb)-128)/4
			cell := PlacementCell{ID: id * 2, Type: PlacementCellType{Width: float64(xb%4) + 1, Height: float64(yb%4) + 1, Pad: id == count-1 && xb&1 != 0}, Fixed: yb&1 != 0, X: x, Y: y}
			if err := p.AddCell(cell); err != nil {
				t.Fatal(err)
			}
			terms[id] = cell.ID
			minX, minY = math.Min(minX, x), math.Min(minY, y)
			maxX, maxY = math.Max(maxX, x), math.Max(maxY, y)
		}
		if err := p.AddNet(PlacementNet{ID: 1, Terms: terms, Weight: 1}); err != nil {
			t.Fatal(err)
		}
		box, err := p.NetBBox(1)
		if err != nil || box != (PlacementRect{X: minX, Y: minY, Width: maxX - minX, Height: maxY - minY}) {
			t.Fatalf("bbox %+v %v", box, err)
		}
		files, err := p.Bookshelf()
		if err != nil {
			t.Fatal(err)
		}
		parsed, err := ParseBookshelf(files.Nodes, files.Placement, PlacementLimits{MaxCellID: 32})
		if err != nil || !reflect.DeepEqual(parsed.CellIDs(), p.CellIDs()) {
			t.Fatalf("roundtrip %v", err)
		}
		if err := parsed.Validate(); err != nil {
			t.Fatal(err)
		}
	})
}
