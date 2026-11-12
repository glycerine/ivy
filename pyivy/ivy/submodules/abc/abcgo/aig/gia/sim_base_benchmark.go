package gia

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/glycerine/abcgo/misc/vec"
)

func Gia_ManSimTest(g *Gia_Man_t) {
	const words = 4
	previous := g.vSimsPi
	g.vSimsPi = vec.Vec_WrdStartRandom(int(Gia_ManCiNum(g)) * words)
	defer func() { g.vSimsPi = previous }()
	start := time.Now()
	for i := 0; i < 20; i++ {
		Gia_ManSimPatSim(g)
	}
	fmt.Printf("Time1: %s\n", time.Since(start))
	start = time.Now()
	for i := 0; i < 20; i++ {
		Gia_ManSimPatSim2(g)
	}
	fmt.Printf("Time2: %s\n", time.Since(start))
}

// giaWriteCompiledSimulation generates a self-contained Go benchmark.
func giaWriteCompiledSimulation(w io.Writer, g *Gia_Man_t) {
	fmt.Fprintln(w, "package main\n\nimport (\"fmt\"; \"math/rand\"; \"time\")\n\nfunc main() {\n start := time.Now()\n var result uint64\n for i:=0;i<2000;i++ {")
	for k := 0; k < 4; k++ {
		fmt.Fprintf(w, "s%07d_%d := uint64(0); _ = s%07d_%d\n", 0, k, 0, k)
	}
	for _, id := range vec.Vec_IntArray(g.vCis) {
		for k := 0; k < 4; k++ {
			fmt.Fprintf(w, "s%07d_%d := rand.Uint64(); _ = s%07d_%d\n", id, k, id, k)
		}
	}
	for _, obj := range g.pObjs {
		if Gia_ObjIsAnd(obj) == 0 {
			continue
		}
		op := "&"
		if Gia_ObjIsXor(obj) != 0 {
			op = "^"
		}
		a, b := "", ""
		if obj.fCompl0 != 0 {
			a = "^"
		}
		if obj.fCompl1 != 0 {
			b = "^"
		}
		for k := 0; k < 4; k++ {
			fmt.Fprintf(w, "s%07d_%d := %ss%07d_%d %s %ss%07d_%d; _ = s%07d_%d\n", obj.id, k, a, Gia_ObjFaninId0(obj, obj.id), k, op, b, Gia_ObjFaninId1(obj, obj.id), k, obj.id, k)
		}
	}
	for _, id := range vec.Vec_IntArray(g.vCos) {
		obj := Gia_ManObj(g, id)
		prefix := ""
		if obj.fCompl0 != 0 {
			prefix = "^"
		}
		for k := 0; k < 4; k++ {
			fmt.Fprintf(w, "result ^= %ss%07d_%d\n", prefix, Gia_ObjFaninId0(obj, id), k)
		}
	}
	io.WriteString(w, "}\n fmt.Printf(\"Res = 0x%08x    Time = %s\\n\", uint32(result), time.Since(start))\n}")
}
func Gia_ManSimGen(g *Gia_Man_t) {
	f, err := os.Create("comp_sim.go")
	if err != nil {
		fmt.Printf("Cannot open compiled simulation: %v\n", err)
		return
	}
	defer f.Close()
	giaWriteCompiledSimulation(f, g)
}
