package bmc

import (
	"bufio"
	"fmt"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func Zyx_TestGetTruthTablePars(filename string, truth []uint64, nVars, nLutSize, nNodes *int32) int32 {
	pos := 0
	for pos < len(filename) && ((filename[pos] >= '0' && filename[pos] <= '9') || (filename[pos] >= 'a' && filename[pos] <= 'f') || (filename[pos] >= 'A' && filename[pos] <= 'F')) {
		pos++
	}
	switch pos {
	case 1:
		*nVars = 2
	case 2:
		*nVars = 3
	case 4:
		*nVars = 4
	case 8:
		*nVars = 5
	case 16:
		*nVars = 6
	case 32:
		*nVars = 7
	case 64:
		*nVars = 8
	default:
		fmt.Println("Invalid truth table size.")
		return 0
	}
	if len(truth) < int(util.Abc_TtWordNum(*nVars)) {
		return 0
	}
	util.Abc_TtReadHex(truth, filename[:pos])
	rest := filename[pos:]
	i := strings.IndexByte(rest, '-')
	if i < 0 {
		return 0
	}
	rest = rest[i+1:]
	j := strings.IndexByte(rest, '-')
	if j < 0 {
		return 0
	}
	lut, err := strconv.ParseInt(rest[:j], 10, 32)
	if err != nil {
		return 0
	}
	rest = rest[j+1:]
	j = 0
	for j < len(rest) && rest[j] >= '0' && rest[j] <= '9' {
		j++
	}
	nodes, err := strconv.ParseInt(rest[:j], 10, 32)
	if err != nil {
		return 0
	}
	*nLutSize, *nNodes = int32(lut), int32(nodes)
	return 1
}
func Zyx_TestTruth(info *vec.Vec_Wrd_t, i, nWords int32) []uint64 {
	data := vec.Vec_WrdArray(info)
	begin := int(i * nWords)
	return data[begin : begin+int(nWords)]
}
func Zyx_TestCreateTruthTables(nVars, nNodes int32) *vec.Vec_Wrd_t {
	words := util.Abc_TtWordNum(nVars)
	info := vec.Vec_WrdStart(int(words * (nVars + nNodes + 1)))
	for i := int32(0); i < nVars; i++ {
		util.Abc_TtIthVar(Zyx_TestTruth(info, i, words), i, nVars)
	}
	return info
}
func Zyx_TestReadNode(line string, truths *vec.Vec_Wrd_t, nVars, nLutSize, iObj int32) int32 {
	line = strings.TrimRight(line, "\r\n")
	if len(line) == 0 {
		return 0
	}
	if nLutSize < 1 || nLutSize > 6 || len(line) != int(1+nLutSize+(1<<nLutSize)) || line[0] != byte('A'+iObj) {
		return 0
	}
	words := util.Abc_TtWordNum(nVars)
	table := line[1 : 1+(1<<nLutSize)]
	names := line[1+(1<<nLutSize):]
	var fans [6][]uint64
	for i := int32(0); i < nLutSize; i++ {
		c := names[i]
		var id int32
		if c >= 'a' && c <= 'z' {
			id = int32(c - 'a')
		} else if c >= 'A' && c <= 'Z' {
			id = int32(c - 'A')
		} else {
			return 0
		}
		if id < 0 || id >= iObj {
			return 0
		}
		fans[i] = Zyx_TestTruth(truths, id, words)
	}
	dest := Zyx_TestTruth(truths, iObj, words)
	clear(dest)
	for k := int32(0); k < 1<<nLutSize; k++ {
		ch := table[(1<<nLutSize)-1-k]
		if ch == '0' {
			continue
		}
		if ch != '1' {
			return 0
		}
		for w := range dest {
			term := ^uint64(0)
			for j := int32(0); j < nLutSize; j++ {
				if k&(1<<j) != 0 {
					term &= fans[j][w]
				} else {
					term &= ^fans[j][w]
				}
			}
			dest[w] |= term
		}
	}
	return 1
}

// Zyx_TestExact reads structures in the source text format and verifies them
// against the truth table encoded in the file name.
func Zyx_TestExact(filename string) {
	f, err := os.Open(filename)
	if err != nil {
		fmt.Printf("Cannot open input file %q.\n", filename)
		return
	}
	defer f.Close()
	truth := make([]uint64, 4)
	var nVars, nLut, nNodes int32
	if Zyx_TestGetTruthTablePars(filepath.Base(filename), truth, &nVars, &nLut, &nNodes) == 0 || nVars > 8 || nLut > 6 || nNodes > 16 {
		return
	}
	info := Zyx_TestCreateTruthTables(nVars, nNodes)
	scanner := bufio.NewScanner(f)
	object, count := nVars, 0
	for scanner.Scan() {
		line := scanner.Text()
		if Zyx_TestReadNode(line, info, nVars, nLut, object) != 0 {
			object++
			continue
		}
		if object != nVars+nNodes {
			fmt.Println("The number of nodes in the structure is not correct.")
			break
		}
		count++
		got := Zyx_TestTruth(info, object-1, util.Abc_TtWordNum(nVars))
		if util.Abc_TtEqual(got, truth, util.Abc_TtWordNum(nVars)) != 0 {
			fmt.Printf("Structure %3d : Verification successful.\n", count)
		} else {
			fmt.Printf("Structure %3d : Verification FAILED.\n", count)
		}
		object = nVars
	}
	if err := scanner.Err(); err != nil {
		fmt.Printf("Cannot read input file %q: %v\n", filename, err)
	}
}
