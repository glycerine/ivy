package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

func Unm_ManPrintPairStats(hash *vec.Hash_IntMan_t, total, initial, final, _ int32) int32 {
	var counts [21]int32
	var pairs, refs int32
	for i := int32(1); i <= int32(vec.Hash_IntManEntryNum(hash)); i++ {
		n := min(int32(20), vec.Hash_IntObjData2(hash, i))
		counts[n]++
		refs += n
		if n > 1 {
			pairs++
		}
	}
	fmt.Println("Statistics for pairs appearing less than 20 times:")
	for i, n := range counts {
		if n > 0 {
			fmt.Printf("%3d : %7d %7.2f %%\n", i, n, 100*float64(n)*float64(i)/float64(max(refs, 1)))
		}
	}
	fmt.Printf("Pairs: Total = %d Init = %d Final = %d Real = %d\n", total, initial, pairs, final)
	return pairs
}
func Unm_ManComputePairs(p *Unm_Man_t, verbose int32) *Vec_Int_t {
	g := p.pGia
	Gia_ManSetRefsMapped(g)
	hash := vec.Hash_IntManStart(1000)
	defer vec.Hash_IntManStop(hash)
	first := vec.Vec_IntStart(1)
	result := vec.Vec_IntAlloc(1000)
	var total, initial int32
	for id := int32(1); id < g.nObjs; id++ {
		if Gia_ObjIsLut(g, id) == 0 {
			continue
		}
		fans := Gia_ObjLutFanins(g, id)
		total += int32(len(fans) * (len(fans) - 1) / 2)
		selected := make([]int32, 0, len(fans))
		for _, fan := range fans {
			if Gia_ObjRefNumId(g, fan) > 1 {
				selected = append(selected, fan)
			}
		}
		initial += int32(len(selected) * (len(selected) - 1) / 2)
		for i := range selected {
			for j := i + 1; j < len(selected); j++ {
				a, b := selected[i], selected[j]
				if a > b {
					a, b = b, a
				}
				num := vec.Hash_Int2ManInsert(hash, a, b, 0)
				refs := vec.Hash_Int2ObjInc(hash, num)
				if refs == 0 {
					vec.Vec_IntPush(first, id)
					continue
				}
				if refs == 1 {
					vec.Vec_IntPushThree(result, vec.Vec_IntEntry(first, int(num)), a, b)
				}
				vec.Vec_IntPushThree(result, id, a, b)
			}
		}
	}
	if verbose != 0 {
		Unm_ManPrintPairStats(hash, total, initial, int32(vec.Vec_IntSize(result)/3), 0)
	}
	return result
}
func Unm_ManCollectDecomp(p *Unm_Man_t, pairs *Vec_Int_t, verbose int32) *Vec_Int_t {
	if vec.Hash_IntManEntryNum(p.pHash) != 0 {
		panic("gia: unmapping hash is not empty")
	}
	first := vec.Vec_IntStart(1)
	result := vec.Vec_IntAlloc(1000)
	data := vec.Vec_IntArray(pairs)
	if len(data)%3 != 0 {
		panic("gia: invalid unmapping pairs")
	}
	for i := 0; i < len(data); i += 3 {
		id, a, b := data[i], data[i+1], data[i+2]
		vec.Vec_IntClear(p.vLeaves)
		for _, fan := range Gia_ObjLutFanins(p.pGia, id) {
			vec.Vec_IntPush(p.vLeaves, fan)
		}
		x, y := vec.Vec_IntFind(p.vLeaves, a), vec.Vec_IntFind(p.vLeaves, b)
		if x < 0 || y < 0 || a >= b {
			panic("gia: invalid unmapping pair")
		}
		used := vec.Vec_IntEntry(p.vId2Used, int(id))
		truth := vec.Vec_WrdEntry(p.vTruths, int(used))
		phase := util.Abc_TtCheckDsdAnd(truth, int32(x), int32(y), nil)
		if phase < 0 {
			continue
		}
		la, lb := 2*a+(phase&1), 2*b+((phase>>1)&1)
		if phase == 4 {
			la, lb = lb, la
		}
		num := vec.Hash_Int2ManInsert(p.pHash, la, lb, 0)
		refs := vec.Hash_Int2ObjInc(p.pHash, num)
		if refs == 0 {
			vec.Vec_IntPush(first, id)
			continue
		}
		if refs == 1 {
			vec.Vec_IntPushThree(result, vec.Vec_IntEntry(first, int(num)), la, lb)
		}
		vec.Vec_IntPushThree(result, id, la, lb)
	}
	if verbose != 0 {
		Unm_ManPrintPairStats(p.pHash, int32(len(data)/3), int32(vec.Hash_IntManEntryNum(p.pHash)), int32(vec.Vec_IntSize(result)/3), 1)
	}
	return result
}
