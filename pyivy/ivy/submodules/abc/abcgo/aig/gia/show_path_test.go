package gia

import (
	"github.com/glycerine/abcgo/misc/vec"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestShowCriticalLutPath(t *testing.T) {
	p := Gia_ManStart(12)
	a, b, c := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	x := Gia_ManAppendAnd(p, a, b)
	y := Gia_ManAppendAnd(p, x, c)
	Gia_ManAppendCo(p, y)
	Gia_ManAppendCo(p, x)
	p.VMapping = vec.Vec_IntStart(int(p.nObjs))
	for _, lut := range []struct {
		id   int32
		fans []int32
	}{{x >> 1, []int32{a >> 1, b >> 1}}, {y >> 1, []int32{x >> 1, c >> 1}}} {
		vec.Vec_IntWriteEntry(p.VMapping, int(lut.id), int32(vec.Vec_IntSize(p.VMapping)))
		vec.Vec_IntPush(p.VMapping, int32(len(lut.fans)))
		for _, fan := range lut.fans {
			vec.Vec_IntPush(p.VMapping, fan)
		}
		vec.Vec_IntPush(p.VMapping, lut.id)
	}
	path := giaCriticalLuts(p)
	if path.depth != 2 || path.nodes != 2 || path.edges != 3 || path.marked[c>>1] {
		t.Fatalf("incorrect critical cone: %+v", path)
	}
	filename := filepath.Join(t.TempDir(), "critical.dot")
	Gia_ShowPath(p, filename)
	data, err := os.ReadFile(filename)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"2 LUTs with 3 critical edges", "Node5 -> Node4 [style = solid]", "Node4 -> Node1 [style = solid]"} {
		if !strings.Contains(string(data), want) {
			t.Fatalf("missing %q in %s", want, data)
		}
	}
	if strings.Contains(string(data), "Node5 -> Node3 [style = solid]") {
		t.Fatal("included noncritical input")
	}
	if p.vLevels != nil {
		t.Fatal("temporary levels retained")
	}
}
