package if_

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/opt/dau"
	"testing"
)

func TestGiaCofactorLutTruthAndMapping(t *testing.T) {
	for _, expression := range []string{"<abc>", "(ab)", "!(ab)", "(!ab)", "!(!ab)"} {
		t.Run(expression, func(t *testing.T) {
			p, cut := ifDelayTestManager(t)
			defer If_ManStop(p)
			truth := dau.Dau_DsdToTruth(expression, 3)
			p.vTtMem[3] = vec.Vec_MemAllocForTT(6, 0)
			defer vec.Vec_MemFree(p.vTtMem[3])
			id := vec.Vec_MemHashInsert(p.vTtMem[3], truth)
			p.nTruth6Words[3] = 1
			cut.iCutFunc = int32(2 * id)
			p.vTtVars[3] = vec.Vec_StrAlloc(id + 1)
			vec.Vec_StrFill(p.vTtVars[3], id+1, 0)
			dst := Gia_ManStart(32)
			leaves, temp, mapping := Vec_IntAlloc(3), Vec_IntAlloc(3), Vec_IntAlloc(0)
			records := vec.Vec_IntAllocArrayCopy([]int32{0}, 1)
			for i := 0; i < 3; i++ {
				Vec_IntPush(leaves, Gia_ManAppendCi(dst))
			}
			result := Gia_ManFromIfLogicCofVars(dst, p, cut, leaves, temp, nil, mapping, records)
			offset := vec.Vec_IntEntry(mapping, int(result>>1))
			count := vec.Vec_IntEntry(records, int(offset))
			marker := vec.Vec_IntEntry(records, int(offset+count+1))
			if offset == 0 || marker >= 0 {
				t.Fatalf("bad mux mapping offset=%d marker=%d records=%v", offset, marker, Vec_IntArray(records))
			}
			for mask := 0; mask < 8; mask++ {
				values := make([]bool, Gia_ManObjNum(dst))
				for i, input := range Vec_IntArray(leaves) {
					values[input>>1] = mask>>uint(i)&1 != 0
				}
				for id := int32(1); id < Gia_ManObjNum(dst); id++ {
					obj := Gia_ManObj(dst, id)
					if gia.Gia_ObjIsAnd(obj) == 0 {
						continue
					}
					a := values[gia.Gia_ObjFaninId0(obj, id)] != (gia.Gia_ObjFaninC0(obj) != 0)
					b := values[gia.Gia_ObjFaninId1(obj, id)] != (gia.Gia_ObjFaninC1(obj) != 0)
					values[id] = a && b
				}
				got := values[result>>1] != (result&1 != 0)
				want := truth[0]>>uint(mask)&1 != 0
				if got != want {
					t.Fatalf("assignment %d: got %v, want %v", mask, got, want)
				}
			}
		})
	}
}

func TestGiaCofactorLutBothInternal(t *testing.T) {
	pars := &If_Par_t_{NLutSize: 5, NCutsMax: 4}
	p := If_ManStart(pars)
	defer If_ManStop(p)
	truth := dau.Dau_DsdToTruth("<a(bc)(de)>", 5)
	p.vTtMem[5] = vec.Vec_MemAllocForTT(6, 0)
	defer vec.Vec_MemFree(p.vTtMem[5])
	id := vec.Vec_MemHashInsert(p.vTtMem[5], truth)
	p.nTruth6Words[5] = 1
	p.vTtVars[5] = vec.Vec_StrAlloc(id + 1)
	vec.Vec_StrFill(p.vTtVars[5], id+1, 0)
	cut := &If_Cut_t{nLeaves: 5, iCutFunc: int32(id * 2)}
	dst := Gia_ManStart(64)
	leaves, temp, mapping := Vec_IntAlloc(5), Vec_IntAlloc(5), Vec_IntAlloc(0)
	records := vec.Vec_IntAllocArrayCopy([]int32{0}, 1)
	for i := 0; i < 5; i++ {
		Vec_IntPush(leaves, Gia_ManAppendCi(dst))
	}
	result := Gia_ManFromIfLogicCofVars(dst, p, cut, leaves, temp, nil, mapping, records)
	offset := vec.Vec_IntEntry(mapping, int(result>>1))
	if size := vec.Vec_IntEntry(records, int(offset)); size != 3 {
		t.Fatalf("mux size %d, want 3", size)
	}
	if marker := vec.Vec_IntEntry(records, int(offset+4)); marker != -(result >> 1) {
		t.Fatalf("mux marker %d", marker)
	}
	for mask := 0; mask < 32; mask++ {
		values := make([]bool, Gia_ManObjNum(dst))
		for i, lit := range Vec_IntArray(leaves) {
			values[lit>>1] = mask>>uint(i)&1 != 0
		}
		for id := int32(1); id < Gia_ManObjNum(dst); id++ {
			obj := Gia_ManObj(dst, id)
			if gia.Gia_ObjIsAnd(obj) == 0 {
				continue
			}
			a := values[gia.Gia_ObjFaninId0(obj, id)] != (gia.Gia_ObjFaninC0(obj) != 0)
			b := values[gia.Gia_ObjFaninId1(obj, id)] != (gia.Gia_ObjFaninC1(obj) != 0)
			values[id] = a && b
		}
		got := values[result>>1] != (result&1 != 0)
		if want := truth[0]>>uint(mask)&1 != 0; got != want {
			t.Fatalf("assignment %d: got %v, want %v", mask, got, want)
		}
	}
}

func TestGiaBiDecompositionTruthAndMapping(t *testing.T) {
	for _, tc := range []struct {
		name, expression string
		or               bool
	}{
		{"and", "(abcd)", false}, {"or", "!(!(ab)!(cd))", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for _, truthPhase := range []int32{0, 1} {
				for _, cutPhase := range []uint32{0, 1} {
					pars := &If_Par_t_{NLutSize: 4, NCutsMax: 4}
					p := If_ManStart(pars)
					truth := dau.Dau_DsdToTruth(tc.expression, 4)
					p.vTtMem[4] = vec.Vec_MemAllocForTT(6, 0)
					id := vec.Vec_MemHashInsert(p.vTtMem[4], truth)
					p.nTruth6Words[4] = 1
					p.vTtDecs[4] = vec.Vec_IntStart(id + 1)
					mask := int32(3 | 12<<16)
					if tc.or {
						mask |= 1 << 30
					}
					vec.Vec_IntWriteEntry(p.vTtDecs[4], id, mask)
					cut := &If_Cut_t{nLeaves: 4, iCutFunc: int32(id*2) | truthPhase, fCompl: cutPhase}
					dst := Gia_ManStart(32)
					leaves, temp, mapping := Vec_IntAlloc(4), Vec_IntAlloc(4), Vec_IntAlloc(0)
					records := vec.Vec_IntAllocArrayCopy([]int32{0}, 1)
					for i := 0; i < 4; i++ {
						Vec_IntPush(leaves, Gia_ManAppendCi(dst))
					}
					lit := Gia_ManFromIfLogicAndVars(dst, p, cut, leaves, temp, nil, mapping, records)
					offset := vec.Vec_IntEntry(mapping, int(lit>>1))
					if offset == 0 || vec.Vec_IntEntry(records, int(offset)) != 2 || vec.Vec_IntEntry(records, int(offset+3)) != -(lit>>1) {
						t.Fatalf("bad mapping: %v", Vec_IntArray(records))
					}
					for x := 0; x < 16; x++ {
						values := make([]bool, Gia_ManObjNum(dst))
						for i, input := range Vec_IntArray(leaves) {
							values[input>>1] = x>>uint(i)&1 != 0
						}
						for id := int32(1); id < Gia_ManObjNum(dst); id++ {
							obj := Gia_ManObj(dst, id)
							if gia.Gia_ObjIsAnd(obj) == 0 {
								continue
							}
							a := values[gia.Gia_ObjFaninId0(obj, id)] != (gia.Gia_ObjFaninC0(obj) != 0)
							b := values[gia.Gia_ObjFaninId1(obj, id)] != (gia.Gia_ObjFaninC1(obj) != 0)
							values[id] = a && b
						}
						got := values[lit>>1] != (lit&1 != 0)
						want := (truth[0]>>uint(x)&1 != 0) != (truthPhase^int32(cutPhase) != 0)
						if got != want {
							t.Fatalf("truthPhase=%d cutPhase=%d assignment=%d got %v want %v", truthPhase, cutPhase, x, got, want)
						}
					}
					vec.Vec_MemFree(p.vTtMem[4])
					If_ManStop(p)
				}
			}
		})
	}
}
