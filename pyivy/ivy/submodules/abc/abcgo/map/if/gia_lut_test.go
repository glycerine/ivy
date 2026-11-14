package if_

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestCreateMappedLut(t *testing.T) {
	dst := Gia_ManStart(12)
	a, b := Gia_ManAppendCi(dst), Gia_ManAppendCi(dst)
	leaves := vec.Vec_IntAllocArrayCopy([]int32{a ^ 1, b}, 2)
	mapping := vec.Vec_IntAlloc(0)
	records := vec.Vec_IntAllocArrayCopy([]int32{0}, 1)
	defer Vec_IntFree(leaves)
	defer Vec_IntFree(mapping)
	defer Vec_IntFree(records)
	// x0 AND x1, where the first supplied leaf is complemented.
	lit := Gia_ManFromIfLogicCreateLut(dst, []word{0x8888888888888888}, leaves, nil, mapping, records)
	if lit>>1 <= b>>1 {
		t.Fatalf("nontrivial function returned %d", lit)
	}
	if got := vec.Vec_IntEntry(mapping, int(lit>>1)); got != 1 {
		t.Fatalf("mapping offset = %d, want 1", got)
	}
	want := []int32{0, 2, a >> 1, b >> 1, lit >> 1}
	got := vec.Vec_IntArray(records)
	if len(got) != len(want) {
		t.Fatalf("records = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("records = %v, want %v", got, want)
		}
	}
	obj := Gia_ManObj(dst, lit>>1)
	if Gia_ObjIsAnd(obj) == 0 {
		t.Fatal("LUT root is not an AND")
	}
	for mask := 0; mask < 4; mask++ {
		vals := make([]bool, Gia_ManObjNum(dst))
		vals[a>>1], vals[b>>1] = mask&1 != 0, mask&2 != 0
		for id := int32(1); id < Gia_ManObjNum(dst); id++ {
			node := Gia_ManObj(dst, id)
			if gia.Gia_ObjIsAnd(node) == 0 {
				continue
			}
			x := vals[gia.Gia_ObjFaninId0(node, id)] != (gia.Gia_ObjFaninC0(node) != 0)
			y := vals[gia.Gia_ObjFaninId1(node, id)] != (gia.Gia_ObjFaninC1(node) != 0)
			vals[id] = x && y
		}
		got := vals[lit>>1] != (lit&1 != 0)
		if want := mask&1 == 0 && mask&2 != 0; got != want {
			t.Fatalf("assignment %d: got %v, want %v", mask, got, want)
		}
	}
	for _, truth := range []word{0, ^word(0), 0xaaaaaaaaaaaaaaaa, 0x5555555555555555} {
		before := vec.Vec_IntSize(records)
		simple := Gia_ManFromIfLogicCreateLut(dst, []word{truth}, leaves, nil, mapping, records)
		if simple>>1 != 0 && simple>>1 != a>>1 && simple>>1 != b>>1 {
			t.Fatalf("simple truth %#x produced %d", truth, simple)
		}
		if vec.Vec_IntSize(records) != before {
			t.Fatalf("simple truth %#x added a LUT", truth)
		}
	}
}

func TestCreateSpecialMappedLuts(t *testing.T) {
	for n := 5; n <= 7; n++ {
		t.Run(fmt.Sprintf("%d inputs", n), func(t *testing.T) {
			dst := Gia_ManStart(64)
			leaves, temp := Vec_IntAlloc(n), Vec_IntAlloc(4)
			mapping, records := Vec_IntAlloc(0), vec.Vec_IntAllocArrayCopy([]int32{0}, 1)
			packing := vec.Vec_IntAllocArrayCopy([]int32{0}, 1)
			for i := 0; i < n; i++ {
				Vec_IntPush(leaves, Gia_ManAppendCi(dst))
			}
			truth := make([]word, (1<<uint(n)+63)/64)
			for x := 0; x < 1<<uint(n); x++ {
				bound := x&15 == 15
				rest := x>>4 == (1<<uint(n-4))-1
				if bound != rest {
					truth[x>>6] |= word(1) << uint(x&63)
				}
			}
			if n < 6 {
				truth[0] |= truth[0] << 32
			}
			lit := Gia_ManFromIfLogicCreateLutSpecial(dst, truth, leaves, temp, nil, mapping, records, packing)
			if got := vec.Vec_IntArray(packing); len(got) != 4 || got[0] != 1 || got[1] != 2 || got[3] != lit>>1 {
				t.Fatalf("bad packing %v", got)
			}
			for mask := 0; mask < 1<<uint(n); mask++ {
				vals := make([]bool, Gia_ManObjNum(dst))
				for i, input := range vec.Vec_IntArray(leaves) {
					vals[input>>1] = mask>>uint(i)&1 != 0
				}
				for id := int32(1); id < Gia_ManObjNum(dst); id++ {
					obj := Gia_ManObj(dst, id)
					if gia.Gia_ObjIsAnd(obj) == 0 {
						continue
					}
					a := vals[gia.Gia_ObjFaninId0(obj, id)] != (gia.Gia_ObjFaninC0(obj) != 0)
					b := vals[gia.Gia_ObjFaninId1(obj, id)] != (gia.Gia_ObjFaninC1(obj) != 0)
					vals[id] = a && b
				}
				got := vals[lit>>1] != (lit&1 != 0)
				want := truth[mask>>6]>>uint(mask&63)&1 != 0
				if got != want {
					t.Fatalf("assignment %d: got %v, want %v", mask, got, want)
				}
			}
		})
	}
}

func TestFindLutReconstructsPartition(t *testing.T) {
	for _, mode := range []string{"selected", "searched"} {
		t.Run(mode, func(t *testing.T) {
			p, cut := ifDelayTestManager(t)
			defer If_ManStop(p)
			p.pPars.PLutStruct = "2"
			if mode == "selected" {
				p.pPars.FDelayOptLut = 1
				cut.uMaskFunc = 1 | 1<<2
			}
			truth := []uint64{0x8080808080808080}
			p.vTtMem[3] = vec.Vec_MemAllocForTT(6, 0)
			defer vec.Vec_MemFree(p.vTtMem[3])
			id := vec.Vec_MemHashInsert(p.vTtMem[3], truth)
			p.nTruth6Words[3] = 1
			cut.iCutFunc = int32(id * 2)
			p.pIfDsdMan = If_DsdManAlloc(3, 2)
			defer If_DsdManFree(p.pIfDsdMan, 0)
			pins := make([]byte, 3)
			dsd := If_DsdManCompute(p.pIfDsdMan, truth, 3, pins, "2")
			p.vTtDsds[3] = vec.Vec_IntStart(id + 1)
			vec.Vec_IntWriteEntry(p.vTtDsds[3], id, dsd)
			p.vTtPerms[3] = vec.Vec_StrAlloc((id + 1) * 6)
			vec.Vec_StrFill(p.vTtPerms[3], (id+1)*6, 0)
			for i, pin := range pins {
				vec.Vec_StrWriteEntry(p.vTtPerms[3], id*6+i, pin)
			}
			dst := Gia_ManStart(20)
			leaves, scratch, mapping := Vec_IntAlloc(3), Vec_IntAlloc(3), Vec_IntAlloc(0)
			records, packing := vec.Vec_IntAllocArrayCopy([]int32{0}, 1), vec.Vec_IntAllocArrayCopy([]int32{0}, 1)
			for i := 0; i < 3; i++ {
				Vec_IntPush(leaves, Gia_ManAppendCi(dst))
			}
			lit := Gia_ManFromIfLogicFindLut(p, dst, cut, nil, leaves, scratch, nil, mapping, records, packing)
			if got := vec.Vec_IntArray(packing); len(got) != 4 || got[0] != 1 || got[1] != 2 || got[3] != lit>>1 {
				t.Fatalf("packing %v", got)
			}
			for mask := 0; mask < 8; mask++ {
				values := make([]bool, Gia_ManObjNum(dst))
				for i, input := range Vec_IntArray(leaves) {
					values[input>>1] = mask>>uint(i)&1 != 0
				}
				for i := int32(1); i < Gia_ManObjNum(dst); i++ {
					obj := Gia_ManObj(dst, i)
					if gia.Gia_ObjIsAnd(obj) == 0 {
						continue
					}
					a := values[gia.Gia_ObjFaninId0(obj, i)] != (gia.Gia_ObjFaninC0(obj) != 0)
					b := values[gia.Gia_ObjFaninId1(obj, i)] != (gia.Gia_ObjFaninC1(obj) != 0)
					values[i] = a && b
				}
				got := values[lit>>1] != (lit&1 != 0)
				want := truth[0]>>uint(mask)&1 != 0
				if got != want {
					t.Fatalf("assignment %d: got %v want %v", mask, got, want)
				}
			}
		})
	}
}
