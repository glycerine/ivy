package gia

import "github.com/glycerine/abcgo/misc/vec"

// Lf_ManFetchSet keeps cut sets live until the final fanout consumes them.
func Lf_ManFetchSet(p *Lf_Man_t, id int32) []Lf_Cut_t {
	obj := Gia_ManObj(p.pGia, id)
	offset := Lf_ObjOff(p, id)
	if Gia_ObjIsAndNotBuf(obj) == 0 || obj.Value == 0 {
		panic("LF invalid cut-set access")
	}
	entry := vec.Vec_IntEntry(&p.vCutSets, int(offset))
	if entry == -1 {
		if vec.Vec_IntSize(&p.vFreeSets) == 0 {
			page := make([][]Lf_Cut_t, 1<<LF_LOG_PAGE)
			shift := int32(vec.Vec_PtrSize(&p.vMemSets)) << LF_LOG_PAGE
			vec.Vec_PtrPush(&p.vMemSets, page)
			for i := int32((1 << LF_LOG_PAGE) - 1); i >= 0; i-- {
				vec.Vec_IntPush(&p.vFreeSets, shift|i)
			}
		}
		entry = vec.Vec_IntPop(&p.vFreeSets)
		vec.Vec_IntWriteEntry(&p.vCutSets, int(offset), entry)
		p.nFrontMax = max(p.nFrontMax, entry+1)
	} else if obj.Value--; obj.Value == 0 {
		vec.Vec_IntPush(&p.vFreeSets, entry)
		vec.Vec_IntWriteEntry(&p.vCutSets, int(offset), -1)
	}
	page := vec.Vec_PtrEntry(&p.vMemSets, int(entry>>LF_LOG_PAGE)).([][]Lf_Cut_t)
	index := entry & ((1 << LF_LOG_PAGE) - 1)
	if page[index] == nil {
		page[index] = make([]Lf_Cut_t, p.pPars.NCutNum)
		for i := range page[index] {
			page[index][i].nLeaves = LF_NO_LEAF
		}
	}
	return page[index]
}
func Lf_ManPrepareSet(p *Lf_Man_t, id, index int32, out *[]Lf_Cut_t) int32 {
	if Lf_ObjOff(p, id) == -1 {
		*out = make([]Lf_Cut_t, 1)
		return Lf_CutCreateUnit(&(*out)[0], id)
	}
	*out = Lf_ManFetchSet(p, id)
	for i := int32(0); i < p.pPars.NCutNum; i++ {
		if (*out)[i].nLeaves == LF_NO_LEAF {
			return i
		}
	}
	return p.pPars.NCutNum
}
