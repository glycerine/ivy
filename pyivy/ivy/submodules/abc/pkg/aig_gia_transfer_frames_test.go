package abc

import (
	"encoding/json"
	"os"
	"slices"
	"testing"
)

func TestGiaTransferFrameSignalsComposition(t *testing.T) {
	source := NewGiaManager()
	input, _ := source.AppendCI()
	ro, _ := source.AppendCI()
	gate, err := source.AppendAnd(input, ro)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = source.AppendCO(gate); err != nil {
		t.Fatal(err)
	}
	if _, err = source.AppendCO(input); err != nil {
		t.Fatal(err)
	}
	if err = source.SetRegisterCount(1); err != nil {
		t.Fatal(err)
	}
	frames, err := source.Unroll(3)
	if err != nil {
		t.Fatal(err)
	}
	target := frames.Graph().Clone()
	mapping := make([]GiaLit, target.ObjectCount())
	for i := range mapping {
		mapping[i] = GiaLit(i) << 1
	}
	signals := []GiaID{ro.ID(), input.ID(), gate.ID(), ro.ID()}
	result, err := TransferGiaFrameSignals(source, frames, target, mapping, signals, 0)
	if err != nil {
		t.Fatal(err)
	}
	expected := make([]GiaID, 0)
	// This fixture's independent temporal structure is x0, x1, x2,
	// x0&x1, x1&x2, in that signal/frame discovery order.
	for _, item := range []struct {
		frame int
		id    GiaID
	}{{1, ro.ID()}, {2, ro.ID()}, {2, input.ID()}, {1, gate.ID()}, {2, gate.ID()}} {
		literal, err := frames.ObjectLiteral(item.frame, item.id)
		if err != nil {
			t.Fatal(err)
		}
		expected = append(expected, literal.ID())
	}
	raw, err := os.ReadFile("test_vectors/oracle/gia_transfer_frames.json")
	if err != nil {
		t.Fatal(err)
	}
	var native []GiaID
	if err = json.Unmarshal(raw, &native); err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(result, native) {
		t.Fatal("native frame transfer", result, native)
	}
	if !slices.Equal(result, expected) {
		t.Fatal(result, expected)
	}
	// Invalid and zero copies disappear; constant-one is a retained ID zero.
	first, _ := frames.ObjectLiteral(1, ro.ID())
	mapping[first.ID()] = GiaConst1
	second, _ := frames.ObjectLiteral(2, ro.ID())
	mapping[second.ID()] = InvalidGiaLit
	result, err = TransferGiaFrameSignals(source, frames, target, mapping, []GiaID{ro.ID()}, 0)
	if err != nil || !slices.Equal(result, []GiaID{0}) {
		t.Fatal(result, err)
	}
	for _, cap := range []int{-1, 1} {
		if _, err := TransferGiaFrameSignals(source, frames, target, mapping, signals, cap); err == nil {
			t.Fatal("invalid cap")
		}
	}
	if _, err := TransferGiaFrameSignals(source, frames, target, mapping[:1], signals, 0); err == nil {
		t.Fatal("mapping dimensions")
	}
	if _, err := TransferGiaFrameSignals(source, frames, target, mapping, []GiaID{source.cos[0]}, 0); err == nil {
		t.Fatal("CO candidate accepted")
	}
	mapping[first.ID()] = GiaLit(target.ObjectCount()) << 1
	if _, err := TransferGiaFrameSignals(source, frames, target, mapping, []GiaID{ro.ID()}, 0); err == nil {
		t.Fatal("invalid target mapping")
	}
}
