package gia

import "testing"

func TestBalanceEnginesDispatch(t *testing.T) {
	giaBalanceEngines.RLock()
	old := giaBalanceEngines.engines
	giaBalanceEngines.RUnlock()
	defer RegisterBalanceEngines(old)
	p := Gia_ManStart(1)
	calls := 0
	RegisterBalanceEngines(BalanceEngines{
		SOP: func(g *Gia_Man_t, cuts, relax, verbose int32) *Gia_Man_t {
			if g != p || cuts != 4 || relax != 5 || verbose != 6 {
				t.Fatal("wrong SOP arguments")
			}
			calls++
			return g
		},
		DSD: func(g *Gia_Man_t, lut, cuts, relax, verbose int32) *Gia_Man_t {
			if g != p || lut != 3 || cuts != 4 || relax != 5 || verbose != 6 {
				t.Fatal("wrong DSD arguments")
			}
			calls++
			return g
		},
	})
	if Gia_ManPerformSopBalance(p, 4, 5, 6) != p || Gia_ManPerformDsdBalance(p, 3, 4, 5, 6) != p || calls != 2 {
		t.Fatal("balance dispatch failed")
	}
}
