package cec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestBackwardPatternRejectsExplicitSpecialGates(t *testing.T) {
	for _, kind := range []string{"xor", "mux"} {
		t.Run(kind, func(t *testing.T) {
			p := gia.Gia_ManStart(16)
			a, b, c := gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p)
			root := int32(0)
			if kind == "xor" {
				root = gia.Gia_ManAppendXorReal(p, a, b)
			} else {
				p.EnableMuxes()
				root = gia.Gia_ManAppendMuxReal(p, c, a, b)
			}
			o := gia.Gia_ManObj(p, root>>1)
			cases := []func(){func() { Cec_ManPatComputePattern_rec(nil, p, o) }, func() { Cec_ManPatComputePattern1_rec(p, o, vec.Vec_IntAlloc(3)) }, func() { Cec_ManPatComputePattern2_rec(p, o, vec.Vec_IntAlloc(3)) }, func() { Cec_ManPatComputePattern3_rec(p, o) }, func() { Cec_ManCheckSat_rec(p, o, 1) }, func() { s := Cec_ManSStart(p, 1); Cec_ManSInsert(s, root>>1) }}
			for i, run := range cases {
				gia.Gia_ManIncrementTravId(p)
				func() {
					defer func() {
						if recover() == nil {
							t.Errorf("path %d silently interpreted %s as AND", i, kind)
						}
					}()
					run()
				}()
			}
		})
	}
}
