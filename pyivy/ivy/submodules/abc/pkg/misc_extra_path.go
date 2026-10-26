package abc

// Grid path enumeration from src/misc/extra/extraUtilPath.c.

import (
	"encoding/binary"
	"fmt"
	"math"
	"math/big"
	"strings"
)

type ExtraGridEdge struct{ From, To int }

type ExtraGridLife struct{ First, Last int }

const extraGridMaximumSize = 1024

// ExtraGridGraph preserves the source edge and frontier order. Frontier[i]
// is the live vertex set immediately before edge i; the final entry is empty.
type ExtraGridGraph struct {
	Size      int
	Nodes     int
	Edges     []ExtraGridEdge
	Life      []ExtraGridLife
	Frontiers [][]int
}

func BuildExtraGridGraph(size int) (ExtraGridGraph, error) {
	if size < 2 {
		return ExtraGridGraph{}, ErrDimension
	}
	if size > extraGridMaximumSize {
		return ExtraGridGraph{}, ErrLimit
	}
	g := ExtraGridGraph{Size: size, Nodes: size * size}
	for row := 0; row < size; row++ {
		for column := 0; column+1 < size; column++ {
			g.Edges = append(g.Edges, ExtraGridEdge{row*size + column, row*size + column + 1})
		}
		if row+1 < size {
			for column := 0; column < size; column++ {
				g.Edges = append(g.Edges, ExtraGridEdge{row*size + column, (row+1)*size + column})
			}
		}
	}
	g.Life = make([]ExtraGridLife, g.Nodes)
	for node := range g.Life {
		g.Life[node] = ExtraGridLife{-1, -1}
	}
	for edge, endpoints := range g.Edges {
		for _, node := range [...]int{endpoints.From, endpoints.To} {
			if g.Life[node].First < 0 {
				g.Life[node].First = edge
			}
			g.Life[node].Last = edge
		}
	}
	g.Frontiers = make([][]int, len(g.Edges)+1)
	live := make([]int, 0, 2*size)
	for edge, endpoints := range g.Edges {
		for _, node := range [...]int{endpoints.From, endpoints.To} {
			if g.Life[node].First == edge {
				live = append(live, node)
			} else if g.Life[node].Last == edge {
				for index, value := range live {
					if value == node {
						live = append(live[:index], live[index+1:]...)
						break
					}
				}
			}
		}
		g.Frontiers[edge+1] = append([]int(nil), live...)
	}
	return g, nil
}

type ExtraGridFrontierOptions struct {
	MaxStatesPerLevel int
	MaxTotalStates    int
	MaxTransitions    int
	Stop              func() bool
}

type ExtraGridFrontier struct {
	Graph       ExtraGridGraph
	States      [][][]uint32
	Transitions [][][2]int
	StateCounts []int
	TotalStates int
	PathCount   *big.Int
}

func extraGridStateKey(state []uint32) string {
	bytes := make([]byte, 4*len(state))
	for index, value := range state {
		binary.LittleEndian.PutUint32(bytes[4*index:], value)
	}
	return string(bytes)
}

func (g ExtraGridGraph) nextState(state []uint32, edgeIndex, selected int) ([]uint32, bool) {
	front := make([]uint32, g.Nodes)
	for index, node := range g.Frontiers[edgeIndex] {
		front[node] = state[index]
	}
	edge := g.Edges[edgeIndex]
	ends := [...]int{edge.From, edge.To}
	for _, node := range ends {
		if g.Life[node].First == edgeIndex {
			front[node] = uint32(node)
		}
	}
	if selected != 0 {
		if front[edge.From]&0xffff == front[edge.To]&0xffff {
			return nil, false
		}
		for _, node := range ends {
			degree := front[node] >> 16
			limit := uint32(2)
			if node == 0 || node == g.Nodes-1 {
				limit = 1
			}
			if degree >= limit {
				return nil, false
			}
			front[node] += 1 << 16
		}
	}
	equivalence := [2]uint32{front[edge.From] & 0xffff, front[edge.To] & 0xffff}
	for endpoint, node := range ends {
		if g.Life[node].Last != edgeIndex {
			continue
		}
		degree := front[node] >> 16
		if node == 0 || node == g.Nodes-1 {
			if degree != 1 {
				return nil, false
			}
		} else if degree != 0 && degree != 2 {
			return nil, false
		}
		first := -1
		for _, remaining := range g.Frontiers[edgeIndex+1] {
			if front[remaining]&0xffff == equivalence[endpoint] {
				if first < 0 {
					first = remaining
				}
				front[remaining] = front[remaining]&0xffff0000 | uint32(first)
			}
		}
		if first >= 0 {
			equivalence[endpoint] = uint32(first)
		}
	}
	if selected != 0 {
		first := -1
		for _, node := range g.Frontiers[edgeIndex+1] {
			component := front[node] & 0xffff
			if component == equivalence[0] || component == equivalence[1] {
				if first < 0 {
					first = node
				}
				front[node] = front[node]&0xffff0000 | uint32(first)
			}
		}
	}
	next := make([]uint32, len(g.Frontiers[edgeIndex+1]))
	for index, node := range g.Frontiers[edgeIndex+1] {
		next[index] = front[node]
	}
	return next, true
}

func BuildExtraGridFrontier(size int, options ExtraGridFrontierOptions) (ExtraGridFrontier, error) {
	fail := func(err error) (ExtraGridFrontier, error) { return ExtraGridFrontier{}, err }
	if options.MaxStatesPerLevel < 0 || options.MaxTotalStates < 0 || options.MaxTransitions < 0 {
		return fail(ErrDimension)
	}
	// Component representatives occupy the low 16 bits of every source state.
	if size > 256 {
		return fail(ErrLimit)
	}
	if options.MaxStatesPerLevel == 0 {
		options.MaxStatesPerLevel = 1 << 20
	}
	if options.MaxTotalStates == 0 {
		options.MaxTotalStates = 8 << 20
	}
	if options.MaxTransitions == 0 {
		options.MaxTransitions = 16 << 20
	}
	g, err := BuildExtraGridGraph(size)
	if err != nil {
		return fail(err)
	}
	if options.Stop != nil && options.Stop() {
		return fail(ErrCancelled)
	}
	result := ExtraGridFrontier{Graph: g, States: make([][][]uint32, len(g.Edges)), Transitions: make([][][2]int, len(g.Edges)), StateCounts: make([]int, len(g.Edges))}
	current := [][]uint32{{}}
	transitions := 0
	for edge := range g.Edges {
		if options.Stop != nil && options.Stop() {
			return fail(ErrCancelled)
		}
		result.States[edge] = current
		result.StateCounts[edge] = len(current)
		result.TotalStates += len(current)
		if len(current) > options.MaxStatesPerLevel || result.TotalStates > options.MaxTotalStates || transitions > options.MaxTransitions-2*len(current) {
			return fail(ErrLimit)
		}
		transitions += 2 * len(current)
		result.Transitions[edge] = make([][2]int, len(current))
		next := make([][]uint32, 0)
		indexes := make(map[string]int)
		for stateIndex, state := range current {
			for selected := 0; selected < 2; selected++ {
				stateNew, valid := g.nextState(state, edge, selected)
				destination := -1
				if valid {
					if edge+1 == len(g.Edges) {
						destination = -2
					} else {
						key := extraGridStateKey(stateNew)
						var found bool
						if destination, found = indexes[key]; !found {
							if len(next) == options.MaxStatesPerLevel || result.TotalStates+len(next)+1 > options.MaxTotalStates {
								return fail(ErrLimit)
							}
							destination = len(next)
							indexes[key] = destination
							next = append(next, stateNew)
						}
					}
				}
				result.Transitions[edge][stateIndex][selected] = destination
			}
		}
		current = next
	}
	memo := make([][]*big.Int, len(result.Transitions))
	for level := range memo {
		memo[level] = make([]*big.Int, len(result.Transitions[level]))
	}
	var count func(int, int) *big.Int
	count = func(level, state int) *big.Int {
		if state == -2 {
			return big.NewInt(1)
		}
		if state == -1 {
			return new(big.Int)
		}
		if memo[level][state] != nil {
			return new(big.Int).Set(memo[level][state])
		}
		value := new(big.Int).Add(count(level+1, result.Transitions[level][state][0]), count(level+1, result.Transitions[level][state][1]))
		memo[level][state] = new(big.Int).Set(value)
		return value
	}
	result.PathCount = count(0, 0)
	return result, nil
}

func (frontier ExtraGridFrontier) DeriveGIA(maxObjects int) (*GiaManager, error) {
	if frontier.PathCount == nil || len(frontier.Transitions) == 0 || maxObjects < 0 {
		return nil, ErrDimension
	}
	if maxObjects == 0 {
		maxObjects = int(MaxGiaID) + 1
	}
	m, err := NewGiaManagerWithLimit(maxObjects)
	if err != nil {
		return nil, err
	}
	inputs := make([]GiaLit, len(frontier.Transitions))
	for index := range inputs {
		if inputs[index], err = m.AppendCI(); err != nil {
			return nil, err
		}
	}
	if err = m.HashStart(); err != nil {
		return nil, err
	}
	memo := make([][]GiaLit, len(frontier.Transitions))
	for level := range memo {
		memo[level] = make([]GiaLit, len(frontier.Transitions[level]))
		for state := range memo[level] {
			memo[level][state] = InvalidGiaLit
		}
	}
	var derive func(int, int) (GiaLit, error)
	derive = func(level, state int) (GiaLit, error) {
		if state == -2 {
			return GiaConst1, nil
		}
		if state == -1 {
			return GiaConst0, nil
		}
		if memo[level][state] != InvalidGiaLit {
			return memo[level][state], nil
		}
		zero, e := derive(level+1, frontier.Transitions[level][state][0])
		if e != nil {
			return InvalidGiaLit, e
		}
		one, e := derive(level+1, frontier.Transitions[level][state][1])
		if e != nil {
			return InvalidGiaLit, e
		}
		memo[level][state], e = m.HashMux(inputs[level], one, zero)
		return memo[level][state], e
	}
	root, err := derive(0, 0)
	if err != nil {
		return nil, err
	}
	if _, err = m.AppendCO(root); err != nil {
		return nil, err
	}
	clean, _, err := m.Cleanup()
	if err != nil {
		return nil, err
	}
	clean.Name = "paths"
	return clean, nil
}

func ExtraGridAIGFilename(size, edges int) string {
	return fmt.Sprintf("grid_%dx%d_e%03d.aig", size, size, edges)
}

func (g ExtraGridGraph) FormatPath(selected []bool) (string, error) {
	if len(selected) != len(g.Edges) {
		return "", ErrDimension
	}
	dimension := 4*(g.Size-1) + 1
	box := make([][]byte, dimension)
	for row := range box {
		box[row] = []byte(strings.Repeat(" ", dimension))
	}
	for row := 0; row < g.Size; row++ {
		for column := 0; column < g.Size; column++ {
			box[4*row][4*column] = '*'
		}
	}
	for index, edge := range g.Edges {
		if !selected[index] {
			continue
		}
		y0, x0 := 4*(edge.From/g.Size), 4*(edge.From%g.Size)
		y1, x1 := 4*(edge.To/g.Size), 4*(edge.To%g.Size)
		if y0 == y1 {
			for x := x0 + 1; x < x1; x++ {
				box[y0][x] = '-'
			}
		} else if x0 == x1 {
			for y := y0 + 1; y < y1; y++ {
				box[y][x0] = '|'
			}
		} else {
			return "", ErrInvariant
		}
	}
	lines := make([]string, len(box))
	for row := range box {
		lines[row] = string(box[row])
	}
	return strings.Join(lines, "\n") + "\n", nil
}

func ExtraMonotoneHorizontalVariable(size, y, x int) (int, error) {
	if size < 1 || y < 0 || y > size || x < 1 || x > size {
		return 0, ErrDimension
	}
	return size*y + x - 1, nil
}

func ExtraMonotoneVerticalVariable(size, y, x int) (int, error) {
	if size < 1 || y < 1 || y > size || x < 0 || x > size {
		return 0, ErrDimension
	}
	return size*(size+1) + size*x + y - 1, nil
}

func BuildExtraMonotonePathsGIA(size, maxObjects int) (*GiaManager, error) {
	if size < 1 || maxObjects < 0 {
		return nil, ErrDimension
	}
	if maxObjects == 0 {
		maxObjects = int(MaxGiaID) + 1
	}
	if size > maxObjects {
		return nil, ErrLimit
	}
	variables, err := checkedSize(0, 2*size, size+1, maxObjects-1)
	if err != nil {
		return nil, err
	}
	m, err := NewGiaManagerWithLimit(maxObjects)
	if err != nil {
		return nil, err
	}
	inputs := make([]GiaLit, variables)
	for index := range inputs {
		if inputs[index], err = m.AppendCI(); err != nil {
			return nil, err
		}
	}
	if err = m.HashStart(); err != nil {
		return nil, err
	}
	nodes := make([]GiaLit, size+1)
	nodes[0] = GiaConst1
	for x := 1; x <= size; x++ {
		v, _ := ExtraMonotoneHorizontalVariable(size, 0, x)
		if nodes[x], err = m.HashAnd(nodes[x-1], inputs[v]); err != nil {
			return nil, err
		}
	}
	for y := 1; y <= size; y++ {
		v, _ := ExtraMonotoneVerticalVariable(size, y, 0)
		if nodes[0], err = m.HashAnd(nodes[0], inputs[v]); err != nil {
			return nil, err
		}
		for x := 1; x <= size; x++ {
			h, _ := ExtraMonotoneHorizontalVariable(size, y, x)
			vertical, _ := ExtraMonotoneVerticalVariable(size, y, x)
			left, e := m.HashAnd(nodes[x-1], inputs[h])
			if e != nil {
				return nil, e
			}
			up, e := m.HashAnd(nodes[x], inputs[vertical])
			if e != nil {
				return nil, e
			}
			if nodes[x], e = m.HashOr(left, up); e != nil {
				return nil, e
			}
		}
	}
	if _, err = m.AppendCO(nodes[size]); err != nil {
		return nil, err
	}
	clean, _, err := m.Cleanup()
	return clean, err
}

func ExtraWordToFloat64(value uint64) float64 { return float64(value) }

type ExtraGraphSolveOptions struct {
	MaxModels, MaxVariables, MaxClauses, MaxLiterals int
	SAT                                              SatLimits
	Stop                                             func() bool
}

type ExtraGraphSolveResult struct {
	Variables, Iterations int
	Models                [][]bool
	First, Last           uint64
	Span, Average         float64
	Estimate              float64
	Complete              bool
}

// SolveExtraGraphPaths reproduces the source's lexicographic SAT experiment,
// but returns every sampled assignment and defined statistics instead of
// printing or dividing by zero.
func SolveExtraGraphPaths(graph *GiaManager, options ExtraGraphSolveOptions) (ExtraGraphSolveResult, error) {
	fail := func(err error) (ExtraGraphSolveResult, error) { return ExtraGraphSolveResult{}, err }
	if graph == nil || graph.PrimaryOutputCount() != 1 || graph.RegisterCount() != 0 || options.MaxModels < 0 {
		return fail(ErrDimension)
	}
	variables := graph.PrimaryInputCount()
	if variables > 64 {
		return fail(ErrUnsupported)
	}
	if options.MaxModels == 0 {
		options.MaxModels = 1000
	}
	loader, err := NewGiaLazyCNF(graph, GiaLazyCNFOptions{MaxVariables: options.MaxVariables, MaxClauses: options.MaxClauses, MaxLiterals: options.MaxLiterals, Stop: options.Stop})
	if err != nil {
		return fail(err)
	}
	defer loader.Close()
	co := graph.cos[0]
	driver, err := graph.objects[co].Fanin0(co)
	if err != nil {
		return fail(err)
	}
	if err = loader.LoadCone(driver); err != nil {
		return fail(err)
	}
	output, err := loader.satLiteral(driver)
	if err != nil {
		return fail(err)
	}
	inputLits := make([]SatLit, variables)
	for index, id := range graph.cis {
		inputLits[index], err = loader.satLiteral(GiaLit(id) << 1)
		if err != nil {
			return fail(err)
		}
	}
	result := ExtraGraphSolveResult{Variables: variables}
	for len(result.Models) < options.MaxModels {
		if options.Stop != nil && options.Stop() {
			return fail(ErrCancelled)
		}
		base, solveErr := loader.solver.Solve([]SatLit{output}, options.SAT)
		if solveErr != nil {
			return fail(solveErr)
		}
		if base.Status == SatUnsatisfiable {
			result.Complete = true
			break
		}
		if base.Status != SatSatisfiable {
			break
		}
		prefix := []SatLit{output}
		model := make([]bool, variables)
		for index, literal := range inputLits {
			zero, _ := MakeSatLit(literal.Var(), true)
			probe := append(append([]SatLit(nil), prefix...), zero)
			answer, e := loader.solver.Solve(probe, options.SAT)
			if e != nil {
				return fail(e)
			}
			if answer.Status == SatSatisfiable {
				prefix = probe
			} else if answer.Status == SatUnsatisfiable {
				model[index] = true
				prefix = append(prefix, zero.Not())
			} else {
				return result, nil
			}
		}
		block := make([]SatLit, variables)
		var number uint64
		for index, literal := range inputLits {
			if model[index] {
				number |= uint64(1) << uint(variables-1-index)
				block[index], _ = MakeSatLit(literal.Var(), true)
			} else {
				block[index], _ = MakeSatLit(literal.Var(), false)
			}
		}
		if _, err = loader.solver.AddClause(block); err != nil {
			return fail(err)
		}
		result.Models = append(result.Models, model)
		result.Iterations++
		if len(result.Models) == 1 {
			result.First = number
		}
		result.Last = number
	}
	if result.Iterations > 0 {
		result.Span = ExtraWordToFloat64(result.Last - result.First)
		result.Average = result.Span / float64(result.Iterations)
		if result.Average > 0 {
			result.Estimate = (math.Exp2(float64(variables)) - ExtraWordToFloat64(result.First)) / result.Average
		}
	}
	return result, nil
}
