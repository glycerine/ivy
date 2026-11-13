package dar

import (
	"errors"
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/tim"
	"github.com/glycerine/abcgo/proof/dch"
	"testing"
)

type testChoiceTiming struct{ value int }

func (t *testChoiceTiming) CloneTiming() any { return &testChoiceTiming{value: t.value} }

func TestNewChoiceWithExplicitOrSourceAIGEngine(t *testing.T) {
	for _, classic := range []bool{false, true} {
		g := aig.NewGraph()
		g.Name = "choices"
		g.Spec = "original.aig"
		timing := &testChoiceTiming{value: 7}
		g.Timing = timing
		a, b, c, d := g.NewInput(), g.NewInput(), g.NewInput(), g.NewInput()
		g.AddOutput(g.AND(g.OR(g.AND(a, c), g.AND(b, c)), d))
		p := ChoiceOptions{Balance: true, UpdateLevel: true, Light: true}
		var result *aig.Graph
		var err error
		if classic {
			result, err = Dar_ManChoiceNewAig(g, &p)
		} else {
			params := dch.DefaultParams()
			p.Engine = dch.Engine{Params: &params}
			result, err = Dar_ManChoiceNew(g, &p)
		}
		if err != nil {
			t.Fatal(err)
		}
		if result.Name != "choices" || result.Spec != "original.aig" || result.Timing != timing || p.SynthesisDuration <= 0 {
			t.Fatal("lost metadata")
		}
		if g.NodeCount() != 0 || g.Timing != nil {
			t.Fatal("C choice path must consume input and transfer timing")
		}
		for n := 0; n < 16; n++ {
			values := []bool{n&1 != 0, n&2 != 0, n&4 != 0, n&8 != 0}
			out, err := result.Evaluate(values)
			want := (values[0] || values[1]) && values[2] && values[3]
			if err != nil || want != out[0] {
				t.Fatal(classic, n, out, err)
			}
		}
	}
}

func TestNewChoiceRequiresUnwiredSourcePathOrExplicitEngine(t *testing.T) {
	g := aig.NewGraph()
	g.AddOutput(g.NewInput())
	p := DefaultChoiceOptions()
	if _, err := Dar_ManChoiceNew(g, &p); err == nil {
		t.Fatal("unwired source choice path did not report an error")
	}
	if _, err := (PartitionedChoiceEngine{}).ComputeChoices([]*aig.Graph{g}); err == nil {
		t.Fatal("partitioned choice without FRAIG backend did not report an error")
	}
}

func TestNewChoiceRestoresTimingBeforeComputingLevels(t *testing.T) {
	g := aig.NewGraph()
	a, b, boxOutput := g.NewInput(), g.NewInput(), g.NewInput()
	g.AddOutput(g.AND(a, b))
	g.AddOutput(g.AND(a, boxOutput))
	manager := tim.Tim_ManStart(3, 2)
	tim.Tim_ManCreateBox(manager, 0, 1, 2, 1, 0, 0)
	g.Timing = manager
	params := dch.DefaultParams()
	options := DefaultChoiceOptions()
	options.Engine = dch.Engine{Params: &params}
	result, err := Dar_ManChoiceNew(g, &options)
	if err != nil {
		t.Fatal(err)
	}
	if result.Timing != manager || aig.Aig_ObjLevel(result, result.Outputs()[1]) != 3 {
		t.Fatalf("timing=%v output level=%d", result.Timing, aig.Aig_ObjLevel(result, result.Outputs()[1]))
	}
}

type rejectedChoiceEngine struct{}

func (rejectedChoiceEngine) ComputeChoices([]*aig.Graph) (*aig.Graph, error) {
	return nil, errors.New("proof failed")
}

type captureChoiceEngine struct{ intermediate *aig.Graph }

func (e *captureChoiceEngine) ComputeChoices(versions []*aig.Graph) (*aig.Graph, error) {
	e.intermediate = aig.Aig_ManDupDfs(versions[0])
	return e.intermediate, nil
}

func TestNewChoiceReleasesIntermediateAfterGuidedDuplication(t *testing.T) {
	g := aig.NewGraph()
	g.AddOutput(g.NewInput())
	engine := &captureChoiceEngine{}
	options := DefaultChoiceOptions()
	options.Engine = engine
	result, err := Dar_ManChoiceNew(g, &options)
	if err != nil {
		t.Fatal(err)
	}
	if engine.intermediate == nil || engine.intermediate.NodeCount() != 0 || result.NodeCount() == 0 {
		t.Fatal("expected consumed intermediate and a live result")
	}
}

type wrongInterfaceEngine struct{}

func (wrongInterfaceEngine) ComputeChoices([]*aig.Graph) (*aig.Graph, error) {
	return aig.NewGraph(), nil
}
func TestNewChoiceRejectsEngineFailures(t *testing.T) {
	g := aig.NewGraph()
	g.AddOutput(g.NewInput())
	for _, engine := range []ChoiceEngine{rejectedChoiceEngine{}, wrongInterfaceEngine{}} {
		p := ChoiceOptions{Engine: engine}
		if _, err := Dar_ManChoiceNew(g, &p); err == nil {
			t.Fatal("lost engine failure")
		}
	}
}
