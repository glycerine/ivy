package abc

import (
	"fmt"
	"io"
	"os"
	"sort"
)

// WriteConeComparison writes support and contained fanout-cone statistics.
func WriteConeComparison(w io.Writer, n *Network) error {
	type candidate struct {
		index   int
		support []*Object
	}
	items := make([]candidate, len(n.cos))
	for i, output := range n.cos {
		items[i] = candidate{i, Abc_NtkNodeSupport(n, []*Object{output}, 1)}
	}
	sort.SliceStable(items, func(i, j int) bool { return len(items[i].support) > len(items[j].support) })
	seen := make(map[*Object]bool)
	// Reverse DFS uses Copy as a linked list; preserve the caller's mappings.
	copies := Abc_NtkSaveCopy(n)
	defer Abc_NtkLoadCopy(n, copies)
	iteration := 0
	for _, item := range items {
		output := n.cos[item.index]
		if seen[output] {
			continue
		}
		iteration++
		nodes := Abc_NtkDfsNodes(n, []*Object{output}, 1)
		levels := Abc_NtkDfsReverseNodesContained(n, item.support, int32(len(item.support)))
		count := 0
		for _, head := range levels[1 : len(levels)-1] {
			for node := head; node != nil; node = node.Copy {
				count++
			}
		}
		outputs, newOutputs := 0, 0
		for node := levels[len(levels)-1]; node != nil; node = node.Copy {
			outputs++
			if !seen[node] {
				newOutputs++
			}
			seen[node] = true
		}
		if len(output.fanins) == 0 {
			return fmt.Errorf("abc: output has no driver")
		}
		if _, err := fmt.Fprintf(w, "%4d CO %5d :  Supp = %5d.  Lev = %3d.  Cone = %5d.  Rev = %5d.  COs = %3d (%3d).\n", iteration, item.index, len(item.support), output.fanins[0].Object.Level, len(nodes), count, outputs, newOutputs); err != nil {
			return err
		}
		if len(item.support) < 10 {
			break
		}
	}
	return nil
}

func Abc_NtkCompareCones(n *Network) {
	if err := WriteConeComparison(os.Stdout, n); err != nil {
		fmt.Println(err)
	}
}

func Abc_NtkCompareSupports(n *Network) {
	if Abc_NtkIsStrash(n) == 0 {
		panic("abc: support comparison requires an AIG")
	}
	for _, node := range n.ObjectsOfKind(ObjectNode) {
		if Abc_AigNodeIsChoice(node) == 0 {
			continue
		}
		original := len(Abc_NtkNodeSupport(n, []*Object{node}, 1))
		for choice := node.Choice; choice != nil; choice = choice.Choice {
			current := len(Abc_NtkNodeSupport(n, []*Object{choice}, 1))
			if current != original {
				fmt.Printf("Choice orig = %3d  Choice new = %3d\n", original, current)
			}
		}
	}
}

func Abc_NtkInvertConstraints(n *Network) {
	outputs := n.ObjectsOfKind(ObjectPO)
	if n.Constraints < 0 || int(n.Constraints) > len(outputs) {
		panic("abc: invalid constraint count")
	}
	for _, output := range outputs[len(outputs)-int(n.Constraints):] {
		Abc_ObjXorFaninC(output, 0)
	}
}

func Abc_NtkPrintCiLevels(n *Network) {
	for i, input := range n.cis {
		fmt.Printf("%c=%d ", 'a'+i, input.Level)
	}
	fmt.Println()
}

func Abc_NtkAddBuffsEval(node, fanin *Object) int32 {
	for _, fanout := range fanin.fanouts {
		if fanout != node && fanout.Level >= node.Level {
			return 0
		}
	}
	return 1
}
func Abc_NtkAddBuffsEval2(node, fanin *Object) int32 {
	for _, fanout := range fanin.fanouts {
		if fanout != node && fanout.Level > node.Level {
			return 1
		}
	}
	return 0
}

func Abc_NtkPrint256() {
	file, err := os.Create("4varfs.txt")
	if err != nil {
		fmt.Println(err)
		return
	}
	for i := 1; i < (1<<16)-1; i++ {
		if _, err = fmt.Fprintf(file, "read_truth %016b; clp; st; w 1.blif; map; cec 1.blif\n", i); err != nil {
			break
		}
	}
	if closeErr := file.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		fmt.Println(err)
	}
}
