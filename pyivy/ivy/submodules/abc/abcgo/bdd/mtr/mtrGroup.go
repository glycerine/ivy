package mtr

import (
	"bufio"
	"fmt"
	"io"
	"os"
)

func Mtr_InitGroupTree(lower, size int) *MtrNode {
	if lower < 0 || size < 1 || uint64(lower)+uint64(size)-1 > uint64(MTR_MAXHIGH) {
		return nil
	}
	return &MtrNode{Low: uint32(lower), Size: uint32(size)}
}
func Mtr_MakeGroup(root *MtrNode, low, size, flags uint32) *MtrNode {
	if !contains(root, low, size) {
		return nil
	}
	if root.Low == low && root.Size == size {
		root.Flags = flags
		return root
	}
	end := uint64(low) + uint64(size)
	var previous *MtrNode
	first := root.Child
	for first != nil && uint64(first.Low)+uint64(first.Size) <= uint64(low) {
		previous = first
		first = first.Younger
	}
	if first != nil && contains(first, low, size) {
		return Mtr_MakeGroup(first, low, size, flags)
	}
	// All overlapping children must be wholly contained. Validate before mutation.
	after := first
	for after != nil && uint64(after.Low) < end {
		if after.Low < low || uint64(after.Low)+uint64(after.Size) > end {
			return nil
		}
		after = after.Younger
	}
	n := &MtrNode{Low: low, Size: size, Flags: flags, Parent: root, Elder: previous, Younger: after}
	if previous == nil {
		root.Child = n
	} else {
		previous.Younger = n
	}
	if after != nil {
		after.Elder = n
	}
	if first != after {
		n.Child = first
		first.Elder = nil
		for c := first; ; c = c.Younger {
			c.Parent = n
			if c.Younger == after {
				c.Younger = nil
				break
			}
		}
	}
	return n
}
func Mtr_DissolveGroup(g *MtrNode) *MtrNode {
	if g == nil || g.Parent == nil || g.Flags&MTR_TERMINAL != 0 || g.Child == nil {
		return nil
	}
	p := g.Parent
	last := g.Child
	for {
		last.Parent = p
		if last.Younger == nil {
			break
		}
		last = last.Younger
	}
	last.Younger = g.Younger
	if g.Younger != nil {
		g.Younger.Elder = last
	}
	g.Child.Elder = g.Elder
	if g.Elder == nil {
		p.Child = g.Child
	} else {
		g.Elder.Younger = g.Child
	}
	Mtr_DeallocNode(g)
	return p
}
func Mtr_FindGroup(root *MtrNode, low, size uint32) *MtrNode {
	if !contains(root, low, size) {
		return nil
	}
	if root.Low == low && root.Size == size {
		return root
	}
	for c := root.Child; c != nil; c = c.Younger {
		if contains(c, low, size) {
			return Mtr_FindGroup(c, low, size)
		}
	}
	return nil
}
func Mtr_SwapGroups(a, b *MtrNode) int {
	if a == nil || b == nil {
		return 0
	}
	if b.Younger == a {
		a, b = b, a
	}
	if a.Younger != b || a.Parent == nil || a.Parent != b.Parent {
		return 0
	}
	sa, sb := int64(a.Size), int64(b.Size)
	if !shiftValid(a, sb) || !shiftValid(b, -sa) {
		return 0
	}
	p := a.Parent
	if a.Elder == nil {
		p.Child = b
	} else {
		a.Elder.Younger = b
	}
	if b.Younger != nil {
		b.Younger.Elder = a
	}
	a.Younger = b.Younger
	b.Elder = a.Elder
	a.Elder = b
	b.Younger = a
	shiftTree(a, sb)
	shiftTree(b, -sa)
	return 1
}

// WriteGroups writes the original parenthesized interval representation and
// validates parent/sibling/range invariants while traversing it.
func WriteGroups(w io.Writer, root *MtrNode) error {
	if root == nil {
		return fmt.Errorf("mtr: nil root")
	}
	var write func(*MtrNode) error
	write = func(n *MtrNode) error {
		if n.Size == 0 || n.Flags & ^uint32(MTR_TERMINAL|MTR_SOFT|MTR_FIXED|MTR_NEWNODE) != 0 {
			return fmt.Errorf("mtr: invalid node")
		}
		if n.Younger != nil && n.Younger.Elder != n || n.Elder != nil && n.Elder.Younger != n {
			return fmt.Errorf("mtr: invalid sibling links")
		}
		if _, err := fmt.Fprintf(w, "(%d", n.Low); err != nil {
			return err
		}
		if n.Flags&MTR_TERMINAL != 0 || n.Child == nil {
			if _, err := io.WriteString(w, ","); err != nil {
				return err
			}
		} else {
			var previous *MtrNode
			for c := n.Child; c != nil; c = c.Younger {
				if c.Parent != n || c.Elder != previous || !contains(n, c.Low, c.Size) || previous != nil && uint64(previous.Low)+uint64(previous.Size) > uint64(c.Low) {
					return fmt.Errorf("mtr: invalid child interval")
				}
				if err := write(c); err != nil {
					return err
				}
				previous = c
			}
		}
		if _, err := fmt.Fprint(w, uint64(n.Low)+uint64(n.Size)-1); err != nil {
			return err
		}
		if n.Flags != 0 {
			if _, err := io.WriteString(w, "|"); err != nil {
				return err
			}
			for _, f := range []struct {
				bit  uint32
				text string
			}{{MTR_FIXED, "F"}, {MTR_NEWNODE, "N"}, {MTR_SOFT, "S"}} {
				if n.Flags&f.bit != 0 {
					if _, err := io.WriteString(w, f.text); err != nil {
						return err
					}
				}
			}
		}
		_, err := io.WriteString(w, ")")
		return err
	}
	if err := write(root); err != nil {
		return err
	}
	if root.Parent == nil {
		_, err := io.WriteString(w, "\n")
		return err
	}
	return nil
}
func Mtr_PrintGroups(root *MtrNode, silent int) {
	w := io.Writer(os.Stdout)
	if silent != 0 {
		w = io.Discard
	}
	if err := WriteGroups(w, root); err != nil {
		panic(err)
	}
}

// ReadGroups reads low/size/flag triples and rejects crossing or invalid ranges.
func ReadGroups(r io.Reader, nleaves int) (*MtrNode, error) {
	root := Mtr_InitGroupTree(0, nleaves)
	if root == nil {
		return nil, fmt.Errorf("mtr: invalid leaf count")
	}
	scan := bufio.NewReader(r)
	for {
		var low, size int
		var attributes string
		n, err := fmt.Fscan(scan, &low, &size, &attributes)
		if err == io.EOF && n == 0 {
			return root, nil
		}
		if err != nil {
			return nil, fmt.Errorf("mtr: incomplete group: %w", err)
		}
		if low < 0 || size < 1 || uint64(low)+uint64(size) > uint64(nleaves) || len(attributes) > 32 {
			return nil, fmt.Errorf("mtr: invalid group range")
		}
		var flags uint32
		for _, c := range attributes {
			switch c {
			case 'D':
			case 'F':
				flags |= MTR_FIXED
			case 'N':
				flags |= MTR_NEWNODE
			case 'S':
				flags |= MTR_SOFT
			case 'T':
				flags |= MTR_TERMINAL
			default:
				return nil, fmt.Errorf("mtr: invalid flag %q", c)
			}
		}
		if Mtr_MakeGroup(root, uint32(low), uint32(size), flags) == nil {
			return nil, fmt.Errorf("mtr: crossing groups")
		}
	}
}
func Mtr_ReadGroups(r io.Reader, nleaves int) *MtrNode {
	root, err := ReadGroups(r, nleaves)
	if err != nil {
		return nil
	}
	return root
}
