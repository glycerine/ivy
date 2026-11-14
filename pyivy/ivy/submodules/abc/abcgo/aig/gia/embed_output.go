package gia

import (
	"fmt"
	"math"
	"os"
	"os/exec"
	"time"

	"github.com/glycerine/abcgo/misc/vec"
)

func Emb_ManPrintStats(p *Emb_Man_t) {
	fmt.Printf("i/o =%7d/%7d  ", Emb_ManPiNum(p), Emb_ManPoNum(p))
	if p.nRegs != 0 {
		fmt.Printf("ff =%7d  ", p.nRegs)
	}
	words := 0
	for _, o := range p.objects {
		words += 5 + len(o.fanins) + len(o.fanouts)
	}
	fmt.Printf("node =%8d  obj =%8d  mem =%5.2f MB\n", Emb_ManNodeNum(p), p.nObjs, 4*float64(words)/(1<<20))
}
func Emb_ManStop(p *Emb_Man_t) {
	if p != nil {
		*p = Emb_Man_t{}
	}
}
func Emb_ManPrintFanio(p *Emb_Man_t) {
	in, out := map[int]int{}, map[int]int{}
	maxIn, maxOut, sumIn, sumOut := 0, 0, 0, 0
	bucket := func(n int) int {
		b := 0
		for n >= 10 {
			n /= 10
			b += 10
		}
		return b + n
	}
	for i, o := range p.objects {
		if i == 0 || Emb_ObjIsTerm(o) != 0 {
			continue
		}
		a, b := len(o.fanins), len(o.fanouts)
		maxIn = max(maxIn, a)
		maxOut = max(maxOut, b)
		sumIn += a
		sumOut += b
		in[bucket(a)]++
		out[bucket(b)]++
	}
	fmt.Println("The distribution of fanins and fanouts in the network:")
	fmt.Println("         Number   Nodes with fanin  Nodes with fanout")
	for k := 0; k <= max(bucket(maxIn), bucket(maxOut)); k++ {
		if in[k] == 0 && out[k] == 0 {
			continue
		}
		label := fmt.Sprint(k)
		if k >= 10 {
			pow := int(math.Pow10(k / 10))
			label = fmt.Sprintf("%d - %d", pow*(k%10), pow*(k%10+1)-1)
		}
		fmt.Printf("%15s : %12d      %12d\n", label, in[k], out[k])
	}
	fmt.Printf("Fanins: Max = %d. Ave = %.2f.  Fanouts: Max = %d. Ave =  %.2f.\n", maxIn, float64(sumIn)/float64(Emb_ManNodeNum(p)), maxOut, float64(sumOut)/float64(Emb_ManNodeNum(p)))
}
func Emb_ManComputeDistance_old(p *Emb_Man_t, pivot *Emb_Obj_t) int32 {
	p.nReached = 0
	Emb_ManIncrementTravId(p)
	Emb_ObjSetTravIdCurrent(p, pivot)
	a, b := vec.Vec_IntAlloc(1), vec.Vec_IntAlloc(1)
	vec.Vec_IntPush(a, pivot.hHandle)
	Emb_ManPerformBfs(p, a, b, nil)
	return p.nDistMax
}
func Gia_ManTestDistanceInternal(p *Emb_Man_t) {
	Gia_ManRandom(1)
	Emb_ManResetTravId(p)
	for _, category := range []string{"inputs", "outputs", "nodes"} {
		start := time.Now()
		fmt.Printf("From %s: ", category)
		for i := 0; i < 20; i++ {
			var pivot *Emb_Obj_t
			for attempts := 0; attempts < 10000 && pivot == nil; attempts++ {
				switch category {
				case "inputs":
					if Emb_ManCiNum(p) > 0 {
						o := Emb_ManCi(p, int32(Gia_ManRandom(0)%uint32(Emb_ManCiNum(p))))
						if len(o.fanouts) > 0 && Emb_ObjIsNode(o.fanouts[0]) != 0 {
							pivot = o
						}
					}
				case "outputs":
					if Emb_ManCoNum(p) > 0 {
						o := Emb_ManCo(p, int32(Gia_ManRandom(0)%uint32(Emb_ManCoNum(p))))
						if len(o.fanins) > 0 && Emb_ObjIsNode(o.fanins[0]) != 0 {
							pivot = o
						}
					}
				default:
					id := int32(Gia_ManRandom(0) % uint32(p.pGia.nObjs))
					handle := Gia_ObjValue(p.pGia.pObjs[id])
					if handle >= 0 && Emb_ObjIsNode(Emb_ManObj(p, handle)) != 0 {
						pivot = Emb_ManObj(p, handle)
					}
				}
			}
			if pivot == nil {
				break
			}
			fmt.Printf("%d ", Emb_ManComputeDistance_old(p, pivot))
		}
		fmt.Printf("Time = %.2f sec\n", time.Since(start).Seconds())
	}
}
func Gia_ManTestDistance(g *Gia_Man_t) {
	start := time.Now()
	p := Emb_ManStart(g)
	Emb_ManPrintStats(p)
	fmt.Printf("Time = %.2f sec\n", time.Since(start).Seconds())
	Gia_ManTestDistanceInternal(p)
	Emb_ManStop(p)
}
func Emb_DumpGraphIntoFile(p *Emb_Man_t) {
	f, err := os.Create("1.g")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()
	for _, o := range p.objects {
		if Emb_ObjIsTravIdCurrent(p, o) == 0 {
			continue
		}
		for _, next := range o.fanouts {
			if Emb_ObjIsTravIdCurrent(p, next) == 0 {
				panic("gia: graph neighbor not visited")
			}
			fmt.Fprintf(f, "%d %d\n", o.Value, next.Value)
		}
	}
}
func Emb_ManPrintSolutions(p *Emb_Man_t, nSols int32) {
	for i := int32(0); i < nSols; i++ {
		for _, v := range Emb_ManSol(p, i) {
			fmt.Printf("%4d ", int(100*v))
		}
		fmt.Println()
	}
}
func Emb_ManDumpGnuplotPrepare(p *Emb_Man_t) *Vec_Int_t {
	const rows, cols = 500, 700
	bitmap := make([][]bool, rows)
	for i := range bitmap {
		bitmap[i] = make([]bool, cols+1)
	}
	for _, o := range p.objects {
		x := int(p.pPlacement[2*o.Value]) * cols / (1 << 16)
		y := int(p.pPlacement[2*o.Value+1]) * rows / (1 << 16)
		bitmap[y][x] = true
	}
	lines := vec.Vec_IntAlloc(1000)
	for y, row := range bitmap {
		in := false
		for x, on := range row {
			if on && !in {
				vec.Vec_IntPush(lines, int32(x))
				vec.Vec_IntPush(lines, int32(y))
				in = true
			}
			if !on && in {
				vec.Vec_IntPush(lines, int32(x-1))
				vec.Vec_IntPush(lines, int32(y))
				in = false
			}
		}
	}
	return lines
}
func Emb_ManDumpGnuplot(p *Emb_Man_t, name string, large, show int32) {
	if p.pPlacement == nil {
		fmt.Println("Emb_ManDumpGnuplot(): Placement is not available.")
		return
	}
	filename := Gia_FileNameGenericAppend(name, ".plt")
	f, err := os.Create(filename)
	if err != nil {
		fmt.Println(err)
		return
	}
	stamp := time.Now().Format(time.RFC1123)
	fmt.Fprintf(f, "# This Gnuplot file was produced by ABC on %s\n\nset nokey\n\n", stamp)
	if show == 0 {
		fmt.Fprintf(f, "set terminal gif font 'arial' 10 size 800,600 xffffff x000000 x000000 x000000\nset output '%s'\n\n", Gia_FileNameGenericAppend(name, ".gif"))
	}
	fmt.Fprintf(f, "set title %q font \"Times, 12\"\n\nplot [:] '-' w l\n\n", fmt.Sprintf("%s :  PI = %d   PO = %d   FF = %d   Node = %d   Obj = %d  HPWL = %.2e\n(image generated by ABC and Gnuplot on %s)", name, Emb_ManPiNum(p), Emb_ManPoNum(p), Emb_ManRegNum(p), Emb_ManNodeNum(p), p.nObjs, Emb_ManComputeHPWL(p), stamp))
	if large != 0 {
		data := vec.Vec_IntArray(Emb_ManDumpGnuplotPrepare(p))
		for i := 0; i < len(data); i += 4 {
			fmt.Fprintf(f, "%5d %5d\n%5d %5d\n\n", data[i], data[i+1], data[i+2], data[i+3])
		}
	} else {
		for _, o := range p.objects {
			if Emb_ObjIsTravIdCurrent(p, o) == 0 {
				continue
			}
			for _, n := range o.fanouts {
				if Emb_ObjIsTravIdCurrent(p, n) == 0 {
					panic("gia: graph neighbor not visited")
				}
				fmt.Fprintf(f, "%5d %5d\n%5d %5d\n\n", p.pPlacement[2*o.Value], p.pPlacement[2*o.Value+1], p.pPlacement[2*n.Value], p.pPlacement[2*n.Value+1])
			}
		}
	}
	fmt.Fprint(f, "EOF\n\n")
	if show != 0 {
		fmt.Fprintln(f, "pause -1 \"Close window\"\nreset")
	} else {
		fmt.Fprintln(f, "# pause -1 \"Close window\"\n# reset")
	}
	if err = f.Close(); err != nil {
		fmt.Println(err)
	}
	if show != 0 {
		cmd := exec.Command("gnuplot", filename)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		if err = cmd.Run(); err != nil {
			fmt.Println(err)
		}
	}
}
func Gia_ManSolveProblem(g *Gia_Man_t, pars *Emb_Par_t) {
	var p *Emb_Man_t
	if pars.FCluster != 0 {
		p = Emb_ManStart(g)
		if pars.FVerbose != 0 {
			fmt.Print("Clustered: ")
			Emb_ManPrintStats(p)
		}
	} else {
		p = Emb_ManStartSimple(g)
	}
	p.fVerbose = pars.FVerbose
	Gia_ManRandom(1)
	Emb_ManResetTravId(p)
	Emb_ManSetValue(p)
	Emb_ManComputeDimensions(p, pars.NDims)
	Emb_ManComputeCovariance(p, pars.NDims)
	Emb_ManComputeEigenvectors(p, pars.NDims, pars.NSols)
	Emb_ManComputeSolutions(p, pars.NDims, pars.NSols)
	Emb_ManDerivePlacement(p, pars.NSols)
	if pars.FRefine != 0 {
		Emb_ManPlacementRefine(p, pars.NIters, pars.FVerbose)
	}
	if (pars.FDump != 0 || pars.FDumpLarge != 0) && pars.NSols == 2 {
		Emb_ManDumpGnuplot(p, g.pName, pars.FDumpLarge, pars.FShowImage)
	}
	if g.nObjs == p.nObjs && p.pPlacement != nil {
		g.pPlacement = make([]Gia_Plc_t, p.nObjs)
		for i := range g.pPlacement {
			g.pPlacement[i].xCoord = uint32(p.pPlacement[2*i])
			g.pPlacement[i].yCoord = uint32(p.pPlacement[2*i+1])
		}
	}
	Emb_ManStop(p)
}
