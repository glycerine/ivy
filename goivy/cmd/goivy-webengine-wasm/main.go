//go:build js && wasm

package main

import (
	"context"
	"encoding/json"
	"syscall/js"

	"github.com/glycerine/ivy/goivy/webengine"
	"github.com/glycerine/ivy/goivy/webui"
)

type wasmRequest struct {
	Type      string            `json:"type"`
	RequestID string            `json:"requestId"`
	ProjectID string            `json:"projectId,omitempty"`
	SessionID string            `json:"sessionId,omitempty"`
	Model     wasmModelDocument `json:"model,omitempty"`
	Intent    wasmCommandIntent `json:"intent,omitempty"`
	Snapshot  map[string]any    `json:"snapshot,omitempty"`
}

type wasmModelDocument struct {
	ID             string `json:"id"`
	ProjectID      string `json:"projectId"`
	Filename       string `json:"filename"`
	Text           string `json:"text"`
	Isolate        string `json:"isolate"`
	EngineRevision int64  `json:"engineRevision"`
}

type wasmCommandIntent struct {
	CommandID string         `json:"commandId"`
	Args      map[string]any `json:"args,omitempty"`
	Target    map[string]any `json:"target,omitempty"`
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
		result, err := engine.LoadModel(ctx, req.SessionID, req.Model.Filename, []byte(req.Model.Text), req.Model.Isolate)
		if err != nil {
			return wasmError(err)
		}
		return wasmResponse{Type: "load-result", Value: result}
	case "run-command":
		result, err := runCommand(ctx, req)
		if err != nil {
			return wasmError(err)
		}
		return wasmResponse{Type: "command-result", Value: result}
	case "get-snapshot":
		snapshot, err := getSnapshot(ctx, req)
		if err != nil {
			return wasmError(err)
		}
		return wasmResponse{Type: "snapshot", Value: snapshot}
	default:
		return wasmResponse{Type: "error", Error: "unknown request type: " + req.Type}
	}
}

func runCommand(ctx context.Context, req wasmRequest) (any, error) {
	commandID := req.Intent.CommandID
	args := req.Intent.Args
	if args == nil {
		args = map[string]any{}
	}
	if commandID == "check.induction" || commandID == "check.bounded" || commandID == "check.pdr" || commandID == "check.concrete" || commandID == "check.abstract" {
		return engine.Check(ctx, req.SessionID, webengine.CheckRequest{
			Mode:  checkMode(commandID),
			Bound: intArg(args["bound"]),
		})
	}
	switch commandID {
	case "concept.split":
		return engine.ConceptSplit(ctx, req.SessionID, stringArg(args["concept"]), firstStringArg(args, "splitBy", "split_by"))
	case "concept.empty":
		return engine.ConceptEmpty(ctx, req.SessionID, stringArg(args["concept"]))
	case "concept.remove":
		return engine.ConceptRemove(ctx, req.SessionID, stringArg(args["concept"]))
	case "concept.undo":
		return engine.ConceptUndo(ctx, req.SessionID)
	case "concept.materializeNode":
		return engine.ConceptMaterialize(ctx, req.SessionID, webuiConceptMaterializeRequest(args, "node"))
	case "concept.materializeEdge":
		return engine.ConceptMaterialize(ctx, req.SessionID, webuiConceptMaterializeRequest(args, "edge"))
	case "concept.projection":
		return engine.ConceptProjection(ctx, req.SessionID, stringArg(args["name"]), stringArg(args["concept"]))
	case "concept.reset":
		return engine.ConceptReset(ctx, req.SessionID)
	case "concept.diagram":
		return engine.ConceptDiagram(ctx, req.SessionID)
	case "toggles.set":
		name := firstStringArg(args, "edge", "label")
		displayClass := firstStringArg(args, "display_class", "displayClass", "class")
		return engine.SetToggle(ctx, req.SessionID, name, displayClass, boolArg(firstPresent(args, "value", "visible")))
	case "proof.action":
		return engine.ProofAction(ctx, req.SessionID, webengine.ProofActionRequest{
			Goal:   firstStringArg(req.Intent.Target, "goalId", "goal"),
			Action: stringArg(args["action"]),
		})
	default:
		if isArgTarget(req.Intent.Target) {
			return engine.ArgAction(ctx, req.SessionID, webengine.ArgActionRequest{
				Node:   firstStringArg(req.Intent.Target, "obj", "nodeId", "node"),
				Action: commandID,
				Args:   args,
			})
		}
		return engine.Action(ctx, req.SessionID, webengine.ActionRequest{
			Action: commandID,
			Args:   args,
		})
	}
}

func getSnapshot(ctx context.Context, req wasmRequest) (map[string]any, error) {
	out := make(map[string]any)
	snapshot := req.Snapshot
	includeDefault := len(snapshot) == 0
	if includeDefault || hasKey(snapshot, "arg") {
		arg, err := engine.ARG(ctx, req.SessionID)
		if err != nil {
			return nil, err
		}
		out["arg"] = arg
	}
	if hasKey(snapshot, "ctiArg") {
		ctiArg, err := engine.CTIARG(ctx, req.SessionID)
		if err != nil {
			return nil, err
		}
		out["ctiArg"] = ctiArg
	}
	if includeDefault || hasKey(snapshot, "concept") {
		conceptReq := webengine.ConceptRequest{}
		if raw, ok := snapshot["concept"].(map[string]any); ok {
			conceptReq.SheetID = stringArg(raw["sheetId"])
			conceptReq.NodeID = firstStringArg(raw, "nodeId", "stateId")
		}
		concept, err := engine.Concept(ctx, req.SessionID, conceptReq)
		if err != nil {
			return nil, err
		}
		out["concept"] = concept
	}
	if hasKey(snapshot, "menus") {
		menus, err := engine.Menus(ctx)
		if err != nil {
			return nil, err
		}
		out["menus"] = menus
	}
	if hasKey(snapshot, "toggles") {
		toggles, err := engine.Toggles(ctx, req.SessionID)
		if err != nil {
			return nil, err
		}
		out["toggles"] = toggles
	}
	if hasKey(snapshot, "proof") {
		proof, err := engine.Proof(ctx, req.SessionID)
		if err != nil {
			return nil, err
		}
		out["proof"] = proof
	}
	return out, nil
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
	case "check.abstract":
		return "abstract"
	default:
		return "pdr"
	}
}

func hasKey(values map[string]any, key string) bool {
	_, ok := values[key]
	return ok
}

func isArgTarget(target map[string]any) bool {
	if stringArg(target["kind"]) == "arg" {
		return true
	}
	return firstStringArg(target, "obj", "nodeId", "node") != ""
}

func webuiConceptMaterializeRequest(args map[string]any, defaultType string) webui.ConceptMaterializeRequest {
	req := webui.ConceptMaterializeRequest{
		Concept:  stringArg(args["concept"]),
		Type:     firstStringArg(args, "type", "_type", "kind", "materializeType", "materialize_type"),
		Relation: stringArg(args["relation"]),
		Source:   stringArg(args["source"]),
		Target:   stringArg(args["target"]),
		Positive: boolArg(args["positive"]),
	}
	if req.Type == "" {
		req.Type = defaultType
	}
	return req
}

func firstPresent(values map[string]any, keys ...string) any {
	for _, key := range keys {
		if value, ok := values[key]; ok {
			return value
		}
	}
	return nil
}

func firstStringArg(values map[string]any, keys ...string) string {
	for _, key := range keys {
		if value, ok := values[key]; ok {
			if text := stringArg(value); text != "" {
				return text
			}
		}
	}
	return ""
}

func stringArg(value any) string {
	switch typed := value.(type) {
	case string:
		return typed
	case nil:
		return ""
	default:
		return ""
	}
}

func boolArg(value any) bool {
	switch typed := value.(type) {
	case bool:
		return typed
	case float64:
		return typed != 0
	case int:
		return typed != 0
	default:
		return false
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
