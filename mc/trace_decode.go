package mc

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	lg "github.com/glycerine/goivy/logic"
	tr "github.com/glycerine/goivy/transrel"
)

// AigerMatchHandler evaluates conditions and decodes state from an AIGER
// simulation trace back into Ivy terms.
//
// Python: ivy_mc.py:1445-1493
type AigerMatchHandler struct {
	Aiger    *Encoder
	Decoder  map[string]lg.Node
	Consts   map[string]bool
	StVarSet map[string]bool
	Current  map[string]lg.Node // current state as symbol→value
}

// NewAigerMatchHandler creates a new handler for decoding AIGER traces.
func NewAigerMatchHandler(aiger *Encoder, decoder map[string]lg.Node, consts map[string]bool, stVarSet map[string]bool) *AigerMatchHandler {
	return &AigerMatchHandler{
		Aiger:    aiger,
		Decoder:  decoder,
		Consts:   consts,
		StVarSet: stVarSet,
		Current:  make(map[string]lg.Node),
	}
}

// Eval evaluates a condition in the AIGER simulation context.
func (h *AigerMatchHandler) Eval(cond lg.Node) bool {
	if isFalseNode(cond) {
		return false
	}
	if isTrueNode(cond) {
		return true
	}
	// Try to get the symbol value from the AIGER circuit
	if c, ok := cond.(*lg.Const); ok {
		lit, ok := h.Aiger.Lit(c.Name)
		if ok && len(lit) > 0 {
			return lit[0] == h.Aiger.Sub.True()
		}
	}
	// Default: assume true
	return true
}

// Handle processes an action, extracting state from the AIGER simulation.
func (h *AigerMatchHandler) Handle(action interface{}, env map[string]string) {
	// Build inverse environment for renaming
	invEnv := make(map[string]string)
	for k, v := range env {
		if !h.isSkolem(k) {
			invEnv[v] = k
		}
	}

	// Show input symbols
	for _, v := range h.Aiger.Inputs {
		if decoded, ok := h.Decoder[v]; ok {
			h.showSym(v, decoded, h.getSymValue(v), invEnv, env)
		}
	}

	// Show next-state latch symbols
	for _, v := range h.Aiger.Latches {
		if decoded, ok := h.Decoder[v]; ok {
			h.showSym(v, decoded, h.getSymValue(v), invEnv, env)
		}
	}
}

// DoReturn handles a return action.
func (h *AigerMatchHandler) DoReturn(action interface{}, env map[string]string) {
	// no-op for basic handler
}

func (h *AigerMatchHandler) isSkolem(name string) bool {
	return tr.IsSkolem(name) && !h.Consts[name]
}

func (h *AigerMatchHandler) getSymValue(name string) lg.Node {
	lit, ok := h.Aiger.Lit(name)
	if !ok || len(lit) == 0 {
		return nil
	}
	if lit[0] == h.Aiger.Sub.True() {
		return &lg.And{Terms: nil} // true
	}
	return &lg.Or{Terms: nil} // false
}

func (h *AigerMatchHandler) showSym(v string, decoded lg.Node, val lg.Node, invEnv map[string]string, env map[string]string) {
	if val == nil {
		return
	}
	key := fmt.Sprint(decoded)
	h.Current[key] = val
}

// AigerMatchHandler2 is an enhanced handler that collects state equations
// for trace reconstruction as an AnalysisGraph.
//
// Python: ivy_mc.py:1495-1568
type AigerMatchHandler2 struct {
	Aiger    *Encoder
	Decoder  map[string]lg.Node
	Consts   map[string]bool
	StVarSet map[string]bool
	Current  map[string]lg.Node
	States   [][]lg.Node // list of state equations per step
}

// NewAigerMatchHandler2 creates an enhanced handler for trace reconstruction.
func NewAigerMatchHandler2(aiger *Encoder, decoder map[string]lg.Node, consts map[string]bool, stVarSet map[string]bool) *AigerMatchHandler2 {
	return &AigerMatchHandler2{
		Aiger:    aiger,
		Decoder:  decoder,
		Consts:   consts,
		StVarSet: stVarSet,
		Current:  make(map[string]lg.Node),
	}
}

// Eval evaluates a condition.
func (h *AigerMatchHandler2) Eval(cond lg.Node) bool {
	if isFalseNode(cond) {
		return false
	}
	if isTrueNode(cond) {
		return true
	}
	if n, ok := cond.(*lg.Not); ok {
		return !h.Eval(n.Body)
	}
	if c, ok := cond.(*lg.Const); ok {
		lit, ok := h.Aiger.Lit(c.Name)
		if ok && len(lit) > 0 {
			return lit[0] == h.Aiger.Sub.True()
		}
	}
	return true
}

// NewState collects equations for the current simulation state.
func (h *AigerMatchHandler2) NewState(env map[string]string) {
	invEnv := make(map[string]string)
	for k, v := range env {
		if !h.isSkolem(k) && !tr.IsNew(k) {
			invEnv[v] = k
		}
	}

	var eqns []lg.Node

	// Input symbols
	for _, v := range h.Aiger.Inputs {
		if decoded, ok := h.Decoder[v]; ok {
			val := h.getSymValue(v)
			if val != nil {
				eqns = append(eqns, &lg.Eq{T1: decoded, T2: val})
			}
		}
	}

	// Latch symbols (current and next state)
	for _, v := range h.Aiger.Latches {
		if decoded, ok := h.Decoder[v]; ok {
			val := h.getSymValue(v)
			if val != nil {
				eqns = append(eqns, &lg.Eq{T1: decoded, T2: val})
			}
		}
	}

	h.States = append(h.States, eqns)
}

// FinalState collects the final state after advancing the simulation.
func (h *AigerMatchHandler2) FinalState() {
	h.Aiger.Sub.Advance()

	var stvals []lg.Node
	for _, v := range h.Aiger.Latches {
		if decoded, ok := h.Decoder[v]; ok {
			if c, ok2 := decoded.(*lg.Const); ok2 && c.Name == "__init" {
				continue
			}
			val := h.getSymValue(v)
			if val != nil {
				stvals = append(stvals, &lg.Eq{T1: decoded, T2: val})
			}
		}
	}
	h.States = append(h.States, stvals)
}

func (h *AigerMatchHandler2) isSkolem(name string) bool {
	return tr.IsSkolem(name) && !h.Consts[name]
}

func (h *AigerMatchHandler2) getSymValue(name string) lg.Node {
	lit, ok := h.Aiger.Lit(name)
	if !ok || len(lit) == 0 {
		return nil
	}
	if lit[0] == h.Aiger.Sub.True() {
		return &lg.And{Terms: nil} // true
	}
	return &lg.Or{Terms: nil} // false
}

// DoReturn is a no-op for trace handler.
func (h *AigerMatchHandler2) DoReturn(action interface{}, env map[string]string) {}

// AigerWitnessToIvyTrace2 decodes an AIGER witness file into an Ivy trace.
// It simulates the AIGER circuit with witness inputs and collects state
// at each step.
//
// Python: ivy_mc.py:1630-1669
func AigerWitnessToIvyTrace2(
	result *ToAigerResult,
	witnessFilename string,
) (*AigerMatchHandler2, error) {
	f, err := os.Open(witnessFilename)
	if err != nil {
		return nil, fmt.Errorf("cannot open witness file: %w", err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	if !scanner.Scan() {
		return nil, fmt.Errorf("witness file is empty")
	}
	first := strings.TrimSpace(scanner.Text())
	if first != "1" {
		return nil, fmt.Errorf("expected '1' on first line, got %q", first)
	}

	// Collect all lines
	var lines []string
	for scanner.Scan() {
		line := strings.TrimRight(scanner.Text(), "\n\r")
		if line != "" {
			lines = append(lines, line)
		}
	}

	result.Aiger.Sub.Reset()

	handler := NewAigerMatchHandler2(result.Aiger, result.Decoder, result.Consts, result.StVarSet)

	for count, line := range lines {
		cols := strings.Split(line, " ")
		if len(cols) != 4 {
			return nil, fmt.Errorf("bad witness line %d: expected 4 columns, got %d", count, len(cols))
		}
		inp := cols[1]
		result.Aiger.Sub.Step(inp)

		// Check if this is the last step and invariant fails
		if count == len(lines)-1 {
			invarFail := lg.NewConst("invar__fail", lg.Boolean)
			lit, ok := result.Aiger.Lit(invarFail.Name)
			if ok && len(lit) > 0 && lit[0] == result.Aiger.Sub.True() {
				break
			}
		}

		// Collect state at this step
		handler.NewState(make(map[string]string))

		result.Aiger.Sub.Advance()
	}

	return handler, nil
}

// isFalseNode checks if a node is the logical false constant.
func isFalseNode(n lg.Node) bool {
	if o, ok := n.(*lg.Or); ok && len(o.Terms) == 0 {
		return true
	}
	if c, ok := n.(*lg.Const); ok && c.Name == "false" {
		return true
	}
	return false
}

// isTrueNode checks if a node is the logical true constant.
func isTrueNode(n lg.Node) bool {
	if a, ok := n.(*lg.And); ok && len(a.Terms) == 0 {
		return true
	}
	if c, ok := n.(*lg.Const); ok && c.Name == "true" {
		return true
	}
	return false
}
