// Proof state management types for interactive proof sessions.
// Ported from Python's proof.py.
package proof

import (
	lg "github.com/glycerine/goivy/logic"
)

// ProofGoal represents a single proof obligation.
type ProofGoal struct {
	Formula lg.Expr
	Node    interface{} // ReachabilityNode in the analysis graph
	Parent  *ProofGoal
	ID      int
}

// ProofGoalStack manages a stack of proof goals.
type ProofGoalStack struct {
	Stack []*ProofGoal
}

// NewProofGoalStack creates an empty goal stack.
func NewProofGoalStack() *ProofGoalStack {
	return &ProofGoalStack{}
}

// Push adds a goal to the stack.
func (s *ProofGoalStack) Push(goal *ProofGoal) {
	goal.ID = len(s.Stack)
	goal.Parent = s.Top()
	s.Stack = append(s.Stack, goal)
}

// Pop removes and returns the top goal.
func (s *ProofGoalStack) Pop() *ProofGoal {
	if len(s.Stack) == 0 {
		return nil
	}
	top := s.Stack[len(s.Stack)-1]
	s.Stack = s.Stack[:len(s.Stack)-1]
	top.ID = -1
	return top
}

// Top returns the top goal without removing it.
func (s *ProofGoalStack) Top() *ProofGoal {
	if len(s.Stack) == 0 {
		return nil
	}
	return s.Stack[len(s.Stack)-1]
}

// Remove removes a specific goal from the stack.
func (s *ProofGoalStack) Remove(goal *ProofGoal) {
	for i, g := range s.Stack {
		if g == goal {
			goal.ID = -1
			s.Stack = append(s.Stack[:i], s.Stack[i+1:]...)
			return
		}
	}
}

// Len returns the number of goals.
func (s *ProofGoalStack) Len() int {
	return len(s.Stack)
}

// --- Reachability Graph ---

// ReachabilityGraph tracks reachable states during analysis.
type ReachabilityGraph struct {
	Nodes []*ReachabilityNode
	Edges []*ReachabilityEdge
}

// NewReachabilityGraph creates an empty reachability graph.
func NewReachabilityGraph() *ReachabilityGraph {
	return &ReachabilityGraph{}
}

// AddNode adds a new node to the graph.
func (g *ReachabilityGraph) AddNode(state interface{}) *ReachabilityNode {
	node := &ReachabilityNode{
		Graph: g,
		State: state,
		ID:    len(g.Nodes),
	}
	g.Nodes = append(g.Nodes, node)
	return node
}

// AddEdge adds a directed edge to the graph.
func (g *ReachabilityGraph) AddEdge(source, target *ReachabilityNode, action interface{}) *ReachabilityEdge {
	edge := &ReachabilityEdge{
		Graph:  g,
		Source: source,
		Target: target,
		Action: action,
	}
	g.Edges = append(g.Edges, edge)
	return edge
}

// ReachabilityNode represents a state in the reachability graph.
type ReachabilityNode struct {
	Graph *ReachabilityGraph
	State interface{}
	ID    int
}

// ReachabilityEdge represents a transition in the reachability graph.
type ReachabilityEdge struct {
	Graph  *ReachabilityGraph
	Source *ReachabilityNode
	Target *ReachabilityNode
	Action interface{}
}

// --- Abstract / Concrete states ---

// AbstractState represents an abstract state as a set of facts.
type AbstractState struct {
	Facts []lg.Expr
}

// NewAbstractState creates an abstract state from facts.
func NewAbstractState(facts []lg.Expr) *AbstractState {
	result := make([]lg.Expr, len(facts))
	copy(result, facts)
	return &AbstractState{Facts: result}
}

// ConcreteState represents a concrete state (model assignment).
type ConcreteState struct {
	Values map[string]interface{}
}

// NewConcreteState creates a concrete state.
func NewConcreteState(values map[string]interface{}) *ConcreteState {
	return &ConcreteState{Values: values}
}

// ProofManager orchestrates proof sessions.
type ProofManager struct {
	Goals *ProofGoalStack
	Graph *ReachabilityGraph
}

// NewProofManager creates a new proof manager.
func NewProofManager() *ProofManager {
	return &ProofManager{
		Goals: NewProofGoalStack(),
		Graph: NewReachabilityGraph(),
	}
}
