package cec

// Structural-isomorphism simulation and matching, ported from cecIso.c by
// Alan Mishchenko (UC Berkeley, 2005).
import (
	"fmt"

	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

func Cec_ManIsoInfo(store []uint32, words, id int32) []uint32 {
	return store[int(words*id):int(words*(id+1))]
}

func Gia_ManIsoSimulate(obj *gia.Gia_Obj_t, id int32, store []uint32, words int32) {
	out := Cec_ManIsoInfo(store, words, id)
	a := Cec_ManIsoInfo(store, words, gia.Gia_ObjFaninId0(obj, id))
	b := Cec_ManIsoInfo(store, words, gia.Gia_ObjFaninId1(obj, id))
	for i := range out {
		x, y := a[i], b[i]
		if gia.Gia_ObjFaninC0(obj) != 0 {
			x = ^x
		}
		if gia.Gia_ObjFaninC1(obj) != 0 {
			y = ^y
		}
		out[i] = x & y
	}
}

func Gia_ManIsoCopy(dst, src int32, store []uint32, words int32) {
	copy(Cec_ManIsoInfo(store, words, dst), Cec_ManIsoInfo(store, words, src))
}

func Gia_ManIsoEqual(a, b int32, store []uint32, words int32) int32 {
	x, y := Cec_ManIsoInfo(store, words, a), Cec_ManIsoInfo(store, words, b)
	for i := range x {
		if x[i] != y[i] {
			return 0
		}
	}
	return 1
}

func Gia_ManIsoRandom(id int32, store []uint32, words int32) {
	for i := range Cec_ManIsoInfo(store, words, id) {
		Cec_ManIsoInfo(store, words, id)[i] = gia.Gia_ManRandom(0)
	}
}

var isoHashPrimes = [...]uint32{1291, 1699, 1999, 2357, 2953, 3313, 3907, 4177, 4831, 5147, 5647, 6343, 6899, 7103, 7873, 8147}

func Gia_ManIsoHashKey(id int32, store []uint32, words, tableSize int32) int32 {
	var hash uint32
	for i, word := range Cec_ManIsoInfo(store, words, id) {
		hash ^= word * isoHashPrimes[i&15]
	}
	return int32(hash % uint32(tableSize))
}

func Gia_ManIsoTableAdd(p *gia.Gia_Man_t, id int32, store []uint32, words int32, table []int32, tableSize int32) {
	color := gia.Gia_ObjColors(p, id)
	if color != 1 && color != 2 {
		panic("cec: invalid isomorphism color")
	}
	key := Gia_ManIsoHashKey(id, store, words, tableSize)
	for entry := table[key]; entry != 0; entry = gia.Gia_ObjValue(gia.Gia_ManObj(p, entry)) {
		obj := gia.Gia_ManObj(p, entry)
		if gia.Gia_ObjColors(p, entry) == color && Gia_ManIsoEqual(id, entry, store, words) != 0 {
			gia.Gia_ObjSetMark0(obj, 1)
			return
		}
	}
	obj := gia.Gia_ManObj(p, id)
	if gia.Gia_ObjValue(obj) != 0 || gia.Gia_ObjMark0(obj) != 0 {
		panic("cec: dirty isomorphism table node")
	}
	gia.Gia_ObjSetValue(obj, table[key])
	table[key] = id
}

func Gia_ManIsoExtractClasses(p *gia.Gia_Man_t, bin int32, store []uint32, words int32, a, b *vec.Vec_Int_t) int32 {
	vec.Vec_IntClear(a)
	vec.Vec_IntClear(b)
	for entry := bin; entry != 0; entry = gia.Gia_ObjValue(gia.Gia_ManObj(p, entry)) {
		obj := gia.Gia_ManObj(p, entry)
		if gia.Gia_ObjMark0(obj) != 0 {
			gia.Gia_ObjSetMark0(obj, 0)
			continue
		}
		if gia.Gia_ObjColors(p, entry) == 1 {
			vec.Vec_IntPush(a, entry)
		} else {
			vec.Vec_IntPush(b, entry)
		}
	}
	if vec.Vec_IntSize(a) > 0 && vec.Vec_IntSize(b) > 0 {
		return 1
	}
	return 0
}

func Gia_ManIsoMatchNodes(iso []int32, store []uint32, words int32, a, b *vec.Vec_Int_t) {
	for _, idA := range vec.Vec_IntArray(a) {
		for _, idB := range vec.Vec_IntArray(b) {
			if Gia_ManIsoEqual(idA, idB, store, words) == 0 {
				continue
			}
			if iso[idA] != 0 || iso[idB] != 0 || idA == idB {
				panic("cec: ambiguous isomorphism pair")
			}
			iso[idA], iso[idB] = idB, idA
		}
	}
}

func Cec_ManTransformClasses(p *gia.Gia_Man_t) {
	iso := p.Isomorphism()
	if !p.HasRepresentatives() || len(iso) != int(gia.Gia_ManObjNum(p)) {
		panic("cec: missing isomorphism pairs")
	}
	p.StartRepresentatives()
	for id, partner := range iso {
		if partner != 0 && partner < int32(id) {
			gia.Gia_ObjSetRepr(p, int32(id), partner)
			gia.Gia_ObjSetNext(p, partner, int32(id))
		}
	}
}

// Cec_ManDetectIsomorphism assigns matching IDs for colored nodes with unique
// simulation signatures. Shared-color nodes map to themselves.
func Cec_ManDetectIsomorphism(p *gia.Gia_Man_t) []int32 {
	const words int32 = 2
	n := gia.Gia_ManObjNum(p)
	iso := make([]int32, n)
	for id := int32(0); id < n; id++ {
		obj := gia.Gia_ManObj(p, id)
		color := gia.Gia_ObjColors(p, id)
		if gia.Gia_ObjIsCo(obj) != 0 {
			if color != 0 {
				panic("cec: colored output in isomorphism input")
			}
			continue
		}
		if color == 0 {
			panic("cec: uncolored isomorphism input")
		}
		if color == 3 {
			iso[id] = id
		}
	}
	store := make([]uint32, n*words)
	tableSize := util.Abc_PrimeCudd(uint32(100 + n/2))
	table := make([]int32, tableSize)
	gia.Gia_ManCleanValue(p)
	for id := int32(1); id < n; id++ {
		obj := gia.Gia_ManObj(p, id)
		if gia.Gia_ObjIsCo(obj) != 0 {
			continue
		}
		if iso[id] == 0 {
			Gia_ManIsoSimulate(obj, id, store, words)
		} else if iso[id] < id {
			Gia_ManIsoCopy(id, iso[id], store, words)
		} else {
			Gia_ManIsoRandom(id, store, words)
		}
		if iso[id] == 0 {
			Gia_ManIsoTableAdd(p, id, store, words, table, tableSize)
		}
	}
	a, b := vec.Vec_IntAlloc(100), vec.Vec_IntAlloc(100)
	for _, bin := range table {
		if Gia_ManIsoExtractClasses(p, bin, store, words, a, b) != 0 {
			Gia_ManIsoMatchNodes(iso, store, words, a, b)
		}
	}
	var count int
	for id, partner := range iso {
		if partner != 0 && partner < int32(id) {
			count++
		}
	}
	fmt.Printf("Computed %d pairs of structurally equivalent nodes.\n", count)
	return iso
}
