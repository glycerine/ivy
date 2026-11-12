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

func TestReadInitCopiesNetwork(t *testing.T) {
	n := abc.Abc_NtkAlloc(abc.ABC_NTK_LOGIC, abc.ABC_FUNC_SOP, 1)
	pi := n.NewObject(abc.ObjectPI)
	output := abc.Abc_NtkAddLatch(n, pi, abc.ABC_INIT_ZERO)
	output.SetName("state")
	abc.Abc_NtkOrderCisCos(n)
	file := filepath.Join(t.TempDir(), "init.txt")
	if err := os.WriteFile(file, []byte("state 1\nmissing 0\n"), 0600); err != nil {
		t.Fatal(err)
	}
	frame := abcmain.NewFrame()
	var out bytes.Buffer
	frame.Out = &out
	frame.ReplaceCurrentNetwork(n)
	if IoCommandReadInit(frame, []string{"read_init", file}) != 0 {
		t.Fatal("read failed")
	}
	copy := frame.Network().(*abc.Network)
	if copy == n || abc.Abc_LatchIsInit0(abc.Abc_ObjFanin0(output)) == 0 || abc.Abc_LatchIsInit1(copy.ObjectsOfKind(abc.ObjectLatch)[0]) == 0 {
		t.Fatal("initial state or ownership incorrect")
	}
	if !strings.Contains(out.String(), "missing") {
		t.Fatal("missing-register diagnostic not emitted")
	}
}
