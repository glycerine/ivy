package gia

import (
	"sort"

	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

// Internal support lists are immutable. Unary nodes and empty unions can share
// them; Gia_ManCreateNodeSupps copies entries into independently owned results.
func giaMergeSupports(a, b []int32) []int32 {
	if len(a) == 0 {
		return b
	}
	if len(b) == 0 {
		return a
	}
	result := make([]int32, 0, len(a)+len(b))
	i, j := 0, 0
	for i < len(a) && j < len(b) {
		switch {
		case a[i] < b[j]:
			result = append(result, a[i])
			i++
		case b[j] < a[i]:
			result = append(result, b[j])
			j++
		default:
			result = append(result, a[i])
			i++
			j++
		}
	}
	return append(append(result, a[i:]...), b[j:]...)
}

func giaAllSupports(p *Gia_Man_t) [][]int32 {
	supports := make([][]int32, p.nObjs)
	for i := int32(0); i < Gia_ManCiNum(p); i++ {
		supports[Gia_ManCi(p, i).id] = []int32{i}
	}
	for _, obj := range p.pObjs {
		if Gia_ObjIsCi(obj) != 0 || obj.id == 0 {
			continue
		}
		a, b, c, kind := p.Gate(obj.id)
		result := supports[a>>1]
		if kind == 2 || kind == 3 || kind == 4 {
			result = giaMergeSupports(result, supports[b>>1])
		}
		if kind == 4 {
			result = giaMergeSupports(result, supports[c>>1])
		}
		supports[obj.id] = result
	}
	return supports
}
func Gia_ManCreateNodeSupps(p *Gia_Man_t, nodes *Vec_Int_t, verbose int32) *Vec_Wec_t {
	clk := util.Abc_Clock()
	all := giaAllSupports(p)
	result := vec.Vec_WecStart(vec.Vec_IntSize(nodes))
	for i, id := range vec.Vec_IntArray(nodes) {
		if id < 0 || id >= p.nObjs {
			panic("gia: invalid support node")
		}
		for _, ci := range all[id] {
			vec.Vec_IntPush(vec.Vec_WecEntry(result, i), ci)
		}
	}
	if verbose != 0 {
		util.Abc_PrintTime(1, "Support computation", util.Abc_Clock()-clk)
	}
	return result
}
func Gia_ManCreateCoSupps(p *Gia_Man_t, verbose int32) *Vec_Wec_t {
	nodes := vec.Vec_IntAlloc(int(Gia_ManCoNum(p)))
	for i := int32(0); i < Gia_ManCoNum(p); i++ {
		vec.Vec_IntPush(nodes, Gia_ManCo(p, i).id)
	}
	return Gia_ManCreateNodeSupps(p, nodes, verbose)
}
func Gia_ManCoSuppSizeMax(p *Gia_Man_t, supports *Vec_Wec_t) int32 {
	if vec.Vec_WecSize(supports) != int(Gia_ManCoNum(p)) {
		panic("gia: invalid output supports")
	}
	maximum := int32(1)
	for i := 0; i < vec.Vec_WecSize(supports); i++ {
		maximum = max(maximum, int32(vec.Vec_IntSize(vec.Vec_WecEntry(supports, i))))
	}
	return maximum
}
func Gia_ManCoLargestSupp(p *Gia_Man_t, supports *Vec_Wec_t) int32 {
	if vec.Vec_WecSize(supports) != int(Gia_ManCoNum(p)) {
		panic("gia: invalid output supports")
	}
	index := int32(-1)
	size := -1
	for i := 0; i < vec.Vec_WecSize(supports); i++ {
		n := vec.Vec_IntSize(vec.Vec_WecEntry(supports, i))
		if n > size {
			index = int32(i)
			size = n
		}
	}
	return index
}
func Gia_ManSortCoBySuppSize(p *Gia_Man_t, supports *Vec_Wec_t) *Vec_Int_t {
	if vec.Vec_WecSize(supports) != int(Gia_ManCoNum(p)) {
		panic("gia: invalid output supports")
	}
	order := make([]int32, vec.Vec_WecSize(supports))
	for i := range order {
		order[i] = int32(i)
	}
	sort.SliceStable(order, func(i, j int) bool {
		return vec.Vec_IntSize(vec.Vec_WecEntry(supports, int(order[i]))) > vec.Vec_IntSize(vec.Vec_WecEntry(supports, int(order[j])))
	})
	result := vec.Vec_IntAlloc(len(order))
	for _, id := range order {
		vec.Vec_IntPush(result, id)
	}
	return result
}
