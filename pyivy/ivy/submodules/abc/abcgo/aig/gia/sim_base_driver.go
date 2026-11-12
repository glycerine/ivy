package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/vec"
)

func Gia_ManPatSatImprove(p *Gia_Man_t, words0, verbose int32) {
	sims := Gia_ManSimPatSim(p)
	miter := Gia_ManSimPatGenMiter(p, sims)
	var status *Vec_Str_t
	store := Cbs2_ManSolveMiterNc(miter, 1000, &status, 0)
	Gia_ManStop(miter)
	var counts [3]int32
	for _, raw := range vec.Vec_StrArray(status) {
		value := int8(raw)
		if value < -1 || value > 1 {
			panic("gia: invalid SAT result")
		}
		counts[value+1]++
	}
	if verbose != 0 {
		fmt.Printf("Total = %d : SAT = %d.  UNSAT = %d.  UNDEC = %d.\n", counts[0]+counts[1]+counts[2], counts[1], counts[2], counts[0])
	}
	if counts[1] == 0 {
		fmt.Println("There are no counter-examples.  No need for more simulation.")
		return
	}
	p.vSimsPi = Gia_ManSimBitPacking(p, store, counts[1], counts[0])
}
func Gia_ManSimRelPrint(p *Gia_Man_t, relation *Vec_Wrd_t, outputMints *Vec_Int_t) {
	words := giaPatternWords(p)
	if words <= 0 || vec.Vec_WrdSize(relation)%int(words) != 0 {
		panic("gia: invalid relation dimensions")
	}
	mints := int32(vec.Vec_WrdSize(relation)) / words
	data := vec.Vec_WrdArray(relation)
	for pattern := int32(0); pattern < 64*words; pattern++ {
		var count int32
		for mint := int32(0); mint < mints; mint++ {
			v := giaSimBit(data, pattern*mints+mint)
			fmt.Print(v)
			count += int32(v)
		}
		fmt.Printf("  Count = %2d \n", count)
	}
}
func Gia_ManSimPatStart(n int32) *Vec_Int_t {
	result := vec.Vec_IntAlloc(int(n))
	for _, v := range []int32{17, 39, 56, 221} {
		vec.Vec_IntPush(result, v)
	}
	return result
}
func Gia_ManSimRelTest(p *Gia_Man_t) {
	objects := Gia_ManSimPatStart(4)
	values := Gia_ManSimPatValues(p)
	relation := Gia_ManSimRel(p, objects, values)
	if relation == nil {
		panic("gia: undefined simulation relation")
	}
	Gia_ManSimRelPrint(p, relation, nil)
}
