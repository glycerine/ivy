package cudd

import (
	"fmt"
	core "github.com/glycerine/abcgo/bdd/cudd/core"
	"io"
	"unsafe"
)

func Cudd_zddPrintMinterm(dd *DdManager, root *DdNode) int32 {
	if dd == nil || root == nil {
		return 0
	}
	out := core.Cudd_zddOutput(dd)
	if out == nil {
		return 0
	}
	path := make([]int, dd.ZddSize)
	for i := range path {
		path[i] = 3
	}
	zddPrintMintermAux(dd, root, 0, path, out)
	return 1
}
func Cudd_zddPrintCover(dd *DdManager, root *DdNode) int32 {
	if dd == nil || root == nil || dd.ZddSize&1 != 0 {
		return 0
	}
	out := core.Cudd_zddOutput(dd)
	if out == nil {
		return 0
	}
	path := make([]int, dd.ZddSize)
	for i := range path {
		path[i] = 3
	}
	zddPrintCoverAux(dd, root, 0, path, out)
	return 1
}
func Cudd_zddPrintDebug(dd *DdManager, f *DdNode, n, pr int32) int32 {
	if dd == nil || f == nil {
		return 0
	}
	out := core.Cudd_zddOutput(dd)
	if out == nil {
		return 0
	}
	if f == dd.Zero && pr > 0 {
		fmt.Fprintln(out, ": is the empty ZDD")
		return 1
	}
	ret := int32(1)
	if pr > 0 {
		nodes := core.Cudd_zddDagSize(f)
		mt := core.Cudd_zddCountMinterm(dd, f, n)
		if nodes < 0 || mt < 0 {
			ret = 0
		}
		fmt.Fprintf(out, ": %d nodes %g minterms\n", nodes, mt)
		if pr > 2 && cuddZddP(dd, f) == 0 {
			ret = 0
		}
		if pr == 2 || pr > 3 {
			if Cudd_zddPrintMinterm(dd, f) == 0 {
				ret = 0
			}
			fmt.Fprintln(out)
		}
	}
	return ret
}
func Cudd_zddCoverPathToString(dd *DdManager, path []int32, str []byte) []byte {
	p := make([]int, len(path))
	for i, v := range path {
		p[i] = int(v)
	}
	return core.Cudd_zddCoverPathToString(dd, p, str)
}

func Cudd_zddFirstPath(dd *DdManager, f *DdNode, path **int32) *DdGen {
	if dd == nil || f == nil || path == nil {
		return nil
	}
	n := dd.ZddSize
	g := &DdGen{manager: dd, genType: CUDD_GEN_ZDD_PATHS, status: CUDD_GEN_EMPTY}
	g.data.cubes.cube = make([]int32, n)
	for i := range g.data.cubes.cube {
		g.data.cubes.cube[i] = 2
	}
	g.stack.stack = make([]*DdNode, n+1)
	g.stack.stack[0] = f
	g.stack.sp = 1
	if !zddPathSeek(g) {
		g.stack.sp = 0
	} else {
		g.status = CUDD_GEN_NONEMPTY
	}
	if n == 0 {
		*path = nil
	} else {
		*path = &g.data.cubes.cube[0]
	}
	return g
}
func Cudd_zddNextPath(g *DdGen, path **int32) int32 {
	if g == nil || path == nil || g.status == CUDD_GEN_EMPTY {
		return 0
	}
	if !zddPathBacktrack(g) || !zddPathSeek(g) {
		g.status = CUDD_GEN_EMPTY
		return 0
	}
	g.status = CUDD_GEN_NONEMPTY
	if len(g.data.cubes.cube) == 0 {
		*path = nil
	} else {
		*path = &g.data.cubes.cube[0]
	}
	return 1
}
func zddPathBacktrack(g *DdGen) bool {
	for {
		if g.stack.sp <= 1 {
			return false
		}
		top := g.stack.stack[g.stack.sp-1]
		prev := Cudd_Regular(g.stack.stack[g.stack.sp-2])
		next := Cudd_T(prev)
		if next != top {
			g.data.cubes.cube[prev.Index] = 1
			g.stack.stack[g.stack.sp-1] = next
			return true
		}
		g.data.cubes.cube[prev.Index] = 2
		g.stack.sp--
	}
}
func zddPathSeek(g *DdGen) bool {
	for {
		top := g.stack.stack[g.stack.sp-1]
		regular := Cudd_Regular(top)
		if !Cudd_IsConstant(regular) {
			g.data.cubes.cube[regular.Index] = 0
			g.stack.stack[g.stack.sp] = Cudd_Not(Cudd_E(regular))
			g.stack.sp++
			continue
		}
		if regular == g.manager.Zero {
			if !zddPathBacktrack(g) {
				return false
			}
			continue
		}
		g.data.cubes.value = CUDD_VALUE_TYPE(regular.Value)
		return true
	}
}
func Cudd_zddDumpDot(dd *DdManager, n int32, roots []*DdNode, inames, onames []string, w io.Writer) int32 {
	if n <= 0 || int(n) > len(roots) {
		return 0
	}
	return int32(core.Cudd_zddDumpDot(dd, roots[:n], inames, onames, w))
}

func zddPrintMintermAux(dd *DdManager, node *DdNode, level int, path []int, out io.Writer) {
	if Cudd_IsConstant(node) {
		if node == dd.One {
			if level != dd.ZddSize {
				if level < dd.ZddSize {
					path[dd.InvPermZ[level]] = 0
					zddPrintMintermAux(dd, node, level+1, path, out)
				}
				return
			}
			for _, v := range path {
				switch v {
				case 0:
					fmt.Fprint(out, "0")
				case 1:
					fmt.Fprint(out, "1")
				case 3:
					fmt.Fprint(out, "@")
				default:
					fmt.Fprint(out, "-")
				}
			}
			fmt.Fprintln(out, " 1")
		}
		return
	}
	idx := int(Cudd_Regular(node).Index)
	if level != dd.PermZ[idx] {
		if level < dd.ZddSize {
			path[dd.InvPermZ[level]] = 0
			zddPrintMintermAux(dd, node, level+1, path, out)
		}
		return
	}
	e, t := Cudd_E(node), Cudd_T(node)
	if t == e {
		path[idx] = 2
		zddPrintMintermAux(dd, e, level+1, path, out)
		return
	}
	path[idx] = 1
	zddPrintMintermAux(dd, t, level+1, path, out)
	path[idx] = 0
	zddPrintMintermAux(dd, e, level+1, path, out)
}
func zddPrintCoverAux(dd *DdManager, node *DdNode, level int, path []int, out io.Writer) {
	if Cudd_IsConstant(node) {
		if node == dd.One {
			if level != dd.ZddSize {
				if level < dd.ZddSize {
					path[dd.InvPermZ[level]] = 0
					zddPrintCoverAux(dd, node, level+1, path, out)
				}
				return
			}
			for i := 0; i < dd.ZddSize; i += 2 {
				switch path[i]*4 + path[i+1] {
				case 0:
					fmt.Fprint(out, "-")
				case 4:
					fmt.Fprint(out, "1")
				case 1:
					fmt.Fprint(out, "0")
				default:
					fmt.Fprint(out, "@")
				}
			}
			fmt.Fprintln(out, " 1")
		}
		return
	}
	idx := int(Cudd_Regular(node).Index)
	if level != dd.PermZ[idx] {
		if level < dd.ZddSize {
			path[dd.InvPermZ[level]] = 0
			zddPrintCoverAux(dd, node, level+1, path, out)
		}
		return
	}
	e, t := Cudd_E(node), Cudd_T(node)
	if t == e {
		path[idx] = 2
		zddPrintCoverAux(dd, e, level+1, path, out)
		return
	}
	path[idx] = 1
	zddPrintCoverAux(dd, t, level+1, path, out)
	path[idx] = 0
	zddPrintCoverAux(dd, e, level+1, path, out)
}
func cuddZddP(dd *DdManager, f *DdNode) int32 {
	out := core.Cudd_zddOutput(dd)
	if out == nil {
		return 0
	}
	seen := map[*DdNode]bool{}
	var walk func(*DdNode) bool
	walk = func(edge *DdNode) bool {
		if edge == nil {
			return false
		}
		n := Cudd_Regular(edge)
		if Cudd_IsConstant(n) {
			fmt.Fprintf(out, "ID = %d\n", boolInt(edge == dd.One))
			return true
		}
		if seen[n] {
			return true
		}
		seen[n] = true
		id := uintptr(unsafe.Pointer(n)) / core.Cudd_zddNodeSize()
		fmt.Fprintf(out, "ID = 0x%x\tindex = %d\tr = %d\t", id, n.Index, n.Ref)
		printZddChild(out, "T", Cudd_T(n), dd.One)
		printZddChild(out, "E", Cudd_E(n), dd.One)
		return (Cudd_IsConstant(Cudd_E(n)) || walk(Cudd_E(n))) && (Cudd_IsConstant(Cudd_T(n)) || walk(Cudd_T(n)))
	}
	ok := walk(f)
	fmt.Fprintln(out)
	if ok {
		return 1
	}
	return 0
}

func printZddChild(out io.Writer, label string, n, one *DdNode) {
	if Cudd_IsConstant(n) {
		fmt.Fprintf(out, "%s = %d\t", label, boolInt(n == one))
	} else {
		fmt.Fprintf(out, "%s = 0x%x\t", label, uintptr(unsafe.Pointer(Cudd_Regular(n)))/core.Cudd_zddNodeSize())
	}
}

func boolInt(v bool) int {
	if v {
		return 1
	}
	return 0
}
