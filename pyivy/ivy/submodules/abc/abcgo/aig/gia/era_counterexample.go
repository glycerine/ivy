package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/proof/cec/core"
)

func Gia_ManAreDeriveCexSatStart(p *Gia_ManAre_t) {
	if p.pSat != nil {
		panic("gia: counterexample solver already active")
	}
	p.pSat = struct{}{}
	p.vCofVars = vec.Vec_IntAlloc(100)
	p.vAssumps = vec.Vec_IntAlloc(100)
}
func Gia_ManAreDeriveCexSatStop(p *Gia_ManAre_t) {
	if p.pSat == nil || p.pTarget == nil {
		panic("gia: counterexample solver is not active")
	}
	p.pSat = nil
	p.pTarget = nil
	p.vCofVars = nil
	p.vAssumps = nil
}
func Gia_ManAreDeriveCexSat(p *Gia_ManAre_t, current, next *Gia_StaAre_t, output int32) {
	if p.pSat == nil {
		panic("gia: counterexample solver is not active")
	}
	g := p.pAig
	q := Gia_ManStart(g.nObjs + 2*g.nRegs + 4)
	Gia_ManHashAlloc(q)
	copies := make([]int32, g.nObjs)
	for i := int32(0); i < Gia_ManCiNum(g); i++ {
		obj := Gia_ManCi(g, i)
		copies[obj.id] = Gia_ManAppendCi(q)
	}
	edge := func(obj *Gia_Obj_t) int32 { return copies[Gia_ObjFaninId0(obj, obj.id)] ^ Gia_ObjFaninC0(obj) }
	for _, obj := range g.pObjs {
		if Gia_ObjIsAnd(obj) != 0 {
			copies[obj.id] = Gia_ManHashAnd(q, edge(obj), copies[Gia_ObjFaninId1(obj, obj.id)]^Gia_ObjFaninC1(obj))
		}
	}
	condition := int32(1)
	for i := int32(0); i < g.nRegs; i++ {
		lit := copies[Gia_ManRo(g, i).id]
		if Gia_StaHasValue0(current, i) != 0 {
			condition = Gia_ManHashAnd(q, condition, lit^1)
		} else if Gia_StaHasValue1(current, i) != 0 {
			condition = Gia_ManHashAnd(q, condition, lit)
		}
	}
	if next != nil {
		for i := int32(0); i < g.nRegs; i++ {
			lit := edge(Gia_ManRi(g, i))
			if Gia_StaHasValue0(next, i) != 0 {
				condition = Gia_ManHashAnd(q, condition, lit^1)
			} else if Gia_StaHasValue1(next, i) != 0 {
				condition = Gia_ManHashAnd(q, condition, lit)
			}
		}
	}
	if output >= 0 {
		if output >= Gia_ManPoNum(g) {
			panic("gia: invalid failing output")
		}
		condition = Gia_ManHashAnd(q, condition, edge(Gia_ManPo(g, output)))
	}
	Gia_ManHashStop(q)
	options := core.DefaultSATOptions()
	options.ConflictLimit = 1000000
	solver, err := core.NewManager(q, options)
	if err != nil {
		panic(err)
	}
	result, err := solver.CheckLiteral(condition)
	if err != nil {
		panic(err)
	}
	if result.Status != core.Different || result.Counterexample == nil {
		panic(fmt.Sprintf("gia: could not reconstruct reachability transition (status %d)", result.Status))
	}
	values := result.Counterexample.Inputs
	vec.Vec_IntClear(p.vCofVars)
	for i := int32(0); i < Gia_ManPiNum(g); i++ {
		if values[i] {
			vec.Vec_IntPush(p.vCofVars, i)
		}
	}
	for i := int32(0); i < g.nRegs; i++ {
		if values[Gia_ManPiNum(g)+i] {
			Gia_StaSetValue1(current, i)
		} else {
			Gia_StaSetValue0(current, i)
		}
	}
	Gia_ManStop(q)
}
func Gia_ManAreDeriveCex(p *Gia_ManAre_t, last *Gia_StaAre_t) *Abc_Cex_t {
	if p.iOutFail < 0 || last == nil {
		panic("gia: missing failing reachability target")
	}
	Gia_ManAreDeriveCexSatStart(p)
	var states []*Gia_StaAre_t
	for s := last; Gia_StaIsGood(p, s) != 0; s = Gia_StaPrev(p, s) {
		if s != last {
			states = append(states, s)
		}
	}
	if len(states) == 0 {
		panic("gia: missing initial state")
	}
	cex := util.Abc_CexAlloc(p.pAig.nRegs, Gia_ManPiNum(p.pAig), int32(len(states)))
	cex.IFrame = int32(len(states) - 1)
	cex.IPo = p.iOutFail
	var previous *Gia_StaAre_t
	for i, s := range states {
		output := int32(-1)
		if i == 0 {
			output = p.iOutFail
		}
		Gia_ManAreDeriveCexSat(p, s, previous, output)
		previous = s
		for _, v := range vec.Vec_IntArray(p.vCofVars) {
			util.Abc_InfoSetBit(cex.PData, p.pAig.nRegs+int32(len(states)-1-i)*Gia_ManPiNum(p.pAig)+v)
		}
	}
	Gia_ManAreDeriveCexSatStop(p)
	return cex
}
