package goivy

import "errors"

// astContextError is Python's ivy_ast.ASTContext.__exit__, applied to the
// error returned while compiling node:
//
//	if isinstance(exc_val,ivy_logic.Error):            # logic-level error
//	    raise IvyError(self.ast,str(exc_val))
//	if isinstance(exc_val,IvyError):
//	    needs_lineno = hasattr(exc_val,'lineno') and not exc_val.lineno.line
//	    if needs_lineno and hasattr(self.ast,'lineno'):
//	        exc_val.lineno = self.ast.lineno
//
// goivy's SortError plays ivy_logic.Error, and IvyError plays
// ivy_utils.IvyError.
func astContextError(node Node, err error) error {
	if err == nil {
		return nil
	}
	var se *SortError
	if errors.As(err, &se) {
		return NewIvyError(node, se.Msg)
	}
	// Python's proof errors are IvyError subclasses: same lineno rule.
	if pe := proofErrorNodeField(err); pe != nil {
		if _, located := proofErrorLoc(*pe); !located {
			if b, ok := node.(interface{ HasLocSet() bool }); ok && b.HasLocSet() {
				*pe = node
			}
		}
		return err
	}
	var ie *IvyError
	if errors.As(err, &ie) && !(ie.HasLoc && ie.Loc.Line != 0) {
		if b, ok := node.(interface{ HasLocSet() bool }); ok && b.HasLocSet() {
			ie.Loc = node.GetLineno()
			ie.HasLoc = true
		}
	}
	return err
}

// pyErrorString is Python's str(e) for an IvyError: str(lineno) + 'error: '
// + msg, so an error without a location still reads "error: msg".
func pyErrorString(err error) string {
	var ie *IvyError
	if errors.As(err, &ie) && ie.HasLoc {
		return ie.Error()
	}
	return "error: " + err.Error()
}

// proofErrorNodeField returns a pointer to the Node field of a proof error
// (Redefinition, Circular, NoMatch, ProofError, CaptureError) in err's
// chain, or nil.
func proofErrorNodeField(err error) *interface{} {
	var r *Redefinition
	var c *Circular
	var n *NoMatch
	var p *ProofError
	var ce *CaptureError
	switch {
	case errors.As(err, &r):
		return &r.Node
	case errors.As(err, &c):
		return &c.Node
	case errors.As(err, &n):
		return &n.Node
	case errors.As(err, &p):
		return &p.Node
	case errors.As(err, &ce):
		return &ce.Node
	}
	return nil
}
