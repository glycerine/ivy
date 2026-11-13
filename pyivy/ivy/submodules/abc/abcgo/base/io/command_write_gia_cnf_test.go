package io

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/glycerine/abcgo/aig/gia"
	abcmain "github.com/glycerine/abcgo/base/main"
)

func TestWriteGiaCNFCommand(t *testing.T) {
	p := gia.Gia_ManStart(8)
	a, b := gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p)
	gia.Gia_ManAppendCo(p, gia.Gia_ManAppendAnd(p, a, b))
	frame := abcmain.NewFrame()
	var errors bytes.Buffer
	frame.Err = &errors
	abcmain.Abc_FrameUpdateGia(frame, p)
	file := filepath.Join(t.TempDir(), "out.cnf")
	if IoCommandWriteCnf2(frame, []string{"&write_cnf", "-K", "4", file}) != 0 {
		t.Fatalf("command failed: %s", errors.String())
	}
	data, err := os.ReadFile(file)
	if err != nil || !strings.Contains(string(data), "p cnf ") {
		t.Fatalf("CNF %q: %v; %s", data, err, errors.String())
	}
}
