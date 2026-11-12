package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/vec"
)

func Gia_ManPrintStateEncoding(codes *Vec_Vec_t, bits int32) {
	for i := 0; i < vec.Vec_VecSize(codes); i++ {
		s := make([]byte, bits)
		for j := range s {
			s[j] = '-'
		}
		for _, b := range vec.Vec_IntArray(vec.Vec_VecEntryInt(codes, i)) {
			if b < 0 || b >= bits {
				panic("gia: invalid state code")
			}
			s[b] = '1'
		}
		fmt.Printf("%6d : %s\n", i+1, s)
	}
}
func Gia_ManCreateOrGate(p *Gia_Man_t, lits *Vec_Int_t) int32 {
	for vec.Vec_IntSize(lits) > 1 {
		data := vec.Vec_IntArray(lits)
		count := 0
		for i := 0; i+1 < len(data); i += 2 {
			data[count] = Gia_ManHashOr(p, data[i], data[i+1])
			count++
		}
		if len(data)%2 != 0 {
			data[count] = data[len(data)-1]
			count++
		}
		vec.Vec_IntShrink(lits, count)
	}
	if vec.Vec_IntSize(lits) == 0 {
		return 0
	}
	return vec.Vec_IntEntry(lits, 0)
}
func Gia_ManAssignCodes(hot, states int32, nBits *int32) *Vec_Vec_t {
	if states <= 0 || hot < 1 || hot > 5 || nBits == nil {
		panic("gia: invalid hot state encoding")
	}
	bits := hot
	choose := func(n, k int32) uint64 {
		value := uint64(1)
		for i := int32(1); i <= k; i++ {
			value = value * uint64(n-k+i) / uint64(i)
			if value >= uint64(states) {
				return value
			}
		}
		return value
	}
	for choose(bits, hot) < uint64(states) {
		bits++
	}
	*nBits = bits
	codes := vec.Vec_VecAlloc(int(states))
	vec.Vec_VecExpandInt(codes, int(states)-1)
	current := make([]int32, 0, hot)
	state := 0
	var generate func(int32)
	generate = func(start int32) {
		if state >= int(states) {
			return
		}
		if int32(len(current)) == hot {
			for _, b := range current {
				vec.Vec_VecPushInt(codes, state, b)
			}
			state++
			return
		}
		for b := start; b < bits; b++ {
			current = append(current, b)
			generate(b + 1)
			current = current[:len(current)-1]
			if state >= int(states) {
				return
			}
		}
	}
	generate(0)
	return codes
}
