package rwr

// Temporary five-variable statistics utility from rwrTemp.c. That C file is
// excluded from ABC's module.make but is kept here as part of the package.
import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
)

var pTruths [13719]uint32
var pFreqs [13719]int32
var pPerm [13719]int32

func Rwr_TempCompare(a, b *int32) int32 {
	if pFreqs[*a] < pFreqs[*b] {
		return 1
	}
	if pFreqs[*a] > pFreqs[*b] {
		return -1
	}
	return 0
}

func rwrTempFiles(input, output string) error {
	in, err := os.Open(input)
	if err != nil {
		return err
	}
	defer in.Close()
	reader := bufio.NewReader(in)
	for i := range pTruths {
		var text string
		if _, err = fmt.Fscan(reader, &text, &pFreqs[i]); err != nil {
			return err
		}
		if len(text) < 2 {
			return fmt.Errorf("rwr: malformed hexadecimal truth")
		}
		v, e := strconv.ParseUint(strings.TrimPrefix(text, "0x"), 16, 32)
		if e != nil {
			return e
		}
		pTruths[i] = uint32(v)
		pPerm[i] = int32(i)
	}
	sort.Slice(pPerm[:], func(i, j int) bool { return Rwr_TempCompare(&pPerm[i], &pPerm[j]) < 0 })
	file, err := os.Create(output)
	if err != nil {
		return err
	}
	w := bufio.NewWriter(file)
	fmt.Fprintln(w, "# Most frequent NPN classes of 5 vars.")
	fmt.Fprintln(w, ".model 5npn\n.inputs a b c d e")
	fmt.Fprint(w, ".outputs")
	for i := 0; i < 100; i++ {
		fmt.Fprintf(w, " %02d", i)
	}
	fmt.Fprintln(w)
	for i := 0; i < 100; i++ {
		fmt.Fprintf(w, ".names a b c d e %02d\n", i)
		truth := pTruths[pPerm[i]]
		for k := 0; k < 32; k++ {
			if truth&(uint32(1)<<uint(k)) != 0 {
				fmt.Fprintf(w, "%05b 1\n", k)
			}
		}
	}
	fmt.Fprintln(w, ".end")
	if err = w.Flush(); err != nil {
		file.Close()
		return err
	}
	return file.Close()
}
func Rwr_Temp() {
	if err := rwrTempFiles("nnclass_stats5.txt", "5npn_100.blif"); err != nil {
		panic(err)
	}
}
