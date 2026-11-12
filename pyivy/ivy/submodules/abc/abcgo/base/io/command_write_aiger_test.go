package io

import (
	"bytes"
	"path/filepath"
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/base/abc"
	abcmain "github.com/glycerine/abcgo/base/main"
)

func TestWriteAigerCommand(t *testing.T) {
	oldLogic := abc.Abc_NtkCreateWithNode("11 1\n")
	n := abc.Abc_NtkStrash(oldLogic, 0, 1, 0)
	frame := abcmain.NewFrame()
	var errors bytes.Buffer
	frame.Err = &errors
	frame.ReplaceCurrentNetwork(n)
	aigerCanonicalizer.RLock()
	old := aigerCanonicalizer.fn
	aigerCanonicalizer.RUnlock()
	defer RegisterAigerCanonicalizer(old)
	called := false
	RegisterAigerCanonicalizer(func(g *aig.Graph, verbose int32) *aig.Graph { called = verbose == 1; return g })
	for _, opts := range [][]string{{"-s"}, {"-uv"}} {
		path := filepath.Join(t.TempDir(), "network.aig")
		args := append([]string{"write_aiger"}, opts...)
		args = append(args, path)
		if IoCommandWriteAiger(frame, args) != 0 {
			t.Fatalf("write %v: %s", opts, errors.String())
		}
		parsed, err := ReadAiger(path, true)
		if err != nil || abc.Abc_NtkPiNum(parsed) != 2 || abc.Abc_NtkPoNum(parsed) != 1 {
			t.Fatalf("read %v: %v", opts, err)
		}
	}
	if !called {
		t.Fatal("canonicalizer not called")
	}
}
