package abc

import "github.com/glycerine/abcgo/misc/vec"

func Abc_ObjLevelNew(o *Object) int32 {
	level := 0
	for _, fanin := range o.fanins {
		if fanin.Object.Level > level {
			level = fanin.Object.Level
		}
	}
	if len(o.fanins) != 0 {
		level++
	}
	return int32(level)
}

func Abc_ObjReverseLevelNew(o *Object) int32 {
	level := 0
	for _, fanout := range o.fanouts {
		if fanout.reverseLevel > level {
			level = fanout.reverseLevel
		}
	}
	return int32(level + 1)
}

func Abc_ObjReverseLevel(o *Object) int32 {
	if o == nil || o.network == nil || !o.network.reverseLevelsActive {
		panic("abc: reverse levels are not active")
	}
	return int32(o.reverseLevel)
}

func Abc_ObjSetReverseLevel(o *Object, level int32) {
	if o == nil || o.network == nil || !o.network.reverseLevelsActive {
		panic("abc: reverse levels are not active")
	}
	o.reverseLevel = int(level)
}

func Abc_ObjRequiredLevel(o *Object) int32 {
	return int32(o.network.levelMax+1) - Abc_ObjReverseLevel(o)
}

// Abc_NtkReverseLevelMax returns the maximum level stored by reverse-level setup.
func Abc_NtkReverseLevelMax(n *Network) int { return n.levelMax }

func Abc_NtkStartReverseLevels(n *Network, maxLevelIncrease int32) {
	if n.reverseLevelsActive {
		panic("abc: reverse levels already active")
	}
	n.levelMax = int(Abc_NtkLevel(n) + maxLevelIncrease)
	n.reverseLevelsActive = true
	for _, o := range n.Objects() {
		o.reverseLevel = 0
	}
	for _, o := range Abc_NtkDfsReverse(n) {
		Abc_ObjSetReverseLevel(o, Abc_ObjReverseLevelNew(o))
	}
}

func Abc_NtkStopReverseLevels(n *Network) {
	if !n.reverseLevelsActive {
		panic("abc: reverse levels are not active")
	}
	for _, o := range n.Objects() {
		o.reverseLevel = 0
	}
	n.reverseLevelsActive = false
	n.levelMax = 0
}

func Abc_NtkUpdateLevel(o *Object, levels *vec.Vec_Vec_t) {
	oldLevel := o.Level
	if int32(oldLevel) == Abc_ObjLevelNew(o) {
		return
	}
	vec.Vec_VecClear(levels)
	vec.Vec_VecPush(levels, oldLevel, o)
	o.MarkA = true
	for level := oldLevel; level < vec.Vec_VecSize(levels); level++ {
		row := vec.Vec_VecEntry(levels, level)
		for i := 0; i < vec.Vec_PtrSize(row); i++ {
			current := vec.Vec_PtrEntry(row, i).(*Object)
			current.MarkA = false
			if current.Level != level {
				panic("abc: stale level in incremental update")
			}
			current.Level = int(Abc_ObjLevelNew(current))
			if current.Level == level {
				continue
			}
			for _, fanout := range current.fanouts {
				if fanout.IsCO() || fanout.MarkA {
					continue
				}
				if fanout.Level < level {
					panic("abc: fanout level precedes updated node")
				}
				vec.Vec_VecPush(levels, fanout.Level, fanout)
				fanout.MarkA = true
			}
		}
	}
}

func Abc_NtkUpdateReverseLevel(o *Object, levels *vec.Vec_Vec_t) {
	if !o.network.reverseLevelsActive {
		panic("abc: reverse levels are not active")
	}
	oldLevel := o.reverseLevel
	if int32(oldLevel) == Abc_ObjReverseLevelNew(o) {
		return
	}
	vec.Vec_VecClear(levels)
	vec.Vec_VecPush(levels, oldLevel, o)
	o.MarkA = true
	for level := oldLevel; level < vec.Vec_VecSize(levels); level++ {
		row := vec.Vec_VecEntry(levels, level)
		for i := 0; i < vec.Vec_PtrSize(row); i++ {
			current := vec.Vec_PtrEntry(row, i).(*Object)
			current.MarkA = false
			old := current.reverseLevel
			if old != level {
				panic("abc: stale reverse level in incremental update")
			}
			current.reverseLevel = int(Abc_ObjReverseLevelNew(current))
			if current.reverseLevel == level {
				continue
			}
			for _, fanin := range current.fanins {
				p := fanin.Object
				if p.IsCI() || p.MarkA {
					continue
				}
				if p.reverseLevel < level {
					panic("abc: fanin reverse level precedes updated node")
				}
				vec.Vec_VecPush(levels, p.reverseLevel, p)
				p.MarkA = true
			}
		}
	}
}

func Abc_NtkUpdate(old, replacement *Object, levels *vec.Vec_Vec_t) {
	replacement.Level = old.Level
	Abc_ObjReplace(old, replacement)
	Abc_NtkUpdateLevel(replacement, levels)
	replacement.reverseLevel = 0
	Abc_NtkUpdateReverseLevel(replacement, levels)
}
