package abci

import (
	"fmt"
	"io"
	"os"
	"slices"

	"github.com/glycerine/abcgo/bool/kit"
	"github.com/glycerine/abcgo/bool/lucky"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/opt/dau"
)

// Abc_TruthHashKey computes CUDD's seven-prime rolling hash over nWords.
func Abc_TruthHashKey(truth []word, nWords, tableSize int32) int32 {
	if tableSize <= 0 || nWords < 0 || int(nWords) > len(truth) {
		return -1
	}
	var key word
	for i := int32(0); i < nWords; i++ {
		key += truth[i] * word(npnTruthHashPrimes[i%int32(len(npnTruthHashPrimes))])
	}
	return int32(key % word(tableSize))
}

var npnTruthHashPrimes = [...]uint32{12582917, 25165843, 50331653, 100663319, 201326611, 402653189, 805306457}

func npnTruthKey(truth []word) string {
	b := make([]byte, 8*len(truth))
	for i, w := range truth {
		for j := 0; j < 8; j++ {
			b[8*i+j] = byte(w >> (8 * j))
		}
	}
	return string(b)
}

// Abc_TruthHashLookup follows the C collision chain for this function's key.
func Abc_TruthHashLookup(funcs [][]word, iThis, nWords int32, table, next []int32, key int32) int32 {
	if iThis < 0 || iThis >= int32(len(funcs)) || key < 0 || key >= int32(len(table)) || int(iThis) >= len(next) || nWords < 0 || int(nWords) > len(funcs[iThis]) {
		return 0
	}
	for that := table[key]; that != -1; that = next[that] {
		if that < 0 || that >= int32(len(funcs)) || int(nWords) > len(funcs[that]) {
			return 0
		}
		if slices.Equal(funcs[that][:nWords], funcs[iThis][:nWords]) {
			return 1
		}
	}
	return 0
}

// Abc_TruthNpnCountUnique drops duplicates while preserving first occurrence
// order, as the original hash-chain implementation does.
func Abc_TruthNpnCountUnique(p *Abc_TtStore_t) int32 {
	if p == nil {
		return 0
	}
	if p.nFuncs == 0 {
		return 0
	}
	tableSize := util.Abc_PrimeCudd(uint32(p.nFuncs))
	if tableSize <= 0 {
		return 0
	}
	table, next := make([]int32, tableSize), make([]int32, tableSize)
	for i := range table {
		table[i], next[i] = -1, -1
	}
	for i, truth := range p.pFuncs[:p.nFuncs] {
		key := Abc_TruthHashKey(truth, p.nWords, tableSize)
		if Abc_TruthHashLookup(p.pFuncs, int32(i), p.nWords, table, next, key) != 0 {
			p.pFuncs[i] = nil
		} else {
			next[i] = table[key]
			table[key] = int32(i)
		}
	}
	unique := p.pFuncs[:0]
	for _, truth := range p.pFuncs[:p.nFuncs] {
		if truth != nil {
			unique = append(unique, truth)
		}
	}
	p.pFuncs = unique
	p.nFuncs = int32(len(unique))
	return p.nFuncs
}

var npnCompareWords int32

func Abc_TruthCompare(a, b []word) int32 {
	words := int(npnCompareWords)
	if words < 0 || words > len(a) || words > len(b) {
		return 0
	}
	return int32(slices.Compare(a[:words], b[:words]))
}

func Abc_TruthNpnCountUniqueSort(p *Abc_TtStore_t) int32 {
	if p == nil {
		return 0
	}
	npnCompareWords = p.nWords
	slices.SortFunc(p.pFuncs[:p.nFuncs], func(a, b []word) int { return int(Abc_TruthCompare(a, b)) })
	unique := p.pFuncs[:0]
	for _, truth := range p.pFuncs[:p.nFuncs] {
		if len(unique) == 0 || !slices.Equal(unique[len(unique)-1][:p.nWords], truth[:p.nWords]) {
			unique = append(unique, truth)
		}
	}
	p.pFuncs = unique
	p.nFuncs = int32(len(unique))
	return p.nFuncs
}

// Abc_TruthNpnPrint writes the variable and phase mapping emitted by C.
func Abc_TruthNpnPrint(permutation []byte, phase uint32, nVars int32) {
	if nVars < 0 || nVars > 16 {
		return
	}
	if len(permutation) < int(nVars) {
		permutation = make([]byte, nVars)
		for i := range permutation {
			permutation[i] = byte('a' + i)
		}
	}
	if phase&(uint32(1)<<uint(nVars)) != 0 {
		fmt.Printf("   Z = ( ")
	} else {
		fmt.Printf("   z = ( ")
	}
	for i := int32(0); i < nVars; i++ {
		v := permutation[i]
		if v < 'a' || v >= 'a'+byte(nVars) {
			v = byte('a' + i)
		}
		if phase&(uint32(1)<<uint(v-'a')) != 0 {
			v -= 'a' - 'A'
		}
		if i+1 == nVars {
			fmt.Printf("%c", v)
		} else {
			fmt.Printf("%c,", v)
		}
	}
	fmt.Printf(" )  ")
}

func truthAsUint32(truth []word, nVars int32) []uint32 {
	n := int(kit.Kit_TruthWordNum(nVars))
	out := make([]uint32, n)
	for i := range out {
		w := i / 2
		if w < len(truth) {
			if i&1 == 0 {
				out[i] = uint32(truth[w])
			} else {
				out[i] = uint32(truth[w] >> 32)
			}
		}
	}
	return out
}

func uint32AsTruth(dst []word, src []uint32) {
	for i := range dst {
		lo, hi := uint32(0), uint32(0)
		if 2*i < len(src) {
			lo = src[2*i]
		}
		if 2*i+1 < len(src) {
			hi = src[2*i+1]
		}
		dst[i] = word(lo) | word(hi)<<32
	}
}

func printNpnTruth(w io.Writer, truth []word, nVars int32) {
	fmt.Fprint(w, "0x")
	if len(truth) > 0 {
		for k := int32(1)<<(uint(nVars-2)) - 1; k >= 0; k-- {
			fmt.Fprintf(w, "%X", Abc_TruthGetHex(&truth[0], k))
		}
	}
}

// Abc_TruthNpnPerform applies the selected source canonicalization method,
// then reports the number of resulting truth-table classes.
func Abc_TruthNpnPerform(p *Abc_TtStore_t, npnType, verbose int32) {
	if p == nil || p.nVars < 0 || p.nVars > 16 || p.nFuncs < 0 {
		return
	}
	names := [...]string{"uniqifying          ", "exact NPN           ", "counting 1s         ", "Jake's hybrid fast  ", "Jake's hybrid good  ", "new hybrid fast     ", "new phase flipping  ", "new hier. matching  ", "new adap. matching  ", "adjustable algorithm (heuristic) ", "adjustable algorithm (exact)     ", "new cost-aware exact algorithm   "}
	if npnType < 0 || npnType >= int32(len(names)) {
		return
	}
	start := util.Abc_Clock()
	fmt.Printf("Applying %-20s to %8d func%s of %2d vars...  ", names[npnType], p.nFuncs, plural(p.nFuncs), p.nVars)
	if verbose != 0 {
		fmt.Println()
	}
	if npnType == 1 {
		Abc_TruthNpnCountUnique(p)
	}
	perm := make([]byte, 16)
	phase := uint32(0)
	var hieManager *dau.Abc_TtHieMan_t
	if npnType >= 7 {
		hieManager = dau.Abc_TtHieManStart(p.nVars, 5)
	}
	for i := int32(0); i < p.nFuncs; i++ {
		truth := p.pFuncs[i]
		if verbose != 0 {
			fmt.Printf("%7d : ", i)
		}
		switch npnType {
		case 0:
			// Uniquifying alone leaves every truth table unchanged.
		case 1:
			if p.nVars > 12 {
				fmt.Printf("Exact permutation search is limited to 12 variables.\n")
				return
			}
			info := lucky.NewPermInfo(p.nVars)
			aux, minimal := make([]uint64, p.nWords), make([]uint64, p.nWords)
			lucky.SimpleMinimal(truth, aux, minimal, info, p.nVars)
			phase = 0
			for j := int32(0); j < p.nVars; j++ {
				perm[j] = byte('a' + j)
			}
		case 2:
			perm = perm[:16]
			lucky.ResetCanonPermutation(perm, p.nVars)
			table := truthAsUint32(truth, p.nVars)
			phase = kit.Kit_TruthSemiCanonicize(table, nil, p.nVars, perm)
			uint32AsTruth(truth, table)
		case 3:
			lucky.ResetCanonPermutation(perm, p.nVars)
			phase = lucky.LuckyCanonicizerFinalFast(truth, p.nVars, perm)
		case 4:
			lucky.ResetCanonPermutation(perm, p.nVars)
			phase = lucky.LuckyCanonicizerFinalFast1(truth, p.nVars, perm)
		case 5:
			phase = dau.Abc_TtCanonicize(truth, p.nVars, perm)
		case 6:
			phase = dau.Abc_TtCanonicizePhase(truth, p.nVars)
		case 7:
			phase = dau.Abc_TtCanonicizeHie(hieManager, truth, p.nVars, perm, 0)
		case 8:
			phase = dau.Abc_TtCanonicizeWrap(dau.Abc_TtCanonicizeAda, hieManager, truth, p.nVars, perm, 125)
		case 9, 10, 11:
			if npnType == 9 {
				phase = dau.Abc_TtCanonicizeWrap(dau.Abc_TtCanonicizeAda, hieManager, truth, p.nVars, perm, 125)
			} else if npnType == 10 {
				phase = dau.Abc_TtCanonicizeWrap(dau.Abc_TtCanonicizeAda, hieManager, truth, p.nVars, perm, 1199)
			} else {
				phase = dau.Abc_TtCanonicizeWrap(dau.Abc_TtCanonicizeCA, hieManager, truth, p.nVars, perm, 1)
			}
		}
		if verbose != 0 {
			if npnType == 7 {
				// The C hierarchical mode intentionally emits only the row break.
				fmt.Println()
			} else {
				printNpnTruth(os.Stdout, truth, p.nVars)
				if npnType != 0 {
					if npnType == 6 {
						Abc_TruthNpnPrint(nil, phase, p.nVars)
					} else {
						Abc_TruthNpnPrint(perm, phase, p.nVars)
					}
				}
				fmt.Println()
			}
		}
	}
	if hieManager != nil {
		dau.Abc_TtHieManStop(hieManager)
	}
	fmt.Printf("Classes =%9d  ", Abc_TruthNpnCountUnique(p))
	util.Abc_PrintTime(1, "Time", util.Abc_Clock()-start)
}

func plural(n int32) string {
	if n == 1 {
		return ""
	}
	return "s"
}

func Abc_TruthNpnTest(name string, npnType, nVarNum, dump, binaryOutput, verbose int32) {
	p := Abc_TtStoreLoad(name, nVarNum)
	if p == nil {
		return
	}
	Abc_TruthNpnPerform(p, npnType, verbose)
	if dump != 0 {
		ext := "_out.txt"
		if binaryOutput != 0 {
			ext = "_out.tt"
		}
		out := Extra_FileNameGenericAppend(name, ext)
		Abc_TtStoreWrite(out, p, binaryOutput)
		if verbose != 0 {
			fmt.Printf("The resulting functions are written into file %q.\n", out)
		}
	}
	Abc_TtStoreFree(p, nVarNum)
}

func Abc_NpnTest(name string, npnType, nVarNum, dump, binaryOutput, verbose int32) int32 {
	if verbose != 0 {
		fmt.Printf("Using truth tables from file %q...\n", name)
	}
	if npnType >= 0 && npnType <= 11 {
		Abc_TruthNpnTest(name, npnType, nVarNum, dump, binaryOutput, verbose)
	} else {
		fmt.Printf("Unknown canonical form value (%d).\n", npnType)
	}
	return 0
}
