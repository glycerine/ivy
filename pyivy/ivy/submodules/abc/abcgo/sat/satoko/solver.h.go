// Copyright Satoko contributors. Distributed under the BSD 2-Clause License.
package satoko

import (
	"github.com/glycerine/abcgo/sat/satoko/utils"
	"github.com/glycerine/abcgo/sat/satoko/utils/vec"
	"time"
)

type Solver struct {
	opts                      Options
	stats                     Stats
	clauses                   []*clause
	watches                   [][]int
	values                    []int8
	levels, reasons           []int
	phases                    []uint8
	trail, limits             []int
	qhead                     int
	assumptions, final, model []int
	ok                        bool
	activity                  *vec.Activities
	order                     *utils.Heap
	actInc                    uint64
	clauseInc                 uint32
	deadline                  time.Time
	stop                      *int
	stopFunc                  func(int) int
	runID                     int
	bookmark                  *solverBookmark
	coneActive                bool
	marks                     []bool
}
type solverBookmark struct {
	clauses, variables int
	ok                 bool
}

func (s *Solver) initOrder()         { s.activity = vec.NewActivities(0); s.order = utils.NewHeap(s.activity) }
func (s *Solver) decisionLevel() int { return len(s.limits) }
func (s *Solver) literalValue(lit int) int8 {
	v := s.values[lit>>1]
	if v < 0 {
		return -1
	}
	return v ^ int8(lit&1)
}
func (s *Solver) enqueue(lit, reason int) bool {
	v := s.literalValue(lit)
	if v >= 0 {
		return v == 1
	}
	variable := lit >> 1
	s.values[variable] = 1 - int8(lit&1)
	s.levels[variable] = s.decisionLevel()
	s.reasons[variable] = reason
	s.trail = append(s.trail, lit)
	return true
}
func (s *Solver) cancel(level int) {
	if s.decisionLevel() <= level {
		return
	}
	at := s.limits[level]
	for _, lit := range s.trail[at:] {
		v := lit >> 1
		s.phases[v] = uint8(lit & 1)
		s.values[v] = -1
		s.reasons[v] = -1
		s.order.Update(uint32(v))
	}
	s.trail = s.trail[:at]
	s.limits = s.limits[:level]
	if s.qhead > at {
		s.qhead = at
	}
}
func (s *Solver) stopped() bool {
	return s.stop != nil && *s.stop != 0 || s.stopFunc != nil && s.stopFunc(s.runID) != 0 || !s.deadline.IsZero() && time.Now().After(s.deadline)
}
