package io

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	abcmain "github.com/glycerine/abcgo/base/main"
)

func TestJSONCommandsRoundTrip(t *testing.T) {
	dir := t.TempDir()
	input, output := filepath.Join(dir, "in.json"), filepath.Join(dir, "out.json")
	if err := os.WriteFile(input, []byte(`{"nodes":["one","two","text"]}`), 0600); err != nil {
		t.Fatal(err)
	}
	frame := abcmain.NewFrame()
	var errors bytes.Buffer
	frame.Err = &errors
	if IoCommandReadJson(frame, []string{"read_json", input}) != 0 {
		t.Fatalf("read: %s", errors.String())
	}
	if IoCommandWriteJson(frame, []string{"write_json", output}) != 0 {
		t.Fatalf("write: %s", errors.String())
	}
	got, err := ReadJSONFile(output)
	if err != nil {
		t.Fatal(err)
	}
	want, err := ReadJSONFile(input)
	if err != nil {
		t.Fatal(err)
	}
	a, _ := got.MarshalJSON()
	b, _ := want.MarshalJSON()
	if !bytes.Equal(a, b) {
		t.Fatalf("round trip: %s != %s", a, b)
	}
}
