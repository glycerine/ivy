package bmc

import (
	"fmt"
	"math/bits"
	"os"
	"strings"
	"time"

	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/sat/bsat"
	"github.com/glycerine/abcgo/sat/glucose"
)

// zyxSolver uses the same incremental Glucose backend as bmcg in ABC.
type zyxSolver struct{ s *glucose.Solver }

func (z *zyxSolver) NewVariable() int32 { return int32(z.s.NewVar(true, true)) }
func (z *zyxSolver) AddClause(lits ...bsat.Literal) bool {
	clause := make([]glucose.Lit, len(lits))
	for i, l := range lits {
		clause[i] = glucose.Lit(l)
	}
	return z.s.AddClause(clause)
}
func (z *zyxSolver) Solve() bsat.Status {
	switch z.s.SolveLimited(nil) {
	case glucose.LTrue:
		return bsat.Satisfiable
	case glucose.LFalse:
		return bsat.Unsatisfiable
	default:
		return bsat.Unknown
	}
}
func (z *zyxSolver) Value(v int32) (bool, bool) {
	if v < 0 || int(v) >= len(z.s.Model) {
		return false, false
	}
	value := z.s.Model[v]
	return value == glucose.LTrue, value != glucose.LUndef
}
func (z *zyxSolver) SetDeadline(t time.Time) { z.s.Deadline = t }

// zyxNetwork holds the topology and truth table of an exact synthesis result.
type zyxNetwork struct {
	fanins       [][]int
	funcs        [][]bool
	complemented bool
}

// zyxSynthesis encodes the same fanin selection and minterm semantics as the
// original lazy encoding. Clauses for minterms are installed on demand.
type zyxSynthesis struct {
	pars        Bmc_EsPar_t
	solver      *zyxSolver
	topo        [][]int32
	funcs       [][]int32
	values      [][]int32
	target      []uint64
	inverted    bool
	combos      [][][]int
	constrained []bool
}

func zyxSelect(n, k int) [][]int {
	var out [][]int
	var visit func(int, []int)
	visit = func(next int, chosen []int) {
		if len(chosen) == k {
			out = append(out, append([]int(nil), chosen...))
			return
		}
		for i := next; i <= n-(k-len(chosen)); i++ {
			visit(i+1, append(chosen, i))
		}
	}
	visit(0, nil)
	return out
}
func (p *zyxSynthesis) clause(lits ...bsat.Literal) bool { return p.solver.AddClause(lits...) }
func (p *zyxSynthesis) set(v int32, val bool) bool {
	if val {
		return p.clause(bsat.Positive(v))
	}
	return p.clause(bsat.Negative(v))
}
func newZyxSynthesis(pars *Bmc_EsPar_t) (*zyxSynthesis, error) {
	if pars == nil || pars.nVars < 2 || pars.nVars > 10 || pars.nNodes < 1 || pars.nLutSize < 2 || pars.nLutSize > 6 || pars.nLutSize > pars.nVars || pars.nVars+pars.nNodes > 26 {
		return nil, fmt.Errorf("invalid synthesis dimensions")
	}
	if pars.fMajority != 0 && pars.nLutSize != 3 {
		return nil, fmt.Errorf("majority gates require three fanins")
	}
	n, nodeCount, k := int(pars.nVars), int(pars.nNodes), int(pars.nLutSize)
	total := n + nodeCount
	minterms := 1 << n
	functions := 1 << k
	p := &zyxSynthesis{pars: *pars, solver: &zyxSolver{s: glucose.Glucose_solver_start()}, target: make([]uint64, util.Abc_TtWordNum(pars.nVars)), topo: make([][]int32, nodeCount), funcs: make([][]int32, nodeCount), values: make([][]int32, minterms), combos: make([][][]int, nodeCount), constrained: make([]bool, minterms)}
	if pars.fMajority != 0 {
		for m := 0; m < minterms; m++ {
			if bits.OnesCount(uint(m)) > n/2 {
				util.Abc_TtSetBit(p.target, int32(m))
			}
		}
	} else {
		if pars.pTtStr == "" {
			return nil, fmt.Errorf("missing truth table")
		}
		util.Abc_TtReadHex(p.target, pars.pTtStr)
		if p.target[0]&1 != 0 {
			p.inverted = true
			for i := range p.target {
				p.target[i] = ^p.target[i]
			}
		}
	}
	for i := 0; i < nodeCount; i++ {
		p.topo[i] = make([]int32, n+i)
		for j := range p.topo[i] {
			p.topo[i][j] = p.solver.NewVariable()
		}
		p.funcs[i] = make([]int32, functions)
		for j := range p.funcs[i] {
			p.funcs[i][j] = p.solver.NewVariable()
		}
		p.combos[i] = zyxSelect(n+i, k)
	}
	for m := 0; m < minterms; m++ {
		p.values[m] = make([]int32, total)
		for i := range p.values[m] {
			p.values[m][i] = p.solver.NewVariable()
		}
		for i := 0; i < n; i++ {
			p.set(p.values[m][i], m&(1<<i) != 0)
		}
		p.set(p.values[m][total-1], util.Abc_TtGetBit(p.target, int32(m)) != 0)
	}
	// Every gate has exactly k distinct earlier fanins.
	for i := 0; i < nodeCount; i++ {
		fans := p.topo[i]
		for _, subset := range zyxSelect(len(fans), k+1) {
			cl := make([]bsat.Literal, 0, len(subset))
			for _, j := range subset {
				cl = append(cl, bsat.Negative(fans[j]))
			}
			p.clause(cl...)
		}
		for _, subset := range zyxSelect(len(fans), len(fans)-k+1) {
			cl := make([]bsat.Literal, 0, len(subset))
			for _, j := range subset {
				cl = append(cl, bsat.Positive(fans[j]))
			}
			p.clause(cl...)
		}
		p.set(p.funcs[i][0], false)
		if pars.fMajority != 0 {
			for v := 1; v < functions; v++ {
				p.set(p.funcs[i][v], bits.OnesCount(uint(v)) > k/2)
			}
		}
		if k == 2 && pars.fMajority == 0 {
			f := p.funcs[i]
			p.clause(bsat.Positive(f[1]), bsat.Positive(f[2]), bsat.Positive(f[3]))
			p.clause(bsat.Negative(f[1]), bsat.Positive(f[2]), bsat.Negative(f[3]))
			p.clause(bsat.Positive(f[1]), bsat.Negative(f[2]), bsat.Negative(f[3]))
			if pars.fOnlyAnd != 0 {
				p.clause(bsat.Negative(f[1]), bsat.Negative(f[2]), bsat.Positive(f[3]))
			}
		}
	}
	// All non-output objects feed at least one gate.
	for j := 0; j < total-1; j++ {
		var clause []bsat.Literal
		for i := 0; i < nodeCount; i++ {
			if j < len(p.topo[i]) {
				clause = append(clause, bsat.Positive(p.topo[i][j]))
			}
		}
		p.clause(clause...)
	}
	p.set(p.topo[nodeCount-1][total-2], true)
	if pars.fMajority != 0 {
		for j, v := range p.topo[0] {
			p.set(v, j < 3)
		}
	}
	if pars.RuntimeLim > 0 {
		p.solver.SetDeadline(time.Now().Add(time.Duration(pars.RuntimeLim) * time.Second))
	}
	return p, nil
}

func (p *zyxSynthesis) addMinterm(m int) bool {
	if p.constrained[m] {
		return true
	}
	p.constrained[m] = true
	k := int(p.pars.nLutSize)
	n := int(p.pars.nVars)
	for gate, sets := range p.combos {
		for _, fans := range sets {
			for pattern := 0; pattern < 1<<k; pattern++ {
				pre := make([]bsat.Literal, 0, 2*k+2)
				for pos, id := range fans {
					pre = append(pre, bsat.Negative(p.topo[gate][id]))
					lit := bsat.Positive(p.values[m][id])
					if pattern&(1<<pos) != 0 {
						lit ^= 1
					}
					pre = append(pre, lit)
				}
				f, out := bsat.Positive(p.funcs[gate][pattern]), bsat.Positive(p.values[m][n+gate])
				c := append(append([]bsat.Literal(nil), pre...), f^1, out)
				if !p.clause(c...) {
					return false
				}
				c = append(append([]bsat.Literal(nil), pre...), f, out^1)
				if !p.clause(c...) {
					return false
				}
			}
		}
	}
	return true
}
func (p *zyxSynthesis) model() zyxNetwork {
	result := zyxNetwork{fanins: make([][]int, len(p.topo)), funcs: make([][]bool, len(p.funcs)), complemented: p.inverted}
	for i, vars := range p.topo {
		for j, v := range vars {
			value, _ := p.solver.Value(v)
			if value {
				result.fanins[i] = append(result.fanins[i], j)
			}
		}
		result.funcs[i] = make([]bool, len(p.funcs[i]))
		for j, v := range p.funcs[i] {
			result.funcs[i][j], _ = p.solver.Value(v)
		}
	}
	return result
}
func (p *zyxSynthesis) mismatch(net zyxNetwork) int {
	n := int(p.pars.nVars)
	vals := make([]bool, n+len(net.fanins))
	for m := range p.values {
		for i := 0; i < n; i++ {
			vals[i] = m&(1<<i) != 0
		}
		for i, fans := range net.fanins {
			pat := 0
			for j, v := range fans {
				if vals[v] {
					pat |= 1 << j
				}
			}
			vals[n+i] = net.funcs[i][pat]
		}
		if vals[len(vals)-1] != (util.Abc_TtGetBit(p.target, int32(m)) != 0) {
			return m
		}
	}
	return -1
}

// ordered matches the source's reverse-lexicographic fanin/function order
// and first-use symmetry ordering for equivalent input variables.
func (p *zyxSynthesis) ordered(net zyxNetwork) bool {
	for i := 1; i < len(net.fanins); i++ {
		a, b := net.fanins[i-1], net.fanins[i]
		comparison := 0
		for j := len(a) - 1; j >= 0; j-- {
			if a[j] < b[j] {
				comparison = -1
				break
			}
			if a[j] > b[j] {
				comparison = 1
				break
			}
		}
		if comparison > 0 {
			return false
		}
		if comparison == 0 && p.pars.fMajority == 0 {
			for j := len(net.funcs[i]) - 1; j >= 0; j-- {
				if net.funcs[i-1][j] != net.funcs[i][j] {
					if net.funcs[i-1][j] {
						return false
					}
					break
				}
			}
		}
	}
	n := int(p.pars.nVars)
	for a := 0; a < n; a++ {
		for b := a + 1; b < n; b++ {
			symmetric := true
			for m := 0; m < 1<<n; m++ {
				x, y := (m>>a)&1, (m>>b)&1
				other := m
				if x != y {
					other ^= 1<<a | 1<<b
				}
				if util.Abc_TtGetBit(p.target, int32(m)) != util.Abc_TtGetBit(p.target, int32(other)) {
					symmetric = false
					break
				}
			}
			if !symmetric {
				continue
			}
			firstA, firstB := -1, -1
			for i, fans := range net.fanins {
				for _, v := range fans {
					if v == a && firstA < 0 {
						firstA = i
					}
					if v == b && firstB < 0 {
						firstB = i
					}
				}
			}
			if firstA > firstB {
				return false
			}
		}
	}
	return true
}
func (p *zyxSynthesis) blockUnordered(net zyxNetwork) bool {
	var clause []bsat.Literal
	for i, vars := range p.topo {
		selected := make(map[int]bool, len(net.fanins[i]))
		for _, j := range net.fanins[i] {
			selected[j] = true
		}
		for j, v := range vars {
			if selected[j] {
				clause = append(clause, bsat.Negative(v))
			} else {
				clause = append(clause, bsat.Positive(v))
			}
		}
	}
	// Equal-fanin gates are ordered by function, so retain function assignments
	// in the blocking clause rather than excluding their entire topology.
	for i, vars := range p.funcs {
		for j, v := range vars {
			if net.funcs[i][j] {
				clause = append(clause, bsat.Negative(v))
			} else {
				clause = append(clause, bsat.Positive(v))
			}
		}
	}
	return p.clause(clause...)
}

func (p *zyxSynthesis) solve() (*zyxNetwork, bsat.Status) {
	for {
		status := p.solver.Solve()
		if status != bsat.Satisfiable {
			return nil, status
		}
		net := p.model()
		if p.pars.fOrderNodes != 0 && !p.ordered(net) {
			if !p.blockUnordered(net) {
				return nil, bsat.Unsatisfiable
			}
			continue
		}
		m := p.mismatch(net)
		if m < 0 {
			return &net, status
		}
		if !p.addMinterm(m) {
			return nil, bsat.Unsatisfiable
		}
	}
}
func (p *zyxSynthesis) blockTopology(net *zyxNetwork) bool {
	var clause []bsat.Literal
	for i, fans := range net.fanins {
		for _, j := range fans {
			clause = append(clause, bsat.Negative(p.topo[i][j]))
		}
	}
	return p.clause(clause...)
}

func (p *zyxSynthesis) writeSolution(net *zyxNetwork, first bool) error {
	var hex strings.Builder
	for i := (1<<int(p.pars.nVars))/4 - 1; i >= 0; i-- {
		var digit int
		for b := 0; b < 4; b++ {
			bit := util.Abc_TtGetBit(p.target, int32(4*i+b)) != 0
			if p.inverted {
				bit = !bit
			}
			if bit {
				digit |= 1 << b
			}
		}
		hex.WriteByte("0123456789abcdef"[digit])
	}
	name := fmt.Sprintf("%s-%d-%d.bool", hex.String(), p.pars.nLutSize, p.pars.nNodes)
	flags := os.O_WRONLY | os.O_CREATE | os.O_APPEND
	if first {
		flags = os.O_WRONLY | os.O_CREATE | os.O_TRUNC
	}
	file, err := os.OpenFile(name, flags, 0666)
	if err != nil {
		return err
	}
	var out strings.Builder
	for i, fans := range net.fanins {
		fmt.Fprintf(&out, "%c", 'A'+int(p.pars.nVars)+i)
		for j := len(net.funcs[i]) - 1; j >= 0; j-- {
			value := net.funcs[i][j]
			if net.complemented && i == len(net.fanins)-1 {
				value = !value
			}
			if value {
				out.WriteByte('1')
			} else {
				out.WriteByte('0')
			}
		}
		for _, v := range fans {
			if v < int(p.pars.nVars) {
				fmt.Fprintf(&out, "%c", 'a'+v)
			} else {
				fmt.Fprintf(&out, "%c", 'A'+v)
			}
		}
		out.WriteByte('\n')
	}
	out.WriteByte('\n')
	_, err = file.WriteString(out.String())
	if closeErr := file.Close(); err == nil {
		err = closeErr
	}
	if err == nil {
		fmt.Printf("Dumped solution into file %q.\n", name)
	}
	return err
}

// Zyx_ManExactSynthesis synthesizes a network of k-input LUTs or majority gates.
func Zyx_ManExactSynthesis(pars *Bmc_EsPar_t) {
	start := time.Now()
	p, err := newZyxSynthesis(pars)
	if err != nil {
		fmt.Printf("Exact synthesis: %v.\n", err)
		return
	}
	fmt.Printf("Running exact synthesis for %d-input function with %d %d-input gates...\n", pars.nVars, pars.nNodes, pars.nLutSize)
	count := 0
	for {
		net, status := p.solve()
		if status != bsat.Satisfiable {
			if count == 0 {
				fmt.Println("The problem has no solution or the solver stopped.")
			}
			break
		}
		count++
		fmt.Printf("Realization of given %d-input function using %d %d-input gates:\n", pars.nVars, pars.nNodes, pars.nLutSize)
		for i := len(net.fanins) - 1; i >= 0; i-- {
			fmt.Printf("%02d = ", int(pars.nVars)+i)
			if pars.fMajority != 0 {
				fmt.Print("MAJ3")
			} else {
				fmt.Printf("%d'b", len(net.funcs[i]))
				for j := len(net.funcs[i]) - 1; j >= 0; j-- {
					value := net.funcs[i][j]
					if net.complemented && i == len(net.fanins)-1 {
						value = !value
					}
					if value {
						fmt.Print("1")
					} else {
						fmt.Print("0")
					}
				}
			}
			fmt.Print("(")
			for _, v := range net.fanins[i] {
				if v < int(pars.nVars) {
					fmt.Printf(" %c", 'a'+v)
				} else {
					fmt.Printf(" %02d", v)
				}
			}
			fmt.Println(" )")
		}
		if pars.fMajority == 0 {
			if err := p.writeSolution(net, count == 1); err != nil {
				fmt.Printf("Cannot write solution: %v\n", err)
			}
		}
		if pars.fEnumSols == 0 || !p.blockTopology(net) {
			break
		}
	}
	if pars.fEnumSols != 0 {
		fmt.Printf("Finished enumerating %d solutions.\n", count)
	}
	fmt.Printf("Total runtime: %s\n", time.Since(start))
}
