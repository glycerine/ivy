package gia

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/glycerine/abcgo/misc/vec"
)

func TestGroupedDotAdder(t *testing.T) {
	p := Gia_ManStart(12)
	a, b, c := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	x := Gia_ManAppendAnd(p, a, b)
	y := Gia_ManAppendAnd(p, x, c)
	Gia_ManAppendCo(p, y)
	Gia_ManAppendCo(p, x)
	adds := vec.Vec_IntAlloc(6)
	for _, v := range []int32{a >> 1, b >> 1, 0, x >> 1, y >> 1, 0} {
		vec.Vec_IntPush(adds, v)
	}
	xors := vec.Vec_IntAlloc(0)
	name := filepath.Join(t.TempDir(), "grouped.dot")
	Gia_ShowProcess(p, name, nil, adds, xors, 0)
	data, err := os.ReadFile(name)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"0 full-adders, and 1 half-adders", "Node5 [label = \"4_5\", shape = octagon]", "Node5 -> Node1 [style = solid]", "Node5 -> Node2 [style = solid]", "Node7 -> Node5 [style = solid]"} {
		if !strings.Contains(string(data), want) {
			t.Fatalf("missing %q in %s", want, data)
		}
	}
	if p.vLevels != nil {
		t.Fatal("temporary levels retained")
	}
}
