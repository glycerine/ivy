package amap

import (
	"bytes"
	"errors"
	"io"
	"testing"

	"github.com/glycerine/abcgo/misc/vec"
)

func TestLibrarySelectionAndGenlibRoundTrip(t *testing.T) {
	lib := Amap_LibReadBuffer("GATE BUF_BIG 2 Y=A;\nPIN A NONINV 1 2 3 4 5 6\nGATE BUF 1 Y=A;\nPIN A NONINV 1 2 3 4 5 6\nGATE INV 1 Y=!A;\nPIN A INV 1 2 3 4 5 6\n", 0)
	if lib == nil || Amap_LibParseEquations(lib, 0) == 0 {
		t.Fatal("parse")
	}
	selected := Amap_LibSelectGates(lib, 0)
	if vec.Vec_PtrSize(selected) != 2 || lib.pGateBuf.pName != "BUF" || lib.pGateInv.pName != "INV" {
		t.Fatal("area/primitive selection")
	}
	var text bytes.Buffer
	if err := Amap_LibWrite(&text, lib, 0); err != nil {
		t.Fatal(err)
	}
	copy := Amap_LibReadBuffer(text.String(), 0)
	if copy == nil || Amap_LibParseEquations(copy, 0) == 0 || len(copy.Gates()) != 3 {
		t.Fatalf("round trip: %s", text.String())
	}
	for i, g := range lib.Gates() {
		other := copy.Gates()[i]
		if g.pName != other.pName || g.dArea != other.dArea || g.pFunc[0] != other.pFunc[0] || g.Pins[0].dDelayFanoutFall != other.Pins[0].dDelayFanoutFall {
			t.Fatal("lost gate data")
		}
	}
}

type amapShortWriter struct{}

func (amapShortWriter) Write(p []byte) (int, error) { return len(p) / 2, nil }
func TestCheckedWriterShortWrite(t *testing.T) {
	w := &amapCheckedWriter{Writer: amapShortWriter{}}
	_, err := w.Write([]byte("word"))
	if !errors.Is(err, io.ErrShortWrite) {
		t.Fatal(err)
	}
	if n, err := w.Write(nil); n != 0 || !errors.Is(err, io.ErrShortWrite) {
		t.Fatal(n, err)
	}
}
