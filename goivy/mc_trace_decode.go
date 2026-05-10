package goivy

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
)

// AigerMatchHandler evaluates conditions and decodes state from an AIGER
// simulation trace back into Ivy terms. Used for verbose path printing.
//
// Python: ivy_mc.py:1477-1526 (class AigerMatchHandler)
type AigerMatchHandler struct {
	Aiger    *Encoder
	Decoder  map[string]Expr
	Consts   map[string]bool
	StVarSet map[string]bool
	Current  map[string]Expr
}

// NewAigerMatchHandler creates a new handler for decoding AIGER traces.
func NewAigerMatchHandler(aiger *Encoder, decoder map[string]Expr, consts map[string]bool, stVarSet map[string]bool) *AigerMatchHandler {
	return &AigerMatchHandler{
		Aiger:    aiger,
		Decoder:  decoder,
		Consts:   consts,
		StVarSet: stVarSet,
		Current:  make(map[string]Expr),
	}
}

// Eval evaluates a condition in the AIGER simulation context.
func (h *AigerMatchHandler) Eval(cond Expr) bool {
	if isFalseNode(cond) {
		return false
	}
	if isTrueNode(cond) {
		return true
	}
	if c, ok := cond.(*Const); ok {
		return isTrueNode(h.Aiger.GetSym(c))
	}
	return true
}

// Handle processes an action, extracting state from the AIGER simulation.
func (h *AigerMatchHandler) Handle(action ActionsAction, env map[NodeKey]Expr) {
	invEnv := make(map[string]string)
	for k, v := range env {
		if c, ok := v.(*Const); ok {
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
func (h *AigerMatchHandler) DoReturn(action ActionsAction, env map[NodeKey]Expr) {}

// Fail marks a failure point.
func (h *AigerMatchHandler) Fail() {}

func (h *AigerMatchHandler) isSkolem(name string) bool {
	return IsSkolem(name) && !h.Consts[name]
}

func (h *AigerMatchHandler) showSym(decoded Expr, val Expr) {
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
	Decoder  map[string]Expr
	Consts   map[string]bool
	StVarSet map[string]bool
	Current  map[string]Expr
	Mod      *Module

	// Trace-building state (replicates TraceBase pattern)
	LastAction  ActionsAction
	Sub         *AigerMatchHandler2
	Returned    *AigerMatchHandler2
	IsFullTrace bool
	States      [][]Expr
}

// NewAigerMatchHandler2 creates an enhanced handler for trace reconstruction.
func NewAigerMatchHandler2(
	aiger *Encoder,
	decoder map[string]Expr,
	consts map[string]bool,
	stVarSet map[string]bool,
	mod *Module,
) *AigerMatchHandler2 {
	return &AigerMatchHandler2{
		Aiger:       aiger,
		Decoder:     decoder,
		Consts:      consts,
		StVarSet:    stVarSet,
		Current:     make(map[string]Expr),
		Mod:         mod,
		IsFullTrace: true,
	}
}

// Eval evaluates a condition.
// Python: ivy_mc.py:1533-1543
func (h *AigerMatchHandler2) Eval(cond Expr) bool {
	if isFalseNode(cond) {
		return false
	}
	if isTrueNode(cond) {
		return true
	}
	if n, ok := cond.(*LogicNot); ok {
		return !h.Eval(n.Body)
	}
	if c, ok := cond.(*Const); ok {
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
func (h *AigerMatchHandler2) Handle(action ActionsAction, env map[NodeKey]Expr) {
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
func (h *AigerMatchHandler2) DoReturn(action ActionsAction, env map[NodeKey]Expr) {
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
	h.LastAction = NewFailAction(h.LastAction)
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
func (h *AigerMatchHandler2) NewState(env map[NodeKey]Expr) {
	if h.Aiger == nil {
		h.AddState(nil)
		return
	}
	invEnv := make(map[string]string)
	envNames := make(map[string]bool)
	for k, v := range env {
		if c, ok := v.(*Const); ok {
			if !h.isSkolem(c.Name) && !IsNew(c.Name) {
				invEnv[c.Name] = string(k)
			}
		}
		envNames[string(k)] = true
	}

	var eqns []Expr

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
				if c, ok := v.(*Const); ok {
					envNameMap[string(k)] = c.Name
				}
			}
			nextDecd := RenameASTByName(decd, rn)
			curDecd := RenameASTByName(decd, envNameMap)
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
	decd Expr,
	val Expr,
	invEnv map[string]string,
	envNames map[string]bool,
	eqns *[]Expr,
) {
	if val == nil {
		return
	}

	if app, ok := decd.(*Apply); ok {
		if c, ok2 := app.Func.(*Const); ok2 && strings.HasPrefix(c.Name, "__new_") {
			newName := strings.TrimPrefix(c.Name, "__")
			newFunc := NewConst(newName, c.CSort)
			decd = TryApply(newFunc, app.Terms...)
		}
	}

	syms := UsedSymbolsAst(decd)
	allOK := true
	for _, sym := range syms.All() {
		c, ok := sym.(*Const)
		if !ok {
			continue
		}
		_, inInv := invEnv[c.Name]
		if !inInv {
			if h.isSkolem(c.Name) || IsNew(c.Name) || envNames[c.Name] {
				allOK = false
				break
			}
		}
	}
	if !allOK {
		return
	}

	expr := RenameASTByName(decd, invEnv)

	if app, ok := expr.(*Apply); ok {
		if c, ok2 := app.Func.(*Const); ok2 && IsNew(c.Name) {
			return
		}
	}

	if IsConstant(expr) {
		if c, ok := expr.(*Const); ok {
			if h.Mod != nil && h.Mod.Sig != nil && h.Mod.Sig.Constructors[c.Name] {
				return
			}
		}
	}

	*eqns = append(*eqns, &Eq{T1: expr, T2: val})
}

// FinalState advances the AIGER simulator and collects the final latch state.
// Python: ivy_mc.py:1587-1599 AigerMatchHandler2.final_state
func (h *AigerMatchHandler2) FinalState() {
	h.Aiger.Sub.Advance()

	post := h.Aiger.Sub.LatchVals()
	stmap := h.Aiger.GetEncoderState(post)

	var stvals []Expr
	for _, v := range h.Aiger.Latches {
		if v.Name == "__init" {
			continue
		}
		if decd, ok := h.Decoder[v.Name]; ok {
			if val, ok2 := stmap[Key(v)]; ok2 && val != nil {
				stvals = append(stvals, &Eq{T1: decd, T2: val})
			}
		}
	}
	h.AddState(stvals)
}

// AddState stores a set of state equations as one trace step.
func (h *AigerMatchHandler2) AddState(eqns []Expr) {
	h.States = append(h.States, eqns)
}

func (h *AigerMatchHandler2) isSkolem(name string) bool {
	return IsSkolem(name) && !h.Consts[name]
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
	mod *Module,
) (*AigerMatchHandler2, error) {
	f, err := os.Open(witnessFilename)
	if err != nil {
		return nil, fmt.Errorf("cannot open witness file: %w", err)
	}
	defer f.Close()
	return AigerWitnessToIvyTrace2Reader(result, f, mod)
}

func AigerWitnessToIvyTrace2Bytes(result *ToAigerResult, witness []byte, mod *Module) (*AigerMatchHandler2, error) {
	return AigerWitnessToIvyTrace2Reader(result, bytes.NewReader(witness), mod)
}

func AigerWitnessToIvyTrace2Reader(result *ToAigerResult, r io.Reader, mod *Module) (*AigerMatchHandler2, error) {
	scanner := bufio.NewScanner(r)
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

	annot, _ := result.Annot.(Annotation)

	for count, line := range lines {
		cols := strings.Split(line, " ")
		if len(cols) != 4 {
			return nil, fmt.Errorf("model checker returned mis-formatted witness")
		}
		inp := cols[1]
		result.Aiger.Sub.Step(inp)

		if count == len(lines)-1 {
			invarFail := NewConst("invar__fail", Boolean)
			sym := result.Aiger.GetSym(invarFail)
			if sym != nil && isTrueNode(sym) {
				break
			}
		}

		if annot != nil && result.Action != nil {
			MatchAnnotation(result.Action, annot, handler, mod)
		}
		handler.End()
	}

	return handler, nil
}

func isCallOrEnvAction(action ActionsAction) bool {
	if action == nil {
		return false
	}
	switch action.(type) {
	case *LogicCallAction, *LogicEnvAction:
		return true
	}
	return false
}

func isCallAction2(action ActionsAction) bool {
	if action == nil {
		return false
	}
	_, ok := action.(*LogicCallAction)
	return ok
}

// isFalseNode checks if a node is the logical false constant.
func isFalseNode(n Expr) bool {
	if o, ok := n.(*LogicOr); ok && len(o.Terms) == 0 {
		return true
	}
	if c, ok := n.(*Const); ok && c.Name == "false" {
		return true
	}
	return false
}

// isTrueNode checks if a node is the logical true constant.
func isTrueNode(n Expr) bool {
	if a, ok := n.(*LogicAnd); ok && len(a.Terms) == 0 {
		return true
	}
	if c, ok := n.(*Const); ok && c.Name == "true" {
		return true
	}
	return false
}
