package acec

// Adder coverage diagnostics from src/proof/acec/acecCover.c.
import (
	"fmt"
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
)

func Gia_AcecExplore(p *Gia_Man_t, verbose int32) {
	if !gia.Gia_ManHasMuxData(p) {
		panic("acec: XOR data required")
	}
	nodes := vec.Vec_IntAlloc(100)
	mapping := vec.Vec_BitStart(int(gia.Gia_ManObjNum(p)))
	fadds := Gia_ManDetectFullAdders(p, verbose, nil)
	hadds := Gia_ManDetectHalfAdders(p, verbose)
	for i := 0; i < vec.Vec_IntSize(hadds); i += 2 {
		vec.Vec_BitWriteEntry(mapping, int(vec.Vec_IntEntry(hadds, i)), 1)
		vec.Vec_BitWriteEntry(mapping, int(vec.Vec_IntEntry(hadds, i+1)), 1)
	}
	for i := 0; i < vec.Vec_IntSize(fadds); i += 5 {
		vec.Vec_BitWriteEntry(mapping, int(vec.Vec_IntEntry(fadds, i+3)), 1)
		vec.Vec_BitWriteEntry(mapping, int(vec.Vec_IntEntry(fadds, i+4)), 1)
	}
	gia.Gia_ManCleanMark01(p)
	for i := int32(0); i < gia.Gia_ManCoNum(p); i++ {
		gia.Gia_ObjSetMark0(gia.Gia_ObjFanin0(gia.Gia_ManCo(p, i)), 1)
	}
	xors := Gia_AcecCollectXors(p, mapping)
	fmt.Print("Collected XORs: ")
	vec.Vec_IntPrint(xors)
	for _, id := range vec.Vec_IntArray(xors) {
		o := gia.Gia_ManObj(p, id)
		gia.Gia_ObjSetMark1(o, 1)
		gia.Gia_ObjSetMark0(gia.Gia_ObjFanin0(o), 1)
		gia.Gia_ObjSetMark0(gia.Gia_ObjFanin1(o), 1)
	}
	for i := 0; i < vec.Vec_IntSize(fadds); i += 5 {
		Gia_AcecMarkFadd(p, vec.Vec_IntArray(fadds)[i:i+5])
	}
	count := 0
	for {
		changed := false
		for i := 0; i < vec.Vec_IntSize(hadds); i += 2 {
			s := vec.Vec_IntArray(hadds)[i : i+2]
			a, b := gia.Gia_ManObj(p, s[0]), gia.Gia_ManObj(p, s[1])
			if gia.Gia_ObjMark0(a) == 0 || gia.Gia_ObjMark0(b) == 0 || gia.Gia_ObjMark1(a) != 0 || gia.Gia_ObjMark1(b) != 0 {
				continue
			}
			Gia_AcecMarkHadd(p, s)
			changed = true
			count++
		}
		if !changed {
			break
		}
	}
	for id := int32(1); id < gia.Gia_ManObjNum(p); id++ {
		o := gia.Gia_ManObj(p, id)
		if gia.Gia_ObjIsAnd(o) == 0 || gia.Gia_ObjMark0(o) == 0 || gia.Gia_ObjMark1(o) != 0 {
			continue
		}
		supp := gia.Gia_ManSuppSize(p, []int32{id}, 1)
		cone := gia.Gia_ManConeSize(p, []int32{id}, 1)
		fmt.Printf("Node %5d : Supp = %5d.  Cone = %5d.\n", id, supp, cone)
		vec.Vec_IntPush(nodes, id)
	}
	fmt.Printf("Fadds = %d. Hadds = %d.  Root nodes found = %d.\n", vec.Vec_IntSize(fadds)/5, count, vec.Vec_IntSize(nodes))
	gia.Gia_ManCleanMark01(p)
	for _, id := range vec.Vec_IntArray(nodes) {
		gia.Gia_ObjSetMark0(gia.Gia_ManObj(p, id), 1)
	}
}
func Gia_AcecCover(p *Gia_Man_t) {
	var cuts *Vec_Int_t
	adds := Gia_ManDetectFullAdders(p, 1, &cuts)
	gia.Gia_ManCleanMark01(p)
	for i := 0; i < vec.Vec_IntSize(adds); i += 5 {
		Gia_AcecMarkFadd(p, vec.Vec_IntArray(adds)[i:i+5])
	}
	for i, id := range vec.Vec_IntArray(cuts) {
		if i%3 == 2 && gia.Gia_ObjMark1(gia.Gia_ManObj(p, id)) == 0 {
			fmt.Printf("%d ", id)
		}
	}
	fmt.Println()
	gia.Gia_ManCleanMark01(p)
}
