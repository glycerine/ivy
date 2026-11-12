package gia

import "github.com/glycerine/abcgo/misc/vec"

const (
	MF_CUT_MAX  = 16
	MF_LOG_PAGE = 12
	MF_NO_LEAF  = 31
	MF_TT_WORDS = 1 << (MF_LEAF_MAX - 6)
	mfPageSize  = 1 << 16
)

type Mf_Obj_t struct {
	iCutSet         int32
	Flow, nFlowRefs float32
	Delay, nMapRefs uint32
}
type Mf_Man_t struct {
	pGia0, pGia               *Gia_Man_t
	pPars                     *Jf_Par_t
	pLfObjs                   []Mf_Obj_t
	vPages                    [][]int32
	vTtMem                    *vec.Vec_Mem_t
	vCnfSizes, vCnfMem, vTemp Vec_Int_t
	iCur, Iter, fUseEla       int32
	clkStart                  int64
	CutCount                  [4]float64
	nCutCounts                [MF_LEAF_MAX + 1]int32
}

func Mf_ManObj(p *Mf_Man_t, i int32) *Mf_Obj_t    { return &p.pLfObjs[i] }
func Mf_ManCutSet(p *Mf_Man_t, id int32) []int32  { return p.vPages[uint32(id)>>16][uint32(id)&0xffff:] }
func Mf_ObjCutSet(p *Mf_Man_t, id int32) []int32  { return Mf_ManCutSet(p, Mf_ManObj(p, id).iCutSet) }
func Mf_ObjCutBest(p *Mf_Man_t, id int32) []int32 { return Mf_ObjCutSet(p, id)[1:] }
func Mf_ObjMapRefNum(p *Mf_Man_t, id int32) int32 { return int32(Mf_ManObj(p, id).nMapRefs) }
func Mf_ObjMapRefInc(p *Mf_Man_t, id int32) int32 {
	o := Mf_ManObj(p, id)
	old := o.nMapRefs
	o.nMapRefs = (old + 1) & 0xffff
	return int32(old)
}
func Mf_ObjMapRefDec(p *Mf_Man_t, id int32) int32 {
	o := Mf_ManObj(p, id)
	o.nMapRefs = (o.nMapRefs - 1) & 0xffff
	return int32(o.nMapRefs)
}
func Mf_CutSize(c []int32) int32     { return c[0] & MF_NO_LEAF }
func Mf_CutFunc(c []int32) int32     { return int32(uint32(c[0]) >> 5) }
func Mf_CutSetBoth(n, f int32) int32 { return n | f<<5 }
func Mf_CutIsTriv(c []int32, id int32) int32 {
	if Mf_CutSize(c) == 1 && c[1] == id {
		return 1
	}
	return 0
}

func Mf_ManSaveCuts(p *Mf_Man_t, cuts []*Mf_Cut_t, n int32) int32 {
	size := int32(1)
	for _, c := range cuts[:n] {
		size += int32(c.nLeaves) + 1
	}
	if (p.iCur&0xffff)+size > 0xffff {
		p.iCur = ((p.iCur >> 16) + 1) << 16
	}
	page := int(uint32(p.iCur) >> 16)
	for len(p.vPages) <= page {
		p.vPages = append(p.vPages, make([]int32, mfPageSize))
	}
	id := p.iCur
	p.iCur += size
	dst := Mf_ManCutSet(p, id)
	dst[0] = n
	pos := 1
	for _, c := range cuts[:n] {
		dst[pos] = Mf_CutSetBoth(int32(c.nLeaves), int32(c.iFunc))
		pos++
		pos += copy(dst[pos:], c.pLeaves[:c.nLeaves])
	}
	return id
}
func Mf_ManPrepareCuts(dst []Mf_Cut_t, p *Mf_Man_t, id, addUnit int32) int32 {
	obj := Mf_ManObj(p, id)
	if obj.iCutSet == 0 {
		return Mf_CutCreateUnit(&dst[0], id)
	}
	list := Mf_ManCutSet(p, obj.iCutSet)
	n := list[0]
	pos := 1
	for i := int32(0); i < n; i++ {
		packed := list[pos:]
		size := Mf_CutSize(packed)
		c := &dst[i]
		c.Delay, c.Flow, c.iFunc, c.nLeaves = 0, 0, uint32(Mf_CutFunc(packed)), uint32(size)
		copy(c.pLeaves[:], packed[1:1+size])
		c.Sign = Mf_CutGetSign(c.pLeaves[:], size)
		pos += int(size) + 1
	}
	if addUnit != 0 && dst[0].nLeaves > 1 {
		n += Mf_CutCreateUnit(&dst[n], id)
	}
	return n
}
func Mf_ObjSetBestCut(cuts []int32, offset int) {
	if offset <= 0 || offset >= len(cuts) {
		panic("gia: invalid cut offset")
	}
	if offset == 1 {
		return
	}
	size := int(Mf_CutSize(cuts[offset:])) + 1
	temp := append([]int32(nil), cuts[1:offset]...)
	copy(cuts[1:], cuts[offset:offset+size])
	copy(cuts[1+size:], temp)
}
