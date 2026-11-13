package fraig

import (
	"math/rand"
	"testing"
)

func TestConstrainedMiterAgainstExhaustiveTruthTables(t *testing.T) {
	r := rand.New(rand.NewSource(838))
	for trial := 0; trial < 30; trial++ {
		p := Fraig_ManCreate(&Fraig_Params_t{NPatsRand: 128, NPatsDyna: 128, NBTLimit: -1})
		a, b, c := Fraig_ManReadIthVar(p, 0), Fraig_ManReadIthVar(p, 1), Fraig_ManReadIthVar(p, 2)
		nodes := []*Fraig_Node_t{a, b, c, Fraig_NodeAnd(p, a, b), Fraig_NodeExor(p, b, c)}
		truth := []uint8{0xaa, 0xcc, 0xf0, 0x88, 0x3c}
		allowed := uint8(0xff)
		for k := 0; k < 2; k++ {
			var clause []*Fraig_Node_t
			var mask uint8
			for {
				clause = nil
				mask = 0
				for z := 0; z < 1+r.Intn(2); z++ {
					i := r.Intn(len(nodes))
					edge, bits := nodes[i], truth[i]
					if r.Intn(2) != 0 {
						edge = Fraig_Not(edge)
						bits = ^bits
					}
					clause = append(clause, edge)
					mask |= bits
				}
				if allowed&mask != 0 {
					break
				}
			}
			Fraig_ManAddClause(p, clause, int32(len(clause)))
			allowed &= mask
		}
		i := r.Intn(len(nodes))
		output, bits := nodes[i], truth[i]
		if r.Intn(2) != 0 {
			output = Fraig_Not(output)
			bits = ^bits
		}
		Fraig_ManSetPo(p, output)
		status := Fraig_ManCheckMiter(p)
		wantSAT := allowed&bits != 0
		if (status == 0) != wantSAT || status < 0 {
			t.Fatalf("trial %d status %d, allowed %08b output %08b", trial, status, allowed, bits)
		}
		if status == 0 {
			assignment := int(p.pModel[0]) | int(p.pModel[1])<<1 | int(p.pModel[2])<<2
			if allowed&(1<<uint(assignment)) == 0 || bits&(1<<uint(assignment)) == 0 {
				t.Fatalf("trial %d invalid witness %v", trial, p.pModel)
			}
		}
		Fraig_ManFree(p)
	}
}
