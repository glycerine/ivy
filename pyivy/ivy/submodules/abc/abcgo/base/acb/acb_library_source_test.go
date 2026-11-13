package acb

import (
	"strings"
	"testing"

	"github.com/glycerine/abcgo/map/mio"
)

// acbFunc.c uses adjacent C string literals as a single first array element,
// followed by zero-initialized entries that terminate the traversal.
func TestSourceGateLibraries(t *testing.T) {
	for _, tc := range []struct {
		name  string
		table [25]string
		gates int32
	}{
		{"full", pLibStr, 18},
		{"two-input", pLibStr2, 10},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if tc.table[1] != "" {
				t.Fatal("missing array sentinel")
			}
			if got := strings.Count(tc.table[0], "GATE "); got != int(tc.gates) {
				t.Fatalf("gate lines = %d, want %d", got, tc.gates)
			}
			lib := mio.Mio_LibraryReadBuffer(tc.table[0], 0, nil, 0)
			if lib == nil {
				t.Fatal("source library failed to parse")
			}
			if got := mio.Mio_LibraryReadGateNum(lib); got != tc.gates {
				t.Fatalf("parsed gates = %d, want %d", got, tc.gates)
			}
		})
	}
}
