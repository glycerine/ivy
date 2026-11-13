package live

import (
	"github.com/glycerine/abcgo/aig/aig"
	"slices"
	"testing"
)

func TestSourceAbstractionShadowChoices(t *testing.T) {
	p := aig.Aig_ManStart(20)
	in, s0, s1 := aig.Aig_ObjCreateCi(p), aig.Aig_ObjCreateCi(p), aig.Aig_ObjCreateCi(p)
	aig.Aig_ObjCreateCo(p, s0)
	aig.Aig_ObjCreateCo(p, s1)
	aig.Aig_ObjCreateCo(p, in)
	aig.Aig_ObjCreateCo(p, s0)
	aig.Aig_ManSetRegNum(p, 2)
	n := &Network{Graph: p}
	live, fair := []aig.Edge{p.Outputs()[0]}, []aig.Edge{p.Outputs()[1]}
	full := LivenessToSafetyTransformation(FULL_BIERE_MODE, n, p, live, fair, nil, nil)
	both := LivenessToSafetyTransformationAbs(FULL_BIERE_MODE, n, p, []int32{0, 1}, live, fair, nil, nil)
	none := LivenessToSafetyTransformationAbs(FULL_BIERE_MODE, n, p, []int32{}, live, fair, nil, nil)
	if registers(full) != registers(both) || registers(full)-registers(none) != 2 {
		t.Fatalf("full=%d both=%d none=%d registers", registers(full), registers(both), registers(none))
	}
	a, err := full.AigerModel()
	if err != nil {
		t.Fatal(err)
	}
	b, err := both.AigerModel()
	if err != nil {
		t.Fatal(err)
	}
	for bits := 0; bits < 256; bits++ {
		if !slices.Equal(evalAiger(a, bits, 4), evalAiger(b, bits, 4)) {
			t.Fatalf("full abstraction changed trace %d", bits)
		}
	}
}
