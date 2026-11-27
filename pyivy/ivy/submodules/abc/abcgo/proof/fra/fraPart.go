package fra

// Support-overlap diagnostics from ABC's fraPart.c.
import (
	"fmt"
	"github.com/glycerine/abcgo/misc/util"

	"github.com/glycerine/abcgo/aig/aig"
)

func fraOutputSupports(graph *aig.Graph) [][]int {
	aig.Aig_ManSetCioIds(graph)
	defer aig.Aig_ManCleanCioIds(graph)
	indices := make(map[int]int, graph.InputCount())
	for i, edge := range graph.Inputs() {
		indices[edge.NodeID()] = i
	}
	result := make([][]int, graph.OutputCount())
	for i, output := range graph.Outputs() {
		for _, edge := range aig.Aig_Support(graph, aig.Aig_ObjFanin0(graph, output)) {
			result[i] = append(result[i], indices[edge.NodeID()])
		}
	}
	return result
}

func fraSupportOverlap(a, b []int) int {
	entries := make(map[int]bool, len(a))
	for _, id := range a {
		entries[id] = true
	}
	count := 0
	for _, id := range b {
		if entries[id] {
			count++
		}
	}
	return count
}

func Fra_ManPartitionTest(graph *aig.Graph, commonLimit int32) {
	started := util.Abc_Clock()
	supports := fraOutputSupports(graph)
	util.Abc_PrintTime(1, "Supports", util.Abc_Clock()-started)
	started = util.Abc_Clock()
	total, totalNew, largest := 0, 0, 0
	ratio := 0.0
	for i, support := range supports {
		if len(support) < 2 {
			continue
		}
		union := make(map[int]bool)
		for _, id := range support {
			union[id] = true
		}
		over, quant := 0, 0
		for _, other := range supports {
			common := fraSupportOverlap(support, other)
			if common < 2 {
				continue
			}
			if common > int(commonLimit) {
				for _, id := range other {
					union[id] = true
				}
				over++
			} else {
				quant++
			}
		}
		size := len(union)
		if size > largest {
			largest = size
		}
		total += len(support)
		totalNew += size
		r := float64(size / len(support)) // C's integer division occurs before conversion.
		ratio += r
		if r >= 5 {
			fmt.Printf("%6d : S = %5d. SNew = %5d. R = %7.2f. Over = %5d. Quant = %5d. \n", i, len(support), size, r, over, quant)
		}
	}
	util.Abc_PrintTime(1, "Scanning", util.Abc_Clock()-started)
	outputs := len(supports)
	fmt.Printf("PIs = %6d. POs = %6d. Lim = %3d.   AveS = %3d. SN = %3d. R = %4.2f Max = %5d.\n", graph.InputCount(), outputs, commonLimit, total/outputs, totalNew/outputs, ratio/float64(outputs), largest)
}

func Fra_ManPartitionTest2(graph *aig.Graph) {
	started := util.Abc_Clock()
	supports := fraOutputSupports(graph)
	util.Abc_PrintTime(1, "Supports", util.Abc_Clock()-started)
	started = util.Abc_Clock()
	inverse := make([][]int, graph.InputCount())
	for i, support := range supports {
		if i >= int(graph.Assertions) {
			break
		}
		for _, id := range support {
			inverse[id] = append(inverse[id], i)
		}
	}
	util.Abc_PrintTime(1, "Inverse ", util.Abc_Clock()-started)
	started = util.Abc_Clock()
	for i, support := range supports {
		if i%50 != 0 {
			continue
		}
		extension := make(map[int]bool)
		for _, id := range support {
			extension[id] = true
			for _, output := range inverse[id] {
				for _, dependent := range supports[output] {
					extension[dependent] = true
				}
			}
		}
		fmt.Printf("%d(%d) ", len(support), len(extension))
	}
	fmt.Printf("\n")
	util.Abc_PrintTime(1, "Extension ", util.Abc_Clock()-started)
}
