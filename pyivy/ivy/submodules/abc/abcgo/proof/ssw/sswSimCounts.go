package ssw

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
	"math/bits"
)

// Ssw_SmlNodeNotEquWeight counts differing simulation bits outside the prefix.
func Ssw_SmlNodeNotEquWeight(p *Ssw_Sml_t, left, right int32) int32 {
	a, b := Ssw_ObjSim(p, left), Ssw_ObjSim(p, right)
	var count int32
	for i := p.nWordsPref; i < p.nWordsTotal; i++ {
		count += int32(bits.OnesCount32(a[i] ^ b[i]))
	}
	return count
}
func Ssw_SmlCountEqual(p *Ssw_Sml_t, li, lo aig.Edge) int32 {
	if p.pAig.Node(lo).Phase {
		panic("ssw: latch output phase must be zero")
	}
	a, b := Ssw_ObjSim(p, int32(li.NodeID())), Ssw_ObjSim(p, int32(lo.NodeID()))
	var count int32
	for i := p.nWordsPref; i < p.nWordsTotal; i++ {
		count += int32(bits.OnesCount32(^(a[i] ^ b[i])))
	}
	return count
}
func Ssw_SmlNodeIsZero(p *Ssw_Sml_t, obj aig.Edge) int32 { return Ssw_SmlObjIsConstWord(p, obj) }
func Ssw_SmlNodeIsZeroFrame(p *Ssw_Sml_t, obj aig.Edge, frame int32) int32 {
	if Ssw_ObjSim(p, int32(obj.NodeID()))[frame] == 0 {
		return 1
	}
	return 0
}
func Ssw_SmlNodeCountOnesReal(p *Ssw_Sml_t, obj aig.Edge) int32 {
	invert := p.pAig.Node(obj).Phase != obj.Complemented()
	var count int32
	for _, word := range Ssw_ObjSim(p, int32(obj.NodeID())) {
		if invert {
			word = ^word
		}
		count += int32(bits.OnesCount32(word))
	}
	return count
}

// Ssw_SmlNodeCountOnesRealVec counts bits set in the OR of real node values.
func Ssw_SmlNodeCountOnesRealVec(p *Ssw_Sml_t, objects *vec.Vec_Ptr_t) int32 {
	var count int32
	for i := int32(0); i < p.nWordsTotal; i++ {
		var combined uint32
		for j := 0; j < vec.Vec_PtrSize(objects); j++ {
			obj := vec.Vec_PtrEntry(objects, j).(aig.Edge)
			value := Ssw_ObjSim(p, int32(obj.NodeID()))[i]
			if p.pAig.Node(obj).Phase != obj.Complemented() {
				value = ^value
			}
			combined |= value
		}
		count += int32(bits.OnesCount32(combined))
	}
	return count
}
