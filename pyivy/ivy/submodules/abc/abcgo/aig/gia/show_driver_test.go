package gia

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/glycerine/abcgo/misc/vec"
)

func TestShowDriver(t *testing.T) {
	for _, adders := range []int32{0, 1} {
		p := Gia_ManStart(8)
		p.pName = filepath.Join(t.TempDir(), "network.aig")
		a, b := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
		x := Gia_ManAppendAnd(p, a, b)
		Gia_ManAppendCo(p, x)
		calls := 0
		compute := func(_ *Gia_Man_t, xors **Vec_Int_t, _ int32) *Vec_Int_t {
			calls++
			*xors = vec.Vec_IntAlloc(0)
			return vec.Vec_IntAlloc(0)
		}
		shown := ""
		show := func(name string) { shown = name }
		giaManShow(p, nil, adders, 0, 0, compute, show)
		if calls != int(adders) || shown != filepath.Join(filepath.Dir(p.pName), "network.dot") {
			t.Fatalf("recognizer calls %d viewer path %q", calls, shown)
		}
		data, err := os.ReadFile(shown)
		if err != nil || !strings.Contains(string(data), "digraph AIG") {
			t.Fatalf("missing DOT output: %s %v", data, err)
		}
	}
}
