package main

import (
	"flag"
	"fmt"
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/ioa"
	"github.com/glycerine/abcgo/aig/saig"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
)

func eval(m *ioa.Model, in, state []bool) ([]bool, []bool) {
	v := make([]bool, int(m.MaxVar)+1)
	for i, l := range m.Inputs {
		v[l>>1] = in[i]
	}
	for i, l := range m.Latches {
		v[l.Literal>>1] = state[i]
	}
	value := func(l uint32) bool { return v[l>>1] != (l&1 != 0) }
	for _, n := range m.ANDs {
		v[n.Literal>>1] = value(n.Left) && value(n.Right)
	}
	o := make([]bool, len(m.Outputs))
	for i, l := range m.Outputs {
		o[i] = value(l)
	}
	next := make([]bool, len(m.Latches))
	for i, l := range m.Latches {
		next[i] = value(l.Next)
	}
	return o, next
}
func main() {
	abcPath := flag.String("abc", "../abc", "path to ABC executable")
	flag.Parse()
	dir, err := os.MkdirTemp("", "saig-retime-diff-")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(dir)
	r := rand.New(rand.NewSource(9917))
	different, unknown, compared := 0, 0, 0
	for trial := 0; trial < 24; trial++ {
		g := aig.NewGraph()
		es := []aig.Edge{g.NewInput(), g.NewInput(), g.NewInput(), g.NewInput(), g.NewInput()}
		for i := 0; i < 10; i++ {
			a, b := es[r.Intn(len(es))], es[r.Intn(len(es))]
			if r.Intn(2) == 0 {
				a = a.Not()
			}
			if r.Intn(2) == 0 {
				b = b.Not()
			}
			es = append(es, g.AND(a, b))
		}
		for i := 0; i < 5; i++ {
			e := es[r.Intn(len(es))]
			if r.Intn(2) == 0 {
				e = e.Not()
			}
			g.AddOutput(e)
		}
		aig.Aig_ManSetRegNum(g, 3)
		aig.Aig_ManCleanup(g)
		src, dst := filepath.Join(dir, "in.aig"), filepath.Join(dir, "out.aig")
		if err := ioa.Ioa_WriteAiger(g, src, 0, 0); err != nil {
			panic(err)
		}
		os.Remove(dst)
		script := fmt.Sprintf("read_aiger %s; retime -M 3; strash; write_aiger %s", src, dst)
		out, err := exec.Command(*abcPath, "-c", script).CombinedOutput()
		if err != nil {
			panic(fmt.Sprintf("ABC trial %d: %v %s", trial, err, out))
		}
		model, err := ioa.Ioa_ReadAiger(dst, 1)
		if err != nil {
			panic(fmt.Sprintf("ABC output trial %d: %v %s", trial, err, out))
		}
		hasUnknown := false
		for _, l := range model.Latches {
			if l.Init != 0 && l.Init != 1 {
				hasUnknown = true
			}
		}
		if hasUnknown {
			unknown++
			continue
		}
		compared++
		translated := saig.Saig_ManRetimeMinArea(g, 1000, 0, 0, 1, 0)
		if int(translated.RegisterCount()) != len(model.Latches) {
			different++
			clean := aig.Aig_ManDupSimple(g)
			aig.Aig_ManSeqCleanup(clean)
			fmt.Printf("trial %d: C registers=%d Go registers=%d cleanup=%d\n", trial, len(model.Latches), translated.RegisterCount(), clean.RegisterCount())
		}
		for seq := 0; seq < 64; seq++ {
			a, b, c := make([]bool, 3), make([]bool, translated.RegisterCount()), make([]bool, len(model.Latches))
			for i, l := range model.Latches {
				if l.Init == 1 {
					c[i] = true
				} else if l.Init != 0 {
					panic(fmt.Sprintf("unknown C initial latch %+v", l))
				}
			}
			for frame := 0; frame < 3; frame++ {
				in := []bool{seq&(1<<(2*frame)) != 0, seq&(1<<(2*frame+1)) != 0}
				old, err := g.Evaluate(append(append([]bool(nil), in...), a...))
				if err != nil {
					panic(err)
				}
				got, err := translated.Evaluate(append(append([]bool(nil), in...), b...))
				if err != nil {
					panic(err)
				}
				ref, next := eval(model, in, c)
				for i := 0; i < 2; i++ {
					if old[i] != got[i] || old[i] != ref[i] {
						panic(fmt.Sprintf("trial %d seq %d frame %d output %d: source=%v go=%v C=%v", trial, seq, frame, i, old, got, ref))
					}
				}
				copy(a, old[len(old)-len(a):])
				copy(b, got[len(got)-len(b):])
				copy(c, next)
			}
		}
	}
	fmt.Printf("%d circuits, 64 three-frame traces each: C and Go outputs match; register counts differ in %d cases; %d C outputs had unknown initial latches\n", compared, different, unknown)
}
