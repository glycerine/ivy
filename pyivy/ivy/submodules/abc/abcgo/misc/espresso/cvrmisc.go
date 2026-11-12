// Cover costs and execution accounting, ported from cvrmisc.c.
package espresso

import (
	"fmt"
	"time"
)

// Elapsed monotonic milliseconds replace the C process-CPU clock. Keeping the
// source millisecond unit preserves timing accumulation and printed precision.
var espressoClockStart = time.Now()

func ptime() int64 { return time.Since(espressoClockStart).Milliseconds() }
func print_time(milliseconds int64) string {
	return fmt.Sprintf("%d.%02d sec", milliseconds/1000, milliseconds%1000/10)
}

func cover_cost(f pcover, cost pcost) {
	*cost = cost_t{cubes: f.count}
	// Count directly rather than constructing a temporary cube list. This is
	// the same zero/one count as massive_count, without mutating its scratch data.
	for row := int32(0); row < f.count; row++ {
		current := GETSET(f, row)
		if TESTP(current, PRIME) {
			cost.primes++
		}
		for variable := int32(0); variable < cube.num_vars; variable++ {
			ones := int32(0)
			for part := cube.first_part[variable]; part <= cube.last_part[variable]; part++ {
				if is_in_set(current, part) {
					ones++
				}
			}
			zeros := cube.part_size[variable] - ones
			switch {
			case variable < cube.num_binary_vars:
				cost.in += zeros
			case variable == cube.num_vars-1:
				cost.out += ones
			case cube.sparse[variable] != 0:
				cost.mv += ones
			default:
				cost.mv += zeros
			}
		}
	}
	cost.total = cost.in + cost.out + cost.mv
}
func fmt_cost(cost pcost) string {
	if cube.num_binary_vars == cube.num_vars-1 {
		return fmt.Sprintf("c=%d(%d) in=%d out=%d tot=%d", cost.cubes, cost.cubes-cost.primes, cost.in, cost.out, cost.total)
	}
	return fmt.Sprintf("c=%d(%d) in=%d mv=%d out=%d", cost.cubes, cost.cubes-cost.primes, cost.in, cost.mv, cost.out)
}
func print_cost(f pcover) string          { var cost cost_t; cover_cost(f, &cost); return fmt_cost(&cost) }
func copy_cost(source, destination pcost) { *destination = *source }
func size_stamp(f pcover, name string)    { fmt.Printf("# %s\tCost is %s\n", name, print_cost(f)) }
func print_trace(f pcover, name string, elapsed int64) {
	fmt.Printf("# %s\tTime was %s, cost is %s\n", name, print_time(elapsed), print_cost(f))
}
func totals(start int64, index int32, f pcover, cost pcost) {
	elapsed := ptime() - start
	total_time[index] += elapsed
	total_calls[index]++
	cover_cost(f, cost)
	if trace {
		fmt.Printf("# %s\tTime was %s, cost is %s\n", total_name[index], print_time(elapsed), fmt_cost(cost))
	}
}

// A library reports fatal source errors through panic rather than terminating
// the embedding process. Callers can recover at their application boundary.
func fatal(message string) { panic("espresso: " + message) }
