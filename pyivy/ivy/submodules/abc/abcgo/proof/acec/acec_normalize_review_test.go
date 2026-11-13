package acec

import "testing"

func TestNormalizeBoothAndOrdinaryCarryChains(t *testing.T) {
	for _, booth := range []int32{0, 1} {
		p := reviewAdderChain(-1)
		original := make([][]bool, 32)
		for mask := 0; mask < 32; mask++ {
			in := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0, mask&8 != 0, mask&16 != 0}
			out, err := p.Evaluate(in)
			if err != nil {
				t.Fatal(err)
			}
			original[mask] = out
		}
		dst := Acec_Normalize(p, booth, 0)
		if dst == nil {
			t.Fatalf("booth %d: nil graph", booth)
		}
		for mask := 0; mask < 32; mask++ {
			in := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0, mask&8 != 0, mask&16 != 0}
			got, err := dst.Evaluate(in)
			if err != nil || len(got) != len(original[mask]) {
				t.Fatalf("booth %d mask %d got %v err %v", booth, mask, got, err)
			}
			for i, want := range original[mask] {
				if got[i] != want {
					t.Fatalf("booth %d input %v got %v want %v", booth, in, got, original[mask])
				}
			}
		}
	}
}
