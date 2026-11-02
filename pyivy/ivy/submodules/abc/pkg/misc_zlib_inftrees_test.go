package abc

import "testing"

func TestZlibInflateTableFixed(t *testing.T) {
	var lens [320]uint16
	var sym int
	for ; sym < 144; sym++ {
		lens[sym] = 8
	}
	for ; sym < 256; sym++ {
		lens[sym] = 9
	}
	for ; sym < 280; sym++ {
		lens[sym] = 7
	}
	for ; sym < 288; sym++ {
		lens[sym] = 8
	}
	work := make([]uint16, 288)
	table := make([]Code, zlibEnoughTables)
	next := table
	bits := uint32(9)
	if r := Inflate_table(LENS, lens[:], 288, &next, &bits, work); r != 0 || bits != 9 {
		t.Fatalf("LENS: r=%d bits=%d", r, bits)
	}
	if len(table)-len(next) != 512 {
		t.Fatalf("used %d entries, want 512", len(table)-len(next))
	}
	// symbol 256 (end of block) has the 7-bit code 0000000
	if table[0].op != 96 || table[0].bits != 7 {
		t.Fatalf("entry 0 = %+v", table[0])
	}
	// symbol 0 has the 8-bit code 00110000 (stored reversed): literal, 8 bits
	idx := 0
	for i, b := 0, uint32(0x30); i < 8; i++ { // reverse the 8-bit code
		idx |= int((b>>uint(i))&1) << uint(7-i)
	}
	if e := table[idx]; e.op != 0 || e.bits != 8 || e.val != 0 {
		t.Fatalf("literal 0 entry = %+v", e)
	}
	// fixed distance code: 30 codes of 5 bits
	for i := range lens[:32] {
		lens[i] = 5
	}
	bits = 5
	dnext := make([]Code, zlibEnoughTables)
	dn := dnext
	if r := Inflate_table(DISTS, lens[:32], 32, &dn, &bits, work); r != 0 || bits != 5 {
		t.Fatalf("DISTS: r=%d bits=%d", r, bits)
	}
	if dnext[0].op != 16 || dnext[0].val != 1 { // distance code 0: base 1, 0 extra bits
		t.Fatalf("dist 0 = %+v", dnext[0])
	}
	// over-subscribed code
	over := []uint16{1, 1, 1}
	nn := make([]Code, 16)
	nb := uint32(2)
	if r := Inflate_table(CODES, over, 3, &nn, &nb, work); r != -1 {
		t.Fatalf("over-subscribed returned %d", r)
	}
}
