package mapper

// Supergate library parsing and truth verification, ported from mapperSuper.c.
import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/glycerine/abcgo/map/mio"
)

func Map_LibraryRead(lib *Map_SuperLib_t, name string) int32 {
	if lib.pGenlib != nil {
		panic("mapper: genlib already set")
	}
	file, err := os.Open(name)
	if err != nil {
		fmt.Printf("Cannot open input file %q.\n", name)
		return 0
	}
	defer file.Close()
	return Map_LibraryReadFile(lib, file)
}

func Map_LibraryReadFile(lib *Map_SuperLib_t, file *os.File) int32 {
	r := bufio.NewReader(file)
	var name string
	for {
		line, err := r.ReadString('\n')
		line = strings.TrimLeft(line, " \r\n")
		if line != "" && line[0] != '#' {
			fields := strings.Fields(line)
			if len(fields) > 0 {
				name = fields[0]
				break
			}
		}
		if err != nil {
			return 0
		}
	}
	if name == "GATE" {
		fmt.Printf("The input file %q looks like a genlib file and not a supergate library file.\n", lib.pName)
		return 0
	}
	f, err := os.Open(name)
	if err != nil {
		fmt.Printf("Cannot open the genlib file %q.\n", name)
		return 0
	}
	f.Close()
	lib.pGenlib = mio.Mio_LibraryRead(name, "", "", 0)
	if lib.pGenlib == nil {
		fmt.Printf("Cannot read genlib file %q.\n", name)
		return 0
	}
	var total int32
	if _, err = fmt.Fscan(r, &lib.nVarsMax, &total); err != nil {
		return 0
	}
	if lib.nVarsMax < 2 || lib.nVarsMax > 10 {
		fmt.Printf("Suspicious number of variables (%d).\n", lib.nVarsMax)
		return 0
	}
	if total < 1 || total > 10000000 {
		fmt.Printf("Suspicious number of gates (%d).\n", total)
		return 0
	}
	var count int32
	for {
		line, e := r.ReadString('\n')
		line = strings.TrimLeft(line, " \r\n")
		if line != "" {
			gate := Map_LibraryReadGate(lib, line, lib.nVarsMax)
			if gate.Num != count+1 {
				panic("mapper: unexpected supergate number")
			}
			gate.nGates = uint32(strings.Count(gate.pFormula, "("))
			if Map_LibraryTruthVerify(lib, gate) == 0 {
				panic("mapper: incorrect supergate truth")
			}
			var canon [2]uint32
			gate.nPhases = uint32(Map_CanonComputeSlow(lib.uTruths, lib.nVarsMax, lib.nVarsMax, gate.uTruth[:], gate.uPhases[:], canon[:]))
			Map_SuperTableInsertC(lib.tTableC, canon[:], gate)
			count++
		}
		if e != nil {
			if e != io.EOF {
				return 0
			}
			break
		}
	}
	lib.nSupersAll = count
	if count != total {
		fmt.Printf("The number of gates read (%d) is different what the file says (%d).\n", count, total)
	}
	return 1
}

func Map_LibraryReadGate(lib *Map_SuperLib_t, line string, nVars int32) *Map_Super_t {
	f := strings.Fields(line)
	if nVars < 0 || nVars > 6 || len(f) < int(nVars)+5 {
		panic("mapper: invalid supergate entry")
	}
	integer := func(s string) int32 {
		n, e := strconv.ParseInt(s, 10, 32)
		if e != nil {
			panic(e)
		}
		return int32(n)
	}
	floating := func(s string) float32 {
		n, e := strconv.ParseFloat(s, 32)
		if e != nil {
			panic(e)
		}
		return float32(n)
	}
	gate := &Map_Super_t{Num: integer(f[0])}
	if lib.nVarsMax < 6 {
		gate.uTruth[0] = Extra_ReadBinary(f[1])
	} else {
		if len(f[1]) < 64 {
			panic("mapper: short supergate truth")
		}
		gate.uTruth[0], gate.uTruth[1] = Extra_ReadBinary(f[1][32:]), Extra_ReadBinary(f[1][:32])
	}
	gate.tDelayMax.Rise = floating(f[2])
	gate.tDelayMax.Fall = gate.tDelayMax.Rise
	for i := int32(0); i < nVars; i++ {
		gate.tDelaysR[i].Rise = floating(f[3+i])
		gate.tDelaysF[i].Fall = gate.tDelaysR[i].Rise
	}
	gate.Area = floating(f[3+nVars])
	gate.pFormula = f[4+nVars]
	if len(f) > int(nVars)+5 {
		fmt.Printf("The following trailing symbols found %q.\n", strings.Join(f[5+nVars:], " "))
	}
	return gate
}

func Map_LibraryReadFormulaStep(formula string, children []string, count *int32) string {
	name := strings.TrimLeft(formula, " ")
	if name == "" {
		panic("mapper: empty formula")
	}
	open := strings.IndexByte(name, '(')
	if open < 0 {
		*count = 0
		return name
	}
	depth, end := 1, open+1
	for ; end < len(name) && depth > 0; end++ {
		if name[end] == '(' {
			depth++
		}
		if name[end] == ')' {
			depth--
		}
	}
	if depth != 0 {
		panic("mapper: unbalanced formula")
	}
	contents := name[open+1 : end-1]
	start, k := 0, 0
	depth = 0
	for i := 0; i <= len(contents); i++ {
		if i == len(contents) || (contents[i] == ',' && depth == 0) {
			if k >= len(children) {
				panic("mapper: too many formula inputs")
			}
			children[k] = contents[start:i]
			k++
			start = i + 1
			continue
		}
		if contents[i] == '(' {
			depth++
		}
		if contents[i] == ')' {
			depth--
		}
	}
	*count = int32(k)
	return name[:open]
}
func Map_LibraryTruthVerify(lib *Map_SuperLib_t, gate *Map_Super_t) int32 {
	var truth [2]uint32
	Map_LibraryComputeTruth(lib, gate.pFormula, truth[:])
	if truth == gate.uTruth {
		return 1
	}
	return 0
}
func Map_LibraryComputeTruth(lib *Map_SuperLib_t, formula string, result []uint32) {
	Map_LibraryComputeTruth_rec(lib, formula, lib.uTruths[:], result)
}
func Map_LibraryComputeTruth_rec(lib *Map_SuperLib_t, formula string, inputs [][2]uint32, result []uint32) {
	var children [6]string
	var count int32
	name := Map_LibraryReadFormulaStep(formula, children[:], &count)
	if count == 0 {
		if len(name) == 0 || name[0] < 'a' || int(name[0]-'a') >= len(inputs) || int32(name[0]-'a') >= lib.nVarsMax {
			panic("mapper: invalid formula variable")
		}
		copy(result, inputs[name[0]-'a'][:])
		return
	}
	var fanins [6][2]uint32
	for i := int32(0); i < count; i++ {
		Map_LibraryComputeTruth_rec(lib, children[i], inputs, fanins[i][:])
	}
	gate := mio.Mio_LibraryReadGateByName(lib.pGenlib, name, "")
	if gate == nil {
		panic(fmt.Sprintf("mapper: gate %q not found", name))
	}
	mio.Mio_DeriveTruthTable(gate, fanins[:], count, lib.nVarsMax, result)
}
func Map_LibraryPrintSupergate(g *Map_Super_t) {
	fmt.Printf("%5d : %5d   A = %5.2f   D = %5.2f/%5.2f/%5.2f   %s\n", g.nUsed, g.Num, g.Area, g.tDelayMax.Rise, g.tDelayMax.Fall, g.tDelayMax.Worst, g.pFormula)
}
func Map_LibraryPrintClasses(*Map_SuperLib_t) {}
