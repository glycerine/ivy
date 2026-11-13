package dar

import (
	"errors"
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/proof/dch"
	"testing"
)

func TestNewChoiceWithExplicitOrSourceAIGEngine(t *testing.T) {
	for _, classic := range []bool{false, true} {
		g := aig.NewGraph()
		g.Name = "choices"
		g.Spec = "original.aig"
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
		if result.Name != g.Name || result.Spec != g.Spec || p.SynthesisDuration <= 0 {
			t.Fatal("lost metadata")
		}
		for n := 0; n < 16; n++ {
			values := []bool{n&1 != 0, n&2 != 0, n&4 != 0, n&8 != 0}
			original, _ := g.Evaluate(values)
			out, err := result.Evaluate(values)
			if err != nil || original[0] != out[0] {
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

type rejectedChoiceEngine struct{}

func (rejectedChoiceEngine) ComputeChoices([]*aig.Graph) (*aig.Graph, error) {
	return nil, errors.New("proof failed")
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
