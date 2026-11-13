package acec

// XOR-tree roots and ranks from src/proof/acec/acecXor.c.
import (
	"fmt"
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
)

func Acec_CheckXors(p *Gia_Man_t, xors *Vec_Int_t) {
	counts := vec.Vec_IntStart(int(gia.Gia_ManObjNum(p)))
	for i := 0; 4*i < vec.Vec_IntSize(xors); i++ {
		if vec.Vec_IntEntry(xors, 4*i+3) == 0 {
			vec.Vec_IntAddToEntry(counts, int(vec.Vec_IntEntry(xors, 4*i)), 1)
		}
	}
	total := 0
	for id, n := range vec.Vec_IntArray(counts) {
		if n > 1 {
			fmt.Printf("*** Obj %d has %d two-input XOR cuts.\n", id, n)
			total++
		} else if n == 1 && gia.Gia_ObjRecognizeExor(gia.Gia_ManObj(p, int32(id)), nil, nil) != 0 {
			fmt.Printf("*** Obj %d cannot be recognized as XOR.\n", id)
		}
	}
	if total == 0 {
		fmt.Println("*** There no multiple two-input XOR cuts.")
	}
}
func Acec_OrderTreeRoots(p *Gia_Man_t, adds, roots, ranks *Vec_Int_t) *Vec_Int_t {
	n := vec.Vec_IntSize(roots)
	order := vec.Vec_IntAlloc(n)
	move := vec.Vec_IntStartFull(n)
	for i := 0; 6*i < vec.Vec_IntSize(adds); i++ {
		node := vec.Vec_IntEntry(adds, 6*i+4)
		rank := vec.Vec_IntEntry(ranks, int(node))
		if rank == -1 {
			continue
		}
		for k := 0; k < 3; k++ {
			r := vec.Vec_IntEntry(ranks, int(vec.Vec_IntEntry(adds, 6*i+k)))
			if r != -1 {
				vec.Vec_IntWriteEntry(move, int(rank), r)
			}
		}
	}
	// The source requires a chain tail referenced by another rank.
	start := -1
	for i := 0; i < n; i++ {
		if vec.Vec_IntEntry(move, i) == -1 && vec.Vec_IntFind(move, int32(i)) >= 0 {
			start = i
			break
		}
	}
	if start < 0 {
		panic("acec: root ordering requires an XOR chain")
	}
	for i := start; ; {
		vec.Vec_IntPush(order, vec.Vec_IntEntry(roots, i))
		next := vec.Vec_IntFind(move, int32(i))
		if next < 0 {
			break
		}
		i = next
	}
	return order
}
func Acec_MapXorOuts(p *Gia_Man_t, xors *Vec_Int_t) *Vec_Bit_t {
	m := vec.Vec_BitStart(int(gia.Gia_ManObjNum(p)))
	for i := 0; 4*i < vec.Vec_IntSize(xors); i++ {
		vec.Vec_BitWriteEntry(m, int(vec.Vec_IntEntry(xors, 4*i)), 1)
	}
	return m
}
func Acec_MapXorOuts2(p *Gia_Man_t, xors, ranks *Vec_Int_t) *Vec_Bit_t {
	m := vec.Vec_BitStart(int(gia.Gia_ManObjNum(p)))
	for i := 0; 4*i < vec.Vec_IntSize(xors); i++ {
		id := vec.Vec_IntEntry(xors, 4*i)
		if vec.Vec_IntEntry(ranks, int(id)) != -1 {
			vec.Vec_BitWriteEntry(m, int(id), 1)
		}
	}
	return m
}
func Acec_MapMajOuts(p *Gia_Man_t, adds *Vec_Int_t) *Vec_Bit_t {
	m := vec.Vec_BitStart(int(gia.Gia_ManObjNum(p)))
	for i := 0; 6*i < vec.Vec_IntSize(adds); i++ {
		vec.Vec_BitWriteEntry(m, int(vec.Vec_IntEntry(adds, 6*i+4)), 1)
	}
	return m
}
func Acec_MapMajOuts2(p *Gia_Man_t, adds, ranks *Vec_Int_t) *Vec_Int_t {
	m := vec.Vec_IntStartFull(int(gia.Gia_ManObjNum(p)))
	for i := 0; 6*i < vec.Vec_IntSize(adds); i++ {
		id := vec.Vec_IntEntry(adds, 6*i+4)
		if vec.Vec_IntEntry(ranks, int(id)) != -1 {
			vec.Vec_IntWriteEntry(m, int(id), int32(i))
		}
	}
	return m
}
func Acec_MapXorIns(p *Gia_Man_t, xors *Vec_Int_t) *Vec_Bit_t {
	m := vec.Vec_BitStart(int(gia.Gia_ManObjNum(p)))
	for i := 0; 4*i < vec.Vec_IntSize(xors); i++ {
		for k := 1; k < 4; k++ {
			vec.Vec_BitWriteEntry(m, int(vec.Vec_IntEntry(xors, 4*i+k)), 1)
		}
	}
	return m
}
func Acec_FindXorRoots(p *Gia_Man_t, xors *Vec_Int_t) *Vec_Int_t {
	ins := Acec_MapXorIns(p, xors)
	roots := vec.Vec_IntAlloc(100)
	for i := 0; 4*i < vec.Vec_IntSize(xors); i++ {
		id := vec.Vec_IntEntry(xors, 4*i)
		if vec.Vec_BitEntry(ins, int(id)) == 0 {
			vec.Vec_IntPushUniqueOrder(roots, id)
		}
	}
	return roots
}
func Acec_RankTrees(p *Gia_Man_t, xors, roots *Vec_Int_t) *Vec_Int_t {
	doubles := vec.Vec_IntAlloc(100)
	ranks := vec.Vec_IntStartFull(int(gia.Gia_ManObjNum(p)))
	for i, id := range vec.Vec_IntArray(roots) {
		vec.Vec_IntWriteEntry(ranks, int(id), int32(i))
	}
	for i := vec.Vec_IntSize(xors)/4 - 1; i >= 0; i-- {
		root := vec.Vec_IntEntry(xors, 4*i)
		rank := vec.Vec_IntEntry(ranks, int(root))
		if rank == -1 {
			continue
		}
		for k := 1; k < 4; k++ {
			id := vec.Vec_IntEntry(xors, 4*i+k)
			if id == 0 {
				continue
			}
			old := vec.Vec_IntEntry(ranks, int(id))
			if old == rank {
				continue
			}
			if old == -1 {
				vec.Vec_IntWriteEntry(ranks, int(id), rank)
			} else {
				vec.Vec_IntPush(doubles, id)
			}
			if old != -1 && gia.Gia_ObjIsAnd(gia.Gia_ManObj(p, id)) != 0 {
				fmt.Printf("Xor node %d belongs to Tree %d and Tree %d.\n", id, old, rank)
			}
		}
	}
	for _, id := range vec.Vec_IntArray(doubles) {
		vec.Vec_IntWriteEntry(ranks, int(id), -1)
	}
	return ranks
}
