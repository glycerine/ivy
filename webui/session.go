package webui

import (
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
			if len(decl.Args()) > 0 {
				if sym, ok := decl.Args()[0].(*ast.Symbol); ok {
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
	s.ConceptSess = NewConceptSession()
	// Add sort concepts (nodes)
	for _, sortName := range sorts {
		s.ConceptSess.Domain.Concepts[sortName] = &Concept{
			Name:      sortName,
			Variables: []string{"X"},
			Formula:   fmt.Sprintf("X:%s", sortName),
			Sorts:     []string{sortName},
			Arity:     1,
		}
	}
	// Add relation concepts (edges)
	for _, rel := range relations {
		var vars []string
		var sortList []string
		var paramStrs []string
		for _, p := range rel.Params {
			vars = append(vars, p.Name)
			sortList = append(sortList, p.Sort)
			paramStrs = append(paramStrs, p.Name+":"+p.Sort)
		}
		name := rel.Name
		if len(paramStrs) > 0 {
			name = rel.Name + "(" + strings.Join(paramStrs, ",") + ")"
		}
		s.ConceptSess.Domain.Concepts[rel.Name] = &Concept{
			Name:      rel.Name,
			Variables: vars,
			Formula:   name,
			Sorts:     sortList,
			Arity:     len(vars),
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

// emit sends an event on the SSE channel (non-blocking drop if full).
func (s *Session) emit(e Event) {
	select {
	case s.Events <- e:
	default:
		// drop if buffer full
	}
}
