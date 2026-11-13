package ssw

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestClassAddCollectRemove(t *testing.T) {
	g := aig.NewGraph()
	a, b, c := g.NewInput(), g.NewInput(), g.NewInput()
	p := Ssw_ClassesStart(g)
	Ssw_ObjAddClass(p, a, []aig.Edge{a, b, c}, 3)
	if Ssw_ClassesClassNum(p) != 1 || Ssw_ClassesLitNum(p) != 2 {
		t.Fatal("wrong class statistics")
	}
	var size int32
	class := Ssw_ClassesReadClass(p, a, &size)
	if size != 3 || class[1] != b {
		t.Fatal("wrong class contents")
	}
	collected := vec.Vec_PtrAlloc(0)
	Ssw_ClassesCollectClass(p, a, collected)
	if vec.Vec_PtrSize(collected) != 2 || vec.Vec_PtrEntry(collected, 0) != b || vec.Vec_PtrEntry(collected, 1) != c {
		t.Fatal("wrong collected class")
	}
	if got := Ssw_ObjRemoveClass(p, a); len(got) != 3 || Ssw_ClassesClassNum(p) != 0 || Ssw_ClassesLitNum(p) != 0 {
		t.Fatal("class removal failed")
	}
	size = 99
	if Ssw_ClassesReadClass(p, a, &size) != nil || size != 99 {
		t.Fatal("absent class changed size")
	}
}
