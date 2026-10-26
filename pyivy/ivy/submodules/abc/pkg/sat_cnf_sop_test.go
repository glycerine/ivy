package abc

import (
	"encoding/json"
	"os"
	"slices"
	"strings"
	"testing"
)

func cnfCoverTruth4(cover []uint8) uint16 {
	var result uint16
	for a := 0; a < 16; a++ {
		for _, cube := range cover {
			match := true
			for b := 0; b < 4; b++ {
				pair := cube >> uint(2*b) & 3
				if pair == 1 && a>>uint(b)&1 != 0 || pair == 2 && a>>uint(b)&1 == 0 {
					match = false
					break
				}
			}
			if match {
				result |= 1 << uint(a)
				break
			}
		}
	}
	return result
}
func TestCNFSOP4AllNative(t *testing.T) {
	data, e := os.ReadFile("test_vectors/oracle/cnf_sop4.bin")
	if e != nil {
		t.Fatal(e)
	}
	offset := 0
	for truth := 0; truth < 65536; truth++ {
		cover, e := CNFSOP4Cover(uint16(truth))
		if e != nil {
			t.Fatal(truth, e)
		}
		if offset+2 > len(data) {
			t.Fatal("truncated header", truth)
		}
		count, literals := int(data[offset]), int(data[offset+1])
		offset += 2
		if offset+count > len(data) || !slices.Equal(cover, data[offset:offset+count]) {
			t.Fatal("native cover", truth, cover)
		}
		offset += count
		var cubes [16]uint32
		for i, c := range cover {
			cubes[i] = uint32(c)
		}
		got, e := CNFISOPLiteralCount(cubes[:len(cover)], 4)
		if e != nil || got != literals {
			t.Fatal(truth, got, literals, e)
		}
		if value := cnfCoverTruth4(cover); value != uint16(truth) {
			t.Fatal("independent truth", truth, value, cover)
		}
	}
	if offset != len(data) {
		t.Fatal("trailing data")
	}
}
func TestCNFISOPCubeNative(t *testing.T) {
	b, e := os.ReadFile("test_vectors/oracle/cnf_isop_cube.json")
	if e != nil {
		t.Fatal(e)
	}
	var rows []struct {
		Cube      uint32
		Variables []SatVar
		Count     int
		Literals  []SatLit
	}
	if e = json.Unmarshal(b, &rows); e != nil {
		t.Fatal(e)
	}
	if len(rows) != 153 {
		t.Fatal(len(rows))
	}
	for _, row := range rows {
		l, e := CNFISOPCubeLiterals(row.Cube, row.Variables)
		if e != nil || !slices.Equal(l, row.Literals) {
			t.Fatal(row, l, e)
		}
		count, e := CNFISOPLiteralCount([]uint32{row.Cube}, len(row.Variables))
		if e != nil || count != row.Count {
			t.Fatal(row, count, e)
		}
	}
}
func TestCNFSOPStorageAndValidation(t *testing.T) {
	cover, e := CNFSOP4Cover(1)
	if e != nil {
		t.Fatal(e)
	}
	cover[0] = 0
	again, e := CNFSOP4Cover(1)
	if e != nil || cnfCoverTruth4(again) != 1 {
		t.Fatal(again, e)
	}
	for _, input := range []string{"", "! ", "\x00", strings.Repeat("!", 17) + " ", strings.Repeat("x", (1<<20)+1)} {
		if _, e := decodeCNFSOP4(input); e == nil {
			t.Fatal("bad table accepted")
		}
	}
	if _, e := decodeCNFSOP4(cnfSOP4Encoded + "ignored padding"); e != nil {
		t.Fatal(e)
	}
	if _, e := CNFISOPCubeLiterals(1, make([]SatVar, 17)); e == nil {
		t.Fatal("wide cube")
	}
	invalid := SatVar(^uint32(0))
	if _, e := CNFISOPCubeLiterals(1, []SatVar{invalid}); e == nil {
		t.Fatal("referenced invalid variable")
	}
	if l, e := CNFISOPCubeLiterals(3, []SatVar{invalid}); e != nil || len(l) != 0 {
		t.Fatal(l, e)
	}
	for _, n := range []int{-1, 17} {
		if _, e := CNFISOPLiteralCount(nil, n); e == nil {
			t.Fatal(n)
		}
	}
}
func FuzzCNFISOPCube(f *testing.F) {
	f.Add(uint32(0x87654321), uint8(16), uint16(32768))
	f.Fuzz(func(t *testing.T, cube uint32, size uint8, assignment uint16) {
		n := int(size % 17)
		vars := make([]SatVar, n)
		for i := range vars {
			vars[i] = SatVar(i)
		}
		l, e := CNFISOPCubeLiterals(cube, vars)
		if e != nil {
			t.Fatal(e)
		}
		matches := true
		for i := 0; i < n; i++ {
			pair := cube >> uint(2*i) & 3
			value := assignment>>uint(i)&1 != 0
			if pair == 1 && value || pair == 2 && !value {
				matches = false
			}
		}
		if satClausesHold([][]SatLit{l}, uint64(assignment)) == matches {
			t.Fatal(cube, n, assignment, l)
		}
		count, e := CNFISOPLiteralCount([]uint32{cube}, n)
		if e != nil || count != len(l) {
			t.Fatal(count, l, e)
		}
	})
}
func FuzzCNFSOP4(f *testing.F) {
	f.Add(uint16(0))
	f.Add(uint16(65535))
	f.Add(uint16(0x6996))
	f.Fuzz(func(t *testing.T, truth uint16) {
		cover, e := CNFSOP4Cover(truth)
		if e != nil || cnfCoverTruth4(cover) != truth {
			t.Fatal(truth, cover, e)
		}
	})
}
