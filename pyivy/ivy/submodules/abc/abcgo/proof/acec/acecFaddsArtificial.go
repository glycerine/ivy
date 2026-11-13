package acec

// Artificial carry-chain annotation and box construction from acecFadds.c.
import (
	"fmt"
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/tim"
	"github.com/glycerine/abcgo/misc/vec"
)

func Gia_ObjFanin0CopyCarry(carries *vec.Vec_Int_t, obj *gia.Gia_Obj_t, id int32) int32 {
	v := gia.Gia_ObjFaninId0(obj, id)
	if carries == nil || vec.Vec_IntEntry(carries, int(v)) == -1 {
		return gia.Gia_ObjFanin0Copy(obj)
	}
	return vec.Vec_IntEntry(carries, int(v)) ^ gia.Gia_ObjFaninC0(obj)
}
func Gia_ObjFanin1CopyCarry(carries *vec.Vec_Int_t, obj *gia.Gia_Obj_t, id int32) int32 {
	v := gia.Gia_ObjFaninId1(obj, id)
	if carries == nil || vec.Vec_IntEntry(carries, int(v)) == -1 {
		return gia.Gia_ObjFanin1Copy(obj)
	}
	return vec.Vec_IntEntry(carries, int(v)) ^ gia.Gia_ObjFaninC1(obj)
}
func Gia_ManDupWithArtificalFaddBoxes(graph *gia.Gia_Man_t, useFanout, xorTrick int32) *gia.Gia_Man_t {
	boxes := gia.Gia_ManBoxNum(graph)
	gia.Gia_ManCleanPhase(graph)
	for id := int32(1); id < gia.Gia_ManObjNum(graph); id++ {
		obj := gia.Gia_ManObj(graph, id)
		m0, m1 := gia.Gia_ObjMark0(obj), gia.Gia_ObjMark1(obj)
		if gia.Gia_ObjIsCi(obj) != 0 || gia.Gia_ObjIsCo(obj) != 0 {
			if m0 != 0 || m1 != 0 {
				panic("box marks on terminal")
			}
			continue
		}
		if m0 != 0 && m1 != 0 {
			panic("overlapping carry annotations")
		}
		if m0 != 0 || m1 != 0 {
			fan := gia.Gia_ObjFanin0(obj)
			if m1 != 0 {
				fan = gia.Gia_ObjFanin1(obj)
			}
			if gia.Gia_ObjPhase(fan) != 0 {
				panic("overlapping carry chains")
			}
			gia.Gia_ObjSetPhaseValue(fan, 1)
		}
	}
	var carries *vec.Vec_Int_t
	if useFanout == 0 {
		carries = vec.Vec_IntStartFull(int(gia.Gia_ManObjNum(graph)))
		if graph.Refs != nil {
			panic("reference counts already exist")
		}
		gia.Gia_ManCreateRefs(graph)
		for id := int32(1); id < gia.Gia_ManObjNum(graph); id++ {
			obj := gia.Gia_ManObj(graph, id)
			if gia.Gia_ObjIsAnd(obj) == 0 {
				continue
			}
			if gia.Gia_ObjMark0(obj) != 0 {
				gia.Gia_ObjRefFanin0Dec(graph, obj)
			} else if gia.Gia_ObjMark1(obj) != 0 {
				gia.Gia_ObjRefFanin1Dec(graph, obj)
			}
		}
	}
	gia.Gia_ManFillValue(graph)
	dst := gia.Gia_ManStart(gia.Gia_ManObjNum(graph))
	dst.CopyIdentityFrom(graph)
	gia.Gia_ObjSetValue(gia.Gia_ManConst0(graph), 0)
	for id := int32(1); id < gia.Gia_ManObjNum(graph); id++ {
		obj := gia.Gia_ManObj(graph, id)
		switch {
		case gia.Gia_ObjIsCi(obj) != 0:
			gia.Gia_ObjSetValue(obj, gia.Gia_ManAppendCi(dst))
		case gia.Gia_ObjIsCo(obj) != 0:
			gia.Gia_ObjSetValue(obj, gia.Gia_ManAppendCo(dst, gia.Gia_ObjFanin0Copy(obj)))
		case gia.Gia_ObjMark0(obj) == 0 && gia.Gia_ObjMark1(obj) == 0:
			gia.Gia_ObjSetValue(obj, gia.Gia_ManAppendAnd(dst, gia.Gia_ObjFanin0Copy(obj), gia.Gia_ObjFanin1Copy(obj)))
		default:
			var carry, other int32
			if gia.Gia_ObjMark0(obj) != 0 {
				carry = Gia_ObjFanin0CopyCarry(carries, obj, id)
				other = gia.Gia_ObjFanin1Copy(obj)
			} else {
				carry = Gia_ObjFanin1CopyCarry(carries, obj, id)
				other = gia.Gia_ObjFanin0Copy(obj)
			}
			if carry < 0 || other < 0 {
				panic("unmapped carry input")
			}
			phase := carry & 1
			gia.Gia_ManAppendCo(dst, carry^phase)
			gia.Gia_ManAppendCo(dst, other^phase)
			gia.Gia_ManAppendCo(dst, phase)
			sum := gia.Gia_ManAppendCi(dst)
			value := gia.Gia_ManAppendCi(dst) ^ phase
			gia.Gia_ObjSetValue(obj, value)
			if carries != nil && gia.Gia_ObjPhase(obj) != 0 {
				vec.Vec_IntWriteEntry(carries, int(id), value)
				if gia.Gia_ObjRefNum(graph, obj) > 0 {
					if xorTrick != 0 {
						value = gia.Gia_ManAppendAnd(dst, sum^(phase^1), other)
					} else {
						value = gia.Gia_ManAppendAnd(dst, gia.Gia_ObjFanin0Copy(obj), gia.Gia_ObjFanin1Copy(obj))
					}
					gia.Gia_ObjSetValue(obj, value)
				}
			}
			boxes++
		}
	}
	gia.Gia_ManCleanPhase(graph)
	graph.Refs = nil
	if gia.Gia_ManHasDangling(dst) != 0 {
		panic("artificial box graph has dangling logic")
	}
	realPI, realPO := gia.Gia_ManCiNum(graph), gia.Gia_ManCoNum(graph)
	if gia.Gia_ManBoxNum(graph) > 0 {
		manager, _ := graph.BoxTiming()
		realPI = int32(tim.Tim_ManPiNum(manager))
		realPO = int32(tim.Tim_ManPoNum(manager))
	}
	dst.SetBoxTiming(Gia_ManGenerateTim(realPI, realPO, boxes, 3, 2), Gia_ManGenerateExtraAig(boxes, 3, 2))
	return dst
}
func Gia_ManDupWithArtificalFaddBoxesTest(graph *gia.Gia_Man_t) *gia.Gia_Man_t {
	gia.Gia_ManCleanMark01(graph)
	for id := int32(1); id < gia.Gia_ManObjNum(graph); id++ {
		obj := gia.Gia_ManObj(graph, id)
		if gia.Gia_ObjIsAnd(obj) == 0 {
			continue
		}
		a, b := id%5, id%7
		if a != 0 && b != 0 {
			a, b = 0, 0
		}
		gia.Gia_ObjSetMark0(obj, a)
		gia.Gia_ObjSetMark1(obj, b)
	}
	result := Gia_ManDupWithArtificalFaddBoxes(graph, 0, 0)
	gia.Gia_ManCleanMark01(graph)
	return result
}
func Gia_ManFindAnnotatedDelay(graph *gia.Gia_Man_t, carryDelay int32, nBoxes *int32, ignoreBoxes int32) int32 {
	gia.Gia_ManCleanLevels(graph, gia.Gia_ManObjNum(graph))
	realPI := gia.Gia_ManCiNum(graph)
	if gia.Gia_ManBoxNum(graph) > 0 {
		manager, _ := graph.BoxTiming()
		realPI = int32(tim.Tim_ManPiNum(manager))
	}
	maximum, count := int32(0), int32(0)
	for id := int32(1); id < gia.Gia_ManObjNum(graph); id++ {
		obj := gia.Gia_ManObj(graph, id)
		if gia.Gia_ObjIsCi(obj) != 0 {
			if ignoreBoxes != 0 {
				continue
			}
			index := gia.Gia_ObjCioId(obj) - realPI
			if index < 0 {
				continue
			}
			box := index / 2
			if box >= gia.Gia_ManBoxNum(graph) {
				panic("invalid box output")
			}
			delay := int32(0)
			for k := int32(0); k < 3; k++ {
				co := gia.Gia_ManCo(graph, box*3+k)
				coID := gia.Gia_ObjId(graph, co)
				if coID >= id {
					panic("box is not topological")
				}
				delay = max(delay, gia.Gia_ObjLevelId(graph, coID))
			}
			if index&1 != 0 {
				delay += carryDelay
			} else {
				delay += 100
			}
			gia.Gia_ObjSetLevelId(graph, id, delay)
			continue
		}
		if gia.Gia_ObjIsCo(obj) != 0 {
			delay := gia.Gia_ObjLevelId(graph, gia.Gia_ObjFaninId0(obj, id))
			gia.Gia_ObjSetLevelId(graph, id, delay)
			maximum = max(maximum, delay)
			continue
		}
		a, b := gia.Gia_ObjLevelId(graph, gia.Gia_ObjFaninId0(obj, id)), gia.Gia_ObjLevelId(graph, gia.Gia_ObjFaninId1(obj, id))
		m0, m1 := gia.Gia_ObjMark0(obj), gia.Gia_ObjMark1(obj)
		if m0 != 0 && m1 != 0 {
			panic("overlapping annotations")
		}
		var delay int32
		if m0 != 0 {
			delay = max(a+carryDelay, b+100)
			count++
		} else if m1 != 0 {
			delay = max(b+carryDelay, a+100)
			count++
		} else {
			delay = max(a+100, b+100)
		}
		gia.Gia_ObjSetLevelId(graph, id, delay)
	}
	if nBoxes != nil {
		*nBoxes = count
	}
	return maximum
}
func giaArtificialUsed(obj *gia.Gia_Obj_t) bool {
	return gia.Gia_ObjMark0(obj) != 0 || gia.Gia_ObjMark1(obj) != 0 || gia.Gia_ObjPhase(obj) != 0
}
func Gia_ManFindChainStart(graph *gia.Gia_Man_t) int32 {
	best, delay := int32(-1), int32(0)
	for id := int32(1); id < gia.Gia_ManObjNum(graph); id++ {
		obj := gia.Gia_ManObj(graph, id)
		if gia.Gia_ObjIsAnd(obj) == 0 || giaArtificialUsed(obj) || delay > gia.Gia_ObjLevelId(graph, id) {
			continue
		}
		best = id
		delay = gia.Gia_ObjLevelId(graph, id)
	}
	return best
}
func Gia_ManFindPath(graph *gia.Gia_Man_t, carryDelay, minPath, maxPath int32, path *vec.Vec_Int_t) int32 {
	id := Gia_ManFindChainStart(graph)
	if id < 0 {
		return -1
	}
	start := id
	vec.Vec_IntClear(path)
	for gia.Gia_ObjIsAnd(gia.Gia_ManObj(graph, id)) != 0 {
		obj := gia.Gia_ManObj(graph, id)
		if giaArtificialUsed(obj) {
			panic("carry path touches used node")
		}
		a, b := gia.Gia_ObjFaninId0(obj, id), gia.Gia_ObjFaninId1(obj, id)
		usedA, usedB := giaArtificialUsed(gia.Gia_ManObj(graph, a)), giaArtificialUsed(gia.Gia_ManObj(graph, b))
		if usedA && usedB {
			break
		}
		choose := int32(0)
		if usedA || !usedB && gia.Gia_ObjLevelId(graph, b) > gia.Gia_ObjLevelId(graph, a) {
			choose = 1
		}
		vec.Vec_IntPush(path, 2*id+choose)
		if choose == 0 {
			id = a
		} else {
			id = b
		}
	}
	if vec.Vec_IntSize(path) < int(minPath) {
		gia.Gia_ObjSetPhaseValue(gia.Gia_ManObj(graph, start), 1)
		return 0
	}
	if vec.Vec_IntSize(path) > int(maxPath) {
		vec.Vec_IntShrink(path, int(maxPath))
	}
	for _, lit := range vec.Vec_IntArray(path) {
		obj := gia.Gia_ManObj(graph, lit>>1)
		var fan *gia.Gia_Obj_t
		if lit&1 != 0 {
			gia.Gia_ObjSetMark1(obj, 1)
			fan = gia.Gia_ObjFanin1(obj)
		} else {
			gia.Gia_ObjSetMark0(obj, 1)
			fan = gia.Gia_ObjFanin0(obj)
		}
		if gia.Gia_ObjPhase(fan) != 0 {
			panic("overlapping carry paths")
		}
		gia.Gia_ObjSetPhaseValue(fan, 1)
	}
	return int32(vec.Vec_IntSize(path))
}
func Gia_ManIteratePaths(graph *gia.Gia_Man_t, carryDelay, minPath, maxPath, limit, ignoreBoxes, verbose int32) int32 {
	gia.Gia_ManCleanLevels(graph, gia.Gia_ManObjNum(graph))
	defer gia.Gia_ManClearLevels(graph)
	gia.Gia_ManCleanMark01(graph)
	gia.Gia_ManCleanPhase(graph)
	for i := int32(0); i < gia.Gia_ManCiNum(graph); i++ {
		gia.Gia_ObjSetPhaseValue(gia.Gia_ManCi(graph, i), 1)
	}
	path := vec.Vec_IntAlloc(100)
	paths := int32(0)
	for i := int32(0); i < limit; i++ {
		var boxes int32
		delay := Gia_ManFindAnnotatedDelay(graph, carryDelay, &boxes, ignoreBoxes)
		result := Gia_ManFindPath(graph, carryDelay, minPath, maxPath, path)
		if result < 0 {
			break
		}
		if result > 0 {
			paths++
		}
		if verbose != 0 {
			fmt.Printf("Iter %5d : Paths = %2d. Boxes = %2d. Total boxes = %6d.  Max delay = %5d.\n", i, paths, result, boxes, delay)
		}
	}
	gia.Gia_ManCleanPhase(graph)
	return 1
}
func Gia_ManDupWithArtificialBoxes(graph *gia.Gia_Man_t, delay, minPath, maxPath, limit, useFanout, xorTrick, ignoreBoxes, verbose int32) *gia.Gia_Man_t {
	Gia_ManIteratePaths(graph, delay, minPath, maxPath, limit, ignoreBoxes, verbose)
	result := Gia_ManDupWithArtificalFaddBoxes(graph, useFanout, xorTrick)
	gia.Gia_ManCleanMark01(graph)
	return result
}
