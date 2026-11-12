// Truth-table simulation, ported from ABC hopTruth.c (Alan Mishchenko).
package hop

import "github.com/glycerine/abcgo/misc/vec"

type Vec_Int_t = vec.Vec_Int_t
type word = uint64

func Hop_ManTruthWordNum(n int32) int32 { return Hop_TruthWordNum(n) }
func Hop_ManTruthCopy(out, in []uint32, n int32) {
	copy(out[:Hop_TruthWordNum(n)], in[:Hop_TruthWordNum(n)])
}
func Hop_ManTruthClear(out []uint32, n int32) { clear(out[:Hop_TruthWordNum(n)]) }
func Hop_ManTruthFill(out []uint32, n int32) {
	for i := int32(0); i < Hop_TruthWordNum(n); i++ {
		out[i] = ^uint32(0)
	}
}
func Hop_ManTruthNot(out, in []uint32, n int32) {
	for i := int32(0); i < Hop_TruthWordNum(n); i++ {
		out[i] = ^in[i]
	}
}
func Hop_ManConvertAigToTruth_rec1(o *Hop_Obj_t) int32 {
	if Hop_ObjIsNode(o) == 0 || o.fMarkA != 0 {
		return 0
	}
	n := Hop_ManConvertAigToTruth_rec1(Hop_ObjFanin0(o)) + Hop_ManConvertAigToTruth_rec1(Hop_ObjFanin1(o))
	o.fMarkA = 1
	return n + 1
}
func Hop_ManConvertAigToTruth_rec2(o *Hop_Obj_t, scratch *Vec_Int_t, n int32) []uint32 {
	if Hop_ObjIsNode(o) == 0 || o.fMarkA == 0 {
		return o.pData.([]uint32)
	}
	a, b := Hop_ManConvertAigToTruth_rec2(Hop_ObjFanin0(o), scratch, n), Hop_ManConvertAigToTruth_rec2(Hop_ObjFanin1(o), scratch, n)
	out := make([]uint32, n)
	for i := range out {
		x, y := a[i], b[i]
		if Hop_ObjFaninC0(o) != 0 {
			x = ^x
		}
		if Hop_ObjFaninC1(o) != 0 {
			y = ^y
		}
		if o.Type == AIG_EXOR {
			out[i] = x ^ y
		} else {
			out[i] = x & y
		}
	}
	o.pData = out
	o.fMarkA = 0
	return out
}
func Hop_ManConvertAigToTruth(p *Hop_Man_t, root *Hop_Obj_t, n int32, scratch *Vec_Int_t, msb int32) []uint32 {
	if n < 0 || n > 30 || (Hop_Regular(root) != p.pConst1 && n > Hop_ManPiNum(p)) {
		panic("invalid truth-table variable count")
	}
	words := Hop_TruthWordNum(n)
	if Hop_Regular(root) == p.pConst1 {
		out := make([]uint32, words)
		if Hop_IsComplement(root) == 0 {
			for i := range out {
				out[i] = ^uint32(0)
			}
		}
		hopTruthSaveScratch(scratch, out)
		return out
	}
	hopValidateTruthSupport(p, root, n)
	for i := int32(0); i < n; i++ {
		table := make([]uint32, words)
		bit := i
		if msb != 0 {
			bit = n - 1 - i
		}
		patterns := [...]uint32{0xaaaaaaaa, 0xcccccccc, 0xf0f0f0f0, 0xff00ff00, 0xffff0000}
		for k := int32(0); k < words; k++ {
			if bit < 5 {
				table[k] = patterns[bit]
			} else if k&(1<<uint(bit-5)) != 0 {
				table[k] = ^uint32(0)
			}
		}
		Hop_ManPi(p, i).pData = table
	}
	constant := make([]uint32, words)
	for i := range constant {
		constant[i] = ^uint32(0)
	}
	p.pConst1.pData = constant
	Hop_ManConvertAigToTruth_rec1(Hop_Regular(root))
	data := Hop_ManConvertAigToTruth_rec2(Hop_Regular(root), scratch, words)
	out := append([]uint32(nil), data...)
	if Hop_IsComplement(root) != 0 {
		for i := range out {
			out[i] = ^out[i]
		}
	}
	hopTruthSaveScratch(scratch, out)
	return out
}

var hopElementary = [8]uint64{0xaaaaaaaaaaaaaaaa, 0xcccccccccccccccc, 0xf0f0f0f0f0f0f0f0, 0xff00ff00ff00ff00, 0xffff0000ffff0000, 0xffffffff00000000, 0, ^uint64(0)}

func Hop_ManComputeTruth6_rec(p *Hop_Man_t, o *Hop_Obj_t) word {
	return hopComputeTruth6Memo(o, make(map[*Hop_Obj_t]uint64))
}
func hopComputeTruth6Memo(o *Hop_Obj_t, memo map[*Hop_Obj_t]uint64) uint64 {
	if v, ok := memo[o]; ok {
		return v
	}
	var value uint64
	switch o.Type {
	case AIG_CONST1:
		value = ^uint64(0)
	case AIG_PI:
		value = hopElementary[o.iData]
	case AIG_AND, AIG_EXOR:
		a, b := hopComputeTruth6Memo(Hop_ObjFanin0(o), memo), hopComputeTruth6Memo(Hop_ObjFanin1(o), memo)
		if Hop_ObjFaninC0(o) != 0 {
			a = ^a
		}
		if Hop_ObjFaninC1(o) != 0 {
			b = ^b
		}
		if o.Type == AIG_EXOR {
			value = a ^ b
		} else {
			value = a & b
		}
	default:
		panic("invalid Hop truth6 node")
	}
	memo[o] = value
	return value
}
func Hop_ManComputeTruth6(p *Hop_Man_t, o *Hop_Obj_t, n int32) word {
	if n < 0 || n > 6 || (Hop_Regular(o) != p.pConst1 && n > Hop_ManPiNum(p)) {
		panic("truth6 requires at most six variables")
	}
	if Hop_Regular(o) == p.pConst1 {
		if Hop_IsComplement(o) != 0 {
			return 0
		}
		return ^uint64(0)
	}
	hopValidateTruthSupport(p, o, n)
	for i := int32(0); i < n; i++ {
		Hop_ManPi(p, i).iData = i
	}
	v := Hop_ManComputeTruth6_rec(p, Hop_Regular(o))
	if Hop_IsComplement(o) != 0 {
		return ^v
	}
	return v
}

// Validate support before mutating scratch fields so repeated calls cannot use
// stale PI truth tables when n omits a reachable input. Also reject cycles.
func hopValidateTruthSupport(p *Hop_Man_t, root *Hop_Obj_t, n int32) {
	allowed := make(map[*Hop_Obj_t]bool, n)
	for i := int32(0); i < n; i++ {
		allowed[Hop_ManPi(p, i)] = true
	}
	state := map[*Hop_Obj_t]uint8{}
	var visit func(*Hop_Obj_t)
	visit = func(edge *Hop_Obj_t) {
		o := Hop_Regular(edge)
		if o == nil {
			panic("nil Hop truth-table fanin")
		}
		if state[o] == 1 {
			panic("cycle in Hop truth-table support")
		}
		if state[o] == 2 {
			return
		}
		switch o.Type {
		case AIG_CONST1:
			state[o] = 2
			return
		case AIG_PI:
			if !allowed[o] {
				panic("Hop truth-table support exceeds variable count")
			}
			state[o] = 2
			return
		case AIG_AND, AIG_EXOR:
		default:
			panic("invalid Hop truth-table object")
		}
		state[o] = 1
		visit(o.pFanin0)
		visit(o.pFanin1)
		state[o] = 2
	}
	visit(root)
}
func hopTruthSaveScratch(scratch *Vec_Int_t, out []uint32) {
	if scratch == nil {
		return
	}
	vec.Vec_IntClear(scratch)
	for _, v := range out {
		vec.Vec_IntPush(scratch, int32(v))
	}
}
