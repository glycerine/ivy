package espresso

// Exact weighted covering from Espresso mincov.c. Reduction, independent-row
// lower bounds, component decomposition and accept/reject branching retain the
// source search order; heuristic mode stops after the accept branch.
import (
	"fmt"
	"math"
	"time"
)

func sm_minimum_cover(matrix *sm_matrix, weight []int32, heuristic, debugLevel int) *sm_row {
	if matrix.nrows <= 0 {
		return sm_row_alloc()
	}
	stats := &stats_t{start_time: time.Now(), debug: debugLevel > 0, max_print_depth: debugLevel, max_depth: -1, no_branching: heuristic != 0, lower_bound: -1}
	elements := 0
	for row := matrix.first_row; row != nil; row = row.next_row {
		elements += row.length
	}
	bound := int64(1)
	for column := matrix.first_col; column != nil; column = column.next_col {
		value := mincovWeight(weight, column.col_num)
		if value < 0 {
			panic("minimum cover requires nonnegative weights")
		}
		bound += int64(value)
	}
	if bound > math.MaxInt32 {
		panic("minimum cover cost exceeds int32")
	}
	selected := solution_alloc()
	copyMatrix := sm_dup(matrix)
	best := sm_mincov(copyMatrix, selected, weight, 0, int32(bound), 0, stats)
	sm_free(copyMatrix)
	solution_free(selected)
	if best == nil {
		return nil
	}
	if stats.debug {
		if stats.no_branching {
			fmt.Printf("**** heuristic covering ...\nlower bound = %d\n", stats.lower_bound)
		}
		fmt.Printf("matrix     = %d by %d with %d elements (%4.3f%%)\ncover size = %d elements\ncover cost = %d\ntime       = %.3f sec\ncomponents = %d\ngimpel     = %d\nnodes      = %d\nmax_depth  = %d\n", matrix.nrows, matrix.ncols, elements, 100*float64(elements)/float64(matrix.nrows*matrix.ncols), best.row.length, best.cost, time.Since(stats.start_time).Seconds(), stats.comp_count, stats.gimpel_count, stats.nodes, stats.max_depth)
	}
	result := sm_row_dup(best.row)
	if !verify_cover(matrix, result) {
		panic("mincov internal error: cover verification failed")
	}
	solution_free(best)
	return result
}
func sm_mincov(matrix *sm_matrix, selected *solution_t, weight []int32, lower, bound int32, depth int, stats *stats_t) *solution_t {
	stats.nodes++
	stats.max_depth = max(stats.max_depth, depth)
	debug := stats.debug && depth <= stats.max_print_depth
	select_essential(matrix, selected, weight, bound)
	if selected.cost >= bound {
		return nil
	}
	if weight == nil {
		var reduced *solution_t
		if gimpel_reduce(matrix, selected, weight, lower, bound, depth, stats, &reduced) {
			return reduced
		}
	}
	independent := sm_maximal_independent_set(matrix, weight)
	newLower := max(selected.cost+independent.cost, lower)
	pick := select_column(matrix, weight, independent)
	solution_free(independent)
	if depth == 0 {
		stats.lower_bound = newLower + int32(stats.gimpel)
	}
	if debug {
		marker := " "
		if stats.component != 0 {
			marker = "*"
		}
		fmt.Printf("ABSMIN[%2d]%s %3dx%3d sel=%3d bnd=%3d lb=%3d %12.3f ", depth, marker, matrix.nrows, matrix.ncols, selected.cost+int32(stats.gimpel), bound+int32(stats.gimpel), newLower+int32(stats.gimpel), time.Since(stats.start_time).Seconds())
	}
	if newLower >= bound {
		if debug {
			fmt.Println("bounded")
		}
		return nil
	}
	if matrix.nrows == 0 {
		if debug {
			fmt.Println("BEST")
		}
		return solution_dup(selected)
	}
	var left, right *sm_matrix
	if sm_block_partition(matrix, &left, &right) {
		if left.ncols > right.ncols {
			left, right = right, left
		}
		if debug {
			fmt.Printf("comp %d %d\n", left.nrows, right.nrows)
		}
		stats.comp_count++
		partial := solution_alloc()
		stats.component++
		first := sm_mincov(left, partial, weight, 0, bound-selected.cost, depth+1, stats)
		stats.component--
		solution_free(partial)
		sm_free(left)
		if first == nil {
			sm_free(right)
			return nil
		}
		for element := first.row.first_col; element != nil; element = element.next_col {
			solution_add(selected, weight, element.col_num)
		}
		solution_free(first)
		best := sm_mincov(right, selected, weight, newLower, bound, depth+1, stats)
		sm_free(right)
		return best
	}
	if debug {
		fmt.Printf("pick=%d\n", pick)
	}
	acceptedMatrix := sm_dup(matrix)
	accepted := solution_dup(selected)
	solution_accept(accepted, acceptedMatrix, weight, pick)
	bestAccepted := sm_mincov(acceptedMatrix, accepted, weight, newLower, bound, depth+1, stats)
	solution_free(accepted)
	sm_free(acceptedMatrix)
	if bestAccepted != nil {
		bound = min(bound, bestAccepted.cost)
	}
	if stats.no_branching || bestAccepted != nil && bestAccepted.cost == newLower {
		return bestAccepted
	}
	rejectedMatrix := sm_dup(matrix)
	rejected := solution_dup(selected)
	solution_reject(rejected, rejectedMatrix, weight, pick)
	bestRejected := sm_mincov(rejectedMatrix, rejected, weight, newLower, bound, depth+1, stats)
	solution_free(rejected)
	sm_free(rejectedMatrix)
	return solution_choose_best(bestAccepted, bestRejected)
}
func select_column(matrix *sm_matrix, weight []int32, independent *solution_t) int {
	candidates := sm_row_alloc()
	if independent != nil {
		for element := independent.row.first_col; element != nil; element = element.next_col {
			row := sm_get_row(matrix, element.col_num)
			for e := row.first_col; e != nil; e = e.next_col {
				sm_row_insert(candidates, e.col_num)
			}
		}
	} else {
		for column := matrix.first_col; column != nil; column = column.next_col {
			sm_row_insert(candidates, column.col_num)
		}
	}
	bestColumn := -1
	best := -1.0
	for candidate := candidates.first_col; candidate != nil; candidate = candidate.next_col {
		column := sm_get_col(matrix, candidate.col_num)
		value := 0.0
		for element := column.first_row; element != nil; element = element.next_row {
			row := sm_get_row(matrix, element.row_num)
			value += 1 / float64(row.length-1)
		}
		value /= float64(mincovWeight(weight, column.col_num))
		if value > best {
			bestColumn = column.col_num
			best = value
		}
	}
	sm_row_free(candidates)
	return bestColumn
}
func select_essential(matrix *sm_matrix, selected *solution_t, weight []int32, bound int32) {
	for {
		deletedColumns := sm_col_dominance(matrix, weight)
		essential := sm_row_alloc()
		for row := matrix.first_row; row != nil; row = row.next_row {
			if row.length == 1 {
				sm_row_insert(essential, row.first_col.col_num)
			}
		}
		for element := essential.first_col; element != nil; element = element.next_col {
			solution_accept(selected, matrix, weight, element.col_num)
			if selected.cost >= bound {
				sm_row_free(essential)
				return
			}
		}
		count := essential.length
		sm_row_free(essential)
		deletedRows := sm_row_dominance(matrix)
		if deletedColumns == 0 && deletedRows == 0 && count == 0 {
			return
		}
	}
}
func verify_cover(matrix *sm_matrix, cover *sm_row) bool {
	for row := matrix.first_row; row != nil; row = row.next_row {
		if !sm_row_intersects(row, cover) {
			return false
		}
	}
	return true
}
