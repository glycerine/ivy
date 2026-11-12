package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/vec"
)

type Str_Edg_t struct{ Fan, fCompl, FanDel, Copy int32 }
type Str_Mux_t struct {
	Id, Delay, Copy, nLutSize int32
	Edge                      [3]Str_Edg_t
	tree                      []Str_Mux_t
}

func Str_MuxFanin(mux *Str_Mux_t, i int32) *Str_Mux_t { return &mux.tree[mux.Edge[i].Fan] }
func Str_MuxHasFanin(mux *Str_Mux_t, i int32) int32 {
	return giaBool(mux.Edge[i].Fan > 0 && Str_MuxFanin(mux, i).Copy != -2)
}
func Str_MuxDelayPrint_rec(mux *Str_Mux_t, i int32) {
	edge := mux.Edge[i]
	if edge.Fan <= 0 {
		fmt.Printf("%d{%d}", -edge.Fan, edge.FanDel)
		return
	}
	child := Str_MuxFanin(mux, i)
	fmt.Print("[ ")
	for k := int32(0); k < 3; k++ {
		if k == 1 {
			fmt.Print("|")
		} else if k == 2 {
			fmt.Print("(")
		}
		if child.Edge[k].fCompl != 0 {
			fmt.Print("!")
		}
		Str_MuxDelayPrint_rec(child, k)
	}
	fmt.Print(") ]")
}
func Str_MuxDelayEdge_rec(mux *Str_Mux_t, i int32) int32 {
	if mux.Edge[i].Fan > 0 {
		child := Str_MuxFanin(mux, i)
		Str_MuxDelayEdge_rec(child, 0)
		Str_MuxDelayEdge_rec(child, 1)
		mux.Edge[i].FanDel = Str_Delay3(child.Edge[0].FanDel, child.Edge[1].FanDel, child.Edge[2].FanDel, child.nLutSize)
	}
	return mux.Edge[i].FanDel
}
func Str_MuxCreate(tree []Str_Mux_t, ntk *Str_Ntk_t, start, count int32, delays *Vec_Int_t, lutSize int32) {
	if count < 2 || len(tree) < int(count+1) {
		panic("gia: invalid mux tree")
	}
	clear(tree[:count+1])
	tree[0].tree = tree
	tree[0].nLutSize = lutSize
	tree[0].Edge[0].Fan = 1
	var inputs int32
	for i := int32(1); i <= count; i++ {
		node := &tree[i]
		node.tree = tree
		node.Id = i
		node.nLutSize = lutSize
		node.Delay = -1
		node.Copy = -1
		obj := Str_NtkObj(ntk, start+count-i)
		if obj.Type != STR_MUX {
			panic("gia: expected structured mux")
		}
		for k := int32(0); k < 3; k++ {
			edge := &node.Edge[k]
			edge.fCompl = Str_ObjFaninC(ntk, obj, k)
			id := Str_ObjFaninId(ntk, obj, k)
			if id >= start {
				edge.Fan = start + count - id
			} else {
				edge.Fan = -inputs
				inputs++
				edge.Copy = Str_ObjFanin(ntk, obj, k).iCopy
				edge.FanDel = vec.Vec_IntEntry(delays, int(edge.Copy>>1))
			}
		}
	}
}
func Str_MuxToGia_rec(dst *Gia_Man_t, mux *Str_Mux_t, i int32, delays *Vec_Int_t) int32 {
	edge := &mux.Edge[i]
	if edge.Fan > 0 {
		child := Str_MuxFanin(mux, i)
		a := Str_MuxToGia_rec(dst, child, 0, delays)
		b := Str_MuxToGia_rec(dst, child, 1, delays)
		if child.Edge[2].Fan > 0 || child.Edge[2].fCompl != 0 {
			panic("gia: invalid mux control")
		}
		edge.Copy = Gia_ManHashMuxReal(dst, child.Edge[2].Copy, b, a)
		Str_ObjDelay(dst, edge.Copy>>1, child.nLutSize, delays)
	}
	return edge.Copy ^ edge.fCompl
}
func Str_MuxFindPathEdge_rec(mux *Str_Mux_t, i int32, path *[]int32) int32 {
	if mux.Edge[i].Fan > 0 && Str_MuxFindPath_rec(Str_MuxFanin(mux, i), path) == 0 {
		return 0
	}
	*path = append(*path, 2*mux.Id+i)
	return 1
}
func Str_MuxFindPath_rec(mux *Str_Mux_t, path *[]int32) int32 {
	delay := max(mux.Edge[0].FanDel, mux.Edge[1].FanDel, mux.Edge[2].FanDel)
	for i := int32(0); i < 2; i++ {
		if mux.Edge[i].FanDel == delay {
			return Str_MuxFindPathEdge_rec(mux, i, path)
		}
	}
	if mux.Edge[2].FanDel == delay {
		return 0
	}
	panic("gia: invalid mux path")
}
func Str_MuxFindBranching(root *Str_Mux_t, i int32) *Str_Mux_t {
	if root.Edge[i].Fan <= 0 {
		return nil
	}
	node := Str_MuxFanin(root, i)
	for {
		if node.Edge[0].Fan <= 0 && node.Edge[1].Fan <= 0 {
			return nil
		}
		if node.Edge[0].Fan > 0 && node.Edge[1].Fan > 0 {
			return node
		}
		if node.Edge[0].Fan > 0 {
			node = Str_MuxFanin(node, 0)
		}
		if node.Edge[1].Fan > 0 {
			node = Str_MuxFanin(node, 1)
		}
	}
}
