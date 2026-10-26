package abc

// Placement partitioning and Gordian flow from src/phys/place/place_genqp.c,
// place_gordian.c, place_gordian.h, place_inc.c, place_partition.c,
// place_test.c, and module.make.

import (
	"math"
	"sort"
)

const (
	gordianCliquePenalty   = 1.0
	gordianIgnoreNetSize   = 20
	gordianLargestFinal    = 20
	gordianDefaultMaxParts = 1 << 20
	gordianDefaultMaxEdges = 1 << 22
)

// GordianBuildUnits is the exact source inventory selected by module.make.
var GordianBuildUnits = [...]string{
	"place_base.c", "place_bin.c", "place_genqp.c", "place_gordian.c",
	"place_legalize.c", "place_pads.c", "place_partition.c",
	"place_qpsolver.c", "place_io.c", "place_inc.c",
}

// GordianOptions bounds analytic placement and deterministic partitioning.
// Zero integer limits select conservative defaults.
type GordianOptions struct {
	LargestFinalSize int
	MaxPartitions    int
	MaxIterations    int
	MaxNetSize       int
	MaxEdges         int
	DensityBins      int
	MaxMovement      float64
	UseCenters       bool
	Stop             func() bool
}

// GordianPartition is an owned binary partition. Member IDs retain source
// identity and are deterministic within each node.
type GordianPartition struct {
	Members       []int
	Bounds        PlacementRect
	Done          bool
	Leaf          bool
	Vertical      bool
	Area          float64
	Level         int
	First, Second *GordianPartition
}

// GordianResult owns the final placement and summary data.
type GordianResult struct {
	Placement  *Placement
	Root       *GordianPartition
	Iterations int
	Partitions int
	Wirelength float64
}

// GordianPlacer owns placement and partition state; no package globals are
// used. Mutating operations commit only after all checks succeed.
type GordianPlacer struct {
	placement  *Placement
	root       *GordianPartition
	options    GordianOptions
	partitions int
	known      map[int]struct{}
	closed     bool
}

func normalizeGordianOptions(options GordianOptions) (GordianOptions, error) {
	if options.LargestFinalSize < 0 || options.MaxPartitions < 0 ||
		options.MaxIterations < 0 || options.MaxNetSize < 0 ||
		options.MaxEdges < 0 || options.DensityBins < 0 || options.MaxMovement < 0 ||
		!placementFinite(options.MaxMovement) {
		return options, ErrDimension
	}
	if options.LargestFinalSize == 0 {
		options.LargestFinalSize = gordianLargestFinal
	}
	if options.MaxPartitions == 0 {
		options.MaxPartitions = gordianDefaultMaxParts
	}
	if options.MaxIterations == 0 {
		options.MaxIterations = 64
	}
	if options.MaxNetSize == 0 {
		options.MaxNetSize = gordianIgnoreNetSize
	}
	if options.MaxEdges == 0 {
		options.MaxEdges = gordianDefaultMaxEdges
	}
	if options.DensityBins == 0 {
		options.DensityBins = 25
	}
	return options, nil
}

func gordianCancelled(options GordianOptions) bool {
	return options.Stop != nil && options.Stop()
}

func gordianClonePartition(part *GordianPartition) *GordianPartition {
	if part == nil {
		return nil
	}
	result := *part
	result.Members = append([]int(nil), part.Members...)
	result.First = gordianClonePartition(part.First)
	result.Second = gordianClonePartition(part.Second)
	return &result
}

func gordianCellArea(cell *PlacementCell) float64 {
	return cell.Type.Width * cell.Type.Height
}

func NewGordianPlacer(placement *Placement, options GordianOptions) (*GordianPlacer, error) {
	options, err := normalizeGordianOptions(options)
	if err != nil || placement == nil {
		return nil, ErrDimension
	}
	if gordianCancelled(options) {
		return nil, ErrCancelled
	}
	if err = placement.Validate(); err != nil {
		return nil, err
	}
	if placement.CoreBounds.Width <= 0 || placement.CoreBounds.Height <= 0 {
		return nil, ErrDimension
	}
	owned, err := placement.Clone()
	if err != nil {
		return nil, err
	}
	root := &GordianPartition{Bounds: owned.CoreBounds, Leaf: true}
	known := make(map[int]struct{})
	for _, id := range owned.CellIDs() {
		cell := owned.cells[id]
		if cell.Fixed || cell.Type.Pad {
			continue
		}
		root.Members = append(root.Members, id)
		root.Area += gordianCellArea(cell)
		known[id] = struct{}{}
	}
	if len(root.Members) == 0 || root.Area <= 0 {
		return nil, ErrDimension
	}
	root.Done = len(root.Members) <= options.LargestFinalSize
	result := &GordianPlacer{placement: owned, root: root, options: options, partitions: 1, known: known}
	if err = result.Validate(); err != nil {
		return nil, err
	}
	return result, nil
}

func (placer *GordianPlacer) check() error {
	if placer == nil || placer.closed || placer.placement == nil || placer.root == nil {
		return ErrInvariant
	}
	if gordianCancelled(placer.options) {
		return ErrCancelled
	}
	return nil
}

func (placer *GordianPlacer) Close() error {
	if placer == nil || placer.closed {
		return ErrInvariant
	}
	placer.placement, placer.root, placer.known = nil, nil, nil
	placer.closed = true
	return nil
}

func (placer *GordianPlacer) Placement() (*Placement, error) {
	if err := placer.check(); err != nil {
		return nil, err
	}
	return placer.placement.Clone()
}

func (placer *GordianPlacer) Root() (*GordianPartition, error) {
	if err := placer.check(); err != nil {
		return nil, err
	}
	return gordianClonePartition(placer.root), nil
}

func (placer *GordianPlacer) validatePartition(part *GordianPartition, parent map[int]bool) (int, error) {
	if part == nil || part.Area <= 0 || !placementFinite(part.Area, part.Bounds.X, part.Bounds.Y, part.Bounds.Width, part.Bounds.Height) || part.Bounds.Width < 0 || part.Bounds.Height < 0 {
		return 0, ErrInvariant
	}
	seen := make(map[int]bool, len(part.Members))
	area := 0.0
	for _, id := range part.Members {
		if seen[id] || id < 0 || id >= len(placer.placement.cells) || placer.placement.cells[id] == nil || placer.placement.cells[id].Fixed || placer.placement.cells[id].Type.Pad {
			return 0, ErrInvariant
		}
		if parent != nil && !parent[id] {
			return 0, ErrInvariant
		}
		seen[id] = true
		area += gordianCellArea(placer.placement.cells[id])
	}
	if math.Abs(area-part.Area) > 1e-9*math.Max(1, area) {
		return 0, ErrInvariant
	}
	if part.Leaf {
		if part.First != nil || part.Second != nil {
			return 0, ErrInvariant
		}
		if part.Done != (len(part.Members) <= placer.options.LargestFinalSize) {
			return 0, ErrInvariant
		}
		return 1, nil
	}
	if part.First == nil || part.Second == nil || part.First.Level != part.Level+1 || part.Second.Level != part.Level+1 {
		return 0, ErrInvariant
	}
	left := make(map[int]bool, len(part.First.Members))
	for _, id := range part.First.Members {
		left[id] = true
	}
	for _, id := range part.Second.Members {
		if left[id] {
			return 0, ErrInvariant
		}
		left[id] = true
	}
	if len(left) != len(seen) {
		return 0, ErrInvariant
	}
	for id := range seen {
		if !left[id] {
			return 0, ErrInvariant
		}
	}
	a, err := placer.validatePartition(part.First, seen)
	if err != nil {
		return 0, err
	}
	b, err := placer.validatePartition(part.Second, seen)
	if err != nil {
		return 0, err
	}
	if part.Done != (part.First.Done && part.Second.Done) {
		return 0, ErrInvariant
	}
	return a + b, nil
}

func (placer *GordianPlacer) Validate() error {
	if placer == nil || placer.closed || placer.placement == nil || placer.root == nil || placer.partitions < 1 || placer.partitions > placer.options.MaxPartitions {
		return ErrInvariant
	}
	if err := placer.placement.Validate(); err != nil {
		return err
	}
	leaves, err := placer.validatePartition(placer.root, nil)
	if err != nil || leaves != placer.partitions {
		return ErrInvariant
	}
	return nil
}

// GordianSplitPenalty preserves splitPenalty.
func GordianSplitPenalty(pins int) (float64, error) {
	if pins < 0 {
		return 0, ErrDimension
	}
	if pins > 1 {
		return 1 + gordianCliquePenalty/float64(pins-1), nil
	}
	return 1 + gordianCliquePenalty, nil
}

func (placer *GordianPlacer) split(part *GordianPartition) error {
	if len(part.Members) <= placer.options.LargestFinalSize {
		part.Done = true
		return nil
	}
	if placer.partitions >= placer.options.MaxPartitions {
		return ErrLimit
	}
	members := append([]int(nil), part.Members...)
	sort.SliceStable(members, func(i, j int) bool {
		a, b := placer.placement.cells[members[i]], placer.placement.cells[members[j]]
		av, bv := a.Y, b.Y
		if part.Vertical {
			av, bv = a.X, b.X
		}
		return av < bv || av == bv && members[i] < members[j]
	})
	half, area, cut := part.Area/2, 0.0, 0
	for cut < len(members)-1 && area < half {
		area += gordianCellArea(placer.placement.cells[members[cut]])
		cut++
	}
	if cut == 0 || cut == len(members) {
		return ErrInvariant
	}
	first := &GordianPartition{Members: append([]int(nil), members[:cut]...), Bounds: part.Bounds, Leaf: true, Vertical: !part.Vertical, Level: part.Level + 1, Area: area}
	second := &GordianPartition{Members: append([]int(nil), members[cut:]...), Bounds: part.Bounds, Leaf: true, Vertical: !part.Vertical, Level: part.Level + 1, Area: part.Area - area}
	first.Done = len(first.Members) <= placer.options.LargestFinalSize
	second.Done = len(second.Members) <= placer.options.LargestFinalSize
	ratio := first.Area / part.Area
	if part.Vertical {
		first.Bounds.Width = part.Bounds.Width * ratio
		second.Bounds.X = part.Bounds.X + first.Bounds.Width
		second.Bounds.Width = part.Bounds.Width - first.Bounds.Width
	} else {
		first.Bounds.Height = part.Bounds.Height * ratio
		second.Bounds.Y = part.Bounds.Y + first.Bounds.Height
		second.Bounds.Height = part.Bounds.Height - first.Bounds.Height
	}
	part.Members = members
	part.First, part.Second, part.Leaf = first, second, false
	part.Done = first.Done && second.Done
	placer.partitions++
	return nil
}

func (placer *GordianPlacer) refine(part *GordianPartition) (bool, error) {
	if gordianCancelled(placer.options) {
		return false, ErrCancelled
	}
	if part.Done {
		return true, nil
	}
	if part.Leaf {
		if err := placer.split(part); err != nil {
			return false, err
		}
		return part.Done, nil
	}
	a, err := placer.refine(part.First)
	if err != nil {
		return false, err
	}
	b, err := placer.refine(part.Second)
	if err != nil {
		return false, err
	}
	part.Done = a && b
	return part.Done, nil
}

// RefinePartitions performs one source-shaped refinement pass atomically.
func (placer *GordianPlacer) RefinePartitions() (bool, error) {
	if err := placer.check(); err != nil {
		return false, err
	}
	root := gordianClonePartition(placer.root)
	parts := placer.partitions
	done, err := placer.refine(root)
	if err != nil {
		placer.partitions = parts
		return false, err
	}
	old := placer.root
	placer.root = root
	if err = placer.Validate(); err != nil {
		placer.root, placer.partitions = old, parts
		return false, err
	}
	return done, nil
}

func gordianLeaves(root *GordianPartition) []*GordianPartition {
	result, stack := make([]*GordianPartition, 0), []*GordianPartition{root}
	for len(stack) != 0 {
		part := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if part.Leaf {
			result = append(result, part)
			continue
		}
		stack = append(stack, part.Second, part.First)
	}
	return result
}

func gordianRefreshDone(part *GordianPartition, largest int) bool {
	if part.Leaf {
		part.Done = len(part.Members) <= largest
		return part.Done
	}
	part.Done = gordianRefreshDone(part.First, largest) && gordianRefreshDone(part.Second, largest)
	return part.Done
}

// BuildQuadraticProblem constructs the source clique model. Nets larger than
// MaxNetSize are ignored and repeated cell pairs have their weights combined.
func (placer *GordianPlacer) BuildQuadraticProblem(useCenters bool) (QPSProblem, error) {
	if err := placer.check(); err != nil {
		return QPSProblem{}, err
	}
	n := len(placer.placement.cells)
	problem := QPSProblem{Positions: make([]QPSPoint, n), Fixed: make([]bool, n), Areas: make([]float64, n)}
	for id := 0; id < n; id++ {
		cell := placer.placement.cells[id]
		if cell == nil {
			problem.Fixed[id], problem.Areas[id] = true, 1
			continue
		}
		problem.Positions[id] = QPSPoint{X: cell.X, Y: cell.Y}
		problem.Fixed[id] = cell.Fixed || cell.Type.Pad
		problem.Areas[id] = gordianCellArea(cell)
	}
	type pair struct{ a, b int }
	weights := make(map[pair]float64)
	pairWork := 0
	for _, net := range placer.placement.nets {
		if gordianCancelled(placer.options) {
			return QPSProblem{}, ErrCancelled
		}
		if net == nil || len(net.Terms) < 2 || len(net.Terms) > placer.options.MaxNetSize {
			continue
		}
		penalty, _ := GordianSplitPenalty(len(net.Terms))
		weight := net.Weight / penalty
		for i, a := range net.Terms {
			for _, b := range net.Terms[i+1:] {
				if pairWork%256 == 0 && gordianCancelled(placer.options) {
					return QPSProblem{}, ErrCancelled
				}
				pairWork++
				x, y := a, b
				if x == y {
					continue
				}
				if x > y {
					x, y = y, x
				}
				key := pair{x, y}
				if _, exists := weights[key]; !exists && len(weights) >= placer.options.MaxEdges {
					return QPSProblem{}, ErrLimit
				}
				weights[key] += weight
			}
		}
	}
	keys := make([]pair, 0, len(weights))
	for key := range weights {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i].a < keys[j].a || keys[i].a == keys[j].a && keys[i].b < keys[j].b })
	for _, key := range keys {
		problem.Edges = append(problem.Edges, QPSEdge{CellA: key.a, CellB: key.b, Weight: weights[key]})
	}
	if useCenters {
		for _, leaf := range gordianLeaves(placer.root) {
			cells := make([]int, 0, len(leaf.Members))
			for _, id := range leaf.Members {
				if !problem.Fixed[id] && problem.Areas[id] > 0 {
					cells = append(cells, id)
				}
			}
			if len(cells) != 0 {
				problem.Centers = append(problem.Centers, QPSCenterConstraint{Cells: cells, Center: QPSPoint{X: leaf.Bounds.X + leaf.Bounds.Width/2, Y: leaf.Bounds.Y + leaf.Bounds.Height/2}})
			}
		}
	}
	return problem, nil
}

func (placer *GordianPlacer) solve(useCenters bool) error {
	problem, err := placer.BuildQuadraticProblem(useCenters)
	if err != nil {
		return err
	}
	result, err := SolveQPS(problem, QPSOptions{MaxCells: len(problem.Positions) + 1, MaxEdges: placer.options.MaxEdges, MaxConstraints: len(problem.Centers) + 1, Stop: placer.options.Stop})
	if err != nil {
		return err
	}
	work, _ := placer.placement.Clone()
	for id, point := range result.Positions {
		if id < len(work.cells) && work.cells[id] != nil && !work.cells[id].Fixed && !work.cells[id].Type.Pad {
			work.cells[id].X, work.cells[id].Y = point.X, point.Y
		}
	}
	if err = work.Validate(); err != nil {
		return err
	}
	placer.placement = work
	return nil
}

// Sanitize moves movable cell centers inside the core while preserving a
// deterministic margin based on source row height.
func (placer *GordianPlacer) Sanitize() error {
	if err := placer.check(); err != nil {
		return err
	}
	work, _ := placer.placement.Clone()
	for _, id := range work.CellIDs() {
		if gordianCancelled(placer.options) {
			return ErrCancelled
		}
		cell := work.cells[id]
		if cell.Fixed || cell.Type.Pad {
			continue
		}
		hx, hy := cell.Type.Width/2, cell.Type.Height/2
		if 2*hx > work.CoreBounds.Width || 2*hy > work.CoreBounds.Height {
			return ErrDimension
		}
		var err error
		cell.X, err = placementClampCenter(cell.X, hx, work.CoreBounds.X, work.CoreBounds.Width)
		if err != nil {
			return err
		}
		cell.Y, err = placementClampCenter(cell.Y, hy, work.CoreBounds.Y, work.CoreBounds.Height)
		if err != nil {
			return err
		}
	}
	placer.placement = work
	return nil
}

// GlobalPlace performs bounded source-shaped solve/refine iterations.
func (placer *GordianPlacer) GlobalPlace() (result GordianResult, err error) {
	if err = placer.check(); err != nil {
		return GordianResult{}, err
	}
	originalPlacement, _ := placer.placement.Clone()
	originalRoot := gordianClonePartition(placer.root)
	originalPartitions := placer.partitions
	defer func() {
		if err != nil {
			placer.placement = originalPlacement
			placer.root = originalRoot
			placer.partitions = originalPartitions
		}
	}()
	for iteration := 0; iteration < placer.options.MaxIterations; iteration++ {
		if err = placer.solve(placer.options.UseCenters); err != nil {
			return GordianResult{}, err
		}
		var done bool
		done, err = placer.RefinePartitions()
		if err != nil {
			return GordianResult{}, err
		}
		if done {
			if err = placer.solve(placer.options.UseCenters); err != nil {
				return GordianResult{}, err
			}
			if err = placer.Sanitize(); err != nil {
				return GordianResult{}, err
			}
			if placer.options.MaxMovement > 0 {
				if err = placer.placement.FixDensity(placer.options.DensityBins, placer.options.MaxMovement); err != nil {
					return GordianResult{}, err
				}
			}
			wire, err := placer.placement.TotalWirelength()
			if err != nil {
				return GordianResult{}, err
			}
			owned, _ := placer.placement.Clone()
			return GordianResult{Placement: owned, Root: gordianClonePartition(placer.root), Iterations: iteration + 1, Partitions: placer.partitions, Wirelength: wire}, nil
		}
	}
	return GordianResult{}, ErrLimit
}

// GlobalIncremental adds newly introduced movable cells to the retained
// partition tree, performs one constrained quadratic solve, sanitizes the
// result, and optionally spreads density. The whole update is transactional.
func (placer *GordianPlacer) GlobalIncremental(updated *Placement) (result GordianResult, err error) {
	if err = placer.check(); err != nil {
		return GordianResult{}, err
	}
	originalPlacement, _ := placer.placement.Clone()
	originalRoot := gordianClonePartition(placer.root)
	originalKnown := make(map[int]struct{}, len(placer.known))
	for id := range placer.known {
		originalKnown[id] = struct{}{}
	}
	defer func() {
		if err != nil {
			placer.placement = originalPlacement
			placer.root = originalRoot
			placer.known = originalKnown
		}
	}()
	if err = placer.AddIncrementalCells(updated); err != nil {
		return GordianResult{}, err
	}
	if err = placer.solve(placer.options.UseCenters); err != nil {
		return GordianResult{}, err
	}
	if err = placer.Sanitize(); err != nil {
		return GordianResult{}, err
	}
	if placer.options.MaxMovement > 0 {
		if err = placer.placement.FixDensity(placer.options.DensityBins, placer.options.MaxMovement); err != nil {
			return GordianResult{}, err
		}
	}
	wire, err := placer.placement.TotalWirelength()
	if err != nil {
		return GordianResult{}, err
	}
	owned, _ := placer.placement.Clone()
	return GordianResult{Placement: owned, Root: gordianClonePartition(placer.root), Iterations: 1, Partitions: placer.partitions, Wirelength: wire}, nil
}

// AddIncrementalCells inserts newly added movable cells through the existing
// partition cuts. Existing members and bounds are retained.
func (placer *GordianPlacer) AddIncrementalCells(updated *Placement) error {
	if err := placer.check(); err != nil {
		return err
	}
	if updated == nil {
		return ErrDimension
	}
	if err := updated.Validate(); err != nil {
		return err
	}
	work, _ := updated.Clone()
	if work.CoreBounds != placer.placement.CoreBounds {
		return ErrDimension
	}
	root := gordianClonePartition(placer.root)
	known := make(map[int]struct{}, len(placer.known))
	for id := range placer.known {
		known[id] = struct{}{}
	}
	var insert func(*GordianPartition, int) error
	insert = func(part *GordianPartition, id int) error {
		if gordianCancelled(placer.options) {
			return ErrCancelled
		}
		part.Members = append(part.Members, id)
		part.Area += gordianCellArea(work.cells[id])
		if part.Leaf {
			part.Done = len(part.Members) <= placer.options.LargestFinalSize
			return nil
		}
		cell := work.cells[id]
		if part.Vertical {
			if cell.X < part.Second.Bounds.X {
				return insert(part.First, id)
			}
		} else if cell.Y < part.Second.Bounds.Y {
			return insert(part.First, id)
		}
		return insert(part.Second, id)
	}
	for _, id := range work.CellIDs() {
		cell := work.cells[id]
		if cell.Fixed || cell.Type.Pad {
			continue
		}
		if _, ok := known[id]; ok {
			continue
		}
		if err := insert(root, id); err != nil {
			return err
		}
		known[id] = struct{}{}
	}
	gordianRefreshDone(root, placer.options.LargestFinalSize)
	oldPlacement, oldRoot, oldKnown := placer.placement, placer.root, placer.known
	placer.placement, placer.root, placer.known = work, root, known
	if err := placer.Validate(); err != nil {
		placer.placement, placer.root, placer.known = oldPlacement, oldRoot, oldKnown
		return err
	}
	return nil
}

// GordianFastEstimate returns the Manhattan distance from a cell to the
// bounding boxes of selected nets, correcting the source's Y/X typo.
func GordianFastEstimate(placement *Placement, cellID int, netIDs []int) (float64, error) {
	if placement == nil || cellID < 0 || cellID >= len(placement.cells) || placement.cells[cellID] == nil {
		return 0, ErrPlacementMissing
	}
	cell := placement.cells[cellID]
	total := 0.0
	for _, id := range netIDs {
		box, err := placement.NetBBox(id)
		if err != nil {
			return 0, err
		}
		if cell.X < box.X {
			total += box.X - cell.X
		} else if cell.X > box.X+box.Width {
			total += cell.X - box.X - box.Width
		}
		if cell.Y < box.Y {
			total += box.Y - cell.Y
		} else if cell.Y > box.Y+box.Height {
			total += cell.Y - box.Y - box.Height
		}
	}
	return total, nil
}

func GordianSquareHashID(id, maximum int) (int, error) {
	if id < 0 || maximum <= 0 {
		return 0, ErrDimension
	}
	value := uint64(id) * (uint64(id) + 17)
	return int(value % uint64(maximum)), nil
}

// PlacementDeleteNetConnections removes every occurrence of a cell from all
// nets transactionally, absorbing the test driver's pointer-array rewrite.
func PlacementDeleteNetConnections(placement *Placement, cellID int) error {
	if placement == nil || cellID < 0 || cellID >= len(placement.cells) || placement.cells[cellID] == nil {
		return ErrPlacementMissing
	}
	work, _ := placement.Clone()
	for _, net := range work.nets {
		if work.cancelled() {
			return ErrCancelled
		}
		if net == nil {
			continue
		}
		terms := net.Terms[:0]
		for _, term := range net.Terms {
			if term != cellID {
				terms = append(terms, term)
			} else {
				work.termCount--
			}
		}
		net.Terms = append([]int(nil), terms...)
	}
	if err := work.Validate(); err != nil {
		return err
	}
	*placement = *work
	return nil
}

func PlacementNameHash(name string, maximum int) (int, error) {
	if maximum <= 0 {
		return 0, ErrDimension
	}
	value := uint64(0)
	for i, b := range []byte(name) {
		value += uint64(b) * uint64(i)
	}
	return int(value % uint64(maximum)), nil
}
