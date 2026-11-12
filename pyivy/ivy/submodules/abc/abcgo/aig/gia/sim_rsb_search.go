package gia

import (
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"math/bits"
)

func Gia_RsbFindMints(p *Gia_RsbMan_t, off, on *int32) {
	active := vec.Vec_IntArray(p.vActive)
	if len(active) == 0 {
		panic("gia: no active pattern sets")
	}
	id := active[util.Abc_Random(0)%uint32(len(active))]
	a, b := p.vSets[0][id], p.vSets[1][id]
	*off = a[util.Abc_Random(0)%uint32(len(a))]
	*on = b[util.Abc_Random(0)%uint32(len(b))]
}
func Gia_RsbFindNode(p *Gia_RsbMan_t) int32 {
	copy(p.pSet[1], p.pSet[0])
	for _, id := range vec.Vec_IntArray(p.vObjs) {
		if giaSimBit(p.pSet[1], id) == 0 {
			panic("gia: invalid selected candidate")
		}
		giaSimFlip(p.pSet[1], id)
	}
	copy(p.pSet[2], p.pSet[1])
	data := vec.Vec_WrdArray(p.vSimsT)
	intersect := func() int {
		var a, b int32
		Gia_RsbFindMints(p, &a, &b)
		count := 0
		for w := int32(0); w < p.nWordsT; w++ {
			p.pSet[1][w] &= data[a*p.nWordsT+w] ^ data[b*p.nWordsT+w]
			count += bits.OnesCount64(p.pSet[1][w])
		}
		return count
	}
	count := intersect()
	for i := 0; i < 5 && count > 1; i++ {
		old := count
		copy(p.pSet[2], p.pSet[1])
		count = intersect()
		if count < old {
			i = 0
		}
	}
	size := 64 * p.nWordsT
	shift := int32(util.Abc_Random(0) % uint32(size))
	for i := int32(0); i < size; i++ {
		id := (i + shift) % size
		if giaSimBit(p.pSet[2], id) != 0 {
			return id
		}
	}
	panic("gia: candidate signals cannot distinguish sampled patterns")
}
func Gia_RsbCollectValid(p *Gia_RsbMan_t) int32 {
	vec.Vec_IntClear(p.vActive)
	for i := range p.vSets[0] {
		if len(p.vSets[0][i]) != 0 && len(p.vSets[1][i]) != 0 {
			vec.Vec_IntPush(p.vActive, int32(i))
		}
	}
	return giaBool(vec.Vec_IntSize(p.vActive) != 0)
}
func Gia_RsbSolve(p *Gia_RsbMan_t) *Vec_Int_t {
	vec.Vec_IntClear(p.vObjs)
	for Gia_RsbCollectValid(p) != 0 {
		Gia_RsbUpdateAdd(p, Gia_RsbFindNode(p))
	}
	if vec.Vec_IntSize(p.vObjs) == 0 {
		return vec.Vec_IntAlloc(0)
	}
	for i := 0; i < 100; i++ {
		undo := 1 + int(util.Abc_Random(0)%uint32(vec.Vec_IntSize(p.vObjs)))
		for k := 0; k < undo; k++ {
			Gia_RsbUpdateRemove(p, Gia_RsbFindNodeToRemove(p, nil))
		}
		for Gia_RsbCollectValid(p) != 0 {
			Gia_RsbUpdateAdd(p, Gia_RsbFindNode(p))
		}
		if vec.Vec_IntSize(p.vObjs2) == 0 || vec.Vec_IntSize(p.vObjs2) > vec.Vec_IntSize(p.vObjs) {
			vec.Vec_IntClear(p.vObjs2)
			vec.Vec_IntAppend(p.vObjs2, p.vObjs)
		}
	}
	return vec.Vec_IntDup(p.vObjs2)
}
func Gia_RsbSetFind(off, on []uint64, sims *Vec_Wrd_t, words int32, transposed *Vec_Wrd_t, wordsT int32, cands *Vec_Int_t) *Vec_Int_t {
	p := Gia_RsbAlloc(nil, off, on, sims, words, transposed, wordsT, cands)
	defer Gia_RsbFree(p)
	result := Gia_RsbSolve(p)
	vec.Vec_IntSort(result, 0)
	return result
}
