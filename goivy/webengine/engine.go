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
	Status   string `json:"status"`
	Filename string `json:"filename,omitempty"`
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

func (e *Engine) LoadModel(ctx context.Context, sessionID, filename string, content []byte) (LoadResult, error) {
	if err := ctx.Err(); err != nil {
		return LoadResult{}, err
	}
	var result LoadResult
	data, backendErr := e.backend.Load(sessionID, filename, content)
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
	data, backendErr := e.backend.Check(sessionID, mode, webui.CheckOptions{Bound: req.Bound})
	if err := decode(data, backendErr, &result); err != nil {
		return CheckResult{}, err
	}
	return result, nil
}

func (e *Engine) ARG(ctx context.Context, sessionID string) (Payload, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return decodePayload(e.backend.GetARG(sessionID))
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
