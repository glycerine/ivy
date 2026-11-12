package sfm

import "testing"

func TestDecompositionGateCnfs(t *testing.T) {
	cnfs := Sfm_DecCreateCnf([]int{0, 1, 2}, []uint64{0, 0xaaaaaaaaaaaaaaaa, 0x8888888888888888})
	for gate, clauses := range cnfs {
		for assignment := 0; assignment < 1<<uint(gate+1); assignment++ {
			valid := true
			for _, clause := range clauses {
				ok := false
				for _, lit := range clause {
					if (assignment>>uint(lit>>1)&1)^(lit&1) == 1 {
						ok = true
						break
					}
				}
				if !ok {
					valid = false
					break
				}
			}
			want := false
			if gate == 1 {
				want = assignment&1 != 0
			}
			if gate == 2 {
				want = assignment&3 == 3
			}
			if valid != ((assignment>>uint(gate)&1 != 0) == want) {
				t.Fatalf("gate %d assignment %d", gate, assignment)
			}
		}
	}
}
