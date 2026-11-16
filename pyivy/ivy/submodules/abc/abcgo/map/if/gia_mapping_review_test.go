package if_

import (
	"fmt"
	"slices"
	"testing"

	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
)

func mappingReviewManager(t testing.TB, mode string, phase int32) (*If_Man_t, *If_Obj_t, []*If_Obj_t, int) {
	t.Helper()
	pars := &If_Par_t{NLutSize: 4, NCutsMax: 4}
	if mode != "structural" {
		pars.FTruth = 1
		pars.FDeriveLuts = 1
	}
	mask := 0
	switch mode {
	case "permuted", "permuted_cache":
		pars.FUseTtPerm = 1
		mask = 5
		if mode == "permuted_cache" {
			pars.FDeriveLuts = 0
		}
	case "structure":
		pars.PLutStruct = "33"
	case "check44":
		pars.FEnableCheck07 = 1
	case "check75":
		pars.FEnableCheck75 = 1
	case "dsd_structure", "dsd_undivided":
		pars.FUseDsd = 1
		pars.PLutStruct = "33"
		if mode == "dsd_undivided" {
			pars.FDeriveLuts = 0
		}
	case "truth_cache_only":
		pars.FDeriveLuts = 0
	case "dsd_truth":
		pars.FUseDsd = 1
		pars.FDeriveLuts = 0
	case "cell_trigger":
		pars.FDeriveLuts = 0
		pars.pFuncCell2 = func(*If_Man_t, []uint64, int32, *Vec_Str_t, *[]string) int32 {
			t.Fatal("reconstruction invoked cost callback")
			return 0
		}
	case "and":
		pars.FUseAndVars = 1
	case "cofactor":
		pars.FUseCofVars = 1
	case "both_and", "both_cofactor":
		pars.FUseAndVars = 1
		pars.FUseCofVars = 1
	}
	p := If_ManStart(pars)
	t.Cleanup(func() { If_ManStop(p) })
	inputs := make([]*If_Obj_t, 4)
	literals := make([]*If_Obj_t, 4)
	ids := make([]int32, 4)
	for pin := range inputs {
		inputs[pin] = If_ManCreateCi(p)
		ids[pin] = inputs[pin].Id
		literals[pin] = inputs[pin]
		if mask>>pin&1 != 0 {
			literals[pin] = If_Not(literals[pin])
		}
	}
	left := If_ManCreateAnd(p, literals[0], literals[1])
	right := If_ManCreateAnd(p, literals[2], literals[3])
	root := If_ManCreateAnd(p, left, right)
	dead := If_ManCreateAnd(p, inputs[0], If_Not(inputs[3]))
	skipped := []*If_Obj_t{left, right, dead}
	// Mapping references are defined by the selected four-input cut, which
	// bypasses both structural children and the unrelated dangling node.
	for _, obj := range skipped {
		obj.nRefs = 0
		obj.iCopy = -99
	}
	root.CutBest = If_Cut_t{nLimit: 4, nLeaves: 4, pLeaves: ids, uMaskFunc: int32(mask)}
	if pars.FTruth != 0 {
		truth := []word{0x8000800080008000}
		id := vec.Vec_MemHashInsert(p.vTtMem[4], truth)
		root.CutBest.iCutFunc = int32(2*id) | phase
		root.CutBest.fCompl = uint32(phase)
		if pars.FUseAndVars != 0 {
			packed := int32(3 | 12<<16)
			if mode == "both_cofactor" {
				packed = 0
			}
			vec.Vec_IntFillExtra(p.vTtDecs[4], id+1, 0)
			vec.Vec_IntWriteEntry(p.vTtDecs[4], id, packed)
		}
		if pars.FUseCofVars != 0 {
			vec.Vec_StrFillExtra(p.vTtVars[4], id+1, 0)
			vec.Vec_StrWriteEntry(p.vTtVars[4], id, 1)
		}
		if pars.FUseDsd != 0 {
			dsd := If_DsdManAlloc(4, 3)
			t.Cleanup(func() { If_DsdManFree(dsd, 0) })
			p.pIfDsdMan = dsd
			perm := make([]byte, 4)
			lit := If_DsdManCompute(dsd, truth, 4, perm, "")
			vec.Vec_IntFillExtra(p.vTtDsds[4], id+1, -1)
			vec.Vec_IntWriteEntry(p.vTtDsds[4], id, lit)
			vec.Vec_StrFillExtra(p.vTtPerms[4], (id+1)*6, 0)
			for i, pin := range perm {
				vec.Vec_StrWriteEntry(p.vTtPerms[4], id*6+i, pin)
			}
		}
	}
	for _, obj := range []*If_Obj_t{root, If_Not(root), If_ManConst1(p), If_Not(If_ManConst1(p)), inputs[0], If_Not(inputs[0])} {
		If_ManCreateCo(p, obj)
	}
	return p, root, skipped, mask
}

func TestMappingLogicDriverBranchesAndMetadata(t *testing.T) {
	for _, mode := range []string{"structural", "truth", "truth_cache_only", "permuted", "permuted_cache", "dsd_truth", "cell_trigger", "structure", "check44", "check75", "dsd_structure", "dsd_undivided", "and", "cofactor", "both_and", "both_cofactor"} {
		for phase := int32(0); phase < 2; phase++ {
			if mode == "structural" && phase != 0 {
				continue
			}
			t.Run(fmt.Sprintf("%s/phase%d", mode, phase), func(t *testing.T) {
				p, root, skipped, mask := mappingReviewManager(t, mode, phase)
				first := Gia_ManFromIfLogic(p)
				if first == nil {
					t.Fatal("reconstruction failed")
				}
				defer Gia_ManStop(first)
				before := slices.Clone(vec.Vec_IntArray(first.VMapping))
				// Repeated reconstruction resets cut scratch and owns independent data.
				second := Gia_ManFromIfLogic(p)
				if second == nil {
					t.Fatal("repeated reconstruction failed")
				}
				defer Gia_ManStop(second)
				if first.VMapping == second.VMapping || !slices.Equal(before, vec.Vec_IntArray(first.VMapping)) {
					t.Fatal("reconstruction reused prior mapping ownership")
				}
				for _, obj := range skipped {
					if obj.iCopy != -99 {
						t.Fatalf("unused node%d reconstructed", obj.Id)
					}
				}
				for _, dst := range []*Gia_Man_t{first, second} {
					if Gia_ManCiNum(dst) != 4 || Gia_ManCoNum(dst) != 6 {
						t.Fatal("interface changed")
					}
					data := vec.Vec_IntArray(dst.VMapping)
					lutCount := 0
					for id := int32(0); id < Gia_ManObjNum(dst); id++ {
						obj := Gia_ManObj(dst, id)
						if gia.Gia_ObjIsCi(obj) != 0 && data[id] != 0 {
							t.Fatalf("CI%d has mapping", id)
						}
						if gia.Gia_ObjIsAnd(obj) == 0 || data[id] == 0 {
							continue
						}
						lutCount++
						at := data[id]
						count := data[at]
						if at < Gia_ManObjNum(dst) || count < 2 || at+count+1 >= int32(len(data)) {
							t.Fatalf("bad LUT%d offset%d", id, at)
						}
						for _, fan := range data[at+1 : at+count+1] {
							if fan < 0 || fan >= id {
								t.Fatalf("non-topological LUT%d fan%d", id, fan)
							}
						}
					}
					wantCount := 1
					switch mode {
					case "structure", "dsd_structure", "cofactor", "both_cofactor":
						wantCount = 2
					case "and", "both_and":
						wantCount = 3
					}
					if lutCount != wantCount {
						t.Fatalf("LUT count%d want%d", lutCount, wantCount)
					}
					rootID := root.iCopy >> 1
					marker := data[data[rootID]+data[data[rootID]]+1]
					if mode == "and" || mode == "both_and" {
						if marker != -rootID {
							t.Fatalf("AND marker%d want%d", marker, -rootID)
						}
					} else if mode == "cofactor" || mode == "both_cofactor" {
						if marker != -2 {
							t.Fatalf("cofactor marker%d want selector ID2", marker)
						}
					} else if marker != rootID {
						t.Fatalf("ordinary root marker%d want%d", marker, rootID)
					}
					var packed []int32
					if dst.VPacking != nil {
						packed = vec.Vec_IntArray(dst.VPacking)
					}
					switch mode {
					case "structure", "dsd_structure":
						if len(packed) != 4 || packed[0] != 1 || packed[1] != 2 || packed[3] != rootID {
							t.Fatalf("packing%v", packed)
						}
					case "check44", "check75":
						if !slices.Equal(packed, []int32{1, 1, rootID}) {
							t.Fatalf("packing%v", packed)
						}
					default:
						if dst.VPacking != nil {
							t.Fatalf("unexpected packing%v", packed)
						}
					}
					for x := 0; x < 16; x++ {
						input := make([]bool, 4)
						for pin := range input {
							input[pin] = x>>pin&1 != 0
						}
						expected := x^mask == 15
						want := []bool{expected, !expected, true, false, input[0], !input[0]}
						got, err := dst.Evaluate(input)
						if err != nil || !slices.Equal(got, want) {
							t.Fatalf("assignment%d got%v want%v err%v", x, got, want, err)
						}
					}
				}
			})
		}
	}
}

func BenchmarkMappingLogicDriver(b *testing.B) {
	for _, mode := range []string{"structural", "truth"} {
		b.Run(mode, func(b *testing.B) {
			p, _, _, _ := mappingReviewManager(b, mode, 0)
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				dst := Gia_ManFromIfLogic(p)
				Gia_ManStop(dst)
			}
		})
	}
}

func TestMappingLogicDriverEmptyInterfaces(t *testing.T) {
	for _, n := range []int{0, 3} {
		for _, outputs := range []bool{false, true} {
			t.Run(fmt.Sprintf("inputs%d/outputs%t", n, outputs), func(t *testing.T) {
				p := If_ManStart(&If_Par_t{NLutSize: 4, NCutsMax: 4})
				defer If_ManStop(p)
				inputs := make([]*If_Obj_t, n)
				for i := range inputs {
					inputs[i] = If_ManCreateCi(p)
				}
				if outputs {
					If_ManCreateCo(p, If_ManConst1(p))
					If_ManCreateCo(p, If_Not(If_ManConst1(p)))
					if n != 0 {
						If_ManCreateCo(p, inputs[0])
						If_ManCreateCo(p, If_Not(inputs[0]))
					}
				}
				dst := Gia_ManFromIfLogic(p)
				if dst == nil {
					t.Fatal("empty/terminal reconstruction failed")
				}
				defer Gia_ManStop(dst)
				count := Gia_ManObjNum(dst)
				data := vec.Vec_IntArray(dst.VMapping)
				suffix := []int32{0}
				offset := int32(0)
				if outputs {
					suffix = []int32{0, 0, 0}
					offset = count + 1
				}
				// Source skips an unreferenced constant: it is not a CI/CO
				// terminal. Referenced constants get the extra LUT record.
				if len(data) != int(count)+len(suffix) || data[0] != offset || !slices.Equal(data[count:], suffix) {
					t.Fatalf("source constant record%v nodes%d", data, count)
				}
				for i := int32(1); i < count; i++ {
					if data[i] != 0 {
						t.Fatalf("terminal%d has a LUT record", i)
					}
				}
				for x := 0; x < 1<<n; x++ {
					input := make([]bool, n)
					for pin := range input {
						input[pin] = x>>pin&1 != 0
					}
					var want []bool
					if outputs {
						want = []bool{true, false}
						if n != 0 {
							want = append(want, input[0], !input[0])
						}
					}
					got, err := dst.Evaluate(input)
					if err != nil || !slices.Equal(got, want) {
						t.Fatalf("assignment%d got%v want%v err%v", x, got, want, err)
					}
				}
			})
		}
	}
}

func TestMappingLogicDriverFailureThenRetry(t *testing.T) {
	p, root, skipped, _ := mappingReviewManager(t, "truth", 0)
	p.pPars.PLutStruct = "3"
	output := dsdReportCapture(t, func() {
		if got := Gia_ManFromIfLogic(p); got != nil {
			Gia_ManStop(got)
			t.Fatal("invalid structure returned a partial graph")
		}
	})
	if output != "Wrong LUT struct (3)\n" || root.iCopy != -1 {
		t.Fatalf("failure output%q copy%d", output, root.iCopy)
	}
	p.pPars.PLutStruct = ""
	dst := Gia_ManFromIfLogic(p)
	if dst == nil {
		t.Fatal("valid reconstruction failed after prior error")
	}
	defer Gia_ManStop(dst)
	for _, obj := range skipped {
		if obj.iCopy != -99 {
			t.Fatalf("unused node%d changed", obj.Id)
		}
	}
	for x := 0; x < 16; x++ {
		input := make([]bool, 4)
		for pin := range input {
			input[pin] = x>>pin&1 != 0
		}
		got, err := dst.Evaluate(input)
		want := []bool{x == 15, x != 15, true, false, input[0], !input[0]}
		if err != nil || !slices.Equal(got, want) {
			t.Fatalf("assignment%d got%v want%v err%v", x, got, want, err)
		}
	}
}
