package dau

import (
	"encoding/binary"
	"fmt"
	"github.com/glycerine/abcgo/internal/testoracle"
	"path/filepath"
	"testing"
)

func TestFunctionEnumerationOriginalABC(t *testing.T) {
	abc := testoracle.NativeABCPath()
	if _, err := testoracle.Stat(abc); err != nil {
		t.Skip("original ABC unavailable")
	}
	for _, test := range []struct {
		inputs, vars, nodes int32
		two, reduce         bool
	}{{3, 3, 3, false, false}, {3, 3, 3, true, false}, {4, 3, 3, false, false}, {4, 4, 3, false, true}} {
		t.Run(fmt.Sprintf("%d_%d_%d_%v_%v", test.inputs, test.vars, test.nodes, test.two, test.reduce), func(t *testing.T) {
			dir := t.TempDir()
			flags := ""
			if test.two {
				flags += " -t"
			}
			if test.reduce {
				flags += " -r"
			}
			command := testoracle.Command(t, abc, "-c", fmt.Sprintf("funenum -S %d -I %d -M %d%s", test.inputs, test.vars, test.nodes, flags))
			command.Dir = dir
			if out, err := command.CombinedOutput(); err != nil {
				t.Fatalf("original function enumeration%v\n%s", err, out)
			}
			data, err := testoracle.ReadFile(filepath.Join(dir, fmt.Sprintf("npn%d%d.ttd", test.inputs, test.vars)))
			if err != nil {
				t.Fatal(err)
			}
			got, err := EnumerateFunctions(test.inputs, test.vars, test.nodes, test.two, test.reduce)
			if err != nil {
				t.Fatal(err)
			}
			if len(data) != 12*len(got.Truths) {
				t.Fatalf("source%d vsGo%d classes", len(data)/12, len(got.Truths))
			}
			for i, truth := range got.Truths {
				want := binary.LittleEndian.Uint64(data[8*i:])
				if truth != want {
					t.Fatalf("entry%d truth%x want%x", i, truth, want)
				}
			}
			offset := 8 * len(got.Truths)
			for i, entry := range got.NodeSupport {
				if entry != int32(binary.LittleEndian.Uint32(data[offset+4*i:])) {
					t.Fatalf("entry%d node/support metadata", i)
				}
			}
		})
	}
}
