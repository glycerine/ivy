package rwr

// Four- and five-input frequency exploration from rwrExp.c.
import (
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/glycerine/abcgo/misc/extra"
	"github.com/glycerine/abcgo/misc/extra/npn"
)

type Rwr_Man4_t struct {
	nFuncs                int
	puCanons              []uint16
	pnCounts              []int
	nConsidered, nClasses int
}
type Rwr_Man5_t struct{ tTableNN, tTableNPN map[uint32]int }

var s_pManRwrExp4 *Rwr_Man4_t
var s_pManRwrExp5 *Rwr_Man5_t

func Rwt_Man4ExploreStart() {
	canonical, _, _, _ := npn.Extra_Truth4VarNPN()
	s_pManRwrExp4 = &Rwr_Man4_t{nFuncs: 1 << 16, puCanons: canonical, pnCounts: make([]int, 1<<16)}
}
func Rwt_Man4ExploreCount(truth uint32) {
	assert(truth < 1<<16, "four-variable truth table")
	s_pManRwrExp4.pnCounts[s_pManRwrExp4.puCanons[truth]]++
}
func Rwt_Man4ExplorePrint() {
	p := s_pManRwrExp4
	maxCount, total, classes := 0, 0, 0
	for _, n := range p.pnCounts {
		maxCount = max(maxCount, n)
		total += n
		if n > 0 {
			classes++
		}
	}
	fmt.Printf("Number of cuts considered       = %8d.\nClasses occurring at least once = %8d.\n", total, classes)
	distribution, representative := make([]int, maxCount+1), make([]int, maxCount+1)
	for i, n := range p.pnCounts {
		distribution[n]++
		representative[n] = i
	}
	fmt.Printf("Occurence = %6d.  Num classes = %4d.  \n", 0, 2288-classes)
	for i := 1; i <= maxCount; i++ {
		if distribution[i] != 0 {
			fmt.Printf("Occurence = %6d.  Num classes = %4d.  Repr = %016b\n", i, distribution[i], representative[i])
		}
	}
	f, err := os.Create("npnclass_stats4.txt")
	if err != nil {
		panic(err)
	}
	written := 0
	for i, n := range p.pnCounts {
		if n > 0 {
			fmt.Fprintf(f, "0x%04X %10d\n", i, n)
			written++
		}
	}
	if err = f.Close(); err != nil {
		panic(err)
	}
	fmt.Printf("%d classes written into file %q.\n", written, "npnclass_stats4.txt")
}
func Rwt_Man5ExploreStart() {
	s_pManRwrExp5 = &Rwr_Man5_t{tTableNN: make(map[uint32]int), tTableNPN: make(map[uint32]int)}
}
func Rwt_Man5ExploreCount(truth uint32) { s_pManRwrExp5.tTableNN[truth]++ }
func rwrSortedTruths(table map[uint32]int) []uint32 {
	out := make([]uint32, 0, len(table))
	for truth := range table {
		out = append(out, truth)
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}
func rwrWriteTruthCounts(path string, keys []uint32, table map[uint32]int) {
	f, err := os.Create(path)
	if err != nil {
		panic(err)
	}
	for _, truth := range keys {
		fmt.Fprintf(f, "0x%08X %10d\n", truth, table[truth])
	}
	if err = f.Close(); err != nil {
		panic(err)
	}
}
func Rwt_Man5ExplorePrint() {
	p := s_pManRwrExp5
	total, maxCount := 0, 0
	for _, n := range p.tTableNN {
		total += n
		maxCount = max(maxCount, n)
	}
	fmt.Printf("Number of cuts considered       = %8d.\nClasses occurring at least once = %8d.\nThe largest number of occurence = %8d.\n", total, len(p.tTableNN), maxCount)
	distribution, representative := make([]int, maxCount+1), make([]uint32, maxCount+1)
	for truth, n := range p.tTableNN {
		distribution[n]++
		representative[n] = truth
	}
	for i := 1; i <= maxCount; i++ {
		if distribution[i] != 0 {
			fmt.Printf("Occurence = %6d.  Num classes = %4d.  Repr = %032b\n", i, distribution[i], representative[i])
		}
	}
	nn := rwrSortedTruths(p.tTableNN)
	rwrWriteTruthCounts("nnclass_stats5.txt", nn, p.tTableNN)
	fmt.Printf("%d classes written into file %q.\n", len(nn), "nnclass_stats5.txt")
	started := time.Now()
	for _, truth := range nn {
		p.tTableNPN[extra.Extra_TruthCanonNPN(truth, 5)] += p.tTableNN[truth]
	}
	fmt.Printf("The numbe of NPN classes = %d.\nComputing NPN classes: %v\n", len(p.tTableNPN), time.Since(started))
	canonical := rwrSortedTruths(p.tTableNPN)
	rwrWriteTruthCounts("npnclass_stats5.txt", canonical, p.tTableNPN)
	fmt.Printf("%d classes written into file %q.\n", len(canonical), "npnclass_stats5.txt")
}
