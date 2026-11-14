package gia

import (
	"fmt"
	"testing"
)

func TestHcdChoicesRejectsPrimaryInputRegisterRepresentative(t *testing.T) {
	p := Gia_ManStart(1)
	a, ro := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	Gia_ManAppendCo(p, ro)
	Gia_ManAppendCo(p, a)
	Gia_ManSetRegNum(p, 1)
	p.StartRepresentatives()
	Gia_ObjSetRepr(p, ro>>1, a>>1)
	defer func() {
		if recover() == nil {
			t.Fatal("accepted PI as a register representative, contrary to C's precondition")
		}
	}()
	Hcd_ManEquivToChoices(p, 1)
}

func TestChoiceRegisterRepresentativesPreserveSequentialTraces(t *testing.T) {
	for name, convert := range map[string]func(*Gia_Man_t, int32) *Gia_Man_t{"gia": Gia_ManEquivToChoices, "hcd": Hcd_ManEquivToChoices} {
		for _, constant := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s_constant%t", name, constant), func(t *testing.T) {
				p := Gia_ManStart(1)
				a, ro0, ro1 := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
				Gia_ManAppendCo(p, ro1)
				Gia_ManAppendCo(p, a)
				if constant {
					Gia_ManAppendCo(p, 0)
				} else {
					Gia_ManAppendCo(p, a)
				}
				Gia_ManSetRegNum(p, 2)
				p.StartRepresentatives()
				if constant {
					Gia_ObjSetRepr(p, ro1>>1, 0)
				} else {
					Gia_ObjSetRepr(p, ro1>>1, ro0>>1)
				}
				q := convert(p, 1)
				if Gia_ManRegNum(q) != 2 {
					t.Fatal("register bookkeeping changed")
				}
				for sequence := 0; sequence < 16; sequence++ {
					state := []bool{false, false}
					for frame := 0; frame < 4; frame++ {
						input := sequence&(1<<frame) != 0
						want, err := p.Evaluate([]bool{input, state[0], state[1]})
						if err != nil {
							t.Fatal(err)
						}
						got, err := q.Evaluate([]bool{input, state[0], state[1]})
						if err != nil || len(got) != len(want) {
							t.Fatalf("trace %d frame %d: %v, %v", sequence, frame, got, err)
						}
						for i := range want {
							if got[i] != want[i] {
								t.Fatalf("trace %d frame %d: %v, want %v", sequence, frame, got, want)
							}
						}
						state = want[1:]
					}
				}
			})
		}
	}
}
