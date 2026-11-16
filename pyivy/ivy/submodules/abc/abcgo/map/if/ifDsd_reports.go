package if_

import (
	"fmt"
	"io"
	"os"
	"unsafe"

	abcutil "github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

func If_DsdManCheckNonDec_rec(p *If_DsdMan_t, id int32) int32 {
	o := If_DsdVecObj(&p.vObjs, id)
	if o.Type == IF_DSD_PRIME {
		return 1
	}
	for _, fan := range o.pFans {
		if If_DsdManCheckNonDec_rec(p, fan>>1) != 0 {
			return 1
		}
	}
	return 0
}

func If_DsdManPrintDistrib(p *If_DsdMan_t) {
	// Keep totals separate from all support slots, including the reserved
	// variable (support one) in a zero-input manager.
	var objects, primes, classes, structures, nonDSD, marked [18]int
	const total = 17
	for v := 3; v <= int(p.nVars); v++ {
		classes[v] = vec.Vec_MemEntryNum(p.vTtMem[v])
		classes[total] += classes[v]
	}
	for id, entry := range Vec_PtrArray(&p.vObjs) {
		o := entry.(*If_DsdObj_t)
		objects[o.nFans]++
		objects[total]++
		structures[o.nSupp]++
		structures[total]++
		if o.Type == IF_DSD_PRIME {
			primes[o.nFans]++
			primes[total]++
		}
		if If_DsdManCheckNonDec_rec(p, int32(id)) != 0 {
			nonDSD[o.nSupp]++
			nonDSD[total]++
		}
		if o.fMark != 0 {
			marked[o.nSupp]++
			marked[total]++
		}
	}
	fmt.Println("***** DSD MANAGER STATISTICS *****")
	fmt.Println("Support     Obj   ObjNDSD            NPNNDSD                  Str   StrNDSD             Marked  ")
	for row := 0; row <= int(p.nVars)+1; row++ {
		i := row
		if row == int(p.nVars)+1 {
			i = total
			fmt.Print("All : ")
		} else {
			fmt.Printf("%3d : ", i)
		}
		fmt.Printf("%9d %9d %6.2f %% %9d %6.2f %%   %9d %9d %6.2f %% %9d %6.2f %%\n",
			objects[i], primes[i], 100*float64(primes[i])/float64(max(1, objects[i])), classes[i], 100*float64(classes[i])/float64(max(1, objects[i])),
			structures[i], nonDSD[i], 100*float64(nonDSD[i])/float64(max(1, structures[i])), marked[i], 100*float64(marked[i])/float64(max(1, structures[i])))
	}
}

func If_DsdManPrintOccurs(out io.Writer, p *If_DsdMan_t) {
	// Counts are 18-bit in the original object representation.
	var bins [70]int
	maximum, sum := int32(0), int64(0)
	for _, entry := range Vec_PtrArray(&p.vObjs)[2:] {
		count := entry.(*If_DsdObj_t).Count
		maximum = max(maximum, count)
		sum += int64(count)
		bin, digit := 0, count
		for digit >= 10 {
			digit /= 10
			bin += 10
		}
		bins[bin+int(digit)]++
	}
	fmt.Fprintln(out, "The distribution of object occurrences:")
	cumulative := 0
	for bin, count := range bins {
		if count == 0 {
			continue
		}
		if bin < 10 {
			fmt.Fprintf(out, "%15d : ", bin)
		} else {
			scale := 1
			for i := 0; i < bin/10; i++ {
				scale *= 10
			}
			fmt.Fprintf(out, "%15s : ", fmt.Sprintf("%d - %d", scale*(bin%10), scale*(bin%10+1)-1))
		}
		cumulative += count
		fmt.Fprintf(out, "%12d   (%6.2f %%)\n", count, 100*float64(cumulative)/float64(If_DsdManObjNum(p)))
	}
	fmt.Fprintf(out, "Fanins: Max = %d. Ave = %.2f.\n", maximum, float64(sum)/float64(If_DsdManObjNum(p)))
}

const ifDsdReportNotations = `*****  NOTATIONS USED BELOW  *****
Support -- the support size
Obj     -- the number of nodes in the DSD manager for each support size
           (the constant node and the primary input node have no support)
ObjNDSD -- the number of prime nodes (that is, nodes whose function has no DSD)
           (percentage is relative to the number of all nodes of that size)
NPNNDSD -- the number of different NPN classes of prime nodes
           (Each NPN class may appear more than once. For example: F1 = 17(ab(cd))
           and F2 = 17(ab[cd]) both have prime majority node (hex TT is 17),
           but in one case the majority node is fed by AND, and in another by XOR.
           These two majority nodes are different nodes in the DSD manager
Str     -- the number of structures for each support size
           (each structure is composed of one or more nodes)
StrNDSD -- the number of DSD structures containing at least one prime node
Marked  -- the number of DSD structures matchable with the LUT structure (say, "44")
`

// ifDsdReportMemory estimates Go storage in the source's five categories. These
// are representation-specific estimates, not C allocation sizes or heap totals.
func ifDsdReportMemory(p *If_DsdMan_t) [5]float64 {
	var bytes [5]uintptr
	slot := unsafe.Sizeof(any(nil))
	for _, entry := range Vec_PtrArray(&p.vObjs) {
		o := entry.(*If_DsdObj_t)
		bytes[0] += unsafe.Sizeof(*o) + uintptr(cap(o.pFans))*4
	}
	bytes[1] = uintptr(vec.Vec_IntCap(&p.vTruths)) * 4
	for v := 3; v <= int(p.nVars); v++ {
		bytes[1] += uintptr(vec.Vec_MemEntrySize(p.vTtMem[v])*vec.Vec_MemEntryNum(p.vTtMem[v])) * 8
		ds := p.vTtDecs[v]
		bytes[3] += unsafe.Sizeof(*ds) + uintptr(vec.Vec_PtrCap(ds))*slot
		for _, entry := range vec.Vec_PtrArray(ds) {
			d := entry.(*vec.Vec_Int_t)
			bytes[3] += unsafe.Sizeof(*d) + uintptr(vec.Vec_IntCap(d))*4
		}
	}
	bytes[2] = uintptr(len(p.pBins)+vec.Vec_IntCap(&p.vNexts)) * 4
	bytes[4] = uintptr(vec.Vec_PtrCap(&p.vObjs)) * slot
	var result [5]float64
	for i, b := range bytes {
		result[i] = float64(b) / (1 << 20)
	}
	return result
}

func If_DsdManPrint(p *If_DsdMan_t, fileName string, number, support, occurs, dump, verbose int32) {
	var out io.Writer = os.Stdout
	if fileName != "" {
		f, err := os.Create(fileName)
		if err != nil {
			fmt.Println("cannot open output file")
			return
		}
		defer f.Close()
		out = f
	}
	if verbose != 0 {
		fmt.Fprint(out, ifDsdReportNotations)
	}
	If_DsdManPrintDistrib(p)
	yesNo := func(flag int32) string {
		if flag != 0 {
			return "yes"
		}
		return "no"
	}
	fmt.Printf("Number of inputs = %d.  LUT size = %d.  Marks = %s.  NewAsUseless = %s.  Bookmark = %d.\n", p.nVars, p.LutSize, yesNo(If_DsdManHasMarks(p)), yesNo(p.fNewAsUseless), p.nObjsPrev)
	if p.pCellStr != "" {
		fmt.Printf("Symbolic cell description: %s\n", p.pCellStr)
	}
	fmt.Fprintf(out, "Unique table misses        = %8d\nUnique table hits          = %8d\n", p.nUniqueMisses, p.nUniqueHits)
	labels := [...]string{"objects   ", "functions ", "hash table", "bound sets", "array     "}
	for i, mb := range ifDsdReportMemory(p) {
		fmt.Fprintf(out, "Memory used for %s = %8.2f MB.\n", labels[i], mb)
	}
	if p.timeDsd != 0 {
		Abc_PrintTime(1, "Time DSD   ", p.timeDsd)
		Abc_PrintTime(1, "Time canon ", p.timeCanon-p.timeCheck)
		Abc_PrintTime(1, "Time check ", p.timeCheck)
		Abc_PrintTime(1, "Time check2", p.timeCheck2)
		Abc_PrintTime(1, "Time verify", p.timeVerify)
	}
	if occurs != 0 {
		If_DsdManPrintOccurs(os.Stdout, p)
	}
	if dump != 0 {
		If_DsdManDumpDsd(p, support)
		If_DsdManDumpAll(p, support)
	}
	if verbose == 0 {
		return
	}
	var ids, costs []int32
	for id, entry := range Vec_PtrArray(&p.vObjs) {
		o := entry.(*If_DsdObj_t)
		if number != 0 && int32(id)%number != 0 || support != 0 && o.nSupp != support {
			continue
		}
		ids = append(ids, int32(id))
		costs = append(costs, -o.Count)
	}
	order := abcutil.Abc_MergeSortCost(costs, int32(len(costs)))
	for rank, index := range order[:min(20, len(order))] {
		// The original sends ranks to stdout even when records go to a file.
		fmt.Printf("%2d : ", rank+1)
		If_DsdManPrintOne(out, p, ids[index], nil, 1)
	}
}
