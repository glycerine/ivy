package abci

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/aig/hop"
	"github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/misc/vec"
)

func dress3AIG(name string, inputNames []string, negateSecond bool) *abc.Network {
	n := abc.Abc_NtkAlloc(abc.ABC_NTK_LOGIC, abc.ABC_FUNC_AIG, 1)
	n.Name = name
	for _, inputName := range inputNames {
		pi := n.NewObject(abc.ObjectPI)
		pi.SetName(inputName)
	}
	manager := n.FunctionManager.(*hop.Hop_Man_t)
	for i := range inputNames {
		hop.Hop_IthVar(manager, int32(i))
	}
	left, right := hop.Hop_IthVar(manager, 0), hop.Hop_IthVar(manager, 1)
	if negateSecond {
		right = hop.Hop_Not(right)
	}
	function := hop.Hop_And(manager, left, right)
	node := n.NewObject(abc.ObjectNode)
	node.Data = function
	for _, pi := range n.Inputs() {
		abc.Abc_ObjAddFanin(node, pi)
	}
	co := n.NewObject(abc.ObjectPO)
	abc.Abc_ObjAddFanin(co, node)
	return n
}

func TestConvertHopToGiaConvertsComplementedSharedDagAndClearsMarks(t *testing.T) {
	manager := hop.Hop_ManStart()
	defer hop.Hop_ManStop(manager)
	a, b := hop.Hop_IthVar(manager, 0), hop.Hop_IthVar(manager, 1)
	shared := hop.Hop_And(manager, a, hop.Hop_Not(b))
	root := hop.Hop_And(manager, shared, hop.Hop_Not(a))
	graph := gia.Gia_ManStart(32)
	defer gia.Gia_ManStop(graph)
	gia.Gia_ManAppendCi(graph)
	gia.Gia_ManAppendCi(graph)
	gia.Gia_ManHashAlloc(graph)
	hop.Hop_ObjSetIData(a, gia.Gia_ManCiLit(graph, 0))
	hop.Hop_ObjSetIData(b, gia.Gia_ManCiLit(graph, 1))
	lit := Abc_ConvertHopToGia(graph, root)
	gia.Gia_ManHashStop(graph)
	if gia.Gia_ManAndNum(graph) != 2 || lit>>1 == 0 {
		t.Fatalf("converted graph has %d ANDs and root literal %d, want 2 ANDs", gia.Gia_ManAndNum(graph), lit)
	}
	for _, object := range hop.Hop_ManObjects(manager) {
		if hop.Hop_ObjIsMarkA(object) != 0 {
			t.Fatalf("temporary mark remains on HOP object %d", object.Id)
		}
	}
}

func TestAigToGiaTwoSharesNamedInputsAndPreservesComplement(t *testing.T) {
	first := dress3AIG("first", []string{"shared", "left"}, false)
	second := dress3AIG("second", []string{"shared", "right"}, true)
	graph := Abc_NtkAigToGiaTwo(first, second, 1)
	defer gia.Gia_ManStop(graph)
	if got := gia.Gia_ManPiNum(graph); got != 3 {
		t.Fatalf("shared manager has %d PIs, want 3", got)
	}
	if got := gia.Gia_ManPoNum(graph); got != 2 {
		t.Fatalf("shared manager has %d POs, want 2", got)
	}
	if got := gia.Gia_ManAndNum(graph); got != 2 {
		t.Fatalf("shared manager has %d ANDs, want 2", got)
	}
	firstOutput := gia.Gia_ObjFaninLit0(gia.Gia_ManPo(graph, 0), gia.Gia_ObjId(graph, gia.Gia_ManPo(graph, 0)))
	secondOutput := gia.Gia_ObjFaninLit0(gia.Gia_ManPo(graph, 1), gia.Gia_ObjId(graph, gia.Gia_ManPo(graph, 1)))
	if firstOutput == secondOutput {
		t.Fatalf("outputs unexpectedly share the same literal %d", firstOutput)
	}
	for _, test := range []struct {
		inputs []bool
		want   [2]bool
	}{
		{[]bool{true, false, false}, [2]bool{false, true}},
		{[]bool{true, true, false}, [2]bool{true, true}},
		{[]bool{false, true, true}, [2]bool{false, false}},
	} {
		for output := int32(0); output < 2; output++ {
			co := gia.Gia_ManPo(graph, output)
			lit := gia.Gia_ObjFaninLit0(co, gia.Gia_ObjId(graph, co))
			if got := evalGiaLiteral(graph, lit, test.inputs); got != test.want[output] {
				t.Fatalf("inputs %v output %d = %v, want %v", test.inputs, output, got, test.want[output])
			}
		}
	}
}

func evalGiaLiteral(graph *Gia_Man_t, lit int32, inputs []bool) bool {
	id := lit >> 1
	var value bool
	if id == 0 {
		value = false
	} else {
		obj := gia.Gia_ManObj(graph, id)
		switch {
		case gia.Gia_ObjIsCi(obj) != 0:
			value = inputs[gia.Gia_ObjCioId(obj)]
		case gia.Gia_ObjIsAnd(obj) != 0:
			value = evalGiaLiteral(graph, gia.Gia_ObjFaninLit0(obj, id), inputs) && evalGiaLiteral(graph, gia.Gia_ObjFaninLit1(obj, id), inputs)
		default:
			panic("unexpected GIA object in literal evaluation")
		}
	}
	if lit&1 != 0 {
		return !value
	}
	return value
}

func TestDumpEquivFileWritesPackedClassesAndPolarity(t *testing.T) {
	first := abc.Abc_NtkAlloc(abc.ABC_NTK_LOGIC, abc.ABC_FUNC_AIG, 1)
	first.Name = "net_a"
	a := first.NewObject(abc.ObjectNode)
	a.SetName("node_a")
	second := abc.Abc_NtkAlloc(abc.ABC_NTK_LOGIC, abc.ABC_FUNC_AIG, 1)
	second.Name = "net_b"
	b := second.NewObject(abc.ObjectNode)
	b.SetName("node_b")
	classes := vec.Vec_IntAlloc(3)
	vec.Vec_IntPush(classes, 2)
	vec.Vec_IntPush(classes, abc.Abc_ObjId(a)<<2)
	vec.Vec_IntPush(classes, abc.Abc_ObjId(b)<<2|1|2)
	path := filepath.Join(t.TempDir(), "equiv.txt")
	Abc_NtkDumpEquivFile(path, classes, []*Abc_Ntk_t{first, second})
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	for _, line := range []string{"1:net_a:node_a", "1:net_b:NOT:node_b"} {
		if !strings.Contains(text, line) {
			t.Fatalf("equivalence file lacks %q:\n%s", line, text)
		}
	}
}
