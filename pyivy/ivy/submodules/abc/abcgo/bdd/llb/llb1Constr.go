package llb

import (
	"fmt"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/bdd/cudd/core"
	"github.com/glycerine/abcgo/misc/vec"
)

// Llb_ManCountEntries counts candidate literals encoded as 0 or 1.
func Llb_ManCountEntries(candidates *Vec_Int_t) int32 {
	if candidates == nil {
		return 0
	}
	var count int32
	for _, entry := range vec.Vec_IntArray(candidates) {
		if entry == 0 || entry == 1 {
			count++
		}
	}
	return count
}

// Llb_ManPrintEntries prints the candidate literals by AIG node ID.
func Llb_ManPrintEntries(g *Aig_Man_t, candidates *Vec_Int_t) {
	if candidates == nil {
		fmt.Println("There are no hints.")
		return
	}
	count := Llb_ManCountEntries(candidates)
	suffix := "s"
	if count == 1 {
		suffix = ""
	}
	fmt.Printf("\n*** Using %d hint%s:\n", count, suffix)
	for id, entry := range vec.Vec_IntArray(candidates) {
		if entry != 0 && entry != 1 {
			continue
		}
		sign := '-'
		if entry == 1 {
			sign = '+'
		}
		kind := "deleted"
		e := aig.Edge(id << 1)
		if g.ValidEdge(e) {
			kind = fmt.Sprint(g.Node(e).Kind)
		}
		fmt.Printf("%c%-6d : node %d (%s)\n", sign, id, id, kind)
	}
}

// llbBuildBDDTable compiles graph nodes in an explicitly supplied CI valuation.
// The inputs slice is indexed in Graph.Inputs order; a nil entry means that CI
// should be named by its corresponding BDD variable.
func llbBuildBDDTable(g *aig.Graph, dd *core.DdManager, inputs []*core.DdNode) []*core.DdNode {
	nodes := g.Nodes()
	funcs := make([]*core.DdNode, len(nodes))
	if len(nodes) == 0 {
		return funcs
	}
	funcs[0] = core.Cudd_ReadOne(dd)
	for i, input := range g.Inputs() {
		if i >= len(inputs) || inputs[i] == nil {
			if i >= len(dd.Vars) {
				return nil
			}
			inputs[i] = core.Cudd_bddIthVar(dd, i)
		}
		funcs[input.NodeID()] = inputs[i]
	}
	var eval func(aig.Edge, []bool) *core.DdNode
	eval = func(edge aig.Edge, active []bool) *core.DdNode {
		id := edge.NodeID()
		if id < 0 || id >= len(funcs) || !g.ValidEdge(aig.Edge(id<<1)) {
			return nil
		}
		base := funcs[id]
		if base == nil {
			if active[id] {
				panic("llb: cycle in AIG node graph")
			}
			active[id] = true
			node := nodes[id]
			left := eval(node.Fanins[0], active)
			switch node.Kind {
			case aig.And:
				right := eval(node.Fanins[1], active)
				base = core.Cudd_bddAnd(dd, left, right)
			case aig.ExclusiveOr:
				right := eval(node.Fanins[1], active)
				base = core.Cudd_bddXor(dd, left, right)
			case aig.Buffer, aig.Output:
				base = left
			case aig.Constant:
				base = core.Cudd_ReadOne(dd)
			case aig.Input:
				panic("llb: input node is missing its BDD variable")
			default:
				panic(fmt.Sprintf("llb: unsupported AIG node kind %d", node.Kind))
			}
			funcs[id] = base
			active[id] = false
		}
		if edge.Complemented() {
			return core.Cudd_Not(base)
		}
		return base
	}
	for _, edge := range g.LiveEdges() {
		if funcs[edge.NodeID()] != nil {
			continue
		}
		// Only invoke recursive evaluation on regular edges; polarity is applied
		// by callers when they consume fanins.
		_ = eval(edge, make([]bool, len(nodes)))
	}
	return funcs
}

func llbBDDForEdge(g *aig.Graph, funcs []*core.DdNode, edge aig.Edge) *core.DdNode {
	if edge.NodeID() < 0 || edge.NodeID() >= len(funcs) || !g.ValidEdge(aig.Edge(edge.NodeID()<<1)) {
		return nil
	}
	f := funcs[edge.NodeID()]
	if edge.Complemented() {
		return core.Cudd_Not(f)
	}
	return f
}

func llbBaseCase(g *aig.Graph, dd *core.DdManager, funcs []*core.DdNode, property *core.DdNode) *Vec_Int_t {
	candidates := vec.Vec_IntStartFull(len(funcs))
	notProperty := core.Cudd_Not(property)
	for _, edge := range g.LiveEdges() {
		node := g.Node(edge)
		if node.Kind != aig.Input && node.Kind != aig.And && node.Kind != aig.ExclusiveOr && node.Kind != aig.Buffer {
			continue
		}
		id := edge.NodeID()
		f := funcs[id]
		if core.Cudd_bddLeq(dd, f, notProperty) {
			vec.Vec_IntWriteEntry(candidates, id, 1)
		} else if core.Cudd_bddLeq(dd, core.Cudd_Not(f), notProperty) {
			vec.Vec_IntWriteEntry(candidates, id, 0)
		}
	}
	return candidates
}

func llbInductiveCase(g *aig.Graph, dd *core.DdManager, global []*core.DdNode, candidates *Vec_Int_t) {
	inputs := g.Inputs()
	outputs := g.Outputs()
	nRegs := int(g.RegisterCount())
	if nRegs < 0 || nRegs > len(inputs) || nRegs > len(outputs) {
		panic("llb: invalid sequential interface")
	}
	nPis := len(inputs) - nRegs
	transitionVars := make([]*core.DdNode, len(inputs))
	for i := 0; i < nPis; i++ {
		transitionVars[i] = core.Cudd_bddIthVar(dd, len(inputs)+i)
	}
	firstLatchOutput := len(outputs) - nRegs
	for i := 0; i < nRegs; i++ {
		nextState := llbBDDForEdge(g, global, outputs[firstLatchOutput+i])
		transitionVars[nPis+i] = nextState
	}
	transition := llbBuildBDDTable(g, dd, transitionVars)
	if transition == nil {
		panic("llb: failed to construct transition BDD table")
	}
	values := vec.Vec_IntArray(candidates)
	for id, candidate := range values {
		if candidate != 0 && candidate != 1 {
			continue
		}
		current, next := global[id], transition[id]
		if current == nil || next == nil {
			vec.Vec_IntWriteEntry(candidates, id, -1)
			continue
		}
		preserved := false
		if candidate == 0 {
			preserved = core.Cudd_bddLeq(dd, core.Cudd_Not(current), core.Cudd_Not(next))
		} else {
			preserved = core.Cudd_bddLeq(dd, current, next)
		}
		if !preserved {
			vec.Vec_IntWriteEntry(candidates, id, -1)
		}
	}
}

// Llb_ManDeriveConstraints returns literals that are both safe initially and
// inductive for the supplied one-step transition relation. A nil result means
// that no candidate survived.
func Llb_ManDeriveConstraints(g *Aig_Man_t) *Vec_Int_t {
	if g == nil {
		return nil
	}
	inputs, outputs := g.Inputs(), g.Outputs()
	nRegs := int(g.RegisterCount())
	if nRegs < 0 || nRegs > len(inputs) || nRegs > len(outputs) {
		return nil
	}
	propertyOutputs := len(outputs) - nRegs
	if propertyOutputs != 1 {
		fmt.Printf("The AIG has %d property outputs.\n", propertyOutputs)
		return nil
	}
	dd := core.NewBDDManager(len(inputs) + len(inputs) - nRegs)
	if dd == nil {
		return nil
	}
	defer core.Cudd_Quit(dd)
	initialVars := make([]*core.DdNode, len(inputs))
	for i := range initialVars {
		initialVars[i] = core.Cudd_bddIthVar(dd, i)
	}
	global := llbBuildBDDTable(g, dd, initialVars)
	if global == nil {
		return nil
	}
	property := llbBDDForEdge(g, global, outputs[0])
	if property == nil {
		return nil
	}
	candidates := llbBaseCase(g, dd, global, property)
	if Llb_ManCountEntries(candidates) != 0 {
		llbInductiveCase(g, dd, global, candidates)
	}
	if Llb_ManCountEntries(candidates) == 0 {
		vec.Vec_IntFree(candidates)
		return nil
	}
	return candidates
}

func Llb_ManConstrTest(g *Aig_Man_t) {
	candidates := Llb_ManDeriveConstraints(g)
	Llb_ManPrintEntries(g, candidates)
	vec.Vec_IntFree(candidates)
}
