package ivy

import (
	"fmt"
	"time"
)

func ivyResynReport(p *Ivy_Man_t, verbose int32, name string, start time.Time) {
	if verbose == 0 {
		return
	}
	fmt.Println()
	fmt.Printf("%s = %9.2f sec\n", name, time.Since(start).Seconds())
	Ivy_ManPrintStats(p)
}

func ivyResynBalance(p *Ivy_Man_t, update, verbose int32) *Ivy_Man_t {
	start := time.Now()
	old := p
	p = Ivy_ManBalance(old, update)
	Ivy_ManStop(old)
	ivyResynReport(p, verbose, "Balance", start)
	return p
}

func ivyResynRewrite(p *Ivy_Man_t, update, zeroCost, verbose int32) {
	start := time.Now()
	Ivy_ManRewritePre(p, update, zeroCost, 0)
	ivyResynReport(p, verbose, "Rewrite", start)
}

// Ivy_ManResyn0 applies balancing, one zero-cost rewrite pass and a final
// balance, preserving ABC's manager replacement and destruction lifecycle.
func Ivy_ManResyn0(p *Ivy_Man_t, update, verbose int32) *Ivy_Man_t {
	if verbose != 0 {
		fmt.Println("Original:")
		Ivy_ManPrintStats(p)
	}
	p = ivyResynBalance(p, update, verbose)
	ivyResynRewrite(p, update, 0, verbose)
	return ivyResynBalance(p, update, verbose)
}

// Ivy_ManResyn performs the ABC balance/rewrite/balance schedule, followed by
// two zero-cost rewrite passes with balancing after each pass.
func Ivy_ManResyn(p *Ivy_Man_t, update, verbose int32) *Ivy_Man_t {
	if verbose != 0 {
		fmt.Println("Original:")
		Ivy_ManPrintStats(p)
	}
	p = ivyResynBalance(p, update, verbose)
	ivyResynRewrite(p, update, 0, verbose)
	p = ivyResynBalance(p, update, verbose)
	ivyResynRewrite(p, update, 1, verbose)
	p = ivyResynBalance(p, update, verbose)
	ivyResynRewrite(p, update, 1, verbose)
	return ivyResynBalance(p, update, verbose)
}

// Ivy_ManRwsat is the SAT-oriented rewrite pass followed by one balance.
func Ivy_ManRwsat(p *Ivy_Man_t, verbose int32) *Ivy_Man_t {
	if verbose != 0 {
		fmt.Println("Original:")
		Ivy_ManPrintStats(p)
	}
	ivyResynRewrite(p, 0, 0, verbose)
	return ivyResynBalance(p, 0, verbose)
}
