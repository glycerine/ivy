package abc

// Majority-network path enumeration from src/misc/extra/extraUtilMaj.c.

import (
	"encoding/binary"
	"fmt"
	"math/bits"
	"slices"
	"strings"
)

type GemLimits struct {
	MaxObjects, MaxWork, MaxWords, MaxOutputBytes int
	Stop                                          func() bool
}

type GemObject struct {
	Variables, Nodes int
	History          uint8
	Groups           uint16
	Predecessor      int
}

type GemStatistics struct {
	Variables, Objects, Capacity, Work int
}

type GemSymmetryReport struct {
	Symmetric, Majority bool
	Polar               uint32
}

type GemStage struct {
	Variable                         int
	Before, AfterExpand, AfterReduce int
}

type GemEnumerationOptions struct {
	DumpTables, Verbose bool
	Limits              GemLimits
}

type GemEnumerationResult struct {
	Objects     []GemObject
	Tables      []*TruthTable
	Stages      []GemStage
	Diagnostics []string
}

type GemManager struct {
	variables int
	verbose   bool
	limits    GemLimits
	objects   []GemObject
	tables    []*TruthTable
	unique    map[string]int
	work      int
	closed    bool
}

func normalizeGemLimits(limits GemLimits) (GemLimits, error) {
	if limits.MaxObjects < 0 || limits.MaxWork < 0 || limits.MaxWords < 0 || limits.MaxOutputBytes < 0 {
		return limits, ErrDimension
	}
	if limits.MaxObjects == 0 {
		limits.MaxObjects = 1 << 20
	}
	if limits.MaxWork == 0 {
		limits.MaxWork = 1 << 28
	}
	if limits.MaxWords == 0 {
		limits.MaxWords = 1 << 20
	}
	if limits.MaxOutputBytes == 0 {
		limits.MaxOutputBytes = 16 << 20
	}
	return limits, nil
}

func gemStopped(limits GemLimits) bool { return limits.Stop != nil && limits.Stop() }

func gemTruthKey(table *TruthTable, variables int) (string, error) {
	stretched, err := table.Stretch(variables)
	if err != nil {
		return "", err
	}
	buffer := make([]byte, 8*len(stretched.words))
	for i, word := range stretched.words {
		binary.LittleEndian.PutUint64(buffer[8*i:], word)
	}
	return string(buffer), nil
}

func NewGemManager(variables int, verbose bool, limits GemLimits) (*GemManager, error) {
	limits, err := normalizeGemLimits(limits)
	if err != nil || variables < 1 || variables > 16 {
		if err != nil {
			return nil, err
		}
		return nil, ErrDimension
	}
	if limits.MaxObjects < 2 {
		return nil, ErrLimit
	}
	if gemStopped(limits) {
		return nil, ErrCancelled
	}
	zero, err := NewConstantTruthTable(0, false, limits.MaxWords)
	if err != nil {
		return nil, err
	}
	variable, err := NewVariableTruthTable(1, 0, limits.MaxWords)
	if err != nil {
		return nil, err
	}
	manager := &GemManager{
		variables: variables,
		verbose:   verbose,
		limits:    limits,
		objects: []GemObject{
			{Predecessor: -1},
			{Variables: 1, Groups: 1, Predecessor: -1},
		},
		tables: []*TruthTable{zero, variable},
		unique: make(map[string]int, 2),
	}
	for i, table := range manager.tables {
		key, keyErr := gemTruthKey(table, variables)
		if keyErr != nil {
			return nil, keyErr
		}
		manager.unique[key] = i
	}
	return manager, nil
}

func (manager *GemManager) check() error {
	if manager == nil {
		return ErrDimension
	}
	if manager.closed || manager.variables < 1 || manager.variables > 16 || len(manager.objects) != len(manager.tables) || len(manager.objects) < 2 || manager.unique == nil {
		return ErrInvariant
	}
	return nil
}

func (manager *GemManager) charge(amount int) error {
	if amount < 0 || manager.work > manager.limits.MaxWork-amount {
		return ErrLimit
	}
	if gemStopped(manager.limits) {
		return ErrCancelled
	}
	manager.work += amount
	return nil
}

func (manager *GemManager) Statistics() (GemStatistics, error) {
	if err := manager.check(); err != nil {
		return GemStatistics{}, err
	}
	return GemStatistics{manager.variables, len(manager.objects), cap(manager.objects), manager.work}, nil
}

func (manager *GemManager) Objects() ([]GemObject, error) {
	if err := manager.check(); err != nil {
		return nil, err
	}
	return slices.Clone(manager.objects), nil
}

func (manager *GemManager) TruthTable(index int) (*TruthTable, error) {
	if err := manager.check(); err != nil {
		return nil, err
	}
	if index < 0 || index >= len(manager.tables) {
		return nil, ErrDimension
	}
	return manager.tables[index].Clone(), nil
}

func (manager *GemManager) ensureObjectCapacity() error {
	if len(manager.objects) >= manager.limits.MaxObjects {
		return ErrLimit
	}
	if len(manager.objects) < cap(manager.objects) {
		return nil
	}
	capacity := 4
	if cap(manager.objects) > 0 {
		capacity = 2 * cap(manager.objects)
	}
	if capacity > manager.limits.MaxObjects {
		capacity = manager.limits.MaxObjects
	}
	objects := make([]GemObject, len(manager.objects), capacity)
	copy(objects, manager.objects)
	tables := make([]*TruthTable, len(manager.tables), capacity)
	copy(tables, manager.tables)
	manager.objects, manager.tables = objects, tables
	return nil
}

func (manager *GemManager) insert(table *TruthTable, object GemObject) (int, bool, error) {
	key, err := gemTruthKey(table, manager.variables)
	if err != nil {
		return 0, false, err
	}
	if index, found := manager.unique[key]; found {
		return index, false, nil
	}
	if err = manager.ensureObjectCapacity(); err != nil {
		return 0, false, err
	}
	index := len(manager.objects)
	manager.objects = append(manager.objects, object)
	manager.tables = append(manager.tables, table.Clone())
	manager.unique[key] = index
	return index, true, nil
}

func gemTruthFromValues(variables, maxWords int, values []bool) (*TruthTable, error) {
	if variables < 0 || variables > 16 || len(values) != 1<<uint(variables) {
		return nil, ErrDimension
	}
	table, err := NewConstantTruthTable(variables, false, maxWords)
	if err != nil {
		return nil, err
	}
	for bit := 0; bit < len(table.words)*64; bit++ {
		assignment := bit
		if variables < 6 {
			assignment &= (1 << uint(variables)) - 1
		}
		if values[assignment] {
			table.words[bit>>6] |= uint64(1) << uint(bit&63)
		}
	}
	return table, nil
}

func GemGroupsDerive(table *TruthTable) (uint16, error) {
	if err := table.Check(); err != nil || table.variables < 1 || table.variables > 16 {
		if err != nil {
			return 0, err
		}
		return 0, ErrDimension
	}
	groups := uint16(1) << uint(table.variables-1)
	for variable := 0; variable < table.variables-1; variable++ {
		symmetry, err := table.InputSymmetry(variable, variable+1, false)
		if err != nil {
			return 0, err
		}
		if symmetry&TruthSymmetryDirect == 0 {
			groups |= uint16(1) << uint(variable)
		}
	}
	return groups, nil
}

func GemGroupVarRemove(groups uint16, index int) (uint16, error) {
	if index < 0 || index >= 16 || groups&(uint16(1)<<uint(index)) == 0 {
		return 0, ErrDimension
	}
	mask := uint16(0)
	if index > 0 {
		mask = uint16(1)<<uint(index) - 1
	}
	return groups&mask | (groups&^mask)>>1, nil
}

func GemGroupVarsInsert1(groups uint16, after int, grouped bool) (uint16, error) {
	if after < -1 || after >= 15 || grouped && after >= 0 && groups&(uint16(1)<<uint(after)) == 0 {
		return 0, ErrDimension
	}
	position := after + 1
	mask := uint16(0)
	if position > 0 {
		mask = uint16(1)<<uint(position) - 1
	}
	result := uint32(groups&mask) | uint32(groups&^mask)<<1
	if grouped {
		result |= uint32(1) << uint(position)
	}
	if result > 0xffff {
		return 0, ErrLimit
	}
	return uint16(result), nil
}

func GemGroupVarsInsert3(groups uint16, after int) (uint16, error) {
	if after < -1 || after >= 13 || after >= 0 && groups&(uint16(1)<<uint(after)) == 0 {
		return 0, ErrDimension
	}
	position := after + 1
	mask := uint16(0)
	if position > 0 {
		mask = uint16(1)<<uint(position) - 1
	}
	result := uint32(groups&mask) | uint32(groups&^mask)<<3 | uint32(4)<<uint(position)
	if result > 0xffff {
		return 0, ErrLimit
	}
	return uint16(result), nil
}

func GemGroupUnpack(groups uint16) []int {
	result := make([]int, 0, bits.OnesCount16(groups))
	for variable := 0; variable < 16; variable++ {
		if groups&(uint16(1)<<uint(variable)) != 0 {
			result = append(result, variable)
		}
	}
	return result
}

func gemMemoryCompare(first, second []uint64) int {
	for i := range first {
		for shift := uint(0); shift < 64; shift += 8 {
			a, b := byte(first[i]>>shift), byte(second[i]>>shift)
			if a < b {
				return -1
			}
			if a > b {
				return 1
			}
		}
	}
	return 0
}

func GemFunctionFindPlace(table *TruthTable, groups uint16, oneVariable bool) (*TruthTable, int, error) {
	if err := table.Check(); err != nil {
		return nil, 0, err
	}
	last := GemGroupUnpack(groups)
	for _, variable := range last {
		if variable >= table.variables {
			return nil, 0, ErrDimension
		}
	}
	best := table.Clone()
	candidate := table.Clone()
	bestPlace := -1
	if len(last) > 0 {
		bestPlace = last[len(last)-1]
	}
	for group := len(last) - 1; group >= 0; group-- {
		limit := -1
		if group > 0 {
			limit = last[group-1]
		}
		for variable := last[group]; variable > limit; variable-- {
			width := 1
			if !oneVariable {
				width = 3
			}
			if variable+width >= table.variables {
				return nil, 0, ErrDimension
			}
			for offset := 0; offset < width; offset++ {
				candidate.swapVariables(variable+offset, variable+offset+1)
			}
		}
		if gemMemoryCompare(best.words, candidate.words) < 0 {
			best = candidate.Clone()
			bestPlace = limit
		}
	}
	return best, bestPlace, nil
}

func (manager *GemManager) canonicalize(table *TruthTable) (*TruthTable, error) {
	remaining := manager.limits.MaxWork - manager.work
	if remaining <= 0 {
		return nil, ErrLimit
	}
	canonical, err := table.CanonicalizeNPN(TruthCanonicalOptions{OnlyPermutation: true, MaxRefinements: remaining, Stop: manager.limits.Stop})
	if err != nil {
		return nil, err
	}
	if err = manager.charge(canonical.Attempts); err != nil {
		return nil, err
	}
	return canonical.Table, nil
}

func (manager *GemManager) Expand(index, variable int) (int, bool, error) {
	if err := manager.check(); err != nil {
		return 0, false, err
	}
	if index < 0 || index >= len(manager.objects) || variable < 0 || variable >= manager.objects[index].Variables || manager.objects[index].Variables+2 > manager.variables {
		return 0, false, ErrDimension
	}
	source := manager.tables[index]
	oldVariables := source.variables
	newVariables := oldVariables + 2
	assignments := 1 << uint(newVariables)
	if err := manager.charge(assignments); err != nil {
		return 0, false, err
	}
	values := make([]bool, assignments)
	for assignment := 0; assignment < assignments; assignment++ {
		if assignment&255 == 0 && gemStopped(manager.limits) {
			return 0, false, ErrCancelled
		}
		last := assignment >> (newVariables - 3) & 7
		replacement := bits.OnesCount(uint(last)) >= 2
		oldAssignment := 0
		for old := 0; old < oldVariables; old++ {
			value := replacement
			if old != variable {
				position := old
				if old > variable {
					position--
				}
				value = assignment&(1<<uint(position)) != 0
			}
			if value {
				oldAssignment |= 1 << uint(old)
			}
		}
		values[assignment], _ = source.Get(uint64(oldAssignment))
	}
	table, err := gemTruthFromValues(newVariables, manager.limits.MaxWords, values)
	if err != nil {
		return 0, false, err
	}
	table, err = manager.canonicalize(table)
	if err != nil {
		return 0, false, err
	}
	groups, err := GemGroupsDerive(table)
	if err != nil {
		return 0, false, err
	}
	object := GemObject{Variables: newVariables, Nodes: manager.objects[index].Nodes + 1, History: uint8(variable), Groups: groups, Predecessor: index}
	return manager.insert(table, object)
}

func (manager *GemManager) Reduce(index, first, second int) (int, bool, error) {
	if err := manager.check(); err != nil {
		return 0, false, err
	}
	if index < 0 || index >= len(manager.objects) || first < 0 || first >= second || second >= manager.objects[index].Variables || second >= 16 {
		return 0, false, ErrDimension
	}
	source := manager.tables[index]
	oldVariables := source.variables
	newVariables := oldVariables - 1
	assignments := 1 << uint(newVariables)
	if err := manager.charge(assignments); err != nil {
		return 0, false, err
	}
	values := make([]bool, assignments)
	for assignment := 0; assignment < assignments; assignment++ {
		if assignment&255 == 0 && gemStopped(manager.limits) {
			return 0, false, ErrCancelled
		}
		replacement := assignment&(1<<uint(newVariables-1)) != 0
		oldAssignment, position := 0, 0
		for old := 0; old < oldVariables; old++ {
			value := replacement
			if old != first && old != second {
				value = assignment&(1<<uint(position)) != 0
				position++
			}
			if value {
				oldAssignment |= 1 << uint(old)
			}
		}
		values[assignment], _ = source.Get(uint64(oldAssignment))
	}
	table, err := gemTruthFromValues(newVariables, manager.limits.MaxWords, values)
	if err != nil {
		return 0, false, err
	}
	table, err = manager.canonicalize(table)
	if err != nil {
		return 0, false, err
	}
	groups, err := GemGroupsDerive(table)
	if err != nil {
		return 0, false, err
	}
	object := GemObject{Variables: newVariables, Nodes: manager.objects[index].Nodes, History: uint8(second<<4 | first), Groups: groups, Predecessor: index}
	return manager.insert(table, object)
}

func (manager *GemManager) CheckMajority(index int) (GemSymmetryReport, error) {
	if err := manager.check(); err != nil {
		return GemSymmetryReport{}, err
	}
	if index < 0 || index >= len(manager.objects) {
		return GemSymmetryReport{}, ErrDimension
	}
	table := manager.tables[index]
	polar, seen := uint32(0), uint32(0)
	for assignment := 0; assignment < 1<<uint(table.variables); assignment++ {
		weight := bits.OnesCount(uint(assignment))
		value, _ := table.Get(uint64(assignment))
		if seen&(uint32(1)<<uint(weight)) != 0 {
			if value != (polar&(uint32(1)<<uint(weight)) != 0) {
				return GemSymmetryReport{}, nil
			}
		} else {
			seen |= uint32(1) << uint(weight)
			if value {
				polar |= uint32(1) << uint(weight)
			}
		}
	}
	half := (table.variables + 1) / 2
	majorityPolar := ((uint32(1) << uint(half)) - 1) << uint(half)
	return GemSymmetryReport{Symmetric: true, Majority: table.variables&1 != 0 && polar == majorityPolar, Polar: polar}, nil
}

func (manager *GemManager) FormatObject(index int, label string, updated bool) (string, error) {
	if err := manager.check(); err != nil {
		return "", err
	}
	if index < 0 || index >= len(manager.objects) {
		return "", ErrDimension
	}
	if !manager.verbose {
		return "", nil
	}
	if len(label) > manager.limits.MaxOutputBytes {
		return "", ErrLimit
	}
	object := manager.objects[index]
	profile := make([]byte, manager.variables)
	for i := range profile {
		if object.Groups&(uint16(1)<<uint(manager.variables-1-i)) != 0 {
			profile[i] = '1'
		} else {
			profile[i] = '0'
		}
	}
	marker := ""
	if updated {
		marker = " *"
	}
	result := fmt.Sprintf("Node %6d : %s  Pred = %6d  Vars = %d  Nodes = %d  History = %d%d  Profile: %s%s\n", index, label, object.Predecessor, object.Variables, object.Nodes, object.History&15, object.History>>4, profile, marker)
	if len(result) > manager.limits.MaxOutputBytes {
		return "", ErrLimit
	}
	return result, nil
}

func (manager *GemManager) Close() bool {
	if manager == nil || manager.closed {
		return false
	}
	manager.objects, manager.tables, manager.unique = nil, nil, nil
	manager.closed = true
	return true
}

func EnumerateGemMajority(variables int, options GemEnumerationOptions) (GemEnumerationResult, error) {
	manager, err := NewGemManager(variables, options.Verbose, options.Limits)
	if err != nil {
		return GemEnumerationResult{}, err
	}
	defer manager.Close()
	result := GemEnumerationResult{}
	previousStop, stop := 1, 1
	for variable := 1; variable <= variables-2; variable++ {
		previousStop, stop = stop, len(manager.objects)
		stage := GemStage{Variable: variable, Before: stop}
		for function := 0; function < stop; function++ {
			object := manager.objects[function]
			if variable != object.Variables && !(variable > object.Variables && function >= previousStop) {
				continue
			}
			for input := 0; input < variable; input++ {
				if object.Groups&(uint16(1)<<uint(input)) == 0 {
					continue
				}
				if _, _, err = manager.Expand(function, input); err != nil {
					return GemEnumerationResult{}, err
				}
			}
		}
		stage.AfterExpand = len(manager.objects)
		for function := stop; function < len(manager.objects); function++ {
			object := manager.objects[function]
			for first := 0; first < object.Variables; first++ {
				if object.Groups&(uint16(1)<<uint(first)) == 0 {
					continue
				}
				for second := first + 1; second < object.Variables; second++ {
					if object.Groups&(uint16(1)<<uint(second)) == 0 {
						continue
					}
					if _, _, err = manager.Reduce(function, first, second); err != nil {
						return GemEnumerationResult{}, err
					}
				}
			}
		}
		stage.AfterReduce = len(manager.objects)
		result.Stages = append(result.Stages, stage)
	}
	result.Objects = slices.Clone(manager.objects)
	if options.DumpTables {
		result.Tables = make([]*TruthTable, len(manager.tables))
		bytes := 0
		for i, table := range manager.tables {
			bytes += len(table.words) * 8
			if bytes > manager.limits.MaxOutputBytes {
				return GemEnumerationResult{}, ErrLimit
			}
			result.Tables[i] = table.Clone()
		}
	}
	if options.Verbose {
		var builder strings.Builder
		for i := range manager.objects {
			line, lineErr := manager.FormatObject(i, "Enumerated", false)
			if lineErr != nil {
				return GemEnumerationResult{}, lineErr
			}
			if builder.Len()+len(line) > manager.limits.MaxOutputBytes {
				return GemEnumerationResult{}, ErrLimit
			}
			builder.WriteString(line)
			result.Diagnostics = append(result.Diagnostics, line)
		}
	}
	return result, nil
}
