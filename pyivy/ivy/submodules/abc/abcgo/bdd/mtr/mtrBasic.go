package mtr

import (
	"fmt"
	"io"
	"os"
)

func Mtr_AllocNode() *MtrNode { return &MtrNode{} }
func Mtr_DeallocNode(n *MtrNode) {
	if n != nil {
		*n = MtrNode{}
	}
}
func Mtr_InitTree() *MtrNode { return Mtr_AllocNode() }
func Mtr_FreeTree(n *MtrNode) {
	if n == nil {
		return
	}
	if n.Flags&MTR_TERMINAL == 0 {
		Mtr_FreeTree(n.Child)
	}
	Mtr_FreeTree(n.Younger)
	Mtr_DeallocNode(n)
}
func Mtr_CopyTree(n *MtrNode, expansion int) *MtrNode {
	if n == nil || expansion < 1 {
		return nil
	}
	factor := uint64(expansion)
	if n.Size > 0 && factor > (uint64(MTR_MAXHIGH)+1)/(uint64(n.Low)+uint64(n.Size)) {
		return nil
	}
	for _, value := range []uint32{n.Low, n.Index} {
		if value > 0 && factor > uint64(MTR_MAXHIGH)/uint64(value) {
			return nil
		}
	}
	c := &MtrNode{Flags: n.Flags, Low: n.Low * uint32(expansion), Size: n.Size * uint32(expansion), Index: n.Index * uint32(expansion)}
	if n.Child != nil {
		c.Child = Mtr_CopyTree(n.Child, expansion)
		if c.Child == nil {
			return nil
		}
		for p := c.Child; p != nil; p = p.Younger {
			p.Parent = c
		}
	}
	if n.Younger != nil {
		c.Younger = Mtr_CopyTree(n.Younger, expansion)
		if c.Younger == nil {
			return nil
		}
		c.Younger.Elder = c
	}
	return c
}
func Mtr_MakeFirstChild(p, c *MtrNode) {
	c.Parent = p
	c.Elder = nil
	c.Younger = p.Child
	if p.Child != nil {
		p.Child.Elder = c
	}
	p.Child = c
}
func Mtr_MakeLastChild(p, c *MtrNode) {
	if p.Child == nil {
		Mtr_MakeFirstChild(p, c)
		return
	}
	last := p.Child
	for last.Younger != nil {
		last = last.Younger
	}
	Mtr_MakeNextSibling(last, c)
}
func Mtr_CreateFirstChild(p *MtrNode) *MtrNode {
	c := Mtr_AllocNode()
	Mtr_MakeFirstChild(p, c)
	return c
}
func Mtr_CreateLastChild(p *MtrNode) *MtrNode {
	c := Mtr_AllocNode()
	Mtr_MakeLastChild(p, c)
	return c
}
func Mtr_MakeNextSibling(a, b *MtrNode) {
	b.Parent = a.Parent
	b.Elder = a
	b.Younger = a.Younger
	if a.Younger != nil {
		a.Younger.Elder = b
	}
	a.Younger = b
}
func WriteTree(w io.Writer, n *MtrNode) error {
	for ; n != nil; n = n.Younger {
		if _, err := fmt.Fprintf(w, "%p: low=%d size=%d index=%d flags=%x parent=%p child=%p elder=%p younger=%p\n", n, n.Low, n.Size, n.Index, n.Flags, n.Parent, n.Child, n.Elder, n.Younger); err != nil {
			return err
		}
		if n.Flags&MTR_TERMINAL == 0 {
			if err := WriteTree(w, n.Child); err != nil {
				return err
			}
		}
	}
	return nil
}
func Mtr_PrintTree(n *MtrNode) { _ = WriteTree(os.Stdout, n) }
