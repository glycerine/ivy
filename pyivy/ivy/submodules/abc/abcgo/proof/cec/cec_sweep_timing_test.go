package cec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"io"
	"os"
	"strings"
	"testing"
)

func TestCecSweeperVerboseStatistics(t *testing.T) {
	for _, engine := range []int{2, 3} {
		g := gia.Gia_ManStart(8)
		x := gia.Gia_ManAppendCi(g)
		gia.Gia_ManAppendCo(g, x)
		read, write, err := os.Pipe()
		if err != nil {
			t.Fatal(err)
		}
		old := os.Stdout
		os.Stdout = write
		if engine == 2 {
			cec2SweepEquivalences(g, cec2SweepSettings{words: 1, iterations: 1, verbose: true})
		} else {
			cec3SweepEquivalences(g, cec3SweepSettings{words: 1, iterations: 1, verbose: true})
		}
		write.Close()
		os.Stdout = old
		data, err := io.ReadAll(read)
		read.Close()
		if err != nil {
			t.Fatal(err)
		}
		text := string(data)
		if !strings.Contains(text, "SAT solving") || !strings.Contains(text, "S =") || !strings.Contains(text, "TOTAL") {
			t.Fatalf("engine %d missing stats: %q", engine, text)
		}
	}
}
