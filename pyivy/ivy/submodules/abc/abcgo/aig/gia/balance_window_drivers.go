package gia

import "fmt"

// These drivers retain their original synthesis engines; the extracted window
// boundary is shared by SOP and DSD balancing without substituting algorithms.
func giaBalanceWindow(p *Gia_Man_t, target, window, verbose int32, balance func(*Gia_Man_t) *Gia_Man_t) *Gia_Man_t {
	if (target == 0) == (window == 0) {
		panic("gia: choose target level or percentage window")
	}
	outputs := Gia_ManFindLatest(p, target, window)
	if verbose != 0 {
		fmt.Printf("Collected %d outputs to extract.\n", Vec_IntSize(outputs))
	}
	if Vec_IntSize(outputs) == 0 {
		return Gia_ManDup(p)
	}
	extracted := Gia_ManExtractWin(p, outputs, 0)
	balanced := balance(extracted)
	Gia_ManStop(extracted)
	result := Gia_ManInsertWin(p, outputs, balanced)
	Gia_ManStop(balanced)
	return result
}
func Gia_ManPerformSopBalanceWin(p *Gia_Man_t, target, window, cuts, relax, verbose int32) *Gia_Man_t {
	return giaBalanceWindow(p, target, window, verbose, func(extracted *Gia_Man_t) *Gia_Man_t {
		return Gia_ManPerformSopBalance(extracted, cuts, relax, verbose)
	})
}
func Gia_ManPerformDsdBalanceWin(p *Gia_Man_t, target, window, lutSize, cuts, relax, verbose int32) *Gia_Man_t {
	return giaBalanceWindow(p, target, window, verbose, func(extracted *Gia_Man_t) *Gia_Man_t {
		return Gia_ManPerformDsdBalance(extracted, lutSize, cuts, relax, verbose)
	})
}
