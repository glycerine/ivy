package bdc

import (
	"fmt"
	"io"
	"os"
)

func Bdc_ManAlloc(par *Bdc_Par_t) *Bdc_Man_t {
	if par == nil || par.NVarsMax <= 1 || par.NVarsMax >= 16 {
		panic("bdc: maximum support must be between 2 and 15")
	}
	return &Bdc_Man_t{Pars: *par, table: make(map[uint32][]*Bdc_Fun_t)}
}
func Bdc_ManFree(p *Bdc_Man_t) {
	if p == nil {
		return
	}
	if p.Pars.FVerbose != 0 {
		fmt.Printf("Bi-decomposition stats: Calls = %d. Nodes = %d. Reuse = %d.\n", p.calls, p.numNodes, p.numReuse)
		fmt.Printf("ANDs = %d. ORs = %d. Weak = %d. Muxes = %d.\n", p.numAnds, p.numOrs, p.numWeaks, p.numMuxes)
	}
	p.nodes = nil
	p.root = nil
	clear(p.table)
}
func Bdc_ManFunc(p *Bdc_Man_t, id int32) *Bdc_Fun_t { return p.nodes[id] }
func Bdc_ManRoot(p *Bdc_Man_t) *Bdc_Fun_t           { return p.root }
func Bdc_ManNodeNum(p *Bdc_Man_t) int32             { return int32(len(p.nodes)) }
func Bdc_ManAndNum(p *Bdc_Man_t) int32              { return int32(len(p.nodes)) - p.nVars - 1 }
func Bdc_ManPrepare(p *Bdc_Man_t, divisors [][]uint32) {
	if len(divisors) > 201 {
		divisors = divisors[:201]
	}
	Bdc_TableClear(p)
	p.nodes = nil
	p.root = nil
	p.nNodesNew = -1 - p.nVars - int32(len(divisors))
	f := Bdc_FunNew(p)
	if f == nil {
		panic("bdc: insufficient node budget for inputs")
	}
	f.Type = BDC_TYPE_CONST1
	f.Func = make([]uint32, truthWords(p.nVars))
	for i := range f.Func {
		f.Func[i] = ^uint32(0)
	}
	Bdc_TableAdd(p, f)
	for v := int32(0); v < p.nVars; v++ {
		f = Bdc_FunNew(p)
		f.Type = BDC_TYPE_PI
		f.Func = make([]uint32, truthWords(p.nVars))
		for a := 0; a < int(truthWords(p.nVars))*32; a++ {
			if a&(1<<uint(v)) != 0 {
				f.Func[a>>5] |= 1 << uint(a&31)
			}
		}
		f.Supp = 1 << uint(v)
		Bdc_TableAdd(p, f)
	}
	for _, truth := range divisors {
		f = Bdc_FunNew(p)
		f.Type = BDC_TYPE_PI
		f.Func = append([]uint32(nil), truth[:truthWords(p.nVars)]...)
		f.Supp = ttSupport(f.Func, p.nVars)
		Bdc_TableAdd(p, f)
	}
}
func Bdc_ManDecompose(p *Bdc_Man_t, truth, care []uint32, n int32, divisors [][]uint32, maxNodes int32) int32 {
	if n < 0 || n > p.Pars.NVarsMax || maxNodes < 0 {
		panic("bdc: invalid decomposition dimensions")
	}
	p.nVars, p.nNodesMax = n, maxNodes
	Bdc_ManPrepare(p, divisors)
	if care != nil && ttZero(care[:truthWords(p.nVars)]) {
		p.root = Bdc_Not(p.nodes[0])
		return 0
	}
	f := Bdc_Isf_t{}
	Bdc_IsfStart(p, &f)
	if care == nil {
		copy(f.On, truth[:truthWords(p.nVars)])
		f.Off = ttNot(truth[:truthWords(p.nVars)])
	} else {
		f.On = ttBinary(care[:truthWords(p.nVars)], truth[:truthWords(p.nVars)], '&')
		f.Off = ttBinary(care[:truthWords(p.nVars)], truth[:truthWords(p.nVars)], '-')
	}
	Bdc_SuppMinimize(p, &f)
	p.root = Bdc_ManDecompose_rec(p, &f)
	p.calls++
	p.numNodes += p.nNodesNew
	if p.root == nil {
		return -1
	}
	if Bdc_ManNodeVerify(p, &f, p.root) == 0 {
		panic("bdc: decomposition verification failed")
	}
	return p.nNodesNew
}
func writeNode(w io.Writer, p *Bdc_Man_t, f *Bdc_Fun_t) {
	if Bdc_IsComplement(f) != 0 {
		fmt.Fprint(w, "!")
	}
	n := Bdc_Regular(f)
	switch n.Type {
	case BDC_TYPE_CONST1:
		fmt.Fprint(w, "1")
	case BDC_TYPE_PI:
		fmt.Fprint(w, n.id-1)
	default:
		fmt.Fprint(w, "(")
		writeNode(w, p, n.Fan0)
		fmt.Fprint(w, " & ")
		writeNode(w, p, n.Fan1)
		fmt.Fprint(w, ")")
	}
}
func Bdc_ManDecPrint(p *Bdc_Man_t) {
	fmt.Fprint(os.Stdout, "F = ")
	writeNode(os.Stdout, p, p.root)
	fmt.Fprintln(os.Stdout)
}
func Bdc_ManDecPrintSimple(p *Bdc_Man_t) {
	for _, f := range p.nodes {
		fmt.Printf("%d : type %d\n", f.id, f.Type)
	}
	fmt.Printf("Root = %d complement %d\n", Bdc_Regular(p.root).id, Bdc_IsComplement(p.root))
}
func Bdc_ManDecPrint_rec(p *Bdc_Man_t, f *Bdc_Fun_t) { writeNode(os.Stdout, p, f) }
func Bdc_ManDecomposeTest(t uint32, n int32) {
	p := Bdc_ManAlloc(&Bdc_Par_t{NVarsMax: 8})
	fmt.Printf("Nodes = %d\n", Bdc_ManDecompose(p, []uint32{t}, nil, n, nil, 1000))
	Bdc_ManFree(p)
}
