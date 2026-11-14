package abci

import (
	"encoding/binary"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/glycerine/abcgo/bool/rpo"
	"github.com/glycerine/abcgo/misc/util"
)

// Rpo_TtStore_t holds the same 64-bit truth-table words used by ABC's word
// representation. The backing slice is contiguous for owned tables; pFuncs
// provides the per-function views used by the original C pointer table.
type Rpo_TtStore_t struct {
	nVars, nWords, nFuncs int32
	pFuncs                [][]word
	backing               []word
	ownsBacking           bool
}

func rpoTruthWordNum64(nVars int32) int32 {
	if nVars < 7 {
		return 1
	}
	return int32(1) << uint(nVars-6)
}

func Abc_RpoTruthStoreAlloc(nVars, nFuncs int32) *Rpo_TtStore_t {
	if nVars < 2 || nVars > 16 || nFuncs < 0 {
		return nil
	}
	nWords := rpoTruthWordNum64(nVars)
	backing := make([]word, int(nWords)*int(nFuncs))
	p := &Rpo_TtStore_t{nVars: nVars, nWords: nWords, nFuncs: nFuncs, backing: backing, ownsBacking: true}
	p.pFuncs = make([][]word, int(nFuncs))
	for i := range p.pFuncs {
		p.pFuncs[i] = backing[i*int(nWords) : (i+1)*int(nWords)]
	}
	return p
}

// Alloc2 wraps caller-owned words, matching the C routine's borrowed-buffer
// path. The wrapper never mutates the supplied tables while creating views.
func Abc_RpoTruthStoreAlloc2(nVars, nFuncs int32, buffer []word) *Rpo_TtStore_t {
	if nVars < 2 || nVars > 16 || nFuncs < 0 {
		return nil
	}
	nWords := rpoTruthWordNum64(nVars)
	need := int(nWords) * int(nFuncs)
	if len(buffer) < need {
		return nil
	}
	p := &Rpo_TtStore_t{nVars: nVars, nWords: nWords, nFuncs: nFuncs, backing: buffer[:need]}
	p.pFuncs = make([][]word, int(nFuncs))
	for i := range p.pFuncs {
		p.pFuncs[i] = p.backing[i*int(nWords) : (i+1)*int(nWords)]
	}
	return p
}

func Abc_RpoTtStoreFree(p *Rpo_TtStore_t, nVarNum int32) {
	if p == nil {
		return
	}
	// C frees the table chunk only for nonnegative nVarNum. Go releases owned
	// memory by dropping references; borrowed storage remains with its caller.
	if nVarNum >= 0 && p.ownsBacking {
		p.backing = nil
	}
	p.pFuncs = nil
	p.backing = nil
	p.nFuncs = 0
}

func Abc_RpoTruthStoreRead(fileName string, p *Rpo_TtStore_t) {
	if p == nil || p.nFuncs == 0 {
		return
	}
	contents := Abc_FileRead(fileName)
	if contents == "" {
		return
	}
	count := int32(0)
	for _, line := range strings.Split(contents, "\n") {
		if line == "" || line[0] == 0 {
			break
		}
		if count >= int32(len(p.pFuncs)) {
			break
		}
		if len(p.pFuncs[count]) == 0 {
			break
		}
		Abc_TruthReadHex(&p.pFuncs[count][0], line, p.nVars)
		count++
	}
	p.nFuncs = count
}

func Abc_RpoTtStoreWrite(fileName string, p *Rpo_TtStore_t, binaryOutput int32) {
	if p == nil {
		return
	}
	f, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		fmt.Printf("Cannot open file %q for writing.\n", fileName)
		return
	}
	defer f.Close()
	for _, truth := range p.pFuncs[:p.nFuncs] {
		if binaryOutput != 0 {
			var encoded [8]byte
			for _, w := range truth {
				binary.LittleEndian.PutUint64(encoded[:], uint64(w))
				_, _ = f.Write(encoded[:])
			}
			continue
		}
		if len(truth) != 0 {
			Abc_TruthWriteHex(f, &truth[0], p.nVars)
		}
		_, _ = fmt.Fprintln(f)
	}
}

func Abc_RpoTtStoreLoad(fileName string, nVarNum int32) *Rpo_TtStore_t {
	if nVarNum < 0 {
		var nVars, nTruths int32
		Abc_TruthGetParams(fileName, &nVars, &nTruths)
		if nVars < 2 || nVars > 16 || nTruths == 0 {
			return nil
		}
		p := Abc_RpoTruthStoreAlloc(nVars, nTruths)
		if p == nil {
			return nil
		}
		Abc_RpoTruthStoreRead(fileName, p)
		return p
	}
	if nVarNum < 6 || nVarNum > 16 {
		return nil
	}
	nFileSize := Abc_FileSize(fileName)
	if nFileSize < 0 {
		return nil
	}
	// Preserve the source's byte count: binary tables are sized by 2^(n-3).
	nBytes := int32(1) << uint(nVarNum-3)
	if nBytes <= 0 {
		return nil
	}
	nTruths := nFileSize / nBytes
	if nFileSize%nBytes != 0 {
		util.Abc_Print(0, "The file size (%d) is divided by the truth table size (%d) with remainder (%d).\n", nFileSize, nBytes, nFileSize%nBytes)
	}
	data, err := os.ReadFile(fileName)
	if err != nil {
		return nil
	}
	p := Abc_RpoTruthStoreAlloc(nVarNum, nTruths)
	if p == nil {
		return nil
	}
	for i := int32(0); i < nTruths; i++ {
		start := int(i * nBytes)
		end := start + int(nBytes)
		if end > len(data) {
			end = len(data)
		}
		truthBytes := data[start:end]
		for j := 0; j < len(p.pFuncs[i]) && j*8 < len(truthBytes); j++ {
			var wordBytes [8]byte
			copy(wordBytes[:], truthBytes[j*8:min(j*8+8, len(truthBytes))])
			p.pFuncs[i][j] = word(binary.LittleEndian.Uint64(wordBytes[:]))
		}
	}
	return p
}

// RPO's factorizer consumes ABC's uint32 truth layout. C passes a word* as an
// unsigned*, which means the low then high halves of each 64-bit word; for up
// to five variables only the first 32-bit word is significant in RPO's layout.
func rpoTruth32(truth []word, nVars int32) []uint32 {
	nWords := 1
	if nVars > 5 {
		nWords = 1 << uint(nVars-5)
	}
	result := make([]uint32, nWords)
	for i := range result {
		wordIndex := i / 2
		if wordIndex >= len(truth) {
			break
		}
		if i&1 == 0 {
			result[i] = uint32(truth[wordIndex])
		} else {
			result[i] = uint32(truth[wordIndex] >> 32)
		}
	}
	return result
}

func Abc_TruthRpoPerform(p *Rpo_TtStore_t, nThreshold, fVerbose int32) {
	if p == nil || p.nFuncs == 0 {
		return
	}
	start := time.Now()
	rpoCount := int32(0)
	for i := int32(0); i < p.nFuncs; i++ {
		if fVerbose != 0 {
			util.Abc_Print(-2, "%d: ", i+1)
		}
		literal := rpo.Rpo_Factorize(rpoTruth32(p.pFuncs[i], p.nVars), p.nVars, nThreshold, fVerbose)
		if literal == nil {
			if fVerbose != 0 {
				util.Abc_Print(-2, "null\n\n")
			}
			continue
		}
		if fVerbose != 0 {
			util.Abc_Print(-2, "Solution : %s\n\n", literal.Expression)
		}
		rpo.Lit_Free(literal)
		rpoCount++
	}
	percent := float64(rpoCount) * 100.0 / float64(p.nFuncs)
	util.Abc_Print(-2, "%d of %d (%.2f %%) functions are RPO.\n", rpoCount, p.nFuncs, percent)
	util.Abc_PrintTime(1, "Time", time.Since(start).Nanoseconds())
}

func Abc_TruthRpoTest(fileName string, nVarNum, nThreshold, fVerbose int32) {
	p := Abc_RpoTtStoreLoad(fileName, nVarNum)
	if p == nil {
		return
	}
	defer Abc_RpoTtStoreFree(p, nVarNum)
	if fVerbose != 0 {
		util.Abc_Print(-2, "Number of variables = %d\n", p.nVars)
	}
	Abc_TruthRpoPerform(p, nThreshold, fVerbose)
}

func Abc_RpoTest(fileName string, nVarNum, nThreshold, fVerbose int32) int32 {
	if fVerbose != 0 {
		fmt.Printf("Using truth tables from file %q...\n", fileName)
	}
	Abc_TruthRpoTest(fileName, nVarNum, nThreshold, fVerbose)
	return 0
}
