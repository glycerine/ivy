package abc

// Go-only model, property, and fuzz tests for the Gordian placement flow.

import (
	"errors"
	"math"
	"slices"
	"testing"
)

func gordianTestPlacement(t *testing.T, movable int) *Placement {
	t.Helper()
	p, err := NewPlacement(PlacementLimits{})
	if err != nil {
		t.Fatal(err)
	}
	p.CoreBounds = PlacementRect{X: 10, Y: 20, Width: 100, Height: 80}
	p.PadBounds = PlacementRect{X: 8, Y: 18, Width: 104, Height: 84}
	p.RowHeight = 2
	if err = p.AddCell(PlacementCell{ID: 0, Label: "fixed", Type: PlacementCellType{Width: 2, Height: 2}, Fixed: true, X: 12, Y: 22}); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < movable; i++ {
		id := i + 1
		if err = p.AddCell(PlacementCell{ID: id, Label: "c", Type: PlacementCellType{Width: float64(1 + i%3), Height: float64(1 + (i+1)%2)}, X: 20 + float64(i*7), Y: 30 + float64((i*13)%50)}); err != nil {
			t.Fatal(err)
		}
	}
	terms := make([]int, 0, movable+1)
	for i := 0; i <= movable; i++ {
		terms = append(terms, i)
	}
	if len(terms) >= 2 {
		if err = p.AddNet(PlacementNet{ID: 0, Terms: terms, Weight: 2}); err != nil {
			t.Fatal(err)
		}
	}
	return p
}

func gordianCollectLeaves(root *GordianPartition) []*GordianPartition {
	if root == nil {
		return nil
	}
	if root.Leaf {
		return []*GordianPartition{root}
	}
	return append(gordianCollectLeaves(root.First), gordianCollectLeaves(root.Second)...)
}

func TestGordianInventoryAndHelpers(t *testing.T) {
	want := []string{"place_base.c", "place_bin.c", "place_genqp.c", "place_gordian.c", "place_legalize.c", "place_pads.c", "place_partition.c", "place_qpsolver.c", "place_io.c", "place_inc.c"}
	if !slices.Equal(GordianBuildUnits[:], want) {
		t.Fatalf("inventory=%v", GordianBuildUnits)
	}
	cases := []struct {
		pins int
		want float64
	}{{0, 2}, {1, 2}, {2, 2}, {3, 1.5}, {5, 1.25}}
	for _, tc := range cases {
		got, err := GordianSplitPenalty(tc.pins)
		if err != nil || got != tc.want {
			t.Fatalf("pins=%d got=%v err=%v", tc.pins, got, err)
		}
	}
	if _, err := GordianSplitPenalty(-1); !errors.Is(err, ErrDimension) {
		t.Fatalf("negative=%v", err)
	}
	if got, err := GordianSquareHashID(7, 97); err != nil || got != (7*24)%97 {
		t.Fatalf("hash=%d err=%v", got, err)
	}
	if got, err := PlacementNameHash("abcd", 101); err != nil || got != (int('b')+2*int('c')+3*int('d'))%101 {
		t.Fatalf("name hash=%d err=%v", got, err)
	}
}

func TestGordianPartitionRefinement(t *testing.T) {
	p := gordianTestPlacement(t, 9)
	g, err := NewGordianPlacer(p, GordianOptions{LargestFinalSize: 2, MaxPartitions: 16})
	if err != nil {
		t.Fatal(err)
	}
	for pass := 0; pass < 8; pass++ {
		done, e := g.RefinePartitions()
		if e != nil {
			t.Fatal(e)
		}
		if done {
			break
		}
	}
	if err = g.Validate(); err != nil {
		t.Fatal(err)
	}
	root, _ := g.Root()
	leaves := gordianCollectLeaves(root)
	if len(leaves) < 5 || len(leaves) > 9 {
		t.Fatalf("leaves=%d", len(leaves))
	}
	seen := map[int]bool{}
	area := 0.0
	for _, leaf := range leaves {
		if len(leaf.Members) > 2 {
			t.Fatalf("large leaf=%v", leaf.Members)
		}
		for _, id := range leaf.Members {
			if seen[id] {
				t.Fatalf("duplicate %d", id)
			}
			seen[id] = true
			a, _ := p.CellArea(id)
			area += a
		}
	}
	if len(seen) != 9 || math.Abs(area-root.Area) > 1e-9 {
		t.Fatalf("members=%d area=%v root=%v", len(seen), area, root.Area)
	}
	// Snapshots are independent.
	root.Members[0] = 999
	again, _ := g.Root()
	if again.Members[0] == 999 {
		t.Fatal("partition snapshot aliases")
	}
}

func TestGordianQuadraticCliqueModel(t *testing.T) {
	p := gordianTestPlacement(t, 3)
	g, err := NewGordianPlacer(p, GordianOptions{LargestFinalSize: 8, UseCenters: true})
	if err != nil {
		t.Fatal(err)
	}
	problem, err := g.BuildQuadraticProblem(true)
	if err != nil {
		t.Fatal(err)
	}
	if len(problem.Edges) != 6 {
		t.Fatalf("edges=%v", problem.Edges)
	}
	// Four pins use penalty 1+1/3; every unordered pair has weight 1.5.
	for _, edge := range problem.Edges {
		if math.Abs(edge.Weight-1.5) > 1e-12 {
			t.Fatalf("edge=%+v", edge)
		}
	}
	if len(problem.Centers) != 1 || !slices.Equal(problem.Centers[0].Cells, []int{1, 2, 3}) {
		t.Fatalf("centers=%+v", problem.Centers)
	}
	before, _ := p.Cell(1)
	problem.Positions[1].X = 999
	after, _ := p.Cell(1)
	if before != after {
		t.Fatal("problem aliases placement")
	}
}

func TestGordianGlobalPlaceAndSanitize(t *testing.T) {
	p := gordianTestPlacement(t, 5)
	// Put a movable cell outside to exercise source cleanup.
	p.cells[1].X, p.cells[1].Y = -100, 500
	g, err := NewGordianPlacer(p, GordianOptions{LargestFinalSize: 2, MaxPartitions: 8, MaxIterations: 8, UseCenters: true})
	if err != nil {
		t.Fatal(err)
	}
	result, err := g.GlobalPlace()
	if err != nil {
		t.Fatal(err)
	}
	if result.Partitions < 3 || result.Partitions > 5 || result.Iterations == 0 || !placementFinite(result.Wirelength) {
		t.Fatalf("result=%+v", result)
	}
	for _, id := range result.Placement.CellIDs() {
		c, _ := result.Placement.Cell(id)
		if c.Fixed || c.Type.Pad {
			continue
		}
		if c.X-c.Type.Width/2 < result.Placement.CoreBounds.X-1e-9 || c.X+c.Type.Width/2 > result.Placement.CoreBounds.X+result.Placement.CoreBounds.Width+1e-9 || c.Y-c.Type.Height/2 < result.Placement.CoreBounds.Y-1e-9 || c.Y+c.Type.Height/2 > result.Placement.CoreBounds.Y+result.Placement.CoreBounds.Height+1e-9 {
			t.Fatalf("outside %+v", c)
		}
	}
	fixed, _ := result.Placement.Cell(0)
	if fixed.X != 12 || fixed.Y != 22 {
		t.Fatalf("fixed moved %+v", fixed)
	}
}

func TestGordianIncrementalAndNetDeletion(t *testing.T) {
	p := gordianTestPlacement(t, 4)
	g, err := NewGordianPlacer(p, GordianOptions{LargestFinalSize: 2, MaxPartitions: 8})
	if err != nil {
		t.Fatal(err)
	}
	_, _ = g.RefinePartitions()
	updated, _ := p.Clone()
	if err = updated.AddCell(PlacementCell{ID: 9, Label: "new", Type: PlacementCellType{Width: 2, Height: 2}, X: 90, Y: 70}); err != nil {
		t.Fatal(err)
	}
	if err = g.AddIncrementalCells(updated); err != nil {
		t.Fatal(err)
	}
	root, _ := g.Root()
	count := 0
	for _, id := range root.Members {
		if id == 9 {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("new count=%d members=%v", count, root.Members)
	}
	if err = PlacementDeleteNetConnections(p, 2); err != nil {
		t.Fatal(err)
	}
	net, _ := p.Net(0)
	if slices.Contains(net.Terms, 2) {
		t.Fatalf("terms=%v", net.Terms)
	}
	if err = p.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestGordianIncrementalReopensCompletedTree(t *testing.T) {
	p := gordianTestPlacement(t, 2)
	g, err := NewGordianPlacer(p, GordianOptions{LargestFinalSize: 2, MaxPartitions: 8})
	if err != nil {
		t.Fatal(err)
	}
	root, _ := g.Root()
	if !root.Done {
		t.Fatal("initial tree should be complete")
	}
	updated, _ := p.Clone()
	for _, cell := range []PlacementCell{
		{ID: 9, Label: "new9", Type: PlacementCellType{Width: 2, Height: 2}, X: 35, Y: 45},
		{ID: 11, Label: "new11", Type: PlacementCellType{Width: 2, Height: 2}, X: 75, Y: 65},
	} {
		if err = updated.AddCell(cell); err != nil {
			t.Fatal(err)
		}
	}
	if err = g.AddIncrementalCells(updated); err != nil {
		t.Fatal(err)
	}
	root, _ = g.Root()
	if root.Done {
		t.Fatal("new members must reopen a completed partition")
	}
	for pass := 0; pass < 8; pass++ {
		done, refineErr := g.RefinePartitions()
		if refineErr != nil {
			t.Fatal(refineErr)
		}
		if done {
			break
		}
	}
	root, _ = g.Root()
	for _, leaf := range gordianCollectLeaves(root) {
		if len(leaf.Members) > 2 {
			t.Fatalf("unrefined leaf=%v", leaf.Members)
		}
	}
	bad, _ := updated.Clone()
	bad.CoreBounds.Width++
	before, _ := g.Root()
	if err = g.AddIncrementalCells(bad); !errors.Is(err, ErrDimension) {
		t.Fatalf("changed core err=%v", err)
	}
	after, _ := g.Root()
	if !gordianPartitionsEqual(before, after) {
		t.Fatal("rejected core change mutated tree")
	}
}

func TestGordianQuadraticLimitsAndInnerCancellation(t *testing.T) {
	p := gordianTestPlacement(t, 24)
	g, err := NewGordianPlacer(p, GordianOptions{LargestFinalSize: 24, MaxNetSize: 32, MaxEdges: 10})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = g.BuildQuadraticProblem(false); !errors.Is(err, ErrLimit) {
		t.Fatalf("edge limit err=%v", err)
	}
	calls, cancelAt := 0, int(^uint(0)>>1)
	g, err = NewGordianPlacer(p, GordianOptions{LargestFinalSize: 24, MaxNetSize: 32, MaxEdges: 1000, Stop: func() bool {
		calls++
		return calls == cancelAt
	}})
	if err != nil {
		t.Fatal(err)
	}
	calls, cancelAt = 0, 4 // check, net, first pair poll, then the 256th pair poll
	if _, err = g.BuildQuadraticProblem(false); !errors.Is(err, ErrCancelled) {
		t.Fatalf("inner cancellation err=%v calls=%d", err, calls)
	}
}

func TestGordianGlobalIncremental(t *testing.T) {
	p := gordianTestPlacement(t, 4)
	g, err := NewGordianPlacer(p, GordianOptions{LargestFinalSize: 2, MaxPartitions: 8, UseCenters: true})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = g.RefinePartitions(); err != nil {
		t.Fatal(err)
	}
	updated, _ := p.Clone()
	if err = updated.AddCell(PlacementCell{ID: 9, Label: "new", Type: PlacementCellType{Width: 2, Height: 2}, X: 80, Y: 60}); err != nil {
		t.Fatal(err)
	}
	if err = updated.AddNet(PlacementNet{ID: 2, Terms: []int{0, 9}, Weight: 1}); err != nil {
		t.Fatal(err)
	}
	result, err := g.GlobalIncremental(updated)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := result.Placement.Cell(9); !ok || result.Iterations != 1 {
		t.Fatalf("result=%+v", result)
	}
	root := result.Root
	if !slices.Contains(root.Members, 9) {
		t.Fatalf("members=%v", root.Members)
	}
}

func TestGordianFastEstimate(t *testing.T) {
	p := gordianTestPlacement(t, 2)
	if err := p.AddNet(PlacementNet{ID: 1, Terms: []int{0, 1}, Weight: 1}); err != nil {
		t.Fatal(err)
	}
	// The second movable is outside the selected net's box.
	p.cells[2].X, p.cells[2].Y = 100, 100
	box, err := p.NetBBox(1)
	if err != nil {
		t.Fatal(err)
	}
	got, err := GordianFastEstimate(p, 2, []int{1})
	if err != nil {
		t.Fatal(err)
	}
	want := 0.0
	if p.cells[2].X < box.X {
		want += box.X - p.cells[2].X
	} else if p.cells[2].X > box.X+box.Width {
		want += p.cells[2].X - box.X - box.Width
	}
	if p.cells[2].Y < box.Y {
		want += box.Y - p.cells[2].Y
	} else if p.cells[2].Y > box.Y+box.Height {
		want += p.cells[2].Y - box.Y - box.Height
	}
	if got != want {
		t.Fatalf("got=%v want=%v box=%+v", got, want, box)
	}
}

func TestGordianCancellationAtomicity(t *testing.T) {
	p := gordianTestPlacement(t, 8)
	calls, cancelAt := 0, int(^uint(0)>>1)
	g, err := NewGordianPlacer(p, GordianOptions{LargestFinalSize: 1, Stop: func() bool {
		calls++
		return calls == cancelAt
	}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = g.RefinePartitions(); err != nil {
		t.Fatal(err)
	}
	before, _ := g.Root()
	calls, cancelAt = 0, 4 // cancel after the first child split on the cloned tree
	_, err = g.RefinePartitions()
	if !errors.Is(err, ErrCancelled) {
		t.Fatalf("err=%v", err)
	}
	after, _ := g.Root()
	if !gordianPartitionsEqual(before, after) {
		t.Fatal("cancelled refinement mutated state")
	}
}

func TestGordianGlobalFailureIsAtomic(t *testing.T) {
	p := gordianTestPlacement(t, 8)
	g, err := NewGordianPlacer(p, GordianOptions{LargestFinalSize: 1, MaxIterations: 1, UseCenters: true})
	if err != nil {
		t.Fatal(err)
	}
	beforePlacement, _ := g.Placement()
	beforeRoot, _ := g.Root()
	if _, err = g.GlobalPlace(); !errors.Is(err, ErrLimit) {
		t.Fatalf("err=%v", err)
	}
	afterPlacement, _ := g.Placement()
	afterRoot, _ := g.Root()
	for _, id := range beforePlacement.CellIDs() {
		a, _ := beforePlacement.Cell(id)
		b, _ := afterPlacement.Cell(id)
		if a != b {
			t.Fatalf("cell %d changed: before=%+v after=%+v", id, a, b)
		}
	}
	if !gordianPartitionsEqual(beforeRoot, afterRoot) {
		t.Fatal("failed global placement changed partitions")
	}
}

func gordianPartitionsEqual(a, b *GordianPartition) bool {
	if a == nil || b == nil {
		return a == b
	}
	return slices.Equal(a.Members, b.Members) && a.Bounds == b.Bounds && a.Done == b.Done && a.Leaf == b.Leaf && a.Vertical == b.Vertical && a.Area == b.Area && a.Level == b.Level && gordianPartitionsEqual(a.First, b.First) && gordianPartitionsEqual(a.Second, b.Second)
}

func TestGordianIndependentPartitionModel5000(t *testing.T) {
	state := uint64(0x8f6e37a1)
	next := func() uint64 { state ^= state << 7; state ^= state >> 9; state ^= state << 8; return state }
	for trial := 0; trial < 5000; trial++ {
		count := 2 + int(next()%15)
		p := gordianTestPlacement(t, count)
		for id := 1; id <= count; id++ {
			p.cells[id].X = 10 + float64(next()%10000)/100
			p.cells[id].Y = 20 + float64(next()%8000)/100
		}
		limit := 1 + int(next()%5)
		g, err := NewGordianPlacer(p, GordianOptions{LargestFinalSize: limit, MaxPartitions: count + 1})
		if err != nil {
			t.Fatalf("trial=%d: %v", trial, err)
		}
		for pass := 0; pass < 16; pass++ {
			done, e := g.RefinePartitions()
			if e != nil {
				t.Fatalf("trial=%d: %v", trial, e)
			}
			if done {
				break
			}
		}
		if err = g.Validate(); err != nil {
			t.Fatalf("trial=%d: %v", trial, err)
		}
		root, _ := g.Root()
		leaves := gordianCollectLeaves(root)
		seen := map[int]bool{}
		for _, leaf := range leaves {
			if len(leaf.Members) > limit {
				t.Fatalf("trial=%d leaf=%d", trial, len(leaf.Members))
			}
			for _, id := range leaf.Members {
				if seen[id] {
					t.Fatalf("trial=%d duplicate=%d", trial, id)
				}
				seen[id] = true
			}
		}
		if len(seen) != count {
			t.Fatalf("trial=%d seen=%d count=%d", trial, len(seen), count)
		}
	}
}

func FuzzGordianPartitionModel(f *testing.F) {
	for _, seed := range []uint64{0, 1, 2, 3, 0xdeadbeef, math.MaxUint32} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, seed uint64) {
		next := func() uint64 { seed = seed*6364136223846793005 + 1442695040888963407; return seed }
		count := 2 + int(next()%24)
		p := gordianTestPlacement(t, count)
		for id := 1; id <= count; id++ {
			p.cells[id].X = 10 + float64(next()%1000)/10
			p.cells[id].Y = 20 + float64(next()%800)/10
		}
		limit := 1 + int(next()%7)
		g, err := NewGordianPlacer(p, GordianOptions{LargestFinalSize: limit, MaxPartitions: count + 1})
		if err != nil {
			t.Fatal(err)
		}
		for pass := 0; pass < 16; pass++ {
			done, e := g.RefinePartitions()
			if e != nil {
				t.Fatal(e)
			}
			if done {
				break
			}
		}
		if err = g.Validate(); err != nil {
			t.Fatal(err)
		}
		root, _ := g.Root()
		seen := map[int]bool{}
		for _, leaf := range gordianCollectLeaves(root) {
			for _, id := range leaf.Members {
				if seen[id] {
					t.Fatalf("duplicate %d", id)
				}
				seen[id] = true
			}
		}
		if len(seen) != count {
			t.Fatalf("seen=%d count=%d", len(seen), count)
		}
	})
}
