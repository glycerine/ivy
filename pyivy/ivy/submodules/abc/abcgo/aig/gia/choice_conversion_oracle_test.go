package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/aig"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func choiceConversionOutput() string {
	snapshots := make([]*Gia_Man_t, 3)
	for i := range snapshots {
		g := Gia_ManStart(16)
		a, b, c := Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g)
		x, y := Gia_ManAppendAnd(g, a, b^1), Gia_ManAppendAnd(g, b, c)
		Gia_ManAppendCo(g, x^int32(i&1))
		Gia_ManAppendCo(g, y^int32((i>>1)&1))
		snapshots[i] = g
	}
	m := Gia_ManChoiceMiter(snapshots)
	var out strings.Builder
	fmt.Fprintf(&out, "gia %d %d %d\n", Gia_ManObjNum(m), Gia_ManCiNum(m), Gia_ManCoNum(m))
	for id := int32(0); id < Gia_ManObjNum(m); id++ {
		o := Gia_ManObj(m, id)
		kind := 0
		if Gia_ObjIsCi(o) != 0 {
			kind = 1
		} else if Gia_ObjIsCo(o) != 0 {
			kind = 3
		} else if Gia_ObjIsAnd(o) != 0 {
			kind = 2
		}
		f0, f1 := int32(-1), int32(-1)
		if kind >= 2 {
			f0 = Gia_ObjFaninLit0(o, id)
		}
		if kind == 2 {
			f1 = Gia_ObjFaninLit1(o, id)
		}
		fmt.Fprintf(&out, "g %d %d %d %d\n", id, kind, f0, f1)
		Gia_ObjSetValue(o, 100+2*id)
	}
	Gia_ManLevelNum(m)
	for i := int32(0); i < Gia_ManCiNum(m); i++ {
		Gia_ObjSetLevelId(m, Gia_ObjId(m, Gia_ManCi(m, i)), 7+i)
	}
	graph := Gia_ManToAigSkip(m, 3)
	for id := int32(0); id < Gia_ManObjNum(m); id++ {
		fmt.Fprintf(&out, "value %d %d\n", id, Gia_ObjValue(Gia_ManObj(m, id)))
	}
	fmt.Fprintf(&out, "aig %d %d %d\n", aig.Aig_ManObjNum(graph), aig.Aig_ManCiNum(graph), aig.Aig_ManCoNum(graph))
	for id := int32(0); id < aig.Aig_ManObjNumMax(graph); id++ {
		node := aig.Edge(id) << 1
		kind := 0
		if id != 0 {
			if aig.Aig_ObjIsCi(graph, node) {
				kind = 1
			} else if aig.Aig_ObjIsCo(graph, node) {
				kind = 3
			} else if aig.Aig_ObjIsNode(graph, node) {
				kind = 2
			}
		}
		f0, f1 := int32(-1), int32(-1)
		if kind >= 2 {
			f0 = 2*aig.Aig_ObjFaninId0(graph, node) + aig.Aig_ObjFaninC0(graph, node)
		}
		if kind == 2 {
			f1 = 2*aig.Aig_ObjFaninId1(graph, node) + aig.Aig_ObjFaninC1(graph, node)
		}
		fmt.Fprintf(&out, "a %d %d %d %d\n", id, kind, f0, f1)
	}
	for i := int32(0); i < aig.Aig_ManCiNum(graph); i++ {
		fmt.Fprintf(&out, "level %d %d\n", i, aig.Aig_ObjLevel(graph, aig.Aig_ManCi(graph, i)))
	}
	for mask := 0; mask < 8; mask++ {
		in := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0}
		gvals, _ := m.Evaluate(in)
		avals, _ := graph.Evaluate(in)
		fmt.Fprintf(&out, "truth %d", mask)
		for _, v := range gvals {
			fmt.Fprintf(&out, " %d", boolInt(v))
		}
		out.WriteString(" |")
		for _, v := range avals {
			fmt.Fprintf(&out, " %d", boolInt(v))
		}
		out.WriteByte('\n')
	}
	return out.String()
}
func boolInt(v bool) int {
	if v {
		return 1
	}
	return 0
}

func TestChoiceConversionAgainstCompiledABC(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("ABC object oracle requires Linux")
	}
	cc, e := exec.LookPath("gcc")
	if e != nil {
		t.Skip("gcc unavailable")
	}
	ar, e := exec.LookPath("ar")
	if e != nil {
		t.Skip("ar unavailable")
	}
	_, file, _, _ := runtime.Caller(0)
	src := filepath.Clean(filepath.Join(filepath.Dir(file), "../../../src"))
	if _, e = os.Stat(filepath.Join(src, "aig/gia/giaDup.o")); e != nil {
		t.Skip("prebuilt ABC objects unavailable")
	}
	dir := t.TempDir()
	var objs []string
	e = filepath.WalkDir(src, func(path string, d os.DirEntry, e error) error {
		if e != nil {
			return e
		}
		if !d.IsDir() && strings.HasSuffix(path, ".o") && path != filepath.Join(src, "misc/st/st.o") {
			objs = append(objs, path)
		}
		return nil
	})
	if e != nil {
		t.Fatal(e)
	}
	st := filepath.Join(dir, "st.o")
	if out, e := exec.Command(cc, "-DABC_USE_STDINT_H", "-I", src, "-c", filepath.Join(src, "misc/st/st.c"), "-o", st).CombinedOutput(); e != nil {
		t.Fatalf("st: %v %s", e, out)
	}
	objs = append(objs, st)
	lib := filepath.Join(dir, "abc.a")
	if out, e := exec.Command(ar, append([]string{"qcs", lib}, objs...)...).CombinedOutput(); e != nil {
		t.Fatalf("archive: %v %s", e, out)
	}
	exe := filepath.Join(dir, "oracle")
	fixture := filepath.Join(filepath.Dir(file), "test_vectors/choice_conversion_oracle.c")
	if out, e := exec.Command(cc, "-DABC_USE_STDINT_H", "-I", src, fixture, lib, "-lstdc++", "-lm", "-ldl", "-lpthread", "-lreadline", "-o", exe).CombinedOutput(); e != nil {
		t.Fatalf("link: %v %s", e, out)
	}
	c, e := exec.Command(exe).CombinedOutput()
	if e != nil {
		t.Fatalf("oracle: %v %s", e, c)
	}
	goResult := choiceConversionOutput()
	if string(c) != goResult {
		t.Fatalf("C:\n%sGo:\n%s", c, goResult)
	}
}
