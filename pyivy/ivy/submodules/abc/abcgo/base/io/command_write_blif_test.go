package io

import (
	"bytes"
	"testing"

	"github.com/glycerine/abcgo/base/abc"
	abcmain "github.com/glycerine/abcgo/base/main"
)

func TestSpecialBlifCommandDispatch(t *testing.T) {
	specialBLIFWriter.RLock()
	old := specialBLIFWriter.fn
	specialBLIFWriter.RUnlock()
	defer RegisterSpecialBLIFWriter(old)
	frame := abcmain.NewFrame()
	var errors bytes.Buffer
	frame.Err = &errors
	n := abc.Abc_NtkCreateWithNode("11 1\n")
	frame.ReplaceCurrentNetwork(n)
	called := false
	RegisterSpecialBLIFWriter(func(got *abc.Network, path, structure string, hierarchy bool) error {
		called = got == n && path == "out.blif" && structure == "44" && hierarchy
		return nil
	})
	if IoCommandWriteBlif(frame, []string{"write_blif", "-a", "-S", "44", "out.blif"}) != 0 || !called {
		t.Fatalf("dispatch failed: %s", errors.String())
	}
	if IoCommandWriteBlif(frame, []string{"write_blif", "-S", "4", "out.blif"}) != 1 {
		t.Fatal("accepted invalid LUT structure")
	}
}
