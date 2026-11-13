package fracomb

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/aig"
	dar "github.com/glycerine/abcgo/opt/dar/core"
)

// Sweep supplies the original FRAIG manager to the combinational CEC driver.
// It must return a new graph; the driver stops its input after the call.
type Sweep func(*aig.Graph, int32, int32) *aig.Graph

func Fra_FraigMiterStatus(g *aig.Graph) int32 {
	if g.Data != nil {
		return 0
	}
	regs := int(g.RegisterCount())
	outputs := g.Outputs()
	undecided := false
	for _, out := range outputs[:len(outputs)-regs] {
		child := g.Node(out).Fanins[0]
		if child == g.Const0() {
			continue
		}
		if child == g.Const1() || aig.Aig_ObjIsCi(g, child) && aig.Aig_ObjCioId(g, child) < g.TruePiCount() || g.Node(child).Phase != child.Complemented() {
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
// attempts to the given combinational miter. Fallback rewrites consume and
// replace the caller's graph, as in ABC.
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
		// ABC returns an all-zero model here without validating it. In
		// particular, a direct positive PI has an invalid zero witness.
		miter.Data = make([]int32, miter.InputCount())
		return 0
	}
	status = Fra_FraigSat(miter, 600, 0, 0, 0, 0, 1, 0, 0, 0)
	if status >= 0 {
		return status
	}
	var err error
	previous := miter
	miter, err = dar.Dar_ManRwsat(previous, true, false)
	aig.Aig_ManStop(previous)
	if err != nil || miter == nil {
		panic("fra: source dereferences failed SAT rewrite")
	}
	nodeLimit, miterLimit := int32(2), int32(300)
	for i := 0; i < 6; i++ {
		if Aig_ManCountXors(miter)*30 > aig.Aig_ManNodeNum(miter)+300 {
			previous = miter
			miter = dar.Dar_ManBalanceXor(previous, true, false, false)
			aig.Aig_ManStop(previous)
		}
		previous = miter
		miter = sweep(previous, nodeLimit, miterLimit)
		aig.Aig_ManStop(previous)
		if miter == nil {
			panic("fra: source dereferences failed FRAIG sweep")
		}
		status = Fra_FraigMiterStatus(miter)
		if status >= 0 {
			break
		}
		previous = miter
		miter, err = dar.Dar_ManRewriteDefault(previous)
		aig.Aig_ManStop(previous)
		if err != nil || miter == nil {
			panic("fra: source dereferences failed rewrite")
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
