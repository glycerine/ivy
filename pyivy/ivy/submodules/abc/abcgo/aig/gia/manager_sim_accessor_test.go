package gia

import (
	"testing"

	"github.com/glycerine/abcgo/misc/vec"
)

func TestGiaManSimObjectDataReturnsMutableObjectRow(t *testing.T) {
	g := Gia_ManStart(2)
	Gia_ManAppendObj(g)
	Gia_ManSetSimData(g, vec.Vec_WrdStart(4), 2)
	row := Gia_ManSimObjectData(g, 1)
	if len(row) != 2 {
		t.Fatalf("object row length = %d, want 2", len(row))
	}
	row[1] = 0x1234
	if got := Gia_ManSimData(g)[3]; got != 0x1234 {
		t.Fatalf("backing word = %#x, want 0x1234", got)
	}
	Gia_ManClearSimData(g)
	if Gia_ManSimWordNum(g) != 0 {
		t.Fatal("clear did not release simulation storage")
	}
}
