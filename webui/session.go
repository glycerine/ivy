package webui

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/glycerine/goivy/ast"
	"github.com/glycerine/goivy/lexer"
	"github.com/glycerine/goivy/parser"
)

// Event is a server-sent event delivered to the browser over SSE.
type Event struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

// Session holds the state for one interactive verification session.
type Session struct {
	ID          string
	Graph       *AnalysisGraphState // ARG state
	ConceptSess *ConceptSession     // concept graph state
	Events      chan Event          // buffered SSE channel
	mu          sync.Mutex
	FilePath    string // last loaded file path
	FileContent string // file content (when uploaded via browser)
	toggles     *Toggles
	ProofStack  *ProofStack
}

// NewSession creates a new verification session with the given id.
func NewSession(id string) *Session {
	return &Session{
		ID:     id,
		Events: make(chan Event, 64),
		Graph:  NewAnalysisGraphState(),
		ConceptSess: NewConceptSession(),
	}
}

// LoadFile loads an Ivy source file by path into this session.
func (s *Session) LoadFile(path string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if path == "" {
		return fmt.Errorf("empty file path")
	}
	s.FilePath = path
	s.emit(Event{Type: "file_loaded", Data: map[string]string{"path": path}})
	return nil
}

// LoadFileContent loads an Ivy source file from in-memory content
// (used when the browser uploads a file via multipart form).
// It parses the file to extract types, relations, and actions
// to populate the concept domain and ARG.
func (s *Session) LoadFileContent(filename string, content []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if filename == "" {
		return fmt.Errorf("empty filename")
	}
	s.FilePath = filename
	s.FileContent = string(content)

	// Parse the Ivy file to extract declarations.
	version := lexer.Version{1, 7}
	// Check for #lang directive
	src := string(content)
	if strings.HasPrefix(src, "#lang ivy") {
		// skip the #lang line for the parser
		if idx := strings.Index(src, "\n"); idx >= 0 {
			src = src[idx+1:]
		}
	}

	p := parser.New(src, version)
	decls, _ := p.Parse() // best-effort: ignore parse errors

	// Extract types, relations, and actions from declarations.
	var sorts []string
	var relations []RelationInfo
	var actions []string

	for _, d := range decls {
		switch decl := d.(type) {
		case *ast.TypeDecl:
			// type client, type server
			// The first arg is a TypeDef whose Name is a Symbol
			if len(decl.Args()) > 0 {
				if td, ok := decl.Args()[0].(*ast.TypeDef); ok {
					if sym, ok := td.Name.(*ast.Symbol); ok {
						sorts = append(sorts, sym.Rep)
					} else if atom, ok := td.Name.(*ast.Atom); ok {
						sorts = append(sorts, atom.Rep)
					}
				} else if sym, ok := decl.Args()[0].(*ast.Symbol); ok {
					sorts = append(sorts, sym.Rep)
				} else if atom, ok := decl.Args()[0].(*ast.Atom); ok {
					sorts = append(sorts, atom.Rep)
				}
			}
		case *ast.RelationDecl:
			// relation link(X:client, Y:server)
			if len(decl.Args()) > 0 {
				if atom, ok := decl.Args()[0].(*ast.Atom); ok {
					ri := RelationInfo{Name: atom.Rep}
					for _, t := range atom.Terms {
						if v, ok := t.(*ast.Variable); ok {
							sortName := ""
							if v.VSort != nil {
								sortName = v.VSort.String()
							}
							ri.Params = append(ri.Params, ParamInfo{
								Name: v.Rep,
								Sort: sortName,
							})
						}
					}
					relations = append(relations, ri)
				}
			}
		case *ast.ActionDecl:
			if len(decl.Args()) > 0 {
				if ad, ok := decl.Args()[0].(*ast.ActionDef); ok {
					if len(ad.Args()) > 0 {
						if atom, ok := ad.Args()[0].(*ast.Atom); ok {
							actions = append(actions, atom.Rep)
						}
					}
				}
			}
		}
	}

	// Build the concept domain from extracted declarations.
	// Mirrors Python's get_initial_concept_domain(sig) in concept.py.
	s.ConceptSess = NewConceptSession()

	// Add sort concepts as nodes (one per sort, matching Python).
	for _, sortName := range sorts {
		s.ConceptSess.Domain.Concepts[sortName] = &Concept{
			Name:      sortName,
			Variables: []string{"X"},
			Formula:   fmt.Sprintf("X = X"), // Eq(X,X) - always true for sort membership
			Sorts:     []string{sortName},
			Arity:     1,
		}
		s.ConceptSess.Domain.Nodes = append(s.ConceptSess.Domain.Nodes, sortName)
	}

	// Add relation concepts, categorized as Python does:
	//   arity 1 → node_labels (e.g., "semaphore")
	//   arity 2 → edges (e.g., "link")
	for _, rel := range relations {
		var vars []string
		var sortList []string
		for _, p := range rel.Params {
			vars = append(vars, p.Name)
			sortList = append(sortList, p.Sort)
		}
		s.ConceptSess.Domain.Concepts[rel.Name] = &Concept{
			Name:      rel.Name,
			Variables: vars,
			Formula:   rel.Name + "(" + strings.Join(vars, ", ") + ")",
			Sorts:     sortList,
			Arity:     len(vars),
		}
		// Categorize exactly as Python does:
		switch len(vars) {
		case 1:
			s.ConceptSess.Domain.NodeLabels = append(s.ConceptSess.Domain.NodeLabels, rel.Name)
		case 2:
			s.ConceptSess.Domain.Edges = append(s.ConceptSess.Domain.Edges, rel.Name)
		}
	}

	// Build initial ARG with state 0
	s.Graph = NewAnalysisGraphState()
	s.Graph.States = append(s.Graph.States, ARGNode{
		ID:    0,
		Label: "0",
	})

	s.emit(Event{Type: "file_loaded", Data: map[string]interface{}{
		"filename":  filename,
		"size":      len(content),
		"sorts":     sorts,
		"relations": relationNames(relations),
		"actions":   actions,
	}})
	return nil
}

// RelationInfo describes a relation declaration.
type RelationInfo struct {
	Name   string
	Params []ParamInfo
}

// ParamInfo describes a parameter of a relation.
type ParamInfo struct {
	Name string
	Sort string
}

func relationNames(rels []RelationInfo) []string {
	names := make([]string, len(rels))
	for i, r := range rels {
		names[i] = r.Name
	}
	return names
}

// ExecuteAction runs a named verification action with the given arguments.
// This is a stub: real logic will dispatch to the checker / art packages.
func (s *Session) ExecuteAction(actionName string, args map[string]interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if actionName == "" {
		return fmt.Errorf("empty action name")
	}
	s.emit(Event{
		Type: "action_started",
		Data: map[string]interface{}{
			"action": actionName,
			"args":   args,
		},
	})
	// TODO: dispatch real verification actions
	s.emit(Event{
		Type: "action_completed",
		Data: map[string]interface{}{
			"action": actionName,
			"status": "ok",
		},
	})
	return nil
}

// ProofStack holds proof goal state for rendering.
// Stub: will be wired to the proof/ package.
func (s *Session) ProofStackData() *ProofStack {
	return s.ProofStack
}

// AddProjection adds a projection concept to the domain.
func (s *Session) AddProjection(name, concept string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	// Stub: real implementation creates a projected binary concept
	s.emit(Event{Type: "concept_updated", Data: nil})
	return nil
}

// ArgNodeAction executes an action on an ARG node (view state, execute, mark, cover, etc.)
func (s *Session) ArgNodeAction(nodeID, action string, args map[string]interface{}) (map[string]interface{}, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	// Stub: will dispatch to art/check packages
	result := map[string]interface{}{
		"status": "ok",
		"node":   nodeID,
		"action": action,
	}
	s.emit(Event{Type: "action_completed", Data: map[string]string{"action": action}})
	return result, nil
}

// ProofGoalAction executes an action on a proof goal node.
func (s *Session) ProofGoalAction(goalID, action string) (map[string]interface{}, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	// Stub: will dispatch to proof/ package
	result := map[string]interface{}{
		"status": "ok",
		"goal":   goalID,
		"action": action,
	}
	return result, nil
}

// SaveState serializes the current session state as JSON.
func (s *Session) SaveState() []byte {
	s.mu.Lock()
	defer s.mu.Unlock()
	data, _ := json.Marshal(map[string]interface{}{
		"session_id":   s.ID,
		"file_path":    s.FilePath,
		"file_content": s.FileContent,
		"toggles":      s.toggles,
	})
	return data
}

// Toggles stores edge/label visibility checkbox state.
type Toggles struct {
	Edges map[string]map[string]bool `json:"edges"` // edge_name → {display_class → checked}
}

// GetToggles returns the current toggle state.
func (s *Session) GetToggles() *Toggles {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.toggles == nil {
		s.toggles = &Toggles{Edges: make(map[string]map[string]bool)}
	}
	return s.toggles
}

// SetToggle updates a single toggle value.
func (s *Session) SetToggle(edge, displayClass string, value bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.toggles == nil {
		s.toggles = &Toggles{Edges: make(map[string]map[string]bool)}
	}
	if s.toggles.Edges[edge] == nil {
		s.toggles.Edges[edge] = make(map[string]bool)
	}
	s.toggles.Edges[edge][displayClass] = value
}

// emit sends an event on the SSE channel (non-blocking drop if full).
func (s *Session) emit(e Event) {
	select {
	case s.Events <- e:
	default:
		// drop if buffer full
	}
}
