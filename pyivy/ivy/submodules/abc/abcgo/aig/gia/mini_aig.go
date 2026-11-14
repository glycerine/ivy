package gia

import (
	"github.com/glycerine/abcgo/aig/miniaig"
	"github.com/glycerine/abcgo/misc/vec"
)

func Gia_ObjFromMiniFanin0Copy(_ *Gia_Man_t, copies *Vec_Int_t, p *miniaig.Mini_Aig_t, id int32) int32 {
	lit := miniaig.Mini_AigNodeFanin0(p, int(id))
	return vec.Vec_IntEntry(copies, lit>>1) ^ int32(lit&1)
}
func Gia_ObjFromMiniFanin1Copy(_ *Gia_Man_t, copies *Vec_Int_t, p *miniaig.Mini_Aig_t, id int32) int32 {
	lit := miniaig.Mini_AigNodeFanin1(p, int(id))
	return vec.Vec_IntEntry(copies, lit>>1) ^ int32(lit&1)
}

// Gia_ManFromMiniAig imports a MiniAIG and optionally returns its object map.
func Gia_ManFromMiniAig(p *miniaig.Mini_Aig_t, out **Vec_Int_t) *Gia_Man_t {
	if p == nil || !miniaig.Mini_AigCheck(p) {
		panic("gia: invalid MiniAIG")
	}
	n := miniaig.Mini_AigNodeNum(p)
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
		vec.Vec_IntPush(copies, lit)
	}
	Gia_ManHashStop(g)
	Gia_ManSetRegNum(g, int32(miniaig.Mini_AigRegNum(p)))
	result := Gia_ManCleanup(g)
	if out != nil {
		Gia_ManDupRemapLiterals(copies, g)
		*out = copies
	}
	Gia_ManStop(g)
	return result
}

// Gia_ManToMiniAig exports GIA objects in input, node and output order.
func Gia_ManToMiniAig(g *Gia_Man_t) *miniaig.Mini_Aig_t {
	p := miniaig.Mini_AigStart()
	g.pObjs[0].Value = 0
	for _, id := range vec.Vec_IntArray(g.vCis) {
		g.pObjs[id].Value = uint32(miniaig.Mini_AigCreatePi(p))
	}
	for _, obj := range g.pObjs[1:] {
		if Gia_ObjIsAnd(obj) == 0 {
			continue
		}
		// MiniAIG stores ordinary ANDs. Expand real gates using its existing
		// constructors rather than treating their stored data fanins as ANDs.
		a, b, c, kind := g.Gate(obj.id)
		copyLit := func(lit int32) int { return int(Gia_ObjLitCopy(g, lit)) }
		var literal int
		switch kind {
		case 2:
			literal = miniaig.Mini_AigAnd(p, copyLit(a), copyLit(b))
		case 3:
			literal = miniaig.Mini_AigXor(p, copyLit(a), copyLit(b))
		case 4:
			literal = miniaig.Mini_AigMux(p, copyLit(a), copyLit(b), copyLit(c))
		case 5:
			literal = copyLit(a)
		default:
			panic("gia: invalid MiniAIG export gate")
		}
		obj.Value = uint32(literal)
	}
	for _, id := range vec.Vec_IntArray(g.vCos) {
		obj := g.pObjs[id]
		obj.Value = uint32(miniaig.Mini_AigCreatePo(p, int(Gia_ObjFanin0Copy(obj))))
	}
	miniaig.Mini_AigSetRegNum(p, int(Gia_ManRegNum(g)))
	return p
}
