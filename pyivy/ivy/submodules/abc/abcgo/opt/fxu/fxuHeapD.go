// Port of ABC/MVSIS FXU (MVSIS Group, UC Berkeley, 2003).
// Original algorithm and licensing: src/opt/fxu; see repository copyright notice.
package fxu

import "io"

type Fxu_HeapDouble struct{ priorityHeap[*Fxu_Double] }

func Fxu_HeapDoubleStart() *Fxu_HeapDouble                  { return new(Fxu_HeapDouble) }
func Fxu_HeapDoubleStop(p *Fxu_HeapDouble)                  { p.clear() }
func Fxu_HeapDoubleInsert(p *Fxu_HeapDouble, x *Fxu_Double) { p.insert(x) }
func Fxu_HeapDoubleUpdate(p *Fxu_HeapDouble, x *Fxu_Double) { p.update(x) }
func Fxu_HeapDoubleDelete(p *Fxu_HeapDouble, x *Fxu_Double) { p.remove(x) }
func Fxu_HeapDoubleReadMax(p *Fxu_HeapDouble) *Fxu_Double   { return p.top() }
func Fxu_HeapDoubleGetMax(p *Fxu_HeapDouble) *Fxu_Double    { return p.pop() }
func Fxu_HeapDoubleReadMaxWeight(p *Fxu_HeapDouble) int32 {
	if len(p.items) == 0 {
		return -1
	}
	return p.items[0].Weight
}
func Fxu_HeapDoubleCheck(p *Fxu_HeapDouble)                    { p.check() }
func Fxu_HeapDoubleCheckOne(p *Fxu_HeapDouble, x *Fxu_Double)  { p.check() }
func Fxu_HeapDoublePrint(w io.Writer, p *Fxu_HeapDouble) error { return p.print(w) }
