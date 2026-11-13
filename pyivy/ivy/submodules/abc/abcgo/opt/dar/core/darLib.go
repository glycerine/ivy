package core
import (
	"fmt"
	"sync"
)

type libraryNode struct {
	fanins [2]uint32
	truth  uint16
}

// Library owns the complete immutable four-input DAG library. Prepared views
// select roots by the original priority arrays without mutating shared nodes.
type Library struct {
	nodes      []libraryNode
	classes    [222][]int
	priorities [222][]int
	canonical  *CanonicalTable
	prepareMu  sync.Mutex
	prepared   map[int]*PreparedLibrary
}
type preparedNode struct {
	fanins    [2]uint32
	libraryID int
}
type preparedRoot struct {
	index      int
	complement bool
	libraryID  int
}
type preparedClass struct {
	nodes []preparedNode
	roots []preparedRoot
}
type PreparedLibrary struct {
	library *Library
	classes [222]preparedClass
}

func Dar_LibAlloc(capacity int) *Library {
	if capacity < 4 {
		panic("dar: library needs four input terminals")
	}
	lib := &Library{nodes: make([]libraryNode, 4, capacity), canonical: Dar_Truth4VarNPN(), prepared: map[int]*PreparedLibrary{}}
	for i, truth := range []uint16{0xaaaa, 0xcccc, 0xf0f0, 0xff00} {
		lib.nodes[i].truth = truth
	}
	return lib
}
func Dar_LibAddNode(lib *Library, first, second int, complementFirst, complementSecond bool) {
	if first < 0 || first >= len(lib.nodes) || second < 0 || second >= len(lib.nodes) {
		panic("dar: library fanin is not an earlier object")
	}
	a, b := lib.nodes[first].truth, lib.nodes[second].truth
	litA, litB := uint32(first)<<1, uint32(second)<<1
	if complementFirst {
		a = ^a
		litA |= 1
	}
	if complementSecond {
		b = ^b
		litB |= 1
	}
	lib.nodes = append(lib.nodes, libraryNode{fanins: [2]uint32{litA, litB}, truth: a & b})
}
func Dar_LibRead() *Library {
	lib := Dar_LibAlloc(len(libraryNodeData)/2 + 4)
	for i := 0; i < len(libraryNodeData); i += 2 {
		a, b := libraryNodeData[i], libraryNodeData[i+1]
		Dar_LibAddNode(lib, int(a>>1), int(b>>1), a&1 != 0, b&1 != 0)
	}
	for _, root := range libraryOutputData {
		node := lib.nodes[root]
		match := lib.canonical.Lookup(node.truth)
		normalized := node.truth
		if normalized > ^normalized {
			normalized = ^normalized
		}
		if match.Canonical != normalized {
			panic("dar: library output is not a canonical representative")
		}
		lib.classes[match.Class] = append(lib.classes[match.Class], int(root))
	}
	offset := 0
	for class, roots := range lib.classes {
		seen := make([]bool, len(roots))
		priorities := make([]int, len(roots))
		for i := range roots {
			index := int(libraryPriorityData[offset+i])
			if index >= len(roots) || seen[index] {
				panic("dar: invalid library priority permutation")
			}
			seen[index] = true
			priorities[i] = index
		}
		lib.priorities[class] = priorities
		offset += len(roots)
	}
	if offset != len(libraryPriorityData) {
		panic("dar: library priority count")
	}
	return lib
}

var sharedLibrary struct {
	sync.Mutex
	library *Library
}

func currentLibrary() *Library {
	sharedLibrary.Lock()
	defer sharedLibrary.Unlock()
	if sharedLibrary.library == nil {
		sharedLibrary.library = Dar_LibRead()
	}
	return sharedLibrary.library
}
func Dar_LibStart() { currentLibrary() }

// Existing prepared views retain their own library reference and remain valid.
func Dar_LibStop() { sharedLibrary.Lock(); sharedLibrary.library = nil; sharedLibrary.Unlock() }
func Dar_LibReturnClass(truth uint32) int {
	return int(currentLibrary().canonical.Lookup(uint16(truth)).Class)
}
func Dar_LibPrepare(subgraphs int) *PreparedLibrary { return currentLibrary().Prepare(subgraphs) }
func (lib *Library) Prepare(subgraphs int) *PreparedLibrary {
	if subgraphs < 1 {
		panic("dar: at least one rewrite subgraph is required")
	}
	lib.prepareMu.Lock()
	defer lib.prepareMu.Unlock()
	if ready := lib.prepared[subgraphs]; ready != nil {
		return ready
	}
	prepared := &PreparedLibrary{library: lib}
	for class, roots := range lib.classes {
		count := min(len(roots), subgraphs)
		if class == 1 {
			count = len(roots)
		} // source favors the four-input AND class
		target := &prepared.classes[class]
		indices := map[int]int{0: 0, 1: 1, 2: 2, 3: 3}
		var collect func(int) int
		collect = func(id int) int {
			if slot, ok := indices[id]; ok {
				return slot
			}
			n := lib.nodes[id]
			a, b := n.fanins[0], n.fanins[1]
			first, second := collect(int(a>>1)), collect(int(b>>1))
			slot := len(target.nodes) + 4
			target.nodes = append(target.nodes, preparedNode{fanins: [2]uint32{uint32(first)<<1 | a&1, uint32(second)<<1 | b&1}, libraryID: id})
			indices[id] = slot
			return slot
		}
		for _, priority := range lib.priorities[class][:count] {
			id := roots[priority]
			slot := collect(id)
			canonical := lib.canonical.representatives[class]
			target.roots = append(target.roots, preparedRoot{index: slot, complement: lib.nodes[id].truth != canonical, libraryID: id})
		}
	}
	lib.prepared[subgraphs] = prepared
	return prepared
}
func Dar_LibPrintStats(lib *Library) string {
	roots := 0
	for _, r := range lib.classes {
		roots += len(r)
	}
	return fmt.Sprintf("DAR library: %d internal nodes, %d roots, %d NPN classes", len(lib.nodes)-4, roots, len(lib.canonical.representatives))
}

// CutNode describes an AND node in a prepared four-input implementation.
// Fanins use ABC literals referring to terminals 0..3 or prior nodes.
type CutNode struct{ Fanins [2]uint32 }

// CutRoot identifies an implementation root and its canonical output phase.
type CutRoot struct {
	Index      int
	Complement bool
}

// CutImplementations returns snapshots of a prepared class and its NPN transform.
func (p *PreparedLibrary) CutImplementations(truth uint16) (NPNMatch, []CutNode, []CutRoot) {
	if p == nil || p.library == nil {
		panic("dar: nil prepared library")
	}
	match := p.library.canonical.Lookup(truth)
	class := p.classes[match.Class]
	nodes := make([]CutNode, len(class.nodes))
	for i, node := range class.nodes {
		nodes[i] = CutNode{Fanins: node.fanins}
	}
	roots := make([]CutRoot, len(class.roots))
	for i, root := range class.roots {
		roots[i] = CutRoot{Index: root.index, Complement: root.complement}
	}
	return match, nodes, roots
}
