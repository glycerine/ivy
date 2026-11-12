package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/vec"
	"math/bits"
)

const (
	MAX_CALL_NUM = 1000000
	MAX_ITEM_NUM = 1 << 20
	MAX_PAGE_NUM = 1 << 11
	MAX_VARS_NUM = 1 << 14
	MAX_CUBE_NUM = 63
)

type Gia_ObjAre_t struct {
	iVar, nStas0, nStas1, nStas2 uint32
	F                            [3]Gia_PtrAre_t
}
type Gia_ManAre_t struct {
	pAig, pNew                                                       *Gia_Man_t
	objects                                                          []*Gia_ObjAre_t
	states                                                           []*Gia_StaAre_t
	fMiter, fStopped, fTree                                          int32
	nWords, nSize, nObjPages, nStaPages, nObjs, nStas, iStaCur       int32
	Root                                                             Gia_PtrAre_t
	vCiTfos, vCiLits                                                 *Vec_Vec_t
	vCubesA, vCubesB                                                 *Vec_Int_t
	pSat                                                             any
	vSatNumCis, vSatNumCos, vCofVars, vAssumps                       *Vec_Int_t
	pTarget                                                          *Gia_StaAre_t
	iOutFail                                                         int32
	nChecks, nEquals, nCompares, nRecCalls, nDisjs, nDisjs2, nDisjs3 int32
	timeAig, timeCube                                                int32
}

func Gia_ObjHasBranch0(q *Gia_ObjAre_t) int32 {
	return giaBool(q.nStas0 == 0 && Gia_Ptr2Int(q.F[0]) != 0)
}
func Gia_ObjHasBranch1(q *Gia_ObjAre_t) int32 {
	return giaBool(q.nStas1 == 0 && Gia_Ptr2Int(q.F[1]) != 0)
}
func Gia_ObjHasBranch2(q *Gia_ObjAre_t) int32 {
	return giaBool(q.nStas2 == 0 && Gia_Ptr2Int(q.F[2]) != 0)
}
func Gia_ManAreObj(p *Gia_ManAre_t, n Gia_PtrAre_t) *Gia_ObjAre_t    { return p.objects[Gia_Ptr2Int(n)] }
func Gia_ManAreSta(p *Gia_ManAre_t, n Gia_PtrAre_t) *Gia_StaAre_t    { return p.states[Gia_Ptr2Int(n)] }
func Gia_ManAreObjInt(p *Gia_ManAre_t, n int32) *Gia_ObjAre_t        { return p.objects[n] }
func Gia_ManAreStaInt(p *Gia_ManAre_t, n int32) *Gia_StaAre_t        { return p.states[n] }
func Gia_ManAreObjLast(p *Gia_ManAre_t) *Gia_ObjAre_t                { return p.objects[p.nObjs-1] }
func Gia_ManAreStaLast(p *Gia_ManAre_t) *Gia_StaAre_t                { return p.states[p.nStas-1] }
func Gia_ObjNextObj0(p *Gia_ManAre_t, q *Gia_ObjAre_t) *Gia_ObjAre_t { return Gia_ManAreObj(p, q.F[0]) }
func Gia_ObjNextObj1(p *Gia_ManAre_t, q *Gia_ObjAre_t) *Gia_ObjAre_t { return Gia_ManAreObj(p, q.F[1]) }
func Gia_ObjNextObj2(p *Gia_ManAre_t, q *Gia_ObjAre_t) *Gia_ObjAre_t { return Gia_ManAreObj(p, q.F[2]) }
func Gia_StaPrev(p *Gia_ManAre_t, s *Gia_StaAre_t) *Gia_StaAre_t     { return Gia_ManAreSta(p, s.iPrev) }
func Gia_StaNext(p *Gia_ManAre_t, s *Gia_StaAre_t) *Gia_StaAre_t     { return Gia_ManAreSta(p, s.iNext) }
func Gia_StaIsGood(p *Gia_ManAre_t, s *Gia_StaAre_t) int32           { return giaBool(s != p.states[0]) }
func Gia_ManAreCreate(g *Gia_Man_t) *Gia_ManAre_t {
	words := (2*g.nRegs + 31) / 32
	tfos := Gia_ManDeriveCiTfo(g)
	return &Gia_ManAre_t{pAig: g, nWords: words, nSize: 2 + words, vCiTfos: tfos, vCiLits: vec.Vec_VecDupInt(tfos), vCubesA: vec.Vec_IntAlloc(100), vCubesB: vec.Vec_IntAlloc(100), iOutFail: -1}
}
func Gia_ManAreFree(p *Gia_ManAre_t) {
	if p != nil {
		Gia_ManStop(p.pAig)
		if p.pNew != nil {
			Gia_ManStop(p.pNew)
		}
		*p = Gia_ManAre_t{}
	}
}
func Gia_ManAreCreateObj(p *Gia_ManAre_t) *Gia_ObjAre_t {
	if p.nObjs == 0 {
		p.objects = append(p.objects, &Gia_ObjAre_t{})
		p.nObjs = 1
	}
	if int64(p.nObjs) >= int64(MAX_ITEM_NUM)*MAX_PAGE_NUM-1 {
		fmt.Println("ERA manager has run out of internal nodes.")
		return nil
	}
	obj := &Gia_ObjAre_t{}
	p.objects = append(p.objects, obj)
	p.nObjs++
	p.nObjPages = (p.nObjs + MAX_ITEM_NUM - 1) / MAX_ITEM_NUM
	return obj
}
func Gia_ManAreCreateSta(p *Gia_ManAre_t) *Gia_StaAre_t {
	if p.nStas == 0 {
		p.states = append(p.states, &Gia_StaAre_t{pData: make([]uint32, p.nWords)})
		p.nStas = 1
	}
	if int64(p.nStas) >= int64(MAX_ITEM_NUM)*MAX_PAGE_NUM-1 {
		fmt.Println("ERA manager has run out of state cubes.")
		return nil
	}
	state := &Gia_StaAre_t{pData: make([]uint32, p.nWords)}
	p.states = append(p.states, state)
	p.nStas++
	p.nStaPages = (p.nStas + MAX_ITEM_NUM - 1) / MAX_ITEM_NUM
	return state
}
func Gia_ManAreRycycleSta(p *Gia_ManAre_t, s *Gia_StaAre_t) {
	clear(s.pData)
	s.iPrev = Gia_PtrAre_t{}
	s.iNext = Gia_PtrAre_t{}
	if p.nStas > 1 && s == Gia_ManAreStaLast(p) {
		p.states = p.states[:p.nStas-1]
		p.nStas--
		p.nStaPages = (p.nStas + MAX_ITEM_NUM - 1) / MAX_ITEM_NUM
	}
}
func Gia_ManAreCreateStaNew(p *Gia_ManAre_t) *Gia_StaAre_t {
	s := Gia_ManAreCreateSta(p)
	if s == nil {
		return nil
	}
	for i := int32(0); i < p.pAig.nRegs; i++ {
		switch Gia_ObjValue(Gia_ManRi(p.pAig, i)) {
		case 0:
			Gia_StaSetValue0(s, i)
		case 1:
			Gia_StaSetValue1(s, i)
		}
	}
	return s
}
func Gia_ManAreCreateStaInit(p *Gia_ManAre_t) *Gia_StaAre_t {
	for i := int32(0); i < p.pAig.nRegs; i++ {
		Gia_ObjSetValue(Gia_ManRi(p.pAig, i), 0)
	}
	return Gia_ManAreCreateStaNew(p)
}
func Gia_ManAreDepth(p *Gia_ManAre_t, id int32) int32 {
	var n int32
	for s := Gia_ManAreStaInt(p, id); Gia_StaIsGood(p, s) != 0; s = Gia_StaPrev(p, s) {
		n++
	}
	return n
}
func Gia_ManCountMinterms(p *Gia_ManAre_t) int32 {
	regs := Gia_ManRegNum(p.pAig)
	if regs > 30 {
		return -1
	}
	memory := make([]uint32, ((1<<uint(regs))+31)/32)
	for _, cube := range p.states[1:] {
		if Gia_StaIsUsed(cube) != 0 {
			Gia_ManCountMintermsInCube(cube, regs, memory)
		}
	}
	var n int32
	for _, word := range memory {
		n += int32(bits.OnesCount32(word))
	}
	return n
}
