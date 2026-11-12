// CASCADE: Decomposition of shared BDDs into a LUT cascade.
// Port of casCore.c, Alan Mishchenko, ABC (2005).
package cas

import (
	"fmt"
	cudd "github.com/glycerine/abcgo/bdd/cudd/core"
	extra "github.com/glycerine/abcgo/bdd/extrab/cascade"
	"io"
	"math/bits"
	"os"
	"sort"
)

type DdManager = cudd.DdManager
type DdNode = cudd.DdNode

func GrayCode(code int) int { return code ^ (code >> 1) }

func GetSingleOutputFunction(dd *DdManager, outputs []*DdNode, nOuts int, encoding []*DdNode, nEncoding int, verbose bool) *DdNode {
	if dd == nil || nOuts < 0 || nOuts > len(outputs) || nEncoding < 0 || nEncoding > len(encoding) {
		return nil
	}
	order := make([]int, nOuts)
	sizes := make([]int, nOuts)
	for i := range order {
		order[i] = i
		sizes[i] = len(dd.Support(outputs[i]))
	}
	sort.SliceStable(order, func(i, j int) bool { return sizes[order[i]] > sizes[order[j]] })
	sorted := make([]*DdNode, nOuts)
	for i, j := range order {
		sorted[i] = outputs[j]
	}
	b := extra.Extra_bddEncodingBinary(dd, sorted, encoding[:nEncoding])
	a := cudd.Cudd_BddToAdd(dd, b)
	if verbose {
		fmt.Printf("Single BDD size = %6d nodes\nMTBDD           = %6d nodes\n", cudd.Cudd_DagSize(b), cudd.Cudd_DagSize(a))
	}
	return a
}
func remapOutputs(dd *DdManager, outputs []*DdNode, byLevel bool) ([]*DdNode, int) {
	result := make([]*DdNode, len(outputs))
	maxSupport := 0
	for i, f := range outputs {
		support := dd.Support(f)
		permutation := identity(len(dd.Vars))
		for k, index := range support {
			dest := k
			if byLevel {
				dest = dd.InvPerm[k]
			}
			permutation[index] = dest
		}
		result[i] = cudd.Cudd_bddPermute(dd, f, permutation)
		if len(support) > maxSupport {
			maxSupport = len(support)
		}
	}
	return result, maxSupport
}
func GetSingleOutputFunctionRemapped(dd *DdManager, outputs []*DdNode, nOuts int, encoding []*DdNode, nEncoding int) *DdNode {
	if dd == nil || nOuts < 0 || nOuts > len(outputs) || nEncoding < 0 || nEncoding > len(encoding) {
		return nil
	}
	remapped, _ := remapOutputs(dd, outputs[:nOuts], false)
	return cudd.Cudd_BddToAdd(dd, extra.Extra_bddEncodingBinary(dd, remapped, encoding[:nEncoding]))
}
func GetSingleOutputFunctionRemappedNewDD(dd *DdManager, outputs []*DdNode, nOuts int, newDD **DdManager) *DdNode {
	if dd == nil || newDD == nil || nOuts < 1 || nOuts > len(outputs) {
		return nil
	}
	remapped, maxSupport := remapOutputs(dd, outputs[:nOuts], true)
	for _, f := range remapped {
		cudd.Cudd_Ref(f)
	}
	defer func() {
		for _, f := range remapped {
			cudd.Cudd_RecursiveDeref(dd, f)
		}
	}()
	nEnc := bits.Len(uint(nOuts - 1))
	encoding := make([]*DdNode, nEnc)
	for i := range encoding {
		encoding[i] = cudd.Cudd_bddNewVarAtLevel(dd, i)
	}
	f := extra.Extra_bddEncodingBinary(dd, remapped, encoding)
	permutation := identity(len(dd.Vars))
	for level := 0; level < maxSupport+nEnc; level++ {
		permutation[dd.InvPerm[level]] = level
	}
	dest := cudd.NewBDDManager(maxSupport + nEnc)
	r := dd.Transfer(f, dest, permutation)
	if r == nil {
		return nil
	}
	*newDD = dest
	return cudd.Cudd_BddToAdd(dest, r)
}
func identity(n int) []int {
	p := make([]int, n)
	for i := range p {
		p[i] = i
	}
	return p
}
func Cudd_bddTransferPermute(source, destination *DdManager, f *DdNode, permutation []int) *DdNode {
	return source.Transfer(f, destination, permutation)
}

// WriteDDintoBLIFfile emits a shared mux network. Stable traversal IDs replace
// C pointer-derived names, preserving sharing without nondeterministic output.
func WriteDDintoBLIFfile(w io.Writer, f *DdNode, output, prefix string, names []string) {
	if f == nil {
		panic("cas: nil diagram")
	}
	ids := make(map[*DdNode]int)
	nodes := []*DdNode{}
	var visit func(*DdNode)
	visit = func(n *DdNode) {
		r := cudd.Cudd_Regular(n)
		if _, ok := ids[r]; ok {
			return
		}
		ids[r] = len(nodes)
		nodes = append(nodes, r)
		if !cudd.Cudd_IsConstant(r) {
			visit(r.E)
			visit(r.T)
		}
	}
	visit(f)
	signal := func(n *DdNode) string { return fmt.Sprintf("%sn%d", prefix, ids[cudd.Cudd_Regular(n)]) }
	fmt.Fprintf(w, ".names %s %s\n", signal(f), output)
	if cudd.Cudd_IsComplement(f) {
		fmt.Fprintln(w, "0 1")
	} else {
		fmt.Fprintln(w, "1 1")
	}
	for _, n := range nodes {
		if cudd.Cudd_IsConstant(n) {
			fmt.Fprintf(w, ".names %s\n", signal(n))
			if n.Value != 0 {
				fmt.Fprintln(w, "1")
			}
			continue
		}
		if int(n.Index) >= len(names) || names[n.Index] == "" {
			panic("cas: missing input name")
		}
		// Encode edge phases in mux cubes; no explicit inverter nodes required.
		low, high := "1", "1"
		if cudd.Cudd_IsComplement(n.E) {
			low = "0"
		}
		if cudd.Cudd_IsComplement(n.T) {
			high = "0"
		}
		fmt.Fprintf(w, ".names %s %s %s %s\n0%s- 1\n1-%s 1\n", names[n.Index], signal(n.E), signal(n.T), signal(n), low, high)
	}
}
func WriteDDintoBLIFfileReorder(dd *DdManager, w io.Writer, f *DdNode, output, prefix string, names []string) {
	dest := cudd.NewBDDManager(len(dd.Vars))
	r := dd.Transfer(f, dest, nil)
	cudd.Cudd_Ref(r)
	defer cudd.Cudd_RecursiveDeref(dest, r)
	cudd.Cudd_ReduceHeap(dest, cudd.CUDD_REORDER_SYMM_SIFT, 1)
	WriteDDintoBLIFfile(w, r, output, prefix, names)
}
func WriteSingleOutputFunctionBlif(dd *DdManager, f *DdNode, names []string, nNames int, fileName string) {
	file, err := os.Create(fileName)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	writeHeader(file, fileName, names[:nNames])
	WriteDDintoBLIFfile(file, f, "F", "", names)
	fmt.Fprintln(file, ".end")
}
func writeHeader(w io.Writer, model string, names []string) {
	fmt.Fprintf(w, ".model %s\n.inputs", model)
	for _, name := range names {
		fmt.Fprintf(w, " %s", name)
	}
	fmt.Fprintln(w, "\n.outputs F")
}
func Abc_CascadeExperiment(base string, dd *DdManager, outputs []*DdNode, nInputs, nOutputs, nLutSize int, check, verbose bool) int {
	if dd == nil || nInputs != len(dd.Vars) || nOutputs < 1 || nOutputs > len(outputs) {
		return 0
	}
	for _, f := range outputs[:nOutputs] {
		cudd.Cudd_Ref(f)
	}
	defer func() {
		for _, f := range outputs[:nOutputs] {
			cudd.Cudd_RecursiveDeref(dd, f)
		}
	}()
	nEnc := bits.Len(uint(nOutputs - 1))
	encoding := make([]*DdNode, nEnc)
	for i := range encoding {
		encoding[i] = cudd.Cudd_bddNewVarAtLevel(dd, i)
	}
	names := make([]string, nInputs+nEnc)
	for i := range names {
		if i < nInputs {
			names[i] = fmt.Sprintf("pi%03d", i)
		} else {
			names[i] = fmt.Sprintf("OutEnc_%02d", i-nInputs)
		}
	}
	f := GetSingleOutputFunction(dd, outputs, nOutputs, encoding, nEnc, verbose)
	if f == nil {
		return 0
	}
	cudd.Cudd_Ref(f)
	defer cudd.Cudd_RecursiveDeref(dd, f)
	cudd.Cudd_ReduceHeap(dd, cudd.CUDD_REORDER_SYMM_SIFT, 1)
	cudd.Cudd_ReduceHeap(dd, cudd.CUDD_REORDER_SYMM_SIFT, 1)
	if check {
		WriteSingleOutputFunctionBlif(dd, f, names, len(names), base+"_ENC.blif")
	}
	return CreateDecomposedNetwork(dd, f, names, len(names), base+"_LUT.blif", nLutSize, check, verbose)
}
