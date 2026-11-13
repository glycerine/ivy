package bmc

import (
	"fmt"
	"time"

	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
)

// Gia_ManMaxiUnfold unfolds the transition relation and returns the final
// register inputs. Controls and data variables precede the timeframe inputs.
func Gia_ManMaxiUnfold(p *gia.Gia_Man_t, nFrames, useVars int32, initial *vec.Vec_Int_t) *gia.Gia_Man_t {
	if p == nil || nFrames <= 0 {
		panic("bmc: invalid maximal initialization unfolding")
	}
	nr, np := int(gia.Gia_ManRegNum(p)), int(gia.Gia_ManPiNum(p))
	if initial == nil || vec.Vec_IntSize(initial) != nr {
		panic("bmc: invalid initial state")
	}
	q := gia.Gia_ManStart(2*int32(nr) + nFrames*gia.Gia_ManObjNum(p))
	gia.Gia_ManHashStart(q)
	inputs := p.InputLiterals()
	outputs := p.OutputLiterals()
	npo := int(gia.Gia_ManPoNum(p))
	controls := make([]int32, nr)
	data := make([]int32, nr)
	for i := range controls {
		controls[i] = gia.Gia_ManAppendCi(q)
	}
	for i := range data {
		data[i] = gia.Gia_ManAppendCi(q)
	}
	values := make([]int32, gia.Gia_ManObjNum(p))
	next := make([]int32, nr)
	for i, lit := range inputs[np:] {
		switch vec.Vec_IntEntry(initial, i) {
		case 0:
			if useVars != 0 {
				values[lit>>1] = gia.Gia_ManHashAnd(q, controls[i], data[i])
			}
		case 1:
			if useVars != 0 {
				values[lit>>1] = gia.Gia_ManHashOr(q, controls[i]^1, data[i])
			} else {
				values[lit>>1] = 1
			}
		case 2:
			values[lit>>1] = gia.Gia_ManHashAnd(q, controls[i], data[i])
		case 3:
			values[lit>>1] = gia.Gia_ManHashOr(q, controls[i]^1, data[i])
		case 4:
			values[lit>>1] = 0
		case 5:
			values[lit>>1] = 1
		default:
			panic("bmc: invalid initial-state value")
		}
	}
	read := func(lit int32) int32 { return values[lit>>1] ^ (lit & 1) }
	for frame := int32(0); frame < nFrames; frame++ {
		for _, lit := range inputs[:np] {
			values[lit>>1] = gia.Gia_ManAppendCi(q)
		}
		for id := int32(1); id < gia.Gia_ManObjNum(p); id++ {
			a, b, c, k := p.Gate(id)
			switch k {
			case 2:
				values[id] = gia.Gia_ManHashAnd(q, read(a), read(b))
			case 3:
				values[id] = gia.Gia_ManHashXor(q, read(a), read(b))
			case 4:
				values[id] = gia.Gia_ManHashMux(q, read(a), read(b), read(c))
			case 5:
				values[id] = read(a)
			}
		}
		for i, lit := range outputs[npo:] {
			next[i] = read(lit)
		}
		for i, lit := range inputs[np:] {
			values[lit>>1] = next[i]
		}
	}
	for _, lit := range next {
		gia.Gia_ManAppendCo(q, lit)
	}
	gia.Gia_ManHashStop(q)
	return gia.Gia_ManCleanup(q)
}

// Gia_ManMaxiPerform weakens initial register constraints using SAT cores.
func Gia_ManMaxiPerform(p *gia.Gia_Man_t, initial *vec.Vec_Int_t, frames, timeout, verbose int32) *vec.Vec_Int_t {
	if p == nil || gia.Gia_ManRegNum(p) == 0 || initial == nil || vec.Vec_IntSize(initial) != int(gia.Gia_ManRegNum(p)) {
		panic("bmc: invalid maximal initialization input")
	}
	start := time.Now()
	a := Gia_ManMaxiUnfold(p, frames, 0, initial)
	b := Gia_ManMaxiUnfold(p, frames, 1, initial)
	m := gia.Gia_ManMiter(a, b, 0, 0, 0, 0, 0)
	if m == nil {
		return nil
	}
	enc, err := gia.Mf_ManGenerateCnfResult(m, 6, true, false, false)
	if err != nil {
		panic(err)
	}
	solver := bsat.NewSolver()
	for solver.NumVars() < int32(enc.Variables) {
		solver.NewVariable()
	}
	if timeout > 0 {
		solver.SetDeadline(time.Now().Add(time.Duration(timeout) * time.Second))
	}
	for _, clause := range enc.Clauses {
		if !solver.AddClause(clause...) {
			panic("bmc: inconsistent maximal initialization encoding")
		}
	}
	var outs []bsat.Literal
	for i := int32(0); i < gia.Gia_ManCoNum(m); i++ {
		id := gia.Gia_ObjId(m, gia.Gia_ManCo(m, i))
		outs = append(outs, bsat.Positive(enc.VarNums[id]))
	}
	solver.AddClause(outs...)
	var assumptions []bsat.Literal
	vars := make([]int32, gia.Gia_ManRegNum(p))
	for i := range vars {
		id := gia.Gia_ObjId(m, gia.Gia_ManCi(m, int32(i)))
		vars[i] = enc.VarNums[id]
		v := vec.Vec_IntEntry(initial, i)
		if v == 0 || v == 1 {
			assumptions = append(assumptions, bsat.Negative(vars[i]))
		}
	}
	for iteration := 0; iteration < 1000000; iteration++ {
		status := solver.Solve(assumptions...)
		if status == bsat.Unknown {
			fmt.Printf("Timeout reached after %d seconds and %d iterations.\n", timeout, iteration)
			break
		}
		if status == bsat.Satisfiable {
			if verbose != 0 {
				fmt.Printf("The problem is SAT after %d iterations.\n", iteration)
			}
			break
		}
		core := solver.UnsatCore()
		if verbose != 0 {
			fmt.Printf("Iter %d: subset = %d\n", iteration+1, len(core))
		}
		if len(core) == len(assumptions) {
			break
		}
		assumptions = assumptions[:0]
		for _, lit := range core {
			assumptions = append(assumptions, lit^1)
		}
	}
	used := make(map[int32]bool, len(assumptions))
	for _, lit := range assumptions {
		used[lit>>1] = true
	}
	result := vec.Vec_IntAlloc(vec.Vec_IntSize(initial))
	for i, v := range vec.Vec_IntArray(initial) {
		if (v == 0 || v == 1) && !used[vars[i]] {
			v |= 2
		}
		vec.Vec_IntPush(result, v)
	}
	if verbose != 0 {
		fmt.Printf("Total runtime: %s\n", time.Since(start))
	}
	return result
}

// Gia_ManMaxiTest supplies an all-zero initial state when none is specified.
func Gia_ManMaxiTest(p *gia.Gia_Man_t, initial *vec.Vec_Int_t, frames, words, timeout, sim, verbose int32) *vec.Vec_Int_t {
	_, _ = words, sim
	if initial == nil {
		initial = vec.Vec_IntStart(int(gia.Gia_ManRegNum(p)))
	}
	return Gia_ManMaxiPerform(p, initial, frames, timeout, verbose)
}
