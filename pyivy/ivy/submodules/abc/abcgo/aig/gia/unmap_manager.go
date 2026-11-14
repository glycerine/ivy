package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/vec"
	"time"
)

type Unm_Man_t struct {
	pGia, pNew      *Gia_Man_t
	pHash           *vec.Hash_IntMan_t
	nNewSize        int32
	vUsed, vId2Used *Vec_Int_t
	vTruths         *Vec_Wrd_t
	vLeaves         *Vec_Int_t
	clkStart        time.Time
}

func Unm_ManAlloc(g *Gia_Man_t) *Unm_Man_t {
	p := &Unm_Man_t{pGia: g, nNewSize: 3 * g.nObjs / 2, clkStart: time.Now(), pHash: vec.Hash_IntManStart(1000), vLeaves: vec.Vec_IntStart(10)}
	p.pNew = Gia_ManStart(p.nNewSize)
	p.pNew.pName, p.pNew.PSpec = g.pName, g.PSpec
	Gia_ManHashStart(p.pNew)
	Gia_ManCleanLevels(p.pNew, p.nNewSize)
	Gia_ManIncrementTravId(p.pNew)
	return p
}

// Unm_ManFree follows the experimental source, which discards its graph.
func Unm_ManFree(p *Unm_Man_t) *Gia_Man_t {
	Gia_ManHashStop(p.pNew)
	p.pNew.vLevels = nil
	Gia_ManSetRegNum(p.pNew, Gia_ManRegNum(p.pGia))
	vec.Hash_IntManStop(p.pHash)
	p.pNew = nil
	return nil
}
func Unm_ManFindUsedNodes(pairs *Vec_Int_t, nObjs int32) *Vec_Int_t {
	data := vec.Vec_IntArray(pairs)
	if nObjs < 0 || len(data)%3 != 0 {
		panic("gia: invalid unmapping pairs")
	}
	marks := make([]bool, nObjs)
	for i := 0; i < len(data); i += 3 {
		if data[i] < 0 || data[i] >= nObjs {
			panic("gia: invalid unmapping node")
		}
		marks[data[i]] = true
	}
	result := vec.Vec_IntAlloc(1000)
	for id, used := range marks {
		if used {
			vec.Vec_IntPush(result, int32(id))
		}
	}
	fmt.Printf("The number of used nodes = %d\n", vec.Vec_IntSize(result))
	return result
}
func Unm_ManComputeTruths(p *Unm_Man_t) *Vec_Wrd_t {
	result := vec.Vec_WrdAlloc(vec.Vec_IntSize(p.vUsed))
	scratch := vec.Vec_WrdStart(int(p.pGia.nObjs))
	for _, id := range vec.Vec_IntArray(p.vUsed) {
		if Gia_ObjIsLut(p.pGia, id) == 0 {
			panic("gia: unmapping node is not LUT")
		}
		vec.Vec_IntClear(p.vLeaves)
		for _, fan := range Gia_ObjLutFanins(p.pGia, id) {
			vec.Vec_IntPush(p.vLeaves, fan)
		}
		if vec.Vec_IntSize(p.vLeaves) > 6 {
			panic("gia: unmapping LUT exceeds six inputs")
		}
		vec.Vec_WrdPush(result, Shr_ManComputeTruth6(p.pGia, Gia_ManObj(p.pGia, id), p.vLeaves, scratch))
	}
	return result
}

// Unm_ManWork identifies repeated LUT fanin pairs that can be decomposed.
func Unm_ManWork(p *Unm_Man_t) {
	pairs := Unm_ManComputePairs(p, 1)
	p.vUsed = Unm_ManFindUsedNodes(pairs, p.pGia.nObjs)
	p.vId2Used = vec.Vec_IntStartFull(int(p.pGia.nObjs))
	for i, id := range vec.Vec_IntArray(p.vUsed) {
		vec.Vec_IntWriteEntry(p.vId2Used, int(id), int32(i))
	}
	p.vTruths = Unm_ManComputeTruths(p)
	Unm_ManCollectDecomp(p, pairs, 1)
}

// Unm_ManTest runs the experimental analysis; the original returns no graph.
func Unm_ManTest(g *Gia_Man_t) *Gia_Man_t {
	p := Unm_ManAlloc(g)
	Unm_ManWork(p)
	fmt.Printf("Time: %s\n", time.Since(p.clkStart))
	return Unm_ManFree(p)
}
