package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/vec"
)

func Gia_Gen2CreateMux_rec(p *Gia_Man_t, controls []int32, n int32, data *Vec_Int_t, offset int32) int32 {
	if n < 0 || int(n) > len(controls) {
		panic("gia: invalid QBF mux controls")
	}
	if n == 0 {
		return vec.Vec_IntEntry(data, int(offset))
	}
	lo := Gia_Gen2CreateMux_rec(p, controls, n-1, data, offset)
	hi := Gia_Gen2CreateMux_rec(p, controls, n-1, data, offset+(1<<uint(n-1)))
	return Gia_ManHashMux(p, controls[n-1], hi, lo)
}
func Gia_Gen2CreateMuxes(p *Gia_Man_t, size, count int32, parameters, inputs *Vec_Int_t) *Vec_Int_t {
	if size < 0 || size > 16 || count < 0 || vec.Vec_IntSize(inputs) != int(size) || vec.Vec_IntSize(parameters) != int(count)*(1<<uint(size)) {
		panic("gia: invalid QBF mux arrays")
	}
	result := vec.Vec_IntAlloc(int(count))
	for i := int32(0); i < count; i++ {
		vec.Vec_IntPush(result, Gia_Gen2CreateMux_rec(p, vec.Vec_IntArray(inputs), size, parameters, i*(1<<uint(size))))
	}
	return result
}
func Gia_Gen2CreateMiter(size, count int32) *Gia_Man_t {
	if size < 0 || size > 16 || count < 0 {
		panic("gia: invalid QBF miter size")
	}
	width := int32(1) << uint(size)
	params := vec.Vec_IntAlloc(int(count * width))
	x, y, xy := vec.Vec_IntAlloc(int(size)), vec.Vec_IntAlloc(int(size)), vec.Vec_IntAlloc(int(size))
	p := Gia_ManStart(1000)
	p.pName = "homoqbf"
	Gia_ManHashAlloc(p)
	for i := int32(0); i < count*width; i++ {
		vec.Vec_IntPush(params, Gia_ManAppendCi(p))
	}
	for i := int32(0); i < size; i++ {
		vec.Vec_IntPush(x, Gia_ManAppendCi(p))
	}
	for i := int32(0); i < size; i++ {
		vec.Vec_IntPush(y, Gia_ManAppendCi(p))
	}
	for i := int32(0); i < size; i++ {
		vec.Vec_IntPush(xy, Gia_ManHashAnd(p, vec.Vec_IntEntry(x, int(i)), vec.Vec_IntEntry(y, int(i)))^1)
	}
	xr, yr, xyr := Gia_Gen2CreateMuxes(p, size, count, params, x), Gia_Gen2CreateMuxes(p, size, count, params, y), Gia_Gen2CreateMuxes(p, size, count, params, xy)
	pairs := vec.Vec_IntAlloc(2 * int(count))
	for i := 0; i < int(count); i++ {
		vec.Vec_IntPush(pairs, vec.Vec_IntEntry(xyr, i))
		vec.Vec_IntPush(pairs, Gia_ManHashAnd(p, vec.Vec_IntEntry(xr, i), vec.Vec_IntEntry(yr, i))^1)
	}
	result := Gia_ManHashDualMiter(p, pairs)
	for i := int32(0); i < width; i++ {
		for j := i + 1; j < width; j++ {
			vec.Vec_IntClear(pairs)
			for k := int32(0); k < count; k++ {
				vec.Vec_IntPush(pairs, vec.Vec_IntEntry(params, int(k*width+i)))
				vec.Vec_IntPush(pairs, vec.Vec_IntEntry(params, int(k*width+j)))
			}
			result = Gia_ManHashOr(p, result, Gia_ManHashDualMiter(p, pairs)^1)
		}
	}
	Gia_ManAppendCo(p, result^1)
	Gia_ManHashStop(p)
	out := Gia_ManCleanup(p)
	Gia_ManStop(p)
	fmt.Printf("Generated QBF miter with %d parameters, %d functional variables, and %d AIG nodes.\n", count*width, 2*size, Gia_ManAndNum(out))
	return out
}
func Gia_Gen2CodeOne(size, count int32, code *Vec_Int_t, x int32) int32 {
	if count < 0 || count > 31 {
		panic("gia: QBF code does not fit int32")
	}
	result := int32(0)
	for k := int32(0); k < count; k++ {
		if vec.Vec_IntEntry(code, int(k*(1<<uint(size))+x)) != 0 {
			result |= 1 << uint(k)
		}
	}
	return result
}
func Gia_Gen2CodeOneP(size, count int32, code *Vec_Int_t, x int32) []word {
	if size < 0 || size > 30 || count < 0 || x < 0 || x >= 1<<uint(size) {
		panic("gia: invalid QBF code index")
	}
	result := make([]word, (int(count)+63)/64)
	for k := int32(0); k < count; k++ {
		if vec.Vec_IntEntry(code, int(k*(1<<uint(size))+x)) != 0 {
			result[k>>6] |= 1 << uint(k&63)
		}
	}
	return result
}
