package rwr

// Rewrite library diagnostics from rwrPrint.c.
import (
	"fmt"
	"io"
	"os"
)

func Rwr_Trav2_rec(p *Rwr_Man_t, node *Rwr_Node_t, volume *int32) {
	if node.fUsed || node.TravId == p.nTravIds {
		return
	}
	node.TravId = p.nTravIds
	*volume++
	Rwr_Trav2_rec(p, node.p0.node, volume)
	Rwr_Trav2_rec(p, node.p1.node, volume)
}
func Rwr_GetBushVolume(p *Rwr_Man_t, entry int32, volume, functions *int32) {
	var v, n int32
	Rwr_ManIncTravId(p)
	for node := p.pTable[entry]; node != nil; node = node.pNext {
		if node.uTruth != p.puCanons[node.uTruth] {
			continue
		}
		n++
		Rwr_Trav2_rec(p, node, &v)
	}
	*volume, *functions = v, n
}
func Rwr_GetBushSumOfVolumes(p *Rwr_Man_t, entry int32) int32 {
	var total int32
	for node := p.pTable[entry]; node != nil; node = node.pNext {
		if node.uTruth != p.puCanons[node.uTruth] {
			continue
		}
		var v int32
		Rwr_ManIncTravId(p)
		Rwr_Trav2_rec(p, node, &v)
		total += v
	}
	return total
}
func Rwr_NodePrint_rec(w io.Writer, node *Rwr_Node_t) {
	if node.Id == 0 {
		fmt.Fprint(w, "Const1")
		return
	}
	if node.Id < 5 {
		fmt.Fprintf(w, "%c", 'a'+node.Id-1)
		return
	}
	child := func(ref Rwr_NodeRef) {
		if ref.node.Id >= 5 {
			fmt.Fprint(w, "(")
		}
		Rwr_NodePrint_rec(w, ref.node)
		if ref.node.Id >= 5 {
			fmt.Fprint(w, ")")
		}
		if ref.complemented {
			fmt.Fprint(w, "'")
		}
	}
	child(node.p0)
	if node.fExor {
		fmt.Fprint(w, "+")
	}
	child(node.p1)
}
func Rwr_NodePrint(w io.Writer, p *Rwr_Man_t, node *Rwr_Node_t) {
	fmt.Fprintf(w, "%5d : 0x%04X tt=%016b lev=%d vol=%d  ", node.Id, node.uTruth, node.uTruth, node.Level, node.Volume)
	Rwr_NodePrint_rec(w, node)
	fmt.Fprintln(w)
}
func Rwr_ManPrint(p *Rwr_Man_t) {
	f, err := os.Create("graph_lib.txt")
	if err != nil {
		panic(err)
	}
	counter := 0
	for i := 0; i < 1<<16; i++ {
		if p.pTable[i] == nil || int(p.puCanons[i]) != i {
			continue
		}
		fmt.Fprintf(f, "\nClass %3d. Func %6d.  ", p.pMap[i], counter)
		counter++
		var volume, functions int32
		Rwr_GetBushVolume(p, int32(i), &volume, &functions)
		fmt.Fprintf(f, "Roots = %3d. Vol = %3d. Sum = %3d.  %016b\n", functions, volume, Rwr_GetBushSumOfVolumes(p, int32(i)), i)
		for node := p.pTable[i]; node != nil; node = node.pNext {
			if node.uTruth == p.puCanons[node.uTruth] {
				Rwr_NodePrint(f, p, node)
			}
		}
	}
	if err = f.Close(); err != nil {
		panic(err)
	}
}
