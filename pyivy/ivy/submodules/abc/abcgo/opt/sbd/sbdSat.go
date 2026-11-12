package sbd

import (
	"fmt"
	"slices"
	"time"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
)

const MAX_M = 8
const MAX_N = 30
const MAX_K = 6
const MAX_D = 8

type TopologyVariables [MAX_N][MAX_M + MAX_N][MAX_K]int32
type DelayVariables [MAX_M + MAX_N][MAX_D]int32
type LUTTopology [MAX_N][MAX_K]int32
type Sbd_Pro_t struct {
	nLuts, nSize, nDivs, nVars, nPars int32
	pPars1                            [SBD_LUTS_MAX][1 << SBD_SIZE_MAX]int32
	pPars2                            [SBD_LUTS_MAX][SBD_SIZE_MAX][SBD_DIV_MAX]int32
	pVars                             [SBD_LUTS_MAX][SBD_SIZE_MAX + 1]int32
	pDivs                             [SBD_DIV_MAX]int32
}

func Sbd_ProblemSetup(p *Sbd_Pro_t, luts, size, divs int32) *vec.Vec_Int_t {
	if luts < 1 || luts > 2 || size < 1 || size > 4 || divs < 1 || divs > 10 {
		panic("invalid parameterized LUT problem")
	}
	*p = Sbd_Pro_t{nLuts: luts, nSize: size, nDivs: divs}
	variable := int32(0)
	for i := int32(0); i < luts; i++ {
		for row := int32(0); row < 1<<size; row++ {
			p.pPars1[i][row] = variable
			variable++
		}
	}
	for i := int32(0); i < luts; i++ {
		for k := int32(0); k < size; k++ {
			for d := int32(0); d < divs; d++ {
				p.pPars2[i][k][d] = variable
				variable++
			}
		}
	}
	p.nPars = variable
	for i := int32(0); i < luts; i++ {
		for k := int32(0); k <= size; k++ {
			if i > 0 && k == 0 {
				p.pVars[i][k] = p.pVars[i-1][size]
			} else {
				p.pVars[i][k] = variable
				variable++
			}
		}
	}
	p.nVars = variable - p.nPars
	for d := int32(0); d < divs; d++ {
		p.pDivs[d] = variable
		variable++
	}
	cnf := vec.Vec_IntAlloc(1000)
	add := func(literals ...int32) {
		for _, lit := range literals {
			vec.Vec_IntPush(cnf, lit)
		}
		vec.Vec_IntPush(cnf, -1)
	}
	for i := int32(0); i < luts; i++ {
		first := int32(0)
		if i > 0 {
			first = 1
		}
		for k := first; k < size; k++ {
			var selected []int32
			for d := int32(0); d < divs; d++ {
				selected = append(selected, 2*p.pPars2[i][k][d]+1)
				for n := int32(0); n < divs; n++ {
					if n < d {
						add(2*p.pPars2[i][k][d], 2*p.pPars2[i][k][n])
					} else if k < size-1 {
						add(2*p.pPars2[i][k][d], 2*p.pPars2[i][k+1][n])
					}
				}
			}
			add(selected...)
		}
	}
	for i := int32(0); i < luts; i++ {
		for row := int32(0); row < 1<<size; row++ {
			for sign := int32(0); sign < 2; sign++ {
				var clause []int32
				for k := int32(0); k < size; k++ {
					clause = append(clause, 2*p.pVars[i][k]+(row>>k&1))
				}
				clause = append(clause, 2*p.pVars[i][size]+sign, 2*p.pPars1[i][row]+1-sign)
				add(clause...)
			}
		}
	}
	for i := int32(0); i < luts; i++ {
		first := int32(0)
		if i > 0 {
			first = 1
		}
		for k := first; k < size; k++ {
			for d := int32(0); d < divs; d++ {
				for sign := int32(0); sign < 2; sign++ {
					add(2*p.pPars2[i][k][d], 2*p.pVars[i][k]+sign, 2*p.pDivs[d]+1-sign)
				}
			}
		}
	}
	return cnf
}
func sbdFlatClauses(cnf *vec.Vec_Int_t) [][]int32 {
	var clauses [][]int32
	start := 0
	for i, v := range vec.Vec_IntArray(cnf) {
		if v == -1 {
			clauses = append(clauses, slices.Clone(vec.Vec_IntArray(cnf)[start:i]))
			start = i + 1
		} else if v < 0 {
			panic("invalid CNF literal")
		}
	}
	if start != vec.Vec_IntSize(cnf) {
		panic("unterminated CNF clause")
	}
	return clauses
}
func Sbd_ProblemLoad1(p *Sbd_Pro_t, cnf *vec.Vec_Int_t, start int32, divVars []int32, top int32, solver *bsat.Solver) {
	root := p.pVars[p.nLuts-1][p.nSize]
	firstDiv := p.nPars + p.nVars
	for _, clause := range sbdFlatClauses(cnf) {
		mapped := make([]bsat.Literal, len(clause))
		for i, lit := range clause {
			v := lit >> 1
			switch {
			case v == root:
				v = top
			case v >= firstDiv:
				v = divVars[v-firstDiv]
			default:
				v += start
			}
			mapped[i] = bsat.Literal(2*v + (lit & 1))
		}
		if !solver.AddClause(mapped...) {
			panic("inconsistent parameterized problem")
		}
	}
}
func Sbd_ProblemLoad2(p *Sbd_Pro_t, cnf *vec.Vec_Wec_t, start int32, divValues []int32, topValue int32, solver *bsat.Solver) {
	root := p.pVars[p.nLuts-1][p.nSize]
	firstDiv := p.nPars + p.nVars
	for level := 0; level < vec.Vec_WecSize(cnf); level++ {
		var mapped []bsat.Literal
		satisfied := false
		for _, lit := range vec.Vec_IntArray(vec.Vec_WecEntry(cnf, level)) {
			v := lit >> 1
			if v == root || v >= firstDiv {
				value := topValue
				if v >= firstDiv {
					value = divValues[v-firstDiv]
				}
				if value != (lit & 1) {
					satisfied = true
					break
				}
				continue
			}
			if v >= p.nPars {
				v += start
			}
			mapped = append(mapped, bsat.Literal(2*v+(lit&1)))
		}
		if !satisfied {
			solver.AddClause(mapped...)
		}
	}
}
func Sbd_SolverTopo(m, n, k int32, variables *TopologyVariables, delays *DelayVariables, inputDelays []int32, required int32) (*bsat.Solver, int32) {
	if m < 1 || m > MAX_M || n < 1 || n > MAX_N || k < 2 || k > MAX_K || required < 0 || required >= MAX_D-1 || len(inputDelays) < int(m) {
		panic("invalid SAT topology dimensions")
	}
	count := int32(0)
	for node := int32(0); node < n; node++ {
		for fan := int32(0); fan < m+n; fan++ {
			for pin := int32(0); pin < k; pin++ {
				variables[node][fan][pin] = -1
			}
		}
		for fan := int32(0); fan < m+node; fan++ {
			for pin := int32(0); pin < k; pin++ {
				variables[node][fan][pin] = count
				count++
			}
		}
	}
	topologyCount := count
	solver := bsat.NewSolver()
	add := func(literals ...bsat.Literal) { solver.AddClause(literals...) }
	for fan := int32(0); fan < m+n-1; fan++ {
		var clause []bsat.Literal
		for node := int32(0); node < n; node++ {
			for pin := int32(0); pin < k; pin++ {
				if v := variables[node][fan][pin]; v >= 0 {
					clause = append(clause, bsat.Positive(v))
				}
			}
		}
		add(clause...)
	}
	for node := int32(0); node < n; node++ {
		for pin := int32(0); pin < k; pin++ {
			var clause []bsat.Literal
			for fan := int32(0); fan < m+node; fan++ {
				clause = append(clause, bsat.Positive(variables[node][fan][pin]))
				for other := fan + 1; other < m+node; other++ {
					add(bsat.Negative(variables[node][fan][pin]), bsat.Negative(variables[node][other][pin]))
				}
			}
			add(clause...)
		}
	}
	for node := int32(0); node < n; node++ {
		for fan := int32(0); fan < m+node; fan++ {
			for pin := int32(0); pin < k; pin++ {
				for other := fan; other < m+node; other++ {
					for pin2 := pin + 1; pin2 < k; pin2++ {
						add(bsat.Negative(variables[node][fan][pin]), bsat.Negative(variables[node][other][pin2]))
					}
				}
			}
		}
	}
	for node := int32(1); node < n; node++ {
		for fan := int32(0); fan < m+node-1; fan++ {
			for other := fan + 1; other < m+node-1; other++ {
				add(bsat.Negative(variables[node][fan][0]), bsat.Negative(variables[node-1][other][0]))
			}
			for second := int32(0); second < fan; second++ {
				for previous := second + 1; previous < fan; previous++ {
					add(bsat.Negative(variables[node][fan][0]), bsat.Negative(variables[node-1][fan][0]), bsat.Negative(variables[node][second][1]), bsat.Negative(variables[node-1][previous][1]))
				}
			}
		}
	}
	if k == 2 {
		for node := int32(1); node < n; node++ {
			for fan := m; fan < m+node; fan++ {
				for other := int32(0); other < fan; other++ {
					for pin := int32(0); pin < k; pin++ {
						v := variables[fan-m][other][pin]
						if v >= 0 {
							add(bsat.Negative(variables[node][fan][0]), bsat.Negative(variables[node][other][1]), bsat.Negative(v))
						}
					}
				}
			}
		}
	}
	for id := int32(0); id < m+n; id++ {
		for delay := int32(0); delay < MAX_D; delay++ {
			delays[id][delay] = count
			count++
		}
	}
	sbdEnsureVariables(solver, count)
	for input := int32(0); input < m; input++ {
		if inputDelays[input] < 0 || inputDelays[input] >= MAX_D-2 {
			panic("input delay exceeds topology encoding")
		}
		add(bsat.Positive(delays[input][inputDelays[input]]))
	}
	for delay := required + 1; delay < MAX_D; delay++ {
		add(bsat.Negative(delays[m+n-1][delay]))
	}
	for node := int32(0); node < n; node++ {
		for fan := int32(0); fan < m+node; fan++ {
			for pin := int32(0); pin < k; pin++ {
				for delay := int32(0); delay < MAX_D-1; delay++ {
					add(bsat.Negative(variables[node][fan][pin]), bsat.Negative(delays[fan][delay]), bsat.Positive(delays[m+node][delay+1]))
				}
			}
		}
	}
	return solver, topologyCount
}
func Sbd_SolverTopoPrint(solver *bsat.Solver, m, n, k int32, variables *TopologyVariables) {
	fmt.Println("Solution:")
	for node := int32(0); node < n; node++ {
		fmt.Printf("%d :", m+node)
		for pin := int32(0); pin < k; pin++ {
			for fan := int32(0); fan < m+node; fan++ {
				if sbdModelValue(solver, variables[node][fan][pin]) != 0 {
					fmt.Printf(" %d", fan)
				}
			}
		}
		fmt.Println()
	}
}
func Sbd_SolverTopoTest() {
	var variables TopologyVariables
	var delayVariables DelayVariables
	solver, count := Sbd_SolverTopo(8, 3, 4, &variables, &delayVariables, []int32{1, 0, 0, 0, 1, 0, 0, 0}, 2)
	solutions := 0
	for iteration := 0; iteration < 1000000 && solver.Solve() == bsat.Satisfiable; iteration++ {
		solutions++
		if iteration < 5 {
			Sbd_SolverTopoPrint(solver, 8, 3, 4, &variables)
		}
		var block []bsat.Literal
		for v := int32(0); v < count; v++ {
			if sbdModelValue(solver, v) != 0 {
				block = append(block, bsat.Negative(v))
			}
		}
		if !solver.AddClause(block...) {
			break
		}
	}
	fmt.Printf("Found %d solutions.\n", solutions)
}
func Sbd_SolverSynth(m, n, k int32, luts *LUTTopology) {
	if m < 2 || m > MAX_M || n < m-1 || n > MAX_N || k != 2 {
		panic("random topology generator requires feasible binary dimensions")
	}
	random := aig.NewAigRandomSource()
	var used [MAX_M + MAX_N]bool
	unused := m
	for node := int32(0); node < n; node++ {
		remaining := n - node
		var a, b int32
		if unused == remaining+1 {
			found := false
			for fan := int32(0); fan < m+node; fan++ {
				if used[fan] {
					continue
				}
				if !found {
					a = fan
					found = true
				} else {
					b = fan
					break
				}
			}
		} else {
			for {
				a = int32(random.Uint32() % uint32(m+node))
				b = int32(random.Uint32() % uint32(m+node))
				if a == b {
					continue
				}
				consumed := int32(0)
				if !used[a] {
					consumed++
				}
				if !used[b] {
					consumed++
				}
				if unused+1-consumed <= remaining {
					break
				}
			}
		}
		luts[node][0], luts[node][1] = a, b
		for _, fan := range []int32{a, b} {
			if !used[fan] {
				used[fan] = true
				unused--
			}
		}
		unused++
	}
}
func sbdCheckTopology(m, n, k int32, luts *LUTTopology) {
	if m < 1 || m > 8 || n < 1 || n > 30 || k < 1 || k > 6 {
		panic("invalid LUT topology")
	}
	for node := int32(0); node < n; node++ {
		for _, fan := range luts[node][:k] {
			if fan < 0 || fan >= m+node {
				panic("non-topological LUT fanin")
			}
		}
	}
}
func Sbd_SolverTruth(m, n, k int32, luts *LUTTopology, values []int32) uint64 {
	if m > 6 {
		panic("word truth exceeds six inputs")
	}
	storage := sbdTruthStorage(m, n)
	return Sbd_SolverTruthWord(m, n, k, luts, values, storage, 0)[0]
}
func sbdTruthStorage(m, n int32) []uint64 {
	words := 1
	if m > 6 {
		words <<= m - 6
	}
	storage := make([]uint64, int(m+n+1)*words)
	for v := int32(0); v < m; v++ {
		for bit := 0; bit < 64*words; bit++ {
			if bit>>v&1 != 0 {
				storage[int(v)*words+bit/64] |= uint64(1) << uint(bit&63)
			}
		}
	}
	return storage
}
func Sbd_SolverTruthWord(m, n, k int32, luts *LUTTopology, values []int32, storage []uint64, compl int32) []uint64 {
	sbdCheckTopology(m, n, k, luts)
	words := 1
	if m > 6 {
		words <<= m - 6
	}
	storage = storage[:int(m+n+1)*words]
	parameters := int32(1<<k) - 1
	for node := int32(0); node < n; node++ {
		truth := storage[int(m+node)*words : int(m+node+1)*words]
		clear(truth)
		for row := int32(1); row <= parameters; row++ {
			if values[node*parameters+row-1] == 0 {
				continue
			}
			for w := 0; w < words; w++ {
				cube := ^uint64(0)
				for pin, fan := range luts[node][:k] {
					value := storage[int(fan)*words+w]
					if row>>pin&1 == 0 {
						value = ^value
					}
					cube &= value
				}
				truth[w] |= cube
			}
		}
	}
	result := storage[int(m+n-1)*words : int(m+n)*words]
	if compl != 0 {
		for i := range result {
			result[i] = ^result[i]
		}
	}
	return result
}
func Sbd_SolverFunc(m, n, k int32, luts *LUTTopology, target []uint64, values []int32) int32 {
	sbdCheckTopology(m, n, k, luts)
	storage := sbdTruthStorage(m, n)
	words := 1
	if m > 6 {
		words <<= m - 6
	}
	truth := slices.Clone(target[:words])
	if m < 6 {
		mask := uint64(1)<<uint(1<<m) - 1
		pattern := truth[0] & mask
		truth[0] = 0
		for shift := 0; shift < 64; shift += 1 << m {
			truth[0] |= pattern << shift
		}
	}
	phase := int32(truth[0] & 1)
	parameters := int32(1<<k) - 1
	solver := bsat.NewSolver()
	sbdEnsureVariables(solver, n*parameters)
	minterm := int32(1)
	for iteration := 0; iteration < 1<<m; iteration++ {
		start := solver.NumVars()
		sbdEnsureVariables(solver, start+n-1)
		for node := int32(0); node < n; node++ {
			for row := int32(0); row <= parameters; row++ {
				var mismatch []bsat.Literal
				possible := true
				for pin, fan := range luts[node][:k] {
					if fan < m {
						if minterm>>fan&1 != row>>pin&1 {
							possible = false
							break
						}
					} else {
						mismatch = append(mismatch, bsat.Literal(2*(start+fan-m)+(row>>pin&1)))
					}
				}
				if !possible {
					continue
				}
				if node == n-1 {
					desired := int32(truth[minterm>>6]>>uint(minterm&63)&1) ^ phase
					if row == 0 {
						if desired != 0 {
							solver.AddClause(mismatch...)
						}
					} else {
						sign := 1 - desired
						solver.AddClause(append(mismatch, bsat.Literal(2*(node*parameters+row-1)+sign))...)
					}
				} else {
					output := bsat.Positive(start + node)
					if row == 0 {
						solver.AddClause(append(mismatch, output^1)...)
					} else {
						parameter := bsat.Positive(node*parameters + row - 1)
						solver.AddClause(append(slices.Clone(mismatch), parameter^1, output)...)
						solver.AddClause(append(mismatch, parameter, output^1)...)
					}
				}
			}
		}
		if solver.Solve() != bsat.Satisfiable {
			return -1
		}
		for v := int32(0); v < n*parameters; v++ {
			values[v] = sbdModelValue(solver, v)
		}
		candidate := Sbd_SolverTruthWord(m, n, k, luts, values, storage, phase)
		if slices.Equal(candidate, truth) {
			return phase
		}
		minterm = -1
		for bit := int32(0); bit < 1<<m; bit++ {
			if candidate[bit>>6]>>uint(bit&63)&1 != truth[bit>>6]>>uint(bit&63)&1 {
				minterm = bit
				break
			}
		}
		if minterm < 0 {
			panic("truth mismatch outside declared inputs")
		}
	}
	return -1
}
func Sbd_SolverFuncTest() {
	var luts LUTTopology
	for i, pair := range [][2]int32{{0, 1}, {2, 3}, {4, 5}, {6, 7}, {8, 9}, {10, 11}, {12, 13}} {
		luts[i][0], luts[i][1] = pair[0], pair[1]
	}
	truth := []uint64{0x0000080000020000, 0x0000000000020000, 0, 0x0000000000020000}
	values := make([]int32, 7*3)
	started := time.Now()
	phase := Sbd_SolverFunc(8, 7, 2, &luts, truth, values)
	fmt.Printf("Result (compl=%d): %v Time=%v\n", phase, values, time.Since(started))
}
