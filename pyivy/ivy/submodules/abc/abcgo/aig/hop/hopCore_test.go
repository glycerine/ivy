package hop

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestComplementedEdgesAndConstructors(t *testing.T) {
	p := Hop_ManStart()
	defer Hop_ManStop(p)
	a, b := Hop_IthVar(p, 0), Hop_IthVar(p, 1)
	if Hop_Not(Hop_Not(a)) != a || Hop_Regular(Hop_Not(a)) != a || Hop_Not(a) != Hop_Not(a) {
		t.Fatal("unstable complemented edge")
	}
	if Hop_And(p, a, b) != Hop_And(p, b, a) {
		t.Fatal("structural sharing")
	}
	if Hop_And(p, a, Hop_Not(a)) != Hop_ManConst0(p) || Hop_And(p, a, a) != a {
		t.Fatal("AND reduction")
	}
	cases := []struct {
		edge *Hop_Obj_t
		want uint64
	}{{Hop_And(p, a, b), 0x8888888888888888}, {Hop_Or(p, a, b), 0xeeeeeeeeeeeeeeee}, {Hop_Exor(p, a, b), 0x6666666666666666}, {Hop_Mux(p, a, b, Hop_Not(b)), 0x9999999999999999}}
	for _, c := range cases {
		if got := Hop_ManComputeTruth6(p, c.edge, 2); got != c.want {
			t.Fatalf("got %x want %x", got, c.want)
		}
	}
	if Hop_ManCheck(p) != 1 {
		t.Fatal("inconsistent graph")
	}
}
func TestStructuralHashResizeAndCleanup(t *testing.T) {
	p := Hop_ManStart()
	defer Hop_ManStop(p)
	p.nTableSize = 3
	p.pTable = make([]*Hop_Obj_t, 3)
	var nodes []*Hop_Obj_t
	for i := int32(0); i < 30; i++ {
		a, b := Hop_IthVar(p, i), Hop_IthVar(p, i+1)
		nodes = append(nodes, Hop_And(p, a, b))
	}
	if p.nTableSize <= 3 || Hop_TableCountEntries(p) != 30 {
		t.Fatal("resize lost nodes")
	}
	Hop_ObjCreatePo(p, nodes[0])
	if removed := Hop_ManCleanup(p); removed != 29 {
		t.Fatal("cleanup", removed)
	}
	if Hop_ManCheck(p) != 1 {
		t.Fatal("invalid graph after cleanup")
	}
}
func TestEquationAdapter(t *testing.T) {
	p := Hop_ManStart()
	defer Hop_ManStop(p)
	edge, err := CompileEquation(p, "a * !b + !a * b", []string{"a", "b"})
	if err != nil {
		t.Fatal(err)
	}
	if got := Hop_ManComputeTruth6(p, edge, 2); got != 0x6666666666666666 {
		t.Fatalf("equation truth %x", got)
	}
	var errors bytes.Buffer
	if ParseEquation(p, &errors, "a + (", []string{"a"}) != nil || errors.Len() == 0 {
		t.Fatal("missing parse failure")
	}
}
func TestMuxRecognitionAndBlif(t *testing.T) {
	p := Hop_ManStart()
	defer Hop_ManStop(p)
	a, b, c := Hop_IthVar(p, 0), Hop_IthVar(p, 1), Hop_IthVar(p, 2)
	edge := Hop_Mux(p, c, a, b)
	var then, otherwise *Hop_Obj_t
	control := Hop_ObjRecognizeMux(Hop_Regular(edge), &then, &otherwise)
	recognized := Hop_Mux(p, control, then, otherwise)
	if Hop_IsComplement(edge) != 0 {
		recognized = Hop_Not(recognized)
	}
	if Hop_ManComputeTruth6(p, recognized, 3) != Hop_ManComputeTruth6(p, edge, 3) {
		t.Fatal("MUX recognition changed truth")
	}
	Hop_ObjCreatePo(p, edge)
	var output bytes.Buffer
	if err := WriteBlif(p, &output); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), ".inputs n1 n2 n3\n") || !strings.HasSuffix(output.String(), ".end\n") {
		t.Fatal(output.String())
	}
}
func TestCuddOperationRecording(t *testing.T) {
	t.Chdir(t.TempDir())
	Cudd2_Init(2, 0, 0, 0, 0, nil)
	defer func() {
		if s_pCuddMan != nil {
			Cudd2_Quit(nil)
		}
	}()
	Cudd2_bddIthVar(nil, 0, uintptr(2))
	Cudd2_bddIthVar(nil, 1, uintptr(4))
	Cudd2_bddXor(nil, uintptr(2), uintptr(4), uintptr(6))
	root := Cudd2_GetArg(uintptr(6))
	if got := Hop_ManComputeTruth6(s_pCuddMan.pAig, root, 2); got != 0x6666666666666666 {
		t.Fatalf("recorded truth %x", got)
	}
	if Cudd2_GetArg(uintptr(7)) != Hop_Not(root) {
		t.Fatal("BDD complement missing")
	}
	Cudd2_bddEqual(nil, uintptr(2), uintptr(4), 0)
	Cudd2_Quit(nil)
	if _, err := os.Stat("aig_temp.blif"); err != nil {
		t.Fatal(err)
	}
}

func TestTraversalIDReset(t *testing.T) {
	p := Hop_ManStart()
	defer Hop_ManStop(p)
	o := Hop_IthVar(p, 0)
	Hop_ObjSetTravIdCurrent(p, o)
	if Hop_ObjIsTravIdCurrent(p, o) != 1 {
		t.Fatal("traversal mark missing")
	}
	p.nTravIds = (1 << 30) - 1
	Hop_ManIncrementTravId(p)
	if p.nTravIds != 2 || o.travId != 0 {
		t.Fatal("traversal ID wrap", p.nTravIds, o.travId)
	}
}
