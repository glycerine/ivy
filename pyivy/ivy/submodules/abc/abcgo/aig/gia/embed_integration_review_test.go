package gia

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"
)

func TestEmbeddingTransfersOriginalObjectCoordinates(t *testing.T) {
	for _, interleaved := range []bool{false, true} {
		for _, cluster := range []int32{0, 1} {
			for _, refine := range []int32{0, 1} {
				t.Run(fmt.Sprintf("interleaved%t/cluster%d/refine%d", interleaved, cluster, refine), func(t *testing.T) {
					t.Chdir(t.TempDir())
					g := Gia_ManStart(1)
					a, b := Gia_ManAppendCi(g), Gia_ManAppendCi(g)
					var c int32
					if !interleaved {
						c = Gia_ManAppendCi(g)
					}
					x := Gia_ManAppendAnd(g, a, b)
					Gia_ManAppendCo(g, x)
					if interleaved {
						c = Gia_ManAppendCi(g)
					}
					y := Gia_ManAppendAnd(g, x, c)
					Gia_ManAppendCo(g, y)
					Gia_ManSetRegNum(g, 1)
					g.SetIdentity("interleaved", "")
					pars := Emb_Par_t{NDims: 5, NSols: 2, NIters: 2, FCluster: cluster, FRefine: refine, FDump: 1}
					Gia_ManSolveProblem(g, &pars)
					if len(g.pPlacement) != int(g.nObjs) {
						t.Fatal("missing placement on source graph")
					}
					plot, err := os.ReadFile("interleaved.plt")
					if err != nil {
						t.Fatal(err)
					}
					var points [][2]int
					for _, line := range strings.Split(string(plot), "\n") {
						fields := strings.Fields(line)
						if len(fields) != 2 {
							continue
						}
						x, ex := strconv.Atoi(fields[0])
						y, ey := strconv.Atoi(fields[1])
						if ex == nil && ey == nil {
							points = append(points, [2]int{x, y})
						}
					}
					if len(points) == 0 || len(points)%2 != 0 {
						t.Fatal("plot has no complete edge segments")
					}
					edges := make(map[[4]int]bool)
					for i := 0; i < len(points); i += 2 {
						edges[[4]int{points[i][0], points[i][1], points[i+1][0], points[i+1][1]}] = true
					}
					checkEdge := func(from, to int32) {
						t.Helper()
						a, b := g.pPlacement[from], g.pPlacement[to]
						if !edges[[4]int{int(a.xCoord), int(a.yCoord), int(b.xCoord), int(b.yCoord)}] {
							t.Errorf("source edge %d -> %d has transferred coordinates absent from the plot", from, to)
						}
					}
					// Compare independent public outputs of the complete pipeline:
					// plotted geometry versus coordinates on the original GIA IDs.
					for _, obj := range g.pObjs {
						for _, fanin := range giaObjectFanins(g, obj) {
							checkEdge(fanin, obj.id)
						}
					}
					checkEdge(Gia_ManRi(g, 0).id, Gia_ManRo(g, 0).id)
				})
			}
		}
	}
}

func BenchmarkEmbeddingPlotPreparation(b *testing.B) {
	p := &Emb_Man_t{nObjs: 256, objects: make([]*Emb_Obj_t, 256), pPlacement: make([]uint16, 512)}
	for i := range p.objects {
		p.objects[i] = &Emb_Obj_t{Value: uint32(i)}
		p.pPlacement[2*i] = uint16(i * 257)
		p.pPlacement[2*i+1] = uint16(i * 7919)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Emb_ManDumpGnuplotPrepare(p)
	}
}
