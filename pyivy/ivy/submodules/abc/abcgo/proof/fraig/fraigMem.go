package fraig

// Fixed-size memory management for FRAIG, adapted from fraigMem.c by Alan
// Mishchenko. Go slices replace C's intrusive free list and raw pointers.
import (
	"unsafe"

	"github.com/glycerine/abcgo/misc/mem"
)

// fraigNodePool keeps Go nodes in typed backing arrays so the garbage
// collector can see their pointer fields while retaining the C fixed-pool
// allocation and accounting policy.
type fraigNodePool struct {
	chunks       [][]Fraig_Node_t
	free         []*Fraig_Node_t
	chunkEntries int
}

func fraigNodePoolStart() *fraigNodePool {
	entryBytes := int(unsafe.Sizeof(Fraig_Node_t{}))
	n := 1024
	if entryBytes*n >= 65536 {
		n = 65536 / entryBytes
	}
	if n < 8 {
		n = 8
	}
	return &fraigNodePool{chunkEntries: n}
}

func fraigNodePoolFetch(p *fraigNodePool) *Fraig_Node_t {
	if p == nil {
		return &Fraig_Node_t{}
	}
	if len(p.free) == 0 {
		chunk := make([]Fraig_Node_t, p.chunkEntries)
		p.chunks = append(p.chunks, chunk)
		for i := len(chunk) - 1; i >= 0; i-- {
			p.free = append(p.free, &chunk[i])
		}
	}
	i := len(p.free) - 1
	node := p.free[i]
	p.free = p.free[:i]
	*node = Fraig_Node_t{}
	return node
}

func fraigNodePoolStop(p *fraigNodePool) {
	if p != nil {
		p.chunks = nil
		p.free = nil
	}
}

func fraigNodePoolMemUsage(p *fraigNodePool) uintptr {
	if p == nil {
		return 0
	}
	return uintptr(len(p.chunks)*p.chunkEntries) * unsafe.Sizeof(Fraig_Node_t{})
}

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

// fraigMemFixedEntryWords views a pool entry as words. The pool's entries are
// word-sized and naturally aligned by Go's allocator, matching the C pool's
// unsigned-word simulation storage.
func fraigMemFixedEntryWords(p *Fraig_MemFixed_t, words int32) []uint32 {
	// A few internal algorithms also operate on synthetic managers assembled
	// directly by tests. Keep those stand-alone fixtures independent of pool
	// setup; production managers always have mmSims initialized.
	if p == nil {
		return make([]uint32, words)
	}
	entry := Fraig_MemFixedEntryFetch(p)
	bytes := int(words) * int(unsafe.Sizeof(uint32(0)))
	if len(entry) < bytes {
		panic("fraig: simulation pool entry is too small")
	}
	if words == 0 {
		return nil
	}
	return unsafe.Slice((*uint32)(unsafe.Pointer(&entry[0])), int(words))
}

func fraigMemFixedRecycleWords(p *Fraig_MemFixed_t, words []uint32) {
	if p == nil || len(words) == 0 {
		return
	}
	entry := unsafe.Slice((*byte)(unsafe.Pointer(&words[0])), cap(words)*int(unsafe.Sizeof(uint32(0))))
	Fraig_MemFixedEntryRecycle(p, entry)
}
