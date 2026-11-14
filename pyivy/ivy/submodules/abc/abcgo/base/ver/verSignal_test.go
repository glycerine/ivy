package ver

import (
	"github.com/glycerine/abcgo/base/abc"
	"os"
	"path/filepath"
	"testing"
)

func TestParseSignalBusOrderAndSuffix(t *testing.T) {
	for _, tc := range []struct {
		decl   string
		want   []string
		hi, lo int32
	}{{"[3:1] bus, other;", []string{"bus[1]", "bus[2]", "bus[3]", "other[1]", "other[2]", "other[3]"}, 3, 1}, {"[1:3] bus;", []string{"bus[3]", "bus[2]", "bus[1]"}, 1, 3}} {
		file := filepath.Join(t.TempDir(), "signals.v")
		if err := os.WriteFile(file, []byte(tc.decl), 0600); err != nil {
			t.Fatal(err)
		}
		reader := Ver_StreamAlloc(file)
		p := &Ver_Man_t{pReader: reader, Output: os.Stdout}
		n := abc.NewNetwork("module")
		if Ver_ParseSignal(p, n, VER_SIG_INPUT) != 1 {
			t.Fatalf("parse %q", tc.decl)
		}
		pis := n.ObjectsOfKind(abc.ObjectPI)
		if len(pis) != len(tc.want) {
			t.Fatalf("%q: got %d PIs", tc.decl, len(pis))
		}
		for i, name := range tc.want {
			net := abc.Abc_NtkFindNet(n, name)
			if net == nil || abc.Abc_ObjFanin0(net) != pis[i] {
				t.Fatalf("%q: position %d expected %s", tc.decl, i, name)
			}
		}
		var hi, lo int32
		Ver_ParseLookupSuffix(p, "bus", &hi, &lo)
		if hi != tc.hi || lo != tc.lo {
			t.Fatalf("range %d:%d", hi, lo)
		}
		Ver_StreamFree(reader)
		Ver_ParseRemoveSuffixTable(p)
	}
}
