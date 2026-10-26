package abc

// Activity decision heap from src/sat/msat/msatOrderH.c.

import (
	"math"
	"slices"
)

type MsatVariableOrder struct {
	activities []float64
	assigned   []bool
	index      []int
	heap       []SatVar
	maximum    int
	closed     bool
}

func NewMsatVariableOrder(variables, maximum int) (*MsatVariableOrder, error) {
	maximum, err := msatVectorMaximum(maximum)
	if err != nil || variables < 0 {
		return nil, ErrDimension
	}
	if variables > maximum {
		return nil, ErrLimit
	}
	return &MsatVariableOrder{
		activities: make([]float64, variables),
		assigned:   make([]bool, variables),
		index:      make([]int, variables),
		maximum:    maximum,
	}, nil
}

func (o *MsatVariableOrder) checkLive() error {
	if o == nil {
		return ErrDimension
	}
	if o.closed || o.maximum <= 0 || len(o.activities) != len(o.assigned) || len(o.index) != len(o.assigned) || len(o.assigned) > o.maximum {
		return ErrInvariant
	}
	return nil
}

func (o *MsatVariableOrder) checkVariable(variable SatVar) error {
	if err := o.checkLive(); err != nil {
		return err
	}
	if uint64(variable) >= uint64(len(o.index)) {
		return ErrDimension
	}
	return nil
}

func (o *MsatVariableOrder) SetBounds(variables int) error {
	if err := o.checkLive(); err != nil {
		return err
	}
	if variables < 0 {
		return ErrDimension
	}
	if variables > o.maximum {
		return ErrLimit
	}
	activities := make([]float64, variables)
	assigned := make([]bool, variables)
	copy(activities, o.activities)
	copy(assigned, o.assigned)
	o.activities, o.assigned = activities, assigned
	o.index = make([]int, variables)
	o.heap = nil
	return nil
}

// Clean installs the source heap layout verbatim. It intentionally does not
// heapify: native preparation uses a cone whose activities have just been reset.
func (o *MsatVariableOrder) Clean(cone []SatVar) error {
	if err := o.checkLive(); err != nil {
		return err
	}
	seen := make([]bool, len(o.index))
	for _, variable := range cone {
		if uint64(variable) >= uint64(len(seen)) || seen[variable] {
			return ErrDimension
		}
		seen[variable] = true
	}
	clear(o.index)
	o.heap = make([]SatVar, len(cone)+1)
	for position, variable := range cone {
		o.heap[position+1] = variable
		o.index[variable] = position + 1
	}
	return nil
}

func (o *MsatVariableOrder) inHeap(variable SatVar) bool {
	return uint64(variable) < uint64(len(o.index)) && o.index[variable] != 0
}

func (o *MsatVariableOrder) better(first, second SatVar) bool {
	return o.activities[first] > o.activities[second]
}

func (o *MsatVariableOrder) percolateUp(position int) {
	variable := o.heap[position]
	for position/2 != 0 && o.better(variable, o.heap[position/2]) {
		o.heap[position] = o.heap[position/2]
		o.index[o.heap[position]] = position
		position /= 2
	}
	o.heap[position] = variable
	o.index[variable] = position
}

func (o *MsatVariableOrder) percolateDown(position int) {
	variable := o.heap[position]
	for 2*position < len(o.heap) {
		child := 2 * position
		if child+1 < len(o.heap) && o.better(o.heap[child+1], o.heap[child]) {
			child++
		}
		if !o.better(o.heap[child], variable) {
			break
		}
		o.heap[position] = o.heap[child]
		o.index[o.heap[position]] = position
		position = child
	}
	o.heap[position] = variable
	o.index[variable] = position
}

func (o *MsatVariableOrder) insert(variable SatVar) {
	o.heap = append(o.heap, variable)
	o.index[variable] = len(o.heap) - 1
	o.percolateUp(len(o.heap) - 1)
}

func (o *MsatVariableOrder) popTop() SatVar {
	result := o.heap[1]
	newTop := o.heap[len(o.heap)-1]
	o.heap = o.heap[:len(o.heap)-1]
	o.index[result] = 0
	if len(o.heap) > 1 {
		o.heap[1] = newTop
		o.index[newTop] = 1
		o.percolateDown(1)
	}
	return result
}

// Select removes heap entries until it finds an unassigned variable. The
// selected variable is removed too, matching Msat_HeapGetTop.
func (o *MsatVariableOrder) Select() (SatVar, bool, error) {
	if err := o.checkLive(); err != nil {
		return 0, false, err
	}
	for len(o.heap) > 1 {
		variable := o.popTop()
		if !o.assigned[variable] {
			return variable, true, nil
		}
	}
	return 0, false, nil
}

func (o *MsatVariableOrder) Assign(variable SatVar) error {
	if err := o.checkVariable(variable); err != nil {
		return err
	}
	o.assigned[variable] = true
	return nil
}

func (o *MsatVariableOrder) Unassign(variable SatVar) error {
	if err := o.checkVariable(variable); err != nil {
		return err
	}
	o.assigned[variable] = false
	if !o.inHeap(variable) {
		o.insert(variable)
	}
	return nil
}

// IncreaseActivity performs the source's heap update, which only percolates
// upward because activity bumps never decrease one variable relative to peers.
func (o *MsatVariableOrder) IncreaseActivity(variable SatVar, amount float64) error {
	if err := o.checkVariable(variable); err != nil {
		return err
	}
	if amount < 0 || math.IsNaN(amount) || math.IsInf(amount, 0) {
		return ErrDimension
	}
	if math.IsInf(o.activities[variable]+amount, 0) {
		return ErrLimit
	}
	o.activities[variable] += amount
	if o.inHeap(variable) {
		o.percolateUp(o.index[variable])
	}
	return nil
}

// RescaleActivities uniformly scales all activities; positive uniform scaling
// preserves heap order and therefore needs no rebuild.
func (o *MsatVariableOrder) RescaleActivities(factor float64) error {
	if err := o.checkLive(); err != nil {
		return err
	}
	if factor < 0 || math.IsNaN(factor) || math.IsInf(factor, 0) {
		return ErrDimension
	}
	for _, activity := range o.activities {
		if math.IsInf(activity*factor, 0) {
			return ErrLimit
		}
	}
	for index := range o.activities {
		o.activities[index] *= factor
	}
	return nil
}

func (o *MsatVariableOrder) Activities() []float64 {
	if o == nil || o.closed {
		return nil
	}
	return slices.Clone(o.activities)
}

func (o *MsatVariableOrder) Heap() []SatVar {
	if o == nil || o.closed {
		return nil
	}
	if len(o.heap) <= 1 {
		return []SatVar{}
	}
	return slices.Clone(o.heap[1:])
}

func (o *MsatVariableOrder) Check() error {
	if err := o.checkLive(); err != nil {
		return err
	}
	if len(o.heap) == 0 {
		return nil
	}
	seen := make([]bool, len(o.index))
	for position := 1; position < len(o.heap); position++ {
		variable := o.heap[position]
		if uint64(variable) >= uint64(len(o.index)) || seen[variable] || o.index[variable] != position {
			return ErrInvariant
		}
		seen[variable] = true
		if position > 1 && o.better(variable, o.heap[position/2]) {
			return ErrInvariant
		}
	}
	for variable, position := range o.index {
		if (position != 0) != seen[variable] {
			return ErrInvariant
		}
	}
	return nil
}

func (o *MsatVariableOrder) Close() {
	if o == nil || o.closed {
		return
	}
	clear(o.activities)
	clear(o.assigned)
	clear(o.index)
	clear(o.heap)
	o.activities, o.assigned, o.index, o.heap = nil, nil, nil, nil
	o.closed = true
}
