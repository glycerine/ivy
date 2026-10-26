package abc

// Stand-alone tests for the satSolver2i interpolation composition.

import (
	"errors"
	"testing"
)

func satSolver2iRelationGraph(t testing.TB) *GiaManager {
	t.Helper()
	graph := NewGiaManager()
	a, _ := graph.AppendCI()
	b, _ := graph.AppendCI()
	c, _ := graph.AppendCI()
	and, _ := graph.AppendAnd(a, b.Not())
	xor, _ := graph.AppendXorReal(and, c)
	mux, _ := graph.AppendMuxReal(a, xor, b)
	if _, err := graph.AppendCO(mux.Not()); err != nil {
		t.Fatal(err)
	}
	return graph
}

func satSolver2iCheckRelation(t testing.TB, source *GiaManager, result GiaOutputRelationInterpolation) {
	t.Helper()
	if result.Graph == nil || result.OriginalInputs != source.InputCount() || result.ClaimInput != source.InputCount() || result.Graph.InputCount() != source.InputCount()+1 || result.Graph.OutputCount() != 1 {
		t.Fatalf("bad relation shape: %+v", result)
	}
	if result.Graph.Name != "interpolant" {
		t.Fatalf("name = %q", result.Graph.Name)
	}
	for assignment := 0; assignment < 1<<source.InputCount(); assignment++ {
		inputs := make([]bool, source.InputCount())
		for index := range inputs {
			inputs[index] = assignment&(1<<index) != 0
		}
		want, err := source.Evaluate(inputs)
		if err != nil {
			t.Fatal(err)
		}
		for _, claim := range []bool{false, true} {
			relationInputs := append(append([]bool(nil), inputs...), claim)
			got, evalErr := result.Graph.Evaluate(relationInputs)
			if evalErr != nil || got[0] != (claim == want[0]) {
				t.Fatalf("inputs=%v claim=%v got=%v err=%v source=%v", inputs, claim, got, evalErr, want[0])
			}
		}
	}
}

func TestGiaInterpolateOutputRelation(t *testing.T) {
	source := satSolver2iRelationGraph(t)
	beforeObjects := source.ObjectCount()
	result, err := source.InterpolateOutputRelation(GiaCofactorOptions{})
	if err != nil {
		t.Fatal(err)
	}
	satSolver2iCheckRelation(t, source, result)
	if result.SourceObjects != beforeObjects || source.ObjectCount() != beforeObjects || result.ResultObjects != result.Graph.ObjectCount() {
		t.Fatalf("ownership/stats: result=%+v source objects=%d", result, source.ObjectCount())
	}
	if err = source.Check(); err != nil {
		t.Fatal(err)
	}
}

func TestGiaInterpolateOutputRelationConstantsAndDangling(t *testing.T) {
	for _, constant := range []GiaLit{GiaConst0, GiaConst1} {
		graph := NewGiaManager()
		_, _ = graph.AppendCI()
		if _, err := graph.AppendCO(constant); err != nil {
			t.Fatal(err)
		}
		result, err := graph.InterpolateOutputRelation(GiaCofactorOptions{})
		if err != nil {
			t.Fatal(err)
		}
		satSolver2iCheckRelation(t, graph, result)
	}
	graph := NewGiaManager()
	a, _ := graph.AppendCI()
	b, _ := graph.AppendCI()
	_, _ = graph.AppendAnd(a, b)
	if _, err := graph.AppendCO(a); err != nil {
		t.Fatal(err)
	}
	result, err := graph.InterpolateOutputRelation(GiaCofactorOptions{})
	if err != nil {
		t.Fatal(err)
	}
	satSolver2iCheckRelation(t, graph, result)
	baseline := NewGiaManager()
	baseA, _ := baseline.AppendCI()
	_, _ = baseline.AppendCI()
	_, _ = baseline.AppendCO(baseA)
	baseResult, err := baseline.InterpolateOutputRelation(GiaCofactorOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if result.ResultObjects != baseResult.ResultObjects {
		t.Fatalf("dangling gate changed result size: got=%d baseline=%d", result.ResultObjects, baseResult.ResultObjects)
	}
}

func TestGiaInterpolateOutputRelationErrorsLimitsAndCancellation(t *testing.T) {
	if _, err := (*GiaManager)(nil).InterpolateOutputRelation(GiaCofactorOptions{}); !errors.Is(err, ErrDimension) {
		t.Fatalf("nil source: %v", err)
	}
	for _, options := range []GiaCofactorOptions{{MaxObjects: -1}, {MaxMappingEntries: -1}} {
		if _, err := satSolver2iRelationGraph(t).InterpolateOutputRelation(options); !errors.Is(err, ErrDimension) {
			t.Fatalf("negative option: %v", err)
		}
	}
	empty := NewGiaManager()
	if _, err := empty.InterpolateOutputRelation(GiaCofactorOptions{}); !errors.Is(err, ErrDimension) {
		t.Fatalf("no output: %v", err)
	}
	two := NewGiaManager()
	a, _ := two.AppendCI()
	_, _ = two.AppendCO(a)
	_, _ = two.AppendCO(a.Not())
	if _, err := two.InterpolateOutputRelation(GiaCofactorOptions{}); !errors.Is(err, ErrDimension) {
		t.Fatalf("two outputs: %v", err)
	}
	sequential := NewGiaManager()
	r, _ := sequential.AppendCI()
	_, _ = sequential.AppendCO(r)
	_ = sequential.SetRegisterCount(1)
	if _, err := sequential.InterpolateOutputRelation(GiaCofactorOptions{}); !errors.Is(err, ErrDimension) {
		t.Fatalf("sequential source: %v", err)
	}
	source := satSolver2iRelationGraph(t)
	if _, err := source.InterpolateOutputRelation(GiaCofactorOptions{MaxMappingEntries: source.ObjectCount() - 1}); !errors.Is(err, ErrLimit) {
		t.Fatalf("mapping limit: %v", err)
	}
	if _, err := source.InterpolateOutputRelation(GiaCofactorOptions{MaxObjects: 1}); !errors.Is(err, ErrLimit) {
		t.Fatalf("object limit: %v", err)
	}
	if _, err := source.InterpolateOutputRelation(GiaCofactorOptions{Stop: func() bool { return true }}); !errors.Is(err, ErrCancelled) {
		t.Fatalf("cancellation: %v", err)
	}
}

func FuzzGiaInterpolateOutputRelation(f *testing.F) {
	for _, seed := range [][]byte{{0}, {1, 2, 3}, {255, 17, 91, 4}} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 48 {
			data = data[:48]
		}
		graph := NewGiaManager()
		signals := []GiaLit{GiaConst0, GiaConst1}
		for range 3 {
			input, err := graph.AppendCI()
			if err != nil {
				t.Fatal(err)
			}
			signals = append(signals, input)
		}
		for index, value := range data {
			first := signals[int(value)%len(signals)].NotIf(value&1 != 0)
			second := signals[(int(value)+index+1)%len(signals)].NotIf(value&2 != 0)
			var next GiaLit
			var err error
			switch value >> 2 & 3 {
			case 0:
				next, err = graph.AppendAndSimplified(first, second)
			case 1:
				next, err = graph.AppendOrSimplified(first, second)
			case 2:
				next, err = graph.AppendXorRealSimplified(first, second)
			default:
				control := signals[(int(value)+2*index+2)%len(signals)]
				next, err = graph.AppendMuxSimplified(control, first, second)
			}
			if err != nil {
				t.Fatal(err)
			}
			signals = append(signals, next)
		}
		if _, err := graph.AppendCO(signals[len(signals)-1]); err != nil {
			t.Fatal(err)
		}
		result, err := graph.InterpolateOutputRelation(GiaCofactorOptions{})
		if err != nil {
			t.Fatal(err)
		}
		satSolver2iCheckRelation(t, graph, result)
	})
}
