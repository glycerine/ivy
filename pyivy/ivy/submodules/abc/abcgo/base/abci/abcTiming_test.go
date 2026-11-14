package abci

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	baseabc "github.com/glycerine/abcgo/base/abc"
)

func TestTimingManagerArrivalRequiredAndPrint(t *testing.T) {
	n := baseabc.NewNetwork("timing")
	pi := baseabc.Abc_NtkCreatePi(n)
	pi.SetName("input")
	po := n.NewObject(baseabc.ObjectPO)
	if err := n.AddFanin(po, pi.Signal()); err != nil {
		t.Fatal(err)
	}
	Abc_NtkTimeSetDefaultArrival(n, 1, 2)
	Abc_NtkTimeSetDefaultRequired(n, 8, 9)
	Abc_NtkTimeSetArrival(n, int32(pi.ID), 3, 4)
	Abc_NtkTimeSetRequired(n, int32(po.ID), 5, 6)
	if got := Abc_NodeReadArrivalWorst(pi); got != 4 {
		t.Fatalf("arrival worst=%v", got)
	}
	if got := Abc_NodeReadRequiredWorst(po); got != 5 {
		t.Fatalf("required worst=%v", got)
	}
	Abc_NtkTimeSetInputDrive(n, 0, 7, 8)
	Abc_NtkTimeSetOutputLoad(n, 0, 9, 10)
	if got := Abc_NodeReadInputDriveWorst(n, 0); got != 8 {
		t.Fatalf("input drive=%v", got)
	}
	if got := Abc_NodeReadOutputLoadWorst(n, 0); got != 10 {
		t.Fatalf("output load=%v", got)
	}

	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w
	Abc_NtkTimePrint(n)
	_ = w.Close()
	os.Stdout = old
	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	_ = r.Close()
	for _, want := range []string{"Default arrival", "Default required", "Inputs (1)", "Outputs (1)", "input", "arrival = 4.000000", "required = 5.000000"} {
		if !strings.Contains(buf.String(), want) {
			t.Fatalf("timing print missing %q: %s", want, buf.String())
		}
	}
}

func TestTimingManagerDupAndNetworkDelete(t *testing.T) {
	old := baseabc.NewNetwork("old")
	pi := baseabc.Abc_NtkCreatePi(old)
	po := old.NewObject(baseabc.ObjectPO)
	_ = old.AddFanin(po, pi.Signal())
	Abc_NtkTimeSetDefaultArrival(old, 1, 2)
	Abc_NtkTimeSetDefaultRequired(old, 3, 4)
	Abc_NtkTimeSetArrival(old, int32(pi.ID), 5, 6)
	Abc_NtkTimeSetRequired(old, int32(po.ID), 7, 8)
	Abc_NtkTimeSetInputDrive(old, 0, 9, 10)
	Abc_NtkTimeSetOutputLoad(old, 0, 11, 12)
	dup := baseabc.NewNetwork("dup")
	pi2 := baseabc.Abc_NtkCreatePi(dup)
	po2 := dup.NewObject(baseabc.ObjectPO)
	_ = dup.AddFanin(po2, pi2.Signal())
	Abc_ManTimeDup(old, dup)
	if got := Abc_NodeReadArrivalWorst(pi2); got != 6 {
		t.Fatalf("duplicated arrival=%v", got)
	}
	if got := Abc_NodeReadRequiredWorst(po2); got != 7 {
		t.Fatalf("duplicated required=%v", got)
	}
	if got := Abc_NodeReadInputDriveWorst(dup, 0); got != 10 {
		t.Fatalf("duplicated drive=%v", got)
	}
	if got := Abc_NodeReadOutputLoadWorst(dup, 0); got != 12 {
		t.Fatalf("duplicated load=%v", got)
	}
	manager := baseabc.NetworkTimingManager(dup)
	baseabc.Abc_NtkDelete(dup)
	if baseabc.NetworkTimingManager(dup) != nil {
		t.Fatal("network deletion left timing manager attached")
	}
	if m := manager.(*Abc_ManTime_t); !m.stopped {
		t.Fatal("network deletion did not stop timing manager")
	}
}

func TestTimingFloatArraysWithoutManagerAndWithDefaults(t *testing.T) {
	n := baseabc.Abc_NtkAlloc(baseabc.ABC_NTK_STRASH, baseabc.ABC_FUNC_AIG, 1)
	defer baseabc.Abc_NtkDelete(n)
	input := baseabc.Abc_NtkCreatePi(n)
	baseabc.Abc_ObjAddFanin(baseabc.Abc_NtkCreatePo(n), input)
	if got := Abc_NtkGetCiArrivalFloats(n); len(got) != 1 || got[0] != 0 {
		t.Fatalf("default arrival %v", got)
	}
	if got := Abc_NtkGetCoRequiredFloats(n); got != nil {
		t.Fatalf("absent required times %v", got)
	}
	Abc_NtkTimeSetDefaultArrival(n, 2, 6)
	Abc_NtkTimeSetDefaultRequired(n, 8, 12)
	if got := Abc_NtkReadDefaultArrivalWorst(n); got != 4 {
		t.Fatalf("C default arrival average=%g", got)
	}
	if got := Abc_NtkReadDefaultRequiredWorst(n); got != 10 {
		t.Fatalf("C default required average=%g", got)
	}
	if got := Abc_NtkGetCiArrivalFloats(n); len(got) != 1 || got[0] != 6 {
		t.Fatalf("arrival %v", got)
	}
	if got := Abc_NtkGetCoRequiredFloats(n); len(got) != 1 || got[0] != 8 {
		t.Fatalf("required %v", got)
	}
}
