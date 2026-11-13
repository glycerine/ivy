package io

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/base/abc"
	abcmain "github.com/glycerine/abcgo/base/main"
)

func TestWriteTruthCommands(t *testing.T) {
	frame := abcmain.NewFrame()
	var errors bytes.Buffer
	frame.Err = &errors
	frame.ReplaceCurrentNetwork(abc.Abc_NtkCreateWithNode("11 1\n"))
	file := filepath.Join(t.TempDir(), "truth.txt")
	if IoCommandWriteTruth(frame, []string{"write_truth", "-x", file}) != 0 {
		t.Fatal("write failed")
	}
	data, err := os.ReadFile(file)
	if err != nil || string(data) != "1000" {
		t.Fatalf("binary truth %q: %v; %s", data, err, errors.String())
	}
	g := gia.Gia_ManStart(8)
	a, b, c := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	x := gia.Gia_ManAppendAnd(g, a, b)
	x = gia.Gia_ManAppendAnd(g, x, c)
	gia.Gia_ManAppendCo(g, x)
	abcmain.Abc_FrameUpdateGia(frame, g)
	if IoCommandWriteTruths(frame, []string{"&write_truths", file}) != 0 {
		t.Fatal("truths failed")
	}
	data, err = os.ReadFile(file)
	if err != nil || string(data) != "0x80\n" {
		t.Fatalf("GIA truth %q: %v; %s", data, err, errors.String())
	}
	if IoCommandWriteTruths(frame, []string{"&write_truths", "-b", file}) != 0 {
		t.Fatal("binary truths failed")
	}
	data, err = os.ReadFile(file)
	if err != nil || len(data) != 8 || data[0] != 0x80 {
		t.Fatalf("binary GIA truth %x: %v", data, err)
	}
}
