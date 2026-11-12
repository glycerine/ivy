package io

import (
	"fmt"
	"os"
	"strings"

	"github.com/glycerine/abcgo/base/abc"
)

// Mop_Man_t owns multi-output PLA cubes. Its Go slices replace the original
// parallel truth-table arrays and free-list bookkeeping.
type Mop_Man_t struct {
	nIns        int32
	nOuts       int32
	inputNames  []string
	outputNames []string
	cubes       []mopCube
}

type mopCube struct {
	inputs  string
	outputs string
}

func Mop_ManAlloc(nIns, nOuts, nCubes int32) *Mop_Man_t {
	if nIns < 0 || nOuts < 0 || nCubes < 0 {
		panic("io: negative MOPLA dimensions")
	}
	return &Mop_Man_t{nIns: nIns, nOuts: nOuts, cubes: make([]mopCube, 0, nCubes)}
}

func Mop_ManStop(p *Mop_Man_t) {
	if p != nil {
		p.cubes = nil
		p.inputNames = nil
		p.outputNames = nil
	}
}

func Mop_ManLoadFile(filename string) string {
	contents, err := os.ReadFile(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Mop_ManLoadFile: %v\n", err)
		return ""
	}
	return string(contents)
}

func Mop_ManReadParams(buffer string, nIns, nOuts *int32) int32 {
	doc, err := parsePLADocument("<buffer>", buffer)
	if err != nil {
		return -1
	}
	if nIns != nil {
		*nIns = int32(doc.nInputs)
	}
	if nOuts != nil {
		*nOuts = int32(doc.nOutputs)
	}
	return int32(len(doc.rows))
}

func Mop_ManRead(filename string) *Mop_Man_t {
	p, err := Mop_ManReadWithError(filename)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return nil
	}
	return p
}

func Mop_ManReadWithError(filename string) (*Mop_Man_t, error) {
	contents, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	doc, err := parsePLADocument(filename, string(contents))
	if err != nil {
		return nil, err
	}
	p := Mop_ManAlloc(int32(doc.nInputs), int32(doc.nOutputs), int32(len(doc.rows)))
	p.inputNames = append([]string(nil), doc.inputNames...)
	p.outputNames = append([]string(nil), doc.outputNames...)
	for _, row := range doc.rows {
		outputs := []byte(row.outputs)
		for i := range outputs {
			if outputs[i] == '-' {
				outputs[i] = '0'
			}
		}
		p.cubes = append(p.cubes, mopCube{inputs: row.inputs, outputs: string(outputs)})
	}
	return p, nil
}

func Mop_ManPrintOne(p *Mop_Man_t, iCube int32) {
	if p == nil || iCube < 0 || int(iCube) >= len(p.cubes) {
		return
	}
	cube := p.cubes[iCube]
	fmt.Printf("%s %s\n", cube.inputs, cube.outputs)
}

func Mop_ManPrint(p *Mop_Man_t) {
	if p == nil {
		return
	}
	fmt.Printf(".i %d\n.o %d\n", p.nIns, p.nOuts)
	for i := range p.cubes {
		Mop_ManPrintOne(p, int32(i))
	}
	fmt.Println(".e")
}

func Mop_ManRemoveEmpty(p *Mop_Man_t) {
	if p == nil {
		return
	}
	kept := p.cubes[:0]
	for _, cube := range p.cubes {
		if strings.ContainsRune(cube.outputs, '1') {
			kept = append(kept, cube)
		}
	}
	p.cubes = kept
}

func Mop_ManCountOutputLits(p *Mop_Man_t) int32 {
	if p == nil {
		return 0
	}
	var count int32
	for _, cube := range p.cubes {
		count += int32(strings.Count(cube.outputs, "1"))
	}
	return count
}

func Mop_ManReduce(p *Mop_Man_t) {
	mopReduce(p, false)
}

func Mop_ManReduce2(p *Mop_Man_t) {
	mopReduce(p, true)
}

func mopReduce(p *Mop_Man_t, merge bool) {
	if p == nil {
		return
	}
	before := len(p.cubes)
	Mop_ManRemoveEmpty(p)
	changed := true
	for changed {
		changed = false
		if mopMergeIdenticalInputs(p) {
			changed = true
		}
		if mopRemoveContained(p) {
			changed = true
		}
		if merge && mopMergeDistanceOne(p) {
			changed = true
		}
	}
	fmt.Printf("MOPLA cubes: %d -> %d; output literals: %d\n", before, len(p.cubes), Mop_ManCountOutputLits(p))
}

func mopMergeIdenticalInputs(p *Mop_Man_t) bool {
	merged := make([]mopCube, 0, len(p.cubes))
	index := make(map[string]int, len(p.cubes))
	changed := false
	for _, cube := range p.cubes {
		if at, ok := index[cube.inputs]; ok {
			outputs := []byte(merged[at].outputs)
			for i := range outputs {
				if cube.outputs[i] == '1' {
					outputs[i] = '1'
				}
			}
			merged[at].outputs = string(outputs)
			changed = true
			continue
		}
		index[cube.inputs] = len(merged)
		merged = append(merged, cube)
	}
	p.cubes = merged
	return changed
}

func mopRemoveContained(p *Mop_Man_t) bool {
	removed := make([]bool, len(p.cubes))
	changed := false
	for i := range p.cubes {
		for j := range p.cubes {
			if i == j || removed[j] || !plaCubeContains(p.cubes[i].inputs, p.cubes[j].inputs) {
				continue
			}
			outputs := []byte(p.cubes[j].outputs)
			for k := range outputs {
				if p.cubes[i].outputs[k] == '1' {
					if outputs[k] != '0' {
						changed = true
					}
					outputs[k] = '0'
				}
			}
			p.cubes[j].outputs = string(outputs)
			if !strings.ContainsRune(p.cubes[j].outputs, '1') {
				removed[j] = true
			}
		}
	}
	result := p.cubes[:0]
	for i, cube := range p.cubes {
		if !removed[i] {
			result = append(result, cube)
		}
	}
	changed = changed || len(result) != len(p.cubes)
	p.cubes = result
	return changed
}

func mopMergeDistanceOne(p *Mop_Man_t) bool {
	used := make([]bool, len(p.cubes))
	result := make([]mopCube, 0, len(p.cubes))
	changed := false
	for i, a := range p.cubes {
		if used[i] {
			continue
		}
		for j := i + 1; j < len(p.cubes); j++ {
			if used[j] || a.outputs != p.cubes[j].outputs {
				continue
			}
			if input, ok := mergePLACubes(a.inputs, p.cubes[j].inputs); ok {
				result = append(result, mopCube{inputs: input, outputs: a.outputs})
				used[i], used[j] = true, true
				changed = true
				break
			}
		}
		if !used[i] {
			used[i] = true
			result = append(result, a)
		}
	}
	p.cubes = result
	return changed
}

func Mop_ManDerive(p *Mop_Man_t, filename string) *Abc_Ntk_t {
	if p == nil {
		return nil
	}
	name := p.inputNames
	if len(name) != int(p.nIns) {
		name = make([]string, p.nIns)
		for i := range name {
			name[i] = fmt.Sprintf("x%d", i)
		}
	}
	outNames := p.outputNames
	if len(outNames) != int(p.nOuts) {
		outNames = make([]string, p.nOuts)
		for i := range outNames {
			outNames[i] = fmt.Sprintf("z%d", i)
		}
	}
	network := abc.Abc_NtkAlloc(abc.ABC_NTK_LOGIC, abc.ABC_FUNC_SOP, 1)
	network.Name = filename
	network.Spec = filename
	inputs := make([]*abc.Object, p.nIns)
	for i, piName := range name {
		inputs[i] = network.NewObject(abc.ObjectPI)
		inputs[i].SetName(piName)
	}
	for output := 0; output < int(p.nOuts); output++ {
		var cover strings.Builder
		for _, cube := range p.cubes {
			if cube.outputs[output] != '1' {
				continue
			}
			cover.WriteString(cube.inputs)
			cover.WriteString(" 1\n")
		}
		data := cover.String()
		if data == "" {
			data = " 0\n"
		}
		node := network.NewObject(abc.ObjectNode)
		node.Data = data
		for _, input := range inputs {
			abc.Abc_ObjAddFanin(node, input)
		}
		po := network.NewObject(abc.ObjectPO)
		po.SetName(outNames[output])
		abc.Abc_ObjAddFanin(po, node)
	}
	return network
}

func Mop_ManTest(filename string, fMerge, fVerbose int32) *Abc_Ntk_t {
	p, err := Mop_ManReadWithError(filename)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return nil
	}
	defer Mop_ManStop(p)
	Mop_ManRemoveEmpty(p)
	before := len(p.cubes)
	if fMerge != 0 {
		mopReduce(p, true)
	} else {
		mopReduce(p, false)
	}
	if fVerbose != 0 {
		fmt.Printf("MOPLA reduced %d cubes to %d\n", before, len(p.cubes))
	}
	return Mop_ManDerive(p, filename)
}
