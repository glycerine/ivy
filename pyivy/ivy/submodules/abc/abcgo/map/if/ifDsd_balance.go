package if_

import "github.com/glycerine/abcgo/misc/vec"

// If_CutDsdBalancePinDelays_rec follows canonical variable occurrences in
// source order. Fanin arity and the complete cut's pin-profile width differ
// when a prime fanin is itself a nontrivial DSD structure.
func If_CutDsdBalancePinDelays_rec(p *If_DsdMan_t, id int32, times []int32, result *word, support *int32, all int32, perm []byte) int32 {
	o := If_DsdVecObj(&p.vObjs, id)
	if o.Type == IF_DSD_VAR {
		pin := int32(perm[*support] >> 1)
		*support++
		*result = If_CutPinDelayInit(pin)
		return times[pin]
	}
	var profiles [16]word
	var delays, counters [16]int32
	if o.Type == IF_DSD_MUX || o.Type == IF_DSD_PRIME {
		if o.Type == IF_DSD_PRIME {
			cover := vec.Vec_WecEntry(p.vIsops[o.nFans], int(If_DsdObjTruthId(p, o)))
			assert(Vec_IntSize(cover) > 0, "prime DSD has a usable ISOP")
		}
		for i, fan := range o.pFans {
			delays[i] = If_CutDsdBalancePinDelays_rec(p, fan>>1, times, &profiles[i], support, all, perm)
		}
		if o.Type == IF_DSD_MUX {
			left := If_CutPinDelayMax(profiles[0], profiles[1], all, 1)
			right := If_CutPinDelayMax(profiles[0], profiles[2], all, 1)
			*result = If_CutPinDelayMax(left, right, all, 1)
			return 2 + max(delays[0], delays[1], delays[2])
		}
		cover := vec.Vec_WecEntry(p.vIsops[o.nFans], int(If_DsdObjTruthId(p, o)))
		return ifCutSopBalancePinDelaysInt(cover, delays[:], profiles[:], o.nFans, all, result)
	}
	assert(o.Type == IF_DSD_AND || o.Type == IF_DSD_XOR, "associative DSD profile node")
	var count, delay int32
	for i, fan := range o.pFans {
		arrival := If_CutDsdBalancePinDelays_rec(p, fan>>1, times, &profiles[i], support, all, perm)
		delay = If_LogCounterPinDelays(counters[:], &count, profiles[:], arrival, profiles[i], all, 0)
	}
	assert(count > 0, "nonempty DSD profile queue")
	*result = If_LogPinDelaysMulti(profiles[:], count, all, 0)
	return delay
}

// If_CutDsdBalanceEval_rec preserves the original interleaving of child
// construction and counter carries. XORs retain the source's unit delay/area
// estimate while emitting their actual three-AND implementation.
func If_CutDsdBalanceEval_rec(p *If_DsdMan_t, id int32, times []int32, support *int32, aig *Vec_Int_t, result *int32, all int32, area *int32, perm []byte) int32 {
	o := If_DsdVecObj(&p.vObjs, id)
	if o.Type == IF_DSD_VAR {
		lit := int32(perm[*support])
		*support++
		if aig != nil {
			*result = lit
		}
		return times[lit>>1]
	}
	if o.Type == IF_DSD_PRIME {
		cover := vec.Vec_WecEntry(p.vIsops[o.nFans], int(If_DsdObjTruthId(p, o)))
		if Vec_IntSize(cover) == 0 {
			return -1
		}
	}
	var delays, literals, counters [16]int32
	var count, delay int32
	for i, fan := range o.pFans {
		delays[i] = If_CutDsdBalanceEval_rec(p, fan>>1, times, support, aig, &literals[i], all, area, perm)
		if delays[i] < 0 {
			return -1
		}
		if aig != nil {
			literals[i] ^= fan & 1
		}
		if o.Type == IF_DSD_AND || o.Type == IF_DSD_XOR {
			xor := int32(0)
			if o.Type == IF_DSD_XOR {
				xor = 1
			}
			lit := int32(-1)
			if aig != nil {
				lit = literals[i]
			}
			delay = If_LogCounterAddAig(counters[:], &count, literals[:], delays[i], lit, aig, all, 0, xor)
		}
	}
	switch o.Type {
	case IF_DSD_MUX:
		if aig != nil {
			*result = If_LogCreateMux(aig, literals[0], literals[1], literals[2], all)
		} else {
			*area += 3
		}
		return 2 + max(delays[0], delays[1], delays[2])
	case IF_DSD_PRIME:
		truthID := If_DsdObjTruthId(p, o)
		cover := vec.Vec_WecEntry(p.vIsops[o.nFans], int(truthID))
		complemented := false
		if int(truthID) < len(p.vIsopComplement[o.nFans]) {
			complemented = p.vIsopComplement[o.nFans][truthID]
		}
		var primeArea int32
		delay := ifCutSopBalanceEvalInt(cover, delays[:], literals[:], aig, result, o.nFans, all, &primeArea, complemented)
		// C accumulates abstract area only in timing-only calls. Construction
		// callers use the emitted vector; do not overwrite the source cost policy.
		if aig == nil {
			*area += primeArea
		}
		return delay
	case IF_DSD_AND, IF_DSD_XOR:
		assert(count > 0, "nonempty DSD construction queue")
		xor := int32(0)
		if o.Type == IF_DSD_XOR {
			xor = 1
		}
		if aig != nil {
			*result = If_LogCreateAndXorMulti(aig, literals[:], count, all, xor)
		} else {
			*area += o.nFans - 1
		}
		return delay
	}
	panic("if: unsupported canonical DSD balancing node")
}
func If_CutDsdBalanceEvalInt(p *If_DsdMan_t, lit int32, times []int32, aig *Vec_Int_t, area *int32, perm []byte) int32 {
	all := dsdObject(p, lit).nSupp
	var support, result int32
	delay := If_CutDsdBalanceEval_rec(p, lit>>1, times, &support, aig, &result, all, area, perm)
	if delay < 0 {
		return -1
	}
	assert(support == all, "balanced DSD visits complete support")
	if aig != nil {
		assert(result>>1 == support+int32(Vec_IntSize(aig)/2)-1, "balanced DSD root is final")
		Vec_IntPush(aig, (result^lit)&1)
		assert(Vec_IntSize(aig)&1 != 0, "balanced DSD has final phase")
	}
	return delay
}
