package gia

import "github.com/glycerine/abcgo/misc/vec"

func Gia_WinTryAddingNode(p *Gia_Man_t, pivot, second int32, levels *Vec_Wec_t, nodes *Vec_Int_t) int32 {
	if vec.Vec_WecSizeSize(levels) != 0 || Gia_ObjIsTravIdCurrentId(p, pivot) != 0 {
		panic("gia: invalid Rsb2 candidate window")
	}
	add := func(id int32) {
		if Gia_ObjIsTravIdCurrentId(p, id) != 0 {
			panic("gia: duplicate Rsb2 candidate")
		}
		Gia_ObjSetTravIdCurrentId(p, id)
		vec.Vec_WecPush(levels, int(Gia_ObjLevelId(p, id)), id)
	}
	add(pivot)
	if second != -1 {
		add(second)
	}
	count := int32(0)
	for level := 0; level < vec.Vec_WecSize(levels); level++ {
		for i := 0; i < vec.Vec_IntSize(vec.Vec_WecEntry(levels, level)); i++ {
			id := vec.Vec_IntEntry(vec.Vec_WecEntry(levels, level), i)
			for f := int32(0); f < min(5, Gia_ObjFanoutNumId(p, id)); f++ {
				fan := Gia_ObjFanoutId(p, id, f)
				obj := Gia_ManObj(p, fan)
				if Gia_ObjIsAnd(obj) == 0 || Gia_ObjIsTravIdCurrentId(p, fan) != 0 {
					continue
				}
				ready := true
				for _, dep := range giaObjectFanins(p, obj) {
					if Gia_ObjIsTravIdCurrentId(p, dep) == 0 {
						ready = false
						break
					}
				}
				if ready {
					add(fan)
					count++
				}
			}
		}
	}
	for level := 0; level < vec.Vec_WecSize(levels); level++ {
		v := vec.Vec_WecEntry(levels, level)
		for _, id := range vec.Vec_IntArray(v) {
			if nodes == nil {
				Gia_ObjSetTravIdPreviousId(p, id)
			} else {
				vec.Vec_IntPush(nodes, id)
			}
		}
		vec.Vec_IntClear(v)
	}
	return count
}
func Gia_WinAddCiWithMaxFanouts(p *Gia_Man_t) int32 {
	best, count := int32(-1), int32(-1)
	for _, id := range vec.Vec_IntArray(p.vCis) {
		if n := Gia_ObjFanoutNumId(p, id); n > count {
			best, count = id, n
		}
	}
	if best < 0 {
		panic("gia: Rsb2 graph has no inputs")
	}
	return best
}
func Gia_WinAddCiWithMaxDivisors(p *Gia_Man_t, levels *Vec_Wec_t) int32 {
	best, count := int32(-1), int32(-1)
	for _, id := range vec.Vec_IntArray(p.vCis) {
		if Gia_ObjIsTravIdCurrentId(p, id) != 0 {
			continue
		}
		n := Gia_WinTryAddingNode(p, id, -1, levels, nil)
		if n > count {
			best, count = id, n
		}
	}
	if best < 0 {
		panic("gia: no remaining Rsb2 input")
	}
	return best
}
func Gia_WinNodeHasUnmarkedFanouts(p *Gia_Man_t, id int32) int32 {
	for i := int32(0); i < Gia_ObjFanoutNumId(p, id); i++ {
		if Gia_ObjIsTravIdCurrentId(p, Gia_ObjFanoutId(p, id, i)) == 0 {
			return 1
		}
	}
	return 0
}
func Gia_RsbCiTranslate(p *Gia_Man_t, objects, mapping *Vec_Int_t) *Vec_Int_t {
	if vec.Vec_IntSize(mapping) != int(p.nObjs) {
		panic("gia: invalid Rsb2 window map")
	}
	result := vec.Vec_IntAlloc(100)
	vec.Vec_IntPush(result, 0)
	vec.Vec_IntPush(result, 0)
	vec.Vec_IntWriteEntry(mapping, 0, 0)
	for _, id := range vec.Vec_IntArray(objects) {
		obj := Gia_ManObj(p, id)
		if Gia_ObjIsTravIdCurrentId(p, id) == 0 {
			panic("gia: unmarked Rsb2 window object")
		}
		a, b := int32(0), int32(0)
		if Gia_ObjIsCi(obj) == 0 {
			if Gia_ObjIsAnd(obj) == 0 || Gia_ObjIsMuxId(p, id) != 0 || Gia_ObjIsXor(obj) != 0 {
				panic("gia: unsupported Rsb2 CI-window gate")
			}
			f0, f1 := Gia_ObjFaninId0(obj, id), Gia_ObjFaninId1(obj, id)
			a = vec.Vec_IntEntry(mapping, int(f0))
			b = vec.Vec_IntEntry(mapping, int(f1))
			if a < 0 || b < 0 {
				panic("gia: missing Rsb2 window fanin")
			}
			a ^= Gia_ObjFaninC0(obj)
			b ^= Gia_ObjFaninC1(obj)
			if a > b {
				a, b = b, a
			}
		}
		vec.Vec_IntWriteEntry(mapping, int(id), int32(vec.Vec_IntSize(result)))
		vec.Vec_IntPush(result, a)
		vec.Vec_IntPush(result, b)
	}
	for _, id := range vec.Vec_IntArray(objects) {
		if Gia_WinNodeHasUnmarkedFanouts(p, id) != 0 {
			lit := vec.Vec_IntEntry(mapping, int(id))
			vec.Vec_IntPush(result, lit)
			vec.Vec_IntPush(result, lit)
		}
	}
	return result
}
func Gia_RsbCiWindow(p *Gia_Man_t, nPis int32) *Vec_Int_t {
	if nPis <= 0 || nPis > Gia_ManCiNum(p) {
		panic("gia: invalid Rsb2 input count")
	}
	nodes := vec.Vec_IntAlloc(100)
	mapping := vec.Vec_IntStartFull(int(p.nObjs))
	levels := vec.Vec_WecStart(int(Gia_ManLevelNum(p) + 1))
	Gia_ManStaticFanoutStart(p)
	defer Gia_ManStaticFanoutStop(p)
	Gia_ManIncrementTravId(p)
	first := Gia_WinAddCiWithMaxFanouts(p)
	Gia_ObjSetTravIdCurrentId(p, first)
	vec.Vec_IntPush(nodes, first)
	for i := int32(1); i < nPis; i++ {
		next := Gia_WinAddCiWithMaxDivisors(p, levels)
		Gia_WinTryAddingNode(p, next, -1, levels, nodes)
	}
	vec.Vec_IntSort(nodes, 0)
	return Gia_RsbCiTranslate(p, nodes, mapping)
}
func Gia_RsbCiWindowTest(p *Gia_Man_t) { Gia_RsbCiWindow(p, 6) }
