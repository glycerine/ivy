//go:build js && wasm

package main

import (
	"context"
	"encoding/json"
	"syscall/js"

	"github.com/glycerine/ivy/goivy/webengine"
)

type wasmRequest struct {
	Type      string            `json:"type"`
	RequestID string            `json:"requestId"`
	ProjectID string            `json:"projectId,omitempty"`
	SessionID string            `json:"sessionId,omitempty"`
	Model     wasmModelDocument `json:"model,omitempty"`
	Intent    wasmCommandIntent `json:"intent,omitempty"`
}

type wasmModelDocument struct {
	ID             string `json:"id"`
	ProjectID      string `json:"projectId"`
	Filename       string `json:"filename"`
	Text           string `json:"text"`
	EngineRevision int64  `json:"engineRevision"`
}

type wasmCommandIntent struct {
	CommandID string         `json:"commandId"`
	Args      map[string]any `json:"args,omitempty"`
}

type wasmResponse struct {
	Type      string `json:"type"`
	RequestID string `json:"requestId,omitempty"`
	Error     string `json:"error,omitempty"`
	Value     any    `json:"value,omitempty"`
}

var engine = webengine.New(nil)

func main() {
	js.Global().Set("goivyWebEngineDispatch", js.FuncOf(dispatch))
	select {}
}

func dispatch(_ js.Value, args []js.Value) any {
	if len(args) == 0 {
		return encode(wasmResponse{Type: "error", Error: "missing request JSON"})
	}
	var req wasmRequest
	if err := json.Unmarshal([]byte(args[0].String()), &req); err != nil {
		return encode(wasmResponse{Type: "error", Error: err.Error()})
	}
	resp := handle(req)
	resp.RequestID = req.RequestID
	return encode(resp)
}

func handle(req wasmRequest) wasmResponse {
	ctx := context.Background()
	switch req.Type {
	case "init":
		return wasmResponse{Type: "ready", Value: map[string]bool{"ok": true}}
	case "new-session":
		session, err := engine.NewSession(ctx)
		if err != nil {
			return wasmError(err)
		}
		return wasmResponse{Type: "session", Value: session}
	case "load-model":
		result, err := engine.LoadModel(ctx, req.SessionID, req.Model.Filename, []byte(req.Model.Text))
		if err != nil {
			return wasmError(err)
		}
		return wasmResponse{Type: "load-result", Value: result}
	case "run-command":
		if req.Intent.CommandID == "check.induction" || req.Intent.CommandID == "check.bounded" || req.Intent.CommandID == "check.pdr" || req.Intent.CommandID == "check.concrete" {
			result, err := engine.Check(ctx, req.SessionID, webengine.CheckRequest{
				Mode:  checkMode(req.Intent.CommandID),
				Bound: intArg(req.Intent.Args["bound"]),
			})
			if err != nil {
				return wasmError(err)
			}
			return wasmResponse{Type: "check-result", Value: result}
		}
		result, err := engine.Action(ctx, req.SessionID, webengine.ActionRequest{
			Action: req.Intent.CommandID,
			Args:   req.Intent.Args,
		})
		if err != nil {
			return wasmError(err)
		}
		return wasmResponse{Type: "action-result", Value: result}
	case "get-snapshot":
		arg, err := engine.ARG(ctx, req.SessionID)
		if err != nil {
			return wasmError(err)
		}
		concept, err := engine.Concept(ctx, req.SessionID, webengine.ConceptRequest{})
		if err != nil {
			return wasmError(err)
		}
		return wasmResponse{Type: "snapshot", Value: map[string]any{"arg": arg, "concept": concept}}
	default:
		return wasmResponse{Type: "error", Error: "unknown request type: " + req.Type}
	}
}

func wasmError(err error) wasmResponse {
	return wasmResponse{Type: "error", Error: err.Error()}
}

func encode(resp wasmResponse) string {
	data, err := json.Marshal(resp)
	if err != nil {
		return `{"type":"error","error":"failed to encode response"}`
	}
	return string(data)
}

func checkMode(commandID string) string {
	switch commandID {
	case "check.induction":
		return "induction"
	case "check.bounded":
		return "bounded"
	case "check.pdr":
		return "pdr"
	case "check.concrete":
		return "concrete"
	default:
		return "pdr"
	}
}

func intArg(value any) int {
	switch typed := value.(type) {
	case float64:
		return int(typed)
	case int:
		return typed
	default:
		return 0
	}
}
