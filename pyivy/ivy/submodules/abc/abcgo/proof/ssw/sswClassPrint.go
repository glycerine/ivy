package ssw

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/aig"
)

func sswPrintClassNode(p *Ssw_Cla_t, obj aig.Edge) {
	fmt.Printf("%d(%d,%d,%d) ", obj.NodeID(), p.pAig.Node(obj).Level, aig.Aig_SupportSize(p.pAig, obj), aig.Aig_NodeMffcSupp(p.pAig, obj, 0, nil))
}
func Ssw_ClassesPrintOne(p *Ssw_Cla_t, repr aig.Edge) {
	fmt.Print("{ ")
	for _, obj := range p.pId2Class[repr.NodeID()] {
		sswPrintClassNode(p, obj)
	}
	fmt.Println("}")
}
func Ssw_ClassesPrint(p *Ssw_Cla_t, verbose int32) {
	fmt.Printf("Equiv classes: Const1 = %5d. Class = %5d. Lit = %5d.\n", p.nCands1, p.nClasses, p.nCands1+p.nLits)
	if verbose == 0 {
		return
	}
	fmt.Print("Constants { ")
	for _, obj := range p.pAig.LiveEdges() {
		if Ssw_ObjIsConst1Cand(p.pAig, obj) != 0 {
			sswPrintClassNode(p, obj)
		}
	}
	fmt.Println("}")
	for i, class := range p.pId2Class {
		if class != nil {
			fmt.Printf("%3d (%3d) : ", i, p.pClassSizes[i])
			Ssw_ClassesPrintOne(p, class[0])
		}
	}
	fmt.Println()
}
