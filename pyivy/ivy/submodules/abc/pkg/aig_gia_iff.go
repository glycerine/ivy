package abc

// Hierarchical LUT packing from src/aig/gia/giaIff.c.

import (
	"math"
	"slices"
)

type GiaIffLUTLibrary struct {
	DelayByInputs []float32
}

type GiaIffOptions struct {
	LUTSize, Degree               int
	MaxObjects, MaxMappingEntries int
	MaxGroups, MaxVisits          int
	Stop                          func() bool
}

type GiaIffResult struct {
	Packing              *GiaLUTPacking
	Timing               *TimManager
	Arrivals             []float32
	Match2, Match3       []int
	MaximumDelay         float32
	Single, Pair, Triple int
}

type GiaIffSession struct {
	manager *GiaManager
	mapping *GiaLUTMapping
	library GiaIffLUTLibrary
	timing  *TimManager
	options GiaIffOptions
	result  GiaIffResult
	ran     bool
	closed  bool
}

func checkGiaIffOptions(options GiaIffOptions) (GiaIffOptions, error) {
	if options.LUTSize < 0 || options.Degree < 0 || options.MaxObjects < 0 || options.MaxMappingEntries < 0 || options.MaxGroups < 0 || options.MaxVisits < 0 {
		return options, ErrDimension
	}
	if options.Degree != 0 && options.Degree != 2 && options.Degree != 3 {
		return options, ErrDimension
	}
	if options.MaxObjects == 0 {
		options.MaxObjects = 1 << 20
	}
	if options.MaxMappingEntries == 0 {
		options.MaxMappingEntries = 4 << 20
	}
	if options.MaxGroups == 0 {
		options.MaxGroups = 1 << 20
	}
	if options.MaxVisits == 0 {
		options.MaxVisits = 4 << 20
	}
	return options, nil
}

func cloneGiaIffResult(result GiaIffResult) GiaIffResult {
	copy := result
	copy.Arrivals = slices.Clone(result.Arrivals)
	copy.Match2 = slices.Clone(result.Match2)
	copy.Match3 = slices.Clone(result.Match3)
	copy.Timing = result.Timing.Clone()
	if result.Packing != nil {
		copy.Packing = &GiaLUTPacking{mapping: result.Packing.mapping, packed: slices.Clone(result.Packing.packed)}
	}
	return copy
}

func NewGiaIffSession(manager *GiaManager, mapping *GiaLUTMapping, library GiaIffLUTLibrary, timing *TimManager, options GiaIffOptions) (*GiaIffSession, error) {
	if manager == nil || mapping == nil || timing == nil {
		return nil, ErrDimension
	}
	var err error
	options, err = checkGiaIffOptions(options)
	if err != nil {
		return nil, err
	}
	if err = mapping.checkGraph(manager); err != nil {
		return nil, err
	}
	if manager.ObjectCount() > options.MaxObjects || len(mapping.packed) > options.MaxMappingEntries {
		return nil, ErrLimit
	}
	cis, cos := timing.Counts()
	if cis != manager.InputCount() || cos != manager.OutputCount() || len(library.DelayByInputs) < 2 {
		return nil, ErrDimension
	}
	for _, delay := range library.DelayByInputs {
		if math.IsNaN(float64(delay)) || math.IsInf(float64(delay), 0) {
			return nil, ErrDimension
		}
	}
	statistics := mapping.Statistics()
	if statistics.LUTs == 0 || statistics.MaxSize < 1 {
		return nil, ErrDimension
	}
	if options.LUTSize == 0 {
		if statistics.MaxSize <= 4 {
			options.LUTSize = 4
		} else if statistics.MaxSize <= 6 {
			options.LUTSize = 6
		} else {
			return nil, ErrUnsupported
		}
	}
	if statistics.MaxSize > options.LUTSize {
		return nil, ErrDimension
	}
	maximumInputs := len(library.DelayByInputs) - 1
	if options.Degree == 0 {
		switch {
		case options.LUTSize == 4 && maximumInputs == 7, options.LUTSize == 6 && maximumInputs == 11:
			options.Degree = 2
		case options.LUTSize == 4 && maximumInputs == 10, options.LUTSize == 6 && maximumInputs == 16:
			options.Degree = 3
		default:
			return nil, ErrDimension
		}
	}
	clone := manager.Clone()
	ownedMapping, err := GiaLUTMappingFromPacked(mapping.objects, mapping.packed, options.MaxMappingEntries)
	if err != nil {
		return nil, err
	}
	return &GiaIffSession{
		manager: clone, mapping: ownedMapping,
		library: GiaIffLUTLibrary{DelayByInputs: slices.Clone(library.DelayByInputs)},
		timing:  timing.Clone(), options: options,
	}, nil
}

func (session *GiaIffSession) usable() bool {
	return session != nil && !session.closed && session.manager != nil && session.mapping != nil && session.timing != nil
}

// Boundary returns the source-order unique input set after optionally absorbing
// one or two direct LUT fanins into root.
func (session *GiaIffSession) Boundary(root GiaID, skip2, skip3 int) ([]GiaID, error) {
	if !session.usable() || uint64(root) >= uint64(session.manager.ObjectCount()) || skip2 < -1 || skip3 < -1 || skip2 == skip3 && skip2 >= 0 {
		return nil, ErrDimension
	}
	record, exists := session.mapping.Record(root)
	if !exists {
		return nil, ErrDimension
	}
	if skip2 >= 0 && !slices.Contains(record.Fanins, GiaID(skip2)) || skip3 >= 0 && !slices.Contains(record.Fanins, GiaID(skip3)) {
		return nil, ErrDimension
	}
	seen := make([]bool, session.manager.ObjectCount())
	boundary := make([]GiaID, 0)
	add := func(id GiaID) error {
		if uint64(id) >= uint64(len(seen)) {
			return ErrDimension
		}
		if !seen[id] {
			seen[id] = true
			boundary = append(boundary, id)
		}
		return nil
	}
	for _, fanin := range record.Fanins {
		if int(fanin) != skip2 && int(fanin) != skip3 {
			if err := add(fanin); err != nil {
				return nil, err
			}
		}
	}
	if skip2 >= 0 {
		child, ok := session.mapping.Record(GiaID(skip2))
		if !ok {
			return nil, ErrDimension
		}
		for _, fanin := range child.Fanins {
			if int(fanin) != skip3 {
				if err := add(fanin); err != nil {
					return nil, err
				}
			}
		}
	}
	if skip3 >= 0 {
		child, ok := session.mapping.Record(GiaID(skip3))
		if !ok {
			return nil, ErrDimension
		}
		for _, fanin := range child.Fanins {
			if int(fanin) != skip2 {
				if err := add(fanin); err != nil {
					return nil, err
				}
			}
		}
	}
	return boundary, nil
}

func (session *GiaIffSession) boundaryArrival(root GiaID, skip2, skip3 int, arrivals []float32) (float32, int, error) {
	boundary, err := session.Boundary(root, skip2, skip3)
	if err != nil || len(boundary) == 0 {
		return 0, 0, ErrDimension
	}
	if len(boundary) >= len(session.library.DelayByInputs) {
		return 0, 0, ErrLimit
	}
	maximum := float32(-math.MaxFloat32)
	for _, id := range boundary {
		if options := session.options; options.Stop != nil && options.Stop() {
			return 0, 0, ErrCancelled
		}
		arrival := arrivals[id]
		if math.IsInf(float64(arrival), -1) || arrival == float32(-math.MaxFloat32) {
			return 0, 0, ErrInvariant
		}
		if arrival > maximum {
			maximum = arrival
		}
	}
	return maximum, len(boundary), nil
}

func (session *GiaIffSession) bestPair(root GiaID, baseline float32, arrivals []float32) (float32, int, error) {
	record, _ := session.mapping.Record(root)
	best, selected := baseline, -1
	for _, fanin := range record.Fanins {
		if session.manager.objects[fanin].IsCI() || fanin == 0 {
			continue
		}
		arrival, size, err := session.boundaryArrival(root, int(fanin), -1, arrivals)
		if err != nil {
			return 0, -1, err
		}
		candidate := arrival + session.library.DelayByInputs[size]
		if best > candidate {
			best, selected = candidate, int(fanin)
		}
	}
	return best, selected, nil
}

func (session *GiaIffSession) bestTriple(root GiaID, baseline float32, arrivals []float32) (float32, int, int, error) {
	record, _ := session.mapping.Record(root)
	best, first, second := baseline, -1, -1
	for _, left := range record.Fanins {
		if left == 0 || session.manager.objects[left].IsCI() {
			continue
		}
		for _, right := range record.Fanins {
			if right == left || right == 0 || session.manager.objects[right].IsCI() {
				continue
			}
			arrival, size, err := session.boundaryArrival(root, int(left), int(right), arrivals)
			if err != nil {
				return 0, -1, -1, err
			}
			candidate := arrival + session.library.DelayByInputs[size]
			if best > candidate {
				best, first, second = candidate, int(left), int(right)
			}
		}
	}
	return best, first, second, nil
}

func (session *GiaIffSession) selectPacking(match2, match3 []int) (*GiaLUTPacking, error) {
	type frame struct {
		id   GiaID
		exit bool
	}
	visited := make([]bool, session.manager.ObjectCount())
	visited[0] = true
	for _, id := range session.manager.cis {
		visited[id] = true
	}
	groups := make([][]GiaID, 0)
	visits := 0
	for _, co := range session.manager.cos {
		driver, err := session.manager.objects[co].Fanin0(co)
		if err != nil {
			return nil, err
		}
		stack := []frame{{id: driver.ID()}}
		for len(stack) > 0 {
			if session.options.Stop != nil && session.options.Stop() {
				return nil, ErrCancelled
			}
			last := len(stack) - 1
			current := stack[last]
			stack = stack[:last]
			if current.exit {
				group := []GiaID{current.id}
				if match2[current.id] >= 0 {
					group = []GiaID{GiaID(match2[current.id]), current.id}
				}
				if match3[current.id] >= 0 {
					group = []GiaID{GiaID(match2[current.id]), GiaID(match3[current.id]), current.id}
				}
				if len(groups) == session.options.MaxGroups {
					return nil, ErrLimit
				}
				groups = append(groups, group)
				continue
			}
			if visited[current.id] {
				continue
			}
			visits++
			if visits > session.options.MaxVisits {
				return nil, ErrLimit
			}
			visited[current.id] = true
			record, ok := session.mapping.Record(current.id)
			if !ok {
				return nil, ErrDimension
			}
			dependencies := make([]GiaID, 0)
			for _, fanin := range record.Fanins {
				if int(fanin) != match2[current.id] && int(fanin) != match3[current.id] {
					dependencies = append(dependencies, fanin)
				}
			}
			if match2[current.id] >= 0 {
				child, ok := session.mapping.Record(GiaID(match2[current.id]))
				if !ok {
					return nil, ErrDimension
				}
				for _, fanin := range child.Fanins {
					if int(fanin) != match3[current.id] {
						dependencies = append(dependencies, fanin)
					}
				}
			}
			if match3[current.id] >= 0 {
				child, ok := session.mapping.Record(GiaID(match3[current.id]))
				if !ok {
					return nil, ErrDimension
				}
				for _, fanin := range child.Fanins {
					if int(fanin) != match2[current.id] {
						dependencies = append(dependencies, fanin)
					}
				}
			}
			stack = append(stack, frame{id: current.id, exit: true})
			for index := len(dependencies) - 1; index >= 0; index-- {
				stack = append(stack, frame{id: dependencies[index]})
			}
		}
	}
	return NewGiaLUTPacking(session.mapping, groups, session.options.MaxMappingEntries)
}

// Run computes arrivals, hierarchical matches and the selected packing once.
func (session *GiaIffSession) Run() (GiaIffResult, error) {
	if !session.usable() {
		return GiaIffResult{}, ErrInvariant
	}
	if session.ran {
		return cloneGiaIffResult(session.result), nil
	}
	arrivals := make([]float32, session.manager.ObjectCount())
	for index := range arrivals {
		arrivals[index] = float32(-math.MaxFloat32)
	}
	arrivals[0] = 0
	match2, match3 := make([]int, len(arrivals)), make([]int, len(arrivals))
	for index := range match2 {
		match2[index], match3[index] = -1, -1
	}
	timing := session.timing.Clone()
	if err := timing.IncrementTraversal(); err != nil {
		return GiaIffResult{}, err
	}
	result := GiaIffResult{MaximumDelay: float32(-math.MaxFloat32)}
	for id, object := range session.manager.objects {
		if session.options.Stop != nil && session.options.Stop() {
			return GiaIffResult{}, ErrCancelled
		}
		if id == 0 {
			continue
		}
		if object.IsCI() {
			ordinal, _ := object.TerminalIndex()
			arrival, err := timing.CIArrival(ordinal)
			if err != nil {
				return GiaIffResult{}, err
			}
			arrivals[id] = arrival
			continue
		}
		if object.IsCO() {
			driver, _ := object.Fanin0(GiaID(id))
			if arrivals[driver.ID()] == float32(-math.MaxFloat32) {
				return GiaIffResult{}, ErrInvariant
			}
			arrivals[id] = arrivals[driver.ID()]
			ordinal, _ := object.TerminalIndex()
			if err := timing.SetCOArrival(ordinal, arrivals[id]); err != nil {
				return GiaIffResult{}, err
			}
			if arrivals[id] > result.MaximumDelay {
				result.MaximumDelay = arrivals[id]
			}
			continue
		}
		if _, mapped := session.mapping.Record(GiaID(id)); !mapped {
			continue
		}
		arrival, size, err := session.boundaryArrival(GiaID(id), -1, -1, arrivals)
		if err != nil {
			return GiaIffResult{}, err
		}
		one := arrival + session.library.DelayByInputs[size]
		two, selected2, err := session.bestPair(GiaID(id), one, arrivals)
		if err != nil {
			return GiaIffResult{}, err
		}
		selected3 := -1
		best := two
		if session.options.Degree == 3 {
			var triple2 int
			best, triple2, selected3, err = session.bestTriple(GiaID(id), two, arrivals)
			if err != nil {
				return GiaIffResult{}, err
			}
			if selected3 >= 0 {
				selected2 = triple2
			}
		}
		arrivals[id] = best
		if selected3 >= 0 {
			match2[id], match3[id] = selected2, selected3
			result.Triple++
		} else if selected2 >= 0 {
			match2[id] = selected2
			result.Pair++
		} else {
			result.Single++
		}
	}
	packing, err := session.selectPacking(match2, match3)
	if err != nil {
		return GiaIffResult{}, err
	}
	result.Packing, result.Timing = packing, timing
	result.Arrivals, result.Match2, result.Match3 = arrivals, match2, match3
	session.result, session.ran = result, true
	return cloneGiaIffResult(result), nil
}

func (session *GiaIffSession) Close() {
	if session == nil || session.closed {
		return
	}
	session.manager, session.mapping, session.timing = nil, nil, nil
	session.library.DelayByInputs = nil
	session.result = GiaIffResult{}
	session.closed = true
}

func BuildGiaIffPacking(manager *GiaManager, mapping *GiaLUTMapping, library GiaIffLUTLibrary, timing *TimManager, options GiaIffOptions) (GiaIffResult, error) {
	session, err := NewGiaIffSession(manager, mapping, library, timing, options)
	if err != nil {
		return GiaIffResult{}, err
	}
	defer session.Close()
	return session.Run()
}
