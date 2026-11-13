package acec

// Arithmetic cut detection from src/proof/acec/acecRe.c.
import (
	"fmt"
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"os"
	"sort"
)

func Ree_TruthPrecompute() {
	truths := [8]uint64{0, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77}
	for _, t := range truths {
		t = util.Abc_Tt6SwapAdjacent(t, 1)
		util.Abc_TtPrintHexRev(os.Stdout, []uint64{t}, 3)
		fmt.Println()
	}
	fmt.Println()
	for _, t := range truths {
		t = util.Abc_Tt6SwapAdjacent(t, 1)
		t = util.Abc_Tt6SwapAdjacent(t, 0)
		util.Abc_TtPrintHexRev(os.Stdout, []uint64{t}, 3)
		fmt.Println()
	}
	fmt.Println()
}
func Ree_TruthPrecompute2() {
	for i := 0; i < 8; i++ {
		t := uint64(0xe8)
		for b := 0; b < 3; b++ {
			if i>>b&1 != 0 {
				t = util.Abc_Tt6Flip(t, int32(b))
			}
		}
		fmt.Printf("%d = %X\n", i, t&0xff)
	}
}
func Ree_ManComputeCuts(p *Gia_Man_t, pxors **Vec_Int_t, verbose int32) *Vec_Int_t {

	h := vec.Hash_IntManStart(1000)
	temp, data := vec.Vec_IntAlloc(1000), vec.Vec_IntAlloc(1000)
	cuts := vec.Vec_IntAlloc(30 * int(gia.Gia_ManAndNum(p)))
	vec.Vec_IntFill(cuts, int(gia.Gia_ManObjNum(p)), 0)
	gia.Gia_ManCleanValue(p)
	for i := int32(0); i < gia.Gia_ManCiNum(p); i++ {
		o := gia.Gia_ManCi(p, i)
		id := gia.Gia_ObjId(p, o)
		vec.Vec_IntWriteEntry(cuts, int(id), int32(vec.Vec_IntSize(cuts)))
		for _, v := range []int32{1, 1, id, 0xaa} {
			vec.Vec_IntPush(cuts, v)
		}
	}
	var xors *Vec_Int_t
	if pxors != nil {
		xors = vec.Vec_IntAlloc(1000)
		*pxors = xors
	}
	total := 0
	for id := int32(1); id < gia.Gia_ManObjNum(p); id++ {
		o := gia.Gia_ManObj(p, id)
		if gia.Gia_ObjIsAnd(o) == 0 {
			continue
		}
		list := vec.Vec_IntArray(cuts)
		a, b := int(vec.Vec_IntEntry(cuts, int(gia.Gia_ObjFaninId0(o, id)))), int(vec.Vec_IntEntry(cuts, int(gia.Gia_ObjFaninId1(o, id))))
		Ree_ManCutMerge(p, id, list[a:], list[b:], temp, h, data, xors)
		vec.Vec_IntWriteEntry(cuts, int(id), int32(vec.Vec_IntSize(cuts)))
		vec.Vec_IntAppend(cuts, temp)
		total += int(vec.Vec_IntEntry(temp, 0))
	}
	if verbose != 0 {
		nodes := float64(gia.Gia_ManAndNum(p))
		fmt.Printf("AIG nodes = %d.  Cuts = %d.  Cuts/Node = %.2f.  Ints/Node = %.2f.\n", gia.Gia_ManAndNum(p), total, float64(total)/nodes, float64(vec.Vec_IntSize(cuts))/nodes)
	}
	adds := Ree_ManDeriveAdds(h, data, verbose)
	raw := vec.Vec_IntArray(adds)
	records := make([][6]int32, len(raw)/6)
	for i := range records {
		copy(records[i][:], raw[6*i:6*i+6])
	}
	sort.SliceStable(records, func(i, j int) bool { return Ree_ManCompare(records[i][:], records[j][:]) < 0 })
	for i := range records {
		copy(raw[6*i:6*i+6], records[i][:])
	}
	if verbose != 0 {
		fmt.Printf("Adders = %d.  Total cuts = %d.  Hashed cuts = %d.  Hashed/Adders = %.2f.\n", len(records), vec.Vec_IntSize(data)/3, vec.Hash_IntManEntryNum(h), 6*float64(vec.Hash_IntManEntryNum(h))/float64(vec.Vec_IntSize(adds)))
	}
	vec.Hash_IntManStop(h)
	Ree_ManRemoveTrivial(p, adds)
	Ree_ManRemoveContained(p, adds)
	return adds
}
