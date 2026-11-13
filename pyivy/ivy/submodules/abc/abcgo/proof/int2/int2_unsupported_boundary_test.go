package int2

import (
	"reflect"
	"testing"

	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
)

func TestUnsupportedPreimageLeavesInputMapsUntouched(t *testing.T) {
	ci := vec.Vec_IntAllocArrayCopy([]int32{0, 3}, 2)
	co := vec.Vec_IntAllocArrayCopy([]int32{1, 5}, 2)
	priority := vec.Vec_IntAllocArrayCopy([]int32{7, 9}, 2)
	for _, graph := range []*Gia_Man_t{nil, testInt2ToggleMachine(), testInt2InputToBadMachine()} {
		image, status := Int2_ManComputePreimage(graph, nil, nil, ci, co, priority)
		if image != nil || status != INT2_TIME_OUT {
			t.Fatalf("image=%v status=%d", image, status)
		}
		if !reflect.DeepEqual(vec.Vec_IntArray(ci), []int32{0, 3}) || !reflect.DeepEqual(vec.Vec_IntArray(co), []int32{1, 5}) || !reflect.DeepEqual(vec.Vec_IntArray(priority), []int32{7, 9}) {
			t.Fatal("unsupported preimage changed an input mapping")
		}
	}
}

func TestSourceFrameInitZeroFramesUsesExistingFaninValue(t *testing.T) {
	g := gia.Gia_ManStart(8)
	input, state := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	gia.Gia_ManAppendCo(g, 0)
	gia.Gia_ManAppendCo(g, input^1)
	gia.Gia_ManSetRegNum(g, 1)
	for _, frames := range []int32{-1, 0} {
		gia.Gia_ObjSetValue(gia.Gia_ManObj(g, input>>1), 0)
		gia.Gia_ObjSetValue(gia.Gia_ManObj(g, state>>1), 1)
		result := Int2_ManFrameInit(g, frames, 0)
		values, err := result.Evaluate(nil)
		if err != nil || len(values) != 1 || !values[0] {
			t.Fatalf("frames=%d result=%v err=%v", frames, values, err)
		}
	}
}

func TestSourceUnrollAndPrefixReturnUnsupported(t *testing.T) {
	graph := testInt2InputToBadMachine()
	count := graph.ObjectCount()
	outputs := append([]int32(nil), graph.OutputLiterals()...)
	for _, frame := range []int32{-1, 0, 1, 10} {
		if result := Int2_ManUnroll(graph, frame); result != nil {
			t.Fatalf("frame %d returned reconstructed graph", frame)
		}
		mapping := vec.Vec_IntAllocArrayCopy([]int32{4, 5}, 2)
		if solver := Int2_ManPreparePrefix(graph, frame, &mapping); solver != nil || mapping != nil {
			t.Fatalf("frame %d returned unsupported prefix state", frame)
		}
	}
	if Int2_ManUnroll(nil, 1) != nil || Int2_ManPreparePrefix(nil, 1, nil) != nil {
		t.Fatal("nil source arguments produced a result")
	}
	if graph.ObjectCount() != count || !reflect.DeepEqual(graph.OutputLiterals(), outputs) {
		t.Fatal("unsupported path changed the source graph")
	}
}

func TestSourceFrameBuilderReportsUnsupportedBeforeMutation(t *testing.T) {
	graph := testInt2InputToBadMachine()
	m := Int2_ManCreate(graph, &Int2_ManPars_t{})
	defer Int2_ManStop(m)
	m.pSatPref = sat_solver_new()
	frames, objects := vec.Vec_PtrSize(m.vMapFrames), m.pGiaPref.ObjectCount()
	defer func() {
		if recover() == nil {
			t.Fatal("unsupported void API must report an error")
		}
		if vec.Vec_PtrSize(m.vMapFrames) != frames || m.pGiaPref.ObjectCount() != objects {
			t.Fatal("unsupported frame builder changed manager state")
		}
	}()
	Int2_ManCreateFrames(m, 0, vec.Vec_IntAllocArrayCopy([]int32{0}, 1))
}
