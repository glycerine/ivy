package gia

import (
	"fmt"

	"github.com/glycerine/abcgo/aig/miniaig"
	"github.com/glycerine/abcgo/misc/vec"
)

func Gia_MiniAigSuperGates_rec(p *miniaig.Mini_Aig_t, id int32, result, mapping *Vec_Int_t) {
	if miniaig.Mini_AigNodeIsPi(p, int(id)) {
		v := vec.Vec_IntEntry(mapping, int(id))
		if v < 0 {
			panic("gia: unmapped MiniAIG input")
		}
		vec.Vec_IntPush(result, v)
		return
	}
	if !miniaig.Mini_AigNodeIsAnd(p, int(id)) {
		panic("gia: supergate requires AND nodes")
	}
	a, b := miniaig.Mini_AigNodeFanin0(p, int(id)), miniaig.Mini_AigNodeFanin1(p, int(id))
	if a&1 != 0 || b&1 != 0 {
		panic("gia: complemented supergate edge")
	}
	Gia_MiniAigSuperGates_rec(p, int32(a>>1), result, mapping)
	Gia_MiniAigSuperGates_rec(p, int32(b>>1), result, mapping)
}
func Gia_MiniAigSuperGates(p *miniaig.Mini_Aig_t) *Vec_Wec_t {
	result := vec.Vec_WecStart(miniaig.Mini_AigPoNum(p))
	mapping := vec.Vec_IntStartFull(miniaig.Mini_AigNodeNum(p))
	index := 0
	for i := 1; i < miniaig.Mini_AigNodeNum(p); i++ {
		if miniaig.Mini_AigNodeIsPi(p, i) {
			vec.Vec_IntWriteEntry(mapping, i, int32(index))
			index++
		}
	}
	index = 0
	for i := 1; i < miniaig.Mini_AigNodeNum(p); i++ {
		if miniaig.Mini_AigNodeIsPo(p, i) {
			fan := miniaig.Mini_AigNodeFanin0(p, i)
			if fan&1 != 0 {
				panic("gia: complemented supergate output")
			}
			Gia_MiniAigSuperGates_rec(p, int32(fan>>1), vec.Vec_WecEntry(result, index), mapping)
			index++
		}
	}
	return result
}
func Gia_MiniAigSuperPrintDouble(p *Vec_Int_t, inputs int32) {
	fmt.Println()
	for _, v := range vec.Vec_IntArray(p) {
		fmt.Printf("%d(%d) ", v%inputs, v/inputs)
	}
	fmt.Printf("  Total = %d\n", vec.Vec_IntSize(p))
}
func Gia_MiniAigSuperMerge(p *Vec_Int_t, inputs int32) int32 {
	if inputs <= 0 {
		panic("gia: invalid input count")
	}
	write, prev, changed := 0, int32(-1), int32(0)
	for _, v := range vec.Vec_IntArray(p) {
		if prev == v {
			vec.Vec_IntWriteEntry(p, write, (v/inputs+1)*inputs+v%inputs)
			write++
			prev = -1
			changed = 1
		} else {
			if prev >= 0 {
				vec.Vec_IntWriteEntry(p, write, prev)
				write++
			}
			prev = v
		}
	}
	if prev >= 0 {
		vec.Vec_IntWriteEntry(p, write, prev)
		write++
	}
	vec.Vec_IntShrink(p, write)
	return changed
}
func Gia_MiniAigSuperPreprocess(_ *miniaig.Mini_Aig_t, super *Vec_Wec_t, inputs, verbose int32) int32 {
	multi := int32(1)
	for i := 0; i < vec.Vec_WecSize(super); i++ {
		row := vec.Vec_WecEntry(super, i)
		vec.Vec_IntSort(row, 0)
		if verbose != 0 {
			fmt.Printf("\nOutput %d\n", i)
			Gia_MiniAigSuperPrintDouble(row, inputs)
		}
		iterations := int32(1)
		for Gia_MiniAigSuperMerge(row, inputs) != 0 {
			if verbose != 0 {
				Gia_MiniAigSuperPrintDouble(row, inputs)
			}
			iterations++
		}
		multi = max(multi, iterations)
	}
	if verbose != 0 {
		fmt.Printf("Multi = %d.\n", multi)
	}
	return multi
}
func Gia_MiniAigSuperDeriveGia(super *Vec_Wec_t, inputs, multi int32) *Gia_Man_t {
	if inputs <= 0 || multi <= 0 {
		panic("gia: invalid supergate dimensions")
	}
	g := Gia_ManStart(1000)
	g.pName = "tree"
	total := inputs * multi
	for i := int32(0); i < total; i++ {
		Gia_ManAppendCi(g)
	}
	Gia_ManHashAlloc(g)
	var outputs []int32
	for i := 0; i < vec.Vec_WecSize(super); i++ {
		lits := vec.Vec_IntAlloc(16)
		for _, v := range vec.Vec_IntArray(vec.Vec_WecEntry(super, i)) {
			if v < 0 || v >= total {
				panic("gia: supergate input out of range")
			}
			vec.Vec_IntPush(lits, 2+2*((v%inputs)*multi+v/inputs))
		}
		outputs = append(outputs, Gia_ManHashAndMulti2(g, lits))
	}
	Gia_ManHashStop(g)
	for _, lit := range outputs {
		Gia_ManAppendCo(g, lit)
	}
	return g
}
func Gia_MiniAigSuperDerive(name string, verbose int32) *Gia_Man_t {
	p := miniaig.Mini_AigLoad(name)
	if p == nil {
		return nil
	}
	super := Gia_MiniAigSuperGates(p)
	n := int32(miniaig.Mini_AigPiNum(p))
	multi := Gia_MiniAigSuperPreprocess(p, super, n, verbose)
	return Gia_MiniAigSuperDeriveGia(super, n, multi)
}
