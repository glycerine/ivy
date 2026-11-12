package amap

import (
	"strings"
	"testing"
)

func TestAmapRemoveCommentsAndContinuedLines(t *testing.T) {
	input := "GATE # ignored\r\nname \\\r\n 1.0\n# trailing comment"
	var dots, lines int32
	got := Amap_RemoveComments(input, &dots, &lines)
	if dots != 1 || lines != 2 {
		t.Fatalf("counts = dots:%d lines:%d, want dots:1 lines:2", dots, lines)
	}
	if got == input || !strings.HasPrefix(got, "GATE") || !strings.Contains(got, "name") || strings.Contains(got, "\\") {
		t.Fatalf("comment removal did not preserve active text: %q", got)
	}
}

func TestAmapReadLibraryAndWildcardPinExpansion(t *testing.T) {
	text := `GATE INVX 1.0 Y A';
PIN A NONINV 1.0 2.0 3.0 4.0 5.0 6.0
GATE NAND2 1.5 Y !(A*B);
PIN A NONINV 1.0 2.0 3.0 4.0 5.0 6.0
PIN B NONINV 1.5 2.5 3.5 4.5 5.5 6.5
GATE WILD 2.0 Y A*B;
PIN * UNKNOWN 0.0 0.0 0.0 0.0 0.0 0.0
.end
`
	lib := Amap_LibReadBuffer(text, 0)
	if lib == nil {
		t.Fatal("valid gate library did not parse")
	}
	defer Amap_LibFree(lib)
	if got := Vec_PtrSize(lib.vGates); got != 3 {
		t.Fatalf("gate count = %d, want 3", got)
	}
	wild := Amap_LibGate(lib, 2)
	if wild.nPins != 2 || wild.Pins[0].pName != "A" || wild.Pins[1].pName != "B" {
		t.Fatalf("wildcard pins = %#v", wild.Pins)
	}
	if got := Amap_LibParseEquations(lib, 0); got != 1 {
		t.Fatal("parsed library equations failed")
	}
	if lib.vGates == nil || Amap_LibGate(lib, 1).pFunc[0]&0xf != 0x7 {
		t.Fatalf("NAND2 truth = %#v, want 0x7", Amap_LibGate(lib, 1).pFunc)
	}
}

func TestAmapDeriveTokensSkipsLatchBody(t *testing.T) {
	tokens := Amap_DeriveTokens("GATE INV 1 Y A; LATCH Q D CLK 0 1 GATE BUF 1 Y A; PIN A UNKNOWN 0 0 0 0 0 0 .end")
	defer Vec_PtrFree(tokens)
	var got []string
	for i := int32(0); i < Vec_PtrSize(tokens); i++ {
		got = append(got, Vec_PtrEntry(tokens, i).(string))
	}
	if len(got) == 0 || got[0] != "GATE" || got[len(got)-1] != ".end" {
		t.Fatalf("tokens = %v", got)
	}
	for _, token := range got {
		if token == "LATCH" || token == "Q" || token == "CLK" {
			t.Fatalf("latch token %q was retained: %v", token, got)
		}
	}
}
