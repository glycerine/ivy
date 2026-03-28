package compiler

import (
	"fmt"
	"strings"

	"github.com/glycerine/goivy/actions"
	"github.com/glycerine/goivy/ast"
	il "github.com/glycerine/goivy/ivylogic"
	lg "github.com/glycerine/goivy/logic"
	"github.com/glycerine/goivy/module"
	"github.com/glycerine/goivy/xtracer"
)

// astTypeName returns the bare AST type name for a node, matching Python's
// type(x).__name__. E.g., *ast.Atom → "Atom", *ast.Range → "Range".
func astTypeName(n ast.Node) string {
	if n == nil {
		return "nil"
	}
	t := fmt.Sprintf("%T", n)
	if i := strings.LastIndex(t, "."); i >= 0 {
		t = t[i+1:]
	}
	return t
}

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
	cc := mod.Cfg.IuCfg.ComposeCharacter
	parts := strings.Split(name, cc)
	if len(parts) == 1 {
		return name
	}
	// Try resolving the parent portion
	parent := strings.Join(parts[:len(parts)-1], cc)
	child := parts[len(parts)-1]
	resolved := resolveAliasInt(parent, mod)
	return resolved + cc + child
}

// compileNativeType applies alias resolution to a NativeType's child reps.
// Corresponds to Python's compile_native_type (ivy_compiler.py:793-794):
//
//	self.clone([self.args[0]] + [x.rename(resolve_alias(x.rep)) for x in self.args[1:]])
func compileNativeType(nt *ast.NativeType, mod *module.Module) *ast.NativeType {
	if len(nt.Elems) <= 1 {
		return nt
	}
	newElems := make([]ast.Node, len(nt.Elems))
	newElems[0] = nt.Elems[0] // keep args[0] unchanged
	for i := 1; i < len(nt.Elems); i++ {
		elem := nt.Elems[i]
		name := extractSortRep(elem)
		if name != "" {
			resolved := ResolveAlias(name, mod)
			if resolved != name {
				if atom, ok := elem.(*ast.Atom); ok {
					newAtom := &ast.Atom{Rep: resolved}
					newAtom.Cfg = atom.Cfg
					newAtom.SetLineno(atom.GetLineno())
					newElems[i] = newAtom
					continue
				}
			}
		}
		newElems[i] = elem
	}
	return &ast.NativeType{Base: nt.Base, Elems: newElems}
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
func (c *Compiler) CompileFieldReference(symbolName string, args []lg.Expr, lineno ast.Location, old bool) (lg.Expr, error) {
	xtracer.Trace("compiler.compile_field_reference ENTER name=%s", symbolName)
	argsCopy := make([]lg.Expr, len(args))
	copy(argsCopy, args)
	result, argsCopy, err := c.compileFieldReferenceRec(symbolName, argsCopy, true, old)
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
func (c *Compiler) compileFieldReferenceRec(symbolName string, args []lg.Expr, top bool, old bool) (lg.Expr, []lg.Expr, error) {
	xtracer.Trace("compiler.compile_field_reference_rec ENTER name=%s", symbolName)
	// Try to find the symbol directly (polymorphic or in signature)
	sym, found := il.FindPolymorphicSymbol(symbolName, c.Module.Cfg.IuCfg)
	if !found {
		if entry, ok := c.Sig.Symbols[symbolName]; ok {
			sym = lg.NewSymbol(symbolName, entry.Sort)
			found = true
		}
	}
	if found {
		xtracer.Trace("compiler.compile_field_reference_rec found name=%s sort=%v", symbolName, sym.CSort)
	} else {
		xtracer.Trace("compiler.compile_field_reference_rec not_found name=%s", symbolName)
	}

	if !found {
		// Split into parent.child
		pc := c.Module.Cfg.IuCfg.ParentChildName(symbolName)
		parentName := pc[0]
		childName := pc[1]

		if parentName == "this" {
			return nil, args, &cfrError{SymbolName: symbolName}
		}

		// Recursively compile the parent
		savedRetCtx := c.ReturnCtx
		c.ReturnCtx = nil
		base, updatedArgs, err := c.compileFieldReferenceRec(parentName, args, false, old)
		args = updatedArgs // Python: args is a shared mutable list; del args[:n] in pull_args is visible here
		c.ReturnCtx = savedRetCtx
		if err != nil {
			if cfrErr, ok := err.(*cfrError); ok {
				// B4-R2: Python checks the caught error's symbol, not the current symbolName
			_, inHier := c.Module.Hierarchy[cfrErr.SymbolName]
				if inHier {
					return nil, args, &cfrError{SymbolName: symbolName}
				}
				return nil, args, cfrErr
			}
			return nil, args, err
		}

		sort := base.NodeSort()
		// Look for the method as a child of the sort
		iuCfg := c.Module.Cfg.IuCfg
		destrName := iuCfg.ComposeNames(il.SortName(sort), childName)
		xtracer.Trace("compiler.compile_field_reference_rec destrName=%s baseSort=%s", destrName, il.SortName(sort))
		if c.TopCtx != nil {
			_, inSig := c.Sig.Symbols[destrName]
			_, inAct := c.TopCtx.Actions[destrName]
			xtracer.Trace("compiler.compile_field_reference_rec destr_check name=%s inSig=%v inAct=%v", destrName, inSig, inAct)
			if !inSig {
				if !inAct {
					// Try sibling of the sort
					sortPC := iuCfg.ParentChildName(il.SortName(sort))
					destrName = iuCfg.ComposeNames(sortPC[0], childName)
					xtracer.Trace("compiler.compile_field_reference_rec sibling_fallback destrName=%s", destrName)
				}
			}
		}

		// Check if it's an action call
		if c.TopCtx != nil {
			if actInfo, ok := c.TopCtx.Actions[destrName]; ok {
				// Python: nformals = len(top_context.actions[actname][0])
				// which is the AST-level formals, not compiled Params.
				nformals := len(actInfo.FormalAST)
				xtracer.Trace("compiler.compile_field_reference_rec action_found name=%s keyPos=%d nParams=%d nArgs=%d", destrName, actInfo.KeyPos, nformals, len(args))
				if c.ExprCtx == nil {
					return nil, args, &lg.IvyError{Msg: fmt.Sprintf(
						"call to action %s not allowed outside an action", destrName)}
				}
				keyPos := actInfo.KeyPos
				// Insert base at key position
				newArgs := make([]lg.Expr, 0, len(args)+1)
				newArgs = append(newArgs, args[:keyPos]...)
				newArgs = append(newArgs, base)
				newArgs = append(newArgs, args[keyPos:]...)
				callArgs, remaining, err := pullArgs(newArgs, nformals, destrName, top)
				if err != nil {
					return nil, args, err
				}
				args = remaining
				cfg := c.Module.Cfg.AstCfg
				atom := cfg.NewAtom(destrName)
				result, err := c.CompileInlineCall(atom, callArgs, true)
				return result, args, err
			} else {
				xtracer.Trace("compiler.compile_field_reference_rec action_NOT_found name=%s", destrName)
			}
		}

		// Find the destructor symbol
		destrSym, err := c.findSymbol(destrName)
		if err != nil {
			return nil, args, &cfrError{SymbolName: symbolName}
		}
		// Prepend base to args
		args = append([]lg.Expr{base}, args...)
		sym = destrSym
		found = true
	}

	if !found {
		return nil, args, &cfrError{SymbolName: symbolName}
	}

	// Apply old_ prefix if needed
	if old {
		sym = lg.NewSymbol("old_"+sym.Name, sym.CSort)
	}

	// Apply to arguments
	if fs, ok := sym.CSort.(*lg.FunctionSort); ok && fs.Arity() > 0 {
		actualArgs, remaining, err := pullArgs(args, fs.Arity(), sym.Name, top)
		if err != nil {
			return nil, args, err
		}
		args = remaining
		// Apply sort-guided inference to each argument against the domain sorts.
		// Python: args = [ivy_logic.sort_infer(arg,sort) for arg,sort in zip(args,sym.sort.dom)]
		dom := fs.Domain()
		for i := 0; i < len(actualArgs) && i < len(dom); i++ {
			inferred, err := il.SortInfer(actualArgs[i], dom[i])
			if err == nil {
				actualArgs[i] = inferred
			}
		}
		result, err := lg.NewApply(sym, actualArgs...)
		if err != nil {
			return nil, args, err
		}
		return result, args, nil
	}

	// 0-arity symbol
	return sym, args, nil
}

// CompileInlineCall compiles an inline action call within an expression.
// This handles the pattern where actions are called on the rhs of
// assignments and their return values become expression values.
// The methodcall parameter controls whether variant dispatch is attempted:
// Python: compile_inline_call(self, args, methodcall=False)
func (c *Compiler) CompileInlineCall(self *ast.Atom, args []lg.Expr, methodcall bool) (lg.Expr, error) {
	xtracer.Trace("compiler.compile_inline_call ENTER")
	rep := ResolveAlias(self.Rep, c.Module)

	if c.TopCtx == nil {
		return nil, lg.NewIvyError(self, fmt.Sprintf("no top context for inline call to %s", rep))
	}

	actInfo, ok := c.TopCtx.Actions[rep]
	if !ok {
		return nil, lg.NewIvyError(self, fmt.Sprintf("unknown action: %s", rep))
	}

	// Python: params, returns, keypos = top_context.actions[rep]
	// These are AST-level formals (not compiled Params/Returns which may be nil).
	params := actInfo.FormalAST
	returns := actInfo.FormalRetAST

	if c.ReturnCtx == nil || c.ReturnCtx.Values == nil {
		if len(returns) != 1 {
			return nil, lg.NewIvyError(self, "wrong number of return values")
		}
		// Create a local symbol for the return value
		// Python: sort = cmpl_sort(returns[0].sort)
		retSort, err := c.CmplSort(ast.GetFormalSortAnnotation(returns[0]))
		if err != nil {
			return nil, lg.NewIvyError(self, fmt.Sprintf("cannot resolve return sort: %v", err))
		}
		locName := fmt.Sprintf("loc:%d", len(c.ExprCtx.LocalSyms))
		locSym := lg.NewSymbol(locName, retSort)
		c.ExprCtx.LocalSyms = append(c.ExprCtx.LocalSyms, locSym)

		// Validate parameter count
		if len(params) != len(args) {
			return nil, lg.NewIvyError(self, fmt.Sprintf(
				"wrong number of input parameters (got %d, expecting %d)",
				len(args), len(params)))
		}

		// Create the CallAction: call(atom(rep, args...), returnValue)
		calleeNode := lg.NewSymbol(rep, lg.TopS)
		var callee lg.Expr = calleeNode
		if len(args) > 0 {
			applied, err := lg.NewApply(calleeNode, args...)
			if err == nil {
				callee = applied
			}
		}
		call := actions.NewCallAction(callee, locSym)
		call.SetLineno(self.GetLineno())
		c.ExprCtx.Code = append(c.ExprCtx.Code, call)
		return locSym, nil
	}

	// Return context has explicit values
	returnValues := c.ReturnCtx.Values
	if len(returns) != len(returnValues) {
		return nil, lg.NewIvyError(self, "wrong number of return values")
	}

	// R2: Apply covariant sort inference to return values
	// Python: return_values = [sort_infer_covariant(a,cmpl_sort(p.sort)) for a,p in zip(return_values,returns)]
	for i := 0; i < len(returnValues) && i < len(returns); i++ {
		pSort, err := c.CmplSort(ast.GetFormalSortAnnotation(returns[i]))
		if err == nil {
			inferred, err := c.SortInferCovariant(returnValues[i], pSort)
			if err == nil {
				returnValues[i] = inferred
			}
		}
	}

	if len(params) != len(args) {
		return nil, lg.NewIvyError(self, fmt.Sprintf(
			"wrong number of input parameters (got %d, expecting %d)",
			len(args), len(params)))
	}

	// R2: Apply contravariant sort inference to args
	// Python: args = [sort_infer_contravariant(a,cmpl_sort(p.sort)) for a,p in zip(args,params)]
	for i := 0; i < len(args) && i < len(params); i++ {
		pSort, err := c.CmplSort(ast.GetFormalSortAnnotation(params[i]))
		if err == nil {
			inferred, err := c.SortInferContravariant(args[i], pSort)
			if err == nil {
				args[i] = inferred
			}
		}
	}

	// Create CallAction with the explicit return values
	calleeNode := lg.NewSymbol(rep, lg.TopS)
	var callee lg.Expr = calleeNode
	if len(args) > 0 {
		applied, err := lg.NewApply(calleeNode, args...)
		if err == nil {
			callee = applied
		}
	}
	var call actions.Action = actions.NewCallAction(callee, returnValues...)
	call.SetLineno(self.GetLineno())

	// Handle variant dispatch for method calls
	// R3: Python uses IfAction directly, NOT wrapped in CallAction
	// B5-R1: Python guards variant dispatch with: if methodcall and args[keypos].sort.name in im.module.variants
	if methodcall && actInfo.KeyPos < len(args) {
		keyArg := args[actInfo.KeyPos]
		keySort := keyArg.NodeSort()
		keySortName := il.SortName(keySort)
		if variants, ok := c.Module.Variants[keySortName]; ok {
			iuCfg2 := c.Module.Cfg.IuCfg
			pcRep := iuCfg2.ParentChildName(rep)
			methodName := pcRep[1]
			for _, vsort := range variants {
				vactName := iuCfg2.ComposeNames(il.SortName(vsort), methodName)
				if _, ok := c.TopCtx.Actions[vactName]; !ok {
					pcVsort := iuCfg2.ParentChildName(il.SortName(vsort))
					parent := pcVsort[0]
					vactName = iuCfg2.ComposeNames(parent, methodName)
					if _, ok := c.TopCtx.Actions[vactName]; !ok || vactName == rep {
						continue
					}
				}
				// Create variant dispatch: if Some(tmpsym, isa_test) then call variant else original
				// Python: call = IfAction(ivy_ast.Some(tmpsym, isa_expr), new_call, call)
				tmpSym := lg.NewSymbol("self:"+il.SortName(vsort), vsort)
				tmpArgs := make([]lg.Expr, len(args))
				copy(tmpArgs, args)
				tmpArgs[actInfo.KeyPos] = tmpSym
				var varCallee lg.Expr = lg.NewSymbol(vactName, lg.TopS)
				if len(tmpArgs) > 0 {
					if applied, err := lg.NewApply(lg.NewSymbol(vactName, lg.TopS), tmpArgs...); err == nil {
						varCallee = applied
					}
				}
				newCall := actions.NewCallAction(varCallee, returnValues...)
				// Build the Some condition: Some(tmpsym, *>(keyArg, tmpsym))
				isaSort := il.RelationSort([]lg.Sort{keySort, vsort})
				isaSym := lg.NewSymbol("*>", isaSort)
				isaApp, _ := lg.NewApply(isaSym, keyArg, tmpSym)
				// Python: ivy_ast.Some(tmpsym, isa_expr)
				someCond := il.Exists([]*lg.Variable{
					{Name: tmpSym.Name, VSort: vsort},
				}, isaApp)
				ifAction := actions.NewIfAction(someCond,
					newCall,
					call)
				// R3: assign IfAction directly to call, do NOT wrap in CallAction
				call = ifAction
			}
		}
	}

	c.ExprCtx.Code = append(c.ExprCtx.Code, call)
	return nil, nil
}

// pullArgs extracts numArgs arguments from the args slice, returning the
// consumed args and the remaining args. This matches Python's pull_args
// which mutates the list via `del args[:num]`.
// Python: pull_args (ivy_compiler.py:148-155)
func pullArgs(args []lg.Expr, numArgs int, sym string, top bool) (consumed []lg.Expr, remaining []lg.Expr, err error) {
	if len(args) < numArgs {
		return nil, nil, &lg.IvyError{Msg: fmt.Sprintf("not enough arguments to %s", sym)}
	}
	if top && len(args) > numArgs {
		return nil, nil, &lg.IvyError{Msg: fmt.Sprintf("too many arguments to %s", sym)}
	}
	return args[:numArgs], args[numArgs:], nil
}
