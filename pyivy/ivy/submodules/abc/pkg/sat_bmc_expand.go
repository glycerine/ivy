package abc

// SOP cube expansion from src/sat/bmc/bmcExpand.c.

import (
	"slices"
	"sort"
	"strings"
)

type SOPCover struct {
	Variables int
	Cubes     []string
	Output    byte // '0', '1', 'n' or 'x'
}

type SOPNode struct {
	Fanins []int
	Cover  SOPCover
}

// SOPNetwork is the concrete portion of an SOP logic network needed by
// Abc_NtkExpandCubes. Fanins are input ordinals for output-driving nodes and
// otherwise remain opaque stable object IDs while readability sorting runs.
type SOPNetwork struct {
	InputCount int
	Nodes      []SOPNode
	Outputs    []int // node indexes; -1 denotes a non-node/constant driver
}

type SOPExpansionOptions struct {
	MaxVariables, MaxClauses, MaxLiterals int
	MaxCubes, MaxCubeEntries              int
	SAT                                   SatLimits
	Stop                                  func() bool
}

type SOPExpansionResult struct {
	Cover    SOPCover
	Constant bool
	Changed  bool
	Status   SatStatus
	Reason   SatStopReason
	Stats    SatStats
}

type SOPNetworkExpansionResult struct {
	Network SOPNetwork
	Outputs []SOPExpansionResult
	Stats   SatStats
}

func cloneSOPCover(cover SOPCover) SOPCover {
	cover.Cubes = slices.Clone(cover.Cubes)
	return cover
}

func cloneSOPNetwork(network SOPNetwork) SOPNetwork {
	result := SOPNetwork{InputCount: network.InputCount, Outputs: slices.Clone(network.Outputs), Nodes: make([]SOPNode, len(network.Nodes))}
	for index, node := range network.Nodes {
		result.Nodes[index] = SOPNode{Fanins: slices.Clone(node.Fanins), Cover: cloneSOPCover(node.Cover)}
	}
	return result
}

func checkSOPExpansionOptions(options SOPExpansionOptions) (SOPExpansionOptions, error) {
	if options.MaxVariables < 0 || options.MaxClauses < 0 || options.MaxLiterals < 0 || options.MaxCubes < 0 || options.MaxCubeEntries < 0 {
		return options, ErrDimension
	}
	if options.MaxVariables == 0 {
		options.MaxVariables = 1 << 20
	}
	if options.MaxClauses == 0 {
		options.MaxClauses = 4 << 20
	}
	if options.MaxLiterals == 0 {
		options.MaxLiterals = 12 << 20
	}
	if options.MaxCubes == 0 {
		options.MaxCubes = 1 << 20
	}
	if options.MaxCubeEntries == 0 {
		options.MaxCubeEntries = 16 << 20
	}
	if options.MaxVariables < 1 || options.MaxClauses < 1 || options.MaxLiterals < 1 || options.MaxCubes < 1 || options.MaxCubeEntries < 1 {
		return options, ErrLimit
	}
	return options, nil
}

func validateSOPCover(cover SOPCover, options SOPExpansionOptions) error {
	if cover.Variables < 0 || len(cover.Cubes) == 0 || !strings.ContainsRune("01nx", rune(cover.Output)) {
		return ErrDimension
	}
	if cover.Variables > options.MaxVariables || len(cover.Cubes) > options.MaxCubes || cover.Variables > options.MaxCubeEntries/len(cover.Cubes) {
		return ErrLimit
	}
	for _, cube := range cover.Cubes {
		if len(cube) != cover.Variables {
			return ErrDimension
		}
		for index := range cube {
			if cube[index] != '-' && cube[index] != '0' && cube[index] != '1' {
				return ErrDimension
			}
		}
	}
	return nil
}

func (cover SOPCover) Evaluate(inputs []bool) (bool, error) {
	if len(inputs) != cover.Variables {
		return false, ErrDimension
	}
	matched := false
	for _, cube := range cover.Cubes {
		if len(cube) != cover.Variables {
			return false, ErrDimension
		}
		this := true
		for variable := range cube {
			switch cube[variable] {
			case '-':
			case '0':
				this = this && !inputs[variable]
			case '1':
				this = this && inputs[variable]
			default:
				return false, ErrDimension
			}
		}
		matched = matched || this
	}
	switch cover.Output {
	case '0':
		return !matched, nil
	case '1':
		return matched, nil
	case 'n', 'x':
		parity := false
		for _, input := range inputs {
			parity = parity != input
		}
		if cover.Output == 'n' {
			parity = !parity
		}
		return parity, nil
	}
	return false, ErrDimension
}

func ParseSOPCover(text string, options SOPExpansionOptions) (SOPCover, error) {
	checked, err := checkSOPExpansionOptions(options)
	if err != nil {
		return SOPCover{}, err
	}
	if text == "" || !strings.HasSuffix(text, "\n") {
		return SOPCover{}, ErrDimension
	}
	lines := strings.Split(strings.TrimSuffix(text, "\n"), "\n")
	if len(lines) > checked.MaxCubes {
		return SOPCover{}, ErrLimit
	}
	cover := SOPCover{Variables: -1, Cubes: make([]string, len(lines))}
	for index, line := range lines {
		if len(line) < 2 || line[len(line)-2] != ' ' || !strings.ContainsRune("01nx", rune(line[len(line)-1])) {
			return SOPCover{}, ErrDimension
		}
		cube := line[:len(line)-2]
		if cover.Variables < 0 {
			cover.Variables = len(cube)
			cover.Output = line[len(line)-1]
		} else if len(cube) != cover.Variables || cover.Output != line[len(line)-1] {
			return SOPCover{}, ErrDimension
		}
		cover.Cubes[index] = cube
	}
	if err = validateSOPCover(cover, checked); err != nil {
		return SOPCover{}, err
	}
	return cover, nil
}

func FormatSOPCover(cover SOPCover) (string, error) {
	options, _ := checkSOPExpansionOptions(SOPExpansionOptions{})
	if err := validateSOPCover(cover, options); err != nil {
		return "", err
	}
	var builder strings.Builder
	for _, cube := range cover.Cubes {
		builder.WriteString(cube)
		builder.WriteByte(' ')
		builder.WriteByte(cover.Output)
		builder.WriteByte('\n')
	}
	return builder.String(), nil
}

func addSatStats(first *SatStats, second SatStats) {
	first.Conflicts += second.Conflicts
	first.Propagations += second.Propagations
	first.Decisions += second.Decisions
	first.Inspections += second.Inspections
	first.Restarts += second.Restarts
	first.Learnt += second.Learnt
	first.Deleted += second.Deleted
}

func sopCubeAssumptions(cube string, variables []SatVar) []SatLit {
	result := make([]SatLit, 0, len(cube))
	for index := range cube {
		if cube[index] == '-' {
			continue
		}
		literal, _ := MakeSatLit(variables[index], cube[index] == '0')
		result = append(result, literal)
	}
	return result
}

// ExpandSOPCubesWithSolver represents Abc_ObjExpandCubesTry. solver must
// already encode the offset; variables map cover columns to solver variables.
func ExpandSOPCubesWithSolver(cover SOPCover, solver *SatSolver, variables []SatVar, options SOPExpansionOptions) (SOPExpansionResult, error) {
	checked, err := checkSOPExpansionOptions(options)
	if err != nil {
		return SOPExpansionResult{}, err
	}
	if solver == nil || len(variables) != cover.Variables {
		return SOPExpansionResult{}, ErrDimension
	}
	if err = validateSOPCover(cover, checked); err != nil {
		return SOPExpansionResult{}, err
	}
	for _, variable := range variables {
		if uint64(variable) >= uint64(solver.VariableCount()) {
			return SOPExpansionResult{}, ErrInvalidLiteral
		}
	}
	result := SOPExpansionResult{Cover: cloneSOPCover(cover), Status: SatSatisfiable}
	for cubeIndex, original := range cover.Cubes {
		candidate := []byte(original)
		for variable := cover.Variables - 1; variable >= 0; variable-- {
			if candidate[variable] == '-' {
				continue
			}
			if checked.Stop != nil && checked.Stop() {
				return SOPExpansionResult{}, ErrCancelled
			}
			saved := candidate[variable]
			candidate[variable] = '-'
			answer, solveErr := solver.Solve(sopCubeAssumptions(string(candidate), variables), checked.SAT)
			if solveErr != nil {
				return SOPExpansionResult{}, solveErr
			}
			addSatStats(&result.Stats, answer.Stats)
			if answer.Status == SatUnknown {
				return SOPExpansionResult{Cover: cloneSOPCover(cover), Status: SatUnknown, Reason: answer.Reason, Stats: result.Stats}, nil
			}
			if answer.Status == SatSatisfiable {
				candidate[variable] = saved
			}
		}
		result.Cover.Cubes[cubeIndex] = string(candidate)
		result.Changed = result.Changed || string(candidate) != original
	}
	return result, nil
}

func sopCoverIrredundant(cover SOPCover, options SOPExpansionOptions) (SOPCover, SatStats, SatStatus, SatStopReason, error) {
	if len(cover.Cubes) <= 1 {
		return cover, SatStats{}, SatSatisfiable, SatFinished, nil
	}
	kept := make([]bool, len(cover.Cubes))
	for index := range kept {
		kept[index] = true
	}
	var stats SatStats
	variables := make([]SatVar, cover.Variables)
	for index := range variables {
		variables[index] = SatVar(index)
	}
	for current, cube := range cover.Cubes {
		if options.Stop != nil && options.Stop() {
			return SOPCover{}, SatStats{}, SatUnknown, SatCancelled, ErrCancelled
		}
		solver, err := NewSatSolverWithLimit(options.MaxVariables)
		if err != nil {
			return SOPCover{}, SatStats{}, SatUnknown, SatFinished, err
		}
		if err = solver.SetVariableCount(cover.Variables); err != nil {
			return SOPCover{}, SatStats{}, SatUnknown, SatFinished, err
		}
		clauses, literals := 0, 0
		for other, otherCube := range cover.Cubes {
			if other == current || !kept[other] {
				continue
			}
			clause := make([]SatLit, 0, cover.Variables)
			for variable := range otherCube {
				if otherCube[variable] == '-' {
					continue
				}
				literal, _ := MakeSatLit(SatVar(variable), otherCube[variable] == '1')
				clause = append(clause, literal)
			}
			if len(clause) == 0 {
				kept[current] = false
				break
			}
			clauses++
			literals += len(clause)
			if clauses > options.MaxClauses || literals > options.MaxLiterals {
				return SOPCover{}, SatStats{}, SatUnknown, SatFinished, ErrLimit
			}
			ok, addErr := solver.AddClause(clause)
			if addErr != nil {
				return SOPCover{}, SatStats{}, SatUnknown, SatFinished, addErr
			}
			if !ok {
				kept[current] = false
				break
			}
		}
		if !kept[current] {
			continue
		}
		answer, err := solver.Solve(sopCubeAssumptions(cube, variables), options.SAT)
		if err != nil {
			return SOPCover{}, SatStats{}, SatUnknown, SatFinished, err
		}
		addSatStats(&stats, answer.Stats)
		if answer.Status == SatUnknown {
			return cloneSOPCover(cover), stats, SatUnknown, answer.Reason, nil
		}
		if answer.Status == SatUnsatisfiable {
			kept[current] = false
		}
	}
	result := cloneSOPCover(cover)
	result.Cubes = result.Cubes[:0]
	for index, cube := range cover.Cubes {
		if kept[index] {
			result.Cubes = append(result.Cubes, cube)
		}
	}
	if len(result.Cubes) == 0 {
		return SOPCover{}, stats, SatUnknown, SatFinished, ErrInvariant
	}
	return result, stats, SatSatisfiable, SatFinished, nil
}

// ExpandGiaSOPCover represents Abc_ObjExpandCubes for one selected GIA output.
func (manager *GiaManager) ExpandGiaSOPCover(output int, inputOrder []int, cover SOPCover, options SOPExpansionOptions) (SOPExpansionResult, error) {
	checked, err := checkSOPExpansionOptions(options)
	if err != nil {
		return SOPExpansionResult{}, err
	}
	if manager == nil || output < 0 || output >= manager.OutputCount() || cover.Variables != manager.InputCount() || len(inputOrder) != manager.InputCount() {
		return SOPExpansionResult{}, ErrDimension
	}
	if err = validateSOPCover(cover, checked); err != nil {
		return SOPExpansionResult{}, err
	}
	seen := make([]bool, manager.InputCount())
	for _, ordinal := range inputOrder {
		if ordinal < 0 || ordinal >= len(seen) || seen[ordinal] {
			return SOPExpansionResult{}, ErrDimension
		}
		seen[ordinal] = true
	}
	if checked.Stop != nil && checked.Stop() {
		return SOPExpansionResult{}, ErrCancelled
	}
	aig, _, err := manager.ToAig()
	if err != nil {
		return SOPExpansionResult{}, err
	}
	cnf, err := EncodeAigCNF(aig)
	if err != nil {
		return SOPExpansionResult{}, err
	}
	clauses := cnf.Clauses()
	literals := 0
	for _, clause := range clauses {
		literals += len(clause)
	}
	if cnf.VariableCount() > checked.MaxVariables || len(clauses)+1 > checked.MaxClauses || literals+1 > checked.MaxLiterals {
		return SOPExpansionResult{}, ErrLimit
	}
	solver, err := NewSatSolverWithLimit(checked.MaxVariables)
	if err != nil {
		return SOPExpansionResult{}, err
	}
	if ok, loadErr := cnf.LoadInto(solver, 0); loadErr != nil || !ok {
		if loadErr != nil {
			return SOPExpansionResult{}, loadErr
		}
		return SOPExpansionResult{}, ErrInvariant
	}
	out := cnf.OutputLiterals()[output]
	result := SOPExpansionResult{Cover: cloneSOPCover(cover)}
	on, err := solver.Solve([]SatLit{out}, checked.SAT)
	if err != nil {
		return SOPExpansionResult{}, err
	}
	addSatStats(&result.Stats, on.Stats)
	if on.Status == SatUnknown {
		result.Status, result.Reason = SatUnknown, on.Reason
		return result, nil
	}
	off, err := solver.Solve([]SatLit{out.Not()}, checked.SAT)
	if err != nil {
		return SOPExpansionResult{}, err
	}
	addSatStats(&result.Stats, off.Stats)
	if off.Status == SatUnknown {
		result.Status, result.Reason = SatUnknown, off.Reason
		return result, nil
	}
	if on.Status == SatUnsatisfiable || off.Status == SatUnsatisfiable {
		value := off.Status == SatUnsatisfiable
		output := byte('0')
		if value {
			output = '1'
		}
		result.Cover = SOPCover{Variables: 0, Cubes: []string{""}, Output: output}
		result.Constant, result.Changed, result.Status = true, true, SatSatisfiable
		return result, nil
	}
	if ok, addErr := solver.AddClause([]SatLit{out.Not()}); addErr != nil || !ok {
		if addErr != nil {
			return SOPExpansionResult{}, addErr
		}
		return SOPExpansionResult{}, ErrInvariant
	}
	inputs := cnf.InputLiterals()
	variables := make([]SatVar, len(inputOrder))
	for column, ordinal := range inputOrder {
		variables[column] = inputs[ordinal].Var()
	}
	expanded, err := ExpandSOPCubesWithSolver(cover, solver, variables, checked)
	if err != nil {
		return SOPExpansionResult{}, err
	}
	addSatStats(&result.Stats, expanded.Stats)
	if expanded.Status == SatUnknown {
		expanded.Stats = result.Stats
		return expanded, nil
	}
	irredundant, irredundantStats, irredundantStatus, irredundantReason, err := sopCoverIrredundant(expanded.Cover, checked)
	if err != nil {
		return SOPExpansionResult{}, err
	}
	addSatStats(&result.Stats, irredundantStats)
	if irredundantStatus == SatUnknown {
		result.Cover = cloneSOPCover(cover)
		result.Status, result.Reason = SatUnknown, irredundantReason
		return result, nil
	}
	result.Cover = irredundant
	result.Status = SatSatisfiable
	result.Changed = expanded.Changed || !slices.Equal(irredundant.Cubes, cover.Cubes)
	return result, nil
}

func sopPermuteNode(node *SOPNode, order []int) {
	newFanins := make([]int, len(order))
	newCubes := make([]string, len(node.Cover.Cubes))
	for index, old := range order {
		newFanins[index] = node.Fanins[old]
	}
	for cubeIndex, cube := range node.Cover.Cubes {
		values := make([]byte, len(order))
		for index, old := range order {
			values[index] = cube[old]
		}
		newCubes[cubeIndex] = string(values)
	}
	node.Fanins, node.Cover.Cubes = newFanins, newCubes
}

func sopSortCubes(node *SOPNode, weighted bool) {
	sort.Slice(node.Cover.Cubes, func(i, j int) bool {
		if weighted {
			left := len(node.Cover.Cubes[i]) - strings.Count(node.Cover.Cubes[i], "-")
			right := len(node.Cover.Cubes[j]) - strings.Count(node.Cover.Cubes[j], "-")
			if left != right {
				return left > right
			}
		}
		return node.Cover.Cubes[i] < node.Cover.Cubes[j]
	})
}

func sopOrderByLiteralCount(node *SOPNode) {
	n := len(node.Fanins)
	order := make([]int, n)
	counts := make([]int, n)
	columns := make([]string, n)
	for variable := 0; variable < n; variable++ {
		order[variable] = variable
		column := make([]byte, len(node.Cover.Cubes))
		for cubeIndex, cube := range node.Cover.Cubes {
			column[cubeIndex] = cube[variable]
			if cube[variable] != '-' {
				counts[variable]++
			}
		}
		columns[variable] = string(column)
	}
	for i := 0; i+1 < n; i++ {
		best := i
		for j := i + 1; j < n; j++ {
			left, right := order[j], order[best]
			if counts[left] < counts[right] || (counts[left] == counts[right] && columns[left] < columns[right]) {
				best = j
			}
		}
		order[i], order[best] = order[best], order[i]
	}
	sopPermuteNode(node, order)
}

func sopOrderSingletonsByCube(node *SOPNode) {
	n := len(node.Fanins)
	order := make([]int, n)
	cost := make([]int, n)
	for variable := 0; variable < n; variable++ {
		order[variable], cost[variable] = variable, int(^uint(0)>>1)
		count, cubeNumber := 0, 0
		for index, cube := range node.Cover.Cubes {
			if cube[variable] != '-' {
				count, cubeNumber = count+1, index
			}
		}
		if count == 1 {
			cost[variable] = cubeNumber
		}
	}
	for i := 0; i+1 < n; i++ {
		best := i
		for j := i + 1; j < n; j++ {
			if cost[order[j]] < cost[order[best]] {
				best = j
			}
		}
		order[i], order[best] = order[best], order[i]
	}
	sopPermuteNode(node, order)
}

func sortSOPNetwork(network *SOPNetwork) {
	for index := range network.Nodes {
		node := &network.Nodes[index]
		if len(node.Fanins) != node.Cover.Variables || len(node.Cover.Cubes) == 0 {
			continue
		}
		sopSortCubes(node, true)
		sopOrderByLiteralCount(node)
		sopSortCubes(node, false)
		sopOrderSingletonsByCube(node)
		sopSortCubes(node, false)
	}
}

// ExpandSOPNetwork represents Abc_NtkExpandCubes and its final SOP sorting.
func (manager *GiaManager) ExpandSOPNetwork(network SOPNetwork, options SOPExpansionOptions) (SOPNetworkExpansionResult, error) {
	checked, err := checkSOPExpansionOptions(options)
	if err != nil {
		return SOPNetworkExpansionResult{}, err
	}
	if manager == nil || network.InputCount != manager.InputCount() || len(network.Outputs) != manager.OutputCount() {
		return SOPNetworkExpansionResult{}, ErrDimension
	}
	result := SOPNetworkExpansionResult{Network: cloneSOPNetwork(network), Outputs: make([]SOPExpansionResult, len(network.Outputs))}
	for output, nodeIndex := range network.Outputs {
		if checked.Stop != nil && checked.Stop() {
			return SOPNetworkExpansionResult{}, ErrCancelled
		}
		if nodeIndex == -1 {
			continue
		}
		if nodeIndex < 0 || nodeIndex >= len(result.Network.Nodes) {
			return SOPNetworkExpansionResult{}, ErrDimension
		}
		node := &result.Network.Nodes[nodeIndex]
		if len(node.Fanins) == 0 {
			continue
		}
		if len(node.Fanins) != manager.InputCount() || node.Cover.Variables != len(node.Fanins) {
			return SOPNetworkExpansionResult{}, ErrDimension
		}
		expanded, expandErr := manager.ExpandGiaSOPCover(output, node.Fanins, node.Cover, checked)
		if expandErr != nil {
			return SOPNetworkExpansionResult{}, expandErr
		}
		result.Outputs[output] = expanded
		addSatStats(&result.Stats, expanded.Stats)
		if expanded.Status == SatUnknown {
			return result, nil
		}
		node.Cover = expanded.Cover
		if expanded.Constant {
			node.Fanins = nil
		}
	}
	sortSOPNetwork(&result.Network)
	return result, nil
}
