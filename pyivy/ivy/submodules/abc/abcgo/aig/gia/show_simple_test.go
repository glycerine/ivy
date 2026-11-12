package gia

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/glycerine/abcgo/misc/vec"
)

func TestWriteDotAigSimple(t *testing.T) {
	p := Gia_ManStart(10)
	a, b := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	x := Gia_ManAppendAnd(p, a, b^1)
	out := Gia_ManAppendCo(p, x)
	Gia_ManAppendCo(p, 0)
	bold := vec.Vec_IntAlloc(1)
	vec.Vec_IntPush(bold, x>>1)
	filename := filepath.Join(t.TempDir(), "graph.dot")
	Gia_WriteDotAigSimple(p, filename, bold)
	data, err := os.ReadFile(filename)
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	for _, want := range []string{"digraph AIG", "Node3 [label = \"3\", shape = ellipse, style = filled]", "Node3 -> Node1 [style = solid]", "Node3 -> Node2 [style = dotted]", "Node0 [label = \"Const0\"", "Node4 -> Node3 [style = solid]"} {
		if !strings.Contains(text, want) {
			t.Fatalf("missing %q in %s", want, text)
		}
	}
	_ = out
	if p.vLevels != nil {
		t.Fatal("levels were not cleared")
	}
}
