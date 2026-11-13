package tim

import "fmt"

func Tim_ManStart(cis, cos int) *Tim_Man_t {
	if cis < 0 || cos < 0 {
		panic("tim: negative terminal count")
	}
	p := &Tim_Man_t{nCis: cis, nCos: cos, pCis: make([]Tim_Obj_t, cis), pCos: make([]Tim_Obj_t, cos), fUseTravId: true}
	for i := range p.pCis {
		p.pCis[i] = Tim_Obj_t{Id: i, iObj2Box: -1, iObj2Num: -1, timeReq: TIM_ETERNITY}
	}
	for i := range p.pCos {
		p.pCos[i] = Tim_Obj_t{Id: i, iObj2Box: -1, iObj2Num: -1, timeReq: TIM_ETERNITY}
	}
	return p
}
func copyTables(p *Tim_Man_t, unit int) [][]float32 {
	if p.vDelayTables == nil {
		return nil
	}
	tables := make([][]float32, len(p.vDelayTables))
	for i, t := range p.vDelayTables {
		if t == nil {
			continue
		}
		tables[i] = append([]float32(nil), t...)
		tables[i][0] = float32(i)
		if unit != 0 {
			for j := 3; j < len(t); j++ {
				if t[j] != -TIM_ETERNITY {
					tables[i][j] = float32(unit)
				}
			}
		}
	}
	return tables
}
func Tim_ManDup(p *Tim_Man_t, unit int) *Tim_Man_t {
	q := Tim_ManStart(p.nCis, p.nCos)
	copy(q.pCis, p.pCis)
	copy(q.pCos, p.pCos)
	for i := range q.pCis {
		q.pCis[i].TravId = 0
	}
	for i := range q.pCos {
		q.pCos[i].TravId = 0
	}
	q.vDelayTables = copyTables(p, unit)
	if unit != 0 {
		Tim_ManInitPiArrivalAll(q, 0)
		Tim_ManInitPoRequiredAll(q, TIM_ETERNITY)
	}
	for _, b := range p.vBoxes {
		black := 0
		if b.fBlack {
			black = 1
		}
		Tim_ManCreateBox(q, b.firstIn, b.nInputs, b.firstOut, b.nOutputs, b.iDelayTable, black)
		q.vBoxes[len(q.vBoxes)-1].iCopy = b.iCopy
	}
	return q
}

// CloneTiming lets AIG graph duplication preserve timing metadata without
// depending on the timing package's concrete type.
func (p *Tim_Man_t) CloneTiming() any { return Tim_ManDup(p, 0) }
func Tim_ManTrim(p *Tim_Man_t, present []int) *Tim_Man_t {
	if len(present) != len(p.vBoxes) {
		panic("tim: trim mask size")
	}
	var boxes []int
	for i, keep := range present {
		if keep != 0 {
			boxes = append(boxes, i)
		}
	}
	if len(boxes) == len(p.vBoxes) {
		return Tim_ManDup(p, 0)
	}
	return Tim_ManReduce(p, boxes, 0)
}
func Tim_ManReduce(p *Tim_Man_t, boxes []int, diff int) *Tim_Man_t {
	if diff < 0 {
		panic("tim: negative terminal reduction")
	}
	pi, po := Tim_ManPiNum(p)-diff, Tim_ManPoNum(p)-diff
	if pi < 0 || po < 0 {
		panic("tim: invalid terminal reduction")
	}
	cis, cos := pi, po
	seen := map[int]bool{}
	for _, id := range boxes {
		if id < 0 || id >= len(p.vBoxes) || seen[id] {
			panic("tim: invalid selected box")
		}
		seen[id] = true
		cis += p.vBoxes[id].nOutputs
		cos += p.vBoxes[id].nInputs
	}
	q := Tim_ManStart(cis, cos)
	copy(q.pCis[:pi], p.pCis[:pi])
	copy(q.pCos[cos-po:], p.pCos[p.nCos-Tim_ManPoNum(p):p.nCos-Tim_ManPoNum(p)+po])
	q.vDelayTables = copyTables(p, 0)
	ci, co := pi, 0
	for _, id := range boxes {
		b := p.vBoxes[id]
		black := 0
		if b.fBlack {
			black = 1
		}
		Tim_ManCreateBox(q, co, b.nInputs, ci, b.nOutputs, b.iDelayTable, black)
		copyID := b.iCopy
		if copyID == -1 {
			copyID = id
		}
		q.vBoxes[len(q.vBoxes)-1].iCopy = copyID
		ci += b.nOutputs
		co += b.nInputs
	}
	// Terminal IDs are positions in the new manager, not the copied old positions.
	for i := range q.pCis {
		q.pCis[i].Id = i
		q.pCis[i].TravId = 0
	}
	for i := range q.pCos {
		q.pCos[i].Id = i
		q.pCos[i].TravId = 0
	}
	return q
}
func Tim_ManAlignTwo(spec, impl *Tim_Man_t) []int {
	if len(spec.vBoxes) <= len(impl.vBoxes) {
		panic("tim: alignment requires a larger specification")
	}
	mask := make([]int, len(spec.vBoxes))
	for _, b := range impl.vBoxes {
		if b.iCopy < 0 || b.iCopy >= len(mask) {
			return nil
		}
		if mask[b.iCopy] != 0 {
			panic("tim: duplicate box copy")
		}
		mask[b.iCopy] = 1
	}
	return mask
}
func Tim_ManStop(p *Tim_Man_t) {
	if p != nil {
		*p = Tim_Man_t{}
	}
}
func Tim_ManStopP(p **Tim_Man_t) {
	if p != nil {
		Tim_ManStop(*p)
		*p = nil
	}
}

// TimingLibrary supplies box data without a dependency on a particular mapper.
type TimingLibrary interface {
	BoxCount() int
	BoxTiming(id int) (inputs, outputs int, black bool, delays []float32)
}

func Tim_ManCreate(p *Tim_Man_t, lib TimingLibrary, arrivals, required []float32) {
	if p.vDelayTables != nil {
		panic("tim: delay tables already set")
	}
	if lib != nil {
		p.vDelayTables = make([][]float32, lib.BoxCount())
	} else {
		p.vDelayTables = make([][]float32, 0)
	}
	for _, b := range p.vBoxes {
		table := make([]float32, 3+b.nInputs*b.nOutputs)
		table[0] = float32(b.iDelayTable)
		table[1] = float32(b.nInputs)
		table[2] = float32(b.nOutputs)
		if b.iDelayTable == -1 || lib == nil {
			for j := 3; j < len(table); j++ {
				table[j] = 1
			}
			b.iDelayTable = len(p.vDelayTables)
			table[0] = float32(b.iDelayTable)
			p.vDelayTables = append(p.vDelayTables, table)
			continue
		}
		ins, outs, black, delays := lib.BoxTiming(b.iDelayTable)
		if ins != b.nInputs || outs != b.nOutputs || len(delays) != ins*outs {
			panic("tim: library box dimensions")
		}
		b.fBlack = black
		if p.vDelayTables[b.iDelayTable] == nil {
			copy(table[3:], delays)
			p.vDelayTables[b.iDelayTable] = table
		}
	}
	if arrivals != nil {
		if len(arrivals) != Tim_ManPiNum(p) {
			panic("tim: PI arrival count")
		}
		k := 0
		for i := range p.pCis {
			if p.pCis[i].iObj2Box < 0 {
				p.pCis[i].timeArr = arrivals[k]
				k++
			}
		}
	}
	if required != nil {
		if len(required) != Tim_ManPoNum(p) {
			panic("tim: PO required count")
		}
		k := 0
		for i := range p.pCos {
			if p.pCos[i].iObj2Box < 0 {
				p.pCos[i].timeReq = required[k]
				k++
			}
		}
	}
}
func Tim_ManGetArrTimes(p *Tim_Man_t) []float32 {
	var result []float32
	for _, o := range p.pCis {
		if o.iObj2Box < 0 && o.timeArr != 0 {
			result = make([]float32, p.nCis)
			break
		}
	}
	if result != nil {
		for i, o := range p.pCis {
			if o.iObj2Box < 0 {
				result[i] = o.timeArr
			}
		}
	}
	return result
}
func Tim_ManGetReqTimes(p *Tim_Man_t) []float32 {
	var result []float32
	for _, o := range p.pCos {
		if o.iObj2Box < 0 && o.timeReq != TIM_ETERNITY {
			result = make([]float32, p.nCos)
			break
		}
	}
	if result != nil {
		k := 0
		for _, o := range p.pCos {
			if o.iObj2Box < 0 {
				result[k] = o.timeReq
				k++
			}
		}
	}
	return result
}
func Tim_ManCiNum(p *Tim_Man_t) int { return p.nCis }
func Tim_ManCoNum(p *Tim_Man_t) int { return p.nCos }
func Tim_ManPiNum(p *Tim_Man_t) int {
	if len(p.vBoxes) == 0 {
		return p.nCis
	}
	return Tim_ManBoxOutputFirst(p, 0)
}
func Tim_ManPoNum(p *Tim_Man_t) int {
	if len(p.vBoxes) == 0 {
		return p.nCos
	}
	i := len(p.vBoxes) - 1
	return p.nCos - Tim_ManBoxInputFirst(p, i) - p.vBoxes[i].nInputs
}
func Tim_ManBoxNum(p *Tim_Man_t) int { return len(p.vBoxes) }
func Tim_ManBlackBoxNum(p *Tim_Man_t) int {
	n := 0
	for _, b := range p.vBoxes {
		if b.fBlack {
			n++
		}
	}
	return n
}
func Tim_ManBlackBoxIoNum(p *Tim_Man_t, ins, outs *int) {
	*ins = 0
	*outs = 0
	for _, b := range p.vBoxes {
		if b.fBlack {
			*ins += b.nInputs
			*outs += b.nOutputs
		}
	}
}
func Tim_ManDelayTableNum(p *Tim_Man_t) int { return len(p.vDelayTables) }
func Tim_ManSetDelayTables(p *Tim_Man_t, tables [][]float32) {
	if p.vDelayTables != nil {
		panic("tim: delay tables already set")
	}
	p.vDelayTables = tables
}
func Tim_ManTravIdDisable(p *Tim_Man_t) { p.fUseTravId = false }
func Tim_ManTravIdEnable(p *Tim_Man_t)  { p.fUseTravId = true }
func Tim_ManPrint(p *Tim_Man_t) {
	if p == nil {
		return
	}
	fmt.Printf("TIMING MANAGER:\nPI = %d. CI = %d. PO = %d. CO = %d. Box = %d.\n", Tim_ManPiNum(p), p.nCis, Tim_ManPoNum(p), p.nCos, len(p.vBoxes))
	for i, o := range p.pCis {
		fmt.Printf("CI %d: arrival = %5.3f required = %5.3f\n", i, o.timeArr, o.timeReq)
	}
	for i, o := range p.pCos {
		fmt.Printf("CO %d: arrival = %5.3f required = %5.3f\n", i, o.timeArr, o.timeReq)
	}
	for i, b := range p.vBoxes {
		fmt.Printf("Box %d: inputs=%d outputs=%d table=%d\n", i, b.nInputs, b.nOutputs, b.iDelayTable)
	}
	for i, t := range p.vDelayTables {
		fmt.Printf("Delay table %d: %v\n", i, t)
	}
}
func Tim_ManPrintBoxCopy(p *Tim_Man_t) {
	if p == nil {
		return
	}
	for _, b := range p.vBoxes {
		fmt.Printf("%d ", b.iCopy)
	}
	fmt.Println()
}
func Tim_ManPrintStats(p *Tim_Man_t, and2 int) {
	if p == nil {
		return
	}
	fmt.Printf("Hierarchy: PI/CI = %d/%d PO/CO = %d/%d Box = %d delay(AND2) = %d\n", Tim_ManPiNum(p), p.nCis, Tim_ManPoNum(p), p.nCos, len(p.vBoxes), and2)
	counts := map[int]int{}
	for _, b := range p.vBoxes {
		counts[b.iDelayTable]++
	}
	for _, b := range p.vBoxes {
		if count := counts[b.iDelayTable]; count > 0 {
			fmt.Printf("Box %d: Num = %d Ins = %d Outs = %d\n", b.iDelayTable, count, b.nInputs, b.nOutputs)
			delete(counts, b.iDelayTable)
		}
	}
}
