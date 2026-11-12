package gia

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadGigLutAndDelay(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sample.gig")
	data := "w2 = PI\nw3 = PI\nw4 = Lut4(w2,w3,-,-) [8888]\nw5 = Bar(w4)\nw6 = PO(w5)\nw7 = Del(w2,w3) [3]\nw8 = Sel(w7)\nw9 = PO(w8)\n"
	if err := os.WriteFile(path, []byte(data), 0600); err != nil {
		t.Fatal(err)
	}
	p := Gia_ManReadGig(path)
	if p == nil {
		t.Fatal("GIG read failed")
	}
	if Gia_ManCiNum(p) != 3 || Gia_ManCoNum(p) != 4 {
		t.Fatalf("inputs=%d outputs=%d", Gia_ManCiNum(p), Gia_ManCoNum(p))
	}
	for bits := 0; bits < 8; bits++ {
		got, err := p.Evaluate([]bool{bits&1 != 0, bits&2 != 0, bits&4 != 0})
		if err != nil {
			t.Fatal(err)
		}
		want := []bool{bits&1 != 0, bits&2 != 0, bits&3 == 3, bits&4 != 0}
		for i := range want {
			if got[i] != want[i] {
				t.Fatalf("input=%d output=%d got=%v want=%v", bits, i, got, want)
			}
		}
	}
}
