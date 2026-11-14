package cudd

import "github.com/glycerine/abcgo/bdd/cudd/core"

// Cudd_SubsetShortPaths extracts a subset by retaining nodes on the shortest
// paths to one, up to the threshold-selected path length.
func Cudd_SubsetShortPaths(dd *DdManager, f *DdNode, numVars, threshold, hardlimit int32) *DdNode {
	if dd == nil || f == nil {
		return nil
	}
	return cuddSubsetShortPaths(dd, f, numVars, threshold, hardlimit)
}

// Cudd_SupersetShortPaths extracts a subset of the complement and complements
// the result, which is the CUDD shortest-path superset construction.
func Cudd_SupersetShortPaths(dd *DdManager, f *DdNode, numVars, threshold, hardlimit int32) *DdNode {
	if dd == nil || f == nil {
		return nil
	}
	return core.Cudd_NotCond(cuddSubsetShortPaths(dd, core.Cudd_Not(f), numVars, threshold, hardlimit), true)
}

const unreachablePath int32 = int32(^uint32(0) >> 1)

type pathMetric struct{ length, bottom int32 }

type pathGraph struct {
	distance map[*DdNode]int32
	edges    []*DdNode
	bottom   map[*DdNode]int32
}

type subsetBuildState struct {
	dd               *DdManager
	metrics          map[*DdNode]pathMetric
	maxPath          int32
	thresholdReached int32
	findShortestPath bool
	results          map[*DdNode]*DdNode
	maxpathTable     map[*DdNode]struct{}
	subsetNodes      map[*DdNode]struct{}
	threshold        int32
}

func cuddSubsetShortPaths(dd *DdManager, f *DdNode, numVars, threshold, hardlimit int32) *DdNode {
	if f == nil || core.Cudd_IsConstant(f) {
		return f
	}
	if numVars == 0 {
		numVars = Cudd_ReadSize(dd)
	}
	if threshold > numVars {
		threshold -= numVars
	}

	graph := collectPathGraph(f)
	metrics := make(map[*DdNode]pathMetric, len(graph.edges))
	counts := make(map[int32]int32)
	for _, edge := range graph.edges {
		if core.Cudd_IsConstant(edge) {
			continue
		}
		reg := core.Cudd_Regular(edge)
		if _, seen := metrics[reg]; seen {
			continue
		}
		odd := pathMetric{length: unreachablePath, bottom: unreachablePath}
		even := pathMetric{length: unreachablePath, bottom: unreachablePath}
		for _, state := range graph.edges {
			if core.Cudd_Regular(state) != reg {
				continue
			}
			bottom := shortestToOne(dd, state, graph.bottom)
			if bottom == unreachablePath || graph.distance[state] == unreachablePath {
				continue
			}
			candidate := pathMetric{length: graph.distance[state] + bottom, bottom: bottom}
			if core.Cudd_IsComplement(state) {
				if candidate.length < odd.length {
					odd = candidate
				}
			} else if candidate.length < even.length {
				even = candidate
			}
		}
		metric := even
		if odd.length <= even.length {
			metric = odd
		}
		metrics[reg] = metric
		if metric.length < unreachablePath {
			counts[metric.length]++
		}
	}

	maxPath, excess := assessPathLength(counts, numVars, threshold)
	if maxPath == numVars+1 {
		return f
	}
	state := &subsetBuildState{
		dd: dd, metrics: metrics, maxPath: maxPath,
		thresholdReached: excess,
		results:          make(map[*DdNode]*DdNode), maxpathTable: make(map[*DdNode]struct{}),
	}
	if state.thresholdReached < 0 {
		state.thresholdReached = 0
	}
	if hardlimit != 0 {
		state.subsetNodes = make(map[*DdNode]struct{})
		state.threshold = threshold
	}
	return state.build(f)
}

// collectPathGraph performs the breadth-first top-distance pass used to
// distinguish regular and complemented paths.
func collectPathGraph(root *DdNode) pathGraph {
	graph := pathGraph{distance: map[*DdNode]int32{root: 0}, edges: []*DdNode{root}, bottom: make(map[*DdNode]int32)}
	queue := []*DdNode{root}
	for len(queue) > 0 {
		edge := queue[0]
		queue = queue[1:]
		if core.Cudd_IsConstant(edge) {
			continue
		}
		t, e := pathChildren(edge)
		for _, child := range []*DdNode{t, e} {
			if _, seen := graph.distance[child]; seen {
				continue
			}
			graph.distance[child] = graph.distance[edge] + 1
			graph.edges = append(graph.edges, child)
			queue = append(queue, child)
		}
	}
	return graph
}

func pathChildren(edge *DdNode) (*DdNode, *DdNode) {
	t, e := core.Cudd_T(edge), core.Cudd_E(edge)
	if core.Cudd_IsComplement(edge) {
		t, e = core.Cudd_Not(t), core.Cudd_Not(e)
	}
	return t, e
}

// shortestToOne computes bottom distances with memoization over edge polarity.
func shortestToOne(dd *DdManager, edge *DdNode, memo map[*DdNode]int32) int32 {
	if edge == Cudd_ReadOne(dd) {
		return 0
	}
	if core.Cudd_IsConstant(edge) {
		return unreachablePath
	}
	if value, ok := memo[edge]; ok {
		return value
	}
	t, e := pathChildren(edge)
	a, b := shortestToOne(dd, t, memo), shortestToOne(dd, e, memo)
	if b < a {
		a = b
	}
	if a < unreachablePath {
		a++
	}
	memo[edge] = a
	return a
}

func assessPathLength(counts map[int32]int32, numVars, threshold int32) (int32, int32) {
	temp, maxPath := threshold, int32(0)
	for i := int32(0); i < numVars+1 && temp > 0; i++ {
		if count := counts[i]; count > 0 {
			maxPath = i
			temp -= count
		}
	}
	if temp >= 0 {
		return maxPath + 1, 0
	}
	return maxPath, temp + counts[maxPath]
}

func (s *subsetBuildState) build(edge *DdNode) *DdNode {
	if core.Cudd_IsConstant(edge) {
		if edge == Cudd_ReadOne(s.dd) && s.findShortestPath {
			s.findShortestPath = false
		}
		return edge
	}
	if result, ok := s.results[edge]; ok {
		return result
	}
	children := [2]*DdNode{}
	children[0], children[1] = pathChildren(edge)
	branches := [2]*DdNode{}
	done := [2]bool{}
	processingDone := 0
	for i, child := range children {
		if core.Cudd_IsConstant(child) {
			branches[i], done[i] = child, true
			processingDone++
			if child == Cudd_ReadOne(s.dd) && s.findShortestPath {
				s.findShortestPath = false
			}
		}
	}
	metrics := [2]pathMetric{
		{length: unreachablePath, bottom: unreachablePath},
		{length: unreachablePath, bottom: unreachablePath},
	}
	for i, child := range children {
		if !done[i] {
			metrics[i] = s.metrics[core.Cudd_Regular(child)]
		}
	}
	tiebreakThen := metrics[0].bottom <= metrics[1].bottom
	for processingDone < 2 {
		idx := -1
		if processingDone == 0 {
			if metrics[0].length < metrics[1].length || (metrics[0].length == metrics[1].length && tiebreakThen) {
				idx = 0
			} else {
				idx = 1
			}
		} else if !done[0] {
			idx = 0
		} else {
			idx = 1
		}
		done[idx] = true
		child, metric := children[idx], metrics[idx]
		var branch *DdNode
		switch {
		case metric.length > s.maxPath:
			branch = Cudd_ReadLogicZero(s.dd)
		case metric.length < s.maxPath:
			if s.findShortestPath {
				s.findShortestPath = false
			}
			branch = s.build(child)
		default:
			regChild := core.Cudd_Regular(child)
			if _, selected := s.maxpathTable[regChild]; selected {
				if s.findShortestPath {
					s.findShortestPath = false
				}
				branch = s.build(child)
			} else if s.thresholdReached <= 0 {
				if s.findShortestPath {
					s.maxpathTable[regChild] = struct{}{}
					s.thresholdReached--
					branch = s.build(child)
				} else {
					branch = Cudd_ReadLogicZero(s.dd)
				}
			} else {
				s.maxpathTable[regChild] = struct{}{}
				s.thresholdReached--
				if s.thresholdReached <= 0 {
					s.findShortestPath = true
				}
				branch = s.build(child)
			}
		}
		branches[idx] = branch
		processingDone++
	}
	s.findShortestPath = false
	result := core.Cudd_bddIte(s.dd, core.Cudd_bddIthVar(s.dd, int(core.Cudd_NodeReadIndex(edge))), branches[0], branches[1])
	if s.subsetNodes != nil {
		regular := core.Cudd_Regular(result)
		if !core.Cudd_IsConstant(regular) {
			if _, exists := s.subsetNodes[regular]; !exists {
				s.subsetNodes[regular] = struct{}{}
				if int32(len(s.subsetNodes)) > s.threshold {
					s.thresholdReached = 0
				}
			}
		}
	}
	s.results[edge] = result
	return result
}
