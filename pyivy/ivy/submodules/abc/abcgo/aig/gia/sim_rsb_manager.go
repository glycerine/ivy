package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

type Gia_RsbMan_t struct {
	pGia                  *Gia_Man_t
	pOffSet, pOnSet       []uint64
	nWords, nWordsT       int32
	vSims, vSimsT         *Vec_Wrd_t
	vCands, vObjs, vObjs2 *Vec_Int_t
	vSets                 [2][][]int32
	pSet                  [3][]uint64
	vActive               *Vec_Int_t
}

func Gia_RsbAlloc(g *Gia_Man_t, off, on []uint64, sims *Vec_Wrd_t, words int32, transposed *Vec_Wrd_t, wordsT int32, cands *Vec_Int_t) *Gia_RsbMan_t {
	if words <= 0 || words > 1024 || wordsT <= 0 || len(off) < int(words) || len(on) < int(words) || vec.Vec_WrdSize(sims) != int(64*words*wordsT) || vec.Vec_WrdSize(sims) != vec.Vec_WrdSize(transposed) {
		panic("gia: invalid resubstitution dimensions")
	}
	p := &Gia_RsbMan_t{pGia: g, pOffSet: off, pOnSet: on, nWords: words, nWordsT: wordsT, vSims: sims, vSimsT: transposed, vCands: cands, vObjs: vec.Vec_IntAlloc(100), vObjs2: vec.Vec_IntAlloc(100), vActive: vec.Vec_IntAlloc(100)}
	for i := range p.pSet {
		p.pSet[i] = make([]uint64, wordsT)
	}
	for _, id := range vec.Vec_IntArray(cands) {
		if id < 0 || id >= 64*wordsT {
			panic("gia: invalid candidate")
		}
		p.pSet[0][id>>6] |= uint64(1) << uint(id&63)
	}
	p.vSets[0] = make([][]int32, 1)
	p.vSets[1] = make([][]int32, 1)
	for bit := int32(0); bit < 64*words; bit++ {
		a, b := giaSimBit(off, bit), giaSimBit(on, bit)
		if a != 0 && b != 0 {
			panic("gia: overlapping on and off sets")
		}
		if a != 0 {
			p.vSets[0][0] = append(p.vSets[0][0], bit)
		} else if b != 0 {
			p.vSets[1][0] = append(p.vSets[1][0], bit)
		}
	}
	util.Abc_Random(1)
	return p
}
func Gia_RsbFree(p *Gia_RsbMan_t) {
	p.vActive = nil
	p.vObjs = nil
	p.vObjs2 = nil
	p.vSets = [2][][]int32{}
	p.pSet = [3][]uint64{}
}
func Gia_RsbCost(p *Gia_RsbMan_t) int32 {
	var cost int32
	for i, off := range p.vSets[0] {
		cost += int32(len(off) * len(p.vSets[1][i]))
	}
	return cost
}
func Gia_RsbPrint(p *Gia_RsbMan_t) {
	fmt.Printf("Database for %d objects and cost %d:\n", vec.Vec_IntSize(p.vObjs), Gia_RsbCost(p))
	for i := range p.vSets[0] {
		for v := 0; v < 2; v++ {
			fmt.Printf("%5d : %b %d %v\n", i, i, v, p.vSets[v][i])
		}
	}
}
func Gia_RsbUpdateAdd(p *Gia_RsbMan_t, id int32) {
	if id < 0 || id >= 64*p.nWordsT {
		panic("gia: invalid resubstitution candidate")
	}
	data := vec.Vec_WrdArray(p.vSims)[id*p.nWords : (id+1)*p.nWords]
	leaves := len(p.vSets[0])
	for side := 0; side < 2; side++ {
		sets := p.vSets[side]
		sets = append(sets, make([][]int32, leaves)...)
		for i := 0; i < leaves; i++ {
			old := sets[i]
			zero := old[:0]
			var one []int32
			for _, pat := range old {
				if giaSimBit(data, pat) != 0 {
					one = append(one, pat)
				} else {
					zero = append(zero, pat)
				}
			}
			sets[i], sets[leaves+i] = zero, one
		}
		p.vSets[side] = sets
	}
	vec.Vec_IntPush(p.vObjs, id)
}
func Gia_RsbUpdateRemove(p *Gia_RsbMan_t, index int32) {
	n := vec.Vec_IntSize(p.vObjs)
	if index < 0 || int(index) >= n {
		panic("gia: invalid support index")
	}
	leaves := len(p.vSets[0])
	bit := 1 << uint(index)
	for side := 0; side < 2; side++ {
		old := p.vSets[side]
		next := make([][]int32, 0, leaves/2)
		for i := 0; i < leaves; i++ {
			if i&bit == 0 {
				merged := append(old[i], old[i|bit]...)
				next = append(next, merged)
			}
		}
		p.vSets[side] = next
	}
	vec.Vec_IntDrop(p.vObjs, int(index))
}
func Gia_RsbRemovalCost(p *Gia_RsbMan_t, index int32) int32 {
	if index < 0 || int(index) >= vec.Vec_IntSize(p.vObjs) {
		panic("gia: invalid support index")
	}
	bit := 1 << uint(index)
	var cost int32
	for i := 0; i < len(p.vSets[0]); i++ {
		if i&bit == 0 {
			cost += int32((len(p.vSets[0][i]) + len(p.vSets[0][i|bit])) * (len(p.vSets[1][i]) + len(p.vSets[1][i|bit])))
		}
	}
	return cost
}
func Gia_RsbFindNodeToRemove(p *Gia_RsbMan_t, minCost *int32) int32 {
	best, cost := int32(-1), int32(0x7fffffff)
	for i := 0; i < vec.Vec_IntSize(p.vObjs); i++ {
		v := Gia_RsbRemovalCost(p, int32(i))
		if v < cost {
			best, cost = int32(i), v
		}
	}
	if minCost != nil {
		*minCost = cost
	}
	return best
}
