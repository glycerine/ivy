package extra

// Majority-function enumeration by majority-gate expansion and variable
// identification. Canonicalization changes only variable order, preserving the
// positive-unate function family and original insertion order.
import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"slices"
	"time"

	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/opt/dau/canon"
)

type Gem_Obj_t struct {
	nVars, nNodes, History, Groups, Predec int32
}
type Gem_Man_t struct {
	nVars, nWords, nObjsAlloc, nObjs int32
	pObjs                            []Gem_Obj_t
	truths                           [][]uint64
	index                            map[string]int32
	pTtElems                         [][]uint64
	fVerbose                         int32
}

func gemTruthKey(truth []uint64) string {
	key := make([]byte, 8*len(truth))
	for i, w := range truth {
		binary.LittleEndian.PutUint64(key[8*i:], w)
	}
	return string(key)
}
func Gem_PrintNode(p *Gem_Man_t, f int32, label string, update int32) {
	if p.fVerbose == 0 {
		return
	}
	obj := p.pObjs[f]
	fmt.Printf("Node %6d : %s  Pred = %6d  Vars = %d  Nodes = %d  History = %d%d  Profile: ", f, label, obj.Predec, obj.nVars, obj.nNodes, obj.History&15, obj.History>>4)
	for i := p.nVars - 1; i >= 0; i-- {
		fmt.Printf("%d", obj.Groups>>i&1)
	}
	if update != 0 {
		fmt.Print(" *")
	}
	fmt.Println()
}
func Gem_ManAlloc(nVars, verbose int32) *Gem_Man_t {
	if nVars < 1 || nVars > 16 {
		panic("majority enumeration requires 1..16 variables")
	}
	words := util.Abc_TtWordNum(nVars)
	p := &Gem_Man_t{nVars: nVars, nWords: words, nObjs: 2, nObjsAlloc: 16, pObjs: make([]Gem_Obj_t, 2, 16), index: map[string]int32{}, fVerbose: verbose}
	p.pObjs[1] = Gem_Obj_t{nVars: 1, Groups: 1}
	p.pTtElems = make([][]uint64, nVars+4)
	for v := range p.pTtElems {
		p.pTtElems[v] = make([]uint64, words)
	}
	for v := int32(0); v < nVars; v++ {
		for bit := int32(0); bit < 64*words; bit++ {
			if bit>>v&1 != 0 {
				p.pTtElems[v][bit>>6] |= uint64(1) << uint(bit&63)
			}
		}
	}
	p.truths = [][]uint64{make([]uint64, words), slices.Clone(p.pTtElems[0])}
	for i, t := range p.truths {
		p.index[gemTruthKey(t)] = int32(i)
	}
	Gem_PrintNode(p, 1, "Original", 0)
	return p
}
func Gem_ManFree(p *Gem_Man_t) int32 {
	if p != nil {
		*p = Gem_Man_t{}
	}
	return 1
}
func Gem_ManRealloc(p *Gem_Man_t) {
	if p.nObjs != p.nObjsAlloc {
		panic("majority storage is not full")
	}
	if p.nObjs == 1<<31-1 {
		panic("majority node limit reached")
	}
	size := min(int64(p.nObjsAlloc)*2, int64(1<<31-1))
	objects := make([]Gem_Obj_t, len(p.pObjs), int(size))
	copy(objects, p.pObjs)
	p.pObjs = objects
	p.nObjsAlloc = int32(size)
}
func Gem_GroupsDerive(truth []uint64, nVars int32, cof0, cof1 []uint64) int32 {
	if nVars < 1 || nVars > 16 {
		panic("invalid majority group width")
	}
	result := int32(1) << (nVars - 1)
	for v := int32(0); v < nVars-1; v++ {
		if util.Abc_TtVarsAreSymmetric(truth, nVars, v, v+1, cof0, cof1) == 0 {
			result |= 1 << v
		}
	}
	return result
}
func Gem_GroupVarRemove(groups, i int32) int32 {
	if i < 0 || i > 15 || groups>>i&1 == 0 {
		panic("invalid removed group boundary")
	}
	mask := int32(1)<<i - 1
	return groups&mask | (groups&^mask)>>1
}
func Gem_GroupVarsInsert1(groups, i, group int32) int32 {
	if i < -1 || i > 15 || group < 0 || group > 1 || group != 0 && i != -1 && groups>>i&1 == 0 {
		panic("invalid inserted group boundary")
	}
	mask := int32(1)<<(i+1) - 1
	return groups&mask | (groups&^mask)<<1 | group<<(i+1)
}
func Gem_GroupVarsInsert3(groups, i int32) int32 {
	if i < -1 || i > 15 || i != -1 && groups>>i&1 == 0 {
		panic("invalid inserted symmetric group")
	}
	mask := int32(1)<<(i+1) - 1
	return groups&mask | (groups&^mask)<<3 | 4<<(i+1)
}
func Gem_GroupUnpack(groups int32, vars []int32) int32 {
	if groups < 0 || groups > 65535 {
		panic("invalid group mask")
	}
	n := int32(0)
	for v := int32(0); groups != 0; v, groups = v+1, groups>>1 {
		if groups&1 != 0 {
			vars[n] = v
			n++
		}
	}
	return n
}

// The C algorithm compares raw bytes. Explicit little-endian encoding preserves
// that ordering independently of the host architecture.
func Gem_FuncFindPlace(truth []uint64, words, groups int32, best []uint64, oneVar int32) int32 {
	truth = truth[:words]
	best = best[:words]
	var last [16]int32
	n := Gem_GroupUnpack(groups, last[:])
	place := int32(-1)
	if n > 0 {
		place = last[n-1]
	}
	copy(best, truth)
	for g := n - 1; g >= 0; g-- {
		limit := int32(-1)
		if g > 0 {
			limit = last[g-1]
		}
		for v := last[g]; v > limit; v-- {
			util.Abc_TtSwapAdjacent(truth, words, v)
			if oneVar == 0 {
				util.Abc_TtSwapAdjacent(truth, words, v+1)
				util.Abc_TtSwapAdjacent(truth, words, v+2)
			}
		}
		if bytes.Compare([]byte(gemTruthKey(best)), []byte(gemTruthKey(truth))) < 0 {
			copy(best, truth)
			place = limit
		}
	}
	return place
}
func gemInsert(p *Gem_Man_t, truth []uint64, obj Gem_Obj_t, label string) bool {
	key := gemTruthKey(truth)
	if _, ok := p.index[key]; ok {
		return false
	}
	if p.nObjs == p.nObjsAlloc {
		Gem_ManRealloc(p)
	}
	index := p.nObjs
	p.index[key] = index
	p.truths = append(p.truths, slices.Clone(truth))
	p.pObjs = append(p.pObjs, obj)
	p.nObjs++
	Gem_PrintNode(p, index, label, 0)
	return true
}
func Gem_FuncExpand(p *Gem_Man_t, f, i int32) {
	obj := p.pObjs[f]
	if i < 0 || i >= obj.nVars || obj.nVars+2 > p.nVars {
		panic("invalid majority expansion")
	}
	result := slices.Clone(p.truths[f])
	cof0, cof1 := make([]uint64, p.nWords), make([]uint64, p.nWords)
	for v := i; v < obj.nVars-1; v++ {
		util.Abc_TtSwapAdjacent(result, p.nWords, v)
	}
	v := obj.nVars - 1
	util.Abc_TtCofactor0p(cof0, result, p.nWords, v)
	util.Abc_TtCofactor1p(cof1, result, p.nWords, v)
	util.Abc_TtMaj(result, p.pTtElems[v], p.pTtElems[v+1], p.pTtElems[v+2], p.nWords)
	util.Abc_TtMux(result, result, cof1, cof0, p.nWords)
	var perm [16]byte
	canon.CanonicalizePermutation(result, obj.nVars+2, perm[:])
	util.Abc_TtStretch6(result, max(6, obj.nVars+2), max(6, p.nVars))
	next := Gem_Obj_t{nVars: obj.nVars + 2, nNodes: obj.nNodes + 1, Groups: Gem_GroupsDerive(result, obj.nVars+2, cof0, cof1), Predec: f, History: i}
	gemInsert(p, result, next, "Expand  ")
}
func Gem_FuncCheckMajority(p *Gem_Man_t, f int32) int32 {
	obj := p.pObjs[f]
	polar := util.Abc_TtIsFullySymmetric(p.truths[f], obj.nVars)
	if polar != -1 {
		fmt.Printf("Found symmetric %d-variable function: ", obj.nVars)
		for bit := obj.nVars; bit >= 0; bit-- {
			fmt.Printf("%d", polar>>bit&1)
		}
		fmt.Print("   ")
		half := (obj.nVars + 1) >> 1
		mask := int32(1)<<half - 1
		if obj.nVars&1 != 0 && polar == mask<<half {
			fmt.Printf("This is majority-%d.\n", obj.nVars)
		} else {
			fmt.Println()
		}
	}
	return 0
}
func Gem_FuncReduce(p *Gem_Man_t, f, i, j int32) int32 {
	obj := p.pObjs[f]
	if i < 0 || i >= j || j >= obj.nVars {
		panic("invalid majority variable identification")
	}
	result := slices.Clone(p.truths[f])
	cof0, cof1 := make([]uint64, p.nWords), make([]uint64, p.nWords)
	for v := j; v < obj.nVars-1; v++ {
		util.Abc_TtSwapAdjacent(result, p.nWords, v)
	}
	for v := i; v < obj.nVars-2; v++ {
		util.Abc_TtSwapAdjacent(result, p.nWords, v)
	}
	v := obj.nVars - 2
	util.Abc_TtCofactor0p(cof0, result, p.nWords, v+1)
	util.Abc_TtCofactor1p(cof1, result, p.nWords, v+1)
	util.Abc_TtCofactor0(cof0, p.nWords, v)
	util.Abc_TtCofactor1(cof1, p.nWords, v)
	util.Abc_TtMux(result, p.pTtElems[v], cof1, cof0, p.nWords)
	var perm [16]byte
	canon.CanonicalizePermutation(result, obj.nVars-1, perm[:])
	util.Abc_TtStretch6(result, max(6, obj.nVars-1), max(6, p.nVars))
	next := Gem_Obj_t{nVars: obj.nVars - 1, nNodes: obj.nNodes, Groups: Gem_GroupsDerive(result, obj.nVars-1, cof0, cof1), Predec: f, History: j<<4 | i}
	if gemInsert(p, result, next, "Crossbar") {
		Gem_FuncCheckMajority(p, p.nObjs-1)
	}
	return 0
}
func gemEnumerate(p *Gem_Man_t, progress bool) {
	started := time.Now()
	stop := int32(1)
	for v := int32(1); v <= p.nVars-2; v++ {
		previous := stop
		stop = p.nObjs
		if progress {
			fmt.Printf("Expanding  var %2d (functions = %10d)  Time = %v\n", v, p.nObjs, time.Since(started))
		}
		for f := int32(0); f < stop; f++ {
			obj := p.pObjs[f]
			if v == obj.nVars || v > obj.nVars && f >= previous {
				for i := int32(0); i < v; i++ {
					if obj.Groups>>i&1 != 0 {
						Gem_FuncExpand(p, f, i)
					}
				}
			}
		}
		if progress {
			fmt.Printf("Connecting var %2d (functions = %10d)  Time = %v\n", v, p.nObjs, time.Since(started))
		}
		for f := stop; f < p.nObjs; f++ {
			obj := p.pObjs[f]
			for i := int32(0); i < obj.nVars; i++ {
				if obj.Groups>>i&1 == 0 {
					continue
				}
				for j := i + 1; j < obj.nVars; j++ {
					if obj.Groups>>j&1 != 0 {
						Gem_FuncReduce(p, f, i, j)
					}
				}
			}
		}
	}
	if progress {
		fmt.Printf("Finished          (functions = %10d)  Time = %v\n", p.nObjs, time.Since(started))
	}
}
func Gem_Enumerate(nVars, dump, verbose int32) int32 {
	p := Gem_ManAlloc(nVars, verbose)
	defer Gem_ManFree(p)
	gemEnumerate(p, true)
	if dump != 0 {
		file, err := os.Create(fmt.Sprintf("tt_enum_%02d.txt", nVars))
		if err != nil {
			panic(err)
		}
		for _, truth := range p.truths {
			for w := len(truth) - 1; w >= 0; w-- {
				if _, err = fmt.Fprintf(file, "%016x", truth[w]); err != nil {
					file.Close()
					panic(err)
				}
			}
			fmt.Fprintln(file)
		}
		if err = file.Close(); err != nil {
			panic(err)
		}
	}
	return 0
}
