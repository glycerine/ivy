package gia

import (
	"github.com/glycerine/abcgo/aig/miniaig"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

// Gia_ManFindComplLuts marks AND roots used only by complemented outputs.
func Gia_ManFindComplLuts(g *Gia_Man_t) *Vec_Bit_t {
	marks := vec.Vec_BitStart(int(g.nObjs))
	for _, id := range vec.Vec_IntArray(g.vCos) {
		o := g.pObjs[id]
		f := Gia_ObjFanin0(o)
		if Gia_ObjIsAnd(f) != 0 && Gia_ObjFaninC0(o) != 0 {
			vec.Vec_BitWriteEntry(marks, int(f.id), 1)
		}
	}
	for _, id := range vec.Vec_IntArray(g.vCos) {
		o := g.pObjs[id]
		f := Gia_ObjFanin0(o)
		if Gia_ObjIsAnd(f) != 0 && Gia_ObjFaninC0(o) == 0 {
			vec.Vec_BitWriteEntry(marks, int(f.id), 0)
		}
	}
	return marks
}

// Gia_ManToMiniLut exports the LUT mapping, accounting for output polarity.
func Gia_ManToMiniLut(g *Gia_Man_t) *miniaig.Mini_Lut_t {
	if Gia_ManHasMapping(g) == 0 {
		panic("gia: MiniLUT export needs a mapping")
	}
	size := max(Gia_ManLutSizeMax(g), 2)
	if size > 16 {
		panic("gia: MiniLUT supports at most 16 inputs")
	}
	p := miniaig.Mini_LutStart(int(size))
	Gia_ManFillValue(g)
	g.pObjs[0].Value = 0
	for _, id := range vec.Vec_IntArray(g.vCis) {
		g.pObjs[id].Value = uint32(miniaig.Mini_LutCreatePi(p))
	}
	marks := Gia_ManFindComplLuts(g)
	Gia_ObjComputeTruthTableStart(g, size)
	defer Gia_ObjComputeTruthTableStop(g)
	nwords := max(1, 1<<uint(max(0, size-6)))
	for id := int32(1); id < g.nObjs; id++ {
		if Gia_ObjIsLut(g, id) == 0 {
			continue
		}
		leaves := append([]int32(nil), Gia_ObjLutFanins(g, id)...)
		if len(leaves) > 6 {
			extra := len(leaves) - 6
			rotated := append([]int32(nil), leaves[extra:]...)
			for k := extra - 1; k >= 0; k-- {
				rotated = append(rotated, leaves[k])
			}
			leaves = rotated
		}
		v := vec.Vec_IntAlloc(len(leaves))
		vars := make([]int, len(leaves))
		for i, leaf := range leaves {
			vec.Vec_IntPush(v, leaf)
			vars[i] = int(g.pObjs[leaf].Value)
		}
		truth := append([]uint64(nil), Gia_ObjComputeTruthTableCut(g, g.pObjs[id], v)...)
		if vec.Vec_BitEntry(marks, int(id)) != 0 {
			util.Abc_TtNot(truth, int32(nwords))
		}
		for i, leaf := range leaves {
			if vec.Vec_BitEntry(marks, int(leaf)) != 0 {
				util.Abc_TtFlip(truth, int32(nwords), int32(i))
			}
		}
		words := make([]uint32, miniaig.Mini_LutWordNum(int(size)))
		for i := range words {
			words[i] = uint32(truth[i/2] >> uint(32*(i%2)))
		}
		g.pObjs[id].Value = uint32(miniaig.Mini_LutCreateNode(p, len(leaves), vars, words))
	}
	inverters := make(map[int]int)
	for _, id := range vec.Vec_IntArray(g.vCos) {
		o := g.pObjs[id]
		f := Gia_ObjFanin0(o)
		var source int
		if f.id == 0 {
			source = int(Gia_ObjFaninC0(o))
		} else {
			source = int(f.Value)
			if int(Gia_ObjFaninC0(o)) != vec.Vec_BitEntry(marks, int(f.id)) {
				inv, ok := inverters[source]
				if !ok {
					words := make([]uint32, miniaig.Mini_LutWordNum(int(size)))
					for i := range words {
						words[i] = 0x55555555
					}
					inv = miniaig.Mini_LutCreateNode(p, 1, []int{source}, words)
					inverters[source] = inv
				}
				source = inv
			}
		}
		o.Value = uint32(miniaig.Mini_LutCreatePo(p, source))
	}
	miniaig.Mini_LutSetRegNum(p, int(Gia_ManRegNum(g)))
	return p
}

func Gia_ManToMiniLutAttr(g *Gia_Man_t, p *miniaig.Mini_Lut_t) []byte {
	attrs := make([]byte, miniaig.Mini_LutNodeNum(p))
	for id := int32(1); id < g.nObjs; id++ {
		if Gia_ObjIsLut(g, id) != 0 && Gia_ObjLutIsMux(g, id) != 0 {
			attrs[g.pObjs[id].Value] = 1
		}
	}
	return attrs
}
