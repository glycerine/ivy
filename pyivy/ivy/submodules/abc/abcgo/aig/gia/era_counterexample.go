package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
	"github.com/glycerine/abcgo/sat/cnf"
)

// Initialize the source algorithm's single-frame CNF once. Register outputs
// are free CIs here; each trace query binds its current and next state by
// assumptions, retaining learned clauses for the following frame.
func Gia_ManAreDeriveCexSatStart(p *Gia_ManAre_t) {
	if p.pSat != nil {
		panic("gia: counterexample solver already active")
	}
	source := p.pAig
	for _, obj := range source.pObjs {
		if Gia_ObjIsXor(obj) != 0 || Gia_ObjIsMux(source, obj) != 0 {
			// The source export API reads raw AND fanins. Supply the
			// equivalent ordinary AIG before exporting real gates to CNF.
			saved := make([]uint32, len(source.pObjs))
			for i, obj := range source.pObjs {
				saved[i] = obj.Value
			}
			original := source
			source = Gia_ManDupNoMuxes(source, 1)
			for i, obj := range original.pObjs {
				obj.Value = saved[i]
			}
			defer Gia_ManStop(source)
			break
		}
	}
	graph := Gia_ManToAig(source, 0)
	defer aig.Aig_ManStop(graph)
	aig.Aig_ManSetRegNum(graph, 0)
	data := cnf.Cnf_Derive(graph, Gia_ManCoNum(p.pAig))
	if data == nil {
		panic("gia: could not derive counterexample CNF")
	}
	defer cnf.Cnf_DataFree(data)
	p.pSat = cnf.Cnf_DataWriteIntoSolver(data, 1, 0)
	if p.pSat == nil {
		panic("gia: inconsistent counterexample CNF")
	}
	p.vSatNumCis = cnf.Cnf_DataCollectCiSatNums(data, graph)
	p.vSatNumCos = cnf.Cnf_DataCollectCoSatNums(data, graph)
	p.vCofVars = vec.Vec_IntAlloc(100)
	p.vAssumps = vec.Vec_IntAlloc(100)
}
func Gia_ManAreDeriveCexSatStop(p *Gia_ManAre_t) {
	if p.pSat == nil || p.pTarget == nil {
		panic("gia: counterexample solver is not active")
	}
	p.pSat.Close()
	p.pSat = nil
	p.pTarget = nil
	p.vSatNumCis, p.vSatNumCos, p.vCofVars, p.vAssumps = nil, nil, nil, nil
}
func Gia_ManAreDeriveCexSat(p *Gia_ManAre_t, current, next *Gia_StaAre_t, output int32) {
	if p.pSat == nil {
		panic("gia: counterexample solver is not active")
	}
	g := p.pAig
	vec.Vec_IntClear(p.vAssumps)
	bind := func(state *Gia_StaAre_t, vars *Vec_Int_t, offset int32) {
		for i := int32(0); i < g.nRegs; i++ {
			variable := vec.Vec_IntEntry(vars, int(offset+i))
			if Gia_StaHasValue0(state, i) != 0 {
				vec.Vec_IntPush(p.vAssumps, bsat.Negative(variable))
			} else if Gia_StaHasValue1(state, i) != 0 {
				vec.Vec_IntPush(p.vAssumps, bsat.Positive(variable))
			}
		}
	}
	bind(current, p.vSatNumCis, Gia_ManPiNum(g))
	if next != nil {
		bind(next, p.vSatNumCos, Gia_ManPoNum(g))
	}
	if output >= 0 {
		if output >= Gia_ManPoNum(g) {
			panic("gia: invalid failing output")
		}
		vec.Vec_IntPush(p.vAssumps, bsat.Positive(vec.Vec_IntEntry(p.vSatNumCos, int(output))))
	}
	status := p.pSat.SolveWithResourceLimits(vec.Vec_IntArray(p.vAssumps), 1000000, 0, 0, 0)
	if status != bsat.Satisfiable {
		panic(fmt.Sprintf("gia: could not reconstruct reachability transition (status %d)", status))
	}
	value := func(index int32) bool {
		v, ok := p.pSat.Value(vec.Vec_IntEntry(p.vSatNumCis, int(index)))
		if !ok {
			panic("gia: missing counterexample SAT model")
		}
		return v
	}
	vec.Vec_IntClear(p.vCofVars)
	for i := int32(0); i < Gia_ManPiNum(g); i++ {
		if value(i) {
			vec.Vec_IntPush(p.vCofVars, i)
		}
	}
	for i := int32(0); i < g.nRegs; i++ {
		if value(Gia_ManPiNum(g) + i) {
			Gia_StaSetValue1(current, i)
		} else {
			Gia_StaSetValue0(current, i)
		}
	}
}
func Gia_ManAreDeriveCex(p *Gia_ManAre_t, last *Gia_StaAre_t) *Abc_Cex_t {
	if p.iOutFail < 0 || last == nil {
		panic("gia: missing failing reachability target")
	}
	if p.pSat == nil {
		Gia_ManAreDeriveCexSatStart(p)
	}
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
