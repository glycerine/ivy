package io

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/glycerine/abcgo/base/abc"
)

func bafFixture(header string, words ...uint32) []byte {
	data := []byte(header)
	for _, word := range words {
		var b [4]byte
		binary.LittleEndian.PutUint32(b[:], word)
		data = append(data, b[:]...)
	}
	return data
}
func TestBAFConstantConventionAndTruncation(t *testing.T) {
	data := bafFixture("constants\x000\x002\x000\x000\x00one\x00zero\x00", 0, 1)
	n, err := ParseBAF("constants.baf", data)
	if err != nil {
		t.Fatal(err)
	}
	for i, po := range n.ObjectsOfKind(abc.ObjectPO) {
		got, err := n.Evaluate(po.Signal(), nil)
		if err != nil || got != (i == 0) {
			t.Fatal(i, got, err)
		}
	}
	for end := 0; end < len(data); end++ {
		if _, err := ParseBAF("truncated.baf", data[:end]); err == nil {
			t.Fatal("accepted truncation", end)
		}
	}
}
func TestBAFANDAndLatchRoundTrip(t *testing.T) {
	data := bafFixture("and\x002\x001\x000\x001\x00a\x00b\x00y\x00", 2, 4, 6)
	n, err := ParseBAF("and.baf", data)
	if err != nil {
		t.Fatal(err)
	}
	var output bytes.Buffer
	if err = WriteBAF(&output, n); err != nil {
		t.Fatal(err)
	}
	copy, err := ParseBAF("copy.baf", output.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	pi := copy.ObjectsOfKind(abc.ObjectPI)
	po := copy.ObjectsOfKind(abc.ObjectPO)[0]
	for bits := 0; bits < 4; bits++ {
		got, err := copy.Evaluate(po.Signal(), map[*abc.Object]bool{pi[0]: bits&1 != 0, pi[1]: bits&2 != 0})
		if err != nil || got != (bits == 3) {
			t.Fatal(bits, got, err)
		}
	}
	latchData := bafFixture("seq\x001\x001\x001\x000\x00d\x00q\x00state\x00next\x00current\x00", 4, 10)
	seq, err := ParseBAF("seq.baf", latchData)
	if err != nil {
		t.Fatal(err)
	}
	output.Reset()
	if err = WriteBAF(&output, seq); err != nil {
		t.Fatal(err)
	}
	seq, err = ParseBAF("seq-copy.baf", output.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	latch := seq.ObjectsOfKind(abc.ObjectLatch)[0]
	if abc.Abc_LatchIsInit1(latch) == 0 || latch.Name() != "state" || latch.Fanins()[0].Object.Name() != "next" || latch.Fanouts()[0].Name() != "current" {
		t.Fatal("latch metadata lost")
	}
}
func TestBAFRejectsForwardGate(t *testing.T) {
	data := bafFixture("bad\x000\x001\x000\x001\x00y\x00", 2, 0, 2)
	if _, err := ParseBAF("bad.baf", data); err == nil {
		t.Fatal("accepted self-dependent gate")
	}
}
