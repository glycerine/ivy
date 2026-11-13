package fracomb

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/opt/dar"
)

// Sweep supplies the original FRAIG manager to the combinational CEC driver.
type Sweep func(*aig.Graph, int32, int32) *aig.Graph

func Fra_FraigMiterStatus(g *aig.Graph) int32 {
	if g.Data != nil {
		return 0
	}
	inputs := g.Inputs()
	regs := int(g.RegisterCount())
	pis := make(map[int]bool, len(inputs)-regs)
	for _, e := range inputs[:len(inputs)-regs] {
		pis[e.NodeID()] = true
	}
	outputs := g.Outputs()
	undecided := false
	for _, out := range outputs[:len(outputs)-regs] {
		child := g.Node(out).Fanins[0]
		if child == g.Const0() {
			continue
		}
		if child == g.Const1() || pis[child.NodeID()] || g.Node(child).Phase != child.Complemented() {
			return 0
		}
		undecided = true
	}
	if undecided {
		return -1
	}
	return 1
}
func Fra_FraigMiterAssertedOutput(g *aig.Graph) int32 {
	outputs := g.Outputs()
	regs := int(g.RegisterCount())
	for i, out := range outputs[:len(outputs)-regs] {
		child := g.Node(out).Fanins[0]
		if child != g.Const0() && (child == g.Const1() || g.Node(child).Phase != child.Complemented()) {
			return int32(i)
		}
	}
	return -1
}

// Aig_ManCountXors counts XOR structures recognized by the AIG manager.
func Aig_ManCountXors(graph *aig.Graph) int32 {
	var count int32
	for _, edge := range graph.LiveEdges() {
		if aig.Aig_ObjIsMuxType(graph, edge) {
			if _, _, ok := aig.Aig_ObjRecognizeExor(graph, edge); ok {
				count++
			}
		}
	}
	return count
}

// Fra_FraigCec applies the original sequence of SAT, rewriting, and FRAIG
// attempts to the given miter. Rewrites replace the caller's graph.
func Fra_FraigCec(graph **aig.Graph, conflictLimit, verbose int32, sweep Sweep) int32 {
	if graph == nil || *graph == nil {
		panic("fra: nil CEC miter")
	}
	miter := *graph
	if verbose != 0 {
		fmt.Printf("Original miter:   Nodes = %6d.\n", aig.Aig_ManNodeNum(miter))
	}
	status := Fra_FraigMiterStatus(miter)
	if status == 0 {
		miter.Data = make([]int32, miter.InputCount())
		return status
	}
	status = Fra_FraigSat(miter, 600, 0, 0, 0, 0, 1, 0, 0, 0)
	if status >= 0 {
		return status
	}
	var err error
	miter, err = dar.Dar_ManRwsat(miter, true, false)
	if err != nil || miter == nil {
		return -1
	}
	nodeLimit, miterLimit := int32(2), int32(300)
	for i := 0; i < 6; i++ {
		if Aig_ManCountXors(miter)*30 > aig.Aig_ManNodeNum(miter)+300 {
			miter = dar.Dar_ManBalanceXor(miter, true, false, false)
		}
		miter = sweep(miter, nodeLimit, miterLimit)
		if miter == nil {
			return -1
		}
		status = Fra_FraigMiterStatus(miter)
		if status >= 0 {
			break
		}
		miter, err = dar.Dar_ManRewriteDefault(miter)
		if err != nil || miter == nil {
			return -1
		}
		status = Fra_FraigMiterStatus(miter)
		if status >= 0 {
			break
		}
		nodeLimit *= 8
		miterLimit *= 2
	}
	if status == -1 {
		status = Fra_FraigSat(miter, int64(conflictLimit), 0, 0, 0, 0, 1, 0, 0, 0)
	}
	*graph = miter
	return status
}
