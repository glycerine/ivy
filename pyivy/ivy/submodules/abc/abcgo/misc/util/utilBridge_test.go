package util

import (
	"bytes"
	"io"
	"reflect"
	"testing"
)

func TestBridgeFramingAndGraphRoundtrip(t *testing.T) {
	p := &BridgeNetlist{Inputs: 2, Registers: 1, Ands: [][2]int32{{2, 4}, {6, 8}}, Outputs: []int32{10, 1}, RegisterInputs: []int32{9}}
	encoded := Gia_ManToBridgeVec(p)
	q, inits, err := DecodeBridgeNetlist(encoded)
	if err != nil || !reflect.DeepEqual(p, q) || !reflect.DeepEqual(inits, []int32{BRIDGE_VALUE_0}) {
		t.Fatal(q, inits, err)
	}
	var buf bytes.Buffer
	if err = WriteBridgePackage(&buf, 107, encoded); err != nil {
		t.Fatal(err)
	}
	if string(buf.Bytes()[:7]) != "000107 " {
		t.Fatal(buf.String())
	}
	typ, data, err := ReadBridgePackage(&buf)
	if err != nil || typ != 107 || !bytes.Equal(data, encoded) {
		t.Fatal(typ, data, err)
	}
}
func TestBridgeTruncationAndResults(t *testing.T) {
	if _, _, err := ReadBridgePackage(bytes.NewReader([]byte("short"))); err != io.ErrUnexpectedEOF {
		t.Fatal(err)
	}
	for _, result := range []int32{-1, 1} {
		var buf bytes.Buffer
		if Gia_ManToBridgeResult(&buf, result, nil, 129) != 1 {
			t.Fatal(result)
		}
		typ, data, err := ReadBridgePackage(&buf)
		if err != nil || typ != 101 || len(data) < 4 {
			t.Fatal(typ, data, err)
		}
	}
	cex := Abc_CexCreate(1, 2, []int32{0, 1, 0, 0, 1}, 1, 0, 0)
	var buf bytes.Buffer
	if Gia_ManToBridgeResult(&buf, 0, cex, 0) != 1 {
		t.Fatal("cex")
	}
	_, data, err := ReadBridgePackage(&buf)
	if err != nil || data[0] != BRIDGE_VALUE_0 {
		t.Fatal(data, err)
	}
}
