package gia

import (
	"fmt"
	"testing"
)

func TestEquivToChoicesPreservesRealGateFunctions(t *testing.T) {
	for name, convert := range map[string]func(*Gia_Man_t, int32) *Gia_Man_t{
		"gia": Gia_ManEquivToChoices,
		"hcd": Hcd_ManEquivToChoices,
	} {
		for _, kind := range []string{"xor", "mux", "buffer"} {
			for phase := int32(0); phase < 16; phase++ {
				t.Run(fmt.Sprintf("%s_%s_%d", name, kind, phase), func(t *testing.T) {
					p := Gia_ManStart(1)
					a, b, c := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
					var root int32
					switch kind {
					case "xor":
						root = Gia_ManAppendXorReal(p, a^(phase&1), b^((phase>>1)&1))
					case "mux":
						p.EnableMuxes()
						root = Gia_ManAppendMuxReal(p, a^(phase&1), b^((phase>>1)&1), c^((phase>>2)&1))
					case "buffer":
						root = Gia_ManAppendBuf(p, a^(phase&1))
					}
					Gia_ManAppendCo(p, root^((phase>>3)&1))
					p.StartRepresentatives()
					q := convert(p, 1)
					if Gia_ManCountChoices(q) != 0 {
						t.Fatal("conversion invented an equivalence class")
					}
					for mask := 0; mask < 8; mask++ {
						inputs := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0}
						x, y, z := inputs[0] != (phase&1 != 0), inputs[1] != (phase&2 != 0), inputs[2] != (phase&4 != 0)
						want := x
						switch kind {
						case "xor":
							want = x != y
						case "mux":
							want = x && y || !x && z
						}
						want = want != (phase&8 != 0)
						got, err := q.Evaluate(inputs)
						if err != nil || len(got) != 1 || got[0] != want {
							t.Fatalf("assignment %d: %v, %v; want %t", mask, got, err, want)
						}
					}
				})
			}
		}
	}
}

func TestEquivToChoicesMixedRepresentationClasses(t *testing.T) {
	for name, convert := range map[string]func(*Gia_Man_t, int32) *Gia_Man_t{
		"gia": Gia_ManEquivToChoices,
		"hcd": Hcd_ManEquivToChoices,
	} {
		for phase := int32(0); phase < 4; phase++ {
			t.Run(fmt.Sprintf("%s_%d", name, phase), func(t *testing.T) {
				p := Gia_ManStart(1)
				a, b := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
				x := Gia_ManAppendXorReal(p, a^(phase&1), b^((phase>>1)&1))
				y := Gia_ManAppendXor(p, a^(phase&1), b^((phase>>1)&1))
				Gia_ManAppendCo(p, x)
				Gia_ManAppendCo(p, y)
				Gia_ManAppendCo(p, y^1)
				p.StartRepresentatives()
				Gia_ObjSetRepr(p, y>>1, x>>1)
				Gia_ObjSetNext(p, x>>1, y>>1)
				q := convert(p, 1)
				Gia_ManSetPhase(q)
				for mask := 0; mask < 4; mask++ {
					inputs := []bool{mask&1 != 0, mask&2 != 0}
					want := (inputs[0] != (phase&1 != 0)) != (inputs[1] != (phase&2 != 0))
					got, err := q.Evaluate(inputs)
					if err != nil || len(got) != 3 || got[0] != want || got[1] != want || got[2] == want {
						t.Fatalf("assignment %d: %v, %v; want [%t %t %t]", mask, got, err, want, want, !want)
					}
					values, err := giaEvaluateNodes(q, inputs)
					if err != nil {
						t.Fatal(err)
					}
					for id := int32(0); id < q.nObjs; id++ {
						if repr := Gia_ObjReprObj(q, id); repr != nil {
							if (values[id] != values[repr.id]) != (q.pObjs[id].fPhase != repr.fPhase) {
								t.Fatalf("assignment %d: invalid choice phase for %d -> %d", mask, id, repr.id)
							}
						}
					}
				}
			})
		}
	}
}
