package io

import (
	"bytes"
	"testing"

	"github.com/glycerine/abcgo/base/abc"
	abcmain "github.com/glycerine/abcgo/base/main"
)

func TestNetworkCNFCommandDispatch(t *testing.T) {
	darCNFWriter.RLock()
	old := darCNFWriter.fn
	darCNFWriter.RUnlock()
	defer RegisterDarCNFWriter(old)
	frame := abcmain.NewFrame()
	var errors bytes.Buffer
	frame.Err = &errors
	n := abc.Abc_NtkCreateWithNode("11 1\n")
	frame.ReplaceCurrentNetwork(n)
	called := false
	RegisterDarCNFWriter(func(got *abc.Network, filename string, fast, polarity, verbose int32) *abc.Network {
		called = got == n && filename == "out.cnf" && fast == 1 && polarity == 0 && verbose == 1
		return nil
	})
	if IoCommandWriteCnf(frame, []string{"write_cnf", "-fcv", "out.cnf"}) != 0 || !called {
		t.Fatalf("dispatch failed: %s", errors.String())
	}
}
