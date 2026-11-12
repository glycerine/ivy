package gia

import (
	"github.com/glycerine/abcgo/aig/ioa"
	abcmain "github.com/glycerine/abcgo/base/main"
	"testing"
)

type testFrameWitness struct {
	bits                             []int32
	registers, inputs, frame, output int
}

func (w testFrameWitness) CounterexampleData() ([]int32, int, int, int, int) {
	return w.bits, w.registers, w.inputs, w.frame, w.output
}

func TestFrameBridgeGraphAndWitnessOwnership(t *testing.T) {
	frame := abcmain.NewFrame()
	InstallFrameBridge(frame)
	p := Gia_ManStart(8)
	pi, ro := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	Gia_ManAppendCo(p, ro)
	Gia_ManAppendCo(p, pi)
	Gia_ManSetRegNum(p, 1)
	frame.SetResource("pGia", p)
	exported, err := frame.Bridge.Export(frame, "gia-aiger")
	if err != nil {
		t.Fatal(err)
	}
	provider, ok := exported.(ioa.AigerProvider)
	if !ok {
		t.Fatal("export is not AIGER")
	}
	model, err := provider.AigerModel()
	if err != nil {
		t.Fatal(err)
	}
	if err := frame.Bridge.Import(frame, "gia-aiger", model); err != nil {
		t.Fatal(err)
	}
	q := frame.Resource("pGia").(*Gia_Man_t)
	if q == p || p.nObjs != 0 {
		t.Fatal("frame did not replace and release original")
	}
	if len(model.Inputs) != 1 || len(model.Latches) != 1 {
		t.Fatal("export snapshot changed")
	}
	classes := []int32{4}
	if err := frame.Bridge.Import(frame, "gia-flop-classes", classes); err != nil {
		t.Fatal(err)
	}
	classes[0] = 7
	if Vec_IntEntry(q.vFlopClasses, 0) != 4 {
		t.Fatal("classes alias input")
	}
	bits := []int32{0, 1, 0}
	witness := testFrameWitness{bits, 1, 1, 1, 0}
	if err := frame.Bridge.Import(frame, "gia-counterexample", witness); err != nil {
		t.Fatal(err)
	}
	bits[1] = 0
	if q.SequentialCounterexample().PData[0] != 2 || frame.Cex.Bits[0] != 2 {
		t.Fatal("witness bits alias input")
	}
	if err := frame.Bridge.Import(frame, "gia-counterexample", witness); err == nil {
		t.Fatal("accepted invalid replay")
	}
	if q.SequentialCounterexample().PData[0] != 2 {
		t.Fatal("invalid witness replaced model")
	}
	witness.frame = int(^uint32(0) >> 1)
	if err := frame.Bridge.Import(frame, "gia-counterexample", witness); err == nil {
		t.Fatal("accepted overflowing frame count")
	}
}
