package int2

import (
	"time"

	"github.com/glycerine/abcgo/aig/gia"
)

// The source interpolation driver is incomplete in the checked-in C source:
// it references undeclared image/map variables and its preimage helper reads
// different objects than its parameters. For small models, exact finite-state
// reachability provides a sound native fallback instead of returning a claim
// based on the incomplete driver.
const (
	int2ReachMaxInputs = 16
	int2ReachMaxStates = 1 << 18
	int2ReachMaxSteps  = 1 << 22
)

type int2ReachState struct {
	bits  []bool
	depth int32
}

func int2StateKey(bits []bool) string {
	key := make([]byte, (len(bits)+7)/8)
	for i, bit := range bits {
		if bit {
			key[i>>3] |= 1 << uint(i&7)
		}
	}
	return string(key)
}

func int2EvalGraph(p *Gia_Man_t, inputs, state, values, next []bool) bool {
	clear(values)
	for i, bit := range inputs {
		obj := gia.Gia_ManPi(p, int32(i))
		values[gia.Gia_ObjId(p, obj)] = bit
	}
	for i, bit := range state {
		obj := gia.Gia_ManRo(p, int32(i))
		values[gia.Gia_ObjId(p, obj)] = bit
	}
	for id := 1; id < len(values); id++ {
		obj := gia.Gia_ManObj(p, int32(id))
		if gia.Gia_ObjIsAnd(obj) == 0 {
			continue
		}
		left := values[gia.Gia_ObjId(p, gia.Gia_ObjFanin0(obj))] != (gia.Gia_ObjFaninC0(obj) != 0)
		right := values[gia.Gia_ObjId(p, gia.Gia_ObjFanin1(obj))] != (gia.Gia_ObjFaninC1(obj) != 0)
		values[id] = left && right
	}
	for i := int32(0); i < gia.Gia_ManPoNum(p); i++ {
		co := gia.Gia_ManPo(p, i)
		bad := values[gia.Gia_ObjId(p, gia.Gia_ObjFanin0(co))] != (gia.Gia_ObjFaninC0(co) != 0)
		if bad {
			return true
		}
	}
	for i := range next {
		co := gia.Gia_ManRi(p, int32(i))
		next[i] = values[gia.Gia_ObjId(p, gia.Gia_ObjFanin0(co))] != (gia.Gia_ObjFaninC0(co) != 0)
	}
	return false
}

// int2Reachability checks the transition system from the all-zero initial
// state. It returns proven, counterexample, or undecided. A depth or resource
// limit can only produce undecided; it never produces a proof.
func int2Reachability(p *Gia_Man_t, maxDepth int32, deadline time.Time) int32 {
	if p == nil || maxDepth < 0 || gia.Gia_ManRegNum(p) <= 0 || gia.Gia_ManPoNum(p) <= 0 {
		return INT2_TIME_OUT
	}
	nPis := gia.Gia_ManPiNum(p)
	if nPis < 0 || nPis > int2ReachMaxInputs {
		return INT2_TIME_OUT
	}
	initial := make([]bool, gia.Gia_ManRegNum(p))
	queue := []int2ReachState{{bits: initial}}
	seen := map[string]struct{}{int2StateKey(initial): {}}
	inputCount := 1 << uint(nPis)
	inputs := make([]bool, nPis)
	values := make([]bool, gia.Gia_ManObjNum(p))
	next := make([]bool, gia.Gia_ManRegNum(p))
	steps := 0
	for head := 0; head < len(queue); head++ {
		if !deadline.IsZero() && time.Now().After(deadline) {
			return INT2_TIME_OUT
		}
		current := queue[head]
		for inputMask := 0; inputMask < inputCount; inputMask++ {
			steps++
			if steps > int2ReachMaxSteps {
				return INT2_TIME_OUT
			}
			if steps&255 == 0 && !deadline.IsZero() && time.Now().After(deadline) {
				return INT2_TIME_OUT
			}
			for i := range inputs {
				inputs[i] = inputMask&(1<<uint(i)) != 0
			}
			if int2EvalGraph(p, inputs, current.bits, values, next) {
				return INT2_FALSE_NEG
			}
			key := int2StateKey(next)
			if _, ok := seen[key]; ok {
				continue
			}
			if maxDepth > 0 && current.depth >= maxDepth {
				return INT2_TIME_OUT
			}
			if len(seen) >= int2ReachMaxStates {
				return INT2_TIME_OUT
			}
			seen[key] = struct{}{}
			queue = append(queue, int2ReachState{bits: append([]bool(nil), next...), depth: current.depth + 1})
		}
	}
	return INT2_COMPUTED
}
