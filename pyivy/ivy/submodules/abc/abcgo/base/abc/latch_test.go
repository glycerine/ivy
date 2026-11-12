package abc

import (
	"github.com/glycerine/abcgo/misc/util"
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

func TestRemoveComplementedSelfFeedPreservesPolarity(t *testing.T) {
	n := NewNetwork("complemented loop")
	output := Abc_NtkAddLatch(n, nil, ABC_INIT_ZERO)
	latch := output.fanins[0].Object
	input := latch.fanins[0].Object
	Abc_ObjAddFanin(input, output.Signal().Not())
	if got := Abc_NtkRemoveSelfFeedLatches(n); got != 1 {
		t.Fatalf("removed %d latches", got)
	}
	if input.fanins[0] != n.One().Not() {
		t.Fatal("complemented self-feed was not replaced with constant zero")
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

func TestOnehotConversionPreservesTransitionTable(t *testing.T) {
	n := NewNetwork("two-bit")
	n.Type = uint8(ABC_NTK_LOGIC)
	n.Function = uint8(ABC_FUNC_SOP)
	pi := n.NewObject(ObjectPI)
	q0 := Abc_NtkAddLatch(n, nil, ABC_INIT_ONE)
	q1 := Abc_NtkAddLatch(n, nil, ABC_INIT_ZERO)
	xor := func(a, b *Object) *Object {
		o := n.NewObject(ObjectNode)
		o.Data = "01 1\n10 1\n"
		Abc_ObjAddFanin(o, a)
		Abc_ObjAddFanin(o, b)
		return o
	}
	Abc_ObjAddFanin(q0.fanins[0].Object.fanins[0].Object, xor(pi, q1))
	Abc_ObjAddFanin(q1.fanins[0].Object.fanins[0].Object, q0.Signal().Not())
	po := n.NewObject(ObjectPO)
	Abc_ObjAddFanin(po, xor(q0, q1))
	Abc_NtkOrderCisCos(n)
	hot := Abc_NtkConvertOnehot(n)
	if got := Abc_NtkCollectLatchValuesStr(hot); got != "0100" {
		t.Fatalf("initial state = %q", got)
	}
	for state := 0; state < 4; state++ {
		for input := 0; input < 2; input++ {
			oldValues := map[*Object]bool{pi: input != 0, q0: state&1 != 0, q1: state&2 != 0}
			values := map[*Object]bool{hot.ObjectsOfKind(ObjectPI)[0]: input != 0}
			for index, output := range hot.ObjectsOfKind(ObjectBO) {
				values[output] = index == state
			}
			want, err := n.Evaluate(po.Signal(), oldValues)
			if err != nil {
				t.Fatal(err)
			}
			got, err := hot.Evaluate(hot.ObjectsOfKind(ObjectPO)[0].Signal(), values)
			if err != nil || got != want {
				t.Fatalf("output state=%d input=%d got=%v want=%v err=%v", state, input, got, want, err)
			}
			next := 0
			for bit, latch := range n.ObjectsOfKind(ObjectLatch) {
				v, err := n.Evaluate(latch.fanins[0], oldValues)
				if err != nil {
					t.Fatal(err)
				}
				if v {
					next |= 1 << bit
				}
			}
			for index, latch := range hot.ObjectsOfKind(ObjectLatch) {
				v, err := hot.Evaluate(latch.fanins[0], values)
				if err != nil || v != (index == next) {
					t.Fatalf("next state=%d input=%d hot=%d got=%v want=%v err=%v", state, input, index, v, index == next, err)
				}
			}
		}
	}
}

func TestCounterexampleReplaysLatchUpdatesSimultaneously(t *testing.T) {
	n := NewNetwork("swap")
	q0 := Abc_NtkAddLatch(n, nil, ABC_INIT_ONE)
	q1 := Abc_NtkAddLatch(n, q0, ABC_INIT_ZERO)
	Abc_ObjAddFanin(q0.fanins[0].Object.fanins[0].Object, q1)
	po := n.NewObject(ObjectPO)
	Abc_ObjAddFanin(po, q1)
	trace := &util.Abc_Cex_t{NRegs: 2, NBits: 2, PData: []uint32{1}}
	if got := Abc_NtkVerifyCex(n, trace); got != -1 {
		t.Fatal("wrong initial output", got)
	}
	trace.IFrame = 1
	if got := Abc_NtkVerifyCex(n, trace); got != 0 {
		t.Fatal("wrong output after simultaneous swap", got)
	}
}

func TestRestoreClassifiedLatchEnable(t *testing.T) {
	n := NewNetwork("enable")
	n.Type = uint8(ABC_NTK_LOGIC)
	n.Function = uint8(ABC_FUNC_SOP)
	control := n.NewObject(ObjectPI)
	driver := n.NewObject(ObjectPI)
	q := Abc_NtkAddLatch(n, driver, ABC_INIT_ZERO)
	po := n.NewObject(ObjectPO)
	Abc_ObjAddFanin(po, q)
	controlPO := n.NewObject(ObjectPO)
	Abc_ObjAddFanin(controlPO, control.Signal().Not())
	Abc_NtkOrderCisCos(n)
	out := Abc_NtkDup(n)
	controls := vec.Vec_PtrAlloc(1)
	vec.Vec_PtrPush(controls, controlPO)
	classes := vec.Vec_IntAlloc(1)
	vec.Vec_IntPush(classes, 0)
	Abc_NtkTransformBack(n, out, controls, classes)
	if Abc_NtkPoNum(out) != 1 {
		t.Fatal("control output not removed")
	}
	for _, enable := range []bool{false, true} {
		for _, data := range []bool{false, true} {
			for _, state := range []bool{false, true} {
				inputs := map[*Object]bool{out.ObjectsOfKind(ObjectPI)[0]: enable, out.ObjectsOfKind(ObjectPI)[1]: data, out.ObjectsOfKind(ObjectBO)[0]: state}
				got, err := out.Evaluate(out.ObjectsOfKind(ObjectBI)[0].Signal(), inputs)
				want := data
				if enable {
					want = state
				}
				if err != nil || got != want {
					t.Fatalf("enable=%v data=%v state=%v got=%v err=%v", enable, data, state, got, err)
				}
			}
		}
	}
	if err := out.Check(); err != nil {
		t.Fatal(err)
	}
}
