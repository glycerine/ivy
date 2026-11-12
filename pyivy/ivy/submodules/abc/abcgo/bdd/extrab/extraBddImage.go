/**CFile****************************************************************

  FileName    [extraBddImage.c]

  PackageName [extra]

  Synopsis    [Various reusable software utilities.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - September 1, 2003.]

  Revision    [$Id: extraBddImage.c,v 1.0 2003/09/01 00:00:00 alanmi Exp $]

***********************************************************************/

package extrab

import (
	"fmt"
	"math"
	"sort"
	"strings"

	cudd "github.com/glycerine/abcgo/bdd/cudd/core"
)

type imagePart struct {
	function *DdNode
	support  map[int]bool
	nodes    int
}
type imageNode struct {
	image, cube *DdNode
	left, right *imageNode
	part        *imagePart
}
type Extra_ImageTree_t struct {
	dd                               *DdManager
	root, care                       *imageNode
	careSupport                      map[int]bool
	verbose, closed                  bool
	NIter, NNodesMax, NNodesMaxTotal int
	sizeLimit                        int
}
type Extra_ImageTree2_t struct {
	dd                    *DdManager
	relation, cube, image *DdNode
	closed                bool
}

func imageSorted(set map[int]bool) []int {
	values := make([]int, 0, len(set))
	for i := range set {
		values = append(values, i)
	}
	sort.Ints(values)
	return values
}
func imageCube(dd *DdManager, vars map[int]bool) *DdNode {
	cube := dd.One
	for _, index := range imageSorted(vars) {
		cube = cudd.Cudd_bddAnd(dd, cube, dd.Vars[index])
	}
	return cube
}
func imageReplace(dd *DdManager, destination **DdNode, f *DdNode) {
	if f != nil {
		cudd.Cudd_Ref(f)
	}
	if *destination != nil {
		cudd.Cudd_RecursiveDeref(dd, *destination)
	}
	*destination = f
}
func imageInputs(dd *DdManager, care *DdNode, nParts int, parts []*DdNode, nVars int, vars []*DdNode) (map[int]bool, bool) {
	if !extraValidBDD(dd, care) || nParts < 0 || nParts > len(parts) || nVars < 0 || nVars > len(vars) {
		return nil, false
	}
	for _, f := range parts[:nParts] {
		if !extraValidBDD(dd, f) {
			return nil, false
		}
	}
	keep := map[int]bool{}
	for _, v := range vars[:nVars] {
		if !extraValidBDD(dd, v) || cudd.Cudd_IsConstant(v) || cudd.Cudd_IsComplement(v) || int(v.Index) >= len(dd.Vars) || dd.Vars[v.Index] != v {
			return nil, false
		}
		keep[int(v.Index)] = true
	}
	return keep, true
}
func imagePartOrder(dd *DdManager, parts map[int]bool) []int {
	ids := imageSorted(parts)
	level := func(id int) int {
		if id < len(dd.Perm) {
			return dd.Perm[id]
		}
		return len(dd.Perm) + id
	}
	sort.SliceStable(ids, func(i, j int) bool { return level(ids[i]) < level(ids[j]) })
	return ids
}
func imageSameParts(a, b map[int]bool) bool {
	if len(a) != len(b) {
		return false
	}
	for id := range a {
		if !b[id] {
			return false
		}
	}
	return true
}
func imageCombine(dd *DdManager, cube *DdNode, left, right *imageNode) *imageNode {
	n := &imageNode{left: left, right: right}
	support := map[int]bool{}
	for index := range left.part.support {
		support[index] = true
	}
	for index := range right.part.support {
		support[index] = true
	}
	for _, index := range dd.Support(cube) {
		delete(support, index)
	}
	function := cudd.Cudd_bddAndAbstract(dd, left.part.function, right.part.function, cube)
	n.part = &imagePart{function, support, cudd.Cudd_DagSize(function)}
	imageReplace(dd, &n.image, function)
	if cube != dd.One {
		imageReplace(dd, &n.cube, cube)
	}
	return n
}

// Extra_bddImageStart builds the source's support-driven quantification tree.
// vars are next-state/parameter projections to retain; all others are quantified.
func Extra_bddImageStart(dd *DdManager, care *DdNode, nParts int, parts []*DdNode, nVars int, vars []*DdNode, verbose bool) *Extra_ImageTree_t {
	return Extra_bddImageStartLimited(dd, care, nParts, parts, nVars, vars, -1, verbose)
}

// Extra_bddImageStartLimited shares the source BBR scheduler's manager-node
// checks after scheduled merges. A negative limit disables these checks.
func Extra_bddImageStartLimited(dd *DdManager, care *DdNode, nParts int, parts []*DdNode, nVars int, vars []*DdNode, limit int, verbose bool) *Extra_ImageTree_t {
	keep, valid := imageInputs(dd, care, nParts, parts, nVars, vars)
	if !valid {
		return nil
	}
	nodes := make([]*imageNode, nParts+1)
	all := append(append([]*DdNode(nil), parts[:nParts]...), care)
	memberships := map[int]map[int]bool{}
	for id, f := range all {
		part := &imagePart{f, extraSupportSet(dd, f), cudd.Cudd_DagSize(f)}
		nodes[id] = &imageNode{part: part}
		for v := range part.support {
			if keep[v] {
				continue
			}
			if memberships[v] == nil {
				memberships[v] = map[int]bool{}
			}
			memberships[v][id] = true
		}
	}
	leafCubes := make([]map[int]bool, len(nodes))
	for v, members := range memberships {
		if len(members) == 1 {
			for id := range members {
				if leafCubes[id] == nil {
					leafCubes[id] = map[int]bool{}
				}
				leafCubes[id][v] = true
			}
			delete(memberships, v)
		}
	}
	for id, n := range nodes {
		f := n.part.function
		if len(leafCubes[id]) > 0 {
			cube := imageCube(dd, leafCubes[id])
			f = cudd.Cudd_bddExistAbstract(dd, f, cube)
			for v := range leafCubes[id] {
				delete(n.part.support, v)
			}
			n.part.function = f
			n.part.nodes = cudd.Cudd_DagSize(f)
			if id == nParts {
				imageReplace(dd, &n.cube, cube)
			}
		}
		imageReplace(dd, &n.image, f)
	}
	tree := &Extra_ImageTree_t{dd: dd, care: nodes[nParts], careSupport: extraSupportSet(dd, care), verbose: verbose, sizeLimit: limit}
	for len(memberships) > 0 {
		best := -1
		costBest := math.Inf(1)
		variables := make(map[int]bool, len(memberships))
		for v := range memberships {
			variables[v] = true
		}
		for _, v := range imageSorted(variables) {
			cost := 0.
			for id := range memberships[v] {
				size := float64(nodes[id].part.nodes)
				cost += size * size
			}
			if cost < costBest {
				best, costBest = v, cost
			}
		}
		members := memberships[best]
		ids := imagePartOrder(dd, members)
		if len(ids) < 2 {
			panic("extrab: inconsistent image schedule")
		}
		quantified := map[int]bool{}
		if len(ids) == 2 {
			for v, m := range memberships {
				if imageSameParts(m, members) {
					quantified[v] = true
					delete(memberships, v)
				}
			}
		} else {
			sort.SliceStable(ids, func(i, j int) bool { return nodes[ids[i]].part.nodes < nodes[ids[j]].part.nodes })
		}
		first, second := ids[0], ids[1]
		left, right := nodes[first], nodes[second]
		nodes[first] = imageCombine(dd, imageCube(dd, quantified), left, right)
		nodes[second] = nil
		for v := range right.part.support {
			if m := memberships[v]; m != nil {
				delete(m, second)
				m[first] = true
			}
		}

		if limit >= 0 && int(cudd.Cudd_ReadKeys(dd)-cudd.Cudd_ReadDead(dd)) > limit {
			for _, n := range nodes {
				if n != nil {
					imageDeleteNode(dd, n)
				}
			}
			return nil
		}
	}
	for {
		first, second := -1, -1
		for id, n := range nodes {
			if n == nil {
				continue
			}
			if first < 0 {
				first = id
			} else {
				second = id
				break
			}
		}
		if second < 0 {
			tree.root = nodes[first]
			break
		}
		nodes[first] = imageCombine(dd, dd.One, nodes[first], nodes[second])
		nodes[second] = nil
	}
	var clean func(*imageNode)
	clean = func(n *imageNode) {
		if n.left != nil {
			clean(n.left)
			clean(n.right)
		}
		n.part = nil
	}
	clean(tree.root)
	if verbose {
		fmt.Printf("Image schedule: %d partitions, %d retained variables\n", nParts, len(keep))
	}
	return tree
}
func imageCompute(tree *Extra_ImageTree_t, n *imageNode) bool {
	dd := tree.dd
	if n.left == nil {
		if n.cube != nil {
			imageReplace(dd, &n.image, cudd.Cudd_bddExistAbstract(dd, n.image, n.cube))
		}
		return true
	}
	if !imageCompute(tree, n.left) || !imageCompute(tree, n.right) {
		return false
	}
	cube := n.cube
	if cube == nil {
		cube = dd.One
	}
	imageReplace(dd, &n.image, cudd.Cudd_bddAndAbstract(dd, n.left.image, n.right.image, cube))
	if tree.verbose {
		tree.NNodesMax = max(tree.NNodesMax, cudd.Cudd_DagSize(n.image))
	}
	return tree.sizeLimit < 0 || int(cudd.Cudd_ReadKeys(dd)-cudd.Cudd_ReadDead(dd)) <= tree.sizeLimit
}
func Extra_bddImageCompute(tree *Extra_ImageTree_t, care *DdNode) *DdNode {
	if tree == nil || tree.closed || !extraValidBDD(tree.dd, care) {
		return nil
	}
	tree.NIter++
	for _, index := range tree.dd.Support(care) {
		if !tree.careSupport[index] {
			return nil
		}
	}
	imageReplace(tree.dd, &tree.care.image, care)
	tree.NNodesMax = 0
	if !imageCompute(tree, tree.root) {
		return nil
	}
	tree.NNodesMaxTotal = max(tree.NNodesMaxTotal, tree.NNodesMax)
	return tree.root.image
}
func Extra_bddImageRead(tree *Extra_ImageTree_t) *DdNode {
	if tree == nil || tree.closed {
		return nil
	}
	return tree.root.image
}
func imageDeleteNode(dd *DdManager, n *imageNode) {
	if n.left != nil {
		imageDeleteNode(dd, n.left)
		imageDeleteNode(dd, n.right)
	}
	imageReplace(dd, &n.cube, nil)
	imageReplace(dd, &n.image, nil)
}
func Extra_bddImageTreeDelete(tree *Extra_ImageTree_t) {
	if tree == nil || tree.closed {
		return
	}
	imageDeleteNode(tree.dd, tree.root)
	tree.root = nil
	tree.care = nil
	tree.careSupport = nil
	tree.closed = true
}
func Extra_bddImagePrintTree(tree *Extra_ImageTree_t) {
	if tree == nil || tree.closed {
		return
	}
	var printNode func(*imageNode, int)
	printNode = func(n *imageNode, depth int) {
		fmt.Printf("%simage=%d", strings.Repeat("    ", depth), cudd.Cudd_DagSize(n.image))
		if n.cube != nil {
			fmt.Printf(" quantify=%v", tree.dd.Support(n.cube))
		}
		fmt.Println()
		if n.left != nil {
			printNode(n.left, depth+1)
			printNode(n.right, depth+1)
		}
	}
	printNode(tree.root, 0)
}
func Extra_bddImageStart2(dd *DdManager, care *DdNode, nParts int, parts []*DdNode, nVars int, vars []*DdNode, verbose bool) *Extra_ImageTree2_t {
	keep, valid := imageInputs(dd, care, nParts, parts, nVars, vars)
	if !valid {
		return nil
	}
	tree := &Extra_ImageTree2_t{dd: dd}
	quantify := map[int]bool{}
	for i := range dd.Vars {
		if !keep[i] {
			quantify[i] = true
		}
	}
	imageReplace(dd, &tree.cube, imageCube(dd, quantify))
	relation := dd.One
	for _, f := range parts[:nParts] {
		relation = cudd.Cudd_bddAnd(dd, relation, f)
	}
	imageReplace(dd, &tree.relation, relation)
	Extra_bddImageCompute2(tree, care)
	return tree
}
func Extra_bddImageCompute2(tree *Extra_ImageTree2_t, care *DdNode) *DdNode {
	if tree == nil || tree.closed || !extraValidBDD(tree.dd, care) {
		return nil
	}
	imageReplace(tree.dd, &tree.image, cudd.Cudd_bddAndAbstract(tree.dd, tree.relation, care, tree.cube))
	return tree.image
}
func Extra_bddImageRead2(tree *Extra_ImageTree2_t) *DdNode {
	if tree == nil || tree.closed {
		return nil
	}
	return tree.image
}
func Extra_bddImageTreeDelete2(tree *Extra_ImageTree2_t) {
	if tree == nil || tree.closed {
		return
	}
	imageReplace(tree.dd, &tree.relation, nil)
	imageReplace(tree.dd, &tree.cube, nil)
	imageReplace(tree.dd, &tree.image, nil)
	tree.closed = true
}
