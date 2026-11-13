package acec

// XOR-tree structural analysis, ported from src/proof/acec/acecStruct.c.
import (
	"fmt"
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
)

func Acec_StructDetectXorRoots(p *Gia_Man_t) *Vec_Int_t {
	roots := vec.Vec_IntAlloc(100)
	ins := make([]bool, gia.Gia_ManObjNum(p))
	for id := int32(1); id < gia.Gia_ManObjNum(p); id++ {
		o := gia.Gia_ManObj(p, id)
		if gia.Gia_ObjIsAnd(o) == 0 {
			continue
		}
		var a, b *Gia_Obj_t
		if gia.Gia_ObjRecognizeExor(o, &a, &b) == 0 {
			continue
		}
		vec.Vec_IntPush(roots, id)
		ins[gia.Gia_ObjId(p, gia.Gia_Regular(a))] = true
		ins[gia.Gia_ObjId(p, gia.Gia_Regular(b))] = true
	}
	k := 0
	for i := 0; i < vec.Vec_IntSize(roots); i++ {
		id := vec.Vec_IntEntry(roots, i)
		if !ins[id] {
			vec.Vec_IntWriteEntry(roots, k, id)
			k++
		}
	}
	vec.Vec_IntShrink(roots, k)
	return roots
}
func Acec_StructAssignRanks(p *Gia_Man_t, roots *Vec_Int_t) *Vec_Int_t {
	ranks := vec.Vec_IntStartFull(int(gia.Gia_ManObjNum(p)))
	doubles := vec.Vec_IntAlloc(100)
	for i := 0; i < vec.Vec_IntSize(roots); i++ {
		vec.Vec_IntWriteEntry(ranks, int(vec.Vec_IntEntry(roots, i)), int32(i))
	}
	for id := gia.Gia_ManObjNum(p) - 1; id > 0; id-- {
		o := gia.Gia_ManObj(p, id)
		if gia.Gia_ObjIsAnd(o) == 0 {
			continue
		}
		var a, b *Gia_Obj_t
		if gia.Gia_ObjRecognizeExor(o, &a, &b) == 0 {
			continue
		}
		rank := vec.Vec_IntEntry(ranks, int(id))
		if rank == -1 {
			continue
		}
		for _, f := range []int32{gia.Gia_ObjId(p, gia.Gia_Regular(a)), gia.Gia_ObjId(p, gia.Gia_Regular(b))} {
			old := vec.Vec_IntEntry(ranks, int(f))
			if old == rank {
				continue
			}
			if old == -1 {
				vec.Vec_IntWriteEntry(ranks, int(f), rank)
			} else {
				vec.Vec_IntPush(doubles, f)
			}
			if old != -1 && gia.Gia_ObjIsAnd(gia.Gia_ManObj(p, f)) != 0 {
				fmt.Printf("Xor node %d belongs to Tree %d and Tree %d.\n", f, old, rank)
			}
		}
	}
	for i := 0; i < vec.Vec_IntSize(doubles); i++ {
		vec.Vec_IntWriteEntry(ranks, int(vec.Vec_IntEntry(doubles, i)), -1)
	}
	return ranks
}
func Acec_FindTreeLeaves(p *Gia_Man_t, roots, ranks *Vec_Int_t) *Vec_Wec_t {
	seen := make([]bool, gia.Gia_ManObjNum(p))
	leaves := vec.Vec_WecStart(vec.Vec_IntSize(roots))
	for id := int32(1); id < gia.Gia_ManObjNum(p); id++ {
		o := gia.Gia_ManObj(p, id)
		if gia.Gia_ObjIsAnd(o) == 0 {
			continue
		}
		var a, b *Gia_Obj_t
		if gia.Gia_ObjRecognizeExor(o, &a, &b) == 0 {
			continue
		}
		seen[id] = true
		rank := vec.Vec_IntEntry(ranks, int(id))
		if rank == -1 {
			continue
		}
		for _, f := range []int32{gia.Gia_ObjId(p, gia.Gia_Regular(a)), gia.Gia_ObjId(p, gia.Gia_Regular(b))} {
			if seen[f] {
				if rank != vec.Vec_IntEntry(ranks, int(f)) {
					panic("acec: XOR tree rank mismatch")
				}
				continue
			}
			vec.Vec_WecPush(leaves, int(rank), f)
		}
	}
	return leaves
}
func Acec_FindShadows(p *Gia_Man_t, ranks *Vec_Int_t) *Vec_Int_t {
	s := vec.Vec_IntDup(ranks)
	for i := int32(0); i < gia.Gia_ManCiNum(p); i++ {
		vec.Vec_IntWriteEntry(s, int(gia.Gia_ObjId(p, gia.Gia_ManCi(p, i))), -1)
	}
	for id := int32(1); id < gia.Gia_ManObjNum(p); id++ {
		o := gia.Gia_ManObj(p, id)
		if gia.Gia_ObjIsAnd(o) == 0 || vec.Vec_IntEntry(s, int(id)) >= 0 {
			continue
		}
		a := vec.Vec_IntEntry(s, int(gia.Gia_ObjFaninId0(o, id)))
		b := vec.Vec_IntEntry(s, int(gia.Gia_ObjFaninId1(o, id)))
		if a == b && a != -1 {
			vec.Vec_IntWriteEntry(s, int(id), a)
		}
	}
	return s
}
func Acec_CollectSupp_rec(p *Gia_Man_t, node, rank int32, ranks *Vec_Int_t) int32 {
	if gia.Gia_ObjIsTravIdCurrentId(p, node) != 0 {
		return 0
	}
	gia.Gia_ObjSetTravIdCurrentId(p, node)
	o := gia.Gia_ManObj(p, node)
	if gia.Gia_ObjIsAnd(o) == 0 {
		panic("acec: expected AND in support")
	}
	if vec.Vec_IntEntry(ranks, int(node)) == rank {
		return 1
	}
	return Acec_CollectSupp_rec(p, gia.Gia_ObjFaninId0(o, node), rank, ranks) + Acec_CollectSupp_rec(p, gia.Gia_ObjFaninId1(o, node), rank, ranks)
}
func Acec_FindNexts(p *Gia_Man_t, ranks, shadows *Vec_Int_t, leaves *Vec_Wec_t) *Vec_Wec_t {
	nexts := vec.Vec_WecStart(vec.Vec_WecSize(leaves))
	for i := 0; i < vec.Vec_WecSize(leaves); i++ {
		tree := vec.Vec_WecEntry(leaves, i)
		for j := 0; j < vec.Vec_IntSize(tree); j++ {
			node := vec.Vec_IntEntry(tree, j)
			o := gia.Gia_ManObj(p, node)
			if gia.Gia_ObjIsAnd(o) == 0 {
				continue
			}
			a := vec.Vec_IntEntry(shadows, int(gia.Gia_ObjFaninId0(o, node)))
			b := vec.Vec_IntEntry(shadows, int(gia.Gia_ObjFaninId1(o, node)))
			if a != b || a == -1 {
				continue
			}
			rank := vec.Vec_IntEntry(ranks, int(node))
			if rank == a {
				panic("acec: unexpected shadow rank")
			}
			gia.Gia_ManIncrementTravId(p)
			n := Acec_CollectSupp_rec(p, node, a, ranks)
			if n <= 1 {
				panic("acec: invalid support size")
			}
			if n <= 3 {
				vec.Vec_IntPushUniqueOrder(vec.Vec_WecEntry(nexts, int(a)), rank)
			}
		}
	}
	return nexts
}
func Acec_StructTest(p *Gia_Man_t) {}
