package if_

import (
	"reflect"
	"testing"

	"github.com/glycerine/abcgo/aig/gia"
	abcmain "github.com/glycerine/abcgo/base/main"
)

func TestGiaBalancingRetainsConfigurationAndTimingMetadata(t *testing.T) {
	for _, kind := range []string{"sop", "dsd", "registered-sop", "registered-dsd"} {
		t.Run(kind, func(t *testing.T) {
			src := Gia_ManStart(16)
			defer Gia_ManStop(src)
			src.SetIdentity("balance-review", "source-spec")
			a, b, c := Gia_ManAppendCi(src), Gia_ManAppendCi(src), Gia_ManAppendCi(src)
			x := gia.Gia_ManAppendAnd(src, a, b^1)
			root := gia.Gia_ManAppendAnd(src, x, c)
			Gia_ManAppendCo(src, root^1)
			configs := gia.Vec_IntAllocArrayCopy([]int32{0, 1}, 2)
			src.VConfigs = configs
			src.PCellStr = "retained-cell"
			frame := abcmain.NewFrame()
			frame.SetResource("pGia", src)
			gia.Abc_FrameMiniAigSetCiArrivals(frame, []int32{7, 11, 13})
			gia.Abc_FrameMiniAigSetCoRequireds(frame, []int32{30})
			src.SetMappingOutputArrivals([]int32{91})
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
			if result.VConfigs != configs || result.PCellStr != "retained-cell" {
				t.Errorf("balancing discarded configuration metadata: configs%v cell%q", result.VConfigs, result.PCellStr)
			}
			if src.VConfigs != nil || src.PCellStr != "" {
				t.Error("configuration ownership did not move")
			}
			arrivals, required, _, _ := result.MappingInputs()
			if !reflect.DeepEqual(arrivals, []int32{7, 11, 13}) || !reflect.DeepEqual(required, []int32{30}) {
				t.Error("input timing metadata changed")
			}
			frame.SetResource("pGia", result)
			if got := gia.Abc_FrameMiniAigReadCoArrivals(frame); !reflect.DeepEqual(got, []int32{91}) {
				t.Errorf("balancing rewrote existing output arrivals: %v want[91]", got)
			}
			if name, spec := result.Identity(); name != "balance-review" || spec != "source-spec" {
				t.Error("balancing changed identity")
			}
			for mask := 0; mask < 8; mask++ {
				values, err := result.Evaluate([]bool{mask&1 != 0, mask&2 != 0, mask&4 != 0})
				want := !(mask&1 != 0 && mask&2 == 0 && mask&4 != 0)
				if err != nil || values[0] != want {
					t.Fatalf("assignment%d result%v want%v err%v", mask, values, want, err)
				}
			}
		})
	}
}

func TestGiaGeneralMappingStillReplacesConfigurationAndArrivalData(t *testing.T) {
	src := Gia_ManStart(16)
	defer Gia_ManStop(src)
	src.SetIdentity("general-map-review", "")
	a, b, c := Gia_ManAppendCi(src), Gia_ManAppendCi(src), Gia_ManAppendCi(src)
	x := gia.Gia_ManAppendAnd(src, a, b^1)
	root := gia.Gia_ManAppendAnd(src, x, c)
	Gia_ManAppendCo(src, root^1)
	src.VConfigs = gia.Vec_IntAllocArrayCopy([]int32{0, 1}, 2)
	src.PCellStr = "old-cell"
	frame := abcmain.NewFrame()
	frame.SetResource("pGia", src)
	gia.Abc_FrameMiniAigSetCiArrivals(frame, []int32{7, 11, 13})
	gia.Abc_FrameMiniAigSetCoRequireds(frame, []int32{30})
	src.SetMappingOutputArrivals([]int32{91})
	var pars If_Par_t
	Gia_ManSetIfParsDefault(&pars)
	pars.NLutSize = 4
	result := Gia_ManPerformMapping(src, &pars)
	if result == nil {
		t.Fatal("mapping failed")
	}
	defer Gia_ManStop(result)
	if result.VConfigs != nil || result.PCellStr != "" || src.VConfigs != nil || src.PCellStr != "" {
		t.Fatal("general mapping retained obsolete cell configuration")
	}
	frame.SetResource("pGia", result)
	if got := gia.Abc_FrameMiniAigReadCoArrivals(frame); !reflect.DeepEqual(got, []int32{14}) {
		t.Fatalf("general mapping did not recompute output arrivals: %v", got)
	}
}

func TestGiaBalancingMetadataAndSequentialInterfaces(t *testing.T) {
	for _, dsd := range []bool{false, true} {
		for regs := int32(0); regs < 2; regs++ {
			for timing := 0; timing < 4; timing++ {
				for config := 0; config < 2; config++ {
					src := Gia_ManStart(16)
					src.SetIdentity("balance-matrix", "matrix-spec")
					a, b, c, d := Gia_ManAppendCi(src), Gia_ManAppendCi(src), Gia_ManAppendCi(src), Gia_ManAppendCi(src)
					x := gia.Gia_ManAppendAnd(src, a, b^1)
					root := gia.Gia_ManAppendAnd(src, x, c)
					other := gia.Gia_ManAppendAnd(src, b, d)
					Gia_ManAppendCo(src, root^1)
					Gia_ManAppendCo(src, other)
					Gia_ManSetRegNum(src, regs)
					var configs *Vec_Int_t
					if config != 0 {
						configs = gia.Vec_IntAllocArrayCopy([]int32{0, 1}, 2)
						src.VConfigs = configs
						src.PCellStr = "matrix-cell"
					}
					frame := abcmain.NewFrame()
					frame.SetResource("pGia", src)
					var arrivals, required []int32
					if timing&1 != 0 {
						arrivals = []int32{1, 7, 3, 9}
						gia.Abc_FrameMiniAigSetCiArrivals(frame, arrivals)
					}
					if timing&2 != 0 {
						required = []int32{20, 30}
						gia.Abc_FrameMiniAigSetCoRequireds(frame, required)
					}
					src.SetMappingOutputArrivals([]int32{91, 92})
					var result *Gia_Man_t
					if dsd {
						result = Gia_ManPerformDsdBalance(src, 6, 8, 0, 0)
					} else {
						result = Gia_ManPerformSopBalance(src, 8, 0, 0)
					}
					if result == nil {
						t.Fatal("matrix balancing failed")
					}
					if result.VConfigs != configs || Gia_ManRegNum(result) != regs || Gia_ManCiNum(result) != 4 || Gia_ManCoNum(result) != 2 {
						t.Fatal("configuration ownership or sequential interface changed")
					}
					if config != 0 && result.PCellStr != "matrix-cell" {
						t.Fatal("cell string lost")
					}
					gotArrivals, gotRequired, _, _ := result.MappingInputs()
					if !reflect.DeepEqual(arrivals, gotArrivals) || !reflect.DeepEqual(required, gotRequired) {
						t.Fatal("matrix input timing changed")
					}
					frame.SetResource("pGia", result)
					if got := gia.Abc_FrameMiniAigReadCoArrivals(frame); !reflect.DeepEqual(got, []int32{91, 92}) {
						t.Fatal("matrix output arrivals changed")
					}
					for mask := 0; mask < 16; mask++ {
						values, err := result.Evaluate([]bool{mask&1 != 0, mask&2 != 0, mask&4 != 0, mask&8 != 0})
						if err != nil || values[0] != !(mask&1 != 0 && mask&2 == 0 && mask&4 != 0) || values[1] != (mask&2 != 0 && mask&8 != 0) {
							t.Fatalf("mask%d result%v err%v", mask, values, err)
						}
					}
					Gia_ManStop(result)
					Gia_ManStop(src)
				}
			}
		}
	}
}
