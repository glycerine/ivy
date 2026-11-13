package live

import (
	"slices"
	"testing"
)

func TestSourceLTLCopyReplacesAndOwnsDatabase(t *testing.T) {
	n := &Network{LTLProperties: []string{":G :F a", ":G :F b"}}
	frame := &Frame{LTLProperties: n.LTLProperties[:1]}
	Abc_FrameCopyLTLDataBase(frame, n)
	if !slices.Equal(frame.LTLProperties, n.LTLProperties) {
		t.Fatalf("copy %v", frame.LTLProperties)
	}
	frame.LTLProperties[0] = "changed"
	if n.LTLProperties[0] != ":G :F a" {
		t.Fatal("frame database aliases network")
	}
	n.LTLProperties = nil
	Abc_FrameCopyLTLDataBase(frame, n)
	if frame.LTLProperties == nil || len(frame.LTLProperties) != 0 {
		t.Fatalf("source allocates a new empty database: %v", frame.LTLProperties)
	}
}
