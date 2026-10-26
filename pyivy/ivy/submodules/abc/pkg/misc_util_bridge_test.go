package abc

// Standalone tests for src/misc/util/utilBridge.c.

import (
	"bytes"
	"errors"
	"slices"
	"testing"
)

func bridgeSequentialFixture(t *testing.T) *GiaManager {
	t.Helper()
	manager, err := NewGiaManagerWithLimit(64)
	if err != nil {
		t.Fatal(err)
	}
	first, _ := manager.AppendCI()
	second, _ := manager.AppendCI()
	state, _ := manager.AppendCI()
	gate, err := manager.AppendAnd(first, second.Not())
	if err != nil {
		t.Fatal(err)
	}
	root, err := manager.AppendAnd(state, gate)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = manager.AppendCO(root); err != nil {
		t.Fatal(err)
	}
	if _, err = manager.AppendCO(gate); err != nil {
		t.Fatal(err)
	}
	if err = manager.SetRegisterCount(1); err != nil {
		t.Fatal(err)
	}
	return manager
}

func TestBridgePackageAndMessages(t *testing.T) {
	message, err := BridgeText([]byte("hello"), 16)
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := message.MarshalBinary(64)
	if err != nil || string(encoded) != "999996 0000000000000005 hello" {
		t.Fatal(string(encoded), err)
	}
	parsed, consumed, err := ParseBridgePackage(append(encoded, 'x'), BridgeLimits{})
	if err != nil || consumed != len(encoded) || parsed.Type != BridgeTextMessage || !bytes.Equal(parsed.Payload, []byte("hello")) {
		t.Fatal(parsed, consumed, err)
	}
	parsed.Payload[0] = 'X'
	if message.Payload[0] != 'h' {
		t.Fatal(message.Payload)
	}
	abortMessage, abortErr := BridgeAbort([]byte{1}, 4)
	progressMessage, progressErr := BridgeProgress([]byte{2}, 4)
	for _, row := range []struct {
		message BridgePackage
		kind    int
	}{
		{mustBridgeMessage(t, abortMessage, abortErr), BridgeAbortMessage},
		{mustBridgeMessage(t, progressMessage, progressErr), BridgeProgressMessage},
		{BridgeBadAbstraction(), BridgeBadAbsMessage},
	} {
		if row.message.Type != row.kind {
			t.Fatal(row)
		}
	}
	for cut := 0; cut < len(encoded); cut++ {
		if _, _, err := ParseBridgePackage(encoded[:cut], BridgeLimits{}); err == nil {
			t.Fatal("accepted truncation", cut)
		}
	}
	bad := slices.Clone(encoded)
	bad[6] = '_'
	if _, _, err := ParseBridgePackage(bad, BridgeLimits{}); !errors.Is(err, ErrFormat) {
		t.Fatal(err)
	}
	if _, _, err := ParseBridgePackage(encoded, BridgeLimits{MaxBytes: 4}); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
}

func mustBridgeMessage(t *testing.T, message BridgePackage, err error) BridgePackage {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
	return message
}

func TestBridgeNetlistBodyAndTriplet(t *testing.T) {
	source := bridgeSequentialFixture(t)
	body, err := MarshalBridgeNetlistBody(source, BridgeLimits{MaxBytes: 128, MaxObjects: 64})
	if err != nil || !bytes.Equal(body, []byte{2, 1, 2, 8, 7, 16, 10, 42, 1, 13}) {
		t.Fatal(body, err)
	}
	message, err := BridgeNetlist(source, BridgeAbsNetlist, BridgeLimits{MaxBytes: 128, MaxObjects: 64})
	if err != nil || message.Type != BridgeAbsNetlist || !bytes.Equal(message.Payload, body) {
		t.Fatal(message, err)
	}
	incoming := append(slices.Clone(body), 1, 0, 0)
	parsed, initials, err := ParseBridgeNetlistBody(incoming, BridgeLimits{MaxBytes: 128, MaxObjects: 64})
	if err != nil || !bytes.Equal(initials, []byte{BridgeValueZero}) || parsed.Name != "temp" || parsed.PrimaryInputCount() != 2 || parsed.RegisterCount() != 1 || parsed.PrimaryOutputCount() != 1 || parsed.AndCount() != 2 {
		t.Fatal(parsed, initials, err)
	}
	for assignment := 0; assignment < 8; assignment++ {
		primary := []bool{assignment&1 != 0, assignment&2 != 0}
		state := []bool{assignment&4 != 0}
		sourceOut, sourceNext, sourceErr := source.Step(primary, state)
		parsedOut, parsedNext, parsedErr := parsed.Step(primary, state)
		if sourceErr != nil || parsedErr != nil || !slices.Equal(sourceOut, parsedOut) || !slices.Equal(sourceNext, parsedNext) {
			t.Fatal(assignment, sourceOut, sourceNext, parsedOut, parsedNext, sourceErr, parsedErr)
		}
	}
	packages := []BridgePackage{{Type: BridgeProgressMessage, Payload: []byte("begin")}, {Type: BridgeNetlistMessage, Payload: incoming}, {Type: BridgeProgressMessage, Payload: []byte("end")}}
	var stream []byte
	for _, item := range packages {
		encoded, err := item.MarshalBinary(256)
		if err != nil {
			t.Fatal(err)
		}
		stream = append(stream, encoded...)
	}
	tripletGraph, tripletInitials, triplet, err := ParseBridgeNetlistTriplet(stream, BridgeLimits{MaxBytes: 256, MaxObjects: 64})
	if err != nil || tripletGraph.RegisterCount() != 1 || !bytes.Equal(tripletInitials, initials) || triplet[0].Type != BridgeProgressMessage || triplet[2].Type != BridgeProgressMessage {
		t.Fatal(tripletGraph, tripletInitials, triplet, err)
	}
	if graph, _, _, err := ParseBridgeNetlistTriplet(append(stream, 0), BridgeLimits{MaxBytes: 256, MaxObjects: 64}); !errors.Is(err, ErrFormat) || graph != nil {
		t.Fatal(graph, err)
	}
}

func TestBridgeResults(t *testing.T) {
	holds, err := BridgeHolds(300, 32)
	if err != nil || holds.Type != BridgeResultMessage || !bytes.Equal(holds.Payload, []byte{3, 1, 0xac, 0x02, 0}) {
		t.Fatal(holds, err)
	}
	unknown, err := BridgeUnknown(300, 32)
	if err != nil || !bytes.Equal(unknown.Payload, []byte{0, 1, 0xac, 0x02}) {
		t.Fatal(unknown, err)
	}
	cex, _ := NewCounterexample(2, 2, 2)
	cex.Property = 3
	_ = cex.SetInitial(0, true)
	_ = cex.SetInput(0, 0, true)
	_ = cex.SetInput(1, 1, true)
	message, err := BridgeCounterexample(cex, 64)
	expected := []byte{2, 1, 3, 1, 1, 1, 2, 2, 3, 2, 2, 2, 3, 1, 2, 2, 2}
	if err != nil || message.Type != BridgeResultMessage || !bytes.Equal(message.Payload, expected) {
		t.Fatal(message, expected, err)
	}
	if result, err := BridgeResult(0, cex, 0, 64); err != nil || !bytes.Equal(result.Payload, expected) {
		t.Fatal(result, err)
	}
	if _, err := BridgeCounterexample(cex, 8); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	for _, result := range []int{-2, 2} {
		if _, err := BridgeResult(result, nil, 0, 8); !errors.Is(err, ErrDimension) {
			t.Fatal(result, err)
		}
	}
}

func TestBridgeNetlistErrors(t *testing.T) {
	manager := bridgeSequentialFixture(t)
	body, _ := MarshalBridgeNetlistBody(manager, BridgeLimits{MaxObjects: 64})
	incoming := append(slices.Clone(body), 1, 0, 0)
	for cut := 0; cut < len(incoming); cut++ {
		if parsed, _, err := ParseBridgeNetlistBody(incoming[:cut], BridgeLimits{MaxObjects: 64}); err == nil || parsed != nil {
			t.Fatal("accepted truncation", cut, parsed)
		}
	}
	for index, value := range map[int]byte{len(incoming) - 3: 2, len(incoming) - 2: 1, len(incoming) - 1: 1} {
		bad := slices.Clone(incoming)
		bad[index] = value
		if _, _, err := ParseBridgeNetlistBody(bad, BridgeLimits{MaxObjects: 64}); !errors.Is(err, ErrFormat) {
			t.Fatal(index, err)
		}
	}
	if _, _, err := ParseBridgeNetlistBody(incoming, BridgeLimits{MaxObjects: manager.ObjectCount() - 1}); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	if _, _, err := ParseBridgeNetlistBody([]byte{0x80, 0x80, 0x80, 0x80, 0x10}, BridgeLimits{MaxObjects: 64}); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	if _, _, err := ParseBridgeNetlistBody([]byte{0, 0, 1, 4, 0, 0, 1, 0, 0}, BridgeLimits{MaxObjects: 64}); !errors.Is(err, ErrInvalidLiteral) {
		t.Fatal(err)
	}
	unsupported, _ := NewGiaManagerWithLimit(8)
	a, _ := unsupported.AppendCI()
	b, _ := unsupported.AppendCI()
	x, _ := unsupported.AppendXorReal(a, b)
	_, _ = unsupported.AppendCO(x)
	if _, err := MarshalBridgeNetlistBody(unsupported, BridgeLimits{MaxObjects: 8}); !errors.Is(err, ErrUnsupported) {
		t.Fatal(err)
	}
	noOutput, _ := NewGiaManagerWithLimit(4)
	_, _ = noOutput.AppendCI()
	if _, err := MarshalBridgeNetlistBody(noOutput, BridgeLimits{MaxObjects: 4}); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
}

func FuzzBridgeParsers(f *testing.F) {
	f.Add([]byte("000003 0000000000000001 x"))
	f.Add([]byte{2, 0, 0, 0, 1, 1, 0, 0})
	f.Add([]byte{})
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 512 {
			data = data[:512]
		}
		limits := BridgeLimits{MaxBytes: 512, MaxObjects: 256, MaxInputs: 128, MaxOutputs: 128}
		if message, consumed, err := ParseBridgePackage(data, limits); err == nil {
			if consumed < bridgeHeaderSize || consumed > len(data) || len(message.Payload) != consumed-bridgeHeaderSize {
				t.Fatal(message, consumed)
			}
			encoded, err := message.MarshalBinary(512)
			if err != nil || !bytes.Equal(encoded, data[:consumed]) {
				t.Fatal(encoded, consumed, err)
			}
		}
		if manager, initials, err := ParseBridgeNetlistBody(data, limits); err == nil {
			if manager == nil || manager.Check() != nil || len(initials) != manager.RegisterCount() {
				t.Fatal(manager, initials)
			}
		}
	})
}
