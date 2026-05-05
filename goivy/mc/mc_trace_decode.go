package mc

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/glycerine/ivy/goivy/actions"
	il "github.com/glycerine/ivy/goivy/ivylogic"
	lg "github.com/glycerine/ivy/goivy/logic"
	"github.com/glycerine/ivy/goivy/module"
)

// AigerMatchHandler evaluates conditions and decodes state from an AIGER
// simulation trace back into Ivy terms. Used for verbose path printing.
//
// Python: ivy_mc.py:1477-1526 (class AigerMatchHandler)
type AigerMatchHandler struct {
	Aiger    *Encoder
	Decoder  map[string]lg.Expr
	Consts   map[string]bool
	StVarSet map[string]bool
	Current  map[string]lg.Expr
}

// NewAigerMatchHandler creates a new handler for decoding AIGER traces.
func NewAigerMatchHandler(aiger *Encoder, decoder map[string]lg.Expr, consts map[string]bool, stVarSet map[string]bool) *AigerMatchHandler {
	return &AigerMatchHandler{
		Aiger:    aiger,
		Decoder:  decoder,
		Consts:   consts,
		StVarSet: stVarSet,
		Current:  make(map[string]lg.Expr),
	}
}

// Eval evaluates a condition in the AIGER simulation context.
func (h *AigerMatchHandler) Eval(cond lg.Expr) bool {
	if isFalseNode(cond) {
		return false
	}
	if isTrueNode(cond) {
		return true
	}
	if c, ok := cond.(*lg.Const); ok {
		return isTrueNode(h.Aiger.GetSym(c))
	}
	return true
}

// Handle processes an action, extracting state from the AIGER simulation.
func (h *AigerMatchHandler) Handle(action actions.Action, env map[lg.NodeKey]lg.Expr) {
	invEnv := make(map[string]string)
	for k, v := range env {
		if c, ok := v.(*lg.Const); ok {
			if !h.isSkolem(c.Name) {
				invEnv[c.Name] = string(k)
			}
		}
	}

	for _, v := range h.Aiger.Inputs {
		if decoded, ok := h.Decoder[v.Name]; ok {
			val := h.Aiger.GetSym(v)
			h.showSym(decoded, val)
		}
	}

	for _, v := range h.Aiger.Latches {
		if decoded, ok := h.Decoder[v.Name]; ok {
			val := h.Aiger.GetSym(v)
			h.showSym(decoded, val)
		}
	}
}

// DoReturn handles a return action.
func (h *AigerMatchHandler) DoReturn(action actions.Action, env map[lg.NodeKey]lg.Expr) {}

// Fail marks a failure point.
func (h *AigerMatchHandler) Fail() {}

func (h *AigerMatchHandler) isSkolem(name string) bool {
	return actions.IsSkolem(name) && !h.Consts[name]
}

func (h *AigerMatchHandler) showSym(decoded lg.Expr, val lg.Expr) {
	if val == nil {
		return
	}
	key := fmt.Sprint(decoded)
	h.Current[key] = val
}

// AigerMatchHandler2 is an enhanced handler that collects state equations
// for trace reconstruction. It implements actions.AnnotationHandler and
// replicates the TraceBase sub-handler delegation pattern.
//
// Python: ivy_mc.py:1527-1599 (class AigerMatchHandler2, extends TraceBase)
type AigerMatchHandler2 struct {
	Aiger    *Encoder
	Decoder  map[string]lg.Expr
	Consts   map[string]bool
	StVarSet map[string]bool
	Current  map[string]lg.Expr
	Mod      *module.Module

	// Trace-building state (replicates TraceBase pattern)
	LastAction  actions.Action
	Sub         *AigerMatchHandler2
	Returned    *AigerMatchHandler2
	IsFullTrace bool
	States      [][]lg.Expr
}

// NewAigerMatchHandler2 creates an enhanced handler for trace reconstruction.
func NewAigerMatchHandler2(
	aiger *Encoder,
	decoder map[string]lg.Expr,
	consts map[string]bool,
	stVarSet map[string]bool,
	mod *module.Module,
) *AigerMatchHandler2 {
	return &AigerMatchHandler2{
		Aiger:       aiger,
		Decoder:     decoder,
		Consts:      consts,
		StVarSet:    stVarSet,
		Current:     make(map[string]lg.Expr),
		Mod:         mod,
		IsFullTrace: true,
	}
}

// Eval evaluates a condition.
// Python: ivy_mc.py:1533-1543
func (h *AigerMatchHandler2) Eval(cond lg.Expr) bool {
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
		return isTrueNode(h.Aiger.GetSym(c))
	}
	return true
}

// Clone creates a copy of the handler for subcall tracking.
// Python: ivy_mc.py:1545-1546
func (h *AigerMatchHandler2) Clone() *AigerMatchHandler2 {
	return NewAigerMatchHandler2(h.Aiger, h.Decoder, h.Consts, h.StVarSet, h.Mod)
}

// Handle processes an action during trace construction.
// Replicates Python TraceBase.handle (ivy_trace.py:201-211).
func (h *AigerMatchHandler2) Handle(action actions.Action, env map[lg.NodeKey]lg.Expr) {
	if h.Sub != nil {
		h.Sub.Handle(action, env)
		return
	}
	if isCallOrEnvAction(h.LastAction) && h.Returned == nil {
		h.Sub = h.Clone()
		h.Sub.Handle(action, env)
		return
	}
	lineno := action.GetLineno()
	if lineno.Filename != "nowhere" {
		h.NewState(env)
		h.LastAction = action
	}
}

// DoReturn handles a return action. No-op for MC traces.
// Python: ivy_mc.py:1551-1552
func (h *AigerMatchHandler2) DoReturn(action actions.Action, env map[lg.NodeKey]lg.Expr) {
	if h.Sub != nil {
		if h.Sub.Sub != nil {
			h.Sub.DoReturn(action, env)
		} else {
			if isCallAction2(h.Sub.LastAction) && h.Sub.Returned == nil {
				h.Sub.DoReturn(action, env)
				return
			}
			h.Returned = h.Sub
			h.Sub = nil
			h.Returned.NewState(env)
		}
	} else if isCallAction2(h.LastAction) && h.Returned == nil {
		h.Sub = h.Clone()
		h.Handle(action, env)
		h.DoReturn(action, env)
	}
}

// Fail marks the last action as failed.
// Python: TraceBase.fail (ivy_trace.py:229-230)
func (h *AigerMatchHandler2) Fail() {
	h.LastAction = actions.NewFailAction(h.LastAction)
}

// End finishes the trace, returning from any unfinished calls.
// Python: TraceBase.end (ivy_trace.py:231-236)
func (h *AigerMatchHandler2) End() {
	if h.Sub != nil {
		h.Sub.End()
		h.Returned = h.Sub
		h.Sub = nil
	}
	h.FinalState()
}

// NewState builds state equations from the current AIGER simulation.
// Python: ivy_mc.py:1554-1585 AigerMatchHandler2.new_state
func (h *AigerMatchHandler2) NewState(env map[lg.NodeKey]lg.Expr) {
	if h.Aiger == nil {
		h.AddState(nil)
		return
	}
	invEnv := make(map[string]string)
	envNames := make(map[string]bool)
	for k, v := range env {
		if c, ok := v.(*lg.Const); ok {
			if !h.isSkolem(c.Name) && !actions.IsNew(c.Name) {
				invEnv[c.Name] = string(k)
			}
		}
		envNames[string(k)] = true
	}

	var eqns []lg.Expr

	for _, v := range h.Aiger.Inputs {
		if decd, ok := h.Decoder[v.Name]; ok {
			val := h.Aiger.GetSym(v)
			h.showSym2(decd, val, invEnv, envNames, &eqns)
		}
	}

	rn := make(map[string]string)
	for x := range h.StVarSet {
		rn[x] = "new_" + x
	}

	for _, v := range h.Aiger.Latches {
		if decd, ok := h.Decoder[v.Name]; ok {
			val := h.Aiger.GetSym(v)
			h.showSym2(decd, val, invEnv, envNames, &eqns)

			envNameMap := make(map[string]string)
			for k, v := range env {
				if c, ok := v.(*lg.Const); ok {
					envNameMap[string(k)] = c.Name
				}
			}
			nextDecd := module.RenameASTByName(decd, rn)
			curDecd := module.RenameASTByName(decd, envNameMap)
			if nextDecd != nil && curDecd != nil && nextDecd.Equal(curDecd) {
				nextVal := h.Aiger.GetNextSym(v)
				h.showSym2(nextDecd, nextVal, invEnv, envNames, &eqns)
			}
		}
	}

	h.AddState(eqns)
}

// showSym2 is the filtering/renaming logic for building trace equations.
// Python: ivy_mc.py:1559-1568 show_sym inner function.
func (h *AigerMatchHandler2) showSym2(
	decd lg.Expr,
	val lg.Expr,
	invEnv map[string]string,
	envNames map[string]bool,
	eqns *[]lg.Expr,
) {
	if val == nil {
		return
	}

	if app, ok := decd.(*lg.Apply); ok {
		if c, ok2 := app.Func.(*lg.Const); ok2 && strings.HasPrefix(c.Name, "__new_") {
			newName := strings.TrimPrefix(c.Name, "__")
			newFunc := lg.NewConst(newName, c.CSort)
			decd = lg.TryApply(newFunc, app.Terms...)
		}
	}

	syms := il.UsedSymbolsAst(decd)
	allOK := true
	for _, sym := range syms.All() {
		c, ok := sym.(*lg.Const)
		if !ok {
			continue
		}
		_, inInv := invEnv[c.Name]
		if !inInv {
			if h.isSkolem(c.Name) || actions.IsNew(c.Name) || envNames[c.Name] {
				allOK = false
				break
			}
		}
	}
	if !allOK {
		return
	}

	expr := module.RenameASTByName(decd, invEnv)

	if app, ok := expr.(*lg.Apply); ok {
		if c, ok2 := app.Func.(*lg.Const); ok2 && actions.IsNew(c.Name) {
			return
		}
	}

	if il.IsConstant(expr) {
		if c, ok := expr.(*lg.Const); ok {
			if h.Mod != nil && h.Mod.Sig != nil && h.Mod.Sig.Constructors[c.Name] {
				return
			}
		}
	}

	*eqns = append(*eqns, &lg.Eq{T1: expr, T2: val})
}

// FinalState advances the AIGER simulator and collects the final latch state.
// Python: ivy_mc.py:1587-1599 AigerMatchHandler2.final_state
func (h *AigerMatchHandler2) FinalState() {
	h.Aiger.Sub.Advance()

	post := h.Aiger.Sub.LatchVals()
	stmap := h.Aiger.GetEncoderState(post)

	var stvals []lg.Expr
	for _, v := range h.Aiger.Latches {
		if v.Name == "__init" {
			continue
		}
		if decd, ok := h.Decoder[v.Name]; ok {
			if val, ok2 := stmap[lg.Key(v)]; ok2 && val != nil {
				stvals = append(stvals, &lg.Eq{T1: decd, T2: val})
			}
		}
	}
	h.AddState(stvals)
}

// AddState stores a set of state equations as one trace step.
func (h *AigerMatchHandler2) AddState(eqns []lg.Expr) {
	h.States = append(h.States, eqns)
}

func (h *AigerMatchHandler2) isSkolem(name string) bool {
	return actions.IsSkolem(name) && !h.Consts[name]
}

// String returns a human-readable representation of the trace.
func (h *AigerMatchHandler2) String() string {
	var b strings.Builder
	for i, eqns := range h.States {
		fmt.Fprintf(&b, "state %d:\n", i)
		for _, eq := range eqns {
			fmt.Fprintf(&b, "    %v\n", eq)
		}
	}
	return b.String()
}

// AigerWitnessToIvyTrace2 decodes an AIGER witness file into an Ivy trace.
// It simulates the AIGER circuit with witness inputs, calls MatchAnnotation
// to walk the action/annotation tree, and collects state equations at each step.
//
// Python: ivy_mc.py:1662-1701
func AigerWitnessToIvyTrace2(
	result *ToAigerResult,
	witnessFilename string,
	mod *module.Module,
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
		return nil, fmt.Errorf("model checker returned mis-formatted witness")
	}

	var lines []string
	for scanner.Scan() {
		line := strings.TrimRight(scanner.Text(), "\n\r")
		if line != "" {
			lines = append(lines, line)
		}
	}

	result.Aiger.Sub.Reset()

	handler := NewAigerMatchHandler2(
		result.Aiger, result.Decoder, result.Consts, result.StVarSet, mod,
	)

	annot, _ := result.Annot.(actions.Annotation)

	for count, line := range lines {
		cols := strings.Split(line, " ")
		if len(cols) != 4 {
			return nil, fmt.Errorf("model checker returned mis-formatted witness")
		}
		inp := cols[1]
		result.Aiger.Sub.Step(inp)

		if count == len(lines)-1 {
			invarFail := lg.NewConst("invar__fail", lg.Boolean)
			sym := result.Aiger.GetSym(invarFail)
			if sym != nil && isTrueNode(sym) {
				break
			}
		}

		if annot != nil && result.Action != nil {
			actions.MatchAnnotation(result.Action, annot, handler, mod)
		}
		handler.End()
	}

	return handler, nil
}

func isCallOrEnvAction(action actions.Action) bool {
	if action == nil {
		return false
	}
	switch action.(type) {
	case *actions.CallAction, *actions.EnvAction:
		return true
	}
	return false
}

func isCallAction2(action actions.Action) bool {
	if action == nil {
		return false
	}
	_, ok := action.(*actions.CallAction)
	return ok
}

// isFalseNode checks if a node is the logical false constant.
func isFalseNode(n lg.Expr) bool {
	if o, ok := n.(*lg.Or); ok && len(o.Terms) == 0 {
		return true
	}
	if c, ok := n.(*lg.Const); ok && c.Name == "false" {
		return true
	}
	return false
}

// isTrueNode checks if a node is the logical true constant.
func isTrueNode(n lg.Expr) bool {
	if a, ok := n.(*lg.And); ok && len(a.Terms) == 0 {
		return true
	}
	if c, ok := n.(*lg.Const); ok && c.Name == "true" {
		return true
	}
	return false
}
