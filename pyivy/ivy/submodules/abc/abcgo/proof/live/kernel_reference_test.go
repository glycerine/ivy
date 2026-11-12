package live

import (
	_ "embed"
	"encoding/json"
	"github.com/glycerine/abcgo/aig/aig"
	"strings"
	"testing"
)

//go:embed test_vectors/kernel_traces.json
var kernelTraces []byte

func kernelFixture(kind string, depth int) *aig.Graph {
	p := aig.Aig_ManStart(30)
	p.Name = "fixture"
	input := aig.Aig_ObjCreateCi(p)
	aig.Aig_ObjCreateCi(p)
	switch kind {
	case "absorber":
		aig.Aig_ObjCreateCo(p, input)
		aig.Aig_ObjCreateCo(p, input)
		aig.Aig_ManSetRegNum(p, 1)
		zero, index := int32(-1), int32(-1)
		p = createNewAigWith0LivePo(p, []aig.Edge{input}, &zero)
		for i := 0; i < depth; i++ {
			p = introduceAbsorberLogic(p, &zero, &index, int32(i))
		}
		return p
	case "combination":
		aig.Aig_ObjCreateCi(p)
		aig.Aig_ObjCreateCo(p, input)
		aig.Aig_ObjCreateCo(p, input)
		aig.Aig_ObjCreateCo(p, input.Not())
		aig.Aig_ManSetRegNum(p, 2)
		return generateGeneralDisjunctiveTester(nil, p, 2)
	case "monotone":
		n := monotoneFixture()
		indexes := &aigPoIndices{0, 1, 3, 4}
		vectors := &monotoneVectorsStruct{attrCandMonotone: []int32{1, 3}, attrHintMonotone: []int32{1, 3}}
		return createMonotoneTester(n.Graph, indexes, vectors, nil)
	case "disjunctive":
		n := monotoneFixture()
		indexes := &aigPoIndices{0, 1, 3, 4}
		vectors := &antecedentConsequentVectorsStruct{attrConsequentCandidates: []int32{1, 3}}
		return createDisjunctiveMonotoneTester(n.Graph, indexes, vectors, nil)
	case "arena":
		for _, edge := range []aig.Edge{input, 0, 0, 0, input, input.Not(), input} {
			aig.Aig_ObjCreateCo(p, edge)
		}
		aig.Aig_ManSetRegNum(p, 1)
		index := int32(-1)
		return createNewAigWith0LivePoWithDSC(p, []aig.Edge{input}, &index, 1, 2, 3, [][]int32{{4}, {5}})
	}
	panic("unknown kernel fixture")
}
func TestOriginalCKernelTraces(t *testing.T) {
	var fixtures []struct {
		Kind                  string
		Depth, Pis, Pos, Regs int
		Traces                []string
	}
	if err := json.Unmarshal(kernelTraces, &fixtures); err != nil {
		t.Fatal(err)
	}
	for _, fixture := range fixtures {
		p := kernelFixture(fixture.Kind, fixture.Depth)
		if trueInputs(p) != fixture.Pis || trueOutputs(p) != fixture.Pos || registers(p) != fixture.Regs {
			t.Fatalf("%s interface mismatch", fixture.Kind)
		}
		for pattern, want := range fixture.Traces {
			state := make([]bool, registers(p))
			var got strings.Builder
			for frame := 0; frame < 6; frame++ {
				inputs := []bool{pattern&(1<<uint(frame)) != 0}
				outputs, next := evaluateStep(p, state, inputs)
				for _, value := range outputs {
					if value {
						got.WriteByte('1')
					} else {
						got.WriteByte('0')
					}
				}
				state = next
			}
			if got.String() != want {
				t.Fatalf("%s depth%d pattern%d trace%s want%s", fixture.Kind, fixture.Depth, pattern, got.String(), want)
			}
		}
	}
}
