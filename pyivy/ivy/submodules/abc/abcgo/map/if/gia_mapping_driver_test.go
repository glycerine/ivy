package if_

import (
	"github.com/glycerine/abcgo/aig/gia"
	"testing"
)

func TestGiaMappingDriverSimpleNetwork(t *testing.T) {
	for _, hash := range []int32{0, 1} {
		t.Run(map[int32]string{0: "unhashed", 1: "hashed"}[hash], func(t *testing.T) {
			src := Gia_ManStart(12)
			src.SetIdentity("simple", "reference")
			a, b, c := Gia_ManAppendCi(src), Gia_ManAppendCi(src), Gia_ManAppendCi(src)
			x := gia.Gia_ManAppendAnd(src, a, b^1)
			y := gia.Gia_ManAppendAnd(src, x, c)
			Gia_ManAppendCo(src, y^1)
			var pars If_Par_t
			Gia_ManSetIfParsDefault(&pars)
			pars.NLutSize = 4
			pars.FHashMapping = hash
			result := Gia_ManPerformMapping(src, &pars)
			if result == nil {
				t.Fatal("mapping failed")
			}
			defer Gia_ManStop(result)
			if gia.Gia_ManHasMapping(result) == 0 || Gia_ManCiNum(result) != 3 || Gia_ManCoNum(result) != 1 {
				t.Fatal("mapped interface missing")
			}
			if name, spec := result.Identity(); name != "simple" || spec != "reference" {
				t.Fatalf("identity %q %q", name, spec)
			}
			co := Gia_ManCo(result, 0)
			if gia.Gia_ObjIsLut(result, gia.Gia_ObjFaninId0p(result, co)) == 0 {
				t.Fatal("output driver is not mapped")
			}
			for mask := 0; mask < 8; mask++ {
				values := make([]bool, Gia_ManObjNum(result))
				for i := int32(0); i < 3; i++ {
					values[gia.Gia_ObjId(result, Gia_ManCi(result, i))] = mask>>uint(i)&1 != 0
				}
				for id := int32(1); id < Gia_ManObjNum(result); id++ {
					obj := Gia_ManObj(result, id)
					if gia.Gia_ObjIsAnd(obj) == 0 {
						continue
					}
					a := values[gia.Gia_ObjFaninId0(obj, id)] != (gia.Gia_ObjFaninC0(obj) != 0)
					b := values[gia.Gia_ObjFaninId1(obj, id)] != (gia.Gia_ObjFaninC1(obj) != 0)
					values[id] = a && b
				}
				got := values[gia.Gia_ObjFaninId0p(result, co)] != (gia.Gia_ObjFaninC0(co) != 0)
				want := !(mask&1 != 0 && mask&2 == 0 && mask&4 != 0)
				if got != want {
					t.Fatalf("assignment %d got %v want %v", mask, got, want)
				}
			}
		})
	}
}

func TestGiaBalancingDriversPreserveTruth(t *testing.T) {
	for _, kind := range []string{"sop", "dsd", "registered-sop", "registered-dsd"} {
		t.Run(kind, func(t *testing.T) {
			src := Gia_ManStart(12)
			a, b, c := Gia_ManAppendCi(src), Gia_ManAppendCi(src), Gia_ManAppendCi(src)
			x := gia.Gia_ManAppendAnd(src, a, b^1)
			y := gia.Gia_ManAppendAnd(src, x, c)
			Gia_ManAppendCo(src, y^1)
			var result *Gia_Man_t
			switch kind {
			case "sop":
				result = Gia_ManPerformSopBalance(src, 8, 0, 0)
			case "dsd":
				result = Gia_ManPerformDsdBalance(src, 6, 8, 0, 0)
			case "registered-sop":
				result = gia.Gia_ManPerformSopBalance(src, 8, 0, 0)
			case "registered-dsd":
				result = gia.Gia_ManPerformDsdBalance(src, 6, 8, 0, 0)
			}
			if result == nil {
				t.Fatal("balancing failed")
			}
			defer Gia_ManStop(result)
			co := Gia_ManCo(result, 0)
			for mask := 0; mask < 8; mask++ {
				values := make([]bool, Gia_ManObjNum(result))
				for i := int32(0); i < 3; i++ {
					values[gia.Gia_ObjId(result, Gia_ManCi(result, i))] = mask>>uint(i)&1 != 0
				}
				for id := int32(1); id < Gia_ManObjNum(result); id++ {
					obj := Gia_ManObj(result, id)
					if gia.Gia_ObjIsAnd(obj) == 0 {
						continue
					}
					x := values[gia.Gia_ObjFaninId0(obj, id)] != (gia.Gia_ObjFaninC0(obj) != 0)
					y := values[gia.Gia_ObjFaninId1(obj, id)] != (gia.Gia_ObjFaninC1(obj) != 0)
					values[id] = x && y
				}
				got := values[gia.Gia_ObjFaninId0p(result, co)] != (gia.Gia_ObjFaninC0(co) != 0)
				want := !(mask&1 != 0 && mask&2 == 0 && mask&4 != 0)
				if got != want {
					t.Fatalf("assignment %d got %v want %v", mask, got, want)
				}
			}
		})
	}
}
