package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/vec"
	"time"
)

func Gia_Iso2ManCollectOrder2_rec(g *Gia_Man_t, id int32, order *Vec_Int_t) {
	if Gia_ObjIsTravIdCurrentId(g, id) != 0 {
		return
	}
	Gia_ObjSetTravIdCurrentId(g, id)
	obj := Gia_ManObj(g, id)
	switch {
	case Gia_ObjIsAnd(obj) != 0:
		a, b := Gia_ObjFaninId0(obj, id), Gia_ObjFaninId1(obj, id)
		if Gia_ManObj(g, a).Value > Gia_ManObj(g, b).Value {
			a, b = b, a
		}
		Gia_Iso2ManCollectOrder2_rec(g, a, order)
		Gia_Iso2ManCollectOrder2_rec(g, b, order)
	case Gia_ObjIsCo(obj) != 0:
		Gia_Iso2ManCollectOrder2_rec(g, Gia_ObjFaninId0(obj, id), order)
	case Gia_ObjIsPi(g, obj) != 0 || Gia_ObjIsConst0(obj) != 0:
	default:
		panic("gia: combinational isomorphism cone has register")
	}
	vec.Vec_IntPush(order, id)
}
func Gia_Iso2ManCollectOrder2(g *Gia_Man_t, outputs []int32, count int32) *Vec_Int_t {
	if count < 0 || int(count) > len(outputs) {
		panic("gia: invalid isomorphism output count")
	}
	order := vec.Vec_IntAlloc(1000)
	Gia_ManIncrementTravId(g)
	for _, output := range outputs[:count] {
		Gia_Iso2ManCollectOrder2_rec(g, Gia_ObjId(g, Gia_ManPo(g, output)), order)
	}
	return order
}
func Gia_Iso2ManCollectOrder_rec(g *Gia_Man_t, id int32, roots, order, mapping *Vec_Int_t) {
	if Gia_ObjIsTravIdCurrentId(g, id) != 0 {
		return
	}
	Gia_ObjSetTravIdCurrentId(g, id)
	obj := Gia_ManObj(g, id)
	switch {
	case Gia_ObjIsAnd(obj) != 0:
		a, b := Gia_ObjFaninId0(obj, id), Gia_ObjFaninId1(obj, id)
		if Gia_ManObj(g, a).Value > Gia_ManObj(g, b).Value {
			a, b = b, a
		}
		Gia_Iso2ManCollectOrder_rec(g, a, roots, order, mapping)
		Gia_Iso2ManCollectOrder_rec(g, b, roots, order, mapping)
	case Gia_ObjIsCo(obj) != 0:
		Gia_Iso2ManCollectOrder_rec(g, Gia_ObjFaninId0(obj, id), roots, order, mapping)
	case Gia_ObjIsCi(obj) != 0:
		if Gia_ObjIsRo(g, obj) != 0 {
			vec.Vec_IntPush(roots, Gia_ObjId(g, Gia_ObjRoToRi(g, obj)))
		}
	case Gia_ObjIsConst0(obj) != 0:
	default:
		panic("gia: invalid isomorphism cone object")
	}
	vec.Vec_IntWriteEntry(mapping, int(id), int32(vec.Vec_IntSize(order)))
	vec.Vec_IntPush(order, id)
}
func Gia_Iso2ManCollectOrder(g *Gia_Man_t, outputs []int32, count int32, roots, order, mapping *Vec_Int_t) {
	if count < 0 || int(count) > len(outputs) || vec.Vec_IntSize(mapping) < int(Gia_ManObjNum(g)) {
		panic("gia: invalid isomorphism order buffers")
	}
	vec.Vec_IntClear(roots)
	for _, output := range outputs[:count] {
		vec.Vec_IntPush(roots, Gia_ObjId(g, Gia_ManPo(g, output)))
	}
	vec.Vec_IntClear(order)
	Gia_ManIncrementTravId(g)
	// Roots discovered through registers are appended during traversal.
	for i := 0; i < vec.Vec_IntSize(roots); i++ {
		Gia_Iso2ManCollectOrder_rec(g, vec.Vec_IntEntry(roots, i), roots, order, mapping)
	}
}
func Gia_Iso2ManCheckIsoPair(g *Gia_Man_t, order0, order1, map0, map1 *Vec_Int_t) int32 {
	a, b := vec.Vec_IntArray(order0), vec.Vec_IntArray(order1)
	if len(a) != len(b) {
		return 0
	}
	for i, id0 := range a {
		id1 := b[i]
		if id0 == id1 {
			continue
		}
		x, y := Gia_ManObj(g, id0), Gia_ManObj(g, id1)
		if x.Value != y.Value {
			return 0
		}
		if Gia_ObjIsAnd(x) == 0 {
			continue
		}
		if Gia_ObjIsAnd(y) == 0 {
			return 0
		}
		type edge struct {
			id, phase int32
			value     uint32
		}
		edges := func(obj *Gia_Obj_t, id int32) [2]edge {
			result := [2]edge{{Gia_ObjFaninId0(obj, id), Gia_ObjFaninC0(obj), Gia_ObjFanin0(obj).Value}, {Gia_ObjFaninId1(obj, id), Gia_ObjFaninC1(obj), Gia_ObjFanin1(obj).Value}}
			if result[0].value > result[1].value {
				result[0], result[1] = result[1], result[0]
			}
			return result
		}
		xx, yy := edges(x, id0), edges(y, id1)
		for k := 0; k < 2; k++ {
			if xx[k].phase != yy[k].phase || vec.Vec_IntEntry(map0, int(xx[k].id)) != vec.Vec_IntEntry(map1, int(yy[k].id)) {
				return 0
			}
		}
	}
	return 1
}
func Gia_Iso2ManCheckIsoClassOneSkip(g *Gia_Man_t, class, roots, order0, order1, map0, map1 *Vec_Int_t) int32 {
	values := vec.Vec_IntArray(class)
	if len(values) < 2 {
		panic("gia: expected nonsingleton class")
	}
	Gia_Iso2ManCollectOrder(g, values[:1], 1, roots, order0, map0)
	for _, output := range values[1:] {
		Gia_Iso2ManCollectOrder(g, []int32{output}, 1, roots, order1, map1)
		if Gia_Iso2ManCheckIsoPair(g, order0, order1, map0, map1) == 0 {
			return 0
		}
	}
	return 1
}
func Gia_Iso2ManCheckIsoClassOne(g *Gia_Man_t, class, roots, order0, order1, map0, map1, newClass *Vec_Int_t) {
	vec.Vec_IntClear(newClass)
	values := vec.Vec_IntArray(class)
	if len(values) <= 1 {
		return
	}
	Gia_Iso2ManCollectOrder(g, values[:1], 1, roots, order0, map0)
	kept := 1
	for _, output := range values[1:] {
		Gia_Iso2ManCollectOrder(g, []int32{output}, 1, roots, order1, map1)
		if Gia_Iso2ManCheckIsoPair(g, order0, order1, map0, map1) != 0 {
			values[kept] = output
			kept++
		} else {
			vec.Vec_IntPush(newClass, output)
		}
	}
	vec.Vec_IntShrink(class, kept)
}
func Gia_Iso2ManCheckIsoClassesSkip(g *Gia_Man_t, classes *Vec_Wec_t) *Vec_Wec_t {
	result := vec.Vec_WecAlloc(2 * vec.Vec_WecSize(classes))
	roots, o0, o1 := vec.Vec_IntAlloc(1000), vec.Vec_IntAlloc(1000), vec.Vec_IntAlloc(1000)
	m0, m1 := vec.Vec_IntStart(int(Gia_ManObjNum(g))), vec.Vec_IntStart(int(Gia_ManObjNum(g)))
	for i := 0; i < vec.Vec_WecSize(classes); i++ {
		class := vec.Vec_WecEntry(classes, i)
		values := vec.Vec_IntArray(class)
		if len(values) < 2 || Gia_Iso2ManCheckIsoClassOneSkip(g, class, roots, o0, o1, m0, m1) != 0 {
			level := vec.Vec_WecPushLevel(result)
			for _, v := range values {
				vec.Vec_IntPush(level, v)
			}
		} else {
			for _, v := range values {
				vec.Vec_IntPush(vec.Vec_WecPushLevel(result), v)
			}
		}
	}
	return result
}
func Gia_Iso2ManCheckIsoClasses(g *Gia_Man_t, classes *Vec_Wec_t) *Vec_Wec_t {
	result := vec.Vec_WecAlloc(2 * vec.Vec_WecSize(classes))
	roots, o0, o1 := vec.Vec_IntAlloc(1000), vec.Vec_IntAlloc(1000), vec.Vec_IntAlloc(1000)
	m0, m1 := vec.Vec_IntStart(int(Gia_ManObjNum(g))), vec.Vec_IntStart(int(Gia_ManObjNum(g)))
	var queue [][]int32
	for i := 0; i < vec.Vec_WecSize(classes); i++ {
		queue = append(queue, append([]int32(nil), vec.Vec_IntArray(vec.Vec_WecEntry(classes, i))...))
	}
	for i := 0; i < len(queue); i++ {
		class := vec.Vec_IntAllocArrayCopy(queue[i], len(queue[i]))
		next := vec.Vec_IntAlloc(0)
		Gia_Iso2ManCheckIsoClassOne(g, class, roots, o0, o1, m0, m1, next)
		level := vec.Vec_WecPushLevel(result)
		for _, v := range vec.Vec_IntArray(class) {
			vec.Vec_IntPush(level, v)
		}
		if vec.Vec_IntSize(next) > 0 {
			queue = append(queue, append([]int32(nil), vec.Vec_IntArray(next)...))
		}
	}
	return result
}

// Gia_ManIsoReduce2 refines structural signature classes by cone comparison.
// As in the source algorithm, the dual flag does not alter this variant.
func Gia_ManIsoReduce2(g *Gia_Man_t, outputGroups, inputPerms **Vec_Ptr_t, estimate, betterQuality, dual, verbose, veryVerbose int32) *Gia_Man_t {
	start := time.Now()
	classes := Gia_Iso2ManPerform(g, veryVerbose)
	var used int
	nontrivial := vec.Vec_WecCountNonTrivial(classes, &used)
	fmt.Printf("Reduced %d outputs to %d candidate classes (%d outputs are in %d non-trivial classes).  Time = %.2f sec\n", Gia_ManPoNum(g), vec.Vec_WecSize(classes), used, nontrivial, time.Since(start).Seconds())
	if estimate != 0 {
		return Gia_ManDup(g)
	}
	var refined *Vec_Wec_t
	if betterQuality != 0 {
		refined = Gia_Iso2ManCheckIsoClasses(g, classes)
	} else {
		refined = Gia_Iso2ManCheckIsoClassesSkip(g, classes)
	}
	vec.Vec_WecSortByFirstInt(refined, 0)
	keep := vec.Vec_WecCollectFirsts(refined)
	result := Gia_ManDupCones(g, vec.Vec_IntArray(keep), int32(vec.Vec_IntSize(keep)), 0)
	nontrivial = vec.Vec_WecCountNonTrivial(refined, &used)
	fmt.Printf("Reduced %d outputs to %d equivalence classes (%d outputs are in %d non-trivial classes).  Time = %.2f sec\n", Gia_ManPoNum(g), vec.Vec_WecSize(refined), used, nontrivial, time.Since(start).Seconds())
	if verbose != 0 {
		for i := 0; i < vec.Vec_WecSize(refined); i++ {
			group := vec.Vec_IntArray(vec.Vec_WecEntry(refined, i))
			if len(group) > 1 {
				fmt.Println(group)
			}
		}
	}
	if inputPerms != nil {
		*inputPerms = nil
	}
	if outputGroups != nil {
		*outputGroups = vec.Vec_WecConvertToVecPtr(refined)
	}
	return result
}
