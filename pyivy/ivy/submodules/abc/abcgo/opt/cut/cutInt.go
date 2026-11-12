package cut

import "time"

type Cut_Stats struct {
	FancyCuts, FancyGood                                                                                  int64
	SeqIterations                                                                                         int32
	SeqConverged                                                                                          bool
	Current, Allocated, Recycled, Peak, Trivial, Filtered, Limited, Nodes, DagNodes, NoCuts, MappingDelay int64
	MergeTime, TruthTime, FilterTime, MappingTime                                                         time.Duration
}
type Cut_Proof struct {
	Node    int32
	Sources [2]int32
	Phases  [2]bool
	Kind    uint8
}
type Cut_Step struct {
	Proof                  Cut_Proof
	Indices                [2]uint16
	Complement, Simulation bool
}
type Cut_Record struct{ Steps []Cut_Step }
type Cut_Man_t struct {
	network                    map[int32]NetworkNode
	networkCounts              []int32
	params                     *Cut_Params_t
	newCuts, oldCuts, tempCuts map[int32]*Cut_Cut_t
	fanCounts, attrs           []int32
	owned                      map[*Cut_Cut_t]struct{}
	records                    map[int32]Cut_Record
	delays                     map[int32]int32
	best                       map[int32]*Cut_Cut_t
	budget                     int32
	xor                        bool
	recorded                   []Cut_TruthRecord
	simulation                 bool
	stats                      Cut_Stats
	stopped                    bool
}

func Cut_NodeSign(node int32) uint32 {
	if node < 0 {
		panic("cut: negative leaf")
	}
	return 1 << uint(node%31)
}
func Cut_TruthWords(vars int32) int32 {
	if vars < 0 || vars > CUT_SIZE_MAX {
		panic("cut: unsupported truth dimension")
	}
	if vars <= 5 {
		return 1
	}
	return 1 << uint(vars-5)
}
func (p *Cut_Man_t) Stats() Cut_Stats { return p.stats }
func (p *Cut_Man_t) check() {
	if p == nil || p.stopped {
		panic("cut: manager stopped")
	}
}

type Cut_TruthRecord struct {
	Leaves []int32
	Truth  []uint32
}

func (p *Cut_Man_t) RecordedTruths() []Cut_TruthRecord {
	records := make([]Cut_TruthRecord, len(p.recorded))
	for i, r := range p.recorded {
		records[i] = Cut_TruthRecord{Leaves: append([]int32(nil), r.Leaves...), Truth: append([]uint32(nil), r.Truth...)}
	}
	return records
}

func cutRecord(list *Cut_Cut_t) Cut_Record {
	var record Cut_Record
	for c := list; c != nil; c = c.Next {
		step := Cut_Step{Indices: [2]uint16{c.Num0, c.Num1}, Complement: c.FCompl, Simulation: c.FSimul}
		if c.proof != nil {
			step.Proof = *c.proof
		}
		record.Steps = append(record.Steps, step)
	}
	return record
}
