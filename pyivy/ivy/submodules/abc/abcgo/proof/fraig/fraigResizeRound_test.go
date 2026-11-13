package fraig

import (
	"reflect"
	"testing"
)

// fraigFeed.c doubles simulation storage, retains PI words, resimulates ANDs
// and clears primary-pattern masks. Check two successive resize boundaries.
func TestSimulationResizePreservesConesAcrossRounds(t *testing.T) {
	p := Fraig_ManCreate(&Fraig_Params_t{NPatsRand: 128, NPatsDyna: 128})
	defer Fraig_ManFree(p)
	a, b := Fraig_ManReadIthVar(p, 0), Fraig_ManReadIthVar(p, 1)
	root := Fraig_Regular(Fraig_NodeAnd(p, a, Fraig_Not(b)))
	p.iWordStart = 4
	for i := 0; i < 4; i++ {
		a.puSimD[i] = 0x12345678 ^ uint32(i*0x11111111)
		b.puSimD[i] = 0xabcdef01 ^ uint32(i*0x01010101)
	}
	root.uHashD = 0
	Fraig_NodeSimulate(root, 0, 4, 0)
	ar, ad := append([]uint32(nil), a.puSimR...), append([]uint32(nil), a.puSimD[:4]...)
	br, bd := append([]uint32(nil), b.puSimR...), append([]uint32(nil), b.puSimD[:4]...)
	rr, rd := append([]uint32(nil), root.puSimR...), append([]uint32(nil), root.puSimD[:4]...)
	hashR, hashD := root.uHashR, root.uHashD
	for round := 0; round < 2; round++ {
		oldWords := p.nWordsDyna
		p.pSimsReal[0] = ^uint32(0)
		Fraig_ReallocateSimulationInfo(p)
		if p.nWordsDyna != 2*oldWords || p.pSimsReal[0] != 0 {
			t.Fatalf("round %d dimensions or mask", round)
		}
		if !reflect.DeepEqual(a.puSimR, ar) || !reflect.DeepEqual(a.puSimD[:4], ad) || !reflect.DeepEqual(b.puSimR, br) || !reflect.DeepEqual(b.puSimD[:4], bd) || !reflect.DeepEqual(root.puSimR, rr) || !reflect.DeepEqual(root.puSimD[:4], rd) || root.uHashR != hashR || root.uHashD != hashD {
			t.Fatalf("round %d lost simulation state", round)
		}
	}
}
