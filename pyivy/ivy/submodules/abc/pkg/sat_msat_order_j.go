package abc

// J-boundary variable order from src/sat/msat/msatOrderJ.c.

import (
	"math"
	"slices"
)

// MsatJBoundaryOrder is the build-disabled MSAT ordering alternative. It keeps
// a circular insertion-ordered boundary of unassigned cone variables having at
// least one assigned neighbor.
type MsatJBoundaryOrder struct {
	activities []float64
	assigned   []bool
	inCone     []bool
	adjacent   [][]SatVar
	next       []int
	previous   []int
	root       int
	items      int
	maximum    int
	closed     bool
}

func NewMsatJBoundaryOrder(variables, maximum int) (*MsatJBoundaryOrder, error) {
	maximum, err := msatVectorMaximum(maximum)
	if err != nil || variables < 0 {
		return nil, ErrDimension
	}
	if variables > maximum {
		return nil, ErrLimit
	}
	o := &MsatJBoundaryOrder{maximum: maximum, root: -1}
	if err = o.resize(variables); err != nil {
		return nil, err
	}
	return o, nil
}

func (o *MsatJBoundaryOrder) checkLive() error {
	if o == nil {
		return ErrDimension
	}
	n := len(o.activities)
	if o.closed || o.maximum <= 0 || n > o.maximum || len(o.assigned) != n || len(o.inCone) != n || len(o.adjacent) != n || len(o.next) != n || len(o.previous) != n {
		return ErrInvariant
	}
	if o.items < 0 || o.items > n || (o.items == 0) != (o.root == -1) {
		return ErrInvariant
	}
	return nil
}

func (o *MsatJBoundaryOrder) checkVariable(variable SatVar) error {
	if err := o.checkLive(); err != nil {
		return err
	}
	if uint64(variable) >= uint64(len(o.activities)) {
		return ErrDimension
	}
	return nil
}

func (o *MsatJBoundaryOrder) hasAssignments() bool {
	return slices.Contains(o.assigned, true)
}

func (o *MsatJBoundaryOrder) resize(variables int) error {
	if variables < 0 {
		return ErrDimension
	}
	if variables > o.maximum {
		return ErrLimit
	}
	activities := make([]float64, variables)
	assigned := make([]bool, variables)
	inCone := make([]bool, variables)
	adjacent := make([][]SatVar, variables)
	copy(activities, o.activities)
	copy(assigned, o.assigned)
	copy(inCone, o.inCone)
	for index := 0; index < min(variables, len(o.adjacent)); index++ {
		for _, other := range o.adjacent[index] {
			if uint64(other) < uint64(variables) {
				adjacent[index] = append(adjacent[index], other)
			}
		}
	}
	o.activities, o.assigned, o.inCone, o.adjacent = activities, assigned, inCone, adjacent
	o.next, o.previous = make([]int, variables), make([]int, variables)
	for index := range variables {
		o.next[index], o.previous[index] = -1, -1
	}
	o.root, o.items = -1, 0
	return nil
}

// SetBounds resizes all variable-indexed state and clears the boundary. State
// for surviving variables is retained; adjacency endpoints outside a shrink
// are discarded.
func (o *MsatJBoundaryOrder) SetBounds(variables int) error {
	if err := o.checkLive(); err != nil {
		return err
	}
	if variables < 0 {
		return ErrDimension
	}
	if variables > o.maximum {
		return ErrLimit
	}
	if o.items != 0 || o.hasAssignments() {
		return ErrInvariant
	}
	return o.resize(variables)
}

// SetCone replaces the active cone and clears the boundary, matching native
// preparation followed by Msat_OrderClean.
func (o *MsatJBoundaryOrder) SetCone(cone []SatVar) error {
	if err := o.checkLive(); err != nil {
		return err
	}
	if o.items != 0 || o.hasAssignments() {
		return ErrInvariant
	}
	seen := make([]bool, len(o.inCone))
	for _, variable := range cone {
		if uint64(variable) >= uint64(len(seen)) || seen[variable] {
			return ErrDimension
		}
		seen[variable] = true
	}
	o.inCone = seen
	o.Clean()
	return nil
}

// SetAdjacency installs one owned adjacency list. Native callers supply a
// symmetric graph; this method permits directed lists so malformed topology is
// observable rather than silently rewritten.
func (o *MsatJBoundaryOrder) SetAdjacency(variable SatVar, adjacent []SatVar) error {
	if err := o.checkVariable(variable); err != nil {
		return err
	}
	if o.items != 0 || o.hasAssignments() {
		return ErrInvariant
	}
	seen := make([]bool, len(o.adjacent))
	for _, other := range adjacent {
		if uint64(other) >= uint64(len(seen)) || seen[other] {
			return ErrDimension
		}
		seen[other] = true
	}
	o.adjacent[variable] = slices.Clone(adjacent)
	return nil
}

// SetActivity writes the shared-solver activity observed by native selection.
func (o *MsatJBoundaryOrder) SetActivity(variable SatVar, activity float64) error {
	if err := o.checkVariable(variable); err != nil {
		return err
	}
	if activity < 0 || math.IsNaN(activity) || math.IsInf(activity, 0) {
		return ErrDimension
	}
	o.activities[variable] = activity
	return nil
}

func (o *MsatJBoundaryOrder) inBoundary(variable int) bool {
	return o.next[variable] != -1
}

func (o *MsatJBoundaryOrder) addLast(variable int) {
	if o.inBoundary(variable) {
		return
	}
	o.items++
	if o.root == -1 {
		o.root = variable
		o.next[variable], o.previous[variable] = variable, variable
		return
	}
	last := o.previous[o.root]
	o.previous[variable], o.next[variable] = last, o.root
	o.next[last], o.previous[o.root] = variable, variable
}

func (o *MsatJBoundaryOrder) remove(variable int) {
	if !o.inBoundary(variable) {
		return
	}
	o.items--
	if o.items == 0 {
		o.root = -1
	} else {
		if o.root == variable {
			o.root = o.next[variable]
		}
		o.next[o.previous[variable]] = o.next[variable]
		o.previous[o.next[variable]] = o.previous[variable]
	}
	o.next[variable], o.previous[variable] = -1, -1
}

// Clean empties the J-boundary without changing cone, adjacency, assignment or
// activity state.
func (o *MsatJBoundaryOrder) Clean() {
	if o == nil || o.closed {
		return
	}
	for index := range o.next {
		o.next[index], o.previous[index] = -1, -1
	}
	o.root, o.items = -1, 0
}

// Select returns the highest-activity boundary variable without removing it.
// Strict comparison preserves insertion order among equal activities.
func (o *MsatJBoundaryOrder) Select() (SatVar, bool, error) {
	if err := o.Check(); err != nil {
		return 0, false, err
	}
	if o.root == -1 {
		return 0, false, nil
	}
	best := o.root
	for variable := o.next[o.root]; variable != o.root; variable = o.next[variable] {
		if o.activities[variable] > o.activities[best] {
			best = variable
		}
	}
	return SatVar(best), true, nil
}

// Assign records the assignment, removes the variable if it was on the
// boundary, and appends eligible neighbors in source adjacency order.
func (o *MsatJBoundaryOrder) Assign(variable SatVar) error {
	if err := o.checkVariable(variable); err != nil {
		return err
	}
	if o.assigned[variable] {
		return ErrInvariant
	}
	o.assigned[variable] = true
	o.remove(int(variable))
	if !o.inCone[variable] {
		return nil
	}
	for _, other := range o.adjacent[variable] {
		if o.inCone[other] && !o.assigned[other] && !o.inBoundary(int(other)) {
			o.addLast(int(other))
		}
	}
	return nil
}

// Unassign records the unassignment, adds the variable when it still has an
// assigned cone neighbor, and removes adjacent boundary entries that no longer
// have any assigned cone neighbor.
func (o *MsatJBoundaryOrder) Unassign(variable SatVar) error {
	if err := o.checkVariable(variable); err != nil {
		return err
	}
	if !o.assigned[variable] || o.inBoundary(int(variable)) {
		return ErrInvariant
	}
	o.assigned[variable] = false
	if o.inCone[variable] && o.hasAssignedConeNeighbor(int(variable)) {
		o.addLast(int(variable))
	}
	for _, other := range o.adjacent[variable] {
		if o.inBoundary(int(other)) && !o.hasAssignedConeNeighbor(int(other)) {
			o.remove(int(other))
		}
	}
	return nil
}

func (o *MsatJBoundaryOrder) hasAssignedConeNeighbor(variable int) bool {
	for _, other := range o.adjacent[variable] {
		if o.inCone[other] && o.assigned[other] {
			return true
		}
	}
	return false
}

func (o *MsatJBoundaryOrder) Boundary() []SatVar {
	if o == nil || o.closed || o.root == -1 {
		return []SatVar{}
	}
	result := make([]SatVar, 0, o.items)
	result = append(result, SatVar(o.root))
	for variable := o.next[o.root]; variable != o.root; variable = o.next[variable] {
		result = append(result, SatVar(variable))
	}
	return result
}

func (o *MsatJBoundaryOrder) Activities() []float64 {
	if o == nil || o.closed {
		return nil
	}
	return slices.Clone(o.activities)
}

func (o *MsatJBoundaryOrder) Assignments() []bool {
	if o == nil || o.closed {
		return nil
	}
	return slices.Clone(o.assigned)
}

// Check validates both ring structure and the semantic J-boundary definition.
func (o *MsatJBoundaryOrder) Check() error {
	if err := o.checkLive(); err != nil {
		return err
	}
	seen := make([]bool, len(o.next))
	count := 0
	if o.root != -1 {
		variable := o.root
		for {
			if variable < 0 || variable >= len(o.next) || seen[variable] || o.next[variable] < 0 || o.previous[variable] < 0 || o.next[o.previous[variable]] != variable || o.previous[o.next[variable]] != variable {
				return ErrInvariant
			}
			if o.assigned[variable] || !o.inCone[variable] || !o.hasAssignedConeNeighbor(variable) {
				return ErrInvariant
			}
			seen[variable] = true
			count++
			variable = o.next[variable]
			if variable == o.root {
				break
			}
		}
	}
	if count != o.items {
		return ErrInvariant
	}
	for variable := range o.next {
		if seen[variable] != o.inBoundary(variable) || o.inCone[variable] && !o.assigned[variable] && o.hasAssignedConeNeighbor(variable) != seen[variable] {
			return ErrInvariant
		}
	}
	return nil
}

func (o *MsatJBoundaryOrder) Close() {
	if o == nil || o.closed {
		return
	}
	o.activities, o.assigned, o.inCone, o.adjacent = nil, nil, nil, nil
	o.next, o.previous = nil, nil
	o.root, o.items, o.closed = -1, 0, true
}
