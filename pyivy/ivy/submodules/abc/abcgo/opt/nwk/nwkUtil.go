package nwk

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/hop"
	"github.com/glycerine/abcgo/misc/vec"
	"io"
	"math"
	"os"
	"sort"
)

func Nwk_ManIncrementTravId(p *Nwk_Man_t) {
	if p.TravID >= math.MaxInt32-3 || p.TravID < 0 {
		for _, o := range p.Objects {
			if o != nil {
				o.TravID = 0
			}
		}
		p.TravID = 1
	} else {
		p.TravID++
	}
}
func Nwk_ManGetFaninMax(p *Nwk_Man_t) int32 {
	var n int
	for _, o := range p.Objects {
		if o != nil && o.Type == NWK_OBJ_NODE {
			n = max(n, len(o.Fanins))
		}
	}
	return int32(n)
}
func Nwk_ManGetTotalFanins(p *Nwk_Man_t) int32 {
	var n int
	for _, o := range p.Objects {
		if o != nil && o.Type == NWK_OBJ_NODE {
			n += len(o.Fanins)
		}
	}
	return int32(n)
}
func Nwk_ManPiNum(p *Nwk_Man_t) int32 {
	var n int32
	for _, o := range p.Inputs {
		n += Nwk_ObjIsPi(o)
	}
	return n
}
func Nwk_ManPoNum(p *Nwk_Man_t) int32 {
	var n int32
	for _, o := range p.Outputs {
		n += Nwk_ObjIsPo(o)
	}
	return n
}
func Nwk_ManGetAigNodeNum(p *Nwk_Man_t) int32 {
	var n int32
	for _, o := range p.Objects {
		if o != nil && o.Type == NWK_OBJ_NODE && o.Function != nil {
			n += hop.Hop_DagSize(o.Function)
		}
	}
	return n
}
func Nwk_NodeCompareLevelsIncrease(a, b **Nwk_Obj_t) int32 {
	if (*a).Level < (*b).Level {
		return -1
	}
	if (*a).Level > (*b).Level {
		return 1
	}
	return 0
}
func Nwk_NodeCompareLevelsDecrease(a, b **Nwk_Obj_t) int32 {
	return -Nwk_NodeCompareLevelsIncrease(a, b)
}
func Nwk_ObjPrint(o *Nwk_Obj_t) {
	fmt.Fprintf(writer(o.Manager), "Object %d (type %d): fanins", o.ID, o.Type)
	for _, n := range o.Fanins {
		fmt.Fprintf(writer(o.Manager), " %d", n.ID)
	}
	fmt.Fprint(writer(o.Manager), "; fanouts")
	for _, n := range o.Fanouts {
		fmt.Fprintf(writer(o.Manager), " %d", n.ID)
	}
	fmt.Fprintln(writer(o.Manager))
}
func truthWords(o *Nwk_Obj_t, scratch *Vec_Int_t) []uint32 {
	return hop.Hop_ManConvertAigToTruth(o.Manager.Hop, o.Function, int32(len(o.Fanins)), scratch, 0)
}
func truthBit(truth []uint32, index int) bool { return truth[index/32]>>uint(index%32)&1 != 0 }
func (p *Nwk_Man_t) WriteBLIF(w io.Writer, ciNames, coNames []string) error {
	if len(ciNames) != 0 && len(ciNames) != len(p.Inputs) || len(coNames) != 0 && len(coNames) != len(p.Outputs) {
		return fmt.Errorf("nwk: name count mismatch")
	}
	name := p.Name
	if name == "" {
		name = "network"
	}
	names := map[*Nwk_Obj_t]string{}
	for _, o := range p.Objects {
		if o != nil {
			names[o] = fmt.Sprintf("n%d", o.ID)
		}
	}
	for i, o := range p.Inputs {
		if len(ciNames) > 0 {
			names[o] = ciNames[i]
		}
	}
	for i, o := range p.Outputs {
		if len(coNames) > 0 {
			names[o] = coNames[i]
		} else {
			names[o] = fmt.Sprintf("o%d", i)
		}
	}
	if _, e := fmt.Fprintf(w, ".model %s\n.inputs", name); e != nil {
		return e
	}
	for _, o := range p.Inputs {
		fmt.Fprintf(w, " %s", names[o])
	}
	fmt.Fprint(w, "\n.outputs")
	for _, o := range p.Outputs {
		fmt.Fprintf(w, " %s", names[o])
	}
	fmt.Fprintln(w)
	scratch := vec.Vec_IntAlloc(0)
	for _, o := range vectorNodes(Nwk_ManDfs(p)) {
		if o.Type == NWK_OBJ_CI {
			continue
		}
		fmt.Fprint(w, ".names")
		for _, f := range o.Fanins {
			fmt.Fprintf(w, " %s", names[f])
		}
		fmt.Fprintf(w, " %s\n", names[o])
		if o.Type == NWK_OBJ_CO {
			if o.Invert {
				fmt.Fprintln(w, "0 1")
			} else {
				fmt.Fprintln(w, "1 1")
			}
			continue
		}
		if len(o.Fanins) > 20 {
			return fmt.Errorf("nwk: BLIF truth expansion exceeds 20 inputs")
		}
		truth := truthWords(o, scratch)
		for m := 0; m < 1<<len(o.Fanins); m++ {
			if !truthBit(truth, m) {
				continue
			}
			for i := range o.Fanins {
				if m>>i&1 != 0 {
					fmt.Fprint(w, "1")
				} else {
					fmt.Fprint(w, "0")
				}
			}
			if len(o.Fanins) > 0 {
				fmt.Fprint(w, " ")
			}
			fmt.Fprintln(w, "1")
		}
	}
	_, e := fmt.Fprintln(w, ".end")
	return e
}
func Nwk_ManDumpBlif(p *Nwk_Man_t, path string, inputs, outputs *Vec_Ptr_t) {
	names := func(v *Vec_Ptr_t) []string {
		if v == nil {
			return nil
		}
		r := make([]string, vec.Vec_PtrSize(v))
		for i := range r {
			r[i] = vec.Vec_PtrEntry(v, i).(string)
		}
		return r
	}
	f, e := os.Create(path)
	if e != nil {
		panic(e)
	}
	defer f.Close()
	if e = p.WriteBLIF(f, names(inputs), names(outputs)); e != nil {
		panic(e)
	}
}
func Nwk_ManPrintFanioNew(p *Nwk_Man_t) {
	type pair struct{ in, out int }
	hist := map[int]pair{}
	bucket := func(n int) int {
		k := 0
		for n >= 10 {
			n /= 10
			k += 10
		}
		return k + n
	}
	var maxIn, maxOut, totalIn, totalOut int
	for _, o := range p.Objects {
		if o == nil || o.Type != NWK_OBJ_NODE {
			continue
		}
		a, b := len(o.Fanins), len(o.Fanouts)
		maxIn = max(maxIn, a)
		maxOut = max(maxOut, b)
		totalIn += a
		totalOut += b
		v := hist[bucket(a)]
		v.in++
		hist[bucket(a)] = v
		v = hist[bucket(b)]
		v.out++
		hist[bucket(b)] = v
	}
	keys := []int{}
	for k := range hist {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	fmt.Fprintln(writer(p), "Fanio distribution (bucket: fanins fanouts)")
	for _, k := range keys {
		v := hist[k]
		fmt.Fprintf(writer(p), "%d: %d %d\n", k, v.in, v.out)
	}
	count := max(int(Nwk_ManNodeNum(p)), 1)
	fmt.Fprintf(writer(p), "Fanins: Max=%d Ave=%.2f. Fanouts: Max=%d Ave=%.2f.\n", maxIn, float64(totalIn)/float64(count), maxOut, float64(totalOut)/float64(count))
}
func Nwk_ManCleanMarks(p *Nwk_Man_t) {
	for _, o := range p.Objects {
		if o != nil {
			o.MarkA = false
			o.MarkB = false
		}
	}
}
func Nwk_ManMinimumBaseNode(o *Nwk_Obj_t, scratch *Vec_Int_t, verbose int32) int32 {
	n := len(o.Fanins)
	if n > 20 {
		panic("nwk: support minimization exceeds 20 variables")
	}
	truth := truthWords(o, scratch)
	keep := []int{}
	for variable := 0; variable < n; variable++ {
		dependent := false
		for m := 0; m < 1<<n; m++ {
			if m>>variable&1 == 0 && truthBit(truth, m) != truthBit(truth, m|1<<variable) {
				dependent = true
				break
			}
		}
		if dependent {
			keep = append(keep, variable)
		}
	}
	if len(keep) == n {
		return 0
	}
	new := Nwk_ManCreateNode(o.Manager, int32(len(keep)), int32(len(o.Fanouts)))
	bindings := make([]*Hop_Obj_t, n)
	for i := range bindings {
		bindings[i] = hop.Hop_ManConst0(o.Manager.Hop)
	}
	for i, oldIndex := range keep {
		Nwk_ObjAddFanin(new, o.Fanins[oldIndex])
		bindings[oldIndex] = hop.Hop_IthVar(o.Manager.Hop, int32(i))
	}
	new.Function = SubstituteFunction(o.Manager.Hop, o.Function, bindings)
	if verbose != 0 {
		fmt.Fprintf(writer(o.Manager), "Reducing node %d fanins from %d to %d.\n", o.ID, n, len(keep))
	}
	Nwk_ObjReplace(o, new)
	return 1
}
func Nwk_ManMinimumBaseInt(p *Nwk_Man_t, verbose int32) int32 {
	scratch := vec.Vec_IntAlloc(0)
	var count int32
	objects := append([]*Nwk_Obj_t(nil), p.Objects...)
	for _, o := range objects {
		if o != nil && o.Type == NWK_OBJ_NODE {
			count += Nwk_ManMinimumBaseNode(o, scratch, verbose)
		}
	}
	return count
}
func Nwk_ManMinimumBaseRec(p *Nwk_Man_t, verbose int32) {
	for Nwk_ManMinimumBaseInt(p, verbose) != 0 {
	}
}
func Nwk_ManMinimumBase(p *Nwk_Man_t, verbose int32) { Nwk_ManMinimumBaseInt(p, verbose) }
func Nwk_ManRemoveDupFaninsNode(o *Nwk_Obj_t, a, b int32, scratch *Vec_Int_t) {
	if o.Fanins[a] != o.Fanins[b] {
		panic("nwk: distinct duplicate fanins")
	}
	o.Function = hop.Hop_Compose(o.Manager.Hop, o.Function, hop.Hop_IthVar(o.Manager.Hop, a), b)
	Nwk_ManMinimumBaseNode(o, scratch, 0)
}
func Nwk_ManRemoveDupFanins(p *Nwk_Man_t, verbose int32) {
	scratch := vec.Vec_IntAlloc(0)
	for _, o := range append([]*Nwk_Obj_t(nil), p.Objects...) {
		if o == nil || o.Type != NWK_OBJ_NODE {
			continue
		}
		found := false
		for i := 0; i < len(o.Fanins) && !found; i++ {
			for j := i + 1; j < len(o.Fanins); j++ {
				if o.Fanins[i] == o.Fanins[j] {
					Nwk_ManRemoveDupFaninsNode(o, int32(i), int32(j), scratch)
					found = true
					break
				}
			}
		}
	}
}
