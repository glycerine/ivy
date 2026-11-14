package core

import (
	"fmt"
	"io"
	"sort"
)

// zddSymmState mirrors subtableZ[].next in CUDD. It is a circular ring of
// positions: a singleton points to itself and a symmetry class links its
// adjacent levels in order.
type zddSymmState = zddGroupState

type zddSymmMove struct{ x, y, size int }

func newZddSymmState(n int) *zddSymmState {
	s := &zddSymmState{next: make([]int, n)}
	for i := range s.next {
		s.next[i] = i
	}
	return s
}

func zddSymmGroupSize(s *zddSymmState, x int) int {
	if x < 0 || x >= len(s.next) {
		return 0
	}
	n, size := s.next[x], 1
	for n != x && size <= len(s.next) {
		size++
		n = s.next[n]
	}
	return size
}

func zddSymmMergeUp(s *zddSymmState, y, x int) {
	gytop := s.next[y]
	s.next[y] = x
	i := s.next[x]
	for s.next[i] != x {
		i = s.next[i]
	}
	s.next[i] = gytop
}

func zddSymmMergeDown(s *zddSymmState, x, y int) {
	gxtop := s.next[x]
	s.next[x] = y
	i := s.next[y]
	for s.next[i] != y {
		i = s.next[i]
	}
	s.next[i] = gxtop
}

func zddSymmSwap(dd *DdManager, s *zddSymmState, x, y int) int {
	if y != x+1 || x < 0 || y >= dd.ZddSize || !swapZddAdjacentInPlace(dd, x, y) {
		return 0
	}
	s.swaps++
	return zddHeapSize(dd)
}

// zddSymmGroupMove is the CUDD symmetric-sifting block movement. Ring links
// are relinked immediately before each physical adjacent swap, matching the
// source algorithm's level semantics.
func zddSymmGroupMove(dd *DdManager, s *zddSymmState, x, y int, moves *[]zddSymmMove) int {
	gxtop, gxbot := s.next[x], x
	gybot := s.next[y]
	for s.next[gybot] != y {
		gybot = s.next[gybot]
	}
	yp, swapx, swapy, size := gybot, -1, -1, 0
	for x <= y {
		for y > gxtop {
			temp := s.next[x]
			if temp == x {
				temp = y
			}
			i := gxtop
			for s.next[i] != x {
				i = s.next[i]
			}
			s.next[i] = y
			if s.next[y] != y {
				s.next[x] = s.next[y]
			} else {
				s.next[x] = x
			}
			if yp != y {
				s.next[yp] = x
			} else {
				yp = x
			}
			s.next[y] = temp
			size = zddSymmSwap(dd, s, x, y)
			if size == 0 {
				return 0
			}
			swapx, swapy = x, y
			y, x = x, x-1
		}
		if s.next[y] <= y {
			gybot = y
		} else {
			y = s.next[y]
		}
		yp, gxtop, gxbot, x = gxtop, gxtop+1, gxbot+1, gxbot+1
	}
	*moves = append([]zddSymmMove{{swapx, swapy, zddHeapSize(dd)}}, *moves...)
	return zddHeapSize(dd)
}

func zddSymmGroupMoveBackward(dd *DdManager, s *zddSymmState, x, y int) int {
	gxtop, gxbot := s.next[x], x
	gybot := s.next[y]
	for s.next[gybot] != y {
		gybot = s.next[gybot]
	}
	yp, size := gybot, -1
	for x <= y {
		for y > gxtop {
			temp := s.next[x]
			if temp == x {
				temp = y
			}
			i := gxtop
			for s.next[i] != x {
				i = s.next[i]
			}
			s.next[i] = y
			if s.next[y] != y {
				s.next[x] = s.next[y]
			} else {
				s.next[x] = x
			}
			if yp != y {
				s.next[yp] = x
			} else {
				yp = x
			}
			s.next[y] = temp
			size = zddSymmSwap(dd, s, x, y)
			if size == 0 {
				return 0
			}
			y, x = x, x-1
		}
		if s.next[y] <= y {
			gybot = y
		} else {
			y = s.next[y]
		}
		yp, gxtop, gxbot, x = gxtop, gxtop+1, gxbot+1, gxbot+1
	}
	return size
}

func zddSymmGroupUp(dd *DdManager, s *zddSymmState, y, low int, moves *[]zddSymmMove) bool {
	limit := zddHeapSize(dd)
	x := y - 1
	for x >= low {
		gtop := s.next[x]
		if zddSymmCheck(dd, x, y) {
			zddSymmMergeUp(s, x, y)
		} else if s.next[x] == x && s.next[y] == y {
			size := zddSymmSwap(dd, s, x, y)
			if size == 0 {
				return false
			}
			*moves = append([]zddSymmMove{{x, y, size}}, *moves...)
			if float64(size) > float64(limit)*dd.maxGrowth {
				return true
			}
			if size < limit {
				limit = size
			}
		} else {
			size := zddSymmGroupMove(dd, s, x, y, moves)
			if size == 0 {
				return false
			}
			if float64(size) > float64(limit)*dd.maxGrowth {
				return true
			}
			if size < limit {
				limit = size
			}
		}
		y = gtop
		x = y - 1
	}
	return true
}

func zddSymmGroupDown(dd *DdManager, s *zddSymmState, x, high int, moves *[]zddSymmMove) bool {
	limit := zddHeapSize(dd)
	y := x + 1
	for y <= high {
		gbot := s.next[y]
		for s.next[gbot] != y {
			gbot = s.next[gbot]
		}
		if zddSymmCheck(dd, x, y) {
			zddSymmMergeDown(s, x, y)
		} else if s.next[x] == x && s.next[y] == y {
			size := zddSymmSwap(dd, s, x, y)
			if size == 0 {
				return false
			}
			*moves = append([]zddSymmMove{{x, y, size}}, *moves...)
			if float64(size) > float64(limit)*dd.maxGrowth {
				return true
			}
			if size < limit {
				limit = size
			}
			x = y
		} else {
			size := zddSymmGroupMove(dd, s, x, y, moves)
			if size == 0 {
				return false
			}
			if float64(size) > float64(limit)*dd.maxGrowth {
				return true
			}
			if size < limit {
				limit = size
			}
		}
		x = gbot
		y = x + 1
	}
	return true
}

func zddSymmBackward(dd *DdManager, s *zddSymmState, moves []zddSymmMove, size int) bool {
	best := -1
	for i, m := range moves {
		if m.size < size {
			best = i
			size = m.size
		}
	}
	for i, m := range moves {
		if i == best {
			break
		}
		if s.next[m.x] == m.x && s.next[m.y] == m.y {
			if zddSymmSwap(dd, s, m.x, m.y) == 0 {
				return false
			}
		} else if zddSymmGroupMoveBackward(dd, s, m.x, m.y) == 0 {
			return false
		}
		if best < 0 && zddHeapSize(dd) == size {
			break
		}
	}
	return true
}

func zddSymmSiftingAux(dd *DdManager, s *zddSymmState, x, low, high int) bool {
	initial := zddHeapSize(dd)
	var down, up []zddSymmMove
	// Discover neighboring symmetry classes around x and splice their rings.
	for i := x; i > low; i-- {
		if !zddSymmCheck(dd, i-1, i) {
			break
		}
		top := s.next[i-1]
		s.next[i-1] = i
		s.next[x] = top
		i = top + 1
	}
	for i := x; i < high; i++ {
		if !zddSymmCheck(dd, i, i+1) {
			break
		}
		bot := i + 1
		for bot < s.next[bot] {
			bot = s.next[bot]
		}
		s.next[bot] = s.next[i]
		s.next[i] = i + 1
		i = bot - 1
	}
	if x == low {
		for x < s.next[x] {
			x = s.next[x]
		}
		initGroup := x - s.next[x] + 1
		if !zddSymmGroupDown(dd, s, x, high, &down) {
			return false
		}
		if len(down) == 0 || s.next[down[0].y] != down[0].y {
			xpos := s.next[x]
			if len(down) > 0 {
				xpos = down[0].y
			}
			i := xpos
			for i < s.next[i] {
				i = s.next[i]
			}
			final := i - xpos + 1
			if initGroup == final {
				return zddSymmBackward(dd, s, down, initial)
			}
			initial = zddHeapSize(dd)
			if !zddSymmGroupUp(dd, s, xpos, low, &up) {
				return false
			}
			return zddSymmBackward(dd, s, up, initial)
		}
		return zddSymmBackward(dd, s, down, initial)
	}
	if x == high {
		for x < s.next[x] {
			x = s.next[x]
		}
		x = s.next[x]
		i := x
		for i < s.next[i] {
			i = s.next[i]
		}
		initGroup := i - x + 1
		if !zddSymmGroupUp(dd, s, x, low, &up) {
			return false
		}
		if len(up) == 0 || s.next[up[0].x] != up[0].x {
			xpos := x
			if len(up) > 0 {
				xpos = up[0].x
			} else {
				for xpos < s.next[xpos] {
					xpos = s.next[xpos]
				}
			}
			i = s.next[xpos]
			final := xpos - i + 1
			if initGroup == final {
				return zddSymmBackward(dd, s, up, initial)
			}
			initial = zddHeapSize(dd)
			if !zddSymmGroupDown(dd, s, xpos, high, &down) {
				return false
			}
			return zddSymmBackward(dd, s, down, initial)
		}
		return zddSymmBackward(dd, s, up, initial)
	}
	if x-low > high-x {
		for x < s.next[x] {
			x = s.next[x]
		}
		if !zddSymmGroupDown(dd, s, x, high, &down) {
			return false
		}
		x = s.next[x]
		if len(down) > 0 {
			x = down[0].y
		}
		i := x
		for i < s.next[i] {
			i = s.next[i]
		}
		initGroup := i - x + 1
		if !zddSymmGroupUp(dd, s, x, low, &up) {
			return false
		}
		if len(up) == 0 || s.next[up[0].x] != up[0].x {
			xpos := x
			if len(up) > 0 {
				xpos = up[0].x
			} else {
				for xpos < s.next[xpos] {
					xpos = s.next[xpos]
				}
			}
			i = s.next[xpos]
			final := xpos - i + 1
			if initGroup == final {
				return zddSymmBackward(dd, s, up, initial)
			}
			down = nil
			initial = zddHeapSize(dd)
			if !zddSymmGroupDown(dd, s, xpos, high, &down) {
				return false
			}
			return zddSymmBackward(dd, s, down, initial)
		}
		return zddSymmBackward(dd, s, up, initial)
	}
	for x < s.next[x] {
		x = s.next[x]
	}
	x = s.next[x]
	if !zddSymmGroupUp(dd, s, x, low, &up) {
		return false
	}
	if len(up) > 0 {
		x = up[0].x
	} else {
		for x < s.next[x] {
			x = s.next[x]
		}
	}
	i := s.next[x]
	initGroup := x - i + 1
	if !zddSymmGroupDown(dd, s, x, high, &down) {
		return false
	}
	if len(down) == 0 || s.next[down[0].y] != down[0].y {
		xpos := s.next[x]
		if len(down) > 0 {
			xpos = down[0].y
		}
		i := xpos
		for i < s.next[i] {
			i = s.next[i]
		}
		final := i - xpos + 1
		if initGroup == final {
			return zddSymmBackward(dd, s, down, initial)
		}
		up = nil
		initial = zddHeapSize(dd)
		if !zddSymmGroupUp(dd, s, xpos, low, &up) {
			return false
		}
		return zddSymmBackward(dd, s, up, initial)
	}
	return zddSymmBackward(dd, s, down, initial)
}

// zddSymmSiftingConvAux is CUDD's convergence-pass variant. Unlike the first
// pass it does not discover symmetries around x before moving it; it starts
// from the representatives already encoded in the ring.
func zddSymmSiftingConvAux(dd *DdManager, s *zddSymmState, x, low, high int) bool {
	initial := zddHeapSize(dd)
	var down, up []zddSymmMove
	if x == low {
		initGroup := x - s.next[x] + 1
		if !zddSymmGroupDown(dd, s, x, high, &down) {
			return false
		}
		if len(down) == 0 || s.next[down[0].y] != down[0].y {
			if len(down) > 0 {
				x = down[0].y
			} else {
				for x < s.next[x] {
					x = s.next[x]
				}
				x = s.next[x]
			}
			i := x
			for i < s.next[i] {
				i = s.next[i]
			}
			finalGroup := i - x + 1
			if initGroup == finalGroup {
				return zddSymmBackward(dd, s, down, initial)
			}
			initial = zddHeapSize(dd)
			if !zddSymmGroupUp(dd, s, x, low, &up) {
				return false
			}
			return zddSymmBackward(dd, s, up, initial)
		}
		return zddSymmBackward(dd, s, down, initial)
	}
	if x == high {
		for x < s.next[x] {
			x = s.next[x]
		}
		x = s.next[x]
		i := x
		for i < s.next[i] {
			i = s.next[i]
		}
		initGroup := i - x + 1
		if !zddSymmGroupUp(dd, s, x, low, &up) {
			return false
		}
		if len(up) == 0 || s.next[up[0].x] != up[0].x {
			if len(up) > 0 {
				x = up[0].x
			} else {
				for x < s.next[x] {
					x = s.next[x]
				}
			}
			i = s.next[x]
			finalGroup := x - i + 1
			if initGroup == finalGroup {
				return zddSymmBackward(dd, s, up, initial)
			}
			initial = zddHeapSize(dd)
			if !zddSymmGroupDown(dd, s, x, high, &down) {
				return false
			}
			return zddSymmBackward(dd, s, down, initial)
		}
		return zddSymmBackward(dd, s, up, initial)
	}
	if x-low > high-x {
		if !zddSymmGroupDown(dd, s, x, high, &down) {
			return false
		}
		if len(down) > 0 {
			x = down[0].y
		} else {
			for x < s.next[x] {
				x = s.next[x]
			}
			x = s.next[x]
		}
		i := x
		for i < s.next[i] {
			i = s.next[i]
		}
		initGroup := i - x + 1
		if !zddSymmGroupUp(dd, s, x, low, &up) {
			return false
		}
		if len(up) == 0 || s.next[up[0].x] != up[0].x {
			if len(up) > 0 {
				x = up[0].x
			} else {
				for x < s.next[x] {
					x = s.next[x]
				}
			}
			i = s.next[x]
			finalGroup := x - i + 1
			if initGroup == finalGroup {
				return zddSymmBackward(dd, s, up, initial)
			}
			down = nil
			initial = zddHeapSize(dd)
			if !zddSymmGroupDown(dd, s, x, high, &down) {
				return false
			}
			return zddSymmBackward(dd, s, down, initial)
		}
		return zddSymmBackward(dd, s, up, initial)
	}
	x = s.next[x]
	if !zddSymmGroupUp(dd, s, x, low, &up) {
		return false
	}
	if len(up) > 0 {
		x = up[0].x
	} else {
		for x < s.next[x] {
			x = s.next[x]
		}
	}
	i := s.next[x]
	initGroup := x - i + 1
	if !zddSymmGroupDown(dd, s, x, high, &down) {
		return false
	}
	if len(down) == 0 || s.next[down[0].y] != down[0].y {
		if len(down) > 0 {
			x = down[0].y
		} else {
			for x < s.next[x] {
				x = s.next[x]
			}
			x = s.next[x]
		}
		i = x
		for i < s.next[i] {
			i = s.next[i]
		}
		finalGroup := i - x + 1
		if initGroup == finalGroup {
			return zddSymmBackward(dd, s, down, initial)
		}
		up = nil
		initial = zddHeapSize(dd)
		if !zddSymmGroupUp(dd, s, x, low, &up) {
			return false
		}
		return zddSymmBackward(dd, s, up, initial)
	}
	return zddSymmBackward(dd, s, down, initial)
}

func zddSymmKeys(dd *DdManager, variable int) int {
	level := dd.PermZ[variable]
	idx := variable
	n := 0
	for _, node := range dd.zddUnique {
		if int(node.Index) == idx && dd.PermZ[node.Index] == level {
			n++
		}
	}
	return n
}

// cuddZddSymmSifting performs source-order symmetric sifting and returns the
// C internal result (one plus the number of variables in nontrivial groups).
func cuddZddSymmSifting(dd *DdManager, lower, upper int, converge bool) int {
	if dd == nil || lower < 0 || upper >= dd.ZddSize || lower > upper {
		return 0
	}
	s := newZddSymmState(dd.ZddSize)
	return zddSymmSiftingWithState(dd, s, lower, upper, converge)
}

func zddSymmSiftingWithState(dd *DdManager, s *zddSymmState, lower, upper int, converge bool) int {
	if dd == nil || s == nil || lower < 0 || upper >= dd.ZddSize || lower > upper {
		return 0
	}
	for level := lower; level <= upper; level++ {
		s.next[level] = level
	}
	n := dd.ZddSize
	initialSize := zddHeapSize(dd)
	vars := make([]int, n)
	for i := 0; i < n; i++ {
		vars[i] = i
	}
	sort.SliceStable(vars, func(i, j int) bool { return zddSymmKeys(dd, vars[i]) > zddSymmKeys(dd, vars[j]) })
	iteration := min(dd.siftMaxVar, n)
	for _, v := range vars[:iteration] {
		if s.swaps >= dd.siftMaxSwap {
			break
		}
		x := dd.PermZ[v]
		if x < lower || x > upper || s.next[x] != x {
			continue
		}
		if !zddSymmSiftingAux(dd, s, x, lower, upper) {
			return 0
		}
	}
	if converge {
		for pass := 0; initialSize > zddHeapSize(dd) && pass < n; pass++ {
			initialSize = zddHeapSize(dd)
			reps := make([]int, 0, n)
			for x := lower; x <= upper; x++ {
				for x < s.next[x] {
					x = s.next[x]
				}
				reps = append(reps, dd.InvPermZ[x])
			}
			sort.SliceStable(reps, func(i, j int) bool { return zddSymmKeys(dd, reps[i]) > zddSymmKeys(dd, reps[j]) })
			for _, v := range reps {
				if s.swaps >= dd.siftMaxSwap {
					break
				}
				x := dd.PermZ[v]
				if x < lower || x > upper || x < s.next[x] {
					continue
				}
				if !zddSymmSiftingConvAux(dd, s, x, lower, upper) {
					return 0
				}
			}
		}
	}
	zddSymmSaveGroups(dd, s, lower, upper)
	return 1 + zddSymmCount(s, lower, upper)
}

func zddSymmSaveGroups(dd *DdManager, s *zddSymmState, lower, upper int) {
	groups := make([][]int, 0)
	for x := lower; x <= upper; {
		group := []int{}
		cur := x
		for {
			group = append(group, dd.InvPermZ[cur])
			cur = s.next[cur]
			if cur == x {
				break
			}
		}
		if len(group) > 1 {
			groups = append(groups, group)
		}
		bottom := x
		for s.next[bottom] != x {
			bottom = s.next[bottom]
		}
		x = bottom + 1
	}
	dd.symmetryGroups = groups
}

// Cudd_zddSymmSifting exposes the source helper result to the outer package.
// The CUDD public dispatcher normalizes any positive result to success.
func Cudd_zddSymmSifting(dd *DdManager, lower, upper int, converge bool) int {
	return cuddZddSymmSifting(dd, lower, upper, converge)
}

// Cudd_zddSymmCheck exposes the adjacent-level symmetry predicate.
func Cudd_zddSymmCheck(dd *DdManager, x, y int) int {
	if zddSymmCheck(dd, x, y) {
		return 1
	}
	return 0
}

func zddSymmCount(s *zddSymmState, low, high int) int {
	count := 0
	for i := low; i <= high; i++ {
		if s.next[i] != i {
			count++
		}
	}
	return count
}

// Cudd_zddSymmProfile prints the current circular symmetry classes.
func Cudd_zddSymmProfile(dd *DdManager, lower, upper int) {
	Cudd_zddSymmProfileTo(dd, lower, upper, nil)
}

// Cudd_zddSymmProfileTo writes the current discovered symmetry groups.
func Cudd_zddSymmProfileTo(dd *DdManager, lower, upper int, out io.Writer) {
	if dd == nil {
		return
	}
	if out == nil {
		out = dd.stdout
	}
	if out == nil {
		out = io.Discard
	}
	count, groups := 0, 0
	for _, group := range dd.SymmetryGroups() {
		if len(group) < 2 {
			continue
		}
		start := -1
		for i, variable := range group {
			if variable >= 0 && variable < len(dd.PermZ) && dd.PermZ[variable] >= lower && dd.PermZ[variable] < upper {
				start = i
				break
			}
		}
		if start < 0 {
			continue
		}
		fmt.Fprint(out, "Group:")
		for offset := range len(group) {
			variable := group[(start+offset)%len(group)]
			fmt.Fprintf(out, "  %d", variable)
			count++
		}
		fmt.Fprintln(out)
		groups++
	}
	fmt.Fprintf(out, "Total Symmetric = %d\nTotal Groups = %d\n", count, groups)
}
