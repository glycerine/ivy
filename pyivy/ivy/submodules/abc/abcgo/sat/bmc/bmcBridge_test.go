package bmc

import (
	"bytes"
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/util"
)

func TestBmcBridgeProgressBytes(t *testing.T) {
	var got bytes.Buffer
	if err := writeBmcBridgeProgress(&got, 7, 12); err != nil {
		t.Fatal(err)
	}
	want := "000003 0000000000000037 property: safe<7>\nbug-free-depth: 12\n"
	if got.String() != want {
		t.Fatalf("bridge bytes\n got %q\nwant %q", got.String(), want)
	}
}

func TestBmcBridgeResultBytesAndOutputRemap(t *testing.T) {
	cex := util.Abc_CexAlloc(1, 1, 1)
	cex.IPo, cex.IFrame = 0, 0
	util.Abc_InfoSetBit(cex.PData, 1)
	var got bytes.Buffer
	if err := writeBmcBridgeResult(&got, 0, cex, 2); err != nil {
		t.Fatal(err)
	}
	want := []byte("000101 0000000000000012 \x02\x01\x02\x01\x00\x01\x01\x01\x03\x01\x01\x02")
	if !bytes.Equal(got.Bytes(), want) {
		t.Fatalf("bridge bytes\n got %v\nwant %v", got.Bytes(), want)
	}
	if cex.IPo != 0 {
		t.Fatalf("writing output 2 mutated shared CEX output to %d", cex.IPo)
	}
}

func TestBmcBridgeTerminalResultBytes(t *testing.T) {
	for _, tc := range []struct {
		result int32
		output int32
		want   []byte
	}{
		{result: 1, output: 5, want: []byte("000101 0000000000000004 \x03\x01\x05\x00")},
		{result: -1, output: 5, want: []byte("000101 0000000000000003 \x00\x01\x05")},
	} {
		var got bytes.Buffer
		if err := writeBmcBridgeResult(&got, tc.result, nil, tc.output); err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(got.Bytes(), tc.want) {
			t.Errorf("result %d bytes\n got %v\nwant %v", tc.result, got.Bytes(), tc.want)
		}
	}
}

func TestScalableBMCBridgeReportsSameModelOutputs(t *testing.T) {
	g := aig.NewGraph()
	pi := g.NewInput()
	g.AddOutput(pi)
	g.AddOutput(pi)
	var stream bytes.Buffer
	params := Saig_ParBmc_t{nFramesMax: 2, fSolveAll: 1, fUseBridge: 1, pBridgeWriter: &stream}
	if result := Saig_ManBmcScalable(g, &params); result != 0 {
		t.Fatalf("result=%d, want SAT witness", result)
	}
	if params.nFailOuts != 2 {
		t.Fatalf("same model should fail both outputs, got %d", params.nFailOuts)
	}
	want := []byte("000101 0000000000000011 \x02\x01\x00\x01\x00\x01\x01\x01\x03\x01\x00" +
		"000101 0000000000000011 \x02\x01\x01\x01\x00\x01\x01\x01\x03\x01\x00")
	if !bytes.Equal(stream.Bytes(), want) {
		t.Fatalf("bridge stream\n got %v\nwant %v", stream.Bytes(), want)
	}
}

func TestScalableBMCBridgeReportsSafeAtCompleteness(t *testing.T) {
	g := aig.NewGraph()
	state := g.NewInput()
	g.AddOutput(state) // next state
	g.AddOutput(state) // property
	aig.Aig_ManSetRegNum(g, 1)
	var stream bytes.Buffer
	params := Saig_ParBmc_t{nFramesMax: 4, fUseBridge: 1, pBridgeWriter: &stream}
	if result := Saig_ManBmcScalable(g, &params); result != 1 {
		t.Fatalf("result=%d, want safe completeness", result)
	}
	want := []byte("000003 0000000000000036 property: safe<0>\nbug-free-depth: 0\n" +
		"000003 0000000000000036 property: safe<0>\nbug-free-depth: 1\n" +
		"000101 0000000000000004 \x03\x01\x00\x00")
	if !bytes.Equal(stream.Bytes(), want) {
		t.Fatalf("bridge stream\n got %q\nwant %q", stream.Bytes(), want)
	}
}

func TestScalableBMCBridgeSingleSolveOmitsSatResult(t *testing.T) {
	g := aig.NewGraph()
	g.AddOutput(g.Const1())
	var stream bytes.Buffer
	params := Saig_ParBmc_t{nFramesMax: 1, fUseBridge: 1, pBridgeWriter: &stream}
	if result := Saig_ManBmcScalable(g, &params); result != 0 {
		t.Fatalf("result=%d, want SAT witness", result)
	}
	// C jumps to finish for !fSolveAll immediately after generating its
	// sequential model, before the bridge serialization in the solve-all path.
	if stream.Len() != 0 {
		t.Fatalf("single-solve bridge emitted %q, want no result package", stream.Bytes())
	}
}
