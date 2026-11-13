package fraig

import "testing"

func TestSimulationHashBeyondOriginalPrimeTable(t *testing.T) {
	const words = FRAIG_MAX_PRIMES + 1
	a := &Fraig_Node_t{Num: 1, NumPi: 0, puSimD: make([]uint32, words)}
	b := &Fraig_Node_t{Num: 2, NumPi: 1, puSimD: make([]uint32, words)}
	a.puSimD[words-1] = 0xdeadbeef
	b.puSimD[words-1] = 0xf0f0f0f0
	n := &Fraig_Node_t{Num: 3, NumPi: -1, p1: a, p2: b, puSimD: make([]uint32, words)}
	Fraig_NodeSimulate(n, words-1, words, 0)
	if want := a.puSimD[words-1] & b.puSimD[words-1]; n.puSimD[words-1] != want || n.uHashD != want*uint32(s_FraigPrimes[0]) {
		t.Fatal("simulation or wrapped hash incorrect")
	}
}
