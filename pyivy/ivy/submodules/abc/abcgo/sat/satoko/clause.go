// Copyright Satoko contributors. Distributed under the BSD 2-Clause License.
package satoko

import "fmt"
import "io"

type clause struct {
	lits                                   []int
	lbd, activity                          uint32
	simplified, learnt, deleted, protected bool
}

func clause_compare(a, b *clause) bool {
	if len(a.lits) == 2 {
		return false
	}
	if len(b.lits) == 2 {
		return true
	}
	if a.lbd != b.lbd {
		return a.lbd > b.lbd
	}
	return a.activity < b.activity
}
func clause_dump(w io.Writer, c *clause, noZero bool) {
	for _, lit := range c.lits {
		v := lit >> 1
		if noZero {
			v++
		}
		if lit&1 != 0 {
			v = -v
		}
		fmt.Fprintf(w, "%d ", v)
	}
	if noZero {
		fmt.Fprintln(w, "0")
	} else {
		fmt.Fprintln(w)
	}
}
