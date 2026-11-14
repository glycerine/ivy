package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/vec"
	"strconv"
	"strings"
	"unicode"
)

// Gia_GenCollectFlopIndexes accepts an empty string for the default sequence.
func Gia_GenCollectFlopIndexes(text string, count, size, flops int32) *Vec_Int_t {
	if count < 0 || size < 0 || flops < 0 || int64(count)*int64(size) > int64(flops) {
		panic("gia: invalid QBF register selection")
	}
	expected := int(count * size)
	result := vec.Vec_IntAlloc(expected)
	if text == "" {
		for i := 0; i < expected; i++ {
			vec.Vec_IntPush(result, int32(i))
		}
		return result
	}
	fields := strings.FieldsFunc(text, func(r rune) bool { return r == ',' || unicode.IsSpace(r) })
	seen := make(map[int32]bool)
	for _, field := range fields {
		n, err := strconv.ParseInt(field, 10, 32)
		if err != nil || n < 0 || n >= int64(flops) {
			fmt.Printf("Invalid flop index (%s); number of flops is %d.\n", field, flops)
			return nil
		}
		id := int32(n)
		if seen[id] {
			fmt.Println("Gia_GenCollectFlopIndexes: There are duplicated flops in the list.")
			return nil
		}
		seen[id] = true
		vec.Vec_IntPush(result, id)
	}
	if len(fields) != expected {
		fmt.Printf("Gia_GenCollectFlopIndexes: Expecting %d flop indexes (instead of %d).\n", expected, len(fields))
		return nil
	}
	return result
}
func Gia_GenCreateMux_rec(p *Gia_Man_t, controls []int32, n int32, data *Vec_Int_t, offset int32) int32 {
	return Gia_Gen2CreateMux_rec(p, controls, n, data, offset)
}
func Gia_GenCreateMuxes(p, q *Gia_Man_t, flops *Vec_Int_t, count, size int32, parameters *Vec_Int_t, useRi int32) *Vec_Int_t {
	if size < 0 || size > 16 || count < 0 || vec.Vec_IntSize(flops) != int(count*size) {
		panic("gia: invalid QBF register muxes")
	}
	result := vec.Vec_IntAlloc(int(count))
	controls := make([]int32, size)
	for i := int32(0); i < count; i++ {
		for k := int32(0); k < size; k++ {
			id := vec.Vec_IntEntry(flops, int(i*size+k))
			if id < 0 || id >= Gia_ManRegNum(p) {
				panic("gia: invalid QBF register")
			}
			obj := Gia_ManRo(p, id)
			if useRi != 0 {
				obj = Gia_ManRi(p, id)
			}
			controls[k] = int32(obj.Value)
		}
		vec.Vec_IntPush(result, Gia_GenCreateMux_rec(q, controls, size, parameters, i*(1<<uint(size))))
	}
	return result
}
func Gia_GenQbfMiter(p *Gia_Man_t, frames, count, size int32, text string, useOutput, verbose int32) *Gia_Man_t {
	_ = frames
	_ = verbose
	if p == nil || size < 0 || size > 16 || count <= 0 {
		panic("gia: invalid QBF sequential miter")
	}
	flops := Gia_GenCollectFlopIndexes(text, count, size, Gia_ManRegNum(p))
	if flops == nil {
		return nil
	}
	nPars := count * (1 << uint(size))
	params := vec.Vec_IntAlloc(int(nPars))
	for i := int32(0); i < nPars; i++ {
		lit := int32(1)
		if i != 0 {
			lit = 2 * (i + 1)
		}
		vec.Vec_IntPush(params, lit)
	}
	q := Gia_ManStart(p.nObjs)
	q.pName, q.PSpec = p.pName, p.PSpec
	Gia_ManHashAlloc(q)
	Gia_ManConst0(p).Value = 0
	for i := int32(0); i < nPars; i++ {
		Gia_ManAppendCi(q)
	}
	for _, id := range vec.Vec_IntArray(p.vCis) {
		Gia_ManObj(p, id).Value = uint32(Gia_ManAppendCi(q))
	}
	for _, obj := range p.pObjs {
		if Gia_ObjIsAnd(obj) != 0 {
			obj.Value = uint32(Gia_ManHashAnd(q, Gia_ObjFanin0Copy(obj), Gia_ObjFanin1Copy(obj)))
		}
	}
	for _, id := range vec.Vec_IntArray(p.vCos) {
		obj := Gia_ManObj(p, id)
		obj.Value = uint32(Gia_ObjFanin0Copy(obj))
	}
	initial := Gia_GenCreateMuxes(p, q, flops, count, size, params, 0)
	next := Gia_GenCreateMuxes(p, q, flops, count, size, params, 1)
	first, second := vec.Vec_IntEntry(initial, 0), vec.Vec_IntEntry(next, 0)
	if useOutput != 0 {
		if Gia_ManPoNum(p) == 0 {
			panic("gia: QBF miter requires an output")
		}
		second = Gia_ManHashAnd(q, second, Gia_ObjFanin0Copy(Gia_ManPo(p, Gia_ManPoNum(p)-1))^1)
	}
	miter := Gia_ManHashAnd(q, first, second^1)
	miter = Gia_ManHashAnd(q, miter^1, 2)
	Gia_ManAppendCo(q, miter)
	result := Gia_ManCleanup(q)
	Gia_ManStop(q)
	return result
}
