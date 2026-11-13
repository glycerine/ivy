package acec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"testing"
)

func TestCutDetectorsRejectNativeMux(t *testing.T) {
	for _, detect := range []struct {
		name string
		call func(*Gia_Man_t)
	}{{"Ree", func(p *Gia_Man_t) { Ree_ManComputeCuts(p, nil, 0) }}, {"Dtc", func(p *Gia_Man_t) { var x, y *Vec_Int_t; Dtc_ManComputeCuts(p, nil, &x, &y, 0) }}} {
		t.Run(detect.name, func(t *testing.T) {
			p := gia.Gia_ManStart(8)
			p.EnableMuxes()
			a, b, c := gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p)
			root := gia.Gia_ManAppendMuxReal(p, a, b, c)
			gia.Gia_ManAppendCo(p, root)
			defer func() {
				if recover() == nil {
					t.Error("detector silently accepted a native mux")
				}
			}()
			detect.call(p)
		})
	}
}
