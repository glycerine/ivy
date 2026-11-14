package ssw

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
)

func Ssw_ManPrintPolarity(g *aig.Graph) {
	for _, obj := range g.LiveEdges() {
		if g.Node(obj).Phase {
			fmt.Print("1")
		} else {
			fmt.Print("0")
		}
	}
	fmt.Println()
}
func sswSetMark(g *aig.Graph, obj aig.Edge, value bool) {
	if value {
		aig.Aig_ObjSetMarkB(g, obj)
	} else {
		aig.Aig_ObjClearMarkB(g, obj)
	}
}
func sswMarkValue(g *aig.Graph, obj aig.Edge) bool { return g.Node(obj).MarkB != obj.Complemented() }

// Ssw_ManRefineByConstrSim simulates the given PI assignment and refines classes.
func Ssw_ManRefineByConstrSim(p *Ssw_Man_t) {
	g := p.pAig
	inputs, outputs := g.Inputs(), g.Outputs()
	regs := int(g.RegisterCount())
	pis, pos := len(inputs)-regs, len(outputs)-regs
	if pis <= 0 || vec.Vec_IntSize(p.vInits)%pis != 0 {
		panic("ssw: invalid constrained simulation inputs")
	}
	for _, li := range outputs[pos:] {
		sswSetMark(g, li, false)
	}
	frames := vec.Vec_IntSize(p.vInits) / pis
	nodes := g.TopologicalObjects()
	offset := 0
	for f := 0; f < frames; f++ {
		sswSetMark(g, g.Const1(), true)
		for _, pi := range inputs[:pis] {
			sswSetMark(g, pi, vec.Vec_IntEntry(p.vInits, offset) != 0)
			offset++
		}
		for i, lo := range inputs[pis:] {
			sswSetMark(g, lo, g.Node(outputs[pos+i]).MarkB)
		}
		for _, obj := range nodes {
			n := g.Node(obj)
			switch n.Kind {
			case aig.And:
				sswSetMark(g, obj, sswMarkValue(g, n.Fanins[0]) && sswMarkValue(g, n.Fanins[1]))
			case aig.ExclusiveOr:
				sswSetMark(g, obj, sswMarkValue(g, n.Fanins[0]) != sswMarkValue(g, n.Fanins[1]))
			case aig.Buffer:
				sswSetMark(g, obj, sswMarkValue(g, n.Fanins[0]))
			}
		}
		for _, out := range outputs {
			sswSetMark(g, out, sswMarkValue(g, g.Node(out).Fanins[0]))
		}
		constr := int(aig.Aig_ManConstrNum(g))
		if constr > 0 {
			for i, out := range outputs[:pos] {
				if g.Node(out).MarkB {
					if i < pos-constr {
						fmt.Printf("output %d failed in frame %d.\n", i, f)
					} else {
						fmt.Printf("constraint %d failed in frame %d.\n", i, f)
					}
				}
			}
		}
		if f == 0 {
			for _, obj := range g.LiveEdges() {
				aig.Aig_ObjSetPhase(g, obj, g.Node(obj).MarkB)
			}
		} else {
			Ssw_ClassesRefineConst1(p.ppClasses, 0)
			Ssw_ClassesRefine(p.ppClasses, 0)
		}
	}
}
