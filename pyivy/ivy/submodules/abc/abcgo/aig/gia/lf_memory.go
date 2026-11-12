package gia

import "github.com/glycerine/abcgo/misc/vec"

// Lf_Mem_t stores delta-encoded cut leaves in reusable fixed-sized pages.
type Lf_Mem_t struct {
	LogPage, MaskPage, nCutWords, iCur int32
	vPages                             Vec_Ptr_t
	vFree                              *Vec_Ptr_t
}

func Lf_MemAlloc(p *Lf_Mem_t, logPage int32, reusable *Vec_Ptr_t, cutWords int32) {
	if logPage < 7 || logPage > 29 || reusable == nil {
		panic("LF invalid cut page configuration")
	}
	*p = Lf_Mem_t{LogPage: logPage, MaskPage: (1 << uint(logPage)) - 1, nCutWords: cutWords, vFree: reusable}
}
func Lf_MemSaveCut(p *Lf_Mem_t, cut *Lf_Cut_t, obj int32) int32 {
	if cut.fMux7 != 0 || cut.nLeaves > LF_LEAF_MAX || (cut.nLeaves < 2 && cut.iFunc > 3) {
		panic("LF invalid stored cut")
	}
	handle := p.iCur
	page := int(handle >> uint(p.LogPage))
	if page == vec.Vec_PtrSize(&p.vPages) {
		var data []byte
		if vec.Vec_PtrSize(p.vFree) > 0 {
			data = vec.Vec_PtrPop(p.vFree).([]byte)
		} else {
			data = make([]byte, int(p.MaskPage)+1)
		}
		if len(data) != int(p.MaskPage)+1 {
			panic("LF incompatible reusable page")
		}
		vec.Vec_PtrPush(&p.vPages, data)
	}
	pos := handle & p.MaskPage
	if p.MaskPage-pos < 4*(LF_LEAF_MAX+2) {
		panic("LF cut too close to page end")
	}
	data := vec.Vec_PtrEntry(&p.vPages, page).([]byte)
	pos = Gia_AigerWriteUnsignedBuffer(data, pos, cut.nLeaves)
	prev := obj
	for i := int(cut.nLeaves) - 1; i >= 0; i-- {
		diff := prev - cut.pLeaves[i]
		if diff < 0 {
			panic("LF leaves not in ascending object order")
		}
		pos = Gia_AigerWriteUnsignedBuffer(data, pos, uint32(diff))
		prev = cut.pLeaves[i]
	}
	if cut.iFunc >= 0 {
		pos = Gia_AigerWriteUnsignedBuffer(data, pos, uint32(cut.iFunc))
	}
	if p.MaskPage-(pos&p.MaskPage) < 4*(LF_LEAF_MAX+2) {
		p.iCur = (int32(page) + 1) << uint(p.LogPage)
	} else {
		p.iCur = (int32(page) << uint(p.LogPage)) | pos
	}
	return handle
}
func Lf_MemLoadCut(p *Lf_Mem_t, handle, obj int32, cut *Lf_Cut_t, truth, recycle int32) *Lf_Cut_t {
	page := int(handle >> uint(p.LogPage))
	if handle < 0 || page >= vec.Vec_PtrSize(&p.vPages) {
		panic("LF invalid cut handle")
	}
	data := vec.Vec_PtrEntry(&p.vPages, page).([]byte)
	input := data[handle&p.MaskPage:]
	cut.nLeaves = Gia_AigerReadUnsigned(&input)
	if cut.nLeaves > LF_LEAF_MAX {
		panic("LF invalid cut size")
	}
	prev := obj
	for i := int(cut.nLeaves) - 1; i >= 0; i-- {
		prev -= int32(Gia_AigerReadUnsigned(&input))
		cut.pLeaves[i] = prev
	}
	cut.iFunc = -1
	if truth != 0 {
		cut.iFunc = int32(Gia_AigerReadUnsigned(&input))
	}
	if cut.nLeaves < 2 && cut.iFunc > 3 {
		panic("LF invalid cut function")
	}
	if recycle != 0 && page > 0 {
		previous := vec.Vec_PtrEntry(&p.vPages, page-1)
		if previous != nil {
			vec.Vec_PtrPush(p.vFree, previous)
			vec.Vec_PtrWriteEntry(&p.vPages, page-1, nil)
		}
	}
	cut.Sign = 0
	if recycle != 0 {
		cut.Sign = Lf_CutGetSign(cut)
	}
	cut.fMux7 = 0
	return cut
}
func Lf_MemRecycle(p *Lf_Mem_t) {
	for i := 0; i < vec.Vec_PtrSize(&p.vPages); i++ {
		if page := vec.Vec_PtrEntry(&p.vPages, i); page != nil {
			vec.Vec_PtrPush(p.vFree, page)
		}
	}
	vec.Vec_PtrClear(&p.vPages)
	p.iCur = 0
}
