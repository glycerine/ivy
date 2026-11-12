package cut

// Recorded steps retain each cut's actual parent nodes, cut indices and phase.
// This extends C's pair-only record so choice-union cuts replay correctly too.
type Cut_Oracle_t struct {
	manager *Cut_Man_t
	records map[int32]Cut_Record
}

func Cut_OracleStart(source *Cut_Man_t) *Cut_Oracle_t {
	source.check()
	if source.params.FRecord == 0 {
		panic("cut: recording was not enabled")
	}
	params := *source.params
	params.FRecord = 0
	oracle := &Cut_Oracle_t{manager: Cut_ManStart(&params), records: source.records}
	source.records = map[int32]Cut_Record{}
	return oracle
}
func Cut_OracleStop(p *Cut_Oracle_t) {
	if p != nil {
		Cut_ManStop(p.manager)
		p.records = nil
	}
}
func Cut_OracleSetFanoutCounts(p *Cut_Oracle_t, counts []int32) {
	Cut_ManSetFanoutCounts(p.manager, counts)
}
func Cut_OracleReadDrop(p *Cut_Oracle_t) int32          { return p.manager.params.FDrop }
func Cut_OracleNodeSetTriv(p *Cut_Oracle_t, node int32) { Cut_NodeSetTriv(p.manager, node) }
func cutSlice(list *Cut_Cut_t) []*Cut_Cut_t {
	var cuts []*Cut_Cut_t
	for c := list; c != nil; c = c.Next {
		cuts = append(cuts, c)
	}
	return cuts
}
func Cut_OracleComputeCuts(p *Cut_Oracle_t, node, node0, node1, compl0, compl1 int32) *Cut_Cut_t {
	record, ok := p.records[node]
	if !ok {
		panic("cut: no recorded node")
	}
	if p.manager.newCuts[node] != nil {
		panic("cut: oracle node already computed")
	}
	var head *Cut_Cut_t
	tail := &head
	for _, step := range record.Steps {
		sources := step.Proof.Sources
		phases := step.Proof.Phases
		if step.Proof.Node == node {
			sources = [2]int32{node0, node1}
			phases = [2]bool{compl0 != 0, compl1 != 0}
		}
		var c *Cut_Cut_t
		if step.Proof.Kind == 0 {
			c = Cut_CutCreateTriv(p.manager, node)
		} else {
			a := cutSlice(p.manager.newCuts[sources[0]])
			if int(step.Indices[0]) >= len(a) {
				panic("cut: recorded first fanin unavailable")
			}
			left := a[step.Indices[0]]
			if step.Proof.Kind == 3 {
				c = Cut_CutDupList(p.manager, &Cut_Cut_t{NVarsMax: left.NVarsMax, leaves: left.leaves, truth: left.truth, FSimul: left.FSimul, FCompl: left.FCompl, USign: left.USign})
				if phases[0] {
					for i := range c.truth {
						c.truth[i] = ^c.truth[i]
					}
				}
			} else {
				b := cutSlice(p.manager.newCuts[sources[1]])
				if int(step.Indices[1]) >= len(b) {
					panic("cut: recorded second fanin unavailable")
				}
				right := b[step.Indices[1]]
				c = Cut_CutMergeTwo(p.manager, left, right)
				if c == nil {
					panic("cut: recorded merge exceeds limit")
				}
				if p.manager.params.FTruth != 0 {
					cutTruthComputeOp(c, left, right, cutBool(phases[0]), cutBool(phases[1]), step.Proof.Kind == 2)
				}
			}
		}
		c.FCompl, c.FSimul = step.Complement, step.Simulation
		*tail = c
		tail = &c.Next
	}
	p.manager.newCuts[node] = head
	return head
}
func Cut_OracleFreeCuts(p *Cut_Oracle_t, node int32)        { Cut_NodeFreeCuts(p.manager, node) }
func Cut_OracleTryDroppingCuts(p *Cut_Oracle_t, node int32) { Cut_NodeTryDroppingCuts(p.manager, node) }
