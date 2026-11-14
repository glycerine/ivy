package gia

import (
	"github.com/glycerine/abcgo/bool/kit"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

func Lf_ManDeriveMapping(p *Lf_Man_t) *Gia_Man_t {
	g := p.pGia
	if p.pPars.FCutMin != 0 || g.VMapping != nil {
		panic("LF invalid structural mapping state")
	}
	mapping := vec.Vec_IntStart(int(g.nObjs))
	for _, obj := range g.pObjs {
		if Gia_ObjIsAndNotBuf(obj) == 0 || Lf_ObjMapRefNum(p, obj.id) == 0 {
			continue
		}
		cut := Lf_ObjCutBest(p, obj.id)
		if cut.fMux7 != 0 {
			panic("LF unexpected MUX cut")
		}
		vec.Vec_IntWriteEntry(mapping, int(obj.id), int32(vec.Vec_IntSize(mapping)))
		vec.Vec_IntPush(mapping, int32(cut.nLeaves))
		for _, leaf := range cut.pLeaves[:cut.nLeaves] {
			vec.Vec_IntPush(mapping, leaf)
		}
		vec.Vec_IntPush(mapping, obj.id)
	}
	g.VMapping = mapping
	return g
}
func Lf_ManDeriveMappingCoarse(p *Lf_Man_t) *Gia_Man_t {
	g := p.pGia
	if p.pPars.FCutMin != 0 || g.PMuxes == nil {
		panic("LF invalid coarse mapping state")
	}
	dst := Gia_ManStart(g.nObjs)
	dst.pName = g.pName
	dst.PSpec = g.PSpec
	copies := make([]int32, g.nObjs)
	copies[0] = 0
	type record struct {
		id     int32
		leaves []int32
		mux    bool
	}
	var records []record
	for _, obj := range g.pObjs[1:] {
		id := obj.id
		if Gia_ObjIsCi(obj) != 0 {
			copies[id] = Gia_ManAppendCi(dst)
			continue
		}
		a := copies[Gia_ObjFanin0(obj).id] ^ Gia_ObjFaninC0(obj)
		if Gia_ObjIsCo(obj) != 0 {
			copies[id] = Gia_ManAppendCo(dst, a)
			continue
		}
		if Gia_ObjIsBuf(obj) != 0 {
			copies[id] = Gia_ManAppendBuf(dst, a)
			continue
		}
		b := copies[Gia_ObjFanin1(obj).id] ^ Gia_ObjFaninC1(obj)
		if Gia_ObjIsMuxId(g, id) != 0 {
			ctrl := copies[Gia_ObjFaninId2(g, id)] ^ Gia_ObjFaninC2(g, obj)
			copies[id] = Gia_ManAppendMux(dst, ctrl, b, a)
		} else if Gia_ObjIsXor(obj) != 0 {
			copies[id] = Gia_ManAppendXor(dst, a, b)
		} else {
			copies[id] = Gia_ManAppendAnd(dst, a, b)
		}
		if Lf_ObjMapRefNum(p, id) == 0 {
			continue
		}
		cut := Lf_ObjCutBest(p, id)
		rec := record{id: copies[id] >> 1, mux: cut.fMux7 != 0}
		for _, leaf := range cut.pLeaves[:cut.nLeaves] {
			rec.leaves = append(rec.leaves, copies[leaf]>>1)
		}
		records = append(records, rec)
	}
	mapping := vec.Vec_IntStart(int(dst.nObjs))
	for _, rec := range records {
		vec.Vec_IntWriteEntry(mapping, int(rec.id), int32(vec.Vec_IntSize(mapping)))
		vec.Vec_IntPush(mapping, int32(len(rec.leaves)))
		for _, leaf := range rec.leaves {
			vec.Vec_IntPush(mapping, leaf)
		}
		marker := rec.id
		if rec.mux {
			marker = -marker
		}
		vec.Vec_IntPush(mapping, marker)
	}
	dst.VMapping = mapping
	Gia_ManSetRegNum(dst, Gia_ManRegNum(g))
	return dst
}
func lfTruthWords(t []word) []uint32 {
	words := make([]uint32, 2*len(t))
	for i, w := range t {
		words[2*i], words[2*i+1] = uint32(w), uint32(w>>32)
	}
	return words
}
func lfRecordMapping(mapping, records *Vec_Int_t, lit int32, leaves []int32, mux bool) {
	id := lit >> 1
	vec.Vec_IntSetEntry(mapping, int(id), int32(vec.Vec_IntSize(records)))
	vec.Vec_IntPush(records, int32(len(leaves)))
	for _, leaf := range leaves {
		vec.Vec_IntPush(records, leaf>>1)
	}
	if mux {
		id = -id
	}
	vec.Vec_IntPush(records, id)
}
func Lf_ManDerivePart(p *Lf_Man_t, dst *Gia_Man_t, mapping, records, copies *Vec_Int_t, cut *Lf_Cut_t, leaves, cover *Vec_Int_t, obj *Gia_Obj_t) int32 {
	truth := Lf_CutTruth(p, cut)
	var scratch []uint32
	synth := func(t []word, ls []int32) int32 {
		return kit.Kit_TruthToGia(dst, lfTruthWords(t), int32(len(ls)), &scratch, ls, 0)
	}
	if p.pPars.NLutSizeMux != 0 && p.pPars.NLutSizeMux == int32(cut.nLeaves) {
		lut := p.pPars.NLutSize
		nWords := int32(len(truth))
		variable := Lf_ManFindCofVar(truth, nWords, int32(cut.nLeaves))
		if variable < 0 || variable >= int32(cut.nLeaves) {
			panic("LF missing cofactor variable")
		}
		var cofs [2]int32
		for phase := 0; phase < 2; phase++ {
			table := make([]word, nWords)
			vars := make([]int32, cut.nLeaves)
			for i := range vars {
				vars[i] = int32(i)
			}
			if phase == 0 {
				util.Abc_TtCofactor0p(table, truth, nWords, variable)
			} else {
				util.Abc_TtCofactor1p(table, truth, nWords, variable)
			}
			n := util.Abc_TtMinBase(table, vars, int32(cut.nLeaves), lut)
			if n <= 0 {
				panic("LF constant MUX cofactor")
			}
			ls := make([]int32, n)
			for i := range ls {
				ls[i] = vec.Vec_IntEntry(copies, int(cut.pLeaves[vars[i]]))
			}
			cofs[phase] = synth(table, ls)
			lfRecordMapping(mapping, records, cofs[phase], ls, false)
		}
		ls := []int32{cofs[0], cofs[1], vec.Vec_IntEntry(copies, int(cut.pLeaves[variable]))}
		lit := synth([]word{0xcacacacacacacaca}, ls)
		lfRecordMapping(mapping, records, lit, ls, true)
		return lit
	}
	ls := make([]int32, cut.nLeaves)
	for i, leaf := range cut.pLeaves[:cut.nLeaves] {
		ls[i] = vec.Vec_IntEntry(copies, int(leaf))
	}
	if cut.fMux7 != 0 {
		if cut.nLeaves != 3 {
			panic("LF invalid MUX cut")
		}
		ls[0] ^= Gia_ObjFaninC0(obj)
		ls[1] ^= Gia_ObjFaninC1(obj)
		ls[2] ^= Gia_ObjFaninC2(p.pGia, obj)
	}
	lit := synth(truth, ls)
	lfRecordMapping(mapping, records, lit, ls, cut.fMux7 != 0)
	return lit
}
func Lf_ManDeriveMappingGia(p *Lf_Man_t) *Gia_Man_t {
	if p.pPars.FCutMin == 0 {
		panic("LF requires minimized cuts")
	}
	g := p.pGia
	dst := Gia_ManStart(g.nObjs)
	dst.pName = g.pName
	dst.PSpec = g.PSpec
	copies := vec.Vec_IntStartFull(int(g.nObjs))
	vec.Vec_IntWriteEntry(copies, 0, 0)
	mapping := vec.Vec_IntStart(2 * int(g.nObjs))
	records := vec.Vec_IntStart(1)
	leaves, cover := vec.Vec_IntAlloc(16), vec.Vec_IntAlloc(1024)
	for _, obj := range g.pObjs[1:] {
		id := obj.id
		if Gia_ObjIsCi(obj) != 0 {
			vec.Vec_IntWriteEntry(copies, int(id), Gia_ManAppendCi(dst))
			continue
		}
		if Gia_ObjIsCo(obj) != 0 || Gia_ObjIsBuf(obj) != 0 {
			lit := vec.Vec_IntEntry(copies, int(Gia_ObjFanin0(obj).id)) ^ Gia_ObjFaninC0(obj)
			if Gia_ObjIsCo(obj) != 0 {
				Gia_ManAppendCo(dst, lit)
			} else {
				vec.Vec_IntWriteEntry(copies, int(id), Gia_ManAppendBuf(dst, lit))
			}
			continue
		}
		if Lf_ObjMapRefNum(p, id) == 0 {
			continue
		}
		cut := Lf_ObjCutBest(p, id)
		if cut.iFunc < 0 {
			panic("LF missing cut truth")
		}
		if cut.nLeaves == 0 {
			if cut.iFunc>>1 != 0 {
				panic("LF invalid constant cut")
			}
			vec.Vec_IntWriteEntry(copies, int(id), cut.iFunc)
			continue
		}
		if cut.nLeaves == 1 {
			if cut.iFunc>>1 != 1 {
				panic("LF invalid unit cut")
			}
			lit := vec.Vec_IntEntry(copies, int(cut.pLeaves[0])) ^ (cut.iFunc & 1)
			vec.Vec_IntWriteEntry(copies, int(id), lit)
			continue
		}
		lit := Lf_ManDerivePart(p, dst, mapping, records, copies, cut, leaves, cover, obj)
		vec.Vec_IntWriteEntry(copies, int(id), lit^(cut.iFunc&1))
	}
	if vec.Vec_IntSize(mapping) > int(dst.nObjs) {
		vec.Vec_IntShrink(mapping, int(dst.nObjs))
	} else {
		vec.Vec_IntFillExtra(mapping, int(dst.nObjs), 0)
	}
	for i, v := range vec.Vec_IntArray(mapping) {
		if v > 0 {
			vec.Vec_IntWriteEntry(mapping, i, v+dst.nObjs)
		}
	}
	vec.Vec_IntAppend(mapping, records)
	dst.VMapping = mapping
	Gia_ManSetRegNum(dst, Gia_ManRegNum(g))
	return dst
}
