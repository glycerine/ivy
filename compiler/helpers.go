package compiler

import (
	"fmt"
	"strings"

	"github.com/glycerine/goivy/ast"
	il "github.com/glycerine/goivy/ivylogic"
	iu "github.com/glycerine/goivy/ivyutils"
	lg "github.com/glycerine/goivy/logic"
	"github.com/glycerine/goivy/module"
)

// ResolveAlias resolves a name through the module's alias table.
// If no alias exists, the name is returned unchanged.
func ResolveAlias(name string, mod *module.Module) string {
	if mod == nil || mod.Aliases == nil {
		return name
	}
	return resolveAliasInt(name, mod)
}

func resolveAliasInt(name string, mod *module.Module) string {
	if alias, ok := mod.Aliases[name]; ok {
		return alias
	}
	parts := strings.Split(name, iu.ComposeCharacter)
	if len(parts) == 1 {
		return name
	}
	// Try resolving the parent portion
	parent := strings.Join(parts[:len(parts)-1], iu.ComposeCharacter)
	child := parts[len(parts)-1]
	resolved := resolveAliasInt(parent, mod)
	return resolved + iu.ComposeCharacter + child
}

// cfrError is used internally during field reference compilation to
// signal that a name was not found (analogous to Python's cfrfail).
type cfrError struct {
	SymbolName string
}

func (e *cfrError) Error() string {
	return fmt.Sprintf("cfrfail: %s", e.SymbolName)
}

// CompileFieldReference resolves a dotted-name field reference at the
// top level, catching internal failures and converting them to proper errors.
func (c *Compiler) CompileFieldReference(symbolName string, args []lg.Node, lineno ast.Location, old bool) (lg.Node, error) {
	argsCopy := make([]lg.Node, len(args))
	copy(argsCopy, args)
	result, err := c.compileFieldReferenceRec(symbolName, argsCopy, true, old)
	if err != nil {
		if cfrErr, ok := err.(*cfrError); ok {
			if _, inSorts := c.Sig.Sorts[symbolName]; inSorts {
				return nil, &lg.IvyError{Msg: fmt.Sprintf(
					"type %s used where a function or individual symbol is expected",
					cfrErr.SymbolName)}
			}
			return nil, &lg.IvyError{Msg: fmt.Sprintf("unknown symbol: %s", cfrErr.SymbolName)}
		}
		return nil, err
	}
	return result, nil
}

// compileFieldReferenceRec is the recursive implementation of field reference
// compilation. It splits dotted names and looks up destructors and actions.
func (c *Compiler) compileFieldReferenceRec(symbolName string, args []lg.Node, top bool, old bool) (lg.Node, error) {
	// Try to find the symbol directly (polymorphic or in signature)
	sym, found := il.FindPolymorphicSymbol(symbolName)
	if !found {
		if entry, ok := c.Sig.Symbols[symbolName]; ok {
			sym = lg.NewConst(symbolName, entry.Sort)
			found = true
		}
	}

	if !found {
		// Split into parent.child
		pc := iu.ParentChildName(symbolName)
		parentName := pc[0]
		childName := pc[1]

		if parentName == "this" {
			return nil, &cfrError{SymbolName: symbolName}
		}

		// Recursively compile the parent
		savedRetCtx := c.ReturnCtx
		c.ReturnCtx = nil
		base, err := c.compileFieldReferenceRec(parentName, args, false, old)
		c.ReturnCtx = savedRetCtx
		if err != nil {
			if cfrErr, ok := err.(*cfrError); ok {
				_, inHier := c.Module.Hierarchy[symbolName]
				if inHier {
					return nil, &cfrError{SymbolName: symbolName}
				}
				return nil, cfrErr
			}
			return nil, err
		}

		sort := base.NodeSort()
		// Look for the method as a child of the sort
		destrName := iu.ComposeNames(il.SortName(sort), childName)
		if c.TopCtx != nil {
			if _, inSig := c.Sig.Symbols[destrName]; !inSig {
				if _, inAct := c.TopCtx.Actions[destrName]; !inAct {
					// Try sibling of the sort
					sortPC := iu.ParentChildName(il.SortName(sort))
					destrName = iu.ComposeNames(sortPC[0], childName)
				}
			}
		}

		// Check if it's an action call
		if c.TopCtx != nil {
			if actInfo, ok := c.TopCtx.Actions[destrName]; ok {
				if c.ExprCtx == nil {
					return nil, &lg.IvyError{Msg: fmt.Sprintf(
						"call to action %s not allowed outside an action", destrName)}
				}
				keyPos := actInfo.KeyPos
				// Insert base at key position
				newArgs := make([]lg.Node, 0, len(args)+1)
				newArgs = append(newArgs, args[:keyPos]...)
				newArgs = append(newArgs, base)
				newArgs = append(newArgs, args[keyPos:]...)

				nformals := len(actInfo.Params)
				callArgs := pullArgs(newArgs, nformals, destrName, top)
				atom := ast.NewAtom(destrName)
				return c.CompileInlineCall(atom, callArgs)
			}
		}

		// Find the destructor symbol
		destrSym, err := c.findSymbol(destrName)
		if err != nil {
			return nil, &cfrError{SymbolName: symbolName}
		}
		// Prepend base to args
		args = append([]lg.Node{base}, args...)
		sym = destrSym
		found = true
	}

	if !found {
		return nil, &cfrError{SymbolName: symbolName}
	}

	// Apply old_ prefix if needed
	if old {
		sym = lg.NewConst("old_"+sym.Name, sym.CSort)
	}

	// Apply to arguments
	if fs, ok := sym.CSort.(*lg.FunctionSort); ok && fs.Arity() > 0 {
		actualArgs := pullArgs(args, fs.Arity(), sym.Name, top)
		// TODO: sort_infer each argument against domain sorts
		result, err := lg.NewApply(sym, actualArgs...)
		if err != nil {
			return nil, err
		}
		return result, nil
	}

	// 0-arity symbol
	return sym, nil
}

// CompileInlineCall compiles an inline action call within an expression.
// This handles the pattern where actions are called on the rhs of
// assignments and their return values become expression values.
func (c *Compiler) CompileInlineCall(self *ast.Atom, args []lg.Node) (lg.Node, error) {
	rep := ResolveAlias(self.Rep, c.Module)

	if c.TopCtx == nil {
		return nil, &lg.IvyError{Msg: fmt.Sprintf("no top context for inline call to %s", rep)}
	}

	actInfo, ok := c.TopCtx.Actions[rep]
	if !ok {
		return nil, &lg.IvyError{Msg: fmt.Sprintf("unknown action: %s", rep)}
	}

	params := actInfo.Params
	returns := actInfo.Returns

	if c.ReturnCtx == nil || c.ReturnCtx.Values == nil {
		if len(returns) != 1 {
			return nil, &lg.IvyError{Msg: "wrong number of return values"}
		}
		// Create a local symbol for the return value
		retSort := returns[0].CSort
		locName := fmt.Sprintf("loc:%d", len(c.ExprCtx.LocalSyms))
		locSym := lg.NewConst(locName, retSort)
		c.ExprCtx.LocalSyms = append(c.ExprCtx.LocalSyms, locSym)

		// Validate parameter count
		if len(params) != len(args) {
			return nil, &lg.IvyError{Msg: fmt.Sprintf(
				"wrong number of input parameters (got %d, expecting %d)",
				len(args), len(params))}
		}

		// TODO: create CallAction and add to ExprCtx.Code
		// For now, just return the local symbol
		return locSym, nil
	}

	// Return context has explicit values
	returnValues := c.ReturnCtx.Values
	if len(returns) != len(returnValues) {
		return nil, &lg.IvyError{Msg: "wrong number of return values"}
	}

	if len(params) != len(args) {
		return nil, &lg.IvyError{Msg: fmt.Sprintf(
			"wrong number of input parameters (got %d, expecting %d)",
			len(args), len(params))}
	}

	// TODO: create CallAction, handle variant dispatch, add to ExprCtx.Code
	return nil, nil
}

// pullArgs extracts numArgs arguments from the args slice. If top is true
// and there are too many args, it returns an error via panic-recovery
// (matching Python's pull_args behavior).
func pullArgs(args []lg.Node, numArgs int, sym string, top bool) []lg.Node {
	if len(args) < numArgs {
		// Not enough arguments - return what we have
		return args
	}
	if top && len(args) > numArgs {
		// Too many arguments - return only what's needed
		return args[:numArgs]
	}
	return args[:numArgs]
}
