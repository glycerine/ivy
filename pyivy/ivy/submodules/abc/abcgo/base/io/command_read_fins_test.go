package io

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/glycerine/abcgo/base/abc"
	abcmain "github.com/glycerine/abcgo/base/main"
)

func TestReadFinsCommand(t *testing.T) {
	n := abc.Abc_NtkCreateWithNode("11 1\n")
	pi := n.ObjectsOfKind(abc.ObjectPI)[0]
	file := filepath.Join(t.TempDir(), "faults.txt")
	if err := os.WriteFile(file, []byte("1 "+abc.Abc_ObjName(pi)+" SA0\n"), 0600); err != nil {
		t.Fatal(err)
	}
	frame := abcmain.NewFrame()
	frame.Out, frame.Err = &bytes.Buffer{}, &bytes.Buffer{}
	frame.ReplaceCurrentNetwork(n)
	if IoCommandReadFins(frame, []string{"read_fins", file}) != 0 {
		t.Fatal("command failed")
	}
	if want := []int32{-1, -1, int32(pi.ID), -99}; !reflect.DeepEqual(n.Faults, want) {
		t.Fatalf("got %v, want %v", n.Faults, want)
	}
	if _, err := ReadFaultList(n, bytes.NewBufferString("2 x SA0\n"), false, &bytes.Buffer{}); err == nil {
		t.Fatal("accepted out-of-order line")
	}
}
