package abc

import (
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestLatchSelfFeedCycles(t *testing.T) {
	n := NewNetwork("cycle")
	outputs := []*Object{Abc_NtkAddLatch(n, nil, ABC_INIT_ZERO), Abc_NtkAddLatch(n, nil, ABC_INIT_ONE), Abc_NtkAddLatch(n, nil, ABC_INIT_DC)}
	for i, output := range outputs {
		latch := output.fanins[0].Object
		Abc_ObjAddFanin(latch.fanins[0].Object, outputs[(i+1)%len(outputs)])
	}
	tail := Abc_NtkAddLatch(n, outputs[0], ABC_INIT_ZERO)
	if got := Abc_NtkCountSelfFeedLatches(n); got != 3 {
		t.Fatalf("cycle count = %d", got)
	}
	if Abc_NtkLatchIsSelfFeed(tail.fanins[0].Object) != 0 {
		t.Fatal("tail entering another cycle is not self-feeding")
	}
	if got := Abc_NtkRemoveSelfFeedLatches(n); got != 1 {
		t.Fatalf("cycle-breaking replacements = %d", got)
	}
	if Abc_NtkCountSelfFeedLatches(n) != 0 {
		t.Fatal("cycle survived replacement")
	}
	if err := n.Check(); err != nil {
		t.Fatal(err)
	}
}

func TestLatchInitializationRoundTrip(t *testing.T) {
	n := NewNetwork("initial")
	for _, init := range []Abc_InitType_t{ABC_INIT_ZERO, ABC_INIT_ONE, ABC_INIT_DC} {
		Abc_NtkAddLatch(n, n.constant, init)
	}
	if got := Abc_NtkCollectLatchValuesStr(n); got != "01x" {
		t.Fatal(got)
	}
	values := Abc_NtkCollectLatchValues(n)
	if vec.Vec_IntSize(values) != 3 || vec.Vec_IntEntry(values, 1) != 1 {
		t.Fatal("wrong collected values")
	}
	Abc_NtkInsertLatchValues(n, values)
	if got := Abc_NtkCollectLatchValuesStr(n); got != "010" {
		t.Fatal(got)
	}
	Abc_NtkInsertLatchValues(n, nil)
	if got := Abc_NtkCollectLatchValuesStr(n); got != "xxx" {
		t.Fatal(got)
	}
}

func TestLatchPipelinePreservesComplementedFanouts(t *testing.T) {
	n := NewNetwork("pipeline")
	pi := n.NewObject(ObjectPI)
	po := n.NewObject(ObjectPO)
	Abc_ObjAddFanin(po, pi.Signal().Not())
	Abc_NtkLatchPipe(n, 2)
	if Abc_NtkLatchNum(n) != 2 || !po.fanins[0].Complement || po.fanins[0].Object.Kind != ObjectBO {
		t.Fatal("pipeline interface or polarity lost")
	}
	if len(pi.fanouts) != 1 || pi.fanouts[0].Kind != ObjectBI {
		t.Fatal("PI fanouts not redirected")
	}
	if Abc_NtkCollectLatchValuesStr(n) != "xx" {
		t.Fatal("pipeline latches should be uninitialized")
	}
	if err := n.Check(); err != nil {
		t.Fatal(err)
	}
}

func TestConvertDcLatchMux(t *testing.T) {
	n := NewNetwork("dc")
	n.Type = uint8(ABC_NTK_LOGIC)
	n.Function = uint8(ABC_FUNC_SOP)
	pi := n.NewObject(ObjectPI)
	output := Abc_NtkAddLatch(n, pi, ABC_INIT_DC)
	po := n.NewObject(ObjectPO)
	Abc_ObjAddFanin(po, output.Signal().Not())
	Abc_NtkConvertDcLatches(n)
	if Abc_NtkLatchNum(n) != 2 || Abc_NtkPiNum(n) != 2 {
		t.Fatal("missing control latch or initialization input")
	}
	control := n.ObjectsOfKind(ObjectBO)[1]
	init := n.ObjectsOfKind(ObjectPI)[1]
	for _, enabled := range []bool{false, true} {
		for _, state := range []bool{false, true} {
			for _, value := range []bool{false, true} {
				got, err := n.Evaluate(po.Signal(), map[*Object]bool{control: enabled, output: state, init: value, pi: false})
				want := value
				if enabled {
					want = state
				}
				if err != nil || got == want {
					t.Fatalf("mux control=%v state=%v init=%v got=%v err=%v", enabled, state, value, got, err)
				}
			}
		}
	}
	if err := n.Check(); err != nil {
		t.Fatal(err)
	}
}
