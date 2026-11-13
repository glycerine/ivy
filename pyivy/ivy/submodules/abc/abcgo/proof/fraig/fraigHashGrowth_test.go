package fraig

import "testing"

func TestSimulationHashRejectsIndexBeyondPrimeTable(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("expected Go bounds check at end of C prime table")
		}
	}()
	_ = fraigSimulationPrime(FRAIG_MAX_PRIMES)
}
