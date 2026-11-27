package if_

import (
	"bytes"
	"io"
	"os"
	"testing"
)

// If_CluHashPrintStats lists busy buckets in bin order, as ifDec16.c does;
// the port iterated its map, so the order changed from run to run.
func TestCluHashPrintStatsBinOrder(t *testing.T) {
	p := &If_Man_t{}
	p.pHashTable[0] = map[uint32]any{}
	for _, bin := range []uint32{41, 3, 77, 15, 60, 8, 99, 23} {
		var head *If_Hte_t
		for i := 0; i < 10+int(bin%3); i++ {
			head = &If_Hte_t{pNext: head}
		}
		p.pHashTable[0][bin] = head
	}
	want := "3=10 8=12 15=10 23=12 41=12 60=10 77=12 99=10 "
	for run := 0; run < 20; run++ {
		r, w, _ := os.Pipe()
		stdout := os.Stdout
		os.Stdout = w
		If_CluHashPrintStats(p, 0)
		os.Stdout = stdout
		w.Close()
		var out bytes.Buffer
		io.Copy(&out, r)
		if out.String() != want {
			t.Fatalf("run %d: %q, want %q", run, out.String(), want)
		}
	}
}
