package gia

import "github.com/glycerine/abcgo/misc/vec"

func Gia_RsbAddSideInputs(p *Gia_Man_t, levels *Vec_Wec_t, window *Vec_Int_t) {
	if vec.Vec_WecSizeSize(levels) != 0 {
		panic("gia: nonempty Rsb2 level scratch")
	}
	for _, id := range vec.Vec_IntArray(window) {
		if Gia_ObjIsTravIdCurrentId(p, id) == 0 {
			panic("gia: unmarked Rsb2 window")
		}
		vec.Vec_WecPush(levels, int(Gia_ObjLevelId(p, id)), id)
	}
	for level := 0; level < vec.Vec_WecSize(levels); level++ {
		for i := 0; i < vec.Vec_IntSize(vec.Vec_WecEntry(levels, level)); i++ {
			id := vec.Vec_IntEntry(vec.Vec_WecEntry(levels, level), i)
			for f := int32(0); f < min(5, Gia_ObjFanoutNumId(p, id)); f++ {
				fan := Gia_ObjFanoutId(p, id, f)
				obj := Gia_ManObj(p, fan)
				if Gia_ObjIsAnd(obj) == 0 || Gia_ObjIsTravIdCurrentId(p, fan) != 0 {
					continue
				}
				ready := true
				for _, dep := range giaObjectFanins(p, obj) {
					if Gia_ObjIsTravIdCurrentId(p, dep) == 0 {
						ready = false
						break
					}
				}
				if ready {
					Gia_ObjSetTravIdCurrentId(p, fan)
					vec.Vec_WecPush(levels, int(Gia_ObjLevelId(p, fan)), fan)
					vec.Vec_IntPush(window, fan)
				}
			}
		}
	}
	for i := 0; i < vec.Vec_WecSize(levels); i++ {
		vec.Vec_IntClear(vec.Vec_WecEntry(levels, i))
	}
}
func Gia_WinCreateFromCut_rec(p *Gia_Man_t, id int32, window *Vec_Int_t) {
	if Gia_ObjIsTravIdCurrentId(p, id) != 0 {
		return
	}
	Gia_ObjSetTravIdCurrentId(p, id)
	obj := Gia_ManObj(p, id)
	if Gia_ObjIsAnd(obj) == 0 {
		panic("gia: Rsb2 cut does not cover cone")
	}
	for _, f := range giaObjectFanins(p, obj) {
		Gia_WinCreateFromCut_rec(p, f, window)
	}
	vec.Vec_IntPush(window, id)
}
func Gia_WinCreateFromCut(p *Gia_Man_t, pivot int32, inputs *Vec_Int_t, levels *Vec_Wec_t, window *Vec_Int_t) {
	if vec.Vec_WecSizeSize(levels) != 0 {
		panic("gia: nonempty Rsb2 level scratch")
	}
	vec.Vec_IntClear(window)
	Gia_ManIncrementTravId(p)
	for _, id := range vec.Vec_IntArray(inputs) {
		Gia_ObjSetTravIdCurrentId(p, id)
		vec.Vec_IntPush(window, id)
	}
	Gia_WinCreateFromCut_rec(p, pivot, window)
	Gia_RsbAddSideInputs(p, levels, window)
	vec.Vec_IntSort(window, 0)
}
func Gia_RsbExpandCut(p *Gia_Man_t, inputs *Vec_Int_t) int32 {
	onlyInputs := int32(0)
	for changed := true; changed; {
		changed = false
		onlyInputs = 1
		for i := 0; i < vec.Vec_IntSize(inputs); i++ {
			id := vec.Vec_IntEntry(inputs, i)
			obj := Gia_ManObj(p, id)
			if Gia_ObjIsTravIdCurrentId(p, id) == 0 {
				panic("gia: unmarked Rsb2 cut")
			}
			if Gia_ObjIsAnd(obj) == 0 {
				continue
			}
			onlyInputs = 0
			a, b := Gia_ObjFaninId0(obj, id), Gia_ObjFaninId1(obj, id)
			haveA, haveB := Gia_ObjIsTravIdCurrentId(p, a) != 0, Gia_ObjIsTravIdCurrentId(p, b) != 0
			if !haveA && !haveB {
				continue
			}
			vec.Vec_IntRemove(inputs, id)
			if !haveA {
				vec.Vec_IntPush(inputs, a)
				Gia_ObjSetTravIdCurrentId(p, a)
			}
			if !haveB {
				vec.Vec_IntPush(inputs, b)
				Gia_ObjSetTravIdCurrentId(p, b)
			}
			changed = true
			break
		}
	}
	return onlyInputs
}
func Gia_RsbFindFaninAdd(fan int32, fanins, counts []int32, n int32) int32 {
	if n < 0 || int(n) >= len(fanins) || len(counts) < len(fanins) {
		panic("gia: Rsb2 fanin capacity")
	}
	i := int32(0)
	for i < n && fanins[i] != fan {
		i++
	}
	fanins[i] = fan
	counts[i]++
	if i == n {
		n++
	}
	return n
}
func Gia_RsbFindFaninToAddToCut(p *Gia_Man_t, inputs *Vec_Int_t) int32 {
	order := make([]int32, 0)
	counts := make(map[int32]int)
	for _, id := range vec.Vec_IntArray(inputs) {
		obj := Gia_ManObj(p, id)
		if Gia_ObjIsAnd(obj) == 0 {
			continue
		}
		for _, fan := range []int32{Gia_ObjFaninId0(obj, id), Gia_ObjFaninId1(obj, id)} {
			if Gia_ObjIsTravIdCurrentId(p, fan) != 0 {
				panic("gia: marked Rsb2 fanin")
			}
			if counts[fan] == 0 {
				order = append(order, fan)
			}
			counts[fan]++
		}
	}
	best := int32(-1)
	bestCount := 0
	for _, fan := range order {
		if counts[fan] > bestCount || counts[fan] == bestCount && best >= 0 && Gia_ObjFanoutNumId(p, fan) > Gia_ObjFanoutNumId(p, best) {
			best, bestCount = fan, counts[fan]
		}
	}
	return best
}
func Gia_RsbWindowGrow2(p *Gia_Man_t, pivot int32, levels *Vec_Wec_t, window, inputs *Vec_Int_t, maxInputs int32) {
	vec.Vec_IntClear(window)
	if Gia_RsbExpandCut(p, inputs) == 0 {
		var best []int32
		if int32(vec.Vec_IntSize(inputs)) <= maxInputs {
			best = append([]int32(nil), vec.Vec_IntArray(inputs)...)
		}
		only := int32(0)
		over := 0
		for only == 0 && (int32(vec.Vec_IntSize(inputs)) <= maxInputs || over < 5) {
			fan := Gia_RsbFindFaninToAddToCut(p, inputs)
			if fan < 0 {
				break
			}
			vec.Vec_IntPush(inputs, fan)
			Gia_ObjSetTravIdCurrentId(p, fan)
			only = Gia_RsbExpandCut(p, inputs)
			if int32(vec.Vec_IntSize(inputs)) > maxInputs {
				over++
			} else {
				over = 0
				if best == nil || len(best) <= vec.Vec_IntSize(inputs) {
					best = append(best[:0], vec.Vec_IntArray(inputs)...)
				}
			}
		}
		if best != nil {
			vec.Vec_IntClear(inputs)
			for _, id := range best {
				vec.Vec_IntPush(inputs, id)
			}
		}
	}
	if int32(vec.Vec_IntSize(inputs)) <= maxInputs {
		vec.Vec_IntSort(inputs, 0)
		Gia_WinCreateFromCut(p, pivot, inputs, levels, window)
	}
}
func Gia_RsbWindowCompute(p *Gia_Man_t, id, maxInputs, maxLevels int32, levels *Vec_Wec_t, paths *Vec_Int_t, window, inputs **Vec_Int_t) int32 {
	*window, *inputs = nil, nil
	win := Gia_RsbWindowInit(p, paths, id, maxLevels)
	if win == nil {
		return 0
	}
	ins := Gia_RsbCreateWindowInputs(p, win)
	if int32(vec.Vec_IntSize(ins)) <= maxInputs+3 {
		Gia_RsbWindowGrow2(p, id, levels, win, ins, maxInputs)
	}
	if int32(vec.Vec_IntSize(ins)) > maxInputs {
		return 0
	}
	vec.Vec_IntSort(win, 0)
	vec.Vec_IntSort(ins, 0)
	*window, *inputs = win, ins
	return 1
}

// Gia_RsbExpandInputs replaces cut nodes whose fanins overlap the window.
func Gia_RsbExpandInputs(p *Gia_Man_t, levels *Vec_Wec_t, window, inputs *Vec_Int_t) {
	for changed := true; changed; {
		changed = false
		for _, id := range append([]int32(nil), vec.Vec_IntArray(inputs)...) {
			obj := Gia_ManObj(p, id)
			if Gia_ObjIsAnd(obj) == 0 {
				continue
			}
			fans := [2]int32{Gia_ObjFaninId0(obj, id), Gia_ObjFaninId1(obj, id)}
			if Gia_ObjIsTravIdCurrentId(p, fans[0]) == 0 && Gia_ObjIsTravIdCurrentId(p, fans[1]) == 0 {
				continue
			}
			vec.Vec_IntRemove(inputs, id)
			for _, fan := range fans {
				if Gia_ObjIsTravIdCurrentId(p, fan) != 0 {
					continue
				}
				vec.Vec_IntPush(inputs, fan)
				Gia_WinTryAddingNode(p, fan, -1, levels, window)
			}
			changed = true
		}
	}
}

func Gia_RsbSelectOneInput(p *Gia_Man_t, levels *Vec_Wec_t, inputs *Vec_Int_t) int32 {
	best, weight := int32(0), int32(-1)
	for _, id := range vec.Vec_IntArray(inputs) {
		obj := Gia_ManObj(p, id)
		if Gia_ObjIsAnd(obj) == 0 {
			continue
		}
		a, b := Gia_ObjFaninId0(obj, id), Gia_ObjFaninId1(obj, id)
		if Gia_ObjIsTravIdCurrentId(p, a) != 0 || Gia_ObjIsTravIdCurrentId(p, b) != 0 {
			panic("gia: marked Rsb2 input fanin")
		}
		w := Gia_WinTryAddingNode(p, a, b, levels, nil)
		if w > weight {
			best, weight = id, w
		}
	}
	return best
}

func Gia_RsbWindowGrow(p *Gia_Man_t, levels *Vec_Wec_t, window, inputs *Vec_Int_t, maxInputs int32) {
	Gia_RsbAddSideInputs(p, levels, window)
	Gia_RsbExpandInputs(p, levels, window, inputs)
	for int32(vec.Vec_IntSize(inputs)) < maxInputs {
		id := Gia_RsbSelectOneInput(p, levels, inputs)
		if id == 0 {
			break
		}
		obj := Gia_ManObj(p, id)
		a, b := Gia_ObjFaninId0(obj, id), Gia_ObjFaninId1(obj, id)
		Gia_WinTryAddingNode(p, a, b, levels, window)
		vec.Vec_IntRemove(inputs, id)
		vec.Vec_IntPush(inputs, a)
		vec.Vec_IntPush(inputs, b)
		Gia_RsbExpandInputs(p, levels, window, inputs)
	}
}
