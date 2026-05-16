package webengine

import (
	"context"
	"encoding/json"
	"fmt"

	goivy "github.com/glycerine/ivy/goivy"
	"github.com/glycerine/ivy/goivy/webui"
)

type Engine struct {
	cfg     *goivy.Config
	backend webui.Backend
}

type SessionInfo struct {
	ID string `json:"id"`
}

type LoadResult struct {
	Status   string   `json:"status"`
	Filename string   `json:"filename,omitempty"`
	Isolate  string   `json:"isolate,omitempty"`
	Isolates []string `json:"isolates,omitempty"`
}

type CheckRequest struct {
	Mode  string `json:"mode"`
	Bound int    `json:"bound,omitempty"`
}

type CheckResult struct {
	Status                string   `json:"status"`
	Result                string   `json:"result"`
	Mode                  string   `json:"mode"`
	Message               string   `json:"message"`
	FailedConjecture      string   `json:"failed_conjecture,omitempty"`
	FailedLabel           string   `json:"failed_label,omitempty"`
	UsedRelations         []string `json:"used_relations,omitempty"`
	Z3Contacted           bool     `json:"z3_contacted,omitempty"`
	CounterexampleTrace   string   `json:"counterexample_trace,omitempty"`
	CounterexampleDetails string   `json:"counterexample_details,omitempty"`
	TraceARG              Payload  `json:"trace_arg,omitempty"`
}

type ConceptRequest struct {
	SheetID string `json:"sheet_id,omitempty"`
	NodeID  string `json:"node_id,omitempty"`
}

type ActionRequest struct {
	Action string         `json:"action"`
	Args   map[string]any `json:"args,omitempty"`
}

type ArgActionRequest struct {
	Node   string         `json:"node"`
	Action string         `json:"action"`
	Args   map[string]any `json:"args,omitempty"`
}

type ProofActionRequest struct {
	Goal   string `json:"goal"`
	Action string `json:"action"`
}

type Payload map[string]any

func New(cfg *goivy.Config, backend ...webui.Backend) *Engine {
	if cfg == nil {
		cfg = goivy.NewConfig()
	}
	var be webui.Backend
	if len(backend) > 0 && backend[0] != nil {
		be = backend[0]
	} else {
		be = webui.NewGoBackend(cfg)
	}
	return &Engine{cfg: cfg, backend: be}
}

func (e *Engine) NewSession(ctx context.Context) (SessionInfo, error) {
	if err := ctx.Err(); err != nil {
		return SessionInfo{}, err
	}
	var raw struct {
		SessionID string `json:"session_id"`
	}
	data, backendErr := e.backend.NewSession(e.cfg)
	if err := decode(data, backendErr, &raw); err != nil {
		return SessionInfo{}, err
	}
	if raw.SessionID == "" {
		return SessionInfo{}, fmt.Errorf("webengine: backend returned empty session_id")
	}
	return SessionInfo{ID: raw.SessionID}, nil
}

func (e *Engine) LoadModel(ctx context.Context, sessionID, filename string, content []byte, isolate ...string) (LoadResult, error) {
	if err := ctx.Err(); err != nil {
		return LoadResult{}, err
	}
	selectedIsolate := ""
	if len(isolate) > 0 {
		selectedIsolate = isolate[0]
	}
	var result LoadResult
	data, backendErr := e.backend.Load(sessionID, filename, content, selectedIsolate)
	if err := decode(data, backendErr, &result); err != nil {
		return LoadResult{}, err
	}
	return result, nil
}

func (e *Engine) Check(ctx context.Context, sessionID string, req CheckRequest) (CheckResult, error) {
	if err := ctx.Err(); err != nil {
		return CheckResult{}, err
	}
	mode := req.Mode
	if mode == "" {
		mode = "pdr"
	}
	var result CheckResult
	data, backendErr := e.backend.Check(sessionID, mode, webui.CheckOptions{Bound: req.Bound, Context: ctx})
	if err := decode(data, backendErr, &result); err != nil {
		return CheckResult{}, err
	}
	return result, nil
}

func (e *Engine) ARG(ctx context.Context, sessionID string) (Payload, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return decodePayload(e.backend.GetARG(sessionID, false))
}

func (e *Engine) CTIARG(ctx context.Context, sessionID string) (Payload, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return decodePayload(e.backend.GetCTIARG(sessionID))
}

func (e *Engine) Concept(ctx context.Context, sessionID string, req ConceptRequest) (Payload, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return decodePayload(e.backend.GetConcept(sessionID, req.SheetID, req.NodeID))
}

func (e *Engine) Toggles(ctx context.Context, sessionID string) (Payload, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return decodePayload(e.backend.GetToggles(sessionID))
}

func (e *Engine) Menus(ctx context.Context) (Payload, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	data, err := json.Marshal(webui.BuildBrowserMenuDescriptors())
	if err != nil {
		return nil, err
	}
	var payload Payload
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, fmt.Errorf("webengine: decode menu json: %w", err)
	}
	return payload, nil
}

func (e *Engine) Action(ctx context.Context, sessionID string, req ActionRequest) (Payload, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return decodePayload(e.backend.Action(sessionID, req.Action, req.Args))
}

func (e *Engine) ArgAction(ctx context.Context, sessionID string, req ArgActionRequest) (Payload, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return decodePayload(e.backend.ArgAction(sessionID, req.Node, req.Action, req.Args))
}

func (e *Engine) ConceptSplit(ctx context.Context, sessionID, concept, splitBy string) (Payload, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return decodePayload(e.backend.ConceptSplit(sessionID, concept, splitBy))
}

func (e *Engine) ConceptEmpty(ctx context.Context, sessionID, concept string) (Payload, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return decodePayload(e.backend.ConceptEmpty(sessionID, concept))
}

func (e *Engine) ConceptRemove(ctx context.Context, sessionID, concept string) (Payload, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return decodePayload(e.backend.ConceptRemove(sessionID, concept))
}

func (e *Engine) ConceptUndo(ctx context.Context, sessionID string) (Payload, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return decodePayload(e.backend.ConceptUndo(sessionID))
}

func (e *Engine) ConceptMaterialize(ctx context.Context, sessionID string, req webui.ConceptMaterializeRequest) (Payload, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return decodePayload(e.backend.ConceptMaterialize(sessionID, req))
}

func (e *Engine) ConceptReset(ctx context.Context, sessionID string) (Payload, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return decodePayload(e.backend.ConceptReset(sessionID))
}

func (e *Engine) ConceptDiagram(ctx context.Context, sessionID string) (Payload, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return decodePayload(e.backend.ConceptDiagram(sessionID))
}

func (e *Engine) ConceptProjection(ctx context.Context, sessionID, name, concept string) (Payload, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return decodePayload(e.backend.ConceptProjection(sessionID, name, concept))
}

func (e *Engine) SetToggle(ctx context.Context, sessionID, name, displayClass string, value bool) (Payload, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return decodePayload(e.backend.SetToggle(sessionID, name, displayClass, value))
}

func (e *Engine) Proof(ctx context.Context, sessionID string) (Payload, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return decodePayload(e.backend.GetProof(sessionID))
}

func (e *Engine) ProofAction(ctx context.Context, sessionID string, req ProofActionRequest) (Payload, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return decodePayload(e.backend.ProofAction(sessionID, req.Goal, req.Action))
}

func (e *Engine) Events(ctx context.Context, sessionID string) (<-chan webui.Event, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return e.backend.Events(sessionID)
}

func (e *Engine) Close() error {
	return e.backend.Close()
}

func decode(data []byte, backendErr error, out any) error {
	if backendErr != nil {
		return backendErr
	}
	if err := json.Unmarshal(data, out); err != nil {
		return fmt.Errorf("webengine: decode backend json: %w", err)
	}
	return nil
}

func decodePayload(data []byte, backendErr error) (Payload, error) {
	var payload Payload
	if err := decode(data, backendErr, &payload); err != nil {
		return nil, err
	}
	return payload, nil
}
