package fraig

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/util"
	"math/bits"
	"unsafe"
)

func Fraig_HashTableCreate(size int32) *Fraig_HashTable_t {
	n := util.Abc_PrimeCudd(uint32(size))
	return &Fraig_HashTable_t{nBins: n, pBins: make([]*Fraig_Node_t, n)}
}
func Fraig_HashTableFree(p *Fraig_HashTable_t) { p.pBins = nil; p.nBins = 0; p.nEntries = 0 }
func fraigStructuralKey(a, b *Fraig_Node_t, n int32) uint32 {
	x := uintptr(unsafe.Pointer(Fraig_Regular(a)))
	y := uintptr(unsafe.Pointer(Fraig_Regular(b)))
	if Fraig_IsComplement(a) {
		x |= 1
	}
	if Fraig_IsComplement(b) {
		y |= 1
	}
	return uint32((x + y*12582917) % uintptr(n))
}
func Fraig_HashTableLookupS(m *Fraig_Man_t, a, b *Fraig_Node_t, result **Fraig_Node_t) int32 {
	p := m.pTableS
	if Fraig_Regular(a).Num > Fraig_Regular(b).Num {
		a, b = b, a
	}
	key := fraigStructuralKey(a, b, p.nBins)
	for e := p.pBins[key]; e != nil; e = e.pNextS {
		if e.p1 == a && e.p2 == b {
			*result = e
			return 1
		}
	}
	if p.nEntries >= 2*p.nBins {
		Fraig_TableResizeS(p)
		key = fraigStructuralKey(a, b, p.nBins)
	}
	e := Fraig_NodeCreate(m, a, b)
	e.pNextS = p.pBins[key]
	p.pBins[key] = e
	p.nEntries++
	*result = e
	return 0
}
func Fraig_HashTableLookupF(m *Fraig_Man_t, node *Fraig_Node_t) *Fraig_Node_t {
	p := m.pTableF
	key := node.uHashR % uint32(p.nBins)
	for e := p.pBins[key]; e != nil; e = e.pNextF {
		if Fraig_CompareSimInfo(node, e, m.nWordsRand, 1) == 0 {
			continue
		}
		for d := e; d != nil; d = d.pNextD {
			if Fraig_CompareSimInfo(node, d, m.iWordStart, 0) != 0 {
				return d
			}
		}
		node.pNextD = e.pNextD
		e.pNextD = node
		return nil
	}
	if p.nEntries >= 2*p.nBins {
		Fraig_TableResizeF(p, 1)
		key = node.uHashR % uint32(p.nBins)
	}
	node.pNextF = p.pBins[key]
	p.pBins[key] = node
	p.nEntries++
	return nil
}
func Fraig_HashTableLookupF0(m *Fraig_Man_t, node *Fraig_Node_t) *Fraig_Node_t {
	p := m.pTableF0
	key := node.uHashD % uint32(p.nBins)
	for e := p.pBins[key]; e != nil; e = e.pNextF {
		if Fraig_CompareSimInfo(node, e, m.iWordStart, 0) != 0 {
			return e
		}
	}
	if p.nEntries >= 2*p.nBins {
		Fraig_TableResizeF(p, 0)
		key = node.uHashD % uint32(p.nBins)
	}
	node.pNextF = p.pBins[key]
	p.pBins[key] = node
	p.nEntries++
	return nil
}
func Fraig_HashTableInsertF0(m *Fraig_Man_t, node *Fraig_Node_t) {
	p := m.pTableF0
	k := node.uHashD % uint32(p.nBins)
	node.pNextF = p.pBins[k]
	p.pBins[k] = node
	p.nEntries++
}
func Fraig_TableResizeS(p *Fraig_HashTable_t) {
	n := util.Abc_PrimeCudd(uint32(2 * p.nBins))
	bins := make([]*Fraig_Node_t, n)
	var count int32
	for _, head := range p.pBins {
		for e := head; e != nil; {
			next := e.pNextS
			k := fraigStructuralKey(e.p1, e.p2, n)
			e.pNextS = bins[k]
			bins[k] = e
			count++
			e = next
		}
	}
	assert(count == p.nEntries, "structural table entry count")
	p.nBins = n
	p.pBins = bins
}
func Fraig_TableResizeF(p *Fraig_HashTable_t, random int32) {
	n := util.Abc_PrimeCudd(uint32(2 * p.nBins))
	bins := make([]*Fraig_Node_t, n)
	var count int32
	for _, head := range p.pBins {
		for e := head; e != nil; {
			next := e.pNextF
			hash := e.uHashD
			if random != 0 {
				hash = e.uHashR
			}
			k := hash % uint32(n)
			e.pNextF = bins[k]
			bins[k] = e
			count++
			e = next
		}
	}
	assert(count == p.nEntries, "functional table entry count")
	p.nBins = n
	p.pBins = bins
}
func fraigSimWords(a, b *Fraig_Node_t, random int32) ([]uint32, []uint32) {
	if random != 0 {
		return a.puSimR, b.puSimR
	}
	return a.puSimD, b.puSimD
}
func Fraig_CompareSimInfo(a, b *Fraig_Node_t, last, random int32) int32 {
	assert(!Fraig_IsComplement(a) && !Fraig_IsComplement(b), "simulation nodes are regular")
	if random != 0 && a.uHashR != b.uHashR || random == 0 && a.uHashD != b.uHashD {
		return 0
	}
	x, y := fraigSimWords(a, b, random)
	for i := int32(0); i < last; i++ {
		if x[i] != y[i] {
			return 0
		}
	}
	return 1
}
func Fraig_FindFirstDiff(a, b *Fraig_Node_t, compl, last, random int32) int32 {
	assert(!Fraig_IsComplement(a) && !Fraig_IsComplement(b), "simulation nodes are regular")
	invert := (compl != 0) != (a.fInv != b.fInv)
	x, y := fraigSimWords(a, b, random)
	for i := int32(0); i < last; i++ {
		d := x[i] ^ y[i]
		if invert {
			d = ^d
		}
		if d != 0 {
			return i*32 + int32(bits.TrailingZeros32(d))
		}
	}
	return -1
}
func Fraig_CompareSimInfoUnderMask(a, b *Fraig_Node_t, last, random int32, mask []uint32) int32 {
	x, y := fraigSimWords(a, b, random)
	for i := int32(0); i < last; i++ {
		if (x[i]^y[i])&mask[i] != 0 {
			return 0
		}
	}
	return 1
}
func Fraig_CollectXors(a, b *Fraig_Node_t, last, random int32, mask []uint32) {
	x, y := fraigSimWords(a, b, random)
	for i := int32(0); i < last; i++ {
		mask[i] = x[i] ^ y[i]
	}
}
func Fraig_TablePrintStatsS(m *Fraig_Man_t) {
	p := m.pTableS
	fmt.Printf("Structural table. Table size = %d. Number of entries = %d.\n", p.nBins, p.nEntries)
	for i, head := range p.pBins {
		n := 0
		for e := head; e != nil; e = e.pNextS {
			n++
		}
		if n > 1 {
			fmt.Printf("%d ", n)
			if n > 50 {
				fmt.Printf("{%d} ", i)
			}
		}
	}
	fmt.Println()
}
func Fraig_TablePrintStatsF(m *Fraig_Man_t) {
	p := m.pTableF
	fmt.Printf("Functional table. Table size = %d. Number of entries = %d.\n", p.nBins, p.nEntries)
	for _, head := range p.pBins {
		n := 0
		for e := head; e != nil; e = e.pNextF {
			n++
		}
		if n > 1 {
			fmt.Printf("{%d} ", n)
		}
	}
	fmt.Println()
}
func Fraig_TablePrintStatsF0(m *Fraig_Man_t) {
	p := m.pTableF0
	fmt.Printf("Zero-node table. Table size = %d. Number of entries = %d.\n\n", p.nBins, p.nEntries)
}
func Fraig_TableRehashF0(m *Fraig_Man_t, linkEquiv int32) int32 {
	p := m.pTableF0
	bins := make([]*Fraig_Node_t, p.nBins)
	var count, ret int32
	for _, head := range p.pBins {
		for f := head; f != nil; {
			nextF := f.pNextF
			for e := f; e != nil; {
				nextD := e.pNextD
				k := e.uHashD % uint32(p.nBins)
				var same *Fraig_Node_t
				if linkEquiv != 0 {
					for same = bins[k]; same != nil; same = same.pNextF {
						if same.uHashD == e.uHashD {
							break
						}
					}
				}
				if same != nil {
					e.pNextD = same.pNextD
					same.pNextD = e
					ret = 1
				} else {
					e.pNextF = bins[k]
					bins[k] = e
					e.pNextD = nil
				}
				count++
				e = nextD
			}
			f = nextF
		}
	}
	assert(count == p.nEntries, "zero table entry count")
	p.pBins = bins
	return ret
}
