package fra

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/opt/dar"
	"time"
)

func Fra_FraigInductionRewrite(p *Fra_Man_t) {
	started := time.Now()
	old := p.pManFraig
	rewritten, err := dar.Dar_ManRewriteDefault(old)
	if err != nil {
		panic(err)
	}
	if old.RegisterCount() != rewritten.RegisterCount() || old.Assertions != rewritten.Assertions {
		panic("fra: rewriting changed induction interface")
	}
	for i := range p.pMemFraig {
		p.pMemFraig[i] = aig.InvalidEdge
	}
	frame := p.pPars.NFramesK
	Fra_ObjSetFraig(p, p.pManAig.Const1(), frame, rewritten.Const1())
	inputs := p.pManAig.Inputs()
	regs := int(p.pManAig.RegisterCount())
	trueInputs := len(inputs) - regs
	rewrittenInputs := rewritten.Inputs()
	for i, input := range inputs[:trueInputs] {
		Fra_ObjSetFraig(p, input, frame, rewrittenInputs[trueInputs*int(frame)+i])
	}
	outputs := rewritten.Outputs()
	if len(outputs)-int(rewritten.Assertions) != regs {
		panic("fra: rewritten induction latch count changed")
	}
	for i, latch := range inputs[trueInputs:] {
		Fra_ObjSetFraig(p, latch, frame, rewritten.Node(outputs[int(rewritten.Assertions)+i]).Fanins[0])
	}
	p.pManFraig = rewritten
	p.timeRwr += time.Since(started)
}

// Fra_FramesAddMore extends the graph's object-copy mapping in place, as in
// fraInd.c. The existing graph interface remains unchanged.
func Fra_FramesAddMore(graph *aig.Graph, frames int32) {
	if frames < 0 {
		panic("fra: negative frame count")
	}
	original := graph.LiveEdges()
	for _, edge := range original {
		aig.Aig_ObjSetCopy(graph, edge, edge)
	}
	inputs, outputs := graph.Inputs(), graph.Outputs()
	regs := int(graph.RegisterCount())
	for f := int32(0); f < frames; f++ {
		for _, edge := range inputs[len(inputs)-regs:] {
			aig.Aig_ObjSetCopy(graph, edge, aig.InvalidEdge)
		}
		for _, edge := range outputs[len(outputs)-regs:] {
			aig.Aig_ObjSetCopy(graph, edge, aig.InvalidEdge)
		}
		latchCopies := make([]aig.Edge, regs)
		for i, out := range outputs[len(outputs)-regs:] {
			latchCopies[i] = aig.Aig_ObjRealCopy(graph, graph.Node(out).Fanins[0])
		}
		for i, latch := range inputs[len(inputs)-regs:] {
			aig.Aig_ObjSetCopy(graph, latch, latchCopies[i])
		}
		for _, edge := range original {
			node := graph.Node(edge)
			if node.Kind != aig.And && node.Kind != aig.ExclusiveOr {
				continue
			}
			a, b := aig.Aig_ObjRealCopy(graph, node.Fanins[0]), aig.Aig_ObjRealCopy(graph, node.Fanins[1])
			if a == aig.InvalidEdge || b == aig.InvalidEdge {
				aig.Aig_ObjSetCopy(graph, edge, aig.InvalidEdge)
				continue
			}
			if node.Kind == aig.And {
				aig.Aig_ObjSetCopy(graph, edge, graph.AND(a, b))
			} else {
				aig.Aig_ObjSetCopy(graph, edge, graph.NewXor(a, b))
			}
		}
	}
}
