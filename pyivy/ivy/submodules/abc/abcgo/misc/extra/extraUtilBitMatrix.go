package extra

import "fmt"

type Extra_BitMat_t = Extra_BitMat_t_

/**CFile****************************************************************

  FileName    [extraUtilBitMatrix.c]

  PackageName [extra]

  Synopsis    [Various reusable software utilities.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - September 1, 2003.]

  Revision    [$Id: extraUtilBitMatrix.c,v 1.0 2003/09/01 00:00:00 alanmi Exp $]

***********************************************************************/

// Extra_BitMat_t_ stores independent upper and lower triangular relations.
// Rows use explicit 32-bit words, matching ABC's unsigned representation.
type Extra_BitMat_t_ struct {
	ppData                       [][]uint32
	nSize, nWords                int32
	nBitShift                    int32
	uMask                        uint32
	nLookups, nInserts, nDeletes int32
}

func Extra_BitMatrixStart(nSize int32) *Extra_BitMat_t_ {
	if nSize < 0 {
		panic("negative bit matrix size")
	}
	p := &Extra_BitMat_t_{nSize: nSize, nWords: int32((int64(nSize) + 31) / 32), nBitShift: 5, uMask: 31}
	data := make([]uint32, int(nSize)*int(p.nWords))
	p.ppData = make([][]uint32, nSize)
	for i := range p.ppData {
		p.ppData[i] = data[i*int(p.nWords) : (i+1)*int(p.nWords)]
	}
	return p
}
func Extra_BitMatrixClean(p *Extra_BitMat_t_) {
	for _, row := range p.ppData {
		clear(row)
	}
}
func Extra_BitMatrixStop(p *Extra_BitMat_t_) {
	if p != nil {
		p.ppData = nil
	}
}
func Extra_BitMatrixReadSize(p *Extra_BitMat_t_) int32 { return p.nSize }
func bitMatrixOrder(i, k int32, upper bool) (int32, int32) {
	if (i > k) == upper {
		return k, i
	}
	return i, k
}
func (p *Extra_BitMat_t_) lookup(i, k int32) int32 {
	if p.ppData[i][k>>5]&(uint32(1)<<uint(k&31)) != 0 {
		return 1
	}
	return 0
}
func Extra_BitMatrixInsert1(p *Extra_BitMat_t_, i, k int32) {
	p.nInserts++
	i, k = bitMatrixOrder(i, k, true)
	p.ppData[i][k>>5] |= uint32(1) << uint(k&31)
}
func Extra_BitMatrixLookup1(p *Extra_BitMat_t_, i, k int32) int32 {
	p.nLookups++
	i, k = bitMatrixOrder(i, k, true)
	return p.lookup(i, k)
}
func Extra_BitMatrixDelete1(p *Extra_BitMat_t_, i, k int32) {
	p.nDeletes++
	i, k = bitMatrixOrder(i, k, true)
	p.ppData[i][k>>5] &^= uint32(1) << uint(k&31)
}
func Extra_BitMatrixInsert2(p *Extra_BitMat_t_, i, k int32) {
	p.nInserts++
	i, k = bitMatrixOrder(i, k, false)
	p.ppData[i][k>>5] |= uint32(1) << uint(k&31)
}
func Extra_BitMatrixLookup2(p *Extra_BitMat_t_, i, k int32) int32 {
	p.nLookups++
	i, k = bitMatrixOrder(i, k, false)
	return p.lookup(i, k)
}
func Extra_BitMatrixDelete2(p *Extra_BitMat_t_, i, k int32) {
	p.nDeletes++
	i, k = bitMatrixOrder(i, k, false)
	p.ppData[i][k>>5] &^= uint32(1) << uint(k&31)
}
func Extra_BitMatrixOr(p *Extra_BitMat_t_, i int32, info []uint32) {
	for w := range p.ppData[i] {
		p.ppData[i][w] |= info[w]
	}
}
func Extra_BitMatrixOrTwo(p *Extra_BitMat_t_, i, j int32) {
	for w := range p.ppData[i] {
		v := p.ppData[i][w] | p.ppData[j][w]
		p.ppData[i][w] = v
		p.ppData[j][w] = v
	}
}
func Extra_BitMatrixCountOnesUpper(p *Extra_BitMat_t_) int32 {
	var n int32
	for i := int32(0); i < p.nSize; i++ {
		for k := i + 1; k < p.nSize; k++ {
			n += p.lookup(i, k)
		}
	}
	return n
}
func Extra_BitMatrixIsDisjoint(p, q *Extra_BitMat_t_) int32 {
	if p.nSize != q.nSize {
		panic("bit matrix sizes differ")
	}
	for i, row := range p.ppData {
		for w, v := range row {
			if v&q.ppData[i][w] != 0 {
				return 0
			}
		}
	}
	return 1
}
func Extra_BitMatrixIsClique(p *Extra_BitMat_t_) int32 {
	for v := int32(0); v < p.nSize; v++ {
		for u := v + 1; u < p.nSize; u++ {
			if Extra_BitMatrixLookup1(p, v, u) == 0 {
				continue
			}
			for i := int32(0); i < p.nSize; i++ {
				if i != v && i != u && Extra_BitMatrixLookup1(p, i, v) != Extra_BitMatrixLookup1(p, i, u) {
					return 0
				}
			}
		}
	}
	return 1
}
func Extra_BitMatrixPrint(p *Extra_BitMat_t_) {
	fmt.Println()
	for i := int32(0); i < p.nSize; i++ {
		for k := int32(0); k <= i; k++ {
			fmt.Print(" ")
		}
		for k := i + 1; k < p.nSize; k++ {
			if Extra_BitMatrixLookup1(p, i, k) != 0 {
				fmt.Print("1")
			} else {
				fmt.Print(".")
			}
		}
		fmt.Println()
	}
}
