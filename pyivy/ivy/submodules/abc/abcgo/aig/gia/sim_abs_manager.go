package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"math/bits"
	"time"
)

type Gia_SimAbsMan_t struct {
	pGia               *Gia_Man_t
	pSet               [2][]uint64
	nCands, nWords     int32
	vSims              *Vec_Wrd_t
	vResub             *Vec_Int_t
	fVerbose           int32
	vValues, vPatPairs *Vec_Int_t
	nWordsTable        int32
	pTableTemp         []uint64
	vCoverTable        *Vec_Wrd_t
	vTtMints           *Vec_Int_t
}

func Gia_Sim5CollectValues(off, on []uint64, words int32) *Vec_Int_t {
	result := vec.Vec_IntAlloc(int(64 * words))
	for bit := int32(0); bit < 64*words; bit++ {
		v := int32(-1)
		if giaSimBit(off, bit) != 0 {
			v = 0
		} else if giaSimBit(on, bit) != 0 {
			v = 1
		}
		vec.Vec_IntPush(result, v)
	}
	return result
}
func Gia_SimAbsAlloc(g *Gia_Man_t, off, on []uint64, sims *Vec_Wrd_t, words int32, resub *Vec_Int_t, verbose int32) *Gia_SimAbsMan_t {
	if words <= 0 || len(off) < int(words) || len(on) < int(words) || vec.Vec_WrdSize(sims)%int(words) != 0 || resub == nil {
		panic("gia: invalid abstraction input")
	}
	return &Gia_SimAbsMan_t{pGia: g, pSet: [2][]uint64{off, on}, nCands: int32(vec.Vec_WrdSize(sims)) / words, nWords: words, vSims: sims, vResub: resub, fVerbose: verbose, vValues: Gia_Sim5CollectValues(off, on, words), vPatPairs: vec.Vec_IntAlloc(100), vCoverTable: vec.Vec_WrdAlloc(10000), vTtMints: vec.Vec_IntAlloc(100)}
}
func Gia_SimAbsFree(p *Gia_SimAbsMan_t) {
	p.vValues = nil
	p.vPatPairs = nil
	p.vCoverTable = nil
	p.vTtMints = nil
	p.pTableTemp = nil
}
func Gia_SimAbsSolve(p *Gia_SimAbsMan_t) {
	start := time.Now()
	pairs := vec.Vec_IntArray(p.vPatPairs)
	if len(pairs)%2 != 0 {
		panic("gia: invalid pattern pairs")
	}
	p.nWordsTable = util.Abc_Bit6WordNum(int32(len(pairs) / 2))
	vec.Vec_WrdFill(p.vCoverTable, int(p.nWordsTable*(p.nCands+1)), 0)
	table := vec.Vec_WrdArray(p.vCoverTable)
	p.pTableTemp = table[p.nWordsTable*p.nCands:]
	sims := vec.Vec_WrdArray(p.vSims)
	for cand := int32(0); cand < p.nCands; cand++ {
		data := sims[cand*p.nWords : (cand+1)*p.nWords]
		cover := table[cand*p.nWordsTable : (cand+1)*p.nWordsTable]
		for i := 0; i < len(pairs); i += 2 {
			a, b := pairs[i], pairs[i+1]
			if vec.Vec_IntEntry(p.vValues, int(a)) != 0 || vec.Vec_IntEntry(p.vValues, int(b)) != 1 {
				panic("gia: invalid distinguishing pair")
			}
			if giaSimBit(data, a) != giaSimBit(data, b) {
				bit := int32(i / 2)
				cover[bit>>6] |= uint64(1) << uint(bit&63)
			}
		}
	}
	vec.Vec_IntClear(p.vResub)
	for bit := int32(0); bit < int32(len(pairs)/2); bit++ {
		p.pTableTemp[bit>>6] |= uint64(1) << uint(bit&63)
	}
	for {
		remaining := false
		for _, word := range p.pTableTemp {
			remaining = remaining || word != 0
		}
		if !remaining {
			break
		}
		best, bestCost := int32(-1), 0
		for cand := int32(0); cand < p.nCands; cand++ {
			score := 0
			for w := int32(0); w < p.nWordsTable; w++ {
				score += bits.OnesCount64(table[cand*p.nWordsTable+w] & p.pTableTemp[w])
			}
			if score > bestCost {
				best, bestCost = cand, score
			}
		}
		if best < 0 {
			panic("gia: candidates cannot distinguish pattern pairs")
		}
		vec.Vec_IntPush(p.vResub, best)
		for w := int32(0); w < p.nWordsTable; w++ {
			p.pTableTemp[w] &^= table[best*p.nWordsTable+w]
		}
	}
	if p.fVerbose != 0 {
		fmt.Printf("Solution %2d for covering problem [%5d x %5d]: %v   Time: %s\n", vec.Vec_IntSize(p.vResub), len(pairs)/2, p.nCands, vec.Vec_IntArray(p.vResub), time.Since(start))
	}
}
func Gia_SimAbsRefine(p *Gia_SimAbsMan_t) int32 {
	chosen := vec.Vec_IntArray(p.vResub)
	if len(chosen) == 0 || len(chosen) > 20 {
		panic("gia: invalid abstraction support")
	}
	vec.Vec_IntFill(p.vTtMints, 1<<uint(len(chosen)), -1)
	sims := vec.Vec_WrdArray(p.vSims)
	var count int32
	for pattern, value := range vec.Vec_IntArray(p.vValues) {
		if value < 0 {
			continue
		}
		var mint int32
		for i, cand := range chosen {
			if giaSimBit(sims[cand*p.nWords:(cand+1)*p.nWords], int32(pattern)) != 0 {
				mint |= 1 << uint(i)
			}
		}
		previous := vec.Vec_IntEntry(p.vTtMints, int(mint))
		if previous < 0 {
			vec.Vec_IntWriteEntry(p.vTtMints, int(mint), int32(pattern))
			continue
		}
		if giaSimBit(p.pSet[value], previous) != 0 {
			continue
		}
		a, b := int32(pattern), previous
		if value != 0 {
			a, b = previous, int32(pattern)
		}
		vec.Vec_IntPushTwo(p.vPatPairs, a, b)
		count++
		if count == 64 {
			return 1
		}
	}
	return giaBool(count != 0)
}
func Gia_SimAbsFind(values *Vec_Int_t, value int32) *Vec_Int_t {
	out := vec.Vec_IntAlloc(100)
	for i, v := range vec.Vec_IntArray(values) {
		if v == value {
			vec.Vec_IntPush(out, int32(i))
		}
	}
	return out
}
