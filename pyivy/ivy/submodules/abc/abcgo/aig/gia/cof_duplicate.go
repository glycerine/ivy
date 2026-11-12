package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/vec"
	"sort"
)

// Gia_ManDupCofInt rebuilds both cofactors and reconnects them with a mux.
func Gia_ManDupCofInt(p *Gia_Man_t, pivot int32) *Gia_Man_t {
	if pivot <= 0 || pivot >= p.nObjs || Gia_ObjIsCand(p.pObjs[pivot]) == 0 {
		return nil
	}
	out := Gia_ManStart(p.nObjs)
	out.CopyIdentityFrom(p)
	Gia_ManHashAlloc(out)
	inputs := make([]int32, p.nObjs)
	for i := int32(0); i < Gia_ManCiNum(p); i++ {
		obj := Gia_ManCi(p, i)
		inputs[obj.id] = Gia_ManAppendCi(out)
	}
	var selector int32 = -1
	cofactors := [2][]int32{make([]int32, p.nObjs), make([]int32, p.nObjs)}
	for phase := 0; phase < 2; phase++ {
		values := cofactors[phase]
		for i := int32(0); i < Gia_ManCiNum(p); i++ {
			obj := Gia_ManCi(p, i)
			values[obj.id] = inputs[obj.id]
			if obj.id == pivot {
				if phase == 0 {
					selector = values[obj.id]
				}
				values[obj.id] = int32(phase)
			}
		}
		lit := func(edge int32) int32 { return values[edge>>1] ^ (edge & 1) }
		for id := int32(1); id < p.nObjs; id++ {
			a, b, c, kind := p.Gate(id)
			switch kind {
			case 1:
				continue
			case 2:
				values[id] = Gia_ManHashAnd(out, lit(a), lit(b))
			case 3:
				values[id] = Gia_ManHashXor(out, lit(a), lit(b))
			case 4:
				values[id] = Gia_ManHashMux(out, lit(a), lit(b), lit(c))
			case 5, 6:
				values[id] = lit(a)
			default:
				panic("gia: invalid cofactor gate")
			}
			if id == pivot {
				if phase == 0 {
					selector = values[id]
				}
				values[id] = int32(phase)
			}
		}
	}
	if selector < 1 {
		Gia_ManStop(out)
		return nil
	}
	for id := int32(0); id < p.nObjs; id++ {
		p.pObjs[id].Value = uint32(cofactors[1][id])
	}
	for i := int32(0); i < Gia_ManCoNum(p); i++ {
		obj := Gia_ManCo(p, i)
		no, yes := cofactors[0][obj.id], cofactors[1][obj.id]
		lit := yes
		if yes != no {
			lit = Gia_ManHashMux(out, selector, yes, no)
		}
		obj.Value = uint32(Gia_ManAppendCo(out, lit))
	}
	Gia_ManHashStop(out)
	Gia_ManSetRegNum(out, Gia_ManRegNum(p))
	return out
}
func Gia_ManDupCof(p *Gia_Man_t, pivot int32) *Gia_Man_t {
	interim := Gia_ManDupCofInt(p, pivot)
	if interim == nil {
		return nil
	}
	result := Gia_ManCleanup(interim)
	Gia_ManStop(interim)
	return result
}
func Gia_ManCofVars(p *Gia_Man_t, limit int32) *Vec_Int_t {
	counts := make([]int32, p.nObjs)
	for _, obj := range p.pObjs {
		for _, fanin := range giaObjectFanins(p, obj) {
			counts[fanin]++
		}
	}
	result := vec.Vec_IntAlloc(100)
	for id, obj := range p.pObjs {
		if Gia_ObjIsCand(obj) != 0 && counts[id] >= limit {
			vec.Vec_IntPush(result, int32(id))
		}
	}
	return result
}
func Gia_ManTransfer(original, cofactored, newGraph *Gia_Man_t, signals *Vec_Int_t) *Vec_Int_t {
	result := vec.Vec_IntAlloc(vec.Vec_IntSize(signals))
	seen := make(map[int32]bool)
	for _, id := range vec.Vec_IntArray(signals) {
		if id < 0 || id >= original.nObjs {
			panic("gia: invalid cofactor signal")
		}
		lit := original.pObjs[id].Value
		if int(lit>>1) >= len(cofactored.pObjs) {
			continue
		}
		mapped := cofactored.pObjs[lit>>1].Value
		if mapped == 0 || mapped == ^uint32(0) {
			continue
		}
		target := int32(mapped >> 1)
		if !seen[target] {
			seen[target] = true
			vec.Vec_IntPush(result, target)
		}
	}
	return result
}
func Gia_ManDupCofAllInt(p *Gia_Man_t, signals *Vec_Int_t, verbose int32) *Gia_Man_t {
	if vec.Vec_IntSize(signals) > 200 {
		fmt.Println("Too many signals to cofactor.")
		return nil
	}
	current := Gia_ManDup(p)
	remaining := vec.Vec_IntDup(signals)
	for vec.Vec_IntSize(remaining) > 0 {
		sort.Slice(vec.Vec_IntArray(remaining), func(i, j int) bool { return vec.Vec_IntArray(remaining)[i] < vec.Vec_IntArray(remaining)[j] })
		pivot := vec.Vec_IntPop(remaining)
		interim := Gia_ManDupCofInt(current, pivot)
		if interim == nil {
			Gia_ManStop(current)
			return nil
		}
		next := Gia_ManCleanup(interim)
		remaining = Gia_ManTransfer(current, interim, next, remaining)
		Gia_ManStop(current)
		Gia_ManStop(interim)
		current = next
		if verbose != 0 {
			fmt.Printf("Cofactored variable %d.\n", pivot)
		}
	}
	return current
}
func Gia_ManDupCofAll(p *Gia_Man_t, limit, verbose int32) *Gia_Man_t {
	return Gia_ManDupCofAllInt(p, Gia_ManCofVars(p, limit), verbose)
}
