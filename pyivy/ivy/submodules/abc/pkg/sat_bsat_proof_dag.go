package abc

// Port of the enabled proof-DAG machinery in src/sat/bsat/satProof.c.

import "slices"

type SatProofDAGLimits struct {
	MaxNodes, MaxAntecedents, MaxClauseID int
	Stop                                  func() bool
}

type SatProofDAGRef struct {
	Node, Clause int // exactly one is nonnegative
	PivotA       bool
}

type SatProofDAGNode struct {
	Learned, Mark, PartA bool
	Label                int
	Antecedents          []SatProofDAGRef
}

type SatProofDAGStats struct {
	Nodes, Antecedents int
	Stopped            bool
}

type SatProofDAGReduction struct {
	Roots, OldToNew []int
	Pivot           int
}

type SatProofDAG struct {
	nodes       []SatProofDAGNode
	antecedents int
	limits      SatProofDAGLimits
	stopped     bool
}

func (d *SatProofDAG) clone() *SatProofDAG {
	if d == nil {
		return nil
	}
	clone := *d
	clone.nodes = make([]SatProofDAGNode, len(d.nodes))
	for index, node := range d.nodes {
		clone.nodes[index] = node
		clone.nodes[index].Antecedents = slices.Clone(node.Antecedents)
	}
	return &clone
}

func MakeSatProofDAGNodeRef(node int, pivotA bool) (SatProofDAGRef, error) {
	if node < 0 {
		return SatProofDAGRef{}, ErrDimension
	}
	return SatProofDAGRef{Node: node, Clause: -1, PivotA: pivotA}, nil
}

func MakeSatProofDAGClauseRef(clause int, pivotA bool) (SatProofDAGRef, error) {
	if clause < 0 {
		return SatProofDAGRef{}, ErrDimension
	}
	return SatProofDAGRef{Node: -1, Clause: clause, PivotA: pivotA}, nil
}

func SatProofDAGRecordWords(antecedents int) (int, error) {
	if antecedents <= 0 {
		return 0, ErrDimension
	}
	return 1 + ((antecedents + 1) >> 1), nil
}

func checkSatProofDAGLimits(limits SatProofDAGLimits) (SatProofDAGLimits, error) {
	if limits.MaxNodes < 0 || limits.MaxAntecedents < 0 || limits.MaxClauseID < 0 {
		return limits, ErrDimension
	}
	if limits.MaxNodes == 0 {
		limits.MaxNodes = 16 << 20
	}
	if limits.MaxAntecedents == 0 {
		limits.MaxAntecedents = 64 << 20
	}
	if limits.MaxClauseID == 0 {
		limits.MaxClauseID = 64 << 20
	}
	return limits, nil
}

func NewSatProofDAG(limits SatProofDAGLimits) (*SatProofDAG, error) {
	checked, err := checkSatProofDAGLimits(limits)
	if err != nil {
		return nil, err
	}
	return &SatProofDAG{limits: checked}, nil
}

func (d *SatProofDAG) tick() error {
	if d.limits.Stop != nil && d.limits.Stop() {
		return ErrCancelled
	}
	return nil
}

func validSatProofDAGRef(reference SatProofDAGRef, nodes, maxClause int) bool {
	node := reference.Node >= 0 && reference.Clause == -1 && reference.Node < nodes
	clause := reference.Clause >= 0 && reference.Node == -1 && reference.Clause <= maxClause
	return node != clause
}

func (d *SatProofDAG) AddNode(node SatProofDAGNode) (int, error) {
	if d == nil || d.stopped {
		return -1, ErrInvariant
	}
	if len(node.Antecedents) == 0 {
		return -1, ErrDimension
	}
	if len(d.nodes) == d.limits.MaxNodes || len(node.Antecedents) > d.limits.MaxAntecedents-d.antecedents {
		return -1, ErrLimit
	}
	for _, reference := range node.Antecedents {
		if !validSatProofDAGRef(reference, len(d.nodes), d.limits.MaxClauseID) {
			if reference.Clause > d.limits.MaxClauseID {
				return -1, ErrLimit
			}
			return -1, ErrInvariant
		}
	}
	node.Antecedents = slices.Clone(node.Antecedents)
	handle := len(d.nodes)
	d.nodes = append(d.nodes, node)
	d.antecedents += len(node.Antecedents)
	return handle, nil
}

func (d *SatProofDAG) Node(handle int) (SatProofDAGNode, error) {
	if d == nil || d.stopped || handle < 0 || handle >= len(d.nodes) {
		return SatProofDAGNode{}, ErrInvalidLiteral
	}
	node := d.nodes[handle]
	node.Antecedents = slices.Clone(node.Antecedents)
	return node, nil
}

func (d *SatProofDAG) ResizeAntecedents(handle, count int) error {
	if d == nil || d.stopped || handle < 0 || handle >= len(d.nodes) {
		return ErrInvalidLiteral
	}
	if count <= 0 || count > len(d.nodes[handle].Antecedents) {
		return ErrDimension
	}
	removed := len(d.nodes[handle].Antecedents) - count
	d.nodes[handle].Antecedents = d.nodes[handle].Antecedents[:count:count]
	d.antecedents -= removed
	return nil
}

func (d *SatProofDAG) SetLabel(handle, label int) error {
	if d == nil || d.stopped || handle < 0 || handle >= len(d.nodes) {
		return ErrInvalidLiteral
	}
	d.nodes[handle].Label = label
	return nil
}

func (d *SatProofDAG) ClearLabels(handles []int) error {
	if d == nil || d.stopped {
		return ErrInvariant
	}
	for _, handle := range handles {
		if handle < 0 || handle >= len(d.nodes) {
			return ErrInvalidLiteral
		}
	}
	for _, handle := range handles {
		d.nodes[handle].Label = 0
	}
	return nil
}

type satProofDAGFrame struct {
	handle, next int
}

func (d *SatProofDAG) collectUsed(roots []int, sorted bool) ([]int, error) {
	if d == nil || d.stopped {
		return nil, ErrInvariant
	}
	visited := make([]bool, len(d.nodes))
	used := make([]int, 0)
	for _, root := range roots {
		if root < 0 {
			continue
		}
		if root >= len(d.nodes) {
			return nil, ErrInvalidLiteral
		}
		if visited[root] {
			continue
		}
		visited[root] = true
		stack := []satProofDAGFrame{{handle: root}}
		for len(stack) > 0 {
			if err := d.tick(); err != nil {
				return nil, err
			}
			frame := &stack[len(stack)-1]
			node := &d.nodes[frame.handle]
			if frame.next < len(node.Antecedents) {
				reference := node.Antecedents[frame.next]
				frame.next++
				if reference.Node >= 0 && !visited[reference.Node] {
					visited[reference.Node] = true
					stack = append(stack, satProofDAGFrame{handle: reference.Node})
				}
				continue
			}
			used = append(used, frame.handle)
			stack = stack[:len(stack)-1]
		}
	}
	if sorted {
		slices.Sort(used)
	}
	return used, nil
}

func (d *SatProofDAG) CollectUsedSorted(roots []int) ([]int, error) {
	return d.collectUsed(roots, true)
}

func (d *SatProofDAG) CollectUsedPostorder(roots []int) ([]int, error) {
	return d.collectUsed(roots, false)
}

func (d *SatProofDAG) MarkUsedCount(roots []int) (int, error) {
	used, err := d.collectUsed(roots, false)
	return len(used), err
}

func (d *SatProofDAG) CoreFromUsed(used []int) ([]int, error) {
	if d == nil || d.stopped {
		return nil, ErrInvariant
	}
	seenNodes := make([]bool, len(d.nodes))
	maximumClause := -1
	for _, handle := range used {
		if err := d.tick(); err != nil {
			return nil, err
		}
		if handle < 0 || handle >= len(d.nodes) || seenNodes[handle] {
			return nil, ErrInvariant
		}
		seenNodes[handle] = true
		for _, reference := range d.nodes[handle].Antecedents {
			maximumClause = max(maximumClause, reference.Clause)
		}
	}
	seenClauses := make([]bool, maximumClause+1)
	core := make([]int, 0)
	for _, handle := range used {
		for _, reference := range d.nodes[handle].Antecedents {
			if reference.Clause >= 0 && !seenClauses[reference.Clause] {
				seenClauses[reference.Clause] = true
				core = append(core, reference.Clause)
			}
		}
	}
	return core, nil
}

func (d *SatProofDAG) DeriveCore(root int) ([]int, error) {
	if root < 0 {
		return nil, ErrInvalidLiteral
	}
	used, err := d.CollectUsedSorted([]int{root})
	if err != nil {
		return nil, err
	}
	core, err := d.CoreFromUsed(used)
	if err != nil {
		return nil, err
	}
	slices.Sort(core)
	return core, nil
}

func (d *SatProofDAG) Reduce(roots []int, pivot int) (SatProofDAGReduction, error) {
	if d == nil || d.stopped || pivot < 0 || pivot >= len(d.nodes) {
		return SatProofDAGReduction{}, ErrInvalidLiteral
	}
	used, err := d.CollectUsedSorted(roots)
	if err != nil {
		return SatProofDAGReduction{}, err
	}
	oldToNew := make([]int, len(d.nodes))
	for i := range oldToNew {
		oldToNew[i] = -1
	}
	for newHandle, oldHandle := range used {
		oldToNew[oldHandle] = newHandle
	}
	if oldToNew[pivot] < 0 {
		return SatProofDAGReduction{}, ErrInvariant
	}
	newNodes := make([]SatProofDAGNode, len(used))
	newAntecedents := 0
	for newHandle, oldHandle := range used {
		node := d.nodes[oldHandle]
		node.Label = 0
		node.Antecedents = slices.Clone(node.Antecedents)
		for i := range node.Antecedents {
			if old := node.Antecedents[i].Node; old >= 0 {
				if oldToNew[old] < 0 {
					return SatProofDAGReduction{}, ErrInvariant
				}
				node.Antecedents[i].Node = oldToNew[old]
			}
		}
		newNodes[newHandle] = node
		newAntecedents += len(node.Antecedents)
	}
	newRoots := slices.Clone(roots)
	for i, root := range newRoots {
		if root < 0 {
			continue
		}
		if root >= len(oldToNew) || oldToNew[root] < 0 {
			return SatProofDAGReduction{}, ErrInvariant
		}
		newRoots[i] = oldToNew[root]
	}
	d.nodes, d.antecedents = newNodes, newAntecedents
	return SatProofDAGReduction{Roots: newRoots, OldToNew: oldToNew, Pivot: oldToNew[pivot]}, nil
}

func (d *SatProofDAG) Verify(roots []int, clauses [][]SatLit) ([][]SatLit, error) {
	used, err := d.CollectUsedSorted(roots)
	if err != nil {
		return nil, err
	}
	resolved := make([][]int, len(d.nodes))
	for _, handle := range used {
		node := d.nodes[handle]
		var current []int
		for index, reference := range node.Antecedents {
			var clause []int
			if reference.Node >= 0 {
				clause = resolved[reference.Node]
				if clause == nil {
					return nil, ErrInvariant
				}
			} else {
				if reference.Clause < 0 || reference.Clause >= len(clauses) {
					return nil, ErrInvalidLiteral
				}
				clause = make([]int, len(clauses[reference.Clause]))
				for i, literal := range clauses[reference.Clause] {
					if !literal.Valid() {
						return nil, ErrInvalidLiteral
					}
					clause[i] = literal.DIMACS()
				}
			}
			if index == 0 {
				current = slices.Clone(clause)
				continue
			}
			current, err = ResolveSatProofClauses(current, clause)
			if err != nil {
				return nil, err
			}
		}
		resolved[handle] = current
	}
	results := make([][]SatLit, len(roots))
	for i, root := range roots {
		if root < 0 || root >= len(resolved) || resolved[root] == nil {
			return nil, ErrInvalidLiteral
		}
		results[i] = make([]SatLit, len(resolved[root]))
		for j, value := range resolved[root] {
			literal, parseErr := ParseSatDIMACSLiteral(value)
			if parseErr != nil {
				return nil, parseErr
			}
			results[i][j] = literal
		}
	}
	return results, nil
}

func (d *SatProofDAG) Check() error {
	if d == nil || d.stopped || len(d.nodes) > d.limits.MaxNodes || d.antecedents > d.limits.MaxAntecedents {
		return ErrInvariant
	}
	total := 0
	for handle, node := range d.nodes {
		if len(node.Antecedents) == 0 {
			return ErrInvariant
		}
		total += len(node.Antecedents)
		for _, reference := range node.Antecedents {
			if !validSatProofDAGRef(reference, handle, d.limits.MaxClauseID) {
				return ErrInvariant
			}
		}
	}
	if total != d.antecedents {
		return ErrInvariant
	}
	return nil
}

func (d *SatProofDAG) Stats() SatProofDAGStats {
	if d == nil {
		return SatProofDAGStats{Stopped: true}
	}
	return SatProofDAGStats{len(d.nodes), d.antecedents, d.stopped}
}

func (d *SatProofDAG) Stop() {
	if d == nil || d.stopped {
		return
	}
	d.nodes, d.antecedents, d.stopped = nil, 0, true
}
