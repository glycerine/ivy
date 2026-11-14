package abci

import (
	"slices"
	"testing"

	"github.com/glycerine/abcgo/aig/hop"
	"github.com/glycerine/abcgo/bool/bdc"
	"github.com/glycerine/abcgo/misc/vec"
)

func TestAbcNodeIfNodeResynPreservesTruth(t *testing.T) {
	h := hop.Hop_ManStart()
	defer hop.Hop_ManStop(h)

	x, y := hop.Hop_IthVar(h, 0), hop.Hop_IthVar(h, 1)
	root := hop.Hop_And(h, x, hop.Hop_Not(y))
	for _, input := range []*hop.Hop_Obj_t{root, hop.Hop_Not(root)} {
		want := append([]uint32(nil), hop.Hop_ManConvertAigToTruth(h, input, 2, nil, 0)...)
		manager := bdc.Bdc_ManAlloc(&bdc.Bdc_Par_t{NVarsMax: 2})
		scratch := vec.Vec_IntAlloc(0)
		gotRoot := Abc_NodeIfNodeResyn(manager, h, input, 2, scratch, nil, -1)
		got := hop.Hop_ManConvertAigToTruth(h, gotRoot, 2, scratch, 0)
		if !slices.Equal(got, want) {
			t.Fatalf("resynthesis truth = %08x, want %08x", got, want)
		}
		vec.Vec_IntFree(scratch)
		bdc.Bdc_ManFree(manager)
	}
}

func TestAbcNodeIfNodeResynPowerCareRestored(t *testing.T) {
	h := hop.Hop_ManStart()
	defer hop.Hop_ManStop(h)
	x, y := hop.Hop_IthVar(h, 0), hop.Hop_IthVar(h, 1)
	root := hop.Hop_And(h, x, y)
	care := []uint32{0}
	manager := bdc.Bdc_ManAlloc(&bdc.Bdc_Par_t{NVarsMax: 2})
	scratch := vec.Vec_IntAlloc(0)
	got := Abc_NodeIfNodeResyn(manager, h, root, 2, scratch, care, 0.5)
	if care[0] != 0 {
		t.Fatalf("care truth was not restored after complement: %#x", care[0])
	}
	want := []uint32{0}
	if truth := hop.Hop_ManConvertAigToTruth(h, got, 2, scratch, 0); !slices.Equal(truth, want) {
		t.Fatalf("power-aware truth = %08x, want %08x", truth, want)
	}
	vec.Vec_IntFree(scratch)
	bdc.Bdc_ManFree(manager)
}
