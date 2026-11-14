package abci

import (
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/misc/vec"
)

func TestDress2DchParametersMatchCDefaults(t *testing.T) {
	var pars Dch_Pars_t
	Dch_ManSetDefaultParams(&pars)
	if pars.nWords != 8 || pars.nBTLimit != 1000 || pars.nSatVarMax != 5000 ||
		pars.fSynthesis != 1 || pars.fPolarFlip != 1 || pars.fSimulateTfo != 1 ||
		pars.nNodesAhead != 1000 || pars.nCallsRecycle != 100 {
		t.Fatalf("DCH defaults differ from abcDress2.c/dchCore.c: %+v", pars)
	}
}

func TestDress2EquivIDPacking(t *testing.T) {
	for _, tc := range []struct {
		id, object, phase, network int32
	}{{0x2d, 11, 0, 1}, {0x36, 13, 1, 0}, {0x7f, 31, 1, 1}} {
		if got := Abc_ObjEquivId2ObjId(tc.id); got != tc.object {
			t.Fatalf("object id %#x = %d, want %d", tc.id, got, tc.object)
		}
		if got := Abc_ObjEquivId2Polar(tc.id); got != tc.phase {
			t.Fatalf("phase %#x = %d, want %d", tc.id, got, tc.phase)
		}
		if got := Abc_ObjEquivId2NtkId(tc.id); got != tc.network {
			t.Fatalf("network %#x = %d, want %d", tc.id, got, tc.network)
		}
	}
}

func TestDress2DualOutputMiterSharesInputsAndPairsOutputs(t *testing.T) {
	first, second := aig.NewGraph(), aig.NewGraph()
	a, b := first.NewInput(), first.NewInput()
	c, d := second.NewInput(), second.NewInput()
	first.AddOutput(first.AND(a, b).Not())
	first.AddOutput(a)
	second.AddOutput(second.AND(c, d))
	second.AddOutput(d.Not())

	miter := Aig_ManCreateDualOutputMiter(first, second)
	if got := aig.Aig_ManCiNum(miter); got != 2 {
		t.Fatalf("miter CI count = %d, want shared inputs 2", got)
	}
	if got := aig.Aig_ManCoNum(miter); got != 4 {
		t.Fatalf("miter CO count = %d, want 4", got)
	}
	for mask := 0; mask < 4; mask++ {
		inputs := []bool{mask&1 != 0, mask&2 != 0}
		got, err := miter.Evaluate(inputs)
		if err != nil {
			t.Fatal(err)
		}
		want := []bool{!(inputs[0] && inputs[1]), inputs[0] && inputs[1], inputs[0], !inputs[1]}
		for i := range want {
			if got[i] != want[i] {
				t.Fatalf("assignment %02b output %d = %t, want %t", mask, i, got[i], want[i])
			}
		}
	}
}

func TestDress2MapClassesUsesAigCopyAndRepresentativeChain(t *testing.T) {
	first, second := aig.NewGraph(), aig.NewGraph()
	a, b := first.NewInput(), first.NewInput()
	c, d := second.NewInput(), second.NewInput()
	root := first.AND(a, b)
	first.AddOutput(root)
	second.AddOutput(second.AND(c, d))
	miter := Aig_ManCreateDualOutputMiter(first, second)
	rootInMiter := aig.Aig_ObjCopy(first, root)
	reprInMiter := aig.Aig_ObjCopy(first, a)
	aig.Aig_ObjSetRepr(miter, rootInMiter, reprInMiter)

	network := abc.NewNetwork("original")
	obj := network.NewObject(abc.ObjectNode)
	strash := abc.NewNetwork("strash")
	strashedObj := strash.NewObject(abc.ObjectNode)
	strashedObj.Temp = int(root)
	strashedObj.Phase = true
	obj.CopyEdge = strashedObj.Signal()

	classes := Abc_NtkDressMapClasses(miter, network, first)
	if got, want := vec.Vec_IntEntry(classes, obj.ID), int32(reprInMiter.NodeID()); got != want {
		t.Fatalf("mapped representative = %d, want %d", got, want)
	}
	Abc_NtkDressMapSetPolarity(network)
	if !obj.Phase {
		t.Fatal("polarity mapping did not transfer the strashed phase")
	}
}

func TestDress2EquivalenceSweepFindsAssociativeAnd(t *testing.T) {
	first, second := aig.NewGraph(), aig.NewGraph()
	a, b, c := first.NewInput(), first.NewInput(), first.NewInput()
	x, y, z := second.NewInput(), second.NewInput(), second.NewInput()
	root1 := first.AND(first.AND(a, b), c)
	root2 := second.AND(x, second.AND(y, z))
	first.AddOutput(root1)
	second.AddOutput(root2)
	miter := Aig_ManCreateDualOutputMiter(first, second)
	root1 = aig.Aig_ObjCopy(first, root1)
	root2 = aig.Aig_ObjCopy(second, root2)
	Dch_ComputeEquivalences2(miter, &Dch_Pars_t{nBTLimit: 1000})
	if got := aig.Aig_ObjRepr(miter, root2); got != root1 {
		t.Fatalf("associative AND representative = %v, want %v", got, root1)
	}
}

func TestDress2TransfersNamesAcrossPolarityAndSkipsNamedTargets(t *testing.T) {
	n0, n1 := abc.NewNetwork("before"), abc.NewNetwork("after")
	targetDirect := n0.NewObject(abc.ObjectNode)
	targetInverse := n0.NewObject(abc.ObjectNode)
	targetAlreadyNamed := n0.NewObject(abc.ObjectNode)
	sourceDirect := n1.NewObject(abc.ObjectNode)
	sourceInverse := n1.NewObject(abc.ObjectNode)
	sourceAlreadyNamed := n1.NewObject(abc.ObjectNode)
	targetAlreadyNamed.SetName("keep")
	sourceDirect.SetName("direct_name")
	sourceInverse.SetName("inverted_name")
	sourceAlreadyNamed.SetName("unused")

	classes := vec.Vec_PtrAlloc(3)
	class := func(values ...int32) {
		v := vec.Vec_IntAlloc(len(values))
		for _, value := range values {
			vec.Vec_IntPush(v, value)
		}
		vec.Vec_PtrPush(classes, v)
	}
	class((int32(targetDirect.ID) << 2), (int32(sourceDirect.ID)<<2)|1)
	class((int32(targetInverse.ID) << 2), (int32(sourceInverse.ID)<<2)|3)
	class((int32(targetAlreadyNamed.ID) << 2), (int32(sourceAlreadyNamed.ID)<<2)|1)

	Abc_NtkDress2Transfer(n0, n1, classes, 0)
	if got := targetDirect.Name(); got != "direct_name" {
		t.Fatalf("direct target name = %q", got)
	}
	if got := targetInverse.Name(); got != "inverted_name_inv" {
		t.Fatalf("inverse target name = %q", got)
	}
	if got := targetAlreadyNamed.Name(); got != "keep" {
		t.Fatalf("pre-named target changed to %q", got)
	}
}
