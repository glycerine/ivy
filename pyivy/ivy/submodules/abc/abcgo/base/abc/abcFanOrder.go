package abc

import (
	"bytes"
	"fmt"
	"slices"
	"sort"
	"strings"
)

type fanSOPRow struct {
	pattern string
	output  byte
}

func fanSOPRows(text string, variables int) []fanSOPRow {
	if index := strings.IndexByte(text, 0); index >= 0 {
		text = text[:index]
	}
	var rows []fanSOPRow
	polarity := byte(0)
	for _, line := range strings.Split(text, "\n") {
		if line == "" {
			continue
		}
		if len(line) != variables+2 || line[variables] != ' ' {
			panic("SOP arity does not match fanins")
		}
		output := line[variables+1]
		if output != '0' && output != '1' && output != 'x' && output != 'n' {
			panic("invalid SOP output")
		}
		if polarity != 0 && polarity != output {
			panic("mixed SOP polarity")
		}
		polarity = output
		for _, c := range line[:variables] {
			if c != '-' && c != '0' && c != '1' {
				panic("invalid SOP literal")
			}
		}
		rows = append(rows, fanSOPRow{line[:variables], output})
	}
	return rows
}
func fanSOPText(rows []fanSOPRow) string {
	var text strings.Builder
	for _, row := range rows {
		text.WriteString(row.pattern)
		text.WriteByte(' ')
		text.WriteByte(row.output)
		text.WriteByte('\n')
	}
	return text.String()
}
func fanNodeRows(o *Object) []fanSOPRow {
	text, ok := o.Data.(string)
	if !ok {
		panic("fanin ordering requires SOP functionality")
	}
	return fanSOPRows(text, len(o.fanins))
}

// Selection sort intentionally retains the source's deterministic tie order.
func fanPermutation(count int, less func(int, int) bool) []int {
	order := make([]int, count)
	for i := range order {
		order[i] = i
	}
	for i := 0; i < count-1; i++ {
		best := i
		for j := i + 1; j < count; j++ {
			if less(order[j], order[best]) {
				best = j
			}
		}
		order[i], order[best] = order[best], order[i]
	}
	return order
}
func permuteFanSOP(o *Object, rows []fanSOPRow, order []int) {
	o.network.detachAnd(o)
	old := o.fanins
	ordered := make([]Edge, len(old))
	for i, index := range order {
		ordered[i] = old[index]
	}
	o.replaceFanins(ordered)
	for i, row := range rows {
		pattern := make([]byte, len(order))
		for j, index := range order {
			pattern[j] = row.pattern[index]
		}
		rows[i].pattern = string(pattern)
	}
	o.Data = fanSOPText(rows)
}
func Abc_NtkOrderFaninsById(n *Network) {
	for _, o := range n.ObjectsOfKind(ObjectNode) {
		if o.Barrier {
			continue
		}
		rows := fanNodeRows(o)
		order := fanPermutation(len(o.fanins), func(a, b int) bool { return o.fanins[a].Object.ID < o.fanins[b].Object.ID })
		permuteFanSOP(o, rows, order)
	}
}
func Abc_NtkSopTranspose(text string, variables int32, columns *[]string, store *[]byte) {
	rows := fanSOPRows(text, int(variables))
	*columns = (*columns)[:0]
	*store = (*store)[:0]
	for variable := 0; variable < int(variables); variable++ {
		start := len(*store)
		for _, row := range rows {
			*store = append(*store, row.pattern[variable])
		}
		*columns = append(*columns, string((*store)[start:]))
		*store = append(*store, 0)
	}
}
func Vec_StrSelectSortCost(columns []string, count int32, permutation *[]int32) {
	if count < 0 || int(count) > len(columns) {
		panic("invalid column count")
	}
	order := fanPermutation(int(count), func(a, b int) bool { return columns[a] < columns[b] })
	old := append([]string(nil), columns[:count]...)
	*permutation = (*permutation)[:0]
	for i, index := range order {
		columns[i] = old[index]
		*permutation = append(*permutation, int32(index))
	}
}
func Vec_StrSelectSortCost2(columns []string, count int32, counts []int32, permutation *[]int32) {
	if count < 0 || int(count) > len(columns) || int(count) > len(counts) {
		panic("invalid column count")
	}
	order := fanPermutation(int(count), func(a, b int) bool { return counts[a] < counts[b] || counts[a] == counts[b] && columns[a] < columns[b] })
	old := append([]string(nil), columns[:count]...)
	*permutation = (*permutation)[:0]
	for i, index := range order {
		columns[i] = old[index]
		*permutation = append(*permutation, int32(index))
	}
}

// fanOrderScratch holds orderFanColumnsFast's reusable buffers.
type fanOrderScratch struct {
	columns []byte
	counts  []int
	last    []int
	text    []byte
}

// orderFanColumnsFast is orderFanColumns on the SOP bytes of one node in
// ABC's layout (nVars literals, a space, the output, a newline per cube):
// no per-row or per-column strings, and the text and fanins are rewritten
// only when the order changes. It reports false (doing nothing) for any
// other layout, which the general path then handles.
func orderFanColumnsFast(o *Object, mode int, sc *fanOrderScratch) bool {
	text, ok := o.Data.(string)
	if !ok {
		return false
	}
	if i := strings.IndexByte(text, 0); i >= 0 {
		text = text[:i]
	}
	vars := len(o.fanins)
	width := vars + 3
	if vars == 0 || len(text) == 0 || len(text)%width != 0 {
		return false
	}
	cubes := len(text) / width
	polarity := text[vars+1]
	for c := 0; c < cubes; c++ {
		row := text[c*width : (c+1)*width]
		if row[vars] != ' ' || row[vars+1] != polarity || row[vars+2] != '\n' {
			return false
		}
		if out := row[vars+1]; out != '0' && out != '1' && out != 'x' && out != 'n' {
			return false
		}
		for _, ch := range []byte(row[:vars]) {
			if ch != '-' && ch != '0' && ch != '1' {
				return false
			}
		}
	}
	sc.columns = append(sc.columns[:0], make([]byte, vars*cubes)...)
	sc.counts = append(sc.counts[:0], make([]int, vars)...)
	sc.last = append(sc.last[:0], make([]int, vars)...)
	columns, counts, last := sc.columns, sc.counts, sc.last
	for v := 0; v < vars; v++ {
		col := columns[v*cubes : (v+1)*cubes]
		for c := 0; c < cubes; c++ {
			ch := text[c*width+v]
			col[c] = ch
			if ch != '-' {
				counts[v]++
				last[v] = c
			}
		}
	}
	if mode == 2 {
		for v := range counts {
			if counts[v] == 1 {
				counts[v] = last[v]
			} else {
				counts[v] = int(^uint(0) >> 1)
			}
		}
	}
	column := func(v int) []byte { return columns[v*cubes : (v+1)*cubes] }
	order := fanPermutation(vars, func(a, b int) bool {
		if mode == 2 {
			return counts[a] < counts[b]
		}
		if mode == 1 && counts[a] != counts[b] {
			return counts[a] < counts[b]
		}
		return bytes.Compare(column(a), column(b)) < 0
	})
	identity := true
	for i, v := range order {
		if i != v {
			identity = false
			break
		}
	}
	if identity {
		o.Data = text // as the general path, without anything past a NUL
		return true
	}
	o.network.detachAnd(o)
	ordered := make([]Edge, vars)
	for i, index := range order {
		ordered[i] = o.fanins[index]
	}
	o.replaceFanins(ordered)
	sc.text = append(sc.text[:0], text...)
	for c := 0; c < cubes; c++ {
		row := sc.text[c*width:]
		src := text[c*width:]
		for i, index := range order {
			row[i] = src[index]
		}
	}
	o.Data = string(sc.text)
	return true
}

func orderFanColumns(n *Network, mode int) {
	var sc fanOrderScratch
	for _, o := range n.ObjectsOfKind(ObjectNode) {
		if o.Barrier {
			continue
		}
		if !orderFanColumnsFast(o, mode, &sc) {
			orderFanColumnsNode(o, mode)
		}
	}
}

// orderFanColumnsNode is the general path for one node.
func orderFanColumnsNode(o *Object, mode int) {
	{
		rows := fanNodeRows(o)
		variables := len(o.fanins)
		columns := make([]string, variables)
		counts, lastCube := make([]int, variables), make([]int, variables)
		for variable := 0; variable < variables; variable++ {
			column := make([]byte, len(rows))
			for cube, row := range rows {
				column[cube] = row.pattern[variable]
				if column[cube] != '-' {
					counts[variable]++
					lastCube[variable] = cube
				}
			}
			columns[variable] = string(column)
		}
		if mode == 2 {
			for variable := range counts {
				if counts[variable] == 1 {
					counts[variable] = lastCube[variable]
				} else {
					counts[variable] = int(^uint(0) >> 1)
				}
			}
		}
		order := fanPermutation(variables, func(a, b int) bool {
			if mode == 2 {
				return counts[a] < counts[b]
			}
			if mode == 1 && counts[a] != counts[b] {
				return counts[a] < counts[b]
			}
			return columns[a] < columns[b]
		})
		permuteFanSOP(o, rows, order)
	}
}
func Abc_NtkOrderFaninsBySortingColumns(n *Network)       { orderFanColumns(n, 0) }
func Abc_NtkOrderFaninsByLitCount(n *Network)             { orderFanColumns(n, 1) }
func Abc_NtkOrderFaninsByLitCountAndCubeCount(n *Network) { orderFanColumns(n, 2) }
func Abc_NodeSplitLarge(o *Object) {
	rows := fanNodeRows(o)
	if len(rows) < 2 {
		panic("cannot split a single SOP cube")
	}
	left, right := Abc_NtkDupObj(o.network, o, 0), Abc_NtkDupObj(o.network, o, 0)
	for _, edge := range o.fanins {
		Abc_ObjAddFanin(left, edge)
		Abc_ObjAddFanin(right, edge)
	}
	Abc_ObjRemoveFanins(o)
	Abc_ObjAddFanin(o, left)
	Abc_ObjAddFanin(o, right)
	if rows[0].output == '0' {
		o.Data = "11 1\n"
	} else {
		o.Data = "1- 1\n-1 1\n"
	}
	cut := len(rows) / 2
	left.Data = fanSOPText(rows[:cut])
	right.Data = fanSOPText(rows[cut:])
}
func Abc_NtkSplitLarge(n *Network, maxFanins, maxCubes int32) {
	for _, o := range n.ObjectsOfKind(ObjectNode) {
		if o.Barrier {
			continue
		}
		rows := fanNodeRows(o)
		if len(rows) > 1 && (len(o.fanins) > int(maxFanins) || len(rows) > int(maxCubes)) {
			Abc_NodeSplitLarge(o)
		}
	}
}
func Abc_NodeCompareCubes1(first, second string) int32 { return int32(strings.Compare(first, second)) }
func Abc_NodeCompareCubes2(first, second string) int32 {
	a, b := len(first)-strings.Count(first, "-"), len(second)-strings.Count(second, "-")
	if a > b {
		return -1
	}
	if a < b {
		return 1
	}
	return Abc_NodeCompareCubes1(first, second)
}
func Abc_NodeSortCubes(o *Object, cubes *[]string, store *[]byte, weight int32) {
	if abcNodeSortCubesFast(o, store, weight) {
		return
	}
	rows := fanNodeRows(o)
	sort.SliceStable(rows, func(i, j int) bool {
		if weight != 0 {
			return Abc_NodeCompareCubes2(rows[i].pattern, rows[j].pattern) < 0
		}
		return rows[i].pattern < rows[j].pattern
	})
	*cubes = (*cubes)[:0]
	for _, row := range rows {
		*cubes = append(*cubes, row.pattern)
	}
	o.Data = fanSOPText(rows)
	*store = append((*store)[:0], o.Data.(string)...)
}
func Abc_NtkSortCubes(n *Network, weight int32) {
	var cubes []string
	var store []byte
	for _, o := range n.ObjectsOfKind(ObjectNode) {
		if o.Barrier {
			continue
		}
		Abc_NodeSortCubes(o, &cubes, &store, weight)
	}
}
func Abc_NtkSortSops(n *Network) {
	Abc_NtkSortCubes(n, 1)
	Abc_NtkOrderFaninsByLitCount(n)
	Abc_NtkSortCubes(n, 0)
	Abc_NtkOrderFaninsByLitCountAndCubeCount(n)
	Abc_NtkSortCubes(n, 0)
}
func Abc_CubeContain(first, second string, variables int32) int32 {
	if variables < 0 || int(variables) > len(first) || int(variables) > len(second) {
		panic("invalid cube arity")
	}
	firstContains, secondContains := true, true
	for i := 0; i < int(variables); i++ {
		if first[i] == second[i] {
			continue
		}
		if first[i] == '-' {
			secondContains = false
		} else if second[i] == '-' {
			firstContains = false
		} else {
			return 0
		}
		if !firstContains && !secondContains {
			return 0
		}
	}
	var result int32
	if firstContains {
		result |= 1
	}
	if secondContains {
		result |= 2
	}
	return result
}
func Abc_NodeMakeSCCFree(o *Object) int32 {
	rows := fanNodeRows(o)
	removed := make([]bool, len(rows))
	changes := false
	for i := range rows {
		for j := i + 1; j < len(rows); j++ {
			if removed[i] || removed[j] {
				continue
			}
			status := Abc_CubeContain(rows[i].pattern, rows[j].pattern, int32(len(o.fanins)))
			if status&1 != 0 {
				removed[j] = true
				changes = true
			} else if status&2 != 0 {
				removed[i] = true
				changes = true
			}
		}
	}
	if !changes {
		return 0
	}
	kept := rows[:0]
	for i, row := range rows {
		if !removed[i] {
			kept = append(kept, row)
		}
	}
	o.Data = fanSOPText(kept)
	return 1
}
func Abc_NodeMakeDist1Free(o *Object) {
	rows := fanNodeRows(o)
	for i := range rows {
		for j := i + 1; j < len(rows); j++ {
			differences, index := 0, -1
			for variable := range rows[i].pattern {
				if rows[i].pattern[variable] != rows[j].pattern[variable] {
					differences++
					index = variable
				}
			}
			if differences == 1 && rows[i].pattern[index] != '-' && rows[j].pattern[index] != '-' {
				first, second := []byte(rows[i].pattern), []byte(rows[j].pattern)
				first[index], second[index] = '-', '-'
				rows[i].pattern, rows[j].pattern = string(first), string(second)
			}
		}
	}
	o.Data = fanSOPText(rows)
}
func Abc_NodeCheckDist1Free(o *Object) {
	rows := fanNodeRows(o)
	for i := range rows {
		for j := i + 1; j < len(rows); j++ {
			differences := 0
			for variable := range rows[i].pattern {
				if rows[i].pattern[variable] != rows[j].pattern[variable] {
					differences++
				}
			}
			if differences <= 1 {
				panic("SOP is not distance-1 free")
			}
		}
	}
}
func Abc_NodeMakeLegit(o *Object) int32 {
	iterations := 0
	for {
		iterations++
		Abc_NodeMakeDist1Free(o)
		if Abc_NodeMakeSCCFree(o) == 0 {
			break
		}
	}
	return abcBool(iterations > 1)
}
func Abc_NtkMakeLegit(n *Network) int32 {
	var count int32
	for _, o := range n.ObjectsOfKind(ObjectNode) {
		if o.Barrier {
			continue
		}
		count += Abc_NodeMakeLegit(o)
	}
	if count > 0 {
		fmt.Printf("%d nodes were made dist1-cube-free and/or single-cube-containment-free.\n", count)
	}
	return 1
}

// abcNodeSortCubesFast sorts the cubes of a well-formed SOP (rows of
// fanins + 3 bytes: literals, ' ', output, '\n') by row offset and writes
// them into one buffer, as abcFanOrder.c sorts in place. It reports false
// for anything else, which the general path validates.
func abcNodeSortCubesFast(o *Object, store *[]byte, weight int32) bool {
	text, ok := o.Data.(string)
	if !ok {
		return false
	}
	if index := strings.IndexByte(text, 0); index >= 0 {
		text = text[:index]
	}
	nVars := len(o.fanins)
	width := nVars + 3
	if len(text) == 0 || len(text)%width != 0 {
		return false
	}
	rows := len(text) / width
	output := text[nVars+1]
	for r := 0; r < rows; r++ {
		line := text[r*width : (r+1)*width]
		if line[nVars] != ' ' || line[nVars+1] != output || line[nVars+2] != '\n' {
			return false
		}
	}
	if output != '0' && output != '1' && output != 'x' && output != 'n' {
		return false
	}
	order := make([]int32, rows)
	for i := range order {
		order[i] = int32(i)
	}
	cube := func(r int32) string { return text[int(r)*width : int(r)*width+nVars] }
	// rows with equal keys are equal text, so the order among them is not
	// observable (C's qsort is not stable either)
	if weight != 0 {
		slices.SortFunc(order, func(a, b int32) int { return int(Abc_NodeCompareCubes2(cube(a), cube(b))) })
	} else {
		slices.SortFunc(order, func(a, b int32) int { return strings.Compare(cube(a), cube(b)) })
	}
	buf := (*store)[:0]
	for _, r := range order {
		buf = append(buf, text[int(r)*width:(int(r)+1)*width]...)
	}
	*store = buf
	o.Data = string(buf)
	return true
}
