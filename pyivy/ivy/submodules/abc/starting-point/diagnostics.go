package startingpoint

import (
	"errors"
	"fmt"
	"strings"
)

type syntaxError struct {
	Message string
	Offset  int
}

func (e *syntaxError) Error() string { return fmt.Sprintf("%s at byte offset %d", e.Message, e.Offset) }

// SourceLocation uses original input bytes: offsets are zero-based; lines and
// byte columns are one-based. Tabs count as one byte, not expanded display cells.
type SourceLocation struct{ Line, Column, ByteOffset int }
type SourceError struct {
	Message  string
	Location SourceLocation
	cause    error
}

func (e *SourceError) Error() string {
	return fmt.Sprintf("%s at source line %d, column %d (byte offset %d)", e.Message, e.Location.Line, e.Location.Column, e.Location.ByteOffset)
}
func (e *SourceError) Unwrap() error { return e.cause }

func sourceError(source string, err error) error {
	var syntax *syntaxError
	if !errors.As(err, &syntax) {
		return err
	}
	offset := syntax.Offset
	if offset < 0 || offset > len(source) {
		return err
	}
	prefix := source[:offset]
	return &SourceError{Message: syntax.Message, Location: SourceLocation{Line: strings.Count(prefix, "\n") + 1, Column: offset - strings.LastIndex(prefix, "\n"), ByteOffset: offset}, cause: err}
}

// Transformed locations must not masquerade as coordinates in the C input.
func transformedDiagnostic(path, phase, draft string, err error) error {
	var syntax *syntaxError
	if errors.As(err, &syntax) && syntax.Offset >= 0 && syntax.Offset <= len(draft) {
		prefix := draft[:syntax.Offset]
		return fmt.Errorf("%s in transformed draft for %s (not original-source coordinates): line %d, column %d: %w", phase, path, strings.Count(prefix, "\n")+1, syntax.Offset-strings.LastIndex(prefix, "\n"), err)
	}
	return fmt.Errorf("%s in transformed draft for %s: %w", phase, path, err)
}
