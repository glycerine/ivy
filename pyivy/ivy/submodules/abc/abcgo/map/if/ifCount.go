package if_

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/vec"
	"math/bits"
)

func If_LogCreateAnd(aig *Vec_Int_t, a, b, nSupport int32) int32 {
	if a>>1 == b>>1 {
		panic("if: log-counter AND operands have same variable")
	}
	id := int32(vec.Vec_IntSize(aig)/2) + nSupport
	vec.Vec_IntPush(aig, a)
	vec.Vec_IntPush(aig, b)
	return id * 2
}
func If_LogCreateMux(aig *Vec_Int_t, control, then, otherwise, nSupport int32) int32 {
	a := If_LogCreateAnd(aig, control, then, nSupport)
	b := If_LogCreateAnd(aig, control^1, otherwise, nSupport)
	return If_LogCreateAnd(aig, a^1, b^1, nSupport) ^ 1
}
func If_LogCreateXor(aig *Vec_Int_t, a, b, nSupport int32) int32 {
	return If_LogCreateMux(aig, a, b^1, b, nSupport)
}
func If_LogCreateAndXor(aig *Vec_Int_t, a, b, nSupport, xor int32) int32 {
	if xor != 0 {
		return If_LogCreateXor(aig, a, b, nSupport)
	}
	return If_LogCreateAnd(aig, a, b, nSupport)
}
func If_LogCreateAndXorMulti(aig *Vec_Int_t, lits []int32, nFanins, nSupport, xor int32) int32 {
	if nFanins <= 0 || int(nFanins) > len(lits) {
		panic("if: invalid multi-gate fanin count")
	}
	for i := nFanins - 1; i > 0; i-- {
		lits[i-1] = If_LogCreateAndXor(aig, lits[i], lits[i-1], nSupport, xor)
	}
	return lits[0]
}

// The timing queue is decreasing. Equal arrivals combine and carry into the
// preceding position, preserving source topology and paired payload movement.
func ifLogCounter(times []int32, n *int32, num, delay int32, onSwap func(int, int), onMerge func(int, int), onShift func(int, int)) int32 {
	size := int(*n)
	if size < 0 || size >= len(times) {
		panic("if: log counter queue capacity exceeded")
	}
	times[size] = num
	size++
	for k := size - 1; k > 0; k-- {
		if times[k] < times[k-1] {
			break
		}
		if times[k] > times[k-1] {
			times[k], times[k-1] = times[k-1], times[k]
			if onSwap != nil {
				onSwap(k, k-1)
			}
			continue
		}
		times[k-1] += delay
		if onMerge != nil {
			onMerge(k, k-1)
		}
		size--
		for i := k; i < size; i++ {
			times[i] = times[i+1]
			if onShift != nil {
				onShift(i, i+1)
			}
		}
	}
	*n = int32(size)
	result := times[0]
	if size > 1 {
		result += delay
	}
	return result
}
func If_LogCounterAdd(times []int32, n *int32, num, xor int32) int32 {
	return ifLogCounter(times, n, num, 1+xor, nil, nil, nil)
}
func If_LogCounterAddAig(times []int32, n *int32, lits []int32, num, lit int32, aig *Vec_Int_t, nSupport, xor, xorFunction int32) int32 {
	if aig == nil {
		return If_LogCounterAdd(times, n, num, xor)
	}
	if int(*n) >= len(lits) {
		panic("if: insufficient log-counter literal storage")
	}
	lits[*n] = lit
	return ifLogCounter(times, n, num, 1+xor, func(a, b int) { lits[a], lits[b] = lits[b], lits[a] }, func(a, b int) { lits[b] = If_LogCreateAndXor(aig, lits[a], lits[b], nSupport, xorFunction) }, func(a, b int) { lits[a] = lits[b] })
}
func If_LogCounterDelayXor(times []int32, n int32) int32 {
	if n <= 0 || int(n) > len(times) {
		panic("if: empty XOR counter")
	}
	for i := n - 1; i > 0; i-- {
		times[i-1] = 2 + max(times[i], times[i-1])
	}
	return times[0]
}
func If_CutPinDelayGet(profile uint64, pin int32) int32 {
	if pin < 0 || pin >= 16 {
		panic("if: pin outside delay profile")
	}
	return int32(profile >> uint(pin*4) & 15)
}
func If_CutPinDelaySet(profile *uint64, pin, delay int32) {
	if pin < 0 || pin >= 16 || delay < 0 || delay >= 16 {
		panic("if: invalid packed pin delay")
	}
	*profile |= uint64(delay) << uint(pin*4)
}
func If_CutPinDelayInit(pin int32) uint64 {
	if pin < 0 || pin >= 16 {
		panic("if: pin outside delay profile")
	}
	return 1 << uint(pin*4)
}
func If_CutPinDelayMax(a, b uint64, nVars, add int32) uint64 {
	if nVars < 0 || nVars > 16 {
		panic("if: invalid delay-profile width")
	}
	var result uint64
	for v := int32(0); v < nVars; v++ {
		delay := max(If_CutPinDelayGet(a, v), If_CutPinDelayGet(b, v))
		if delay != 0 {
			If_CutPinDelaySet(&result, v, min(delay+add, 15))
		}
	}
	return result
}
func If_CutPinDelayDecrement(a uint64, nVars int32) uint64 {
	var result uint64
	for v := int32(0); v < nVars; v++ {
		if delay := If_CutPinDelayGet(a, v); delay > 0 {
			If_CutPinDelaySet(&result, v, delay-1)
		}
	}
	return result
}
func If_CutPinDelayEqual(a, b uint64, nVars int32) int32 {
	for v := int32(0); v < nVars; v++ {
		if If_CutPinDelayGet(a, v) != If_CutPinDelayGet(b, v) {
			return 0
		}
	}
	return 1
}
func If_CutPinDelayDom(a, b uint64, nVars int32) int32 {
	for v := int32(0); v < nVars; v++ {
		if If_CutPinDelayGet(a, v) > If_CutPinDelayGet(b, v) {
			return 0
		}
	}
	return 1
}
func If_CutPinDelayTranslate(profile uint64, nVars int32, permutation []byte) {
	if int(nVars) > len(permutation) {
		panic("if: insufficient delay permutation storage")
	}
	for v := int32(0); v < nVars; v++ {
		delay := If_CutPinDelayGet(profile, v)
		if delay <= 1 {
			panic("if: missing gate delay")
		}
		permutation[v] = byte(delay - 1)
	}
}
func If_CutPinDelayPrint(profile uint64, nVars int32) {
	fmt.Print("Delay profile = {")
	for v := int32(0); v < nVars; v++ {
		fmt.Printf(" %d", If_CutPinDelayGet(profile, v))
	}
	fmt.Println(" }")
}
func If_LogCounterPinDelays(times []int32, n *int32, profiles []uint64, num int32, profile uint64, nSupport, xor int32) int32 {
	if int(*n) >= len(profiles) {
		panic("if: insufficient pin-delay counter storage")
	}
	profiles[*n] = profile
	return ifLogCounter(times, n, num, 1+xor, func(a, b int) { profiles[a], profiles[b] = profiles[b], profiles[a] }, func(a, b int) { profiles[b] = If_CutPinDelayMax(profiles[a], profiles[b], nSupport, 1+xor) }, func(a, b int) { profiles[a] = profiles[b] })
}
func If_LogPinDelaysMulti(profiles []uint64, nFanins, nSupport, xor int32) uint64 {
	if nFanins <= 0 || int(nFanins) > len(profiles) {
		panic("if: invalid pin-delay fanin count")
	}
	for i := nFanins - 1; i > 0; i-- {
		profiles[i-1] = If_CutPinDelayMax(profiles[i], profiles[i-1], nSupport, 1+xor)
	}
	return profiles[0]
}
func If_AigVerifyArray(aig *Vec_Int_t, nLeaves int32) uint64 {
	a := vec.Vec_IntArray(aig)
	if len(a) == 0 || a[len(a)-1] < 0 || a[len(a)-1] > 1 || nLeaves < 0 || nLeaves > 6 {
		panic("if: invalid small AIG array")
	}
	if len(a) == 1 {
		if a[0] != 0 {
			return ^uint64(0)
		}
		return 0
	}
	elementary := func(variable int32) uint64 {
		var result uint64
		for bit := 0; bit < 64; bit++ {
			if bit&(1<<uint(variable)) != 0 {
				result |= 1 << uint(bit)
			}
		}
		return result
	}
	if len(a) == 2 {
		if a[0] != 0 {
			panic("if: invalid variable AIG array")
		}
		result := elementary(0)
		if a[1] != 0 {
			result = ^result
		}
		return result
	}
	if len(a)%2 != 1 {
		panic("if: malformed gate array")
	}
	truth := make([]uint64, (len(a)-1)/2)
	for i := range truth {
		value := func(lit int32) uint64 {
			variable := lit >> 1
			var result uint64
			if variable < nLeaves {
				result = elementary(variable)
			} else {
				index := variable - nLeaves
				if index < 0 || int(index) >= i {
					panic("if: non-topological AIG array")
				}
				result = truth[index]
			}
			if lit&1 != 0 {
				return ^result
			}
			return result
		}
		truth[i] = value(a[2*i]) & value(a[2*i+1])
	}
	result := truth[len(truth)-1]
	if a[len(a)-1] != 0 {
		result = ^result
	}
	vec.Vec_IntClear(aig)
	return result
}
func If_AigPrintArray(aig *Vec_Int_t, nLeaves int32) {
	a := vec.Vec_IntArray(aig)
	if len(a) == 0 {
		panic("if: empty AIG array")
	}
	switch len(a) {
	case 1:
		fmt.Printf("Const %d\n", a[0])
	case 2:
		fmt.Printf("Variable phase=%d\n", a[1])
	default:
		for i := 0; i < len(a)-1; i += 2 {
			fmt.Printf("%d %d\n", a[i], a[i+1])
		}
		fmt.Printf("Phase %d\n", a[len(a)-1])
	}
}
func If_LogCounter64Eval(count uint64) int32 {
	if count == 0 {
		panic("if: empty logarithmic counter")
	}
	return int32(bits.Len64(count - 1))
}
func If_LogCounter64Add(count uint64, num int32) uint64 {
	if num < 0 || num >= 48 {
		panic("if: log-counter exponent outside range")
	}
	return count + 1<<uint(num)
}
func If_LogCounter32Eval(count uint32, start int32) int32 {
	if count == 0 {
		panic("if: empty logarithmic counter")
	}
	result := int32(bits.Len32(count)-1) + start/2
	if start&1 != 0 || count&(count-1) != 0 {
		result++
	}
	return result
}
func If_LogCounter32Add(count uint32, start *int32, num int32) uint32 {
	base := *start >> 1
	if num < base {
		*start |= 1
		return count
	}
	if num > base+16 {
		shift := num - (base + 16)
		discarded := count
		if shift < 32 {
			discarded = count & ((1 << uint(shift)) - 1)
		}
		if discarded != 0 {
			*start |= 1
		}
		if shift >= 32 {
			count = 0
		} else {
			count >>= uint(shift)
		}
		base += shift
		*start = base*2 + (*start & 1)
	}
	return count + 1<<uint(num-base)
}
