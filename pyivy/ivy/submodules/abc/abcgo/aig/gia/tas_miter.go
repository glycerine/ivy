package gia

import (
	"fmt"
	"time"

	"github.com/glycerine/abcgo/misc/vec"
)

func Tas_ManSatPrintStats(p *Tas_Man_t) {
	fmt.Printf("CO = %8d  AND = %8d  Conf = %6d  JustMax = %5d\n", Gia_ManCoNum(p.pAig), Gia_ManAndNum(p.pAig), p.Pars.nBTLimit, p.Pars.nJustLimit)
	for _, row := range []struct {
		name             string
		calls, conflicts int32
		elapsed          time.Duration
	}{
		{"Unsat", p.nSatUnsat, p.nConfUnsat, p.timeSatUnsat}, {"Sat", p.nSatSat, p.nConfSat, p.timeSatSat}, {"Undef", p.nSatUndec, p.nConfUndec, p.timeSatUndec},
	} {
		var percent, average float64
		if p.nSatTotal != 0 {
			percent = 100 * float64(row.calls) / float64(p.nSatTotal)
		}
		if row.calls != 0 {
			average = float64(row.conflicts) / float64(row.calls)
		}
		fmt.Printf("%-5s calls %6d  (%6.2f %%)   Ave conf = %8.1f   Time = %s\n", row.name, row.calls, percent, average, row.elapsed)
	}
	fmt.Printf("Total time = %s\n", p.timeTotal)
}

func tasAppendCounterexample(store *vec.Vec_Int_t, model *vec.Vec_Int_t, output int32) {
	vec.Vec_IntPush(store, output)
	if model == nil {
		vec.Vec_IntPush(store, -1)
		return
	}
	vec.Vec_IntPush(store, int32(vec.Vec_IntSize(model)))
	for _, lit := range vec.Vec_IntArray(model) {
		vec.Vec_IntPush(store, lit)
	}
}

// Tas_ManSolveMiterNc solves each output and records status and SAT cubes.
func Tas_ManSolveMiterNc(aig *Gia_Man_t, conflicts int32, status **vec.Vec_Str_t, verbose int32) *vec.Vec_Int_t {
	if aig == nil || status == nil || Gia_ManRegNum(aig) != 0 {
		panic("gia: miter solver requires a combinational AIG and status destination")
	}
	started := time.Now()
	Gia_ManCreateRefs(aig)
	Gia_ManCleanMark0(aig)
	Gia_ManCleanMark1(aig)
	Gia_ManFillValue(aig)
	Gia_ManCleanPhase(aig)
	p := Tas_ManAlloc(aig, conflicts)
	defer Tas_ManStop(p)
	result := vec.Vec_IntAlloc(10000)
	states := vec.Vec_StrAlloc(int(Gia_ManPoNum(aig)))
	for i := int32(0); i < Gia_ManCoNum(aig); i++ {
		output := Gia_ManCo(aig, i)
		root := Gia_ObjChild0(output)
		vec.Vec_IntClear(p.vModel)
		if Gia_Regular(root) == Gia_ManConst0(aig) {
			if Gia_IsComplement(root) != 0 {
				tasAppendCounterexample(result, p.vModel, i)
				vec.Vec_StrPush(states, 0)
			} else {
				vec.Vec_StrPush(states, 1)
			}
			continue
		}
		start := time.Now()
		code := Tas_ManSolve(p, root, nil)
		elapsed := time.Since(start)
		vec.Vec_StrPush(states, byte(int8(code)))
		switch code {
		case -1:
			p.nSatUndec++
			p.nConfUndec += p.Pars.nBTThis
			p.timeSatUndec += elapsed
			tasAppendCounterexample(result, nil, i)
		case 1:
			p.nSatUnsat++
			p.nConfUnsat += p.Pars.nBTThis
			p.timeSatUnsat += elapsed
		case 0:
			p.nSatSat++
			p.nConfSat += p.Pars.nBTThis
			p.timeSatSat += elapsed
			tasAppendCounterexample(result, p.vModel, i)
		default:
			panic("gia: invalid SAT status")
		}
	}
	p.nSatTotal = Gia_ManPoNum(aig)
	p.timeTotal = time.Since(started)
	if verbose != 0 {
		Tas_ManSatPrintStats(p)
	}
	*status = states
	return result
}

func Tas_ManPrintClause(p *Tas_Man_t, level, handle int32) {
	if Tas_QueIsEmpty(&p.pClauses) == 0 {
		panic("gia: clause queue is not empty")
	}
	fmt.Printf("Level %2d : ", level)
	for i := handle; p.pClauses.pData[i] != nil; i++ {
		node := p.pClauses.pData[i]
		fmt.Printf("%d=%d(%d) ", Gia_ObjId(p.pAig, node), Tas_VarValue(node), Tas_VarDecLevel(p, node))
	}
	fmt.Println()
}

func Tas_ManPrintClauseNew(p *Tas_Man_t, level, handle int32) {
	if Tas_QueIsEmpty(&p.pClauses) == 0 {
		panic("gia: clause queue is not empty")
	}
	fmt.Printf("Level %2d : ", level)
	for i := handle; p.pClauses.pData[i] != nil; i++ {
		node := p.pClauses.pData[i]
		sign := '-'
		if Tas_VarValue(node) != 0 {
			sign = '+'
		}
		fmt.Printf("%c%d ", sign, Gia_ObjId(p.pAig, node))
	}
	fmt.Println()
}

// Tas_ManSolveMiterNc2 packs SAT patterns and marks proved old roots.
func Tas_ManSolveMiterNc2(aig *Gia_Man_t, conflicts int32, old *Gia_Man_t, roots []*Gia_Obj_t, simInfo [][]uint32) {
	if aig == nil || old == nil || Gia_ManRegNum(aig) != 0 || len(roots) < int(Gia_ManCoNum(aig)) || len(simInfo) < int(Gia_ManCiNum(aig)) {
		panic("gia: invalid pattern miter inputs")
	}
	for _, words := range simInfo[:Gia_ManCiNum(aig)] {
		if len(words) < 1 {
			panic("gia: missing simulation word")
		}
	}
	started := time.Now()
	Gia_ManCreateRefs(aig)
	Gia_ManCleanMark0(aig)
	Gia_ManCleanMark1(aig)
	Gia_ManFillValue(aig)
	Gia_ManCleanPhase(aig)
	p := Tas_ManAlloc(aig, conflicts)
	defer Tas_ManStop(p)
	presence := make([][]uint32, Gia_ManCiNum(aig))
	for i := range presence {
		presence[i] = make([]uint32, 1)
	}
	var tried, stored int
	step := int(Gia_ManCoNum(aig))/1000 + 1
	for i := 0; i < int(Gia_ManCoNum(aig)); i += step {
		root := Gia_ObjChild0(Gia_ManCo(aig, int32(i)))
		if Gia_Regular(root) == Gia_ManConst0(aig) {
			panic("gia: pattern miter output must be nonconstant")
		}
		start := time.Now()
		status := Tas_ManSolve(p, root, nil)
		elapsed := time.Since(start)
		switch status {
		case -1:
			p.nSatUndec++
			p.nConfUndec += p.Pars.nBTThis
			p.timeSatUndec += elapsed
		case 1:
			p.nSatUnsat++
			p.nConfUnsat += p.Pars.nBTThis
			p.timeSatUnsat += elapsed
			id := Gia_ObjId(old, roots[i])
			if Gia_ObjProved(old, id) != 0 {
				panic("gia: root already proved")
			}
			Gia_ObjSetProved(old, id)
		case 0:
			p.nSatSat++
			p.nConfSat += p.Pars.nBTThis
			p.timeSatSat += elapsed
			tried++
			stored += int(Tas_StorePattern(simInfo, presence, vec.Vec_IntArray(p.vModel)))
		default:
			panic("gia: invalid SAT status")
		}
	}
	fmt.Printf("Tried = %d  Stored = %d\n", tried, stored)
	p.nSatTotal = Gia_ManPoNum(aig)
	p.timeTotal = time.Since(started)
	Tas_ManSatPrintStats(p)
}
