package amap

import (
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
)

func TestAmapGraphCopyUsesAllocatedSourceChoices(t *testing.T) {
	source := aig.NewGraph()
	a, b, c := source.NewInput(), source.NewInput(), source.NewInput()
	root := source.AND(source.AND(a, b), c)
	alternative := source.AND(a, source.AND(b, c))
	source.AddOutput(root.Not())
	aig.Aig_ObjSetEquiv(source, root, alternative)
	manager := Amap_ManStart(int32(source.NodeCount()))
	defer Amap_ManStop(manager)
	Amap_ManCreate(manager, source)
	if manager.nChoicesGiven != 1 {
		t.Fatalf("given choices = %d, want 1", manager.nChoicesGiven)
	}
	mappedRoot := manager.sourceCopies[root.NodeID()]
	mappedAlternative := manager.sourceCopies[alternative.NodeID()]
	if mappedRoot == nil || mappedAlternative == nil || Amap_ObjChoice(manager, mappedRoot) != mappedAlternative {
		t.Fatal("source choice was not copied before its representative")
	}
	for assignment := 0; assignment < 8; assignment++ {
		inputs := []bool{assignment&1 != 0, assignment&2 != 0, assignment&4 != 0}
		if got := evalAmapOutput(manager, inputs); got != (assignment != 7) {
			t.Fatalf("assignment %03b: output = %v", assignment, got)
		}
	}
}

func TestAmapGraphCopyPreservesCombinationalFunctions(t *testing.T) {
	source := aig.NewGraph()
	a, b, c := source.NewInput(), source.NewInput(), source.NewInput()
	x := source.NewXor(a, b)
	m := source.Mux(c, x, a.Not())
	root := source.AND(m, b.Not())
	source.AddOutput(root)

	manager := Amap_ManStart(int32(source.NodeCount()))
	manager.fUseXor = 1
	manager.fUseMux = 1
	Amap_ManCreate(manager, source)
	defer Amap_ManStop(manager)

	if got, want := Amap_ManPiNum(manager), int32(3); got != want {
		t.Fatalf("PI count = %d, want %d", got, want)
	}
	if got, want := Amap_ManPoNum(manager), int32(1); got != want {
		t.Fatalf("PO count = %d, want %d", got, want)
	}
	for assignment := 0; assignment < 8; assignment++ {
		inputs := []bool{assignment&1 != 0, assignment&2 != 0, assignment&4 != 0}
		want, err := source.Evaluate(inputs)
		if err != nil {
			t.Fatal(err)
		}
		got := evalAmapOutput(manager, inputs)
		if len(want) != 1 || got != want[0] {
			t.Fatalf("assignment %03b: Amap=%v, source=%v", assignment, got, want)
		}
	}
}

func evalAmapOutput(p *Amap_Man_t, inputs []bool) bool {
	cache := make(map[uint32]bool, len(p.objects))
	seen := make(map[uint32]bool, len(p.objects))
	var evalLit func(int32) bool
	var evalObj func(*Amap_Obj_t) bool
	evalLit = func(lit int32) bool {
		value := evalObj(Amap_ManObj(p, Abc_Lit2Var(lit)))
		return value != (Abc_LitIsCompl(lit) != 0)
	}
	evalObj = func(obj *Amap_Obj_t) bool {
		obj = Amap_Regular(obj)
		if seen[obj.Id] {
			return cache[obj.Id]
		}
		seen[obj.Id] = true
		var value bool
		switch obj.Type {
		case AMAP_OBJ_CONST1:
			value = true
		case AMAP_OBJ_PI:
			value = inputs[obj.IdPio]
		case AMAP_OBJ_PO:
			value = evalLit(obj.Fan[0])
		case AMAP_OBJ_AND:
			value = evalLit(obj.Fan[0]) && evalLit(obj.Fan[1])
		case AMAP_OBJ_XOR:
			value = evalLit(obj.Fan[0]) != evalLit(obj.Fan[1])
		case AMAP_OBJ_MUX:
			if evalLit(obj.Fan[2]) {
				value = evalLit(obj.Fan[1])
			} else {
				value = evalLit(obj.Fan[0])
			}
		default:
			panic("unexpected Amap object kind")
		}
		cache[obj.Id] = value
		return value
	}
	return evalObj(Amap_ManPo(p, 0))
}
