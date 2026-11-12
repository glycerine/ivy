// Port of ABC/MVSIS FXU (MVSIS Group, UC Berkeley, 2003).
// Original algorithm and licensing: src/opt/fxu; see repository copyright notice.
package fxu

import "io"

type Fxu_HeapSingle struct{ priorityHeap[*Fxu_Single] }

func Fxu_HeapSingleStart() *Fxu_HeapSingle                  { return new(Fxu_HeapSingle) }
func Fxu_HeapSingleStop(p *Fxu_HeapSingle)                  { p.clear() }
func Fxu_HeapSingleInsert(p *Fxu_HeapSingle, x *Fxu_Single) { p.insert(x) }
func Fxu_HeapSingleUpdate(p *Fxu_HeapSingle, x *Fxu_Single) { p.update(x) }
func Fxu_HeapSingleDelete(p *Fxu_HeapSingle, x *Fxu_Single) { p.remove(x) }
func Fxu_HeapSingleReadMax(p *Fxu_HeapSingle) *Fxu_Single   { return p.top() }
func Fxu_HeapSingleGetMax(p *Fxu_HeapSingle) *Fxu_Single    { return p.pop() }
func Fxu_HeapSingleReadMaxWeight(p *Fxu_HeapSingle) int32 {
	if len(p.items) == 0 {
		return -1
	}
	return p.items[0].Weight
}
func Fxu_HeapSingleCheck(p *Fxu_HeapSingle)                    { p.check() }
func Fxu_HeapSingleCheckOne(p *Fxu_HeapSingle, x *Fxu_Single)  { p.check() }
func Fxu_HeapSinglePrint(w io.Writer, p *Fxu_HeapSingle) error { return p.print(w) }
