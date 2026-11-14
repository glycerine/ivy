package abs

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
)

// Ga2_ManAddToAbs adds selected cuts and their input nodes to every frame.
func Ga2_ManAddToAbs(p *Ga2_Man_t, add *vec.Vec_Int_t) {
	for i := 0; i < vec.Vec_IntSize(add); i++ {
		id := vec.Vec_IntEntry(add, i)
		obj := gia.Gia_ManObj(p.pGia, id)
		Ga2_ManSetupNode(p, obj, 1)
		if p.proofActive {
			vec.Vec_IntWriteEntry(p.vProofIds, int(id), p.nProofIds)
			p.nProofIds++
		}
	}
	for i := 0; i < vec.Vec_IntSize(add); i++ {
		obj := gia.Gia_ManObj(p.pGia, vec.Vec_IntEntry(add, i))
		leaves := Ga2_ObjLeaves(p.pGia, obj)
		for k := 0; k < vec.Vec_IntSize(leaves); k++ {
			leaf := gia.Gia_ManObj(p.pGia, vec.Vec_IntEntry(leaves, k))
			if Ga2_ObjId(p, leaf) == -1 {
				Ga2_ManSetupNode(p, leaf, 0)
			}
		}
	}
	for frame := int32(0); frame <= p.pPars.IFrame; frame++ {
		vec.Vec_IntFillExtra(Ga2_MapFrameMap(p, frame), vec.Vec_IntSize(p.vValues), -1)
		for i := 0; i < vec.Vec_IntSize(add); i++ {
			obj := gia.Gia_ManObj(p.pGia, vec.Vec_IntEntry(add, i))
			Ga2_ManAddToAbsOneStatic(p, obj, frame, 1)
		}
	}
}
