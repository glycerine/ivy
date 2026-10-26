package abc

// Port of the stable-ID-safe object mutation paths in src/aig/aig/aigObj.c.

import "fmt"

type AigObjectOptions struct {
	MaxObjects int // zero: one million deleted objects
	Stop       func() bool
}

type aigObjectBudget struct {
	options AigObjectOptions
	used    int
}

func newAigObjectBudget(options AigObjectOptions) (*aigObjectBudget, error) {
	if options.MaxObjects < 0 {
		return nil, ErrDimension
	}
	if options.MaxObjects == 0 {
		options.MaxObjects = 1 << 20
	}
	return &aigObjectBudget{options: options}, nil
}

func (b *aigObjectBudget) step() error {
	if b.options.Stop != nil && b.options.Stop() {
		return ErrCancelled
	}
	if b.used == b.options.MaxObjects {
		return ErrLimit
	}
	b.used++
	return nil
}

func (g *AigManager) choiceContains(id AigID) bool {
	if g.equivs == nil {
		return false
	}
	if g.equivs[id] != InvalidAigID {
		return true
	}
	for _, next := range g.equivs {
		if next == id {
			return true
		}
	}
	return false
}

func (g *AigManager) deleteInternal(id AigID) ([]AigID, error) {
	object := g.objects[id]
	if (object.Type != AigAND && object.Type != AigBuffer) || object.RefCount != 0 || g.choiceContains(id) {
		return nil, ErrInvariant
	}
	if object.Type == AigAND {
		delete(g.hash, aigHashKey{object.Fanin0, object.Fanin1})
	}
	fanins := []AigID{object.Fanin0.ID()}
	if object.Type == AigAND {
		fanins = append(fanins, object.Fanin1.ID())
	}
	for _, fanin := range fanins {
		if g.objects[fanin].RefCount == 0 {
			return nil, ErrInvariant
		}
		g.objects[fanin].RefCount--
	}
	g.objects[id] = AigObject{}
	if g.equivs != nil {
		g.equivs[id] = InvalidAigID
	}
	return fanins, nil
}

func (g *AigManager) deleteConeInPlace(root AigID, budget *aigObjectBudget) (int, error) {
	stack := []AigID{root}
	deleted := 0
	for len(stack) != 0 {
		id := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if g.objects[id].Type == AigConstant1 || g.objects[id].Type == AigCI {
			continue
		}
		if err := budget.step(); err != nil {
			return 0, err
		}
		fanins, err := g.deleteInternal(id)
		if err != nil {
			return 0, err
		}
		deleted++
		// Push in reverse so fanin zero is processed first, matching C recursion.
		for index := len(fanins) - 1; index >= 0; index-- {
			fanin := fanins[index]
			object := g.objects[fanin]
			if (object.Type == AigAND || object.Type == AigBuffer) && object.RefCount == 0 {
				stack = append(stack, fanin)
			}
		}
	}
	return deleted, nil
}

// DeleteCone implements Aig_ObjDelete_rec with fFreeTop set. The root must be
// an unreferenced internal object. Deleted IDs remain permanent tombstones.
func (g *AigManager) DeleteCone(root AigID, options AigObjectOptions) (int, error) {
	if g == nil {
		return 0, ErrDimension
	}
	if err := g.Check(); err != nil {
		return 0, err
	}
	if uint64(root) >= uint64(len(g.objects)) || g.objects[root].Type == AigNone {
		return 0, ErrInvalidLiteral
	}
	object := g.objects[root]
	if object.Type != AigAND && object.Type != AigBuffer {
		return 0, ErrUnsupported
	}
	if object.RefCount != 0 {
		return 0, ErrInvariant
	}
	budget, err := newAigObjectBudget(options)
	if err != nil {
		return 0, err
	}
	work := g.Clone()
	deleted, err := work.deleteConeInPlace(root, budget)
	if err != nil {
		return 0, err
	}
	if err = work.Check(); err != nil {
		return 0, err
	}
	*g = *work
	return deleted, nil
}

func (g *AigManager) removeOutputIndex(index int) {
	copy(g.cos[index:], g.cos[index+1:])
	g.cos = g.cos[:len(g.cos)-1]
}

// DeleteOutput implements Aig_ObjDeletePo and also removes the tombstoned CO
// from Go's authoritative ordered output slice.
func (g *AigManager) DeleteOutput(output AigID) error {
	if g == nil {
		return ErrDimension
	}
	if err := g.Check(); err != nil {
		return err
	}
	index := -1
	for i, id := range g.cos {
		if id == output {
			index = i
			break
		}
	}
	if index < 0 {
		return ErrInvalidLiteral
	}
	work := g.Clone()
	object := work.objects[output]
	driver := object.Fanin0.ID()
	if work.objects[driver].RefCount == 0 {
		return ErrInvariant
	}
	work.objects[driver].RefCount--
	primaryOutputs := len(work.cos) - work.registers
	if index >= primaryOutputs {
		work.registers--
	} else if work.constraints > 0 && index >= primaryOutputs-work.constraints {
		work.constraints--
	}
	work.removeOutputIndex(index)
	work.objects[output] = AigObject{}
	if work.equivs != nil {
		work.equivs[output] = InvalidAigID
	}
	if err := work.Check(); err != nil {
		return err
	}
	*g = *work
	return nil
}

// PatchOutput implements Aig_ObjPatchFanin0 and recursively removes the old
// driver cone when it becomes unused.
func (g *AigManager) PatchOutput(output AigID, driver AigLit, options AigObjectOptions) (int, error) {
	if g == nil {
		return 0, ErrDimension
	}
	if err := g.Check(); err != nil {
		return 0, err
	}
	if uint64(output) >= uint64(len(g.objects)) || g.objects[output].Type != AigCO || !g.validSignal(driver) {
		return 0, ErrInvalidLiteral
	}
	budget, err := newAigObjectBudget(options)
	if err != nil {
		return 0, err
	}
	work := g.Clone()
	old := work.objects[output].Fanin0.ID()
	if work.objects[old].RefCount == 0 || work.objects[driver.ID()].RefCount == ^uint32(0) {
		return 0, ErrLimit
	}
	work.objects[old].RefCount--
	work.objects[driver.ID()].RefCount++
	work.objects[output].Fanin0 = driver
	work.objects[output].Level = work.objects[driver.ID()].Level
	work.objects[output].Phase = work.phase(driver)
	deleted := 0
	oldObject := work.objects[old]
	if (oldObject.Type == AigAND || oldObject.Type == AigBuffer) && oldObject.RefCount == 0 {
		deleted, err = work.deleteConeInPlace(old, budget)
		if err != nil {
			return 0, err
		}
	}
	if err = work.Check(); err != nil {
		return 0, err
	}
	*g = *work
	return deleted, nil
}

// FormatObject is the owned, non-printing replacement for Aig_ObjPrint.
func (g *AigManager) FormatObject(literal AigLit) (string, error) {
	if g == nil || !literal.Valid() || uint64(literal.ID()) >= uint64(len(g.objects)) || g.objects[literal.ID()].Type == AigNone {
		return "", ErrInvalidLiteral
	}
	object := g.objects[literal.ID()]
	prefix := ""
	if literal.Complemented() {
		prefix = "Compl "
	}
	name := ""
	switch object.Type {
	case AigConstant1:
		name = "constant 1"
	case AigCI:
		name = "PI"
	case AigCO:
		name = fmt.Sprintf("PO( %4d%s )", object.Fanin0.ID(), aigObjectPrime(object.Fanin0))
	case AigBuffer:
		name = fmt.Sprintf("BUF( %d%s )", object.Fanin0.ID(), aigObjectPrime(object.Fanin0))
	case AigAND:
		name = fmt.Sprintf("AND( %4d%s, %4d%s )", object.Fanin0.ID(), aigObjectPrime(object.Fanin0), object.Fanin1.ID(), aigObjectPrime(object.Fanin1))
	default:
		return "", ErrUnsupported
	}
	return fmt.Sprintf("%sNode %4d : %s (refs = %3d)", prefix, literal.ID(), name, object.RefCount), nil
}

func aigObjectPrime(literal AigLit) string {
	if literal.Complemented() {
		return "'"
	}
	return " "
}
