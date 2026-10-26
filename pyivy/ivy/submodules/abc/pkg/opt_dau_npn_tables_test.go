package abc

// Stand-alone codec, summary and dump tests for DAU NPN tables.

import (
	"encoding/binary"
	"errors"
	"slices"
	"testing"
)

func TestDauFunctionTableCodec(t *testing.T) {
	enumeration, err := EnumerateDauNetworks(DauNetworkEnumerationOptions{Variables: 4, MaxNodes: 31})
	if err != nil {
		t.Fatal(err)
	}
	classes := enumeration.Classes()
	encoded, err := EncodeDauFunctionTable(4, classes)
	if err != nil {
		t.Fatal(err)
	}
	if string(encoded[:4]) != "DNP1" || encoded[4] != 4 || binary.LittleEndian.Uint32(encoded[8:]) != 222 || len(encoded) != dauFunctionTableHeaderSize+222*dauFunctionTableRecordSize {
		t.Fatal("header", encoded[:12], len(encoded))
	}
	inputs, decoded, err := DecodeDauFunctionTable(encoded, DauFunctionTableDecodeOptions{})
	if err != nil || inputs != 4 || !slices.Equal(decoded, classes) {
		t.Fatal(inputs, len(decoded), err)
	}
	decoded[0].Truth = 1
	_, again, err := DecodeDauFunctionTable(encoded, DauFunctionTableDecodeOptions{})
	if err != nil || again[0].Truth != 0 {
		t.Fatal("decoded table aliases input or earlier output", err)
	}
	if inputs, decoded, err := DecodeDauFunctionTable(encoded, DauFunctionTableDecodeOptions{MaxClasses: 221}); inputs != 0 || decoded != nil || !errors.Is(err, ErrLimit) {
		t.Fatal(inputs, decoded, err)
	}
}

func TestDauFunctionTableCodecRejectsMalformed(t *testing.T) {
	valid, err := EncodeDauFunctionTable(3, []DauFunctionClass{{Truth: 0, Nodes: 0, Support: 0}, {Truth: 0xaa, Nodes: 0, Support: 1}})
	if err != nil {
		t.Fatal(err)
	}
	cases := [][]byte{nil, valid[:11], append(slices.Clone(valid), 0), slices.Clone(valid), slices.Clone(valid), slices.Clone(valid), slices.Clone(valid)}
	cases[3][0] = 'X'
	cases[4][5] = 1
	cases[5][dauFunctionTableHeaderSize+12] = 2
	binary.LittleEndian.PutUint64(cases[6][dauFunctionTableHeaderSize:], 1<<8)
	for index, data := range cases {
		if inputs, decoded, err := DecodeDauFunctionTable(data, DauFunctionTableDecodeOptions{}); inputs != 0 || decoded != nil || !errors.Is(err, ErrFormat) {
			t.Fatal(index, inputs, decoded, err)
		}
	}
	if encoded, err := EncodeDauFunctionTable(3, []DauFunctionClass{{Truth: 0xaa, Nodes: 0, Support: 2}}); encoded != nil || !errors.Is(err, ErrFormat) {
		t.Fatal(encoded, err)
	}
}

func TestSummarizeAndDumpDauFunctions(t *testing.T) {
	enumeration, err := EnumerateDauNetworks(DauNetworkEnumerationOptions{Variables: 4, MaxNodes: 31})
	if err != nil {
		t.Fatal(err)
	}
	levels := enumeration.Levels()
	summary, err := SummarizeDauExactNPN(enumeration.Classes(), 4, 4, levels[len(levels)-1].Nodes)
	if err != nil {
		t.Fatal(err)
	}
	if summary[len(summary)-1].Total != 222 {
		t.Fatal(summary)
	}
	for _, row := range summary {
		if row.Canonical != row.SemiCanonical {
			t.Fatal("already exact frontier changed under exact summary", row)
		}
	}
	dumps, err := FormatDauFunctionDumps(enumeration.Classes(), 4, 4, 31)
	if err != nil {
		t.Fatal(err)
	}
	count := 0
	for _, dump := range dumps {
		for _, line := range dump.Lines {
			if len(line) != 4 {
				t.Fatal(dump.Nodes, line)
			}
			count++
		}
	}
	proper := 0
	for _, class := range enumeration.Classes() {
		if class.Support < 4 {
			proper++
		}
	}
	if count != 222-proper {
		t.Fatal(count, proper)
	}
	if _, err := FormatDauFunctionDumps(enumeration.Classes(), 4, 3, 2); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
}

func TestSummarizeDauExactNPNWide(t *testing.T) {
	truth := uint64(0xd6a59c3e70b4f821)
	permutation := []int{1, 2, 3, 4, 5, 0}
	other := transformDauNPNTruth(truth, 6, permutation, 0x2b)
	classes := []DauFunctionClass{
		{Truth: truth, Nodes: 0, Support: len(dauLogicalSupport(truth, 6))},
		{Truth: other, Nodes: 0, Support: len(dauLogicalSupport(other, 6))},
	}
	rows, err := SummarizeDauExactNPN(classes, 6, 6, 0)
	if err != nil || rows[0] != (DauExactNPNSummary{Nodes: 0, SemiCanonical: 2, Canonical: 1, Total: 1}) {
		t.Fatal(rows, err)
	}
	rows, err = SummarizeDauExactNPNWithOptions(classes, 6, 6, 0, DauExactNPNSummaryOptions{MaxCanonicalCandidates: 2*92160 - 1})
	if rows != nil || !errors.Is(err, ErrLimit) {
		t.Fatal(rows, err)
	}
	rows, err = SummarizeDauExactNPNWithOptions(classes, 6, 6, 0, DauExactNPNSummaryOptions{Stop: func() bool { return true }})
	if rows != nil || !errors.Is(err, ErrCancelled) {
		t.Fatal(rows, err)
	}
}

func FuzzDecodeDauFunctionTable(f *testing.F) {
	seed, err := EncodeDauFunctionTable(3, []DauFunctionClass{{Truth: 0, Nodes: 0, Support: 0}, {Truth: 0xaa, Nodes: 0, Support: 1}})
	if err != nil {
		f.Fatal(err)
	}
	f.Add(seed)
	f.Add([]byte("DNP1"))
	f.Fuzz(func(t *testing.T, data []byte) {
		inputs, classes, err := DecodeDauFunctionTable(data, DauFunctionTableDecodeOptions{MaxClasses: 1000})
		if err != nil {
			if inputs != 0 || classes != nil {
				t.Fatal("non-atomic decode", inputs, classes, err)
			}
			return
		}
		reencoded, err := EncodeDauFunctionTable(inputs, classes)
		if err != nil || !slices.Equal(reencoded, data) {
			t.Fatal(inputs, classes, err)
		}
	})
}
