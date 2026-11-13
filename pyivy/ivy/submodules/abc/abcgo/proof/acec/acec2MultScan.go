package acec

// Cut pattern detection from src/proof/acec/acec2Mult.c.
import (
	"fmt"
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
)

func Sdb_StoDiffExactlyOne(cuts *Vec_Wec_t, limit int32, cut []int32) int32 {
	for i := 0; i < vec.Vec_WecSize(cuts); i++ {
		level := vec.Vec_WecEntry(cuts, i)
		found := true
		for _, id := range cut[1 : 1+cut[0]] {
			if vec.Vec_IntFind(level, id) < 0 {
				found = false
				break
			}
		}
		if found {
			return -1
		}
	}
	for i := 0; i < vec.Vec_WecSize(cuts) && int32(i) != limit; i++ {
		id := Sdb_StoDiffExactlyOne2(vec.Vec_WecEntry(cuts, i), cut)
		if id != -1 {
			return id
		}
	}
	return -1
}
func Sdb_StoDiffExactlyOne3(cuts *Vec_Wec_t, limit int32, cut []int32, count *int32) int32 {
	all := int32(-1)
	*count = 0
	for i := 0; i < vec.Vec_WecSize(cuts) && int32(i) != limit; i++ {
		id := Sdb_StoDiffExactlyOne2(vec.Vec_WecEntry(cuts, i), cut)
		if id != -1 {
			if all == -1 {
				all = id
			}
			if all == id {
				*count++
			}
		}
	}
	return all
}
func Sdb_StoFindAll(cuts *Vec_Wec_t) *Vec_Int_t {
	all := vec.Vec_IntAlloc(100)
	for i := 0; i < vec.Vec_WecSize(cuts); i++ {
		for _, id := range vec.Vec_IntArray(vec.Vec_WecEntry(cuts, i)) {
			vec.Vec_IntPushUnique(all, id)
		}
	}
	return all
}
func Sdb_StoDiffExactlyOne2(all *Vec_Int_t, cut []int32) int32 {
	missing := int32(-1)
	for _, id := range cut[1 : 1+cut[0]] {
		if vec.Vec_IntFind(all, id) >= 0 {
			continue
		}
		if missing != -1 {
			return -1
		}
		missing = id
	}
	return missing
}
func Sdb_StoFindInputs(cuts *Vec_Wec_t, front int32) *Vec_Int_t {
	res := vec.Vec_IntAlloc(100)
	first := vec.Vec_IntAlloc(100)
	second := vec.Vec_IntAlloc(100)
	max := int32(0)
	for i := 0; i < vec.Vec_WecSize(cuts); i++ {
		for _, id := range vec.Vec_IntArray(vec.Vec_WecEntry(cuts, i)) {
			if id > max {
				max = id
			}
		}
	}
	if front < 0 || front > max {
		return res
	}
	counts := vec.Vec_IntStart(int(max + 1))
	for i := 0; i < vec.Vec_WecSize(cuts); i++ {
		for _, id := range vec.Vec_IntArray(vec.Vec_WecEntry(cuts, i)) {
			vec.Vec_IntAddToEntry(counts, int(id), 1)
		}
	}
	vec.Vec_IntWriteEntry(counts, 0, 0)
	min := int32(0)
	for _, n := range vec.Vec_IntArray(counts) {
		if n != 0 && (min == 0 || n < min) {
			min = n
		}
	}
	if min == 0 {
		return res
	}
	index := vec.Vec_IntFind(counts, min)
	vec.Vec_IntPush(first, int32(index))
	vec.Vec_IntWriteEntry(counts, index, 0)
	for i := 0; i < vec.Vec_IntSize(counts); i++ {
		n := vec.Vec_IntEntry(counts, i)
		if n == min || n == 2*min {
			vec.Vec_IntPush(first, int32(i))
			vec.Vec_IntWriteEntry(counts, i, 0)
		}
	}
	min = vec.Vec_IntEntry(counts, int(front))
	if min == 0 {
		return res
	}
	vec.Vec_IntPush(second, front)
	vec.Vec_IntWriteEntry(counts, int(front), 0)
	for i := 0; i < vec.Vec_IntSize(counts); i++ {
		n := vec.Vec_IntEntry(counts, i)
		if n == min || n == 2*min {
			vec.Vec_IntPush(second, int32(i))
			vec.Vec_IntWriteEntry(counts, i, 0)
		}
	}
	vec.Vec_IntPrint(first)
	vec.Vec_IntPrint(second)
	vec.Vec_IntAppend(res, first)
	vec.Vec_IntAppend(res, second)
	return res
}
func Sdb_StoIterCutsOne(p *Sdb_Sto_t, id, size int32, cut *[]int32) int32 {
	list := vec.Vec_IntArray(vec.Vec_WecEntry(p.vCuts, int(id)))
	off := 1
	for i := int32(0); i < list[0]; i++ {
		n := list[off]
		entry := list[off : off+int(n)+2]
		off += int(n) + 2
		if n != size {
			continue
		}
		truth := uint32(vec.Vec_MemReadEntry(p.vTtMem, int(entry[n+1]>>1))[0])
		if n == 5 {
			for _, v := range s_FuncTruths5 {
				if v == truth {
					*cut = entry
					return 1
				}
			}
		}
		if n == 4 {
			truth &= 0xffff
			for _, v := range s_FuncTruths4 {
				if v == truth {
					*cut = entry
					return 2
				}
			}
			for _, v := range s_FuncTruths4a {
				if v == truth {
					*cut = entry
					return 3
				}
			}
			for _, v := range s_FuncTruths4b {
				if v == truth {
					*cut = entry
					return 4
				}
			}
		}
	}
	return 0
}
func Sdb_StoIterCuts(p *Sdb_Sto_t) *Vec_Int_t {
	cuts := vec.Vec_WecAlloc(100)
	var cut []int32
	for id := int32(1); id < gia.Gia_ManObjNum(p.pGia); id++ {
		if gia.Gia_ObjIsAnd(gia.Gia_ManObj(p.pGia, id)) == 0 {
			continue
		}
		r := Sdb_StoIterCutsOne(p, id, 5, &cut)
		if r == 0 {
			continue
		}
		if r != 1 || cut[0] != 5 {
			panic("acec: expected 5-cut")
		}
		level := vec.Vec_WecPushLevel(cuts)
		for _, v := range cut[1:6] {
			vec.Vec_IntPush(level, v)
		}
	}
	limit := int32(vec.Vec_WecSize(cuts))
	fmt.Printf("Detected %d  5-cuts.\n", limit)
	_ = Sdb_StoFindAll(cuts)
	front, back, signed := int32(-1), int32(-1), int32(0)
	for id := int32(1); id < gia.Gia_ManObjNum(p.pGia); id++ {
		if gia.Gia_ObjIsAnd(gia.Gia_ManObj(p.pGia, id)) == 0 {
			continue
		}
		r := Sdb_StoIterCutsOne(p, id, 4, &cut)
		if r == 0 {
			continue
		}
		if r < 2 || r > 4 || cut[0] != 4 {
			panic("acec: expected 4-cut")
		}
		n := Sdb_StoDiffExactlyOne(cuts, limit, cut)
		if n == -1 {
			continue
		}
		if r == 2 {
			front = n
		} else {
			back = n
		}
		if r == 4 {
			signed = 1
		}
		level := vec.Vec_WecPushLevel(cuts)
		vec.Vec_IntPush(level, 0)
		for _, v := range cut[1:5] {
			vec.Vec_IntPush(level, v)
		}
	}
	vec.Vec_WecPrint(cuts, 0)
	if front != 0 {
		fmt.Printf("Front  = %d\n", front)
	}
	if back != 0 {
		fmt.Printf("Back   = %d\n", back)
	}
	if signed != 0 {
		fmt.Printf("Sign   = %d\n", signed)
	}
	return Sdb_StoFindInputs(cuts, front)
}
