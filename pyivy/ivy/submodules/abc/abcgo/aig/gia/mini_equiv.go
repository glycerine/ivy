package gia

import (
	"github.com/glycerine/abcgo/aig/miniaig"
	"github.com/glycerine/abcgo/misc/vec"
)

// Gia_ManMapEquivAfterScorr maps GIA representatives back to MiniAIG IDs.
func Gia_ManMapEquivAfterScorr(g *Gia_Man_t, mapping *Vec_Int_t) *Vec_Int_t {
	result := vec.Vec_IntStartFull(vec.Vec_IntSize(mapping))
	first := vec.Vec_IntStartFull(int(g.nObjs))
	for i, lit := range vec.Vec_IntArray(mapping) {
		if lit < 0 {
			continue
		}
		repr := Gia_ObjReprSelf(g, lit>>1)
		prev := vec.Vec_IntEntry(first, int(repr))
		if prev < 0 {
			vec.Vec_IntWriteEntry(first, int(repr), int32(i))
		} else if prev != int32(i) {
			vec.Vec_IntWriteEntry(result, i, prev)
		}
	}
	Gia_ManSetPhase(g)
	for i, prev := range vec.Vec_IntArray(result) {
		if prev < 0 {
			continue
		}
		a, b := vec.Vec_IntEntry(mapping, i), vec.Vec_IntEntry(mapping, int(prev))
		phase := int32(g.pObjs[a>>1].fPhase) ^ int32(g.pObjs[b>>1].fPhase) ^ (a & 1) ^ (b & 1)
		vec.Vec_IntWriteEntry(result, i, prev*2+phase)
	}
	return result
}

// Gia_MiniAigReduce substitutes each supplied earlier equivalence literal.
func Gia_MiniAigReduce(p *miniaig.Mini_Aig_t, equivs []int32) *Gia_Man_t {
	if p == nil || !miniaig.Mini_AigCheck(p) || len(equivs) != miniaig.Mini_AigNodeNum(p) {
		panic("gia: invalid MiniAIG equivalences")
	}
	n := len(equivs)
	g := Gia_ManStart(int32(n))
	g.pName = "MiniAig"
	copies := vec.Vec_IntAlloc(n)
	vec.Vec_IntPush(copies, 0)
	Gia_ManHashAlloc(g)
	for id := 1; id < n; id++ {
		var lit int32
		switch {
		case miniaig.Mini_AigNodeIsPi(p, id):
			lit = Gia_ManAppendCi(g)
		case miniaig.Mini_AigNodeIsPo(p, id):
			lit = Gia_ManAppendCo(g, Gia_ObjFromMiniFanin0Copy(g, copies, p, int32(id)))
		case miniaig.Mini_AigNodeIsAnd(p, id):
			lit = Gia_ManHashAnd(g, Gia_ObjFromMiniFanin0Copy(g, copies, p, int32(id)), Gia_ObjFromMiniFanin1Copy(g, copies, p, int32(id)))
		default:
			panic("gia: invalid MiniAIG node")
		}
		if e := equivs[id]; e >= 0 {
			if int(e>>1) >= id {
				panic("gia: forward equivalence")
			}
			lit = vec.Vec_IntEntry(copies, int(e>>1)) ^ (e & 1)
		}
		vec.Vec_IntPush(copies, lit)
	}
	Gia_ManHashStop(g)
	Gia_ManSetRegNum(g, int32(miniaig.Mini_AigRegNum(p)))
	result := Gia_ManSeqCleanup(g)
	Gia_ManStop(g)
	return result
}

// Gia_MiniAigMiter creates one failing output for each equivalence assertion.
func Gia_MiniAigMiter(p *miniaig.Mini_Aig_t, equivs []int32) *Gia_Man_t {
	if p == nil || !miniaig.Mini_AigCheck(p) || len(equivs) != miniaig.Mini_AigNodeNum(p) {
		panic("gia: invalid MiniAIG equivalences")
	}
	n := len(equivs)
	g := Gia_ManStart(int32(2 * n))
	g.pName = "MiniAig"
	copies := vec.Vec_IntAlloc(n)
	vec.Vec_IntPush(copies, 0)
	Gia_ManHashAlloc(g)
	for id := 1; id < n; id++ {
		var lit int32
		switch {
		case miniaig.Mini_AigNodeIsPi(p, id):
			lit = Gia_ManAppendCi(g)
		case miniaig.Mini_AigNodeIsPo(p, id):
			lit = -1
		case miniaig.Mini_AigNodeIsAnd(p, id):
			lit = Gia_ManHashAnd(g, Gia_ObjFromMiniFanin0Copy(g, copies, p, int32(id)), Gia_ObjFromMiniFanin1Copy(g, copies, p, int32(id)))
		default:
			panic("gia: invalid MiniAIG node")
		}
		vec.Vec_IntPush(copies, lit)
	}
	for id := 1; id < n; id++ {
		e := equivs[id]
		if e < 0 {
			continue
		}
		if int(e>>1) >= id || vec.Vec_IntEntry(copies, id) < 0 || vec.Vec_IntEntry(copies, int(e>>1)) < 0 {
			panic("gia: invalid miter equivalence")
		}
		a := vec.Vec_IntEntry(copies, id)
		b := vec.Vec_IntEntry(copies, int(e>>1)) ^ (e & 1)
		Gia_ManAppendCo(g, Gia_ManHashXor(g, a, b))
	}
	npos := miniaig.Mini_AigPoNum(p)
	regs := miniaig.Mini_AigRegNum(p)
	pos := 0
	for id := 1; id < n; id++ {
		if miniaig.Mini_AigNodeIsPo(p, id) {
			if pos >= npos-regs {
				Gia_ManAppendCo(g, Gia_ObjFromMiniFanin0Copy(g, copies, p, int32(id)))
			}
			pos++
		}
	}
	Gia_ManSetRegNum(g, int32(regs))
	Gia_ManHashStop(g)
	result := Gia_ManCleanup(g)
	Gia_ManStop(g)
	return result
}
