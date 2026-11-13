package fraig

// Fixed-size memory management for FRAIG, adapted from fraigMem.c by Alan
// Mishchenko. Go slices replace C's intrusive free list and raw pointers.
import "github.com/glycerine/abcgo/misc/mem"

type Fraig_MemFixed_t = mem.Mem_Fixed_t

func Fraig_MemFixedStart(entrySize int32) *Fraig_MemFixed_t {
	return mem.Mem_FixedStart(entrySize)
}

func Fraig_MemFixedStop(p *Fraig_MemFixed_t, verbose int32) {
	mem.Mem_FixedStop(p, verbose)
}

func Fraig_MemFixedEntryFetch(p *Fraig_MemFixed_t) []byte {
	return mem.Mem_FixedEntryFetch(p)
}

func Fraig_MemFixedEntryRecycle(p *Fraig_MemFixed_t, entry []byte) {
	mem.Mem_FixedEntryRecycle(p, entry)
}

func Fraig_MemFixedRestart(p *Fraig_MemFixed_t) {
	mem.Mem_FixedRestart(p)
}

func Fraig_MemFixedReadMemUsage(p *Fraig_MemFixed_t) int32 {
	return mem.Mem_FixedReadMemUsage(p)
}
