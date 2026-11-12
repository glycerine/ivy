package fsim

import (
	"encoding/binary"
	"fmt"
	"github.com/glycerine/abcgo/aig/aig"
)

func Fsim_ManFrontFindNext(p *Manager, used []bool) (int, error) {
	for tries := 0; tries < len(used); tries++ {
		if p.nextNumber%len(used) == 0 {
			p.nextNumber++
		}
		slot := p.nextNumber % len(used)
		if !used[slot] {
			used[slot] = true
			return p.nextNumber, nil
		}
		p.nextNumber++
	}
	return 0, fmt.Errorf("fsim: frontier has no free slot")
}
func Fsim_ManFront(p *Manager, compressed bool) error {
	if p.Graph == nil {
		return fmt.Errorf("fsim: deleted manager")
	}
	used := make([]bool, p.Frontier)
	ids := make([]int, len(p.ops))
	refs := append([]int(nil), p.references...)
	p.nextNumber = 1
	p.constSlot = 0
	p.program = nil
	p.encoded = nil
	p.compressed = compressed
	if refs[1] > 0 {
		id, err := Fsim_ManFrontFindNext(p, used)
		if err != nil {
			return err
		}
		ids[1] = id
		p.constSlot = id
	}
	for _, original := range p.ops[2:] {
		o := original
		if o.Kind != aig.Output {
			node, err := Fsim_ManFrontFindNext(p, used)
			if err != nil {
				return err
			}
			o.Node = node
		} else {
			o.Node = 0
		}
		o.Fan0, o.Fan1 = 0, 0
		fans := fsimFanins(original)
		for i, lit := range fans {
			if ids[lit>>1] == 0 {
				return fmt.Errorf("fsim: expired frontier fanin")
			}
			copied := ids[lit>>1]*2 + (lit & 1)
			if i == 0 {
				o.Fan0 = copied
			} else {
				o.Fan1 = copied
			}
		}
		if o.Kind == aig.And || o.Kind == aig.ExclusiveOr {
			if o.Fan0 < o.Fan1 {
				o.Fan0, o.Fan1 = o.Fan1, o.Fan0
			}
		}
		p.program = append(p.program, o)
		for _, lit := range fans {
			refs[lit>>1]--
			if refs[lit>>1] == 0 {
				used[ids[lit>>1]%p.Frontier] = false
				ids[lit>>1] = 0
			}
		}
		if o.Kind != aig.Output {
			ids[original.Node] = o.Node
			if refs[original.Node] == 0 {
				used[o.Node%p.Frontier] = false
				ids[original.Node] = 0
			}
		}
	}
	if compressed {
		var previous int
		for _, o := range p.program {
			p.encoded = appendOperation(p.encoded, o, &previous)
		}
		p.program = nil
	}
	return nil
}

// The source delta/varint representation is extended with explicit operation
// kinds so native XORs and buffers preserve their semantics.
func appendOperation(out []byte, o Fsim_Obj_t, previous *int) []byte {
	tag := uint64(o.Kind)
	if o.Kind == aig.Output {
		out = binary.AppendUvarint(out, tag)
		return binary.AppendUvarint(out, uint64((*previous*2|1)-o.Fan0))
	}
	out = binary.AppendUvarint(out, uint64(o.Node-*previous)<<3|tag)
	*previous = o.Node
	if o.Kind != aig.Input {
		out = binary.AppendUvarint(out, uint64(o.Node*2-o.Fan0))
	}
	if o.Kind == aig.And || o.Kind == aig.ExclusiveOr {
		out = binary.AppendUvarint(out, uint64(o.Fan0-o.Fan1))
	}
	return out
}
func decodeOperation(data []byte, offset *int, previous *int) (Fsim_Obj_t, error) {
	read := func() (int, error) {
		if *offset >= len(data) {
			return 0, fmt.Errorf("fsim: truncated operation")
		}
		value, n := binary.Uvarint(data[*offset:])
		if n <= 0 || value > uint64(^uint(0)>>1) {
			return 0, fmt.Errorf("fsim: invalid operation varint")
		}
		*offset += n
		return int(value), nil
	}
	header, err := read()
	if err != nil {
		return Fsim_Obj_t{}, err
	}
	o := Fsim_Obj_t{Kind: aig.NodeKind(header & 7)}
	switch o.Kind {
	case aig.Input, aig.And, aig.ExclusiveOr, aig.Buffer:
		o.Node = *previous + (header >> 3)
		*previous = o.Node
		if o.Kind != aig.Input {
			delta, e := read()
			if e != nil {
				return o, e
			}
			o.Fan0 = o.Node*2 - delta
		}
		if o.Kind == aig.And || o.Kind == aig.ExclusiveOr {
			delta, e := read()
			if e != nil {
				return o, e
			}
			o.Fan1 = o.Fan0 - delta
		}
	case aig.Output:
		delta, e := read()
		if e != nil {
			return o, e
		}
		o.Fan0 = (*previous*2 | 1) - delta
	default:
		return o, fmt.Errorf("fsim: invalid operation kind")
	}
	if o.Node < 0 || o.Fan0 < 0 || o.Fan1 < 0 {
		return o, fmt.Errorf("fsim: negative decoded frontier index")
	}
	return o, nil
}
func (p *Manager) forEachOperation(visit func(Fsim_Obj_t)) error {
	if !p.compressed {
		for _, o := range p.program {
			visit(o)
		}
		return nil
	}
	offset, previous := 0, 0
	for i := 2; i < len(p.ops); i++ {
		o, err := decodeOperation(p.encoded, &offset, &previous)
		if err != nil {
			return err
		}
		o.Index, o.Source = p.ops[i].Index, p.ops[i].Source
		visit(o)
	}
	if offset != len(p.encoded) {
		return fmt.Errorf("fsim: extra compressed operations")
	}
	return nil
}
func Fsim_ManVerifyFront(p *Manager) error {
	owners := make([]int, p.Frontier)
	if p.constSlot > 0 {
		owners[p.constSlot%p.Frontier] = 1
	}
	index := 2
	var failure error
	err := p.forEachOperation(func(o Fsim_Obj_t) {
		original := p.ops[index]
		for j, lit := range fsimFanins(o) {
			want := fsimFanins(original)[j]
			if owners[(lit>>1)%p.Frontier] != want>>1 || (lit&1) != (want&1) {
				failure = fmt.Errorf("fsim: frontier mismatch at operation %d", index)
			}
		}
		if o.Kind != aig.Output {
			owners[o.Node%p.Frontier] = original.Node
		}
		index++
	})
	if err != nil {
		return err
	}
	return failure
}
