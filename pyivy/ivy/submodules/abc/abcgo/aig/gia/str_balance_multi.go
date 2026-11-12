package gia

import "github.com/glycerine/abcgo/misc/vec"

func Str_NtkBalanceMulti2(dst *Gia_Man_t, ntk *Str_Ntk_t, obj *Str_Obj_t, delay *Vec_Int_t, lutSize int32) {
	obj.iCopy = 0
	if obj.Type == STR_AND {
		obj.iCopy = 1
	}
	for i := int32(0); i < int32(obj.nFanins); i++ {
		lit := Str_ObjFaninCopy(ntk, obj, i)
		if obj.Type == STR_AND {
			obj.iCopy = Gia_ManHashAnd(dst, obj.iCopy, lit)
		} else {
			obj.iCopy = Gia_ManHashXorReal(dst, obj.iCopy, lit)
		}
		Str_ObjDelay(dst, obj.iCopy>>1, lutSize, delay)
	}
}
func Str_NtkBalanceTwo(dst *Gia_Man_t, ntk *Str_Ntk_t, obj *Str_Obj_t, i, j int32, delay *Vec_Int_t, costs, super []int32, matrix []uint64, size, lutSize, best int32) int32 {
	if i < 0 || i >= j || j >= size {
		panic("gia: invalid structured merge")
	}
	var result int32
	if obj.Type == STR_AND {
		result = Gia_ManHashAnd(dst, super[i], super[j])
	} else {
		result = Gia_ManHashXorReal(dst, super[i], super[j])
	}
	costs[i] = Str_ObjDelay(dst, result>>1, lutSize, delay)
	super[i] = result
	matrix[i] |= matrix[j]
	copy(costs[j:size-1], costs[j+1:size])
	copy(super[j:size-1], super[j+1:size])
	copy(matrix[j:size-1], matrix[j+1:size])
	for k := int32(0); k < size-2; k++ {
		if costs[k] <= costs[k+1] {
			break
		}
		costs[k], costs[k+1] = costs[k+1], costs[k]
		super[k], super[k+1] = super[k+1], super[k]
		matrix[k], matrix[k+1] = matrix[k+1], matrix[k]
	}
	return result
}
func Str_NtkBalanceMulti(dst *Gia_Man_t, ntk *Str_Ntk_t, obj *Str_Obj_t, delay *Vec_Int_t, lutSize int32) {
	superVec := vec.Vec_IntAlloc(256)
	for i := int32(0); i < int32(obj.nFanins); i++ {
		vec.Vec_IntPush(superVec, Str_ObjFaninCopy(ntk, obj, i))
	}
	vec.Vec_IntSort(superVec, 0)
	if obj.Type == STR_AND {
		Gia_ManSimplifyAnd(superVec)
	} else {
		Gia_ManSimplifyXor(superVec)
	}
	size := vec.Vec_IntSize(superVec)
	if size == 0 {
		panic("gia: empty balanced supergate")
	}
	if size == 1 {
		obj.iCopy = vec.Vec_IntEntry(superVec, 0)
		return
	}
	super := make([]int32, 256)
	costs := make([]int32, 256)
	matrix := make([]uint64, 256)
	copy(super, vec.Vec_IntArray(superVec))
	if size > 256 {
		panic("gia: supergate exceeds balancing capacity")
	}
	if size == 2 {
		obj.iCopy = Str_NtkBalanceTwo(dst, ntk, obj, 0, 1, delay, costs, super, matrix, 2, lutSize, -1)
		return
	}
	for i := 0; i < size; i++ {
		costs[i] = vec.Vec_IntEntry(delay, int(super[i]>>1))
	}
	vec.Vec_IntSelectSortCost2(super, size, costs)
	affinity := func() {
		vec.Vec_IntClear(superVec)
		vec.Vec_IntPushArray(superVec, super, size)
		vCosts := vec.Vec_IntAlloc(size)
		vec.Vec_IntPushArray(vCosts, costs, size)
		Str_ManVectorAffinity(dst, superVec, vCosts, matrix, 256)
	}
	if size < 64 {
		affinity()
	}
	for size > 2 {
		left, right := 0, 1
		if size <= 64 {
			if size == 64 {
				affinity()
			}
			if costs[0] != 17 || costs[1] != 17 {
				stop := 0
				for stop < size && costs[stop]>>4 == costs[0]>>4 {
					stop++
				}
				if stop == 1 {
					bestOcc := Str_CountBits(matrix[0] & matrix[1])
					for i := 2; i < size && costs[i]>>4 == costs[1]>>4; i++ {
						occ := Str_CountBits(matrix[0] & matrix[i])
						if occ > bestOcc {
							right, bestOcc = i, occ
						}
					}
				} else {
					bestCost, bestOcc := int32(-1), int32(-1)
					best := -1
					for i := 1; i < stop; i++ {
						cost := (costs[0] & 15) + (costs[i] & 15)
						if cost > lutSize {
							continue
						}
						occ := Str_CountBits(matrix[0] & matrix[i])
						if cost > bestCost || cost == bestCost && occ > bestOcc {
							best, bestCost, bestOcc = i, cost, occ
						}
					}
					if bestCost == lutSize {
						right = best
					} else {
						feasibleI, feasibleJ, overflowI, overflowJ := -1, -1, -1, -1
						bestCost, bestOcc = -1, -1
						overflowCost, overflowOcc := int32(-1), int32(-1)
						for i := 0; i < stop; i++ {
							for j := i + 1; j < stop; j++ {
								cost := (costs[i] & 15) + (costs[j] & 15)
								occ := Str_CountBits(matrix[i] & matrix[j])
								if cost <= lutSize {
									if occ > bestOcc || occ == bestOcc && cost > bestCost {
										feasibleI, feasibleJ, bestCost, bestOcc = i, j, cost, occ
									}
								} else if occ > overflowOcc || occ == overflowOcc && cost > overflowCost {
									overflowI, overflowJ, overflowCost, overflowOcc = i, j, cost, occ
								}
							}
						}
						if feasibleI >= 0 {
							left, right = feasibleI, feasibleJ
						} else {
							left, right = overflowI, overflowJ
						}
					}
				}
			}
		}
		if left < 0 || right < 0 {
			panic("gia: failed to choose balanced pair")
		}
		Str_NtkBalanceTwo(dst, ntk, obj, int32(left), int32(right), delay, costs, super, matrix, int32(size), lutSize, -1)
		size--
	}
	obj.iCopy = Str_NtkBalanceTwo(dst, ntk, obj, 0, 1, delay, costs, super, matrix, 2, lutSize, -1)
}
