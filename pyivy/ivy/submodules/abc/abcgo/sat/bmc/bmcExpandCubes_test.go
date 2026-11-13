package bmc

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestExpandCubes(t *testing.T) {
	g := gia.Gia_ManStart(16)
	a := gia.Gia_ManAppendCi(g)
	gia.Gia_ManAppendCi(g)
	gia.Gia_ManAppendCo(g, a)
	raw := []byte("11 1\n\x00")
	sop := vec.Vec_StrAllocArrayCopy(raw, len(raw))
	if got := Abc_ObjExpandCubes(sop, g, 2); got != 0 || string(vec.Vec_StrArray(sop)) != "1- 1\n\x00" {
		t.Fatalf("status=%d sop=%q", got, vec.Vec_StrArray(sop))
	}
}
func TestExpandConstant(t *testing.T) {
	g := gia.Gia_ManStart(4)
	gia.Gia_ManAppendCo(g, 1)
	raw := []byte(" 0\n\x00")
	sop := vec.Vec_StrAllocArrayCopy(raw, len(raw))
	if got := Abc_ObjExpandCubes(sop, g, 0); got != 1 || string(vec.Vec_StrArray(sop)) != " 1\n\x00" {
		t.Fatalf("status=%d sop=%q", got, vec.Vec_StrArray(sop))
	}
}

func TestNetworkExpandCubes(t *testing.T) {
	n := abc.Abc_NtkAlloc(abc.ABC_NTK_LOGIC, abc.ABC_FUNC_SOP, 1)
	a, b := n.NewObject(abc.ObjectPI), n.NewObject(abc.ObjectPI)
	node := n.NewObject(abc.ObjectNode)
	node.Data = "11 1\n"
	abc.Abc_ObjAddFanin(node, a)
	abc.Abc_ObjAddFanin(node, b)
	out := n.NewObject(abc.ObjectPO)
	abc.Abc_ObjAddFanin(out, node)
	g := gia.Gia_ManStart(16)
	x := gia.Gia_ManAppendCi(g)
	gia.Gia_ManAppendCi(g)
	gia.Gia_ManAppendCo(g, x)
	Abc_NtkExpandCubes(n, g, 0)
	if got := node.Data.(string); got != "1- 1\n" {
		t.Fatalf("sop %q", got)
	}
}

func TestNetworkExpandConstantRemovesFanins(t *testing.T) {
	n := abc.Abc_NtkAlloc(abc.ABC_NTK_LOGIC, abc.ABC_FUNC_SOP, 1)
	a := n.NewObject(abc.ObjectPI)
	node := n.NewObject(abc.ObjectNode)
	node.Data = "1 1\n"
	abc.Abc_ObjAddFanin(node, a)
	out := n.NewObject(abc.ObjectPO)
	abc.Abc_ObjAddFanin(out, node)
	g := gia.Gia_ManStart(8)
	gia.Gia_ManAppendCi(g)
	gia.Gia_ManAppendCo(g, 1)
	Abc_NtkExpandCubes(n, g, 0)
	if abc.Abc_ObjFaninNum(node) != 0 || node.Data.(string) != " 1\n" {
		t.Fatalf("fanins=%d sop=%q", abc.Abc_ObjFaninNum(node), node.Data)
	}
}
