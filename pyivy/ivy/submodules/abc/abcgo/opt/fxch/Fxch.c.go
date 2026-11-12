// Port of ABC FXCH, Bruno Schmitt, UFRGS (2016); original Fxch.c.
package fxch

import (
	"fmt"
	"io"
	"sort"
)

func Fxch_CubesGruping(p *Fxch_Man_t) {
	indices := make(map[string]int)
	var cubes Cubes
	var outputs []map[int32]bool
	for _, c := range p.Cubes {
		key := divKey(c[1:])
		if i, ok := indices[key]; ok {
			outputs[i][c[0]] = true
		} else {
			indices[key] = len(cubes)
			row := copyInts(c)
			owner := row[0]
			row[0] = 0
			cubes = append(cubes, row)
			outputs = append(outputs, map[int32]bool{owner: true})
		}
	}
	p.Cubes = cubes
	p.outputs = outputs
}
func Fxch_CubesUnGruping(p *Fxch_Man_t) {
	var out Cubes
	for i, c := range p.Cubes {
		if len(c) == 0 {
			continue
		}
		if c[0] != 0 {
			out = append(out, copyInts(c))
			continue
		}
		ids := make([]int, 0, len(p.outputs[i]))
		for o := range p.outputs[i] {
			ids = append(ids, int(o))
		}
		sort.Ints(ids)
		for _, o := range ids {
			row := copyInts(c)
			row[0] = int32(o)
			out = append(out, row)
		}
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i][0] < out[j][0] })
	p.Cubes = out
}

// Extract mutates cubes only on success. Zero maximum means unlimited, as in C.
func Extract(cubes *Cubes, objMax, maxDiv int32, log io.Writer) (int32, error) {
	if cubes == nil || objMax < 0 || maxDiv < 0 {
		return 0, fmt.Errorf("invalid FXCH arguments")
	}
	work := make(Cubes, len(*cubes))
	for i, c := range *cubes {
		if len(c) == 0 || c[0] < 0 || c[0] >= objMax {
			return 0, fmt.Errorf("invalid FXCH owner at row %d", i)
		}
		work[i] = copyInts(c)
		last := int32(-1)
		for _, l := range c[1:] {
			if l < 0 || l>>1 >= objMax || l <= last || last >= 0 && l>>1 == last>>1 {
				return 0, fmt.Errorf("invalid/unsorted FXCH literal at row %d", i)
			}
			last = l
		}
	}
	if err := validateDAG(work); err != nil {
		return 0, err
	}
	if len(work) == 0 {
		return 0, nil
	}
	p := Fxch_ManAlloc(work)
	p.Log = log
	Fxch_CubesGruping(p)
	Fxch_ManMapLiteralsIntoCubes(p, objMax)
	Fxch_ManGenerateLitHashKeys(p)
	Fxch_ManComputeLevel(p)
	Fxch_ManDivCreate(p)
	for maxDiv == 0 || p.Extracted < maxDiv {
		best := int32(-1)
		weight := float32(0)
		for i, d := range p.divisors {
			if d.weight > weight {
				best = int32(i)
				weight = d.weight
			}
		}
		if best < 0 {
			break
		}
		Fxch_ManUpdate(p, best)
	}
	count := p.Extracted
	Fxch_CubesUnGruping(p)
	*cubes = p.Cubes
	return count, nil
}
func Fxch_FastExtract(cubes *Cubes, objMax, maxDiv, verbose, veryVerbose int32) int32 {
	var log io.Writer = io.Discard
	count, err := Extract(cubes, objMax, maxDiv, log)
	if err != nil {
		return 0
	}
	if verbose != 0 {
		fmt.Printf("FXCH extracted %d divisors\n", count)
	}
	return 1
}

// FXCHNetwork adapters translate native SOP graph objects into owner/literal rows.
type FXCHNetwork interface {
	FXCHCubes() (Cubes, int32, error)
	ReplaceFXCHCubes(Cubes) error
}

func Abc_NtkFxchPerform(n FXCHNetwork, maxDiv, verbose, veryVerbose int32) int32 {
	c, limit, err := n.FXCHCubes()
	if err != nil {
		return 0
	}
	if Fxch_FastExtract(&c, limit, maxDiv, verbose, veryVerbose) == 0 {
		return 0
	}
	if n.ReplaceFXCHCubes(c) != nil {
		return 0
	}
	return 1
}
func Abc_NtkFxRetrieve(n FXCHNetwork) (Cubes, error) { c, _, e := n.FXCHCubes(); return c, e }
func Abc_NtkFxInsert(n FXCHNetwork, c Cubes) error   { return n.ReplaceFXCHCubes(c) }
func Abc_NtkFxCheck(n FXCHNetwork) int32             { _, _, e := n.FXCHCubes(); return boolInt(e == nil) }

// Validate ownership dependencies independently of numeric owner order.
func validateDAG(c Cubes) error {
	edges := make(map[int32][]int32)
	for _, row := range c {
		if _, ok := edges[row[0]]; !ok {
			edges[row[0]] = nil
		}
		for _, l := range row[1:] {
			edges[row[0]] = append(edges[row[0]], l>>1)
		}
	}
	state := make(map[int32]byte)
	var visit func(int32) error
	visit = func(id int32) error {
		if state[id] == 1 {
			return fmt.Errorf("FXCH combinational cycle at node %d", id)
		}
		if state[id] == 2 {
			return nil
		}
		state[id] = 1
		for _, dep := range edges[id] {
			if _, ok := edges[dep]; ok {
				if err := visit(dep); err != nil {
					return err
				}
			}
		}
		state[id] = 2
		return nil
	}
	for id := range edges {
		if err := visit(id); err != nil {
			return err
		}
	}
	return nil
}
