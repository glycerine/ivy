// Package ivylsp provides a minimal Language Server Protocol stub for Ivy.
// This is a port of Python's ivy_lsp.py (19 lines).
//
// The Python version uses pygls to implement a minimal LSP server with
// text document completion. This Go version provides the same stub structure
// using Go LSP patterns.
package ivylsp

import (
	"fmt"
	"io"
	"os"
)

// Server represents the Ivy LSP server.
type Server struct {
	Name    string
	Version string
	In      io.Reader
	Out     io.Writer
}

// NewServer creates a new Ivy LSP server.
func NewServer() *Server {
	return &Server{
		Name:    "ivy-lsp-server",
		Version: "v0.1",
		In:      os.Stdin,
		Out:     os.Stdout,
	}
}

// CompletionItem represents a single completion suggestion.
type CompletionItem struct {
	Label  string
	Kind   int
	Detail string
}

// Complete returns completion items for the given document URI and position.
// This is a stub that returns no completions.
func (s *Server) Complete(uri string, line, col int) []CompletionItem {
	// Stub: no completions provided yet.
	// Full implementation would parse the document, determine context,
	// and suggest Ivy keywords, symbols, actions, etc.
	return nil
}

// StartIO starts the LSP server reading from stdin and writing to stdout.
// This is a stub that prints a message and exits.
func (s *Server) StartIO() error {
	fmt.Fprintf(os.Stderr, "%s %s: LSP server not yet implemented\n", s.Name, s.Version)
	return nil
}

// Main is the entry point for the LSP server.
func Main() {
	s := NewServer()
	s.StartIO()
}
