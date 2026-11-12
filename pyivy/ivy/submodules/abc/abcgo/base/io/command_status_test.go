package io

import (
	"bytes"
	"path/filepath"
	"testing"

	abcmain "github.com/glycerine/abcgo/base/main"
)

func TestVerificationLogCommands(t *testing.T) {
	frame := abcmain.NewFrame()
	var errors bytes.Buffer
	frame.Err = &errors
	frame.SetVerificationResult(0, 2, &abcmain.Counterexample{PIs: 2, Registers: 1, Output: 3, Frame: 1, Bits: []uint64{0b10101}})
	file := filepath.Join(t.TempDir(), "status.log")
	if IoCommandWriteStatus(frame, []string{"write_status", file}) != 0 {
		t.Fatal("write failed")
	}
	frame.SetVerificationResult(-1, -1, nil)
	if IoCommandReadStatus(frame, []string{"read_status", file}) != 0 {
		t.Fatalf("read failed: %s", errors.String())
	}
	status, depth, cex := frame.VerificationResult()
	if status != 0 || depth != 2 || cex == nil || cex.Output != 3 || cex.Frame != 1 || cex.PIs != 2 || cex.Registers != 1 || cex.Bits[0] != 0b10101 {
		t.Fatalf("incorrect result: %d %d %+v", status, depth, cex)
	}
}

func TestVerificationLogRejectsMalformedTrace(t *testing.T) {
	if _, _, _, err := readVerificationLog(bytes.NewBufferString("snl_SAT 3 engine 0\n1\n101\n")); err == nil {
		t.Fatal("accepted malformed trace")
	}
}
