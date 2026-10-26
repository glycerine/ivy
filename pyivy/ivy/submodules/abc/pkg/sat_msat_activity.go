package abc

// Variable and clause activity control from src/sat/msat/msatActivity.c.

import "math"

type MsatActivityConfig struct {
	VariableIncrement float64
	VariableDecay     float64
	ClauseIncrement   float64
	ClauseDecay       float64
}

type MsatActivitySnapshot struct {
	Config             MsatActivityConfig
	VariableActivities []float64
	ClauseActivities   []float32
}

// MsatActivityState connects one concrete decision heap and learned-clause
// database. It does not own either object; their stable identity is shared with
// the future solver core.
type MsatActivityState struct {
	order   *MsatVariableOrder
	learned *MsatClauseVector
	config  MsatActivityConfig
	closed  bool
}

func NewMsatActivityState(order *MsatVariableOrder, learned *MsatClauseVector, config MsatActivityConfig) (*MsatActivityState, error) {
	if order == nil || learned == nil || order.checkLive() != nil || learned.checkLive() != nil ||
		config.VariableIncrement < 0 || math.IsNaN(config.VariableIncrement) || math.IsInf(config.VariableIncrement, 0) ||
		math.IsNaN(config.VariableDecay) || math.IsInf(config.VariableDecay, 0) ||
		config.ClauseIncrement < 0 || math.IsNaN(config.ClauseIncrement) || math.IsInf(config.ClauseIncrement, 0) ||
		config.ClauseDecay < 0 || math.IsNaN(config.ClauseDecay) || math.IsInf(config.ClauseDecay, 0) {
		return nil, ErrDimension
	}
	return &MsatActivityState{order: order, learned: learned, config: config}, nil
}

func (s *MsatActivityState) checkLive() error {
	if s == nil {
		return ErrDimension
	}
	if s.closed || s.order == nil || s.learned == nil || s.order.checkLive() != nil || s.learned.checkLive() != nil {
		return ErrInvariant
	}
	return nil
}

func (s *MsatActivityState) VariableBump(literal SatLit) error {
	if err := s.checkLive(); err != nil {
		return err
	}
	if !literal.Valid() || uint64(literal.Var()) >= uint64(len(s.order.activities)) {
		return ErrInvalidLiteral
	}
	if s.config.VariableDecay < 0 {
		return nil
	}
	variable := literal.Var()
	if err := s.order.IncreaseActivity(variable, s.config.VariableIncrement); err != nil {
		return err
	}
	if s.order.activities[variable] > 1e100 {
		return s.VariableRescale()
	}
	return nil
}

func (s *MsatActivityState) VariableDecay() error {
	if err := s.checkLive(); err != nil {
		return err
	}
	if s.config.VariableDecay < 0 {
		return nil
	}
	next := s.config.VariableIncrement * s.config.VariableDecay
	if math.IsInf(next, 0) || math.IsNaN(next) {
		return ErrLimit
	}
	s.config.VariableIncrement = next
	return nil
}

func (s *MsatActivityState) VariableRescale() error {
	if err := s.checkLive(); err != nil {
		return err
	}
	next := s.config.VariableIncrement * 1e-100
	if math.IsInf(next, 0) || math.IsNaN(next) {
		return ErrLimit
	}
	if err := s.order.RescaleActivities(1e-100); err != nil {
		return err
	}
	s.config.VariableIncrement = next
	return nil
}

func (s *MsatActivityState) learnedClauses() ([]*MsatClause, error) {
	clauses := s.learned.Values()
	for _, clause := range clauses {
		if clause == nil || clause.closed || math.IsNaN(float64(clause.activity)) || math.IsInf(float64(clause.activity), 0) {
			return nil, ErrInvariant
		}
	}
	return clauses, nil
}

func (s *MsatActivityState) ClauseBump(clause *MsatClause) error {
	if err := s.checkLive(); err != nil {
		return err
	}
	if clause == nil || clause.closed || math.IsNaN(float64(clause.activity)) || math.IsInf(float64(clause.activity), 0) {
		return ErrDimension
	}
	rescale := float64(clause.activity)+s.config.ClauseIncrement > 1e20
	increment := s.config.ClauseIncrement
	base := clause.activity
	if rescale {
		clauses, err := s.learnedClauses()
		if err != nil {
			return err
		}
		increment *= 1e-20
		for _, learned := range clauses {
			if learned == clause {
				base *= float32(1e-20)
				break
			}
		}
	}
	delta := float32(increment)
	result := base + delta
	if math.IsInf(float64(delta), 0) || math.IsNaN(float64(delta)) || math.IsInf(float64(result), 0) || math.IsNaN(float64(result)) {
		return ErrLimit
	}
	if rescale {
		if err := s.ClauseRescale(); err != nil {
			return err
		}
	}
	clause.activity += delta
	return nil
}

func (s *MsatActivityState) ClauseDecay() error {
	if err := s.checkLive(); err != nil {
		return err
	}
	next := s.config.ClauseIncrement * s.config.ClauseDecay
	if math.IsInf(next, 0) || math.IsNaN(next) {
		return ErrLimit
	}
	s.config.ClauseIncrement = next
	return nil
}

func (s *MsatActivityState) ClauseRescale() error {
	if err := s.checkLive(); err != nil {
		return err
	}
	clauses, err := s.learnedClauses()
	if err != nil {
		return err
	}
	nextIncrement := s.config.ClauseIncrement * 1e-20
	if math.IsInf(nextIncrement, 0) || math.IsNaN(nextIncrement) {
		return ErrLimit
	}
	rescaled := make([]float32, len(clauses))
	for index, clause := range clauses {
		rescaled[index] = clause.activity * float32(1e-20)
		if math.IsInf(float64(rescaled[index]), 0) || math.IsNaN(float64(rescaled[index])) {
			return ErrLimit
		}
	}
	for index, clause := range clauses {
		clause.activity = rescaled[index]
	}
	s.config.ClauseIncrement = nextIncrement
	return nil
}

func (s *MsatActivityState) Snapshot() (MsatActivitySnapshot, error) {
	if err := s.checkLive(); err != nil {
		return MsatActivitySnapshot{}, err
	}
	clauses, err := s.learnedClauses()
	if err != nil {
		return MsatActivitySnapshot{}, err
	}
	result := MsatActivitySnapshot{Config: s.config, VariableActivities: s.order.Activities(), ClauseActivities: make([]float32, len(clauses))}
	for index, clause := range clauses {
		result.ClauseActivities[index] = clause.activity
	}
	return result, nil
}

func (s *MsatActivityState) Close() {
	if s == nil || s.closed {
		return
	}
	s.order, s.learned = nil, nil
	s.closed = true
}
