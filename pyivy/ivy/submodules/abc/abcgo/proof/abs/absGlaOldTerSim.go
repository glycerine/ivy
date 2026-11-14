package abs

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
)

// Gla_ManVerifyUsingTerSim checks the selected pseudo-inputs with all other
// pseudo-inputs unknown, and clears the temporary GIA ternary marks afterward.
func Gla_ManVerifyUsingTerSim(p *Gla_Man_t, pis, ppis, nodes, outputs, selected *vec.Vec_Int_t) {
	g := p.pGia
	for frame := int32(0); frame <= p.pPars.IFrame; frame++ {
		gia.Gia_ObjTerSimSet0(gia.Gia_ManConst0(g))
		setValues := func(v *vec.Vec_Int_t) {
			for _, id := range vec.Vec_IntArray(v) {
				o := gia.Gia_ManObj(g, id)
				if Gla_ObjSatValue(p, id, frame) != 0 {
					gia.Gia_ObjTerSimSet1(o)
				} else {
					gia.Gia_ObjTerSimSet0(o)
				}
			}
		}
		setValues(pis)
		for _, id := range vec.Vec_IntArray(ppis) {
			gia.Gia_ObjTerSimSetX(gia.Gia_ManObj(g, id))
		}
		setValues(selected)
		for _, id := range vec.Vec_IntArray(nodes) {
			o := gia.Gia_ManObj(g, id)
			if gia.Gia_ObjIsAnd(o) != 0 {
				gia.Gia_ObjTerSimAnd(o)
			} else if frame == 0 {
				gia.Gia_ObjTerSimSet0(o)
			} else {
				gia.Gia_ObjTerSimRo(g, o)
			}
		}
		for _, id := range vec.Vec_IntArray(outputs) {
			gia.Gia_ObjTerSimCo(gia.Gia_ManObj(g, id))
		}
	}
	if gia.Gia_ObjTerSimGet1(gia.Gia_ManPo(g, 0)) == 0 {
		fmt.Print("\nRefinement verification has failed!!!\n")
	}
	gia.Gia_ObjTerSimSetC(gia.Gia_ManConst0(g))
	for _, v := range []*vec.Vec_Int_t{pis, ppis, nodes, outputs} {
		for _, id := range vec.Vec_IntArray(v) {
			gia.Gia_ObjTerSimSetC(gia.Gia_ManObj(g, id))
		}
	}
}
