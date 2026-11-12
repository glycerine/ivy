package saig

import (
	"github.com/glycerine/abcgo/aig/aig"
	"testing"
)

func TestSaigMultivalueWrapperSharesAigSemantics(t *testing.T) {
	p := aig.NewGraph()
	a, b := p.NewInput(), p.NewInput()
	p.AddOutput(p.OR(a, b))
	p.AddOutput(a.Not())
	p.AddOutput(a.Not())
	aig.Aig_ManSetRegNum(p, 2)
	mapping := Saig_MvManSimulate(p, 1, 3, 0, 0)
	if len(mapping) != 2 || mapping[0] != a || mapping[1] != a {
		t.Fatal("symbolic equivalent registers were not mapped", mapping)
	}
	direct := aig.Aig_MvManSimulate(p, 1, 3, 0, 0)
	for i := range mapping {
		if mapping[i] != direct[i] {
			t.Fatal("SAIG and AIG simulator semantics differ")
		}
	}
}
func TestSaigMultivalueHashWrapper(t *testing.T) {
	values := []uint32{1, 2, uint32(SAIG_UNDEF_VALUE)}
	if Saig_MvSimHash(values, 3, 503) != aig.Aig_MvSimHash(values, 3, 503) {
		t.Fatal("hash wrapper differs")
	}
}
