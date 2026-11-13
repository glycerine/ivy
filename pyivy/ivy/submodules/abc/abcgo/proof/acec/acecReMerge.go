package acec

// Cut enumeration from src/proof/acec/acecRe.c.
import (
	"fmt"
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
)

func Ree_ManCutPrint(cut []int32, count int32, truth uint64, id int32) {
	fmt.Printf("%d : %d : ", count, id)
	i := int32(1)
	for ; i <= cut[0]; i++ {
		fmt.Printf("%3d ", cut[i])
	}
	for ; i <= 4; i++ {
		fmt.Print("    ")
	}
	fmt.Printf("0x%02x\n", truth&0xff)
}
func Ree_ManCutMerge(p *Gia_Man_t, id int32, list0, list1 []int32, cuts *Vec_Int_t, h *vec.Hash_IntMan_t, data, xors *Vec_Int_t) {
	vec.Vec_IntFill(cuts, 2, 1)
	vec.Vec_IntPush(cuts, id)
	vec.Vec_IntPush(cuts, 0xaa)
	xor2, xor3 := -1, -1
	off0 := 1
	for i := int32(0); i < list0[0]; i++ {
		n0 := list0[off0]
		a := list0[off0 : off0+int(n0)+2]
		off0 += int(n0) + 2
		off1 := 1
		for j := int32(0); j < list1[0]; j++ {
			n1 := list1[off1]
			b := list1[off1 : off1+int(n1)+2]
			off1 += int(n1) + 2
			var cut [6]int32
			if Ree_ManCutMergeOne(a, b, cut[:]) == 0 || Ree_ManCutCheckEqual(cuts, cut[:]) != 0 {
				continue
			}
			truth := Ree_ManCutTruth(gia.Gia_ManObj(p, id), a, b, cut[:])
			normalized := truth
			if truth&0x80 != 0 {
				normalized = ^truth & 0xff
			}
			if normalized == 0x66 && xor2 == -1 {
				xor2 = vec.Vec_IntSize(cuts)
			} else if normalized == 0x69 && xor3 == -1 {
				xor3 = vec.Vec_IntSize(cuts)
			}
			vec.Vec_IntAddToEntry(cuts, 0, 1)
			for _, v := range cut[:1+cut[0]] {
				vec.Vec_IntPush(cuts, v)
			}
			vec.Vec_IntPush(cuts, truth)
			if (normalized == 0x66 || normalized == 0x11 || normalized == 0x22 || normalized == 0x44 || normalized == 0x77) && cut[0] == 2 {
				key := vec.Hsh_Int3ManInsert(h, cut[1], cut[2], 0)
				vec.Vec_IntPushThree(data, id, key, truth)
			} else if normalized == 0x69 || normalized == 0x17 || normalized == 0x2b || normalized == 0x4d || normalized == 0x71 {
				if cut[0] != 3 {
					panic("acec: expected three-leaf truth")
				}
				key := vec.Hsh_Int3ManInsert(h, cut[1], cut[2], cut[3])
				vec.Vec_IntPushThree(data, id, key, truth)
			}
		}
	}
	if xors == nil {
		return
	}
	off := xor2
	if off < 0 {
		off = xor3
	}
	if off < 0 {
		return
	}
	packed := vec.Vec_IntArray(cuts)
	n := packed[off]
	vec.Vec_IntPush(xors, id)
	for k := int32(1); k <= n; k++ {
		vec.Vec_IntPush(xors, packed[off+int(k)])
	}
	if n == 2 {
		vec.Vec_IntPush(xors, 0)
	}
}
func Ree_ManCompare(a, b []int32) int32 {
	if a[3] < b[3] {
		return -1
	}
	if a[3] > b[3] {
		return 1
	}
	if a[4] < b[4] {
		return -1
	}
	if a[4] > b[4] {
		return 1
	}
	return 0
}
