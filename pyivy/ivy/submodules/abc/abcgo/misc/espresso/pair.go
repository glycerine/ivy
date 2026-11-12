package espresso

import "fmt"

func set_pair(pla pPLA) { set_pair1(pla, true) }
func set_pair1(pla pPLA, adjustLabels bool) {
	pair := pla.pair
	if pair == nil {
		fatal("missing pairing")
	}
	if adjustLabels {
		makeup_labels(pla)
	}
	paired := make([]bool, cube.num_binary_vars)
	for i := int32(0); i < pair.cnt; i++ {
		a, b := pair.var1[i]-1, pair.var2[i]-1
		if a < 0 || b < 0 || a >= cube.num_binary_vars || b >= cube.num_binary_vars || a == b || paired[a] || paired[b] {
			fatal("pairing requires distinct binary-valued variables")
		}
		paired[a], paired[b] = true, true
	}
	oldSize, oldVariables, oldBinary := cube.size, cube.num_vars, cube.num_binary_vars
	oldMV := oldBinary * 2
	pla.F = delvar(pairvar(pla.F, pair), paired)
	pla.R = delvar(pairvar(pla.R, pair), paired)
	pla.D = delvar(pairvar(pla.D, pair), paired)
	newBinary := int32(0)
	for _, used := range paired {
		if !used {
			newBinary++
		}
	}
	newVariables := newBinary + pair.cnt + oldVariables - oldBinary
	parts := make([]int32, newVariables)
	for i := int32(0); i < pair.cnt; i++ {
		parts[newBinary+i] = 4
	}
	copy(parts[newBinary+pair.cnt:], cube.part_size[oldBinary:oldVariables])
	setdown_cube()
	cube.num_vars, cube.num_binary_vars, cube.part_size = newVariables, newBinary, parts
	cube_setup()
	if adjustLabels {
		old := pla.label
		pla.label = make([]string, cube.size)
		for i := int32(0); i < pair.cnt; i++ {
			base := newBinary*2 + i*4
			a, b := (pair.var1[i]-1)*2, (pair.var2[i]-1)*2
			pla.label[base] = old[a] + "+" + old[b]
			pla.label[base+1] = old[a] + "+" + old[b+1]
			pla.label[base+2] = old[a+1] + "+" + old[b]
			pla.label[base+3] = old[a+1] + "+" + old[b+1]
		}
		destination := 0
		for v := int32(0); v < oldBinary; v++ {
			if !paired[v] {
				pla.label[2*destination], pla.label[2*destination+1] = old[2*v], old[2*v+1]
				destination++
			}
		}
		copy(pla.label[newBinary*2+pair.cnt*4:], old[oldMV:oldSize])
	}
	for i := int32(0); i < pair.cnt; i++ {
		cube.sparse[newBinary+i] = 0
	}
}
func pairvar(a pcover, pair ppair) pcover {
	// Paired parts precede all original MV variables in the new cube schema.
	// The original insertion before only the output misplaced preexisting MV parts.
	insert := cube.num_binary_vars * 2
	a = sf_delcol(a, insert, -4*pair.cnt)
	for row := int32(0); row < a.count; row++ {
		p := GETSET(a, row)
		for j := int32(0); j < pair.cnt; j++ {
			first, second := cube.first_part[pair.var1[j]-1], cube.first_part[pair.var2[j]-1]
			base := insert + j*4
			for av := int32(0); av < 2; av++ {
				for bv := int32(0); bv < 2; bv++ {
					if is_in_set(p, first+av) && is_in_set(p, second+bv) {
						set_insert(p, base+3-2*av-bv)
					}
				}
			}
		}
	}
	return a
}
func delvar(a pcover, paired []bool) pcover {
	offset := int32(0)
	for v := int32(0); v < cube.num_binary_vars; {
		if !paired[v] {
			v++
			continue
		}
		first := cube.first_part[v]
		length := int32(0)
		for v < cube.num_binary_vars && paired[v] {
			length += cube.part_size[v]
			v++
		}
		a = sf_delcol(a, first-offset, length)
		offset += length
	}
	return a
}
func pair_new(n int32) ppair {
	if n < 0 {
		fatal("negative pair capacity")
	}
	return &pair_t{var1: make([]int32, n), var2: make([]int32, n)}
}
func pair_save(pair ppair, n int32) ppair {
	if n < pair.cnt {
		fatal("pair capacity smaller than count")
	}
	result := pair_new(n)
	result.cnt = pair.cnt
	copy(result.var1, pair.var1[:pair.cnt])
	copy(result.var2, pair.var2[:pair.cnt])
	return result
}
func pair_free(pair ppair) {
	if pair != nil {
		*pair = pair_t{}
	}
}
func print_pair(pair ppair) {
	fmt.Print("pair is")
	for i := int32(0); i < pair.cnt; i++ {
		fmt.Printf(" (%d %d)", pair.var1[i], pair.var2[i])
	}
	fmt.Println()
}
func generate_all_pairs(pair ppair, n int32, candidate pset, action func(ppair)) {
	if set_ord(candidate) < 2 {
		action(pair)
		return
	}
	work := pair_save(pair, n)
	remaining := set_save(candidate)
	first := int32(0)
	for first < n && !is_in_set(candidate, first) {
		first++
	}
	for second := first + 1; second < n; second++ {
		if !is_in_set(candidate, second) {
			continue
		}
		set_remove(remaining, first)
		set_remove(remaining, second)
		work.var1[work.cnt], work.var2[work.cnt] = first+1, second+1
		work.cnt++
		generate_all_pairs(work, n, remaining, action)
		work.cnt--
		set_insert(remaining, first)
		set_insert(remaining, second)
	}
	if set_ord(candidate)%2 == 1 {
		set_remove(remaining, first)
		generate_all_pairs(work, n, remaining, action)
	}
}
func greedy_best_cost(costs [][]int32, result *ppair) int32 {
	pair := pair_new(cube.num_binary_vars)
	candidate := set_full(cube.num_binary_vars)
	total := int32(0)
	for set_ord(candidate) >= 2 {
		best := int32(-2147483648)
		a, b := int32(-1), int32(-1)
		for i := int32(0); i < cube.num_binary_vars; i++ {
			if !is_in_set(candidate, i) {
				continue
			}
			for j := i + 1; j < cube.num_binary_vars; j++ {
				if is_in_set(candidate, j) && costs[i][j] > best {
					best, a, b = costs[i][j], i, j
				}
			}
		}
		pair.var1[pair.cnt], pair.var2[pair.cnt] = a+1, b+1
		pair.cnt++
		set_remove(candidate, a)
		set_remove(candidate, b)
		total += best
	}
	*result = pair
	return total
}

var pairingBestCost int32
var pairingCosts [][]int32
var pairingBest ppair
var pairingBestPhase pset
var pairingPLA pPLA
var pairingBestF, pairingBestD, pairingBestR pcover
var pairingStrategy int32

func pair_best_cost(costs [][]int32) ppair {
	pairingBestCost = -2147483648
	pairingBest = nil
	pairingCosts = costs
	generate_all_pairs(pair_new(cube.num_binary_vars), cube.num_binary_vars, set_full(cube.num_binary_vars), find_best_cost)
	return pairingBest
}
func find_best_cost(pair ppair) {
	cost := int32(0)
	for i := int32(0); i < pair.cnt; i++ {
		cost += pairingCosts[pair.var1[i]-1][pair.var2[i]-1]
	}
	if cost > pairingBestCost {
		pairingBestCost = cost
		pairingBest = pair_save(pair, pair.cnt)
	}
	if debug&MINCOV != 0 && trace {
		fmt.Printf("cost is %d ", cost)
		print_pair(pair)
	}
}
func find_pairing_cost(pla pPLA, strategy int32) [][]int32 {
	if strategy < 0 || strategy > 3 {
		fatal("invalid pairing cost strategy")
	}
	n := cube.num_binary_vars
	costs := make([][]int32, n)
	for i := range costs {
		costs[i] = make([]int32, n)
	}
	pla.pair = pair_new(1)
	pla.pair.cnt = 1
	for a := int32(0); a < n-1; a++ {
		for b := a + 1; b < n; b++ {
			var f, d, r pcover
			oldBinary, oldVariables := cube.num_binary_vars, cube.num_vars
			oldParts := append([]int32(nil), cube.part_size...)
			if strategy > 0 {
				f, d, r = sf_save(pla.F), sf_save(pla.D), sf_save(pla.R)
				pla.pair.var1[0], pla.pair.var2[0] = a+1, b+1
				set_pair1(pla, false)
			}
			cost := int32(0)
			switch strategy {
			case 3:
				pla.F = minimize_exact(pla.F, pla.D, pla.R, true)
				cost = f.count - pla.F.count
			case 2:
				pla.F = espresso(pla.F, pla.D, pla.R)
				cost = f.count - pla.F.count
			case 1:
				pla.F = reduce(pla.F, pla.D)
				pla.F = expand(pla.F, pla.R, false)
				pla.F = irredundant(pla.F, pla.D)
				cost = f.count - pla.F.count
			case 0:
				mask := set_or(new_cube(), cube.var_mask[a], cube.var_mask[b])
				merged := dist_merge(sf_save(pla.F), mask)
				cost = pla.F.count - merged.count
			}
			costs[a][b] = cost
			if strategy > 0 {
				setdown_cube()
				cube.num_binary_vars, cube.num_vars, cube.part_size = oldBinary, oldVariables, oldParts
				cube_setup()
				pla.F, pla.D, pla.R = f, d, r
			}
		}
	}
	pla.pair = nil
	return costs
}
func find_optimal_pairing(pla pPLA, strategy int32) {
	costs := find_pairing_cost(pla, strategy)
	if summary {
		fmt.Print("    ")
		for i := int32(0); i < cube.num_binary_vars; i++ {
			fmt.Printf("%3d ", i+1)
		}
		fmt.Println()
		for i, row := range costs {
			fmt.Printf("%3d ", i+1)
			for _, cost := range row {
				fmt.Printf("%3d ", cost)
			}
			fmt.Println()
		}
	}
	if cube.num_binary_vars <= 14 {
		pla.pair = pair_best_cost(costs)
	} else {
		greedy_best_cost(costs, &pla.pair)
	}
	fmt.Print("# ")
	print_pair(pla.pair)
	set_pair(pla)
	pla.F = espresso(pla.F, pla.D, pla.R)
}
func pair_all(pla pPLA, strategy int32) {
	pairingPLA = pla
	pairingStrategy = strategy
	pairingBestCost = pla.F.count + 1
	pairingBest, pairingBestPhase = nil, nil
	pairingBestF, pairingBestD, pairingBestR = nil, nil, nil
	generate_all_pairs(pair_new(cube.num_binary_vars), cube.num_binary_vars, set_full(cube.num_binary_vars), minimize_pair)
	pla.pair, pla.phase = pairingBest, pairingBestPhase
	set_pair(pla)
	fmt.Print("# ")
	print_pair(pla.pair)
	pla.F, pla.D, pla.R = pairingBestF, pairingBestD, pairingBestR
}
func minimize_pair(pair ppair) {
	pla := pairingPLA
	f, d, r := sf_save(pla.F), sf_save(pla.D), sf_save(pla.R)
	oldBinary, oldVariables := cube.num_binary_vars, cube.num_vars
	oldParts := append([]int32(nil), cube.part_size...)
	pla.pair = pair
	set_pair1(pla, false)
	if summary {
		print_pair(pair)
	}
	switch pairingStrategy {
	case 2:
		phase_assignment(pla, 0)
		if summary {
			fmt.Printf("# phase is %s\n", pc1(pla.phase))
		}
	case 1:
		pla.F = minimize_exact(pla.F, pla.D, pla.R, true)
	case 0:
		pla.F = espresso(pla.F, pla.D, pla.R)
	}
	if pla.F.count < pairingBestCost {
		pairingBestCost = pla.F.count
		pairingBest = pair_save(pair, pair.cnt)
		pairingBestPhase = nil
		if pla.phase != nil {
			pairingBestPhase = set_save(pla.phase)
		}
		pairingBestF, pairingBestD, pairingBestR = sf_save(pla.F), sf_save(pla.D), sf_save(pla.R)
	}
	setdown_cube()
	cube.num_binary_vars, cube.num_vars, cube.part_size = oldBinary, oldVariables, oldParts
	cube_setup()
	pla.F, pla.D, pla.R = f, d, r
	pla.pair, pla.phase = nil, nil
}
