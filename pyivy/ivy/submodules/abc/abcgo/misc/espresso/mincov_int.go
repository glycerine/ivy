package espresso

import "time"

// Statistics and selected columns shared by the exact minimum-cover search.
type stats_t struct {
	debug                                       bool
	max_print_depth, max_depth, nodes           int
	component, comp_count, gimpel_count, gimpel int
	start_time                                  time.Time
	no_branching                                bool
	lower_bound                                 int32
}
type solution_t struct {
	row  *sm_row
	cost int32
}

func mincovWeight(weight []int32, column int) int32 {
	if weight == nil {
		return 1
	}
	return weight[column]
}
