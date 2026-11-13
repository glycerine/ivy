package cec

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/glycerine/abcgo/aig/gia"
)

func TestVerifyOriginalABC(t *testing.T) {
	_, file, _, _ := runtime.Caller(0)
	binary := filepath.Clean(filepath.Join(filepath.Dir(file), "../../../abc"))
	if _, err := os.Stat(binary); err != nil {
		t.Skip("original ABC executable unavailable")
	}
	for _, unequal := range []bool{false, true} {
		g := gia.Gia_ManStart(24)
		a, b, c := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
		x := gia.Gia_ManAppendAnd(g, gia.Gia_ManAppendAnd(g, a, b), c)
		y := gia.Gia_ManAppendAnd(g, a, gia.Gia_ManAppendAnd(g, b, c))
		gia.Gia_ManAppendCo(g, x)
		if unequal {
			gia.Gia_ManAppendCo(g, gia.Gia_ManAppendAnd(g, a, b))
		} else {
			gia.Gia_ManAppendCo(g, y)
		}
		data, err := gia.MarshalAiger(g, gia.AigerWriteOptions{})
		if err != nil {
			t.Fatal(err)
		}
		path := filepath.Join(t.TempDir(), "input.aig")
		if err = os.WriteFile(path, data, 0600); err != nil {
			t.Fatal(err)
		}
		cmd := fmt.Sprintf("&r -s %s; &cec -m -d", path)
		log, err := exec.Command(binary, "-c", cmd).CombinedOutput()
		if err != nil {
			t.Fatalf("ABC: %v %s", err, log)
		}
		text := string(log)
		abcStatus := int32(-1)
		if strings.Contains(text, "Networks are NOT EQUIVALENT") {
			abcStatus = 0
		} else if strings.Contains(text, "Networks are equivalent") {
			abcStatus = 1
		}
		var pars Cec_ParCec_t
		Cec_ManCecSetDefaultParams(&pars)
		pars.fSilent = 1
		status := Cec_ManVerify(g, &pars)
		if status != abcStatus || pars.Error != nil {
			t.Fatalf("unequal %v: Go %d (%v), ABC %d: %s", unequal, status, pars.Error, abcStatus, text)
		}
		if status == 0 {
			w := g.CombinationalCounterexample()
			if w == nil {
				t.Fatal("missing witness")
			}
			in := []bool{w.PData[0]&1 != 0, w.PData[0]&2 != 0, w.PData[0]&4 != 0}
			out, err := g.Evaluate(in)
			if err != nil || out[0] == out[1] {
				t.Fatalf("invalid witness %v outputs %v error %v", in, out, err)
			}
		}
	}
}
