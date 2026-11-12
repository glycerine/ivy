package lucky

func setSwapInfoPtr(vars int32) *swapInfo {
	x := &swapInfo{posArray: make([]varInfo, vars+2), realArray: make([]int32, vars+2), varN: vars}
	x.realArray[0] = vars + 100
	for i := int32(1); i <= vars; i++ {
		x.posArray[i] = varInfo{direction: -1, position: i}
		x.realArray[i] = i
	}
	x.realArray[vars+1] = vars + 10
	return x
}
func freeSwapInfoPtr(x *swapInfo) {
	if x != nil {
		x.posArray = nil
		x.realArray = nil
	}
}
func nextSwap(x *swapInfo) int32 {
	for i := x.varN; i > 1; i-- {
		v := &x.posArray[i]
		if i > x.realArray[v.position+v.direction] {
			v.position += v.direction
			other := x.realArray[v.position]
			x.realArray[v.position] = i
			x.realArray[v.position-v.direction] = other
			x.posArray[other].position = v.position - v.direction
			for j := x.varN; j > i; j-- {
				x.posArray[j].direction *= -1
			}
			x.positionToSwap1 = x.posArray[other].position - 1
			x.positionToSwap2 = v.position - 1
			return 1
		}
	}
	return 0
}
func fillInSwapArray(info *permInfo) {
	counter := info.totalSwaps - 1
	state := setSwapInfoPtr(info.varN)
	for nextSwap(state) == 1 {
		info.swapArray[counter] = min(state.positionToSwap1, state.positionToSwap2)
		counter--
	}
}
func oneBitPosition(x, size int32) int32 {
	for i := int32(0); i < size; i++ {
		if x>>uint(i)&1 != 0 {
			return i
		}
	}
	return -1
}
func fillInFlipArray(info *permInfo) {
	var previous int32
	for i := int32(1); i <= info.totalFlips; i++ {
		gray := i ^ (i >> 1)
		info.flipArray[info.totalFlips-i] = oneBitPosition(previous^gray, info.varN)
		previous = gray
	}
}
func factorial(n int32) int32 {
	if n < 0 || n > 12 {
		panic("permutation count exceeds int32")
	}
	result := int32(1)
	for i := int32(2); i <= n; i++ {
		result *= i
	}
	return result
}
func setPermInfoPtr(vars int32) *permInfo {
	if vars < 0 || vars > 12 {
		panic("exhaustive permutation generation supports up to twelve variables")
	}
	info := &permInfo{varN: vars, totalFlips: (1 << uint(vars)) - 1, totalSwaps: factorial(vars) - 1}
	info.flipArray = make([]int32, info.totalFlips)
	info.swapArray = make([]int32, info.totalSwaps)
	fillInSwapArray(info)
	fillInFlipArray(info)
	return info
}
func freePermInfoPtr(info *permInfo) {
	if info != nil {
		info.swapArray = nil
		info.flipArray = nil
	}
}
func minWord(a, b, minimal []uint64, vars int32) {
	if memCompare(a, b, vars) < 0 {
		Kit_TruthCopy_64bit(minimal, a, vars)
	} else {
		Kit_TruthCopy_64bit(minimal, b, vars)
	}
}
func minWord1(a, minimal []uint64, vars int32) {
	if memCompare(a, minimal, vars) <= 0 {
		Kit_TruthCopy_64bit(minimal, a, vars)
	}
}
func minWord3(a, b, minimal []uint64, vars int32) {
	if memCompare(a, b, vars) <= 0 {
		if memCompare(a, minimal, vars) < 0 {
			Kit_TruthCopy_64bit(minimal, a, vars)
		} else {
			return
		}
	}
	if memCompare(b, minimal, vars) <= 0 {
		Kit_TruthCopy_64bit(minimal, b, vars)
	}
}
func simpleMinimal(table, aux, minimal []uint64, info *permInfo, vars int32) {
	if vars == 0 {
		if table[0]>>63 != 0 {
			table[0] = ^table[0]
		}
		minimal[0] = table[0]
		return
	}
	Kit_TruthCopy_64bit(aux, table, vars)
	Kit_TruthNot_64bit(table, vars)
	minWord(table, aux, minimal, vars)
	permute := func() {
		for i := info.totalSwaps - 1; i >= 0; i-- {
			Kit_TruthSwapAdjacentVars_64bit(table, vars, info.swapArray[i])
			Kit_TruthSwapAdjacentVars_64bit(aux, vars, info.swapArray[i])
			minWord3(table, aux, minimal, vars)
		}
	}
	permute()
	for j := info.totalFlips - 1; j >= 0; j-- {
		if vars > 1 {
			Kit_TruthSwapAdjacentVars_64bit(table, vars, 0)
			Kit_TruthSwapAdjacentVars_64bit(aux, vars, 0)
		}
		Kit_TruthChangePhase_64bit(table, vars, info.flipArray[j])
		Kit_TruthChangePhase_64bit(aux, vars, info.flipArray[j])
		minWord3(table, aux, minimal, vars)
		permute()
	}
	Kit_TruthCopy_64bit(table, minimal, vars)
}
func simpleMinimalGroups(table, aux, minimal []uint64, groups []int32, nGroups int32, infos []*permInfo, vars, flipOutput, flipInput int32) {
	offsets := make([]int32, nGroups)
	for i := int32(1); i < nGroups; i++ {
		offsets[i] = offsets[i-1] + groups[i-1]
	}
	if flipOutput != 0 {
		Kit_TruthCopy_64bit(aux, table, vars)
		Kit_TruthNot_64bit(table, vars)
		minWord(table, aux, minimal, vars)
	} else {
		Kit_TruthCopy_64bit(minimal, table, vars)
	}
	n := 2*nGroups + 1
	values, changed, radices := make([]int32, n), make([]int32, n), make([]int32, n)
	radices[0] = 2
	for i := int32(1); i <= nGroups; i++ {
		radices[i] = infos[groups[i-1]].totalFlips + 1
		radices[nGroups+i] = infos[groups[i-1]].totalSwaps + 1
	}
	updateMinimum := func() {
		if flipOutput != 0 {
			minWord3(table, aux, minimal, vars)
		} else {
			minWord1(table, minimal, vars)
		}
	}
	for {
		for i := int32(1); i <= nGroups; i++ {
			if changed[i] == 0 || flipInput == 0 && groups[i-1] == 1 {
				continue
			}
			info := infos[groups[i-1]]
			j := int32(0)
			if values[i] != 0 {
				j = info.totalFlips - values[i]
			}
			Kit_TruthChangePhase_64bit(table, vars, offsets[i-1]+info.flipArray[j])
			if flipOutput != 0 {
				Kit_TruthChangePhase_64bit(aux, vars, offsets[i-1]+info.flipArray[j])
			}
			updateMinimum()
		}
		for i := int32(1); i <= nGroups; i++ {
			if changed[nGroups+i] == 0 || groups[i-1] == 1 {
				continue
			}
			info := infos[groups[i-1]]
			j := int32(0)
			if values[nGroups+i] != info.totalSwaps {
				j = info.swapArray[info.totalSwaps-values[nGroups+i]-1]
			}
			Kit_TruthSwapAdjacentVars_64bit(table, vars, offsets[i-1]+j)
			if flipOutput != 0 {
				Kit_TruthSwapAdjacentVars_64bit(aux, vars, offsets[i-1]+j)
			}
			updateMinimum()
		}
		clear(changed)
		j := n - 1
		for values[j] == radices[j]-1 {
			changed[j] = 1
			values[j] = 0
			j--
		}
		if j == 0 {
			break
		}
		changed[j] = 1
		values[j]++
	}
	Kit_TruthCopy_64bit(table, minimal, vars)
}
func SimpleMinimal(table, aux, minimal []uint64, info *PermInfo, vars int32) {
	simpleMinimal(table, aux, minimal, info, vars)
}
func SimpleMinimalGroups(table, aux, minimal []uint64, groups []int32, infos []*PermInfo, vars int32, flipOutput, flipInput bool) {
	a, b := int32(0), int32(0)
	if flipOutput {
		a = 1
	}
	if flipInput {
		b = 1
	}
	simpleMinimalGroups(table, aux, minimal, groups, int32(len(groups)), infos, vars, a, b)
}
