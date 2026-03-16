// Package iupdr provides an interactive UPDR (Universal Property-Directed
// Reachability) interface.
// This is a port of Python's iupdr.py (203 lines).
//
// The Python version uses IPython widgets for interactive proof exploration.
// In Go, this functionality is provided through the web UI instead.
// This package provides the core logic that the web UI calls into.
package iupdr

import (
	"fmt"

	"github.com/glycerine/goivy/art"
	"github.com/glycerine/goivy/clauseops"
	lg "github.com/glycerine/goivy/logic"
	"github.com/glycerine/goivy/module"
	"github.com/glycerine/goivy/proof"
	"github.com/glycerine/goivy/tactics"
)

// Session holds the state for an interactive UPDR session.
type Session struct {
	AG      *art.AnalysisGraph
	Mod     *module.Module
	TC      *tactics.TacticsContext
	Frames  []*art.State
	History []StepInfo
}

// StepInfo records information about a single interactive step.
type StepInfo struct {
	Type    string
	Goal    *proof.ProofGoal
	Active  *proof.ProofGoal
	Message string
	Data    map[string]interface{}
}

// NewSession creates a new interactive UPDR session.
func NewSession(mod *module.Module) *Session {
	ag := art.NewAnalysisGraph(mod)
	tc := tactics.NewTacticsContext(ag, mod)
	return &Session{
		AG:  ag,
		Mod: mod,
		TC:  tc,
	}
}

// Initialize sets up the initial state and adds it to the graph.
func (s *Session) Initialize(initClauses *clauseops.Clauses) {
	state := art.NewState(s.Mod, initClauses)
	s.AG.Add(state, nil)
	s.Frames = append(s.Frames, state)
}

// AddFrame adds a new frame by executing the big action.
func (s *Session) AddFrame() (*art.State, error) {
	if len(s.Frames) == 0 {
		return nil, fmt.Errorf("no initial frame")
	}
	lastFrame := s.Frames[len(s.Frames)-1]
	action := tactics.GetBigAction(s.AG)
	newFrame := s.AG.Execute(action, lastFrame, nil, "")
	if newFrame == nil {
		return nil, fmt.Errorf("failed to execute action")
	}
	s.Frames = append(s.Frames, newFrame)
	return newFrame, nil
}

// PushBadStates pushes the negation of the safety property as a goal
// at the given frame.
func (s *Session) PushBadStates(badStates lg.Node, frame *art.State) {
	goal := tactics.GoalAtArgNode(badStates, frame)
	s.TC.PushGoal(goal)
}

// Step performs one UPDR step: process the top goal.
// Returns step info and whether the proof is complete.
func (s *Session) Step() (*StepInfo, bool, error) {
	goal := s.TC.TopGoal()
	if goal == nil {
		return &StepInfo{Type: "no_goals", Message: "goal stack empty"}, false, nil
	}

	// Check if goal is refuted
	if s.TC.RefutedGoal(goal) {
		s.TC.RemoveGoal(goal)
		info := &StepInfo{
			Type:    "refuted",
			Goal:    goal,
			Message: "goal refuted and removed",
		}
		s.History = append(s.History, *info)
		return info, false, nil
	}

	// Check if at initial frame (counterexample found)
	if goalNode, ok := goal.Node.(*art.State); ok {
		if len(s.Frames) > 0 && goalNode == s.Frames[0] {
			info := &StepInfo{
				Type:    "counterexample",
				Goal:    goal,
				Message: "goal reached initial state — no invariant exists",
			}
			s.History = append(s.History, *info)
			return info, true, nil
		}
	}

	// Try refine or reverse
	refined, result := s.TC.RefineOrReverse(goal)
	if refined {
		// Learned new fact
		if newFact, ok := result.(lg.Node); ok {
			if goalNode, ok := goal.Node.(*art.State); ok {
				factClauses := clauseops.FormulaToClauses(newFact, nil)
				tactics.ArgAddFacts(goalNode, factClauses)
			}
		}
		// Remove refuted goals
		for s.TC.Goals.Len() > 0 {
			g := s.TC.TopGoal()
			if g == nil || !s.TC.RefutedGoal(g) {
				break
			}
			s.TC.RemoveGoal(g)
		}
		info := &StepInfo{
			Type:    "refined",
			Goal:    goal,
			Active:  s.TC.TopGoal(),
			Message: "learned new fact",
		}
		s.History = append(s.History, *info)
		return info, false, nil
	}

	// Push new goal from backward image
	if newGoal, ok := result.(*proof.ProofGoal); ok {
		s.TC.PushGoal(newGoal)
		info := &StepInfo{
			Type:    "reversed",
			Goal:    goal,
			Active:  newGoal,
			Message: "pushed new goal from backward image",
		}
		s.History = append(s.History, *info)
		return info, false, nil
	}

	return &StepInfo{Type: "stuck", Message: "cannot make progress"}, false, nil
}

// CheckInductive checks if any consecutive frame pair is inductive.
// Returns the frame index if found, -1 otherwise.
func (s *Session) CheckInductive() int {
	for i := 0; i < len(s.Frames)-1; i++ {
		if s.AG.Cover(s.Frames[i+1], s.Frames[i]) {
			return i
		}
	}
	return -1
}

// RunToCompletion runs the UPDR algorithm until completion or max iterations.
func (s *Session) RunToCompletion(maxIter int) (bool, error) {
	for iter := 0; iter < maxIter; iter++ {
		// Check for inductive invariant
		if idx := s.CheckInductive(); idx >= 0 {
			return true, nil // Found invariant
		}

		// Add new frame if goal stack is empty
		if s.TC.Goals.Len() == 0 {
			_, err := s.AddFrame()
			if err != nil {
				return false, err
			}
		}

		// Process goals
		info, done, err := s.Step()
		if err != nil {
			return false, err
		}
		if done {
			return false, fmt.Errorf("counterexample: %s", info.Message)
		}
	}
	return false, fmt.Errorf("reached iteration limit (%d)", maxIter)
}
