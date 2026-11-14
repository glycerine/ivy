package abci

import (
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/base/abc"
)

func TestAbcNtkDarObjectAndCopyEdgeMapping(t *testing.T) {
	n := abc.Abc_NtkAlloc(abc.ABC_NTK_STRASH, abc.ABC_FUNC_AIG, 1)
	x, y := abc.Abc_NtkCreatePi(n), abc.Abc_NtkCreatePi(n)
	and := abc.Abc_AigAnd(n.FunctionManager.(*abc.Abc_Aig_t), x.Signal(), y.Signal().Not())
	po := abc.Abc_NtkCreatePo(n)
	if err := n.AddFanin(po, and.Not()); err != nil {
		t.Fatal(err)
	}
	g := abc.Abc_NtkToDar(n, 0, 0)
	if g == nil {
		t.Fatal("Abc_NtkToDar returned nil")
	}
	got, ok := Abc_NtkDarObjLit(and.Object)
	if !ok || got != aig.Edge(uint32(and.Object.Temp)) {
		t.Fatalf("direct node mapping = (%v,%t), Temp=%d", got, ok, and.Object.Temp)
	}
	// The AND literal is regular; its complemented output is recorded on the
	// output edge and must not contaminate the node's copy mapping.
	if got.Complemented() {
		t.Fatal("regular AND node was recorded as complemented")
	}
	outputs := g.Outputs()
	if len(outputs) != 1 || outputs[0] != got.Not() {
		t.Fatalf("complemented output mapping = %v, want [%v]", outputs, got.Not())
	}

	original := abc.Abc_NtkCreatePi(n)
	intermediate := abc.Abc_NtkCreateNet(n)
	original.SetCopyEdge(abc.Edge{Object: intermediate, Complement: true})
	intermediate.SetCopyEdge(abc.Edge{Object: and.Object, Complement: true})
	mapped, ok := Abc_NtkDarMappedObjLit(original)
	if !ok || mapped != got {
		t.Fatalf("two-hop copy-edge mapping = (%v,%t), want %v", mapped, ok, got)
	}
	if _, ok := Abc_NtkDarMappedObjLit(nil); ok {
		t.Fatal("nil object unexpectedly mapped")
	}
}

func TestAbcNtkDarFrameOrderAndInitialOnePhase(t *testing.T) {
	n := abc.Abc_NtkAlloc(abc.ABC_NTK_STRASH, abc.ABC_FUNC_AIG, 0)
	pi := abc.Abc_NtkCreatePi(n)
	lo := abc.Abc_NtkAddLatch(n, nil, abc.ABC_INIT_ONE)
	li := abc.Abc_ObjFanin0(abc.Abc_ObjFanin0(lo))
	abc.Abc_ObjAddFanin(li, pi)
	po := abc.Abc_NtkCreatePo(n)
	abc.Abc_ObjAddFanin(po, lo)
	abc.Abc_NtkOrderCisCos(n)

	g := Abc_NtkToDar(n, 0, 1)
	if g == nil || g.InputCount() != 2 || g.OutputCount() != 2 || g.RegisterCount() != 1 {
		t.Fatalf("unexpected frame interface: graph=%v", g)
	}
	// ABC orders primary inputs before latch outputs, and primary outputs
	// before latch inputs. Init-one normalization complements both the
	// register state input and its next-state output.
	for pattern := 0; pattern < 4; pattern++ {
		in := []bool{pattern&1 != 0, pattern&2 != 0}
		out, err := g.Evaluate(in)
		if err != nil || out[0] != !in[1] || out[1] != !in[0] {
			t.Fatalf("input %v output %v error %v", in, out, err)
		}
	}
}

func TestVerifyPatternSliceReturnsOrderedOutputs(t *testing.T) {
	n := abc.Abc_NtkAlloc(abc.ABC_NTK_STRASH, abc.ABC_FUNC_AIG, 1)
	x := abc.Abc_NtkCreatePi(n)
	po0, po1 := abc.Abc_NtkCreatePo(n), abc.Abc_NtkCreatePo(n)
	abc.Abc_ObjAddFanin(po0, x)
	abc.Abc_ObjAddFanin(po1, x.Signal().Not())
	if got := Abc_NtkVerifySimulatePatternSlice(n, []int32{1}); len(got) != 2 || got[0] != 1 || got[1] != 0 {
		t.Fatalf("simulated outputs = %v", got)
	}
	if got := Abc_NtkVerifySimulatePatternSlice(n, nil); got != nil {
		t.Fatalf("short model unexpectedly produced outputs: %v", got)
	}
}
