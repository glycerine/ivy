package if_

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGiaFromIfLogicMapping(t *testing.T) {
	for _, useTruth := range []bool{false, true} {
		t.Run(map[bool]string{false: "structural", true: "truth"}[useTruth], func(t *testing.T) {
			pars := &If_Par_t_{NLutSize: 4, NCutsMax: 4}
			if useTruth {
				pars.FTruth, pars.FDeriveLuts = 1, 1
			}
			p := If_ManStart(pars)
			defer If_ManStop(p)
			a, b := If_ManCreateCi(p), If_ManCreateCi(p)
			root := If_ManCreateAnd(p, a, If_Not(b))
			root.CutBest = If_Cut_t{nLeaves: 2, pLeaves: []int32{a.Id, b.Id}}
			if useTruth {
				id := vec.Vec_MemHashInsert(p.vTtMem[2], []uint64{0x2222222222222222})
				root.CutBest.iCutFunc = int32(id * 2)
			}
			If_ManCreateCo(p, If_Not(root))
			result := Gia_ManFromIfLogic(p)
			if result == nil {
				t.Fatal("mapping failed")
			}
			defer Gia_ManStop(result)
			if Gia_ManCiNum(result) != 2 || Gia_ManCoNum(result) != 1 || gia.Gia_ManHasMapping(result) == 0 {
				t.Fatal("interface or mapping missing")
			}
			co := Gia_ManCo(result, 0)
			rootID := gia.Gia_ObjFaninId0p(result, co)
			if gia.Gia_ObjIsLut(result, rootID) == 0 {
				t.Fatal("output driver has no LUT")
			}
			leaves := gia.Gia_ObjLutFanins(result, rootID)
			if len(leaves) != 2 {
				t.Fatalf("LUT leaves %v", leaves)
			}
			for x := 0; x < 4; x++ {
				vals := make([]bool, Gia_ManObjNum(result))
				for i := int32(0); i < Gia_ManCiNum(result); i++ {
					vals[gia.Gia_ObjId(result, Gia_ManCi(result, i))] = x>>uint(i)&1 != 0
				}
				for id := int32(1); id < Gia_ManObjNum(result); id++ {
					obj := Gia_ManObj(result, id)
					if gia.Gia_ObjIsAnd(obj) == 0 {
						continue
					}
					a := vals[gia.Gia_ObjFaninId0(obj, id)] != (gia.Gia_ObjFaninC0(obj) != 0)
					b := vals[gia.Gia_ObjFaninId1(obj, id)] != (gia.Gia_ObjFaninC1(obj) != 0)
					vals[id] = a && b
				}
				got := vals[rootID] != (gia.Gia_ObjFaninC0(co) != 0)
				if want := !(x&1 != 0 && x&2 == 0); got != want {
					t.Fatalf("assignment %d got %v want %v", x, got, want)
				}
			}
		})
	}
}

func TestGiaFromIfLogicCellConfiguration(t *testing.T) {
	manager := If_DsdManAlloc(2, 2)
	defer If_DsdManFree(manager, 0)
	truth := []word{0x8888888888888888}
	pins := make([]byte, 2)
	dsd := If_DsdManCompute(manager, truth, 2, pins, "")
	Id_DsdManTuneStr1(manager, "{ab}", 0, 0)
	pars := &If_Par_t_{NLutSize: 4, NCutsMax: 4, FTruth: 1, FUseDsd: 1, FUseDsdTune: 1, FDeriveLuts: 1}
	p := If_ManStart(pars)
	defer If_ManStop(p)
	p.pIfDsdMan = manager
	p.pName = filepath.Join(t.TempDir(), "mapped.aig")
	a, b := If_ManCreateCi(p), If_ManCreateCi(p)
	root := If_ManCreateAnd(p, a, b)
	truthID := vec.Vec_MemHashInsert(p.vTtMem[2], truth)
	root.CutBest = If_Cut_t{nLeaves: 2, pLeaves: []int32{a.Id, b.Id}, iCutFunc: int32(truthID * 2)}
	vec.Vec_IntSetEntry(p.vTtDsds[2], truthID, dsd)
	vec.Vec_StrFillExtra(p.vTtPerms[2], (truthID+1)*6, 0)
	for i, pin := range pins {
		vec.Vec_StrWriteEntry(p.vTtPerms[2], truthID*6+i, pin)
	}
	If_ManCreateCo(p, root)
	result := Gia_ManFromIfLogic(p)
	if result == nil {
		t.Fatal("cell mapping failed")
	}
	defer Gia_ManStop(result)
	cell, configs := result.CellConfiguration()
	if cell != "{ab}" || configs == nil || Vec_IntSize(configs) != 3 || Vec_IntEntry(configs, 0) != 1 {
		t.Fatalf("cell=%q configs=%v", cell, configs)
	}
	file := strings.TrimSuffix(p.pName, ".aig") + "_configs.txt"
	if data, err := os.ReadFile(file); err != nil || len(data) == 0 {
		t.Fatalf("configuration report %q: %v", data, err)
	}
	cellNetwork := Ifn_NtkParse(cell)
	Ifn_Prepare(cellNetwork, nil, 2)
	reconstructed := Gia_ManStart(12)
	leaves, cover := Vec_IntAlloc(2), Vec_IntAlloc(8)
	Vec_IntPush(leaves, Gia_ManAppendCi(reconstructed))
	Vec_IntPush(leaves, Gia_ManAppendCi(reconstructed))
	Gia_ManHashStart(reconstructed)
	lit := If_ManSatDeriveGiaFromBits(reconstructed, cellNetwork, ifnConfigWords(configs, 2, 1), leaves, cover)
	Gia_ManHashStop(reconstructed)
	for x := 0; x < 4; x++ {
		vals := make([]bool, Gia_ManObjNum(reconstructed))
		for i, input := range Vec_IntArray(leaves) {
			vals[input>>1] = x>>uint(i)&1 != 0
		}
		for id := int32(1); id < Gia_ManObjNum(reconstructed); id++ {
			obj := Gia_ManObj(reconstructed, id)
			if gia.Gia_ObjIsAnd(obj) == 0 {
				continue
			}
			a := vals[gia.Gia_ObjFaninId0(obj, id)] != (gia.Gia_ObjFaninC0(obj) != 0)
			b := vals[gia.Gia_ObjFaninId1(obj, id)] != (gia.Gia_ObjFaninC1(obj) != 0)
			vals[id] = a && b
		}
		if got, want := vals[lit>>1] != (lit&1 != 0), x == 3; got != want {
			t.Fatalf("assignment %d got %v want %v", x, got, want)
		}
	}
}

func TestTruthCacheSizesForSmallLuts(t *testing.T) {
	for size := int32(1); size <= 6; size++ {
		p := If_ManStart(&If_Par_t_{NLutSize: size, NCutsMax: 4, FTruth: 1})
		for n := int32(0); n <= size; n++ {
			if p.nTruth6Words[n] != 1 || len(p.puTempW) != 1 {
				t.Fatalf("LUT size %d, support %d: words %d temporary %d", size, n, p.nTruth6Words[n], len(p.puTempW))
			}
		}
		If_ManStop(p)
	}
}
