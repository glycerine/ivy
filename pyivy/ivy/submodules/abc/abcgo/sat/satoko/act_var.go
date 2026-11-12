// Copyright Satoko contributors. Distributed under the BSD 2-Clause License.
package satoko

import "github.com/glycerine/abcgo/sat/satoko/utils"

func (s *Solver) bumpVariable(v int) {
	s.activity.Set(uint32(v), utils.ActivityAdd(s.activity.At(uint32(v)), s.actInc))
	if s.activity.At(uint32(v)) > s.opts.VariableLimit {
		for i := range s.values {
			s.activity.Set(uint32(i), utils.ActivityRescale(s.activity.At(uint32(i)), s.opts.VariableRescale))
		}
		s.actInc = utils.ActivityRescale(s.actInc, s.opts.VariableRescale)
	}
	if s.order.Contains(uint32(v)) {
		s.order.Update(uint32(v))
	}
}
func (s *Solver) decayVariables() {
	decay := s.opts.VariableDecay
	if decay <= 0 || decay > 1 {
		decay = 0.95
	}
	s.actInc = utils.ActivityMultiply(s.actInc, utils.ActivityFromFloat(1/decay))
}
