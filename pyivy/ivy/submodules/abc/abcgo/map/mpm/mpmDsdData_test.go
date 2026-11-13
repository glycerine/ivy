package mpm

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestSixInputDSDLibrarySourceData(t *testing.T) {
	if len(s_DsdClass6) != 595 {
		t.Fatalf("classes=%d", len(s_DsdClass6))
	}
	h := sha256.New()
	for _, c := range s_DsdClass6 {
		var b [20]byte
		binary.LittleEndian.PutUint32(b[0:], uint32(c.nVars))
		binary.LittleEndian.PutUint32(b[4:], uint32(c.nAnds))
		binary.LittleEndian.PutUint32(b[8:], uint32(c.nClauses))
		binary.LittleEndian.PutUint64(b[12:], uint64(c.uTruth))
		h.Write(b[:])
		h.Write([]byte(c.pStr))
		h.Write([]byte{0})
	}
	if got := fmt.Sprintf("%x", h.Sum(nil)); got != "83c1a14c21772484efd079c8dd7bb30c4cb157df26fd9014a23bf432cbae7f54" {
		t.Fatalf("DSD table digest %s", got)
	}
}

func TestDSDTruthTablesFilterByClauseLimit(t *testing.T) {
	for _, limit := range []int32{-1, 0, 1, 4, 8, 20, 64, 1000} {
		result := Mpm_ManGetTruthWithCnf(limit)
		expected := 0
		for _, class := range s_DsdClass6 {
			if class.nClauses <= limit {
				if expected >= vec.Vec_WrdSize(result) || vec.Vec_WrdEntry(result, expected) != uint64(class.uTruth) {
					t.Fatalf("limit %d entry %d", limit, expected)
				}
				expected++
			}
		}
		if vec.Vec_WrdSize(result) != expected {
			t.Fatalf("limit %d: got %d expected %d", limit, vec.Vec_WrdSize(result), expected)
		}
	}
}
