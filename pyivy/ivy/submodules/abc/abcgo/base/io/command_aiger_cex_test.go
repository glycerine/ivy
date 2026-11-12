package io

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/glycerine/abcgo/base/abc"
	abcmain "github.com/glycerine/abcgo/base/main"
)

func TestFrameCounterexampleWordConversion(t *testing.T) {
	input := &abcmain.Counterexample{PIs: 65, Frame: 0, Bits: []uint64{0x8000000100000001, 1}}
	got, err := frameCounterexample(input)
	if err != nil || len(got.PData) != 3 || got.PData[0] != 1 || got.PData[1] != 0x80000001 || got.PData[2] != 1 {
		t.Fatalf("conversion %+v: %v", got, err)
	}
	input.Bits = nil
	if _, err := frameCounterexample(input); err == nil {
		t.Fatal("accepted truncated data")
	}
}

func TestWriteAigerCounterexampleCommand(t *testing.T) {
	logic := abc.Abc_NtkCreateWithNode("11 1\n")
	n := abc.Abc_NtkStrash(logic, 0, 1, 0)
	frame := abcmain.NewFrame()
	var errors bytes.Buffer
	frame.Err = &errors
	frame.ReplaceCurrentNetwork(n)
	frame.SetVerificationResult(0, 0, &abcmain.Counterexample{PIs: 2, Output: 0, Frame: 0, Bits: []uint64{3}})
	filename := filepath.Join(t.TempDir(), "trace.cex")
	if IoCommandWriteAigerCex(frame, []string{"write_aiger_cex", filename}) != 0 {
		t.Fatal("command failed")
	}
	data, err := os.ReadFile(filename)
	if err != nil || !strings.HasPrefix(string(data), "1\n 11") {
		t.Fatalf("trace %q: %v; %s", data, err, errors.String())
	}
}
