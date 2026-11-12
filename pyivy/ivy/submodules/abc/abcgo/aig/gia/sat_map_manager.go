package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
)

type Sbm_Man_t struct {
	pSat                                                       *bsat.Solver
	vCardVars                                                  *Vec_Int_t
	LogN, FirstVar, LitShift, nInputs                          int32
	vRoots                                                     *Vec_Int_t
	vCuts, vObjCuts                                            *Vec_Wec_t
	vSolCuts, vCutGates                                        *Vec_Int_t
	vCutAreas                                                  *Vec_Wrd_t
	vAssump, vPolar, vArrs, vReqs, vLit2Used, vDelays, vReason *Vec_Int_t
}

func Sbm_ManAlloc(logN int32) *Sbm_Man_t {
	p := &Sbm_Man_t{LogN: logN}
	p.pSat = Sbm_AddCardinSolver(logN, &p.vCardVars)
	p.FirstVar = p.pSat.NumVars()
	p.vRoots = vec.Vec_IntAlloc(100)
	p.vCuts = vec.Vec_WecAlloc(1000)
	p.vObjCuts = vec.Vec_WecAlloc(1000)
	p.vSolCuts = vec.Vec_IntAlloc(100)
	p.vCutGates = vec.Vec_IntAlloc(100)
	p.vCutAreas = vec.Vec_WrdAlloc(100)
	p.vAssump = vec.Vec_IntAlloc(100)
	p.vPolar = vec.Vec_IntAlloc(100)
	p.vArrs = vec.Vec_IntAlloc(100)
	p.vReqs = vec.Vec_IntAlloc(100)
	p.vLit2Used = vec.Vec_IntAlloc(100)
	p.vDelays = vec.Vec_IntAlloc(100)
	p.vReason = vec.Vec_IntAlloc(100)
	return p
}
func Sbm_ManStop(p *Sbm_Man_t) { p.pSat = nil }
func Sbm_ManCheckSol(p *Sbm_Man_t, _ *Vec_Int_t) int32 {
	vec.Vec_IntClear(p.vPolar)
	vec.Vec_IntFill(p.vLit2Used, vec.Vec_WecSize(p.vObjCuts)+int(p.nInputs), 0)
	for _, cut := range vec.Vec_IntArray(p.vSolCuts) {
		if cut < 0 {
			vec.Vec_IntWriteEntry(p.vLit2Used, int(-cut), 1)
			vec.Vec_IntPush(p.vPolar, -cut)
			continue
		}
		vec.Vec_IntPush(p.vPolar, p.FirstVar+cut)
		record := vec.Vec_WecEntry(p.vCuts, int(cut))
		lit := vec.Vec_IntEntry(record, 0) - p.LitShift
		if vec.Vec_IntEntry(p.vLit2Used, int(lit)) == 0 {
			vec.Vec_IntWriteEntry(p.vLit2Used, int(lit), 1)
			vec.Vec_IntPush(p.vPolar, lit)
		}
	}
	result := int32(1)
	for _, lit := range vec.Vec_IntArray(p.vRoots) {
		if vec.Vec_IntEntry(p.vLit2Used, int(lit)) == 0 {
			fmt.Printf("Output literal %d has no cut.\n", lit)
			result = 0
		}
	}
	for _, cut := range vec.Vec_IntArray(p.vSolCuts) {
		if cut < 0 {
			continue
		}
		data := vec.Vec_IntArray(vec.Vec_WecEntry(p.vCuts, int(cut)))
		for _, lit := range data[1:] {
			index := lit - p.LitShift
			if index < 0 {
				if lit&1 == 0 {
					panic("gia: expected complemented input")
				}
				index = int32(vec.Vec_WecSize(p.vObjCuts)) + lit/2 - 1
			}
			if vec.Vec_IntEntry(p.vLit2Used, int(index)) == 0 {
				fmt.Printf("Literal %d of cut %d is not mapped.\n", lit, cut)
				result = 0
			}
		}
		vec.Vec_IntPush(p.vPolar, p.FirstVar+cut)
	}
	return result
}
