package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/util"
	ceccore "github.com/glycerine/abcgo/proof/cec/core"
)

// allOutputConjunction adds virtual ANDs for the legacy fAndOuts SAT mode.
// It borrows the graph without mutating object storage, numbering or outputs.
type allOutputConjunction struct {
	base *Gia_Man_t
	ands [][2]int32
	root int32
}

func (n *allOutputConjunction) ObjectCount() int        { return n.base.ObjectCount() + len(n.ands) }
func (n *allOutputConjunction) InputLiterals() []int32  { return n.base.InputLiterals() }
func (n *allOutputConjunction) OutputLiterals() []int32 { return []int32{n.root} }
func (n *allOutputConjunction) RegisterCount() int      { return n.base.RegisterCount() }
func (n *allOutputConjunction) Gate(id int32) (a, b, c int32, kind int) {
	if int(id) < n.base.ObjectCount() {
		return n.base.Gate(id)
	}
	index := int(id) - n.base.ObjectCount()
	if index < 0 || index >= len(n.ands) {
		return 0, 0, 0, -1
	}
	pair := n.ands[index]
	return pair[0], pair[1], 0, 2
}
func giaConjunction(p *Gia_Man_t) *allOutputConjunction {
	n := &allOutputConjunction{base: p, root: 1}
	for _, output := range p.OutputLiterals() {
		if output == 1 {
			continue
		}
		if output == 0 {
			n.root = 0
			break
		}
		if n.root == 1 {
			n.root = output
			continue
		}
		id := int32(n.ObjectCount())
		n.ands = append(n.ands, [2]int32{n.root, output})
		n.root = id << 1
	}
	return n
}
func giaSaveCounterexample(p *Gia_Man_t, witness *ceccore.Counterexample) {
	p.PCexComb = nil
	if witness == nil {
		return
	}
	cex := util.Abc_CexAlloc(p.nRegs, Gia_ManPiNum(p), 1)
	cex.IPo = int32(witness.Output)
	set := func(index int, value bool) {
		if value {
			cex.PData[index>>5] |= 1 << uint(index&31)
		}
	}
	for i := 0; i < int(p.nRegs); i++ {
		set(i, witness.Inputs[int(Gia_ManPiNum(p))+i])
	}
	for i := 0; i < int(Gia_ManPiNum(p)); i++ {
		set(int(p.nRegs)+i, witness.Inputs[i])
	}
	p.PCexComb = cex
}

// Gia_ManSolveSatResult solves the conjunction of all outputs, as the original
// Fra_FraigSat fAndOuts call did. Sequential state is free only when explicitly
// requested through TreatRegistersAsInputs; the legacy entry point rejects it.
func Gia_ManSolveSatResult(p *Gia_Man_t, provided ...ceccore.Options) (ceccore.Result, error) {
	p.PCexComb = nil
	options := ceccore.DefaultSATOptions()
	options.ConflictLimit = 10000000
	if len(provided) > 1 {
		return ceccore.Result{}, fmt.Errorf("gia: expected at most one SAT option set")
	}
	if len(provided) == 1 {
		options = provided[0]
	}
	network := giaConjunction(p)
	manager, err := ceccore.NewManager(network, options)
	if err != nil {
		return ceccore.Result{}, err
	}
	result, err := manager.CheckLiteral(network.root)
	if err != nil || result.Status != ceccore.Different {
		return result, err
	}
	if result.Counterexample == nil {
		return ceccore.Result{}, fmt.Errorf("gia: SAT result has no witness")
	}
	outputs, err := p.Evaluate(result.Counterexample.Inputs)
	if err != nil {
		return ceccore.Result{}, err
	}
	for _, value := range outputs {
		if !value {
			return ceccore.Result{}, fmt.Errorf("gia: SAT witness does not satisfy every output")
		}
	}

	values, err := manager.Model().Values(result.Counterexample.Inputs)
	if err != nil {
		return ceccore.Result{}, err
	}
	for id, obj := range p.pObjs {
		obj.fMark0 = uint32(giaBool(values[id]))
	}
	if len(outputs) > 0 {
		result.Counterexample.Output = 0
	}
	giaSaveCounterexample(p, result.Counterexample)
	return result, nil
}

func Gia_ManSolveSat(p *Gia_Man_t) int32 {
	result, err := Gia_ManSolveSatResult(p)
	if err != nil {
		fmt.Printf("GIA SAT failed: %v\n", err)
		return -1
	}
	return int32(result.Status)
}

func Cec_ManVerifyTwo(left, right *Gia_Man_t, verbose int32) int32 {
	left.PCexComb = nil
	result, err := ceccore.VerifyTwo(left, right)
	if err != nil {
		if verbose != 0 {
			fmt.Printf("GIA equivalence check failed: %v\n", err)
		}
		return -1
	}
	if result.Status == ceccore.Different {
		giaSaveCounterexample(left, result.Counterexample)
	}
	if verbose != 0 {
		switch result.Status {
		case ceccore.Equivalent:
			fmt.Println("Networks are equivalent.")
		case ceccore.Different:
			fmt.Println("Networks are not equivalent.")
		default:
			fmt.Println("Equivalence is undecided.")
		}
	}
	return int32(result.Status)
}
