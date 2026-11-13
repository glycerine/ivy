package core
import "testing"

func TestNPNAllFourVariableFunctions(t *testing.T) {
	table := Dar_Truth4VarNPN()
	if len(table.Representatives()) != 222 {
		t.Fatal("class count")
	}
	for truth := 0; truth < 65536; truth++ {
		match := table.Lookup(uint16(truth))
		rebuilt := uint16(0)
		for x := 0; x < 16; x++ {
			y := 0
			for i, p := range match.Permutation {
				bit := (x >> p) & 1
				if match.InputPhase&(1<<i) != 0 {
					bit ^= 1
				}
				y |= bit << i
			}
			value := match.Canonical&(1<<y) != 0
			if match.OutputComplement {
				value = !value
			}
			if value {
				rebuilt |= 1 << x
			}
		}
		if rebuilt != uint16(truth) {
			t.Fatalf("truth %04x canonical transform reconstructs %04x", truth, rebuilt)
		}
	}
}
func TestTruthPermutationRoundtrip(t *testing.T) {
	for _, p := range Dar_Permutations(4) {
		for _, truth := range []uint32{0, 1, 0x6996, 0x8000, 0xffff} {
			permuted := Dar_TruthPermute(truth, p, 4, false)
			if Dar_TruthPermute(permuted, p, 4, true) != truth {
				t.Fatal("inverse permutation", p, truth)
			}
		}
	}
	for phase := 0; phase < 16; phase++ {
		truth := uint32(0x69966996)
		if Dar_TruthPolarize(Dar_TruthPolarize(truth, phase, 4), phase, 4) != truth {
			t.Fatal("phase involution")
		}
	}
}
