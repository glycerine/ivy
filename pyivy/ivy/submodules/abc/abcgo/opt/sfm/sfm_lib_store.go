package sfm

import (
	"github.com/glycerine/abcgo/map/mio"
	"github.com/glycerine/abcgo/misc/vec"
)

type Sfm_Fun_t struct {
	Next, Area     int32
	pFansT, pFansB [SFM_SUPP_MAX + 1]byte
}
type Sfm_Lib_t struct {
	nVars, nWords            int32
	fVerbose, fDelay         int32
	pCells                   []mio.Mio_Cell2_t
	pObjs                    []Sfm_Fun_t
	vTtMem                   *vec.Vec_Mem_t
	lists, counts, hits      []int32
	profiles                 [][]int32
	nObjSkipped, nObjRemoved int32
}

func Sfm_LibStart(nVars, fDelay, fVerbose int32) *Sfm_Lib_t {
	if nVars < 0 || nVars > SFM_SUPP_MAX {
		panic("sfm: invalid support size")
	}
	words := int32(1)
	if nVars > 6 {
		words = 1 << uint(nVars-6)
	}
	return &Sfm_Lib_t{nVars: nVars, nWords: words, fDelay: fDelay, fVerbose: fVerbose, vTtMem: vec.Vec_MemAllocForTT(int(nVars), 0), lists: []int32{-1, -1}, counts: []int32{-1, -1}, hits: []int32{-1, -1}}
}
func Sfm_LibStop(p *Sfm_Lib_t) {
	if p == nil {
		return
	}
	vec.Vec_MemHashFree(p.vTtMem)
	vec.Vec_MemFree(p.vTtMem)
	p.vTtMem = nil
	p.pObjs = nil
	p.pCells = nil
	p.lists = nil
	p.counts = nil
	p.hits = nil
	p.profiles = nil
}
func Sfm_LibFun(p *Sfm_Lib_t, id int32) *Sfm_Fun_t {
	if id < 0 {
		return nil
	}
	return &p.pObjs[id]
}
func Sfm_LibPrepareAdd(p *Sfm_Lib_t, truth []uint64, permutation []int32, bottom, top *mio.Mio_Cell2_t, pin int32) {
	if bottom == nil || len(permutation) > SFM_SUPP_MAX {
		panic("sfm: invalid implementation")
	}
	id := vec.Vec_MemHashInsert(p.vTtMem, truth)
	for len(p.lists) <= id {
		p.lists = append(p.lists, -1)
		p.counts = append(p.counts, 0)
		p.hits = append(p.hits, 0)
	}
	area := int32(bottom.AreaW)
	if top != nil {
		area += int32(top.AreaW)
	}
	var profile []int32
	if p.fDelay != 0 {
		profile = Sfm_LibCellProfile(bottom, top, permutation)
		var retained []int32
		for i := p.lists[id]; i >= 0; i = p.pObjs[i].Next {
			o := &p.pObjs[i]
			if Sfm_LibNewIsContained(o.Area, p.profiles[i], area, profile) {
				p.nObjSkipped++
				return
			}
			retained = append(retained, i)
		}
		kept := retained[:0]
		for _, i := range retained {
			if Sfm_LibNewContains(p.pObjs[i].Area, p.profiles[i], area, profile) {
				p.nObjRemoved++
			} else {
				kept = append(kept, i)
			}
		}
		p.lists[id] = -1
		if len(kept) > 0 {
			p.lists[id] = kept[0]
			for i := 0; i < len(kept)-1; i++ {
				p.pObjs[kept[i]].Next = kept[i+1]
			}
			p.pObjs[kept[len(kept)-1]].Next = -1
		}
	} else {
		for i := p.lists[id]; i >= 0; i = p.pObjs[i].Next {
			if area >= p.pObjs[i].Area {
				return
			}
		}
	}
	inverse := make([]int32, len(permutation))
	for i, v := range permutation {
		if v < 0 || int(v) >= len(inverse) {
			panic("sfm: invalid pin permutation")
		}
		inverse[v] = int32(i)
	}
	if bottom.Id >= 128 || top != nil && top.Id >= 128 {
		panic("sfm: gate index overflow")
	}
	obj := Sfm_Fun_t{Area: area, Next: p.lists[id]}
	obj.pFansB[0] = byte(bottom.Id)
	for k := int32(0); k < bottom.Fanins(); k++ {
		obj.pFansB[k+1] = byte(inverse[k])
	}
	if top != nil {
		obj.pFansT[0] = byte(top.Id)
		k := bottom.Fanins()
		for i := int32(0); i < top.Fanins(); i++ {
			if i == pin {
				obj.pFansT[i+1] = 16
			} else {
				obj.pFansT[i+1] = byte(inverse[k])
				k++
			}
		}
		if int(k) != len(permutation) {
			panic("sfm: invalid composed fanins")
		}
	}
	p.lists[id] = int32(len(p.pObjs))
	p.pObjs = append(p.pObjs, obj)
	if p.fDelay != 0 {
		p.profiles = append(p.profiles, profile)
	}
	p.counts[id]++
}
func Sfm_LibFunId(p *Sfm_Lib_t, obj *Sfm_Fun_t) int32 {
	for i := range p.pObjs {
		if &p.pObjs[i] == obj {
			return int32(i)
		}
	}
	return -1
}
