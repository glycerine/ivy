package gia

import "github.com/glycerine/abcgo/misc/vec"

func Gia_ManStgKHot(lines *Vec_Int_t, ins, outs, states, hot, verbose int32) *Gia_Man_t {
	var bits int32
	codes := Gia_ManAssignCodes(hot, states, &bits)
	if verbose != 0 {
		Gia_ManPrintStateEncoding(codes, bits)
	}
	return giaBuildStg(lines, ins, outs, states, hot, bits, codes, false)
}
func Gia_ManStgOneHot(lines *Vec_Int_t, ins, outs, states int32) *Gia_Man_t {
	var bits int32
	codes := Gia_ManAssignCodes(1, states, &bits)
	return giaBuildStg(lines, ins, outs, states, 1, bits, codes, true)
}
func giaBuildStg(lines *Vec_Int_t, ins, outs, states, hot, bits int32, codes *Vec_Vec_t, oneBased bool) *Gia_Man_t {
	data := vec.Vec_IntArray(lines)
	if len(data)%4 != 0 || ins < 0 || ins > 20 || outs < 0 || outs > 30 {
		panic("gia: invalid state transition data")
	}
	p := Gia_ManStart(10000)
	p.pName = "stg"
	for i := int32(0); i < ins+bits; i++ {
		Gia_ManAppendCi(p)
	}
	Gia_ManHashStart(p)
	minterms := make([]int32, 1<<uint(ins))
	for i := range minterms {
		lit := int32(1)
		for b := int32(0); b < ins; b++ {
			input := 2 * (b + 1)
			if i&(1<<uint(b)) == 0 {
				input ^= 1
			}
			lit = Gia_ManHashAnd(p, lit, input)
		}
		minterms[i] = lit
	}
	curs := make([]int32, states)
	for i := range curs {
		lit := int32(1)
		for _, b := range vec.Vec_IntArray(vec.Vec_VecEntryInt(codes, i)) {
			input := 2 * (1 + ins + b)
			if b < hot {
				input ^= 1
			}
			lit = Gia_ManHashAnd(p, lit, input)
		}
		curs[i] = lit
	}
	next := make([]*Vec_Int_t, bits)
	outputs := make([]*Vec_Int_t, outs)
	for i := range next {
		next[i] = vec.Vec_IntAlloc(0)
	}
	for i := range outputs {
		outputs[i] = vec.Vec_IntAlloc(0)
	}
	for i := 0; i < len(data); i += 4 {
		mint, cur, dest, out := data[i], data[i+1], data[i+2], data[i+3]
		if oneBased {
			cur--
			dest--
		}
		if mint < 0 || int(mint) >= len(minterms) || cur < 0 || cur >= states || dest < 0 || dest >= states || out < 0 || uint32(out) >= uint32(1)<<uint(outs) {
			panic("gia: invalid state transition")
		}
		condition := Gia_ManHashAnd(p, minterms[mint], curs[cur])
		for _, b := range vec.Vec_IntArray(vec.Vec_VecEntryInt(codes, int(dest))) {
			vec.Vec_IntPush(next[b], condition)
		}
		for b := int32(0); b < outs; b++ {
			if out&(1<<uint(b)) != 0 {
				vec.Vec_IntPush(outputs[b], condition)
			}
		}
	}
	for _, lits := range outputs {
		Gia_ManAppendCo(p, Gia_ManCreateOrGate(p, lits))
	}
	for i, lits := range next {
		lit := Gia_ManCreateOrGate(p, lits)
		if int32(i) < hot {
			lit ^= 1
		}
		Gia_ManAppendCo(p, lit)
	}
	Gia_ManSetRegNum(p, bits)
	Gia_ManHashStop(p)
	return Gia_ManCleanup(p)
}
