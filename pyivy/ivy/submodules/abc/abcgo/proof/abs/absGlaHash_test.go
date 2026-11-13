package abs

import "testing"

func TestGlaHashHitMissAndCollision(t *testing.T) {
	p := &Ga2_Man_t{nTable: 1, pTable: make([]int32, 6)}
	a := []int32{2, 4, 6, 8, 10}
	b := []int32{3, 4, 6, 8, 10}
	*Saig_ManBmcLookup(p, a) = 23
	if got := *Saig_ManBmcLookup(p, a); got != 23 || p.nHashHit != 1 || p.nHashMiss != 1 {
		t.Fatalf("hit %d counters %d %d", got, p.nHashHit, p.nHashMiss)
	}
	if got := *Saig_ManBmcLookup(p, b); got != 0 || p.nHashOver != 1 {
		t.Fatalf("collision %d over %d", got, p.nHashOver)
	}
}
