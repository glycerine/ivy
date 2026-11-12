package gia

import "sync"

// BalanceEngines supplies the IF-owned algorithms used by GIA scripts. IF
// depends on GIA, so the engines are registered by IF at package startup.
type BalanceEngines struct {
	SOP func(*Gia_Man_t, int32, int32, int32) *Gia_Man_t
	DSD func(*Gia_Man_t, int32, int32, int32, int32) *Gia_Man_t
}

var giaBalanceEngines struct {
	sync.RWMutex
	engines BalanceEngines
}

func RegisterBalanceEngines(engines BalanceEngines) {
	giaBalanceEngines.Lock()
	giaBalanceEngines.engines = engines
	giaBalanceEngines.Unlock()
}

func Gia_ManPerformSopBalance(p *Gia_Man_t, cuts, relax, verbose int32) *Gia_Man_t {
	giaBalanceEngines.RLock()
	balance := giaBalanceEngines.engines.SOP
	giaBalanceEngines.RUnlock()
	if balance == nil {
		panic("gia: IF SOP balance engine is not registered")
	}
	return balance(p, cuts, relax, verbose)
}

func Gia_ManPerformDsdBalance(p *Gia_Man_t, lutSize, cuts, relax, verbose int32) *Gia_Man_t {
	giaBalanceEngines.RLock()
	balance := giaBalanceEngines.engines.DSD
	giaBalanceEngines.RUnlock()
	if balance == nil {
		panic("gia: IF DSD balance engine is not registered")
	}
	return balance(p, lutSize, cuts, relax, verbose)
}
