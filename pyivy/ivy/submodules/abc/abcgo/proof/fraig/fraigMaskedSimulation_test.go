package fraig

import (
	"fmt"
	"strings"
	"testing"
)

func TestMaskedSimulationRequiresRegularNodes(t *testing.T) {
	a, b := &Fraig_Node_t{}, &Fraig_Node_t{}
	for _, tc := range []struct {
		name string
		run  func()
	}{
		{"compare first", func() { Fraig_CompareSimInfoUnderMask(Fraig_Not(a), b, 0, 1, nil) }},
		{"compare second", func() { Fraig_CompareSimInfoUnderMask(a, Fraig_Not(b), 0, 1, nil) }},
		{"xor first", func() { Fraig_CollectXors(Fraig_Not(a), b, 0, 1, nil) }},
		{"xor second", func() { Fraig_CollectXors(a, Fraig_Not(b), 0, 1, nil) }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				p := recover()
				if p == nil || !strings.Contains(fmt.Sprint(p), "simulation nodes are regular") {
					t.Fatalf("want source regular-node assertion, got %v", p)
				}
			}()
			tc.run()
		})
	}
}
