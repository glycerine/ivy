package abs

import "testing"

func TestCofactorTruthExhaustive(t *testing.T) {
	for variables := int32(1); variables <= 4; variables++ {
		for varied := int32(1); varied <= variables; varied++ {
			bits := 1 << uint(variables)
			for table := uint64(0); table < uint64(1)<<uint(bits); table++ {
				want := int32(1)
				step := 1 << uint(varied)
				for base := 0; base < bits; base += step {
					ones := 0
					for bit := base; bit < base+step; bit++ {
						ones += int(table >> uint(bit) & 1)
					}
					if ones == 0 || ones == step {
						want = 0
						break
					}
				}
				if got := Abs_GiaCheckTruth([]uint64{table}, variables, varied); got != want {
					t.Fatalf("variables=%d varied=%d truth=%x got=%d want=%d", variables, varied, table, got, want)
				}
			}
		}
	}
}
