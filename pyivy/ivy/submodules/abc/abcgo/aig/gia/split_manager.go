package gia

import "github.com/glycerine/abcgo/misc/vec"

type Spl_Man_t struct {
	pGia                                                      *Gia_Man_t
	iObj, Limit, fReverse                                     int32
	vMarksCIO, vMarksIn, vMarksNo, vMarksAnd                  *Vec_Bit_t
	vRoots, vNodes, vLeaves, vAnds, vFanouts, vCands, vInputs *Vec_Int_t
}

func Spl_ManAlloc(g *Gia_Man_t, limit, reverse int32) *Spl_Man_t {
	if Gia_ManHasMapping(g) == 0 || Gia_ManHasMapping2(g) != 0 {
		panic("gia: invalid mapping for window manager")
	}
	p := &Spl_Man_t{pGia: g, Limit: limit, fReverse: reverse}
	p.vMarksCIO = vec.Vec_BitStart(int(g.nObjs))
	p.vMarksIn = vec.Vec_BitStart(int(g.nObjs))
	p.vMarksNo = vec.Vec_BitStart(int(g.nObjs))
	p.vMarksAnd = vec.Vec_BitStart(int(g.nObjs))
	p.vRoots = vec.Vec_IntAlloc(100)
	p.vNodes = vec.Vec_IntAlloc(100)
	p.vLeaves = vec.Vec_IntAlloc(100)
	p.vAnds = vec.Vec_IntAlloc(100)
	p.vFanouts = vec.Vec_IntAlloc(100)
	p.vCands = vec.Vec_IntAlloc(100)
	p.vInputs = vec.Vec_IntAlloc(100)
	vec.Vec_BitWriteEntry(p.vMarksCIO, 0, 1)
	for i := int32(0); i < Gia_ManCiNum(g); i++ {
		vec.Vec_BitWriteEntry(p.vMarksCIO, int(Gia_ObjId(g, Gia_ManCi(g, i))), 1)
	}
	for i := int32(0); i < Gia_ManCoNum(g); i++ {
		vec.Vec_BitWriteEntry(p.vMarksCIO, int(Gia_ObjId(g, Gia_ManCo(g, i))), 1)
	}
	g.Refs = nil
	Gia_ManCreateRefs(g)
	Gia_ManSetLutRefs(g)
	g.vMapping2 = Spl_ManToWecMapping(g)
	g.VMapping = nil
	Gia_ManStaticFanoutStart(g)
	return p
}
func Spl_ManStop(p *Spl_Man_t) {
	Gia_ManStaticFanoutStop(p.pGia)
	if Gia_ManHasMapping(p.pGia) != 0 || Gia_ManHasMapping2(p.pGia) == 0 {
		panic("gia: invalid window manager mapping")
	}
	p.pGia.VMapping = Spl_ManFromWecMapping(p.pGia, p.pGia.vMapping2)
	p.pGia.vMapping2 = nil
}
