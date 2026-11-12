package glucose

import (
	"bytes"
	abcmain "github.com/glycerine/abcgo/base/main"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLegacyCommandIntegration(t *testing.T) {
	f := abcmain.NewFrame()
	var out, errs bytes.Buffer
	f.Out = &out
	f.Err = &errs
	Glucose_Init(f)
	command, ok := f.LookupCommand("&glucose")
	if !ok {
		t.Fatal("command not registered")
	}
	g, _ := truthGraph(0x96)
	f.SetResource("pGia", g)
	if command.Handler(f, []string{"&glucose", "-p"}) != 0 || !strings.Contains(out.String(), "SATISFIABLE") || f.Cex == nil {
		t.Fatal(out.String(), errs.String())
	}
	bits := 0
	for i := 0; i < 3; i++ {
		if f.Cex.Bits[0]&(1<<i) != 0 {
			bits |= 1 << i
		}
	}
	if uint8(0x96)&(1<<bits) == 0 {
		t.Fatal("command produced invalid witness")
	}
	before := f.Cex
	name := filepath.Join(t.TempDir(), "input.cnf")
	if err := os.WriteFile(name, []byte("p cnf 1 1\n1 0\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if command.Handler(f, []string{"&glucose", name}) != 0 || f.Cex != before {
		t.Fatal("CNF-file command changed graph witness")
	}
	for _, args := range [][]string{{"&glucose", "-C"}, {"&glucose", "-C", "-1"}, {"&glucose", "-h"}, {"&glucose", "-unknown"}} {
		if command.Handler(f, args) != 1 {
			t.Fatal("invalid options accepted", args)
		}
	}
	Glucose_End(f)
}
