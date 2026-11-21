package gia

import (
	"encoding/binary"
	"fmt"
	"math"
	"strings"
	"testing"
)

func TestAigerTextDriversUseBoundedScratch(t *testing.T) {
	wire := []byte("aig 1 1 0 1024 0\n" + strings.Repeat("2\n", 1024))
	allocs := testing.AllocsPerRun(5, func() {
		p, err := ParseAiger(wire, AigerReadOptions{SkipStrash: true})
		if err != nil || Gia_ManPoNum(p) != 1024 {
			t.Fatal("output table", err)
		}
		Gia_ManStop(p)
	})
	if allocs >= 100 {
		t.Fatalf("1024 decimal output records used %.0f allocations", allocs)
	}
}

func TestAigerTextRecordWhitespaceAndPhases(t *testing.T) {
	wire := []byte("aag 8 2 1 1 1\r\n\t0002 \r\n\u00a00008\r\n0012\t0016\u20030012\r\n \t0017\t\r\n0016\u00a00012 0002\r\n")
	p, err := ParseAiger(wire, AigerReadOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if Gia_ManPiNum(p) != 2 || p.nRegs != 1 || p.aigerLatchInits[0] != -1 || Gia_ManAndNum(p) != 1 {
		t.Fatal("incorrect numeric records")
	}
	for i := range 8 {
		bits := []bool{i&1 != 0, i&2 != 0, i&4 != 0}
		got := aigerOutputs(p, bits)
		if got[1] != (bits[0] && bits[2]) || got[0] == got[1] {
			t.Fatalf("assignment %d outputs %v", i, got)
		}
	}
}

func TestAigerDeltaCursorAcrossWidths(t *testing.T) {
	values := []uint32{0, 1, 127, 128, 16383, 16384, 1<<21 - 1, 1 << 21, 1<<28 - 1, 1 << 28, math.MaxUint32}
	var data []byte
	for _, v := range values {
		data = binary.AppendUvarint(data, uint64(v))
	}
	c := aigerCursor{data: data}
	for _, want := range values {
		got, err := c.unsigned()
		if err != nil || got != want {
			t.Fatalf("delta got %d %v, want %d", got, err, want)
		}
	}
	if c.pos != len(data) {
		t.Fatal("cursor ended at", c.pos)
	}
	for _, suffix := range [][]byte{nil, {0x80}, {0x80, 0x80}, {0xff, 0xff, 0xff, 0xff, 0x10}, {0x80, 0x80, 0x80, 0x80, 0x80, 0}} {
		cursor := aigerCursor{data: append([]byte{3}, suffix...), pos: 1}
		if _, err := cursor.unsigned(); err == nil || cursor.pos != 1 {
			t.Fatal("cursor after incomplete or oversized delta", cursor.pos)
		}
	}
}

func TestAigerNamesCoverRegisterInputsAndKeepPOIds(t *testing.T) {
	wire := []byte("aag 2 1 1 1 0 1 1\n2\n4 2\n4\n2\n3\ni0 pi\nl0 reg\no0 out\nb0 @o7\nc0 assume\n")
	p, err := ParseAiger(wire, AigerReadOptions{})
	if err != nil {
		t.Fatal(err)
	}
	names := Vec_PtrArray(p.VNamesOut)
	if len(names) != int(Gia_ManCoNum(p)) || names[0] != "out" || names[1] != "@o7" || names[2] != "assume" || names[3] != nil {
		t.Fatalf("output names %v do not cover the four COs", names)
	}
	ids := Vec_IntArray(p.vUserPoIds)
	if len(ids) != 3 || ids[0] != -1 || ids[1] != 7 || ids[2] != -1 {
		t.Fatal("public output IDs", ids)
	}
	clear(wire)
	if p.aigerSymbols['l'][0] != "reg" || p.aigerSymbols['c'][0] != "assume" {
		t.Fatal("symbol points into mutable input")
	}
}

func TestAigerDuplicateAfterEmptySymbol(t *testing.T) {
	for _, second := range []string{"", "next"} {
		wire := []byte("aig 1 1 0 1 0\n2\no0 \no0 " + second + "\n")
		if _, err := ParseAiger(wire, AigerReadOptions{}); err == nil {
			t.Fatal("accepted a second label for the same output")
		}
	}
}

func TestAigerNamedPortsShareSymbolText(t *testing.T) {
	var text strings.Builder
	text.WriteString("aig 1 1 0 1024 0\n" + strings.Repeat("2\n", 1024))
	for i := range 1024 {
		fmt.Fprintf(&text, "o%d output %d\n", i, i)
	}
	wire := []byte(text.String())
	n := testing.AllocsPerRun(5, func() {
		p, err := ParseAiger(wire, AigerReadOptions{SkipStrash: true})
		if err != nil || p.aigerSymbols['o'][1023] != "output 1023" {
			t.Fatal("named outputs", err)
		}
		Gia_ManStop(p)
	})
	// The public []any name vector boxes each string once.
	if n >= 1152 {
		t.Fatalf("1024 named ports used %.0f allocations", n)
	}
}
