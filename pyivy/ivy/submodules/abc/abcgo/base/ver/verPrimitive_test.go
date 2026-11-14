package ver

import (
	"github.com/glycerine/abcgo/aig/hop"
	"github.com/glycerine/abcgo/base/abc"
	"os"
	"path/filepath"
	"testing"
)

func TestStandardGateFunctions(t *testing.T) {
	for _, tc := range []struct {
		kind  Ver_GateType_t
		unary bool
		truth uint32
	}{{VER_GATE_AND, false, 8}, {VER_GATE_OR, false, 14}, {VER_GATE_XOR, false, 6}, {VER_GATE_NAND, false, 7}, {VER_GATE_NOR, false, 1}, {VER_GATE_XNOR, false, 9}, {VER_GATE_BUF, true, 2}, {VER_GATE_NOT, true, 1}} {
		dir := t.TempDir()
		file := filepath.Join(dir, "gate.v")
		text := "(out,a,b);"
		if tc.unary {
			text = "(out,a);"
		}
		if err := os.WriteFile(file, []byte(text), 0600); err != nil {
			t.Fatal(err)
		}
		p := Ver_ParseStart(file, nil)
		n := Ver_ParseFindOrCreateNetwork(p, "module")
		abc.Abc_NtkFindOrCreateNet(n, "out")
		abc.Abc_NtkFindOrCreateNet(n, "a")
		abc.Abc_NtkFindOrCreateNet(n, "b")
		if Ver_ParseGateStandard(p, n, tc.kind) != 1 {
			t.Fatalf("kind %d parse", tc.kind)
		}
		out := abc.Abc_NtkFindNet(n, "out")
		node := abc.Abc_ObjFanin0(out)
		function := node.Data.(*hop.Hop_Obj_t)
		vars := int32(2)
		mask := uint32(15)
		if tc.unary {
			vars = 1
			mask = 3
		}
		truth := hop.Hop_ManConvertAigToTruth(p.pDesign.FunctionManager, function, vars, nil, 0)[0] & mask
		if truth != tc.truth {
			t.Fatalf("kind %d got %#x want %#x", tc.kind, truth, tc.truth)
		}
		Ver_ParseFreeData(p)
		Ver_ParseStop(p)
	}
}

func TestStandardFlopConnectionsAndInitialValue(t *testing.T) {
	file := filepath.Join(t.TempDir(), "flop.v")
	if err := os.WriteFile(file, []byte("(q,d);"), 0600); err != nil {
		t.Fatal(err)
	}
	p := Ver_ParseStart(file, nil)
	n := Ver_ParseFindOrCreateNetwork(p, "module")
	q := abc.Abc_NtkFindOrCreateNet(n, "q")
	d := abc.Abc_NtkFindOrCreateNet(n, "d")
	if Ver_ParseFlopStandard(p, n) != 1 {
		t.Fatal("flop parse")
	}
	bo := abc.Abc_ObjFanin0(q)
	latch := abc.Abc_ObjFanin0(bo)
	bi := abc.Abc_ObjFanin0(latch)
	if abc.Abc_ObjFanin0(bi) != d || abc.Abc_LatchIsInit0(latch) != 1 {
		t.Fatal("flop connectivity or initial value")
	}
	Ver_ParseFreeData(p)
	Ver_ParseStop(p)
}
