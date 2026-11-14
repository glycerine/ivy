package abci

import (
	"encoding/binary"
	"fmt"
	"os"
	"strings"
	"unsafe"

	"github.com/glycerine/abcgo/bool/bdc"
	"github.com/glycerine/abcgo/bool/dec"
	"github.com/glycerine/abcgo/bool/deco"
	"github.com/glycerine/abcgo/bool/kit"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/opt/dau"
	"github.com/glycerine/abcgo/opt/dsc"
)

// Abc_TtStore_t owns a set of 64-bit truth tables. pFuncs mirrors the C
// pointer table, while backing keeps a single contiguous allocation for the
// ordinary allocator and for borrowed binary input buffers.
type Abc_TtStore_t struct {
	nVars, nWords, nFuncs int32
	pFuncs                [][]word
	backing               []word
	ownsBacking           bool
}

func abcTruthWordNum(nVars int32) int32 {
	if nVars < 7 {
		return 1
	}
	return int32(1) << uint(nVars-6)
}

func wordPointerSlice(p *word, index int32) []word {
	if p == nil || index < 0 {
		return nil
	}
	return unsafe.Slice(p, int(index)+1)
}

func Abc_TruthGetBit(p *word, i int32) int32 {
	return int32(wordPointerSlice(p, i>>6)[i>>6]>>(uint(i)&63)) & 1
}
func Abc_TruthSetBit(p *word, i int32) {
	x := wordPointerSlice(p, i>>6)
	x[i>>6] |= word(1) << (uint(i) & 63)
}
func Abc_TruthXorBit(p *word, i int32) {
	x := wordPointerSlice(p, i>>6)
	x[i>>6] ^= word(1) << (uint(i) & 63)
}
func Abc_TruthGetHex(p *word, k int32) int32 {
	return int32(wordPointerSlice(p, k>>4)[k>>4]>>(uint(k&15)*4)) & 15
}
func Abc_TruthSetHex(p *word, k, d int32) {
	x := wordPointerSlice(p, k>>4)
	x[k>>4] |= word(d) << (uint(k&15) * 4)
}
func Abc_TruthXorHex(p *word, k, d int32) {
	x := wordPointerSlice(p, k>>4)
	x[k>>4] ^= word(d) << (uint(k&15) * 4)
}

func Abc_TruthReadHexDigit(c byte) int32 {
	switch {
	case c >= '0' && c <= '9':
		return int32(c - '0')
	case c >= 'A' && c <= 'F':
		return int32(c-'A') + 10
	case c >= 'a' && c <= 'f':
		return int32(c-'a') + 10
	default:
		return -1
	}
}

func Abc_TruthWriteHexDigit(f *os.File, digit int32) {
	if digit < 0 || digit > 15 || f == nil {
		return
	}
	if digit < 10 {
		_, _ = fmt.Fprintf(f, "%d", digit)
	} else {
		_, _ = fmt.Fprintf(f, "%c", 'A'+digit-10)
	}
}

// Abc_TruthReadHex reads the source's most-significant-digit-first textual
// encoding into a zero-initialized table.
func Abc_TruthReadHex(truth *word, text string, nVars int32) {
	if truth == nil || nVars < 2 || nVars > 16 {
		return
	}
	if strings.HasPrefix(text, "0x") || strings.HasPrefix(text, "0X") {
		text = text[2:]
	}
	digits := 1 << uint(nVars-2)
	if len(text) < digits {
		return
	}
	for k := 0; k < digits; k++ {
		digit := Abc_TruthReadHexDigit(text[digits-1-k])
		if digit < 0 {
			return
		}
		Abc_TruthSetHex(truth, int32(k), digit)
	}
}

func Abc_TruthWriteHex(f *os.File, truth *word, nVars int32) {
	if f == nil || truth == nil || nVars < 2 || nVars > 16 {
		return
	}
	for k, digits := int32(0), int32(1)<<uint(nVars-2); k < digits; k++ {
		Abc_TruthWriteHexDigit(f, Abc_TruthGetHex(truth, k))
	}
}

func Abc_TruthStoreAlloc(nVars, nFuncs int32) *Abc_TtStore_t {
	if nVars < 2 || nVars > 16 || nFuncs < 0 {
		return nil
	}
	nWords := abcTruthWordNum(nVars)
	backing := make([]word, int(nWords)*int(nFuncs))
	p := &Abc_TtStore_t{nVars: nVars, nWords: nWords, nFuncs: nFuncs, backing: backing, ownsBacking: true}
	p.pFuncs = make([][]word, nFuncs)
	for i := range p.pFuncs {
		p.pFuncs[i] = backing[i*int(nWords) : (i+1)*int(nWords)]
	}
	return p
}

// Abc_TruthStoreAlloc2 slices a caller-owned contiguous truth-table buffer.
func Abc_TruthStoreAlloc2(nVars, nFuncs int32, buffer *word) *Abc_TtStore_t {
	if nVars < 2 || nVars > 16 || nFuncs < 0 || (nFuncs > 0 && buffer == nil) {
		return nil
	}
	nWords := abcTruthWordNum(nVars)
	all := unsafe.Slice(buffer, int(nWords*nFuncs))
	p := &Abc_TtStore_t{nVars: nVars, nWords: nWords, nFuncs: nFuncs, backing: all}
	p.pFuncs = make([][]word, nFuncs)
	for i := range p.pFuncs {
		p.pFuncs[i] = all[i*int(nWords) : (i+1)*int(nWords)]
	}
	return p
}

func Abc_TtStoreFree(p *Abc_TtStore_t, nVarNum int32) {
	if p == nil {
		return
	}
	// The C API's nVarNum controls whether pFuncs[0] is freed. Go's allocator
	// owns all slices, so dropping the backing reference is the corresponding
	// release operation; borrowed storage is never cleared.
	if nVarNum >= 0 && p.ownsBacking {
		p.backing = nil
	}
	p.pFuncs = nil
	p.backing = nil
	p.nFuncs = 0
}

func Abc_FileSize(name string) int32 {
	info, err := os.Stat(name)
	if err != nil {
		fmt.Printf("Cannot open file %q for reading.\n", name)
		return -1
	}
	if info.Size() > int64(^uint32(0)>>1) {
		return -1
	}
	return int32(info.Size())
}

// Extra_FileNameGenericAppend keeps the suffix after the final filename
// extension, matching the helper used by the original truth-table tools.
func Extra_FileNameGenericAppend(name, suffix string) string {
	slash := strings.LastIndexAny(name, "/\\")
	if dot := strings.LastIndexByte(name, '.'); dot > slash {
		name = name[:dot]
	}
	return name + suffix
}

// Abc_FileRead preserves the C helper's two trailing newlines and NUL
// terminator because its line-count and parser callers depend on them.
func Abc_FileRead(name string) string {
	b, err := os.ReadFile(name)
	if err != nil {
		fmt.Printf("Cannot open file %q for reading.\n", name)
		return ""
	}
	return string(b) + "\n\n\x00"
}

func Abc_TruthGetParams(name string, pnVars, pnTruths *int32) {
	if pnVars != nil {
		*pnVars = 0
	}
	if pnTruths != nil {
		*pnTruths = 0
	}
	contents := Abc_FileRead(name)
	if contents == "" {
		return
	}
	line := contents
	if i := strings.IndexAny(line, " \r\n\x00"); i >= 0 {
		line = line[:i]
	}
	line = strings.TrimPrefix(strings.TrimPrefix(line, "0x"), "0X")
	if len(line) == 0 {
		fmt.Println("Does not look like the input file contains truth tables...")
		return
	}
	nVars := int32(-1)
	for n := int32(2); n <= 16; n++ {
		if len(line) == 1<<uint(n-2) {
			nVars = n
			break
		}
	}
	if nVars < 2 || nVars > 16 {
		fmt.Println("Does not look like the input file contains truth tables...")
		return
	}
	if pnVars != nil {
		*pnVars = nVars
	}
	if pnTruths != nil {
		*pnTruths = int32(strings.Count(contents, "\n"))
	}
}

func Abc_TruthStoreRead(name string, p *Abc_TtStore_t) {
	if p == nil || p.nVars < 2 {
		return
	}
	data, err := os.ReadFile(name)
	if err != nil {
		fmt.Printf("Cannot open file %q for reading.\n", name)
		return
	}
	count := int32(0)
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSuffix(line, "\r")
		if line == "" {
			break
		}
		if count >= int32(len(p.pFuncs)) {
			break
		}
		Abc_TruthReadHex(&p.pFuncs[count][0], line, p.nVars)
		count++
	}
	p.nFuncs = count
}

func Abc_TtStoreWrite(name string, p *Abc_TtStore_t, binaryOutput int32) {
	if p == nil {
		return
	}
	f, err := os.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		fmt.Printf("Cannot open file %q for writing.\n", name)
		return
	}
	defer f.Close()
	for _, truth := range p.pFuncs[:p.nFuncs] {
		if binaryOutput != 0 {
			var buf [8]byte
			for _, w := range truth {
				binary.LittleEndian.PutUint64(buf[:], uint64(w))
				_, _ = f.Write(buf[:])
			}
			continue
		}
		if len(truth) != 0 {
			Abc_TruthWriteHex(f, &truth[0], p.nVars)
		}
		_, _ = fmt.Fprint(f, "    ")
		text := []byte{}
		dau.Dau_DsdDecompose(truth, p.nVars, 0, boolInt(p.nVars <= 10), &text)
		_, _ = fmt.Fprintf(f, "%s\n", text)
	}
}

func boolInt(v bool) int32 {
	if v {
		return 1
	}
	return 0
}

func Abc_TtStoreLoad(name string, nVarNum int32) *Abc_TtStore_t {
	if nVarNum < 0 {
		var nVars, nFuncs int32
		Abc_TruthGetParams(name, &nVars, &nFuncs)
		if nVars < 2 || nFuncs == 0 {
			return nil
		}
		p := Abc_TruthStoreAlloc(nVars, nFuncs)
		Abc_TruthStoreRead(name, p)
		return p
	}
	if nVarNum < 6 || nVarNum > 16 {
		return nil
	}
	data, err := os.ReadFile(name)
	if err != nil {
		fmt.Printf("Cannot open file %q for reading.\n", name)
		return nil
	}
	nBytes := 1 << uint(nVarNum-3)
	if len(data)%nBytes != 0 {
		fmt.Printf("The file size (%d) is divided by the truth table size (%d) with remainder (%d).\n", len(data), nBytes, len(data)%nBytes)
	}
	nFuncs := len(data) / nBytes
	p := Abc_TruthStoreAlloc(nVarNum, int32(nFuncs))
	for i := 0; i < nFuncs; i++ {
		for w := range p.pFuncs[i] {
			p.pFuncs[i][w] = word(binary.LittleEndian.Uint64(data[i*nBytes+w*8:]))
		}
	}
	return p
}

func Abc_TtStoreLoadSave(name string) {
	p := Abc_TtStoreLoad(name, -1)
	if p == nil {
		return
	}
	out := Extra_FileNameGenericAppend(name, "_binary.data")
	Abc_TtStoreWrite(out, p, 1)
	Abc_TtStoreFree(p, -1)
	fmt.Printf("Input file %q was copied into output file %q.\n", name, out)
}

func Abc_TtStoreTest(name string) {
	p := Abc_TtStoreLoad(name, -1)
	if p == nil {
		return
	}
	Abc_TtStoreWrite("out.txt", p, 0)
	Abc_TtStoreFree(p, -1)
}

func Abc_TtStoreLoadSaveBin(name string) {
	data, err := os.ReadFile(name)
	if err != nil {
		return
	}
	lines := strings.Split(strings.TrimSuffix(string(data), "\n"), "\n")
	var output []byte
	var nVars int = -1
	for _, line := range lines {
		if nVars < 0 {
			nVars = len(line)
		} else if len(line) != nVars {
			return
		}
		words := make([]uint32, (len(line)+31)/32)
		for i, bit := range line {
			if bit == '1' {
				words[i>>5] |= uint32(1) << uint(i&31)
			} else if bit != '0' {
				return
			}
		}
		var b [4]byte
		for _, w := range words {
			binary.LittleEndian.PutUint32(b[:], w)
			output = append(output, b[:]...)
		}
	}
	out := Extra_FileNameGenericAppend(name, "_binary.data")
	if err := os.WriteFile(out, output, 0666); err == nil {
		fmt.Printf("Input file %q was copied into output file %q.\n", name, out)
	}
}

func Abc_TruthDecPerform(p *Abc_TtStore_t, decType, verbose int32) {
	if p == nil {
		return
	}
	start := util.Abc_Clock()
	var nodes int32
	for i, truth := range p.pFuncs[:p.nFuncs] {
		if verbose != 0 {
			fmt.Printf("%7d : ", i)
		}
		switch decType {
		case 1:
			words32 := make([]uint32, len(truth)*2)
			for j, w := range truth {
				words32[2*j], words32[2*j+1] = uint32(w), uint32(w>>32)
			}
			cover, buffer := make([]uint32, 0, 1<<16), make([]byte, 0, 10000)
			sop := kit.Kit_PlaFromTruthNew(words32, p.nVars, &cover, &buffer)
			g := deco.Dec_Factor(sop)
			nodes += deco.Dec_GraphNodeNum(g)
			if verbose != 0 {
				dec.Dec_GraphPrint(os.Stdout, g, nil, "")
			}
			deco.Dec_GraphFree(g)
		case 2:
			pars := bdc.Bdc_Par_t{NVarsMax: p.nVars}
			manager := bdc.Bdc_ManAlloc(&pars)
			bdc.Bdc_ManDecompose(manager, words32FromTruth(truth), nil, p.nVars, nil, 1000)
			nodes += bdc.Bdc_ManAndNum(manager)
			if verbose != 0 {
				bdc.Bdc_ManDecPrint(manager)
			}
			bdc.Bdc_ManFree(manager)
		case 3:
			ntk := kit.Kit_DsdDecomposeMux(words32FromTruth(truth), p.nVars, 3)
			nodes += kit.Kit_DsdCountAigNodes(ntk)
			if verbose != 0 {
				kit.Kit_DsdPrintExpanded(ntk)
				fmt.Println()
			}
			kit.Kit_DsdNtkFree(ntk)
		case 4:
			text := []byte{}
			dau.Dau_DsdDecompose(truth, p.nVars, 0, 1, &text)
			nodes += dau.Dau_DsdCountAnds(string(text))
			if verbose != 0 {
				fmt.Printf("%s\n", text)
			}
		case 5:
			dau.Dau_DecTrySets(truth, util.Abc_TtSupportSize(truth, p.nVars), verbose)
		case 6:
			pool := dsc.Dsc_AllocPool(p.nVars)
			text, _ := dsc.Dsc_Decompose(truth, p.nVars, pool)
			nodes += dsc.Dsc_CountAnds(text)
			if verbose != 0 {
				fmt.Println(text)
			}
			dsc.Dsc_FreePool(pool)
		default:
			return
		}
	}
	fmt.Printf("AIG nodes =%9d  ", nodes)
	util.Abc_PrintTime(1, "Time", util.Abc_Clock()-start)
}

func words32FromTruth(truth []word) []uint32 {
	words := make([]uint32, len(truth)*2)
	for i, w := range truth {
		words[2*i], words[2*i+1] = uint32(w), uint32(w>>32)
	}
	return words
}

func Abc_TruthDecTest(name string, decType, nVarNum, verbose int32) {
	p := Abc_TtStoreLoad(name, nVarNum)
	if p == nil {
		return
	}
	Abc_TruthDecPerform(p, decType, verbose)
	Abc_TtStoreFree(p, nVarNum)
}

func Abc_DecTest(name string, decType, nVarNum, verbose int32) int32 {
	if verbose != 0 {
		fmt.Printf("Using truth tables from file %q...\n", name)
	}
	if decType == 0 {
		if nVarNum < 0 {
			Abc_TtStoreTest(name)
		}
	} else if decType >= 1 && decType <= 6 {
		Abc_TruthDecTest(name, decType, nVarNum, verbose)
	} else {
		fmt.Printf("Unknown decomposition type value (%d).\n", decType)
	}
	return 0
}
