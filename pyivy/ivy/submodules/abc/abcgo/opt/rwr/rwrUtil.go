package rwr

// Forest serialization and practical-class setup from rwrUtil.c.
import (
	"encoding/binary"
	"fmt"
	"os"
	"time"
)

func rwrForestEntry(node *Rwr_Node_t) (uint32, uint32) {
	a := uint32(node.p0.node.Id)<<1 | uint32(Rwr_IsComplement(node.p0))
	b := uint32(node.p1.node.Id)<<1 | uint32(Rwr_IsComplement(node.p1))
	a <<= 1
	if node.fExor {
		a |= 1
	}
	return a, b
}
func rwrLoadEntry(p *Rwr_Man_t, first, second uint32) {
	xor := int32(first & 1)
	first >>= 1
	a, b := p.vForest[first>>1], p.vForest[second>>1]
	level := int32(1 + max(a.Level, b.Level))
	volume := 1 + Rwr_ManNodeVolume(p, a, b)
	Rwr_ManAddNode(p, Rwr_NotCond(Rwr_NodeRef{node: a}, int32(first&1)), Rwr_NotCond(Rwr_NodeRef{node: b}, int32(second&1)), xor, level, volume+xor)
}
func Rwr_ManLoadFromArray(p *Rwr_Man_t, verbose int32) {
	started := time.Now()
	i := 0
	for ; ; i++ {
		a, b := s_RwtAigSubgraphs[2*i], s_RwtAigSubgraphs[2*i+1]
		if a == 0 && b == 0 {
			break
		}
		rwrLoadEntry(p, uint32(a), uint32(b))
	}
	if verbose != 0 {
		fmt.Printf("The number of classes = %d. Canonical nodes = %d.\nThe number of nodes loaded = %d.  Loading: %v\n", p.nClasses, p.nAdded, i-1, time.Since(started))
	}
}
func Rwr_ManWriteToArray(p *Rwr_Man_t) {
	started := time.Now()
	f, err := os.Create("npn4_aig_array.txt")
	if err != nil {
		panic(err)
	}
	fmt.Fprint(f, "static unsigned short s_RwtAigSubgraphs[] = \n{")
	count := len(p.vForest) - 5
	for i, node := range p.vForest[5:] {
		if i%5 == 0 {
			fmt.Fprint(f, "\n    ")
		}
		a, b := rwrForestEntry(node)
		fmt.Fprintf(f, "0x%04X,0x%04X, ", a, b)
	}
	if count%5 == 0 {
		fmt.Fprint(f, "\n    ")
	}
	fmt.Fprint(f, "0x0000,0x0000 \n};\n")
	if err = f.Close(); err != nil {
		panic(err)
	}
	fmt.Printf("The number of nodes saved = %d.   Saving: %v\n", count, time.Since(started))
}
func Rwr_ManWriteToFile(p *Rwr_Man_t, path string) {
	started := time.Now()
	f, err := os.Create(path)
	if err != nil {
		panic(err)
	}
	write := func(x uint32) {
		if e := binary.Write(f, binary.NativeEndian, x); e != nil {
			panic(e)
		}
	}
	count := len(p.vForest) - 5
	write(uint32(count))
	for _, node := range p.vForest[5:] {
		a, b := rwrForestEntry(node)
		write(a)
		write(b)
	}
	if err = f.Close(); err != nil {
		panic(err)
	}
	fmt.Printf("The number of nodes saved = %d.   Saving: %v\n", count, time.Since(started))
}
func Rwr_ManLoadFromFile(p *Rwr_Man_t, path string) {
	started := time.Now()
	f, err := os.Open(path)
	if err != nil {
		fmt.Printf("Rwr_ManLoadFromFile: Cannot open file %q.\n", path)
		return
	}
	defer f.Close()
	read := func() uint32 {
		var x uint32
		if e := binary.Read(f, binary.NativeEndian, &x); e != nil {
			panic(e)
		}
		return x
	}
	count := read()
	for i := uint32(0); i < count; i++ {
		rwrLoadEntry(p, read(), read())
	}
	fmt.Printf("The number of classes = %d. Canonical nodes = %d.\nThe number of nodes loaded = %d.   Loading: %v\n", p.nClasses, p.nAdded, count, time.Since(started))
}
func Rwr_ListAddToTail(head **Rwr_Node_t, node *Rwr_Node_t) { rwrAddTail(head, node) }
func Rwr_ManGetPractical(p *Rwr_Man_t) []bool {
	practical := make([]bool, p.nFuncs)
	practical[0] = true
	for _, truth := range s_RwrPracticalClasses[1:] {
		if truth == 0 {
			break
		}
		practical[truth] = true
	}
	return practical
}
