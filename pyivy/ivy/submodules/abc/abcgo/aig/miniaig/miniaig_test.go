package miniaig

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func evalMini(p *Mini_Aig_t, inputs []bool) []bool {
	v := make([]bool, Mini_AigNodeNum(p))
	j := 0
	lit := func(l int) bool { return v[l>>1] != (l&1 != 0) }
	for i := 1; i < len(v); i++ {
		switch {
		case Mini_AigNodeIsPi(p, i):
			v[i] = inputs[j]
			j++
		case Mini_AigNodeIsAnd(p, i):
			v[i] = lit(p.pArray[2*i]) && lit(p.pArray[2*i+1])
		case Mini_AigNodeIsPo(p, i):
			v[i] = lit(p.pArray[2*i])
		}
	}
	return v
}
func TestMiniTruth(t *testing.T) {
	for _, truth := range []uint32{0, 0xffffffff, 0x66666666, 0x88888888, 0x96969696, 0xe8e8e8e8} {
		p := Mini_AigStart()
		vars := []int{Mini_AigCreatePi(p), Mini_AigCreatePi(p), Mini_AigCreatePi(p)}
		out := Mini_AigCreatePo(p, Mini_AigTruth(p, vars, 3, truth))
		if !Mini_AigCheck(p) {
			t.Fatal("invalid graph")
		}
		for a := 0; a < 8; a++ {
			v := evalMini(p, []bool{a&1 != 0, a&2 != 0, a&4 != 0})
			if v[out>>1] != (truth>>a&1 != 0) {
				t.Fatalf("truth %x input %d", truth, a)
			}
		}
	}
}
func TestMiniSerialization(t *testing.T) {
	p := Mini_AigStart()
	a, b := Mini_AigCreatePi(p), Mini_AigCreatePi(p)
	Mini_AigCreatePo(p, Mini_AigXor(p, a, b))
	name := filepath.Join(t.TempDir(), "graph.aig")
	Mini_AigDump(p, name)
	q := Mini_AigLoad(name)
	if q == nil || len(q.pArray) != len(p.pArray) {
		t.Fatal("round trip")
	}
	for i := range p.pArray {
		if p.pArray[i] != q.pArray[i] {
			t.Fatal(i)
		}
	}
	Mini_AigDumpVerilog(name, "top", p)
	s, _ := os.ReadFile(name)
	if !strings.Contains(string(s), "module top") || !strings.Contains(string(s), "assign o") {
		t.Fatal(string(s))
	}
	if Mini_AigLoad(name) != nil {
		t.Fatal("accepted invalid binary")
	}
}
func TestMiniLut(t *testing.T) {
	p := Mini_LutStart(6)
	a, b := Mini_LutCreatePi(p), Mini_LutCreatePi(p)
	n := Mini_LutCreateNode(p, 2, []int{a, b}, []uint32{0x88888888, 0x88888888})
	o := Mini_LutCreatePo(p, n)
	if !Mini_LutCheck(p) || !Mini_LutNodeIsPo(p, o) || Mini_LutNodeTruth(p, n)[0] != 0x88888888 {
		t.Fatal("invalid LUT graph")
	}
	name := filepath.Join(t.TempDir(), "lut.bin")
	Mini_LutDump(p, name)
	q := Mini_LutLoad(name)
	if q == nil || len(q.pArray) != len(p.pArray) {
		t.Fatal("LUT round trip")
	}
	for i, v := range p.pArray {
		if q.pArray[i] != v {
			t.Fatal(i)
		}
	}
	for i, v := range p.pTruths {
		if q.pTruths[i] != v {
			t.Fatal(i)
		}
	}
}
func TestOperatorNames(t *testing.T) {
	if Abc_OperName(ABC_OPER_ARI_ADD) != "+" || Abc_OperName(ABC_OPER_BIT_MUX) != "mux" {
		t.Fatal("operator mapping")
	}
}
func TestNdrRecordsAndRoundTrip(t *testing.T) {
	p := Ndr_Create(1)
	m := Ndr_AddModule(p, 1)
	Ndr_AddObject(p, m, ABC_OPER_CI, 0, 3, 0, 0, 0, nil, 1, []int{2}, nil)
	Ndr_AddObject(p, m, ABC_OPER_CONST, 0, 3, 0, 0, 0, nil, 1, []int{4}, "4'b1010")
	Ndr_AddObject(p, m, ABC_OPER_ARI_ADD, 0, 3, 0, 0, 2, []int{2, 4}, 1, []int{3}, nil)
	Ndr_AddObject(p, m, ABC_OPER_CO, 0, 3, 0, 0, 1, []int{3}, 0, nil, nil)
	mod := m - 256
	if Ndr_DataCiNum(p, mod) != 1 || Ndr_DataCoNum(p, mod) != 1 || Ndr_DataObjNum(p, mod) != 4 {
		t.Fatal("record iteration")
	}
	var b strings.Builder
	Ndr_WriteVerilogModule(&b, p, mod, []string{"", "add10", "a", "s", "const10"})
	for _, want := range []string{"module add10", "input [3:0] a", "assign s = a + const10", "4'b1010"} {
		if !strings.Contains(b.String(), want) {
			t.Fatal(want, b.String())
		}
	}
	name := filepath.Join(t.TempDir(), "net.ndr")
	Ndr_Write(name, p)
	q := Ndr_Read(name)
	if q == nil || len(q.pBody) != len(p.pBody) {
		t.Fatal("read")
	}
	for i, v := range p.pBody {
		if uint32(q.pBody[i]) != uint32(v) || q.pHead[i] != p.pHead[i] {
			t.Fatal(i)
		}
	}
	data, _ := os.ReadFile(name)
	os.WriteFile(name, data[:len(data)-1], 0644)
	if Ndr_Read(name) != nil {
		t.Fatal("accepted truncation")
	}
}
func TestNdrExamples(t *testing.T) {
	t.Chdir(t.TempDir())
	for _, example := range []func(){Ndr_ModuleTest, Ndr_ModuleTestAdder, Ndr_ModuleTestHierarchy, Ndr_ModuleTestMemory, Ndr_ModuleTestFlop, Ndr_ModuleTestSelSel, Ndr_ModuleTestDec, Ndr_ModuleTestAddSub, Ndr_ModuleTestLut} {
		example()
	}
	files, _ := filepath.Glob("*.ndr")
	if len(files) < 8 {
		t.Fatal(files)
	}
	for _, name := range files {
		p := Ndr_Read(name)
		if p == nil {
			t.Fatal(name)
		}
	}
}

func TestMiniAigRejectsMalformedNodes(t *testing.T) {
	for _, data := range [][]int{{0, MINI_AIG_NULL}, {MINI_AIG_NULL, MINI_AIG_NULL, MINI_AIG_NULL, 0}} {
		p := &Mini_Aig_t{pArray: data}
		if Mini_AigCheck(p) {
			t.Fatal("accepted malformed node", data)
		}
	}
}
func TestMiniAigVerilogOutputFanin(t *testing.T) {
	p := Mini_AigStart()
	a := Mini_AigCreatePi(p)
	po := Mini_AigCreatePo(p, a)
	node := Mini_AigAnd(p, po, a)
	Mini_AigCreatePo(p, node)
	name := filepath.Join(t.TempDir(), "output_fanin.v")
	Mini_AigDumpVerilog(name, "top", p)
	text, _ := os.ReadFile(name)
	if strings.Contains(string(text), "n2") || !strings.Contains(string(text), "o2") {
		t.Fatal(string(text))
	}
}
func TestNdrHierarchyWritesOutputConnections(t *testing.T) {
	p := Ndr_Create(1)
	child := Ndr_AddModule(p, 1)
	Ndr_AddObject(p, child, ABC_OPER_CI, 0, 0, 0, 0, 0, nil, 1, []int{2}, nil)
	Ndr_AddObject(p, child, ABC_OPER_CO, 0, 0, 0, 0, 1, []int{2}, 0, nil, nil)
	parent := Ndr_AddModule(p, 3)
	Ndr_AddObject(p, parent, ABC_OPER_CI, 0, 0, 0, 0, 0, nil, 1, []int{4}, nil)
	Ndr_AddObject(p, parent, child, 0, 0, 0, 0, 1, []int{4}, 1, []int{5}, nil)
	Ndr_AddObject(p, parent, ABC_OPER_CO, 0, 0, 0, 0, 1, []int{5}, 0, nil, nil)
	var text strings.Builder
	Ndr_WriteVerilogModule(&text, p, parent-256, []string{"", "child", "x", "parent", "in", "out"})
	if !strings.Contains(text.String(), "child  (in, out);") {
		t.Fatal(text.String())
	}
}
func TestMiniLutRejectsMalformedSentinels(t *testing.T) {
	p := Mini_LutStart(3)
	id := Mini_LutCreatePi(p)
	p.pArray[id*p.LutSize+2] = 0
	if Mini_LutCheck(p) {
		t.Fatal("accepted PI with hidden fanin")
	}
	p = Mini_LutStart(3)
	a := Mini_LutCreatePi(p)
	id = Mini_LutCreateNode(p, 1, []int{a}, []uint32{0xaaaaaaaa})
	p.pArray[id*p.LutSize+2] = MINI_LUT_NULL2
	if Mini_LutCheck(p) {
		t.Fatal("accepted misplaced PO sentinel")
	}
}
func TestNdrRejectsContainerCrossing(t *testing.T) {
	p := Ndr_Create(1)
	child := Ndr_AddModule(p, 2)
	Ndr_AddObject(p, child, ABC_OPER_CI, 0, 0, 0, 0, 0, nil, 1, []int{3}, nil)
	next := Ndr_AddModule(p, 4)
	p.pBody[child-256+2] = len(p.pBody) - (child - 256 + 2)
	name := filepath.Join(t.TempDir(), "invalid.ndr")
	Ndr_Write(name, p)
	if Ndr_Read(name) != nil {
		t.Fatal("object crossed its module boundary", next)
	}
}
