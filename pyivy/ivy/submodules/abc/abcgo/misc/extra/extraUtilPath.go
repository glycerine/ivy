package extra

// Frontier enumeration of simple corner-to-corner grid paths. State entries
// pack degree in the high 16 bits and component representative in the low 16.
import (
	"encoding/binary"
	"fmt"
	"math"
	"slices"
	"time"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/ioa"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
)

func Abc_NodeVarX(size, y, x int32) int32 { return 2 * (size*y + x) }
func Abc_NodeVarY(size, y, x int32) int32 { return 2 * (size*(size+1) + size*x + y) }
func Abc_EnumeratePaths(size int32) *aig.Graph {
	if size < 1 || size > 1000 {
		panic("invalid directed grid size")
	}
	graph := aig.NewGraph()
	inputs := make([]aig.Edge, 2*size*(size+1))
	for i := range inputs {
		inputs[i] = graph.NewInput()
	}
	// Original literals number CI objects starting at one; X/Y return those
	// object literals, so convert to the corresponding zero-based input index.
	edge := func(lit int32) aig.Edge { return inputs[lit/2-1] }
	nodes := make([]aig.Edge, size+1)
	nodes[0] = graph.Const1()
	for x := int32(1); x <= size; x++ {
		nodes[x] = graph.AND(nodes[x-1], edge(Abc_NodeVarX(size, 0, x)))
	}
	for y := int32(1); y <= size; y++ {
		nodes[0] = graph.AND(nodes[0], edge(Abc_NodeVarY(size, y, 0)))
		for x := int32(1); x <= size; x++ {
			horizontal := graph.AND(nodes[x-1], edge(Abc_NodeVarX(size, y, x)))
			vertical := graph.AND(nodes[x], edge(Abc_NodeVarY(size, y, x)))
			nodes[x] = graph.OR(horizontal, vertical)
		}
	}
	graph.AddOutput(nodes[size])
	aig.Aig_ManCleanup(graph)
	return graph
}
func Abc_EnumeratePathsTest() error {
	model, err := Abc_EnumeratePaths(2).AigerModel()
	if err != nil {
		return err
	}
	return ioa.Ioa_WriteAiger(model, "testpath.aig", 0, 0)
}
func Abc_GraphGrid(n int32) *vec.Vec_Int_t {
	if n < 2 || n > 256 {
		panic("frontier grid requires 2..256 vertices per side")
	}
	edges := vec.Vec_IntAlloc(int(4 * n * (n - 1)))
	for row := int32(0); row < n; row++ {
		for x := int32(0); x < n-1; x++ {
			vec.Vec_IntPush(edges, row*n+x)
			vec.Vec_IntPush(edges, row*n+x+1)
		}
		if row == n-1 {
			break
		}
		for x := int32(0); x < n; x++ {
			vec.Vec_IntPush(edges, row*n+x)
			vec.Vec_IntPush(edges, row*n+x+n)
		}
	}
	return edges
}
func Abc_GraphNodeLife(edges *vec.Vec_Int_t, n int32) *vec.Vec_Int_t {
	if n < 2 || n > 256 || vec.Vec_IntSize(edges)%2 != 0 {
		panic("invalid grid lifetimes")
	}
	life := vec.Vec_IntStartFull(int(2 * n * n))
	for e := 0; e < vec.Vec_IntSize(edges)/2; e++ {
		for _, node := range vec.Vec_IntArray(edges)[2*e : 2*e+2] {
			if node < 0 || node >= n*n {
				panic("invalid grid vertex")
			}
			if vec.Vec_IntEntry(life, int(2*node)) == -1 {
				vec.Vec_IntWriteEntry(life, int(2*node), int32(e))
			}
			vec.Vec_IntWriteEntry(life, int(2*node+1), int32(e))
		}
	}
	return life
}
func Abc_GraphFrontiers(edges, life *vec.Vec_Int_t) *vec.Vec_Wec_t {
	fronts := vec.Vec_WecAlloc(vec.Vec_IntSize(edges)/2 + 1)
	vec.Vec_WecPushLevel(fronts)
	var current []int32
	for e := 0; e < vec.Vec_IntSize(edges)/2; e++ {
		for _, node := range vec.Vec_IntArray(edges)[2*e : 2*e+2] {
			if vec.Vec_IntEntry(life, int(2*node)) == int32(e) {
				current = append(current, node)
			} else if vec.Vec_IntEntry(life, int(2*node+1)) == int32(e) {
				at := slices.Index(current, node)
				if at < 0 {
					panic("missing frontier vertex")
				}
				current = append(current[:at], current[at+1:]...)
			}
		}
		level := vec.Vec_WecPushLevel(fronts)
		for _, node := range current {
			vec.Vec_IntPush(level, node)
		}
	}
	return fronts
}
func graphPathPrint(buffer []int32, edges *vec.Vec_Int_t, size int32) {
	width := int(4*(size-1) + 1)
	box := make([][]byte, width)
	for y := range box {
		box[y] = make([]byte, width)
		for x := range box[y] {
			box[y][x] = ' '
			if y%4 == 0 && x%4 == 0 {
				box[y][x] = '*'
			}
		}
	}
	for e := 0; e < vec.Vec_IntSize(edges)/2; e++ {
		if buffer[e] == 0 {
			continue
		}
		a, b := vec.Vec_IntEntry(edges, 2*e), vec.Vec_IntEntry(edges, 2*e+1)
		y0, x0, y1, x1 := 4*(a/size), 4*(a%size), 4*(b/size), 4*(b%size)
		if y0 == y1 {
			for x := x0 + 1; x < x1; x++ {
				box[y0][x] = '-'
			}
		} else if x0 == x1 {
			for y := y0 + 1; y < y1; y++ {
				box[y][x0] = '|'
			}
		} else {
			panic("non-grid path edge")
		}
	}
	for _, row := range box {
		fmt.Println(string(row))
	}
	fmt.Print("\n\n=================================\n\n")
}
func Abc_GraphPathPrint4(buffer []int32, edges *vec.Vec_Int_t) { graphPathPrint(buffer, edges, 4) }
func Abc_GraphPathPrint5(buffer []int32, edges *vec.Vec_Int_t) { graphPathPrint(buffer, edges, 5) }
func Abc_GraphCountPaths_rec(level, node int32, nodes *vec.Vec_Wec_t, cache [][]float64, buffer []int32, edges *vec.Vec_Int_t) float64 {
	if node == -2 {
		return 1
	}
	if node == -1 {
		return 0
	}
	if cache[level][node] != -1 {
		return cache[level][node]
	}
	buffer[level] = 0
	low := Abc_GraphCountPaths_rec(level+1, vec.Vec_WecEntryEntry(nodes, int(level), int(2*node)), nodes, cache, buffer, edges)
	buffer[level] = 1
	high := Abc_GraphCountPaths_rec(level+1, vec.Vec_WecEntryEntry(nodes, int(level), int(2*node+1)), nodes, cache, buffer, edges)
	cache[level][node] = low + high
	return low + high
}
func Abc_GraphCountPaths(nodes *vec.Vec_Wec_t, edges *vec.Vec_Int_t) float64 {
	cache := make([][]float64, vec.Vec_WecSize(nodes))
	for level := range cache {
		cache[level] = make([]float64, vec.Vec_IntSize(vec.Vec_WecEntry(nodes, level))/2)
		for i := range cache[level] {
			cache[level][i] = -1
		}
	}
	return Abc_GraphCountPaths_rec(0, 0, nodes, cache, make([]int32, len(cache)), edges)
}
func Abc_GraphDeriveGia_rec(graph *aig.Graph, level, node int32, nodes *vec.Vec_Wec_t, cache [][]aig.Edge, buffer []int32, edges *vec.Vec_Int_t) aig.Edge {
	if node == -2 {
		return graph.Const1()
	}
	if node == -1 {
		return graph.Const0()
	}
	if cache[level][node] != aig.InvalidEdge {
		return cache[level][node]
	}
	buffer[level] = 0
	low := Abc_GraphDeriveGia_rec(graph, level+1, vec.Vec_WecEntryEntry(nodes, int(level), int(2*node)), nodes, cache, buffer, edges)
	buffer[level] = 1
	high := Abc_GraphDeriveGia_rec(graph, level+1, vec.Vec_WecEntryEntry(nodes, int(level), int(2*node+1)), nodes, cache, buffer, edges)
	result := graph.Mux(graph.Inputs()[level], high, low)
	cache[level][node] = result
	return result
}
func Abc_GraphDeriveGia(nodes *vec.Vec_Wec_t, edges *vec.Vec_Int_t) *aig.Graph {
	graph := aig.NewGraph()
	graph.Name = "paths"
	for e := 0; e < vec.Vec_IntSize(edges)/2; e++ {
		graph.NewInput()
	}
	cache := make([][]aig.Edge, vec.Vec_WecSize(nodes))
	for level := range cache {
		cache[level] = make([]aig.Edge, vec.Vec_IntSize(vec.Vec_WecEntry(nodes, level))/2)
		for i := range cache[level] {
			cache[level][i] = aig.InvalidEdge
		}
	}
	graph.AddOutput(Abc_GraphDeriveGia_rec(graph, 0, 0, nodes, cache, make([]int32, len(cache)), edges))
	aig.Aig_ManCleanup(graph)
	return graph
}
func Abc_GraphDeriveGiaDump(nodes *vec.Vec_Wec_t, edges *vec.Vec_Int_t, size int32) error {
	graph := Abc_GraphDeriveGia(nodes, edges)
	model, err := graph.AigerModel()
	if err != nil {
		return err
	}
	name := fmt.Sprintf("grid_%dx%d_e%03d.aig", size, size, vec.Vec_IntSize(edges)/2)
	if err = ioa.Ioa_WriteAiger(model, name, 0, 0); err != nil {
		return err
	}
	fmt.Printf("Finished dumping AIG into file %q.\n", name)
	return nil
}
func Abc_GraphBuildState(state *vec.Vec_Int_t, e, x int32, edges, life *vec.Vec_Int_t, fronts *vec.Vec_Wec_t, front []int32, next *vec.Vec_Int_t, verbose int32) int32 {
	before, after := vec.Vec_IntArray(vec.Vec_WecEntry(fronts, int(e))), vec.Vec_IntArray(vec.Vec_WecEntry(fronts, int(e+1)))
	pair := vec.Vec_IntArray(edges)[2*e : 2*e+2]
	count := int32(vec.Vec_IntSize(life) / 2)
	if pair[0] >= pair[1] || x < 0 || x > 1 || len(before) != vec.Vec_IntSize(state) {
		panic("invalid frontier state")
	}
	for i, node := range before {
		front[node] = vec.Vec_IntEntry(state, i)
	}
	for _, node := range pair {
		if vec.Vec_IntEntry(life, int(2*node)) == e {
			front[node] = node
		}
	}
	if x != 0 {
		if front[pair[0]]&65535 == front[pair[1]]&65535 {
			return -1
		}
		for _, node := range pair {
			degree := front[node] >> 16
			limit := int32(2)
			if node == 0 || node == count-1 {
				limit = 1
			}
			if degree >= limit {
				return -1
			}
			front[node] += 1 << 16
		}
	}
	equiv := [2]int32{front[pair[0]] & 65535, front[pair[1]] & 65535}
	for n, node := range pair {
		if vec.Vec_IntEntry(life, int(2*node+1)) != e {
			continue
		}
		degree := front[node] >> 16
		if node == 0 || node == count-1 {
			if degree != 1 {
				return -1
			}
		} else if degree != 0 && degree != 2 {
			return -1
		}
		first := int32(-1)
		for _, other := range after {
			if other == node {
				panic("departing node retained in frontier")
			}
			if front[other]&65535 == equiv[n] {
				if first == -1 {
					first = other
				}
				front[other] = front[other]&^65535 | first
			}
		}
		if first != -1 {
			equiv[n] = first
		}
	}
	if x != 0 {
		first := int32(-1)
		for _, node := range after {
			if front[node]&65535 == equiv[0] || front[node]&65535 == equiv[1] {
				if first == -1 {
					first = node
				}
				front[node] = front[node]&^65535 | first
			}
		}
	}
	vec.Vec_IntClear(next)
	for _, node := range after {
		vec.Vec_IntPush(next, front[node])
	}
	if verbose != 0 {
		fmt.Printf("Edge = %d. Arc = %d. Next state: %v\n", e, x, vec.Vec_IntArray(next))
	}
	return 1
}
func graphStateKey(state *vec.Vec_Int_t) string {
	data := make([]byte, 4*vec.Vec_IntSize(state))
	for i, v := range vec.Vec_IntArray(state) {
		binary.LittleEndian.PutUint32(data[4*i:], uint32(v))
	}
	return string(data)
}
func graphBuildFrontier(edges, life *vec.Vec_Int_t, fronts *vec.Vec_Wec_t, verbose int32) (*vec.Vec_Wec_t, []int32) {
	edgeCount, nodeCount := vec.Vec_IntSize(edges)/2, vec.Vec_IntSize(life)/2
	nodes := vec.Vec_WecAlloc(edgeCount)
	state := vec.Vec_IntAlloc(nodeCount)
	front := make([]int32, nodeCount)
	current := []*vec.Vec_Int_t{vec.Vec_IntAlloc(0)}
	counts := make([]int32, edgeCount)
	for e := 0; e < edgeCount; e++ {
		level := vec.Vec_WecPushLevel(nodes)
		counts[e] = int32(len(current))
		var following []*vec.Vec_Int_t
		indices := map[string]int32{}
		for _, s := range current {
			for x := int32(0); x < 2; x++ {
				result := Abc_GraphBuildState(s, int32(e), x, edges, life, fronts, front, state, verbose)
				if result == 1 {
					if e == edgeCount-1 {
						result = -2
					} else {
						key := graphStateKey(state)
						var ok bool
						result, ok = indices[key]
						if !ok {
							result = int32(len(following))
							indices[key] = result
							copyState := vec.Vec_IntAlloc(vec.Vec_IntSize(state))
							for _, v := range vec.Vec_IntArray(state) {
								vec.Vec_IntPush(copyState, v)
							}
							following = append(following, copyState)
						}
					}
				}
				vec.Vec_IntPush(level, result)
			}
		}
		current = following
	}
	return nodes, counts
}
func Abc_GraphBuildFrontier(size int32, edges, life *vec.Vec_Int_t, fronts *vec.Vec_Wec_t, verbose int32) error {
	started := time.Now()
	nodes, counts := graphBuildFrontier(edges, life, fronts, verbose)
	total := int64(0)
	for _, c := range counts {
		total += int64(c)
	}
	fmt.Printf("States = %8d   Paths = %24.0f  Time = %v\n", total, Abc_GraphCountPaths(nodes, edges), time.Since(started))
	return Abc_GraphDeriveGiaDump(nodes, edges, size)
}
func Abc_EnumerateFrontierTest(size int32) error {
	edges := Abc_GraphGrid(size)
	life := Abc_GraphNodeLife(edges, size)
	fronts := Abc_GraphFrontiers(edges, life)
	return Abc_GraphBuildFrontier(size, edges, life, fronts, 0)
}
func Abc_Word2Double(w uint64) float64 { return float64(w) }

// Tseitin clauses preserve the graph's logic while avoiding the unrelated
// mapping/CNF managers used only to feed the original lexicographic diagnostic.
func graphPathSolver(graph *aig.Graph) (*bsat.Solver, []bsat.Literal) {
	if err := graph.Check(); err != nil {
		panic(err)
	}
	solver := bsat.NewSolver()
	for range graph.Nodes() {
		solver.NewVariable()
	}
	literal := func(edge aig.Edge) bsat.Literal { return bsat.Literal(edge) }
	solver.AddClause(bsat.Positive(0)) // aig.Graph's regular constant is true.
	for _, edge := range graph.LiveEdges() {
		node := graph.Node(edge)
		y := literal(edge)
		switch node.Kind {
		case aig.And:
			a, b := literal(node.Fanins[0]), literal(node.Fanins[1])
			solver.AddClause(y^1, a)
			solver.AddClause(y^1, b)
			solver.AddClause(y, a^1, b^1)
		case aig.ExclusiveOr:
			a, b := literal(node.Fanins[0]), literal(node.Fanins[1])
			solver.AddClause(a, b, y^1)
			solver.AddClause(a^1, b^1, y^1)
			solver.AddClause(a, b^1, y)
			solver.AddClause(a^1, b, y)
		case aig.Buffer, aig.Output:
			a := literal(node.Fanins[0])
			solver.AddClause(y^1, a)
			solver.AddClause(y, a^1)
		}
	}
	for _, edge := range graph.Outputs() {
		solver.AddClause(literal(edge))
	}
	inputs := graph.Inputs()
	assignment := make([]bsat.Literal, len(inputs))
	for i, edge := range inputs {
		assignment[i] = literal(edge) ^ 1
	}
	return solver, assignment
}

type graphPathEstimate struct {
	Samples           int
	First, Last       uint64
	Average, Estimate float64
}

func graphSolvePaths(graph *aig.Graph, limit int) graphPathEstimate {
	if graph.InputCount() < 1 || graph.InputCount() > 64 || limit < 1 {
		panic("invalid lexicographic path sample dimensions")
	}
	solver, literals := graphPathSolver(graph)
	result := graphPathEstimate{}
	for iteration := 0; iteration < limit; iteration++ {
		if solver.SolveLexicographic(literals) != bsat.Satisfiable {
			break
		}
		var minterm uint64
		for k, lit := range literals {
			if lit&1 == 0 {
				minterm |= uint64(1) << uint(len(literals)-1-k)
			}
		}
		if result.Samples == 0 {
			result.First = minterm
		}
		result.Last = minterm
		result.Samples++
		block := make([]bsat.Literal, len(literals))
		for i, lit := range literals {
			block[i] = lit ^ 1
		}
		if !solver.AddClause(block...) {
			break
		}
	}
	if result.Samples > 1 {
		result.Average = float64(result.Last-result.First) / float64(result.Samples-1)
		if result.Average > 0 {
			result.Estimate = (math.Ldexp(1, graph.InputCount()) - float64(result.First)) / result.Average
		}
	}
	return result
}
func Abc_GraphSolve(graph *aig.Graph) {
	result := graphSolvePaths(graph, 1000)
	fmt.Printf("Vars = %d   Iters = %d   Ave = %.0f   Total = %.0f  Estimate = %.0f\n", graph.InputCount(), result.Samples, float64(result.Last-result.First), result.Average, result.Estimate)
}
