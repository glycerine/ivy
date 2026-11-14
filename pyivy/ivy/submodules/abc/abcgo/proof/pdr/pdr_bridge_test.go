package pdr

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/util"
)

func TestPdrBridgeReportsEachOutputAndAbort(t *testing.T) {
	for _, incremental := range []bool{false, true} {
		g := aig.NewGraph()
		g.AddOutput(g.Const0())
		g.AddOutput(g.Const1())
		aig.Aig_ManSetRegNum(g, 0)
		var pars Pdr_Par_t
		Pdr_ManSetDefaultParams(&pars)
		pars.FSilent, pars.FSolveAll, pars.FStoreCex, pars.FUseBridge = 1, 1, 1, 1
		pars.FNotVerbose = 1
		file, err := os.CreateTemp(t.TempDir(), "bridge")
		if err != nil {
			t.Fatal(err)
		}
		old := os.Stdout
		os.Stdout = file
		var status int32
		func() {
			defer func() { os.Stdout = old }()
			if incremental {
				status = IPdr_ManSolve(g, &pars)
			} else {
				status = Pdr_ManSolve(g, &pars)
			}
		}()
		if err := file.Close(); err != nil {
			t.Fatal(err)
		}
		data, err := os.ReadFile(file.Name())
		if err != nil {
			t.Fatal(err)
		}
		reader := bytes.NewReader(data)
		var kinds []int32
		var payloads [][]byte
		for reader.Len() != 0 {
			kind, payload, err := util.ReadBridgePackage(reader)
			if err != nil && err != io.EOF {
				t.Fatalf("incremental=%v: malformed bridge stream %q: %v", incremental, data, err)
			}
			kinds, payloads = append(kinds, kind), append(payloads, payload)
		}
		if status != 0 || len(kinds) != 3 || kinds[0] != util.BRIDGE_RESULTS || kinds[1] != util.BRIDGE_RESULTS || kinds[2] != util.BRIDGE_ABORT {
			t.Fatalf("incremental=%v: status=%d types=%v", incremental, status, kinds)
		}
		if !bytes.Equal(payloads[0], []byte{util.BRIDGE_VALUE_1, 1, 0, 0}) || !bytes.Equal(payloads[2], []byte("timeout")) {
			t.Fatalf("incremental=%v: payloads=%v", incremental, payloads)
		}
		if len(payloads[1]) < 4 || payloads[1][0] != util.BRIDGE_VALUE_0 || payloads[1][2] != 1 {
			t.Fatalf("incremental=%v: failed-output payload=%v", incremental, payloads[1])
		}
	}
}
