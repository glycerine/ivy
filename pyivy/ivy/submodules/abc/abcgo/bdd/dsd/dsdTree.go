package dsd

import (
	"fmt"
	"io"
	"strings"

	bdd "github.com/glycerine/abcgo/bdd/cudd/core"
	"github.com/glycerine/abcgo/opt/dau/canon"
)

func Dsd_TreeNodeCreate(kind, nDecs, blockNum int) *Dsd_Node_t {
	if kind < int(DSD_NODE_CONST1) || kind > int(DSD_NODE_PRIME) || nDecs < 0 {
		panic("dsd: invalid node dimensions")
	}
	p := &Dsd_Node_t{Type: Dsd_Type_t(kind), pDecs: make([]*Dsd_Node_t, nDecs)}
	p.inverse = &Dsd_Node_t{regular: p, inverse: p}
	return p
}
func Dsd_TreeNodeDelete(dd *DdManager, p *Dsd_Node_t) {
	p = Dsd_Regular(p)
	if p == nil {
		return
	}
	if p.G != nil {
		bdd.Cudd_RecursiveDeref(dd, p.G)
	}
	if p.S != nil {
		bdd.Cudd_RecursiveDeref(dd, p.S)
	}
	p.G, p.S = nil, nil
	p.pDecs = nil
}

// Dsd_TreeUnmark resets traversal visits; the user-visible Mark survives.
func Dsd_TreeUnmark(m *Dsd_Manager_t) {
	seen := map[*Dsd_Node_t]bool{}
	var clearVisits func(*Dsd_Node_t)
	clearVisits = func(p *Dsd_Node_t) {
		p = Dsd_Regular(p)
		if p == nil || seen[p] {
			return
		}
		seen[p] = true
		p.visits = 0
		if p.Type != DSD_NODE_BUF {
			for _, c := range p.pDecs {
				clearVisits(c)
			}
		}
	}
	for _, p := range m.roots {
		clearVisits(p)
	}
}
func Dsd_TreeNodeGetInfoOne(p *Dsd_Node_t, depth, gate *int) {
	d, g := 0, 0
	var visit func(*Dsd_Node_t, int)
	visit = func(p *Dsd_Node_t, rank int) {
		p = Dsd_Regular(p)
		size := len(p.pDecs)
		if p.Type == DSD_NODE_OR || p.Type == DSD_NODE_EXOR {
			size = 2
		}
		g = max(g, size)
		if len(p.pDecs) < 2 {
			return
		}
		d = max(d, rank+1)
		for _, c := range p.pDecs {
			visit(c, rank+1)
		}
	}
	visit(p, 0)
	if depth != nil {
		*depth = d
	}
	if gate != nil {
		*gate = g
	}
}
func Dsd_TreeNodeGetInfo(m *Dsd_Manager_t, depth, gate *int) {
	d, g := 0, 0
	for _, p := range m.roots {
		pd, pg := 0, 0
		Dsd_TreeNodeGetInfoOne(p, &pd, &pg)
		d = max(d, pd)
		g = max(g, pg)
	}
	if depth != nil {
		*depth = d
	}
	if gate != nil {
		*gate = g
	}
}
func Dsd_TreeGetAigCost(p *Dsd_Node_t) int {
	p = Dsd_Regular(p)
	if len(p.pDecs) < 2 {
		return 0
	}
	cost := 0
	switch p.Type {
	case DSD_NODE_OR:
		cost = len(p.pDecs) - 1
	case DSD_NODE_EXOR:
		cost = 3 * (len(p.pDecs) - 1)
	case DSD_NODE_PRIME:
		if len(p.pDecs) == 3 {
			cost = 3
		}
	}
	for _, c := range p.pDecs {
		cost += Dsd_TreeGetAigCost(c)
	}
	return cost
}
func collectNodes(roots []*Dsd_Node_t) []*Dsd_Node_t {
	result := []*Dsd_Node_t{}
	touched := []*Dsd_Node_t{}
	var visit func(*Dsd_Node_t)
	visit = func(p *Dsd_Node_t) {
		p = Dsd_Regular(p)
		if p == nil || len(p.pDecs) < 2 || p.visits != 0 {
			return
		}
		p.visits++
		touched = append(touched, p)
		for _, c := range p.pDecs {
			visit(c)
		}
		result = append(result, p)
	}
	for _, p := range roots {
		visit(p)
	}
	for _, p := range touched {
		p.visits = 0
	}
	return result
}
func Dsd_TreeCountNonTerminalNodesOne(p *Dsd_Node_t) int { return len(collectNodes([]*Dsd_Node_t{p})) }
func Dsd_TreeCountNonTerminalNodes(m *Dsd_Manager_t) int { return len(collectNodes(m.roots)) }
func countPrimes(nodes []*Dsd_Node_t) int {
	n := 0
	for _, p := range nodes {
		if p.Type == DSD_NODE_PRIME {
			n++
		}
	}
	return n
}
func Dsd_TreeCountPrimeNodesOne(p *Dsd_Node_t) int {
	return countPrimes(collectNodes([]*Dsd_Node_t{p}))
}
func Dsd_TreeCountPrimeNodes(m *Dsd_Manager_t) int { return countPrimes(collectNodes(m.roots)) }
func Dsd_TreeCollectNodesDfs(m *Dsd_Manager_t, n *int) []*Dsd_Node_t {
	out := collectNodes(m.roots)
	*n = len(out)
	return out
}
func Dsd_TreeCollectNodesDfsOne(m *Dsd_Manager_t, p *Dsd_Node_t, n *int) []*Dsd_Node_t {
	out := collectNodes([]*Dsd_Node_t{p})
	*n = len(out)
	return out
}
func Dsd_TreeCollectDecomposableVars(m *Dsd_Manager_t, vars []int) int {
	if len(m.roots) == 0 {
		return 0
	}
	count := 0
	var visit func(*Dsd_Node_t) bool
	visit = func(p *Dsd_Node_t) bool {
		p = Dsd_Regular(p)
		if len(p.pDecs) <= 1 {
			return false
		}
		skip := false
		for _, c := range p.pDecs {
			skip = visit(c) || skip
		}
		if skip || (p.Type != DSD_NODE_OR && p.Type != DSD_NODE_EXOR && len(p.pDecs) > 4) {
			return true
		}
		for _, c := range p.pDecs {
			r := Dsd_Regular(c)
			if r.Type == DSD_NODE_BUF {
				if vars != nil {
					vars[count] = int(r.S.Index)
				}
				count++
			}
		}
		return false
	}
	visit(m.roots[0])
	return count
}
func inputName(names []string, index int, short bool) string {
	if short || index >= len(names) {
		return string(rune('a' + index))
	}
	return names[index]
}
func outputName(names []string, index int) string {
	if index < len(names) {
		return names[index]
	}
	return fmt.Sprintf("out%d", index)
}
func selectedRoots(m *Dsd_Manager_t, output int) []int {
	if output >= 0 {
		if output >= len(m.roots) {
			panic("dsd: output index out of range")
		}
		return []int{output}
	}
	if output != -1 {
		panic("dsd: invalid output selection")
	}
	indices := make([]int, len(m.roots))
	for i := range indices {
		indices[i] = i
	}
	return indices
}

// Dsd_TreePrint writes a line-oriented decomposition with numbered subblocks.
func Dsd_TreePrint(w io.Writer, m *Dsd_Manager_t, inputs, outputs []string, short, output int) {
	counter := 1
	for _, i := range selectedRoots(m, output) {
		printNode(w, m.roots[i], outputName(outputs, i), inputs, short != 0, 0, &counter)
	}
}
func printNode(w io.Writer, edge *Dsd_Node_t, name string, inputs []string, short bool, indent int, counter *int) {
	p := Dsd_Regular(edge)
	fmt.Fprint(w, strings.Repeat(" ", indent))
	if Dsd_IsComplement(edge) {
		fmt.Fprintf(w, "NOT(%s) = ", name)
	} else {
		fmt.Fprintf(w, "%s = ", name)
	}
	if p.Type == DSD_NODE_CONST1 {
		fmt.Fprintln(w, " Constant 1.")
		return
	}
	if p.Type == DSD_NODE_BUF {
		fmt.Fprintln(w, inputName(inputs, int(p.S.Index), short))
		return
	}
	kind := map[Dsd_Type_t]string{DSD_NODE_OR: "OR", DSD_NODE_EXOR: "EXOR", DSD_NODE_PRIME: "PRIME"}[p.Type]
	fmt.Fprintf(w, "%s(", kind)
	numbers := make([]int, len(p.pDecs))
	for i, c := range p.pDecs {
		if i > 0 {
			fmt.Fprint(w, ",")
		}
		fmt.Fprint(w, " ")
		if Dsd_IsComplement(c) {
			fmt.Fprint(w, "NOT(")
		}
		r := Dsd_Regular(c)
		if r.Type == DSD_NODE_BUF {
			fmt.Fprint(w, inputName(inputs, int(r.S.Index), short))
		} else {
			numbers[i] = *counter
			*counter++
			fmt.Fprintf(w, "<%d>", numbers[i])
		}
		if Dsd_IsComplement(c) {
			fmt.Fprint(w, ")")
		}
	}
	fmt.Fprintln(w, " )")
	for i, n := range numbers {
		if n != 0 {
			printNode(w, Dsd_Regular(p.pDecs[i]), fmt.Sprintf("<%d>", n), inputs, short, indent+6, counter)
		}
	}
}
func Dsd_NodePrint(w io.Writer, p *Dsd_Node_t) {
	counter := 1
	printNode(w, p, "F", nil, true, 0, &counter)
}
func Dsd_TreeFunc2Truth_rec(dd *DdManager, f *DdNode) uint64 {
	patterns := [6]uint64{0xaaaaaaaaaaaaaaaa, 0xcccccccccccccccc, 0xf0f0f0f0f0f0f0f0, 0xff00ff00ff00ff00, 0xffff0000ffff0000, 0xffffffff00000000}
	memo := map[*DdNode]uint64{}
	var visit func(*DdNode) uint64
	visit = func(f *DdNode) uint64 {
		if f == dd.One {
			return ^uint64(0)
		}
		if f == dd.LogicZero {
			return 0
		}
		if t, ok := memo[f]; ok {
			return t
		}
		lev := level(dd, f)
		if lev >= 6 {
			panic("dsd: truth display has more than six variables")
		}
		lo, hi := split(dd, f, lev)
		v := patterns[lev]
		t := v&visit(hi) | ^v&visit(lo)
		memo[f] = t
		return t
	}
	return visit(f)
}
func Dsd_TreePrint2(w io.Writer, m *Dsd_Manager_t, inputs, outputs []string, output int) {
	for _, i := range selectedRoots(m, output) {
		fmt.Fprintf(w, "%8s = ", outputName(outputs, i))
		printExpression(w, m.dd, m.roots[i], inputs)
		fmt.Fprintln(w)
	}
}
func printExpression(w io.Writer, dd *DdManager, edge *Dsd_Node_t, inputs []string) {
	p := Dsd_Regular(edge)
	comp := Dsd_IsComplement(edge)
	switch p.Type {
	case DSD_NODE_CONST1:
		if comp {
			fmt.Fprint(w, "Const0")
		} else {
			fmt.Fprint(w, "Const1")
		}
	case DSD_NODE_BUF:
		if comp {
			fmt.Fprint(w, "!")
		}
		fmt.Fprint(w, inputName(inputs, int(p.S.Index), false))
	case DSD_NODE_OR:
		if !comp {
			fmt.Fprint(w, "!")
		}
		fmt.Fprint(w, "(")
		for _, c := range p.pDecs {
			printExpression(w, dd, Dsd_Not(c), inputs)
		}
		fmt.Fprint(w, ")")
	case DSD_NODE_EXOR:
		if comp {
			fmt.Fprint(w, "!")
		}
		fmt.Fprint(w, "[")
		for _, c := range p.pDecs {
			printExpression(w, dd, c, inputs)
		}
		fmt.Fprint(w, "]")
	case DSD_NODE_PRIME:
		n := len(p.pDecs)
		fmt.Fprint(w, " ")
		if n <= 6 {
			truth := []uint64{Dsd_TreeFunc2Truth_rec(dd, Dsd_TreeGetPrimeFunction(dd, p))}
			perm := make([]byte, n)
			phase := canon.Canonicalize(truth, int32(n), perm)
			if comp != (phase>>uint(n)&1 != 0) {
				fmt.Fprint(w, "!")
			}
			digits := max(1, (1<<uint(n))/4)
			mask := ^uint64(0)
			if n < 6 {
				mask = (uint64(1) << uint(1<<uint(n))) - 1
			}
			fmt.Fprintf(w, "%0*x{", digits, truth[0]&mask)
			for i, index := range perm {
				printExpression(w, dd, Dsd_NotCond(p.pDecs[index], phase>>uint(i)&1 != 0), inputs)
			}
		} else {
			if comp {
				fmt.Fprint(w, "!")
			}
			fmt.Fprintf(w, "|%d|{", n)
			for _, c := range p.pDecs {
				printExpression(w, dd, c, inputs)
			}
		}
		fmt.Fprint(w, "} ")
	}
}
