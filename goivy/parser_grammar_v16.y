// parser_grammar_v16.y — goyacc LALR(1) grammar for FULL Ivy file parsing, version <=1.6.
//
// This is intentionally a thin syntax adapter. It mirrors the existing
// v16_logicgrammar_v16.y term/formula split, then hands declarations to the
// shared parser accumulator and builders.

%{
package goivy

import "github.com/glycerine/ivy/goivy/xtracer"

%}

%union {
	node     Node
	nodes    []Node
	str      string
	tok      TokenInfo
	bval     bool
	accum    *ivyAccum
}

%token <tok>  PARSER16_TOK_PRESYMBOL PARSER16_TOK_VARIABLE
%token <tok>  PARSER16_TOK_LPAREN PARSER16_TOK_RPAREN PARSER16_TOK_LB PARSER16_TOK_RB PARSER16_TOK_LCB PARSER16_TOK_RCB
%token <tok>  PARSER16_TOK_COMMA PARSER16_TOK_SEMI PARSER16_TOK_COLON PARSER16_TOK_DOT PARSER16_TOK_DOTS
%token <tok>  PARSER16_TOK_PLUS PARSER16_TOK_MINUS PARSER16_TOK_TIMES PARSER16_TOK_DIV
%token <tok>  PARSER16_TOK_EQ PARSER16_TOK_TILDAEQ PARSER16_TOK_TILDA PARSER16_TOK_LE PARSER16_TOK_LT PARSER16_TOK_GE PARSER16_TOK_GT
%token <tok>  PARSER16_TOK_AND PARSER16_TOK_OR PARSER16_TOK_ARROW PARSER16_TOK_IFF
%token <tok>  PARSER16_TOK_PTO PARSER16_TOK_DOLLAR PARSER16_TOK_ASSIGN
%token <tok>  PARSER16_TOK_FORALL PARSER16_TOK_EXISTS
%token <tok>  PARSER16_TOK_TRUE PARSER16_TOK_FALSE
%token <tok>  PARSER16_TOK_OLD PARSER16_TOK_THIS
%token <tok>  PARSER16_TOK_IF PARSER16_TOK_ELSE PARSER16_TOK_WHILE PARSER16_TOK_INVARIANT
%token <tok>  PARSER16_TOK_GLOBALLY PARSER16_TOK_EVENTUALLY
%token <tok>  PARSER16_TOK_DOTDOTDOT
%token <tok>  PARSER16_TOK_TYPE PARSER16_TOK_INDIV PARSER16_TOK_VAR PARSER16_TOK_FUNCTION PARSER16_TOK_RELATION PARSER16_TOK_DERIVED PARSER16_TOK_STRUCT
%token <tok>  PARSER16_TOK_AXIOM PARSER16_TOK_PROPERTY PARSER16_TOK_CONJECTURE PARSER16_TOK_SCHEMA PARSER16_TOK_ASSERT PARSER16_TOK_DEFINITION PARSER16_TOK_PROOF PARSER16_TOK_WITH PARSER16_TOK_INIT
%token <tok>  PARSER16_TOK_MODULE PARSER16_TOK_CLASS PARSER16_TOK_OBJECT
%token <tok>  PARSER16_TOK_INCLUDE PARSER16_TOK_ACTION PARSER16_TOK_METHOD PARSER16_TOK_CALL PARSER16_TOK_RETURNS
%token <tok>  PARSER16_TOK_IMPORT PARSER16_TOK_EXPORT PARSER16_TOK_PRIVATE PARSER16_TOK_DELEGATE
%token <tok>  PARSER16_TOK_MACRO PARSER16_TOK_ALIAS PARSER16_TOK_PROGRESS PARSER16_TOK_RELY PARSER16_TOK_MIXORD PARSER16_TOK_INTERPRET
%token <tok>  PARSER16_TOK_MIXIN PARSER16_TOK_BEFORE PARSER16_TOK_AFTER PARSER16_TOK_IMPLEMENT
%token <tok>  PARSER16_TOK_TRUSTED PARSER16_TOK_ISOLATE
%token <tok>  PARSER16_TOK_ATTRIBUTE PARSER16_TOK_VARIANT PARSER16_TOK_OF
%token <tok>  PARSER16_TOK_REQUIRES PARSER16_TOK_MODIFIES
%token <tok>  PARSER16_TOK_NAMED PARSER16_TOK_TEMPORAL
%token <tok>  PARSER16_TOK_ASSUME PARSER16_TOK_ENSURES PARSER16_TOK_SET PARSER16_TOK_INSTANTIATE
%token <tok>  PARSER16_TOK_LOCAL PARSER16_TOK_LET PARSER16_TOK_IN
%token <tok>  PARSER16_TOK_SOME PARSER16_TOK_MINIMIZING PARSER16_TOK_MAXIMIZING PARSER16_TOK_DECREASES

%type <accum> top
%type <node>  labeledfmla fmla term aterm var simplevar atype tterm typesymbol symdecl constantdecl rel sort
%type <node>  fun defn defnrhs optproof proofstep match opttemporal optactiondef sequence action simpleact complexact callatom atom optimpex assert_rhs
%type <node>  lparam param optinit termtuple somefmla eqn lit topseq optdelegee oper attributeval requires ensures
%type <node>  optskolem schdefn schdefnrhs schconc objsym objectend
%type <nodes> terms vars simplevars tterms rels funs defns matches optargs optreturns optactualreturns names actseq actseqrev callatoms schdecl schdecls objectargs
%type <nodes> lparams params bounds invariants decreases eqns moresymbols atoms modifies
%type <bval>  optdotdotdot opttrusted
%type <str>   relop SYMBOLx
%type <tok>   labelname

%left         PARSER16_TOK_SEMI
%left         PARSER16_TOK_GLOBALLY PARSER16_TOK_EVENTUALLY
%left         PARSER16_TOK_IF
%left         PARSER16_TOK_ELSE
%left         PARSER16_TOK_OR
%left         PARSER16_TOK_AND
%left         PARSER16_TOK_TILDA
%left         PARSER16_TOK_EQ PARSER16_TOK_LE PARSER16_TOK_LT PARSER16_TOK_GE PARSER16_TOK_GT PARSER16_TOK_PTO
%left         PARSER16_TOK_TILDAEQ
%left         PARSER16_TOK_COLON
%left         PARSER16_TOK_PLUS
%left         PARSER16_TOK_MINUS
%left         PARSER16_TOK_TIMES
%left         PARSER16_TOK_DIV
%left         PARSER16_TOK_DOLLAR

%start        top

%%

top:
    /* empty */
    {
        xtracer.Trace("parser.p_top ENTER (top)")
        lex := parser16lex.(*parser16LexAdapter)
        parent := lex.accum
        $$ = newIvyAccum(parent, lex.parentObjName)
        $$.filename = lex.filename
        lex.parentObjName = ""
        $$.parent = parent
        if $$.astCfg == nil {
            $$.astCfg = lex.astCfg
        }
        lex.accum = $$
    }
    | top PARSER16_TOK_INCLUDE SYMBOLx
    {
        xtracer.Trace("parser.p_top_include_symbol ENTER (top)")
        $$ = $1
        lex := parser16lex.(*parser16LexAdapter)
        parserDeclareInclude(parser16Acfg(parser16lex), $$, lex.accum, lex.importer, $3, tok16Lineno(lex, $2))
    }
    | top PARSER16_TOK_TYPE typesymbol
    {
        xtracer.Trace("parser.p_top_type_symbol ENTER (top)")
        $$ = $1
        parser16DeclareType(parser16Acfg(parser16lex), $$, $3.(*Atom), tok16Lineno(parser16lex.(*parser16LexAdapter), $2))
    }
    | top PARSER16_TOK_TYPE typesymbol PARSER16_TOK_EQ sort
    {
        xtracer.Trace("parser.p_top_type_symbol_eq_sort ENTER (top)")
        $$ = $1
        lex := parser16lex.(*parser16LexAdapter)
        scnst := parser16Acfg(parser16lex).NewAtom($3.(*Atom).Rep)
        scnst.SetLineno(nodeLineno($3))
        sortNode := $5
        _, isRange := sortNode.(*Range)
        if isRange {
            sortNode = parser16Acfg(parser16lex).NewUninterpretedSortAST()
        }
        tdfn := parser16Acfg(parser16lex).NewTypeDef(scnst, sortNode)
        tdfn.SetLineno(tok16Lineno(lex, $4))
        $$.declare(parser16Acfg(parser16lex).NewTypeDecl(tdfn))
        if isRange {
            imp := parser16Acfg(parser16lex).NewImplies(scnst, $5)
            imp.SetLineno(tok16Lineno(lex, $2))
            thing := parser16Acfg(parser16lex).NewInterpretDecl(parser17MkLF(parser16Acfg(parser16lex), imp))
            thing.SetLineno(tok16Lineno(lex, $2))
            $$.declare(thing)
        }
    }
    | top symdecl
    {
        xtracer.Trace("parser.p_top_symdecl ENTER (top)")
        $$ = $1
        $$.declare($2)
    }
    | top PARSER16_TOK_RELATION rels
    {
        xtracer.Trace("parser.p_top_relation_rels ENTER (top)")
        $$ = $1
        for _, d := range $3 {
            $$.declare(d)
        }
    }
    | top PARSER16_TOK_FUNCTION funs
    {
        xtracer.Trace("parser.p_top_function_tapp_colon_atype ENTER (top)")
        $$ = $1
        for _, d := range $3 {
            $$.declare(d)
        }
    }
    | top PARSER16_TOK_DERIVED defns
    {
        xtracer.Trace("parser.p_top_derived_defns ENTER (top)")
        $$ = $1
        args := make([]Node, len($3))
        for i, x := range $3 {
            args[i] = parser17MkLF(parser16Acfg(parser16lex), x)
        }
        dd := parser16Acfg(parser16lex).NewDerivedDecl(args...)
        $$.declare(dd)
    }
    | top PARSER16_TOK_PROGRESS defns
    {
        xtracer.Trace("parser.p_top_progress_defns ENTER (top)")
        $$ = $1
        pd := parser16Acfg(parser16lex).NewProgressDecl($3...)
        $$.declare(pd)
    }
    | top PARSER16_TOK_RELY atom PARSER16_TOK_ARROW atom
    {
        xtracer.Trace("parser.p_top_rely_atom_arrow_atom ENTER (top)")
        $$ = $1
        imp := parser16Acfg(parser16lex).NewImplies($3, $5)
        rd := parser16Acfg(parser16lex).NewRelyDecl(imp)
        $$.declare(rd)
    }
    | top PARSER16_TOK_RELY atom
    {
        xtracer.Trace("parser.p_top_rely_atom ENTER (top)")
        $$ = $1
        rd := parser16Acfg(parser16lex).NewRelyDecl($3)
        $$.declare(rd)
    }
    | top PARSER16_TOK_MIXORD callatom PARSER16_TOK_ARROW callatom
    {
        xtracer.Trace("parser.p_top_mixord_callatom_arrow_callatom ENTER (top)")
        $$ = $1
        imp := parser16Acfg(parser16lex).NewImplies($3, $5)
        md := parser16Acfg(parser16lex).NewMixOrdDecl(imp)
        $$.declare(md)
    }
    | top PARSER16_TOK_MACRO atom PARSER16_TOK_EQ sequence
    {
        xtracer.Trace("parser.p_top_macro_atom_eq_lcb_action_rcb ENTER (top)")
        $$ = $1
        d := parser16Acfg(parser16lex).NewDefinition(AppToAtom($3), $5)
        md := parser16Acfg(parser16lex).NewMacroDecl(d)
        $$.declare(md)
    }
    | top PARSER16_TOK_SCHEMA schdefn
    {
        xtracer.Trace("parser.p_top_schema_defn ENTER (top)")
        $$ = $1
        sch := parser16Acfg(parser16lex).NewSchema($3)
        sd := parser16Acfg(parser16lex).NewSchemaDecl(sch)
        $$.declare(sd)
    }
    | top PARSER16_TOK_MODULE atom PARSER16_TOK_EQ PARSER16_TOK_LCB top PARSER16_TOK_RCB
    {
        xtracer.Trace("parser.p_top_module_atom_eq_lcb_top_rcb ENTER (top)")
        $$ = $1
        lex := parser16lex.(*parser16LexAdapter)
        modAccum := $6
        d := parser16Acfg(parser16lex).NewDefinition(AppToAtom($3), modAccum)
        $$.declare(parser16Acfg(parser16lex).NewModuleDecl(d))
        lex.accum = $$
        $$.isModule = false
    }
    | top PARSER16_TOK_OBJECT objsym objectargs PARSER16_TOK_EQ PARSER16_TOK_LCB optdotdotdot top PARSER16_TOK_RCB objectend
    {
        xtracer.Trace("parser.p_top_object_symbol_eq_lcb_top_rcb ENTER (top)")
        $$ = $1
        objAccum := $8
        pref := $3.(*Atom)
        createObject(parser16Acfg(parser16lex), $$, pref, $4, objAccum, nodeLineno($3), $7)
        parser16lex.(*parser16LexAdapter).accum = $$
    }
    | top PARSER16_TOK_CLASS objsym objectargs PARSER16_TOK_EQ PARSER16_TOK_LCB optdotdotdot top PARSER16_TOK_RCB objectend
    {
        xtracer.Trace("parser.p_top__top_class_symbol_objectargs_eq_lcb_optdo ENTER (top)")
        $$ = $1
        lex := parser16lex.(*parser16LexAdapter)
        objAccum := $8
        scnst := parser16Acfg(parser16lex).NewAtom("this")
        scnst.SetLineno(tok16Lineno(lex, $2))
        tdfn := parser16Acfg(parser16lex).NewTypeDef(scnst, parser16Acfg(parser16lex).NewUninterpretedSortAST())
        tdfn.SetLineno(tok16Lineno(lex, $2))
        objAccum.declare(parser16Acfg(parser16lex).NewTypeDecl(tdfn))
        if n := len(objAccum.decls); n > 1 {
            last := objAccum.decls[n-1]
            copy(objAccum.decls[1:], objAccum.decls[:n-1])
            objAccum.decls[0] = last
        }
        createObject(parser16Acfg(parser16lex), $$, $3.(*Atom), $4, objAccum, nodeLineno($3), $7)
        lex.accum = $$
    }
    | top opttemporal PARSER16_TOK_AXIOM labeledfmla
    {
        xtracer.Trace("parser.p_top_axiom_optlabel_gprop ENTER (top)")
        $$ = $1
        parser16DeclareAxiom(parser16Acfg(parser16lex), $$, $4.(*LabeledFormula), $2 != nil, tok16Lineno(parser16lex.(*parser16LexAdapter), $3))
    }
    | top opttemporal PARSER16_TOK_PROPERTY labeledfmla optskolem optproof
    {
        xtracer.Trace("parser.p_top_property_labeledfmla ENTER (top)")
        $$ = $1
        parser16DeclareProperty(parser16Acfg(parser16lex), $$, $4.(*LabeledFormula), $2 != nil, tok16Lineno(parser16lex.(*parser16LexAdapter), $3))
        if $5 != nil {
            $$.declare(parser16Acfg(parser16lex).NewNamedDecl($5))
        }
        if $6 != nil {
            $$.declare(parser16Acfg(parser16lex).NewProofDecl($6))
        }
    }
    | top PARSER16_TOK_CONJECTURE labeledfmla
    {
        xtracer.Trace("parser.p_top_conjecture_labeledfmla ENTER (top)")
        $$ = $1
        parser16DeclareConjecture(parser16Acfg(parser16lex), $$, $3.(*LabeledFormula), tok16Lineno(parser16lex.(*parser16LexAdapter), $2))
    }
    | top PARSER16_TOK_ASSERT SYMBOLx PARSER16_TOK_ARROW assert_rhs
    {
        xtracer.Trace("parser.p_top_assert_symbol_arrow_assert_rhs ENTER (top)")
        $$ = $1
        thing := parser16Acfg(parser16lex).NewImplies(parser16Acfg(parser16lex).NewAtom($3), $5)
        thing.SetLineno(tok16Lineno(parser16lex.(*parser16LexAdapter), $4))
        d := parser16Acfg(parser16lex).NewAssertDecl(thing)
        d.SetLineno(tok16Lineno(parser16lex.(*parser16LexAdapter), $2))
        $$.declare(d)
    }
    | top PARSER16_TOK_INIT labeledfmla
    {
        xtracer.Trace("parser.p_top_init_labeledfmla ENTER (top)")
        $$ = $1
        parser16DeclareInit(parser16Acfg(parser16lex), $$, $3.(*LabeledFormula), tok16Lineno(parser16lex.(*parser16LexAdapter), $2))
    }
    | top optimpex PARSER16_TOK_ACTION SYMBOLx optargs optreturns optactiondef
    {
        xtracer.Trace("parser.p_top_optimpex_action_symbol_optargs_optreturns_eq_action ENTER (top)")
        $$ = $1
        parser17DeclareAction(parser16Acfg(parser16lex), $$, $2, false, $4, $5, $6, $7, tok16Lineno(parser16lex.(*parser16LexAdapter), $3))
    }
    | top optimpex PARSER16_TOK_METHOD SYMBOLx optargs optreturns optactiondef
    {
        xtracer.Trace("parser.p_top_optimpex_action_symbol_optargs_optreturns_eq_action ENTER (top)")
        $$ = $1
        parser17DeclareAction(parser16Acfg(parser16lex), $$, $2, true, $4, $5, $6, $7, tok16Lineno(parser16lex.(*parser16LexAdapter), $3))
    }
    | top PARSER16_TOK_MIXIN callatom PARSER16_TOK_BEFORE callatom
    {
        xtracer.Trace("parser.p_top_mixin_callatom_before_callatom ENTER (top)")
        $$ = $1
        m := parser16Acfg(parser16lex).NewMixinBeforeDef($3, $5)
        $$.declare(parser16Acfg(parser16lex).NewMixinDecl(m))
    }
    | top PARSER16_TOK_MIXIN callatom PARSER16_TOK_AFTER callatom
    {
        xtracer.Trace("parser.p_top_mixin_callatom_after_callatom ENTER (top)")
        $$ = $1
        m := parser16Acfg(parser16lex).NewMixinAfterDef($3, $5)
        $$.declare(parser16Acfg(parser16lex).NewMixinDecl(m))
    }
    | top PARSER16_TOK_BEFORE atype optargs optreturns topseq
    {
        xtracer.Trace("parser.p_top_before_callatom_lcb_action_rcb ENTER (top)")
        $$ = $1
        atom := atypeToAtom(parser16Acfg(parser16lex), $3)
        atom.SetLineno(tok16Lineno(parser16lex.(*parser16LexAdapter), $2))
        handleBeforeAfterV16(parser16Acfg(parser16lex), "before", atom, $6, $$, $4, $5)
    }
    | top PARSER16_TOK_AFTER atype optargs optreturns topseq
    {
        xtracer.Trace("parser.p_top_after_callatom_lcb_action_rcb ENTER (top)")
        $$ = $1
        atom := atypeToAtom(parser16Acfg(parser16lex), $3)
        atom.SetLineno(tok16Lineno(parser16lex.(*parser16LexAdapter), $2))
        handleBeforeAfterV16(parser16Acfg(parser16lex), "after", atom, $6, $$, $4, $5)
    }
    | top PARSER16_TOK_AFTER PARSER16_TOK_INIT optargs topseq
    {
        xtracer.Trace("parser.p_top_after_init_optargs_lcb_action_rcb ENTER (top)")
        $$ = $1
        atom := parser16Acfg(parser16lex).NewAtom("init")
        atom.SetLineno(tok16Lineno(parser16lex.(*parser16LexAdapter), $2))
        handleBeforeAfterV16(parser16Acfg(parser16lex), "after", atom, $5, $$, $4, nil)
    }
    | top PARSER16_TOK_IMPLEMENT atype optargs optreturns topseq
    {
        xtracer.Trace("parser.p_top_implement_callatom_lcb_action_rcb ENTER (top)")
        $$ = $1
        atom := atypeToAtom(parser16Acfg(parser16lex), $3)
        atom.SetLineno(tok16Lineno(parser16lex.(*parser16LexAdapter), $2))
        handleBeforeAfterV16(parser16Acfg(parser16lex), "implement", atom, $6, $$, $4, $5)
    }
    | top opttrusted PARSER16_TOK_ISOLATE SYMBOLx optargs PARSER16_TOK_EQ callatoms
    {
        xtracer.Trace("parser.p_top_opttrusted_isolate_callatom_eq_callatoms ENTER (top)")
        $$ = $1
        lex := parser16lex.(*parser16LexAdapter)
        nameAtom := parser16Acfg(parser16lex).NewAtom($4, $5...)
        elems := append([]Node{nameAtom}, $7...)
        idef := parser16Acfg(parser16lex).NewIsolateDef(elems, 0)
        idef.Trusted = $2
        idef.Elems[0].SetLineno(tok16Lineno(lex, $3))
        idef.SetLineno(tok16Lineno(lex, $3))
        $$.declare(parser16Acfg(parser16lex).NewIsolateDecl(idef))
    }
    | top opttrusted PARSER16_TOK_ISOLATE SYMBOLx optargs PARSER16_TOK_EQ callatoms PARSER16_TOK_WITH callatoms
    {
        xtracer.Trace("parser.p_top_opttrusted_isolate_callatom_eq_callatoms_with_callatoms ENTER (top)")
        $$ = $1
        lex := parser16lex.(*parser16LexAdapter)
        nameAtom := parser16Acfg(parser16lex).NewAtom($4, $5...)
        elems := append(append([]Node{nameAtom}, $7...), $9...)
        idef := parser16Acfg(parser16lex).NewIsolateDef(elems, len($9))
        idef.Trusted = $2
        idef.Elems[0].SetLineno(tok16Lineno(lex, $3))
        idef.SetLineno(tok16Lineno(lex, $3))
        $$.declare(parser16Acfg(parser16lex).NewIsolateDecl(idef))
    }
    | top PARSER16_TOK_DELEGATE callatoms optdelegee
    {
        xtracer.Trace("parser.p_top_delegate_callatom_opt ENTER (top)")
        $$ = $1
        args := make([]Node, len($3))
        for i, s := range $3 {
            if $4 != nil {
                args[i] = parser16Acfg(parser16lex).NewDelegateDef([]Node{s, $4})
            } else {
                args[i] = parser16Acfg(parser16lex).NewDelegateDef([]Node{s})
            }
        }
        $$.declare(parser16Acfg(parser16lex).NewDelegateDecl(args...))
    }
    | top PARSER16_TOK_INTERPRET oper PARSER16_TOK_ARROW oper
    {
        xtracer.Trace("parser.p_top_interpret_symbol_arrow_symbol ENTER (top)")
        $$ = $1
        lex := parser16lex.(*parser16LexAdapter)
        imp := parser16Acfg(parser16lex).NewImplies($3, $5)
        imp.SetLineno(tok16Lineno(lex, $4))
        d := parser16Acfg(parser16lex).NewInterpretDecl(parser17MkLF(parser16Acfg(parser16lex), imp))
        d.SetLineno(tok16Lineno(lex, $4))
        $$.declare(d)
    }
    | top PARSER16_TOK_INTERPRET oper PARSER16_TOK_ARROW PARSER16_TOK_LCB term PARSER16_TOK_DOTS term PARSER16_TOK_RCB
    {
        xtracer.Trace("parser.p_top_interpret_symbol_arrow_lcb_symbol_dots_symbol_rcb ENTER (top)")
        $$ = $1
        lex := parser16lex.(*parser16LexAdapter)
        rng := parser16Acfg(parser16lex).NewRange($6, $8)
        imp := parser16Acfg(parser16lex).NewImplies($3, rng)
        imp.SetLineno(tok16Lineno(lex, $4))
        d := parser16Acfg(parser16lex).NewInterpretDecl(parser17MkLF(parser16Acfg(parser16lex), imp))
        d.SetLineno(tok16Lineno(lex, $4))
        $$.declare(d)
    }
    | top PARSER16_TOK_INTERPRET oper PARSER16_TOK_ARROW PARSER16_TOK_LCB SYMBOLx moresymbols PARSER16_TOK_RCB
    {
        xtracer.Trace("parser.p_top_interpret_symbol_arrow_lcb_symbol_moresymbols_rcb ENTER (top)")
        $$ = $1
        vals := append([]Node{parser16Acfg(parser16lex).NewAtom($6)}, $7...)
        es := parser16Acfg(parser16lex).NewEnumeratedSort(vals...)
        imp := parser16Acfg(parser16lex).NewImplies($3, es)
        imp.SetLineno(tok16Lineno(parser16lex.(*parser16LexAdapter), $4))
        d := parser16Acfg(parser16lex).NewInterpretDecl(parser17MkLF(parser16Acfg(parser16lex), imp))
        d.SetLineno(tok16Lineno(parser16lex.(*parser16LexAdapter), $4))
        $$.declare(d)
    }
    | top PARSER16_TOK_ATTRIBUTE callatom PARSER16_TOK_EQ attributeval
    {
        xtracer.Trace("parser.p_top_attribute_callatom_eq_attributeval ENTER (top)")
        $$ = $1
        lex := parser16lex.(*parser16LexAdapter)
        adef := parser16Acfg(parser16lex).NewAttributeDef($3, $5)
        adef.SetLineno(tok16Lineno(lex, $2))
        d := parser16Acfg(parser16lex).NewAttributeDecl(adef)
        d.SetLineno(tok16Lineno(lex, $2))
        $$.declare(d)
    }
    | top PARSER16_TOK_VARIANT typesymbol PARSER16_TOK_OF atype
    {
        xtracer.Trace("parser.p_top_variant_symbol_of_atype ENTER (top)")
        $$ = $1
        scnst := parser16Acfg(parser16lex).NewAtom($3.(*Atom).Rep)
        scnst.SetLineno(nodeLineno($3))
        tdfn := parser16Acfg(parser16lex).NewTypeDef(scnst, parser16Acfg(parser16lex).NewUninterpretedSortAST())
        tdfn.SetLineno(tok16Lineno(parser16lex.(*parser16LexAdapter), $4))
        $$.declare(parser16Acfg(parser16lex).NewTypeDecl(tdfn))
        vdfn := parser16Acfg(parser16lex).NewVariantDef(scnst, parser16Acfg(parser16lex).NewAtom(NodeRep($5)))
        $$.declare(parser16Acfg(parser16lex).NewVariantDecl(vdfn))
    }
    | top PARSER16_TOK_VARIANT typesymbol PARSER16_TOK_OF atype PARSER16_TOK_EQ sort
    {
        xtracer.Trace("parser.p_top_variant_symbol_of_symbol_eq_sort ENTER (top)")
        $$ = $1
        scnst := parser16Acfg(parser16lex).NewAtom($3.(*Atom).Rep)
        scnst.SetLineno(nodeLineno($3))
        tdfn := parser16Acfg(parser16lex).NewTypeDef(scnst, $7)
        tdfn.SetLineno(tok16Lineno(parser16lex.(*parser16LexAdapter), $4))
        $$.declare(parser16Acfg(parser16lex).NewTypeDecl(tdfn))
        vdfn := parser16Acfg(parser16lex).NewVariantDef(scnst, parser16Acfg(parser16lex).NewAtom(NodeRep($5)))
        $$.declare(parser16Acfg(parser16lex).NewVariantDecl(vdfn))
    }
    | top PARSER16_TOK_EXPORT callatom
    {
        xtracer.Trace("parser.p_top_export_callatom ENTER (top)")
        $$ = $1
        d := parser16Acfg(parser16lex).NewExportDecl(parser16Acfg(parser16lex).NewExportDef($3, parser16Acfg(parser16lex).NewAtom("")))
        d.SetLineno(tok16Lineno(parser16lex.(*parser16LexAdapter), $2))
        $$.declare(d)
    }
    | top PARSER16_TOK_IMPORT callatom
    {
        xtracer.Trace("parser.p_top_import_callatom ENTER (top)")
        $$ = $1
        d := parser16Acfg(parser16lex).NewImportDecl(parser16Acfg(parser16lex).NewImportDef($3, parser16Acfg(parser16lex).NewAtom("")))
        d.SetLineno(tok16Lineno(parser16lex.(*parser16LexAdapter), $2))
        $$.declare(d)
    }
    | top PARSER16_TOK_PRIVATE callatom
    {
        xtracer.Trace("parser.p_top_private_callatom ENTER (top)")
        $$ = $1
        d := parser16Acfg(parser16lex).NewPrivateDecl(parser16Acfg(parser16lex).NewPrivateDef($3))
        d.SetLineno(tok16Lineno(parser16lex.(*parser16LexAdapter), $2))
        $$.declare(d)
    }
    | top PARSER16_TOK_ALIAS SYMBOLx PARSER16_TOK_EQ callatom
    {
        xtracer.Trace("parser.p_top_aliase_symbol_eq_callatom ENTER (top)")
        $$ = $1
        d := parser16Acfg(parser16lex).NewAliasDecl(parser16Acfg(parser16lex).NewDefinition(parser16Acfg(parser16lex).NewAtom($3), $5))
        d.SetLineno(tok16Lineno(parser16lex.(*parser16LexAdapter), $2))
        $$.declare(d)
    }
    | top PARSER16_TOK_DEFINITION defns optproof
    {
        xtracer.Trace("parser.p_top_definition_defns ENTER (top)")
        $$ = $1
        lfs := make([]Node, 0, len($3))
        for _, def := range $3 {
            lfs = append(lfs, parser17MkLF(parser16Acfg(parser16lex), def))
        }
        d := parser16Acfg(parser16lex).NewDefinitionDecl(lfs...)
        d.SetLineno(tok16Lineno(parser16lex.(*parser16LexAdapter), $2))
        $$.declare(d)
        if $4 != nil {
            $$.declare(parser16Acfg(parser16lex).NewProofDecl($4))
        }
    }
    ;

optproof:
    /* empty */
    {
        xtracer.Trace("parser.p_optproof ENTER (optproof)")
        $$ = nil
    }
    | PARSER16_TOK_PROOF proofstep
    {
        xtracer.Trace("parser.p_optproof_symbol ENTER (optproof)")
        $$ = $2
    }
    ;

proofstep:
    SYMBOLx
    {
        xtracer.Trace("parser.p_proofstep_symbol ENTER (proofstep)")
        a := parser16Acfg(parser16lex).NewAtom($1)
        si := parser16Acfg(parser16lex).NewSchemaInstantiation(a, parser16Acfg(parser16lex).NewRenaming(nil))
        $$ = si
    }
    | SYMBOLx PARSER16_TOK_WITH matches
    {
        xtracer.Trace("parser.p_proofstep_symbol_with_defns ENTER (proofstep)")
        a := parser16Acfg(parser16lex).NewAtom($1)
        si := parser16Acfg(parser16lex).NewSchemaInstantiationWithMatches(a, parser16Acfg(parser16lex).NewRenaming(nil), $3)
        $$ = si
    }
    ;

match:
    defn
    {
        xtracer.Trace("parser.p_match_defn ENTER (match)")
        $$ = $1
    }
    | var PARSER16_TOK_EQ fmla
    {
        xtracer.Trace("parser.p_match_var_eq_fmla ENTER (match)")
        $$ = parser16Acfg(parser16lex).NewDefinition($1, checkNonTemporal($3))
        $$.SetLineno(tok16Lineno(parser16lex.(*parser16LexAdapter), $2))
    }
    ;

matches:
    match
    {
        xtracer.Trace("parser.p_matches ENTER (matches)")
        $$ = []Node{$1}
    }
    | matches PARSER16_TOK_COMMA match
    {
        xtracer.Trace("parser.p_matches_matches_comma_match ENTER (matches)")
        $$ = append($1, $3)
    }
    ;

optskolem:
    /* empty */
    {
        xtracer.Trace("parser.p_optskolem ENTER (optskolem)")
        $$ = nil
    }
    | PARSER16_TOK_NAMED tterm
    {
        xtracer.Trace("parser.p_optskolem_symbol ENTER (optskolem)")
        $$ = AppToAtom($2)
        $$.SetLineno(tok16Lineno(parser16lex.(*parser16LexAdapter), $1))
    }
    ;

opttemporal:
    /* empty */
    {
        xtracer.Trace("parser.p_opttemporal ENTER (opttemporal)")
        $$ = nil
    }
    | PARSER16_TOK_TEMPORAL
    {
        xtracer.Trace("parser.p_opttemporal_symbol ENTER (opttemporal)")
        $$ = parser16Acfg(parser16lex).NewAtom("temporal")
    }
    ;

optimpex:
    /* empty */
    {
        xtracer.Trace("parser.p_optimpex ENTER (optimpex)")
        $$ = nil
    }
    | PARSER16_TOK_EXPORT
    {
        xtracer.Trace("parser.p_optimpex_export ENTER (optimpex)")
        $$ = parser16Acfg(parser16lex).NewExportDecl()
    }
    | PARSER16_TOK_IMPORT
    {
        xtracer.Trace("parser.p_optimpex_import ENTER (optimpex)")
        $$ = parser16Acfg(parser16lex).NewImportDecl()
    }
    ;

topseq:
    sequence
    {
        xtracer.Trace("parser.p_topseq_sequence ENTER (topseq)")
        $$ = $1
    }
    ;

opttrusted:
    /* empty */
    {
        xtracer.Trace("parser.p_opttrusted ENTER (opttrusted)")
        $$ = false
    }
    | PARSER16_TOK_TRUSTED
    {
        xtracer.Trace("parser.p_opttrusted_trusted ENTER (opttrusted)")
        $$ = true
    }
    ;

optdelegee:
    /* empty */
    {
        xtracer.Trace("parser.p_optdelegee ENTER (optdelegee)")
        $$ = nil
    }
    | PARSER16_TOK_ARROW callatom
    {
        xtracer.Trace("parser.p_optdelegee_callatom ENTER (optdelegee)")
        $$ = $2
    }
    ;

oper:
    atype
    {
        xtracer.Trace("parser.p_oper_symbol ENTER (oper)")
        $$ = atypeToAtom(parser16Acfg(parser16lex), $1)
    }
    | relop
    {
        xtracer.Trace("parser.p_oper_relop ENTER (oper)")
        $$ = parser16Acfg(parser16lex).NewAtom($1)
    }
    ;

attributeval:
    callatom
    {
        xtracer.Trace("parser.p_top_attributeval_callatom ENTER (attributeval)")
        $$ = $1
    }
    | PARSER16_TOK_TRUE
    {
        xtracer.Trace("parser.p_top_attributeval_true ENTER (attributeval)")
        $$ = parser16Acfg(parser16lex).NewAtom("true")
        $$.SetLineno(tok16Lineno(parser16lex.(*parser16LexAdapter), $1))
    }
    | PARSER16_TOK_FALSE
    {
        xtracer.Trace("parser.p_top_attributeval_false ENTER (attributeval)")
        $$ = parser16Acfg(parser16lex).NewAtom("false")
        $$.SetLineno(tok16Lineno(parser16lex.(*parser16LexAdapter), $1))
    }
    ;

moresymbols:
    /* empty */
    {
        xtracer.Trace("parser.p_moresymbols ENTER (moresymbols)")
        $$ = nil
    }
    | moresymbols PARSER16_TOK_COMMA SYMBOLx
    {
        xtracer.Trace("parser.p_moresymbols_more_symbols_comma_symbol ENTER (moresymbols)")
        $$ = append($1, parser16Acfg(parser16lex).NewAtom($3))
    }
    ;

objectend:
    /* empty */
    {
        xtracer.Trace("parser.p_objectend ENTER (objectend)")
        parser16lex.(*parser16LexAdapter).accum.isObject = false
        $$ = nil
    }
    ;

optdotdotdot:
    /* empty */
    {
        xtracer.Trace("parser.p_optdotdotdot ENTER (optdotdotdot)")
        parser16lex.(*parser16LexAdapter).parentObjName = ""
        $$ = false
    }
    | PARSER16_TOK_DOTDOTDOT
    {
        xtracer.Trace("parser.p_optdotdotdot_dotdotdot ENTER (optdotdotdot)")
        $$ = true
    }
    ;

objectargs:
    optargs
    {
        xtracer.Trace("parser.p_objectargs_optargs ENTER (objectargs)")
        $$ = $1
        parser16lex.(*parser16LexAdapter).accum.params = $1
    }
    ;

objsym:
    SYMBOLx
    {
        xtracer.Trace("parser.p_objsym ENTER (objsym)")
        lex := parser16lex.(*parser16LexAdapter)
        $$ = parser16Acfg(parser16lex).NewAtom($1)
        lex.parentObjName = $1
    }
    ;

optactiondef:
    /* empty */
    {
        xtracer.Trace("parser.p_optactiondef ENTER (optactiondef)")
        $$ = parser16Acfg(parser16lex).NewSequence()
    }
    | PARSER16_TOK_EQ sequence
    {
        xtracer.Trace("parser.p_optactiondef_eq_sequence ENTER (optactiondef)")
        $$ = $2
    }
    ;

sequence:
    PARSER16_TOK_LCB PARSER16_TOK_RCB
    {
        xtracer.Trace("parser.p_sequence_lcb_rcb ENTER (sequence)")
        $$ = parser16Acfg(parser16lex).NewSequence()
        $$.SetLineno(tok16Lineno(parser16lex.(*parser16LexAdapter), $1))
    }
    | PARSER16_TOK_LCB actseq PARSER16_TOK_RCB
    {
        xtracer.Trace("parser.p_sequence_lcb_actseq_rcb ENTER (sequence)")
        stmts := lowerVarStmts($2)
        seq := parser17MakeSequence(parser16Acfg(parser16lex), stmts)
        if s, ok := seq.(*Sequence); ok {
            s.SetLineno(tok16Lineno(parser16lex.(*parser16LexAdapter), $1))
        }
        $$ = seq
    }
    | PARSER16_TOK_LCB actseq PARSER16_TOK_SEMI PARSER16_TOK_RCB
    {
        xtracer.Trace("parser.p_sequence_lcb_actseq_semi_rcb ENTER (sequence)")
        stmts := lowerVarStmts($2)
        seq := parser16Acfg(parser16lex).NewSequence(stmts...)
        seq.SetLineno(tok16Lineno(parser16lex.(*parser16LexAdapter), $1))
        $$ = seq
    }
    ;

actseq:
    actseqrev
    {
        xtracer.Trace("parser.p_actseq_actseqrev ENTER (actseq)")
        for i, j := 0, len($1)-1; i < j; i, j = i+1, j-1 {
            $1[i], $1[j] = $1[j], $1[i]
        }
        $$ = $1
    }
    ;

actseqrev:
    simpleact
    {
        xtracer.Trace("parser.p_actseqrev_simpleact ENTER (actseqrev)")
        $$ = []Node{$1}
    }
    | complexact
    {
        xtracer.Trace("parser.p_actseqrev_complexact ENTER (actseqrev)")
        $$ = []Node{$1}
    }
    | simpleact PARSER16_TOK_SEMI actseqrev
    {
        xtracer.Trace("parser.p_actseqrev_simpact_semi_actseqrev ENTER (actseqrev)")
        $$ = append($3, $1)
    }
    | simpleact PARSER16_TOK_SEMI
    {
        xtracer.Trace("parser.p_actseqrev_simpact_semi ENTER (actseqrev)")
        $$ = []Node{$1}
    }
    | complexact actseqrev
    {
        xtracer.Trace("parser.p_actseqrev_complexact_actseqrev ENTER (actseqrev)")
        $$ = append($2, $1)
    }
    | complexact PARSER16_TOK_SEMI actseqrev
    {
        xtracer.Trace("parser.p_actseqrev_complexact_semi_actseqrev ENTER (actseqrev)")
        $$ = append($3, $1)
    }
    | complexact PARSER16_TOK_SEMI
    {
        xtracer.Trace("parser.p_actseqrev_complexact_semi ENTER (actseqrev)")
        $$ = []Node{$1}
    }
    ;

action:
    simpleact
    {
        xtracer.Trace("parser.p_action_simpleact ENTER (action)")
        $$ = $1
    }
    | complexact
    {
        xtracer.Trace("parser.p_action_complexact ENTER (action)")
        $$ = $1
    }
    ;

simpleact:
    PARSER16_TOK_ASSUME labeledfmla
    {
        xtracer.Trace("parser.p_action_assume ENTER (simpleact)")
        a := parser16Acfg(parser16lex).NewAssumeAction(checkNonTemporal($2))
        a.SetLineno(tok16Lineno(parser16lex.(*parser16LexAdapter), $1))
        $$ = a
    }
    | PARSER16_TOK_ASSERT labeledfmla
    {
        xtracer.Trace("parser.p_action_assert ENTER (simpleact)")
        a := parser16Acfg(parser16lex).NewAssertAction(checkNonTemporal($2))
        a.SetLineno(tok16Lineno(parser16lex.(*parser16LexAdapter), $1))
        $$ = a
    }
    | PARSER16_TOK_ENSURES labeledfmla
    {
        xtracer.Trace("parser.p_action_ensures ENTER (simpleact)")
        a := parser16Acfg(parser16lex).NewEnsuresAction(checkNonTemporal($2))
        a.SetLineno(tok16Lineno(parser16lex.(*parser16LexAdapter), $1))
        $$ = a
    }
    | PARSER16_TOK_CALL optactualreturns callatom
    {
        xtracer.Trace("parser.p_action_call_callatom ENTER (simpleact)")
        callArgs := append([]Node{$3}, $2...)
        $$ = parser16Acfg(parser16lex).NewCallAction(callArgs...)
        $$.SetLineno(tok16Lineno(parser16lex.(*parser16LexAdapter), $1))
    }
    | PARSER16_TOK_CALL callatom
    {
        xtracer.Trace("parser.p_action_call_callatom ENTER (simpleact)")
        $$ = parser16Acfg(parser16lex).NewCallAction($2)
        $$.SetLineno(tok16Lineno(parser16lex.(*parser16LexAdapter), $1))
    }
    | PARSER16_TOK_SET lit
    {
        xtracer.Trace("parser.p_action_set_lit ENTER (simpleact)")
        a := parser16Acfg(parser16lex).NewSetAction($2)
        a.SetLineno(tok16Lineno(parser16lex.(*parser16LexAdapter), $1))
        $$ = a
    }
    | PARSER16_TOK_INSTANTIATE callatom
    {
        xtracer.Trace("parser.p_action_instantiate_atom ENTER (simpleact)")
        a := parser16Acfg(parser16lex).NewInstantiateAction($2)
        a.SetLineno(tok16Lineno(parser16lex.(*parser16LexAdapter), $1))
        $$ = a
    }
    | PARSER16_TOK_VAR tterm optinit
    {
        xtracer.Trace("parser.p_action_var_opttypedsym_assign_fmla ENTER (simpleact)")
        if $3 != nil {
            $$ = parser16Acfg(parser16lex).NewVarAction($2, $3)
        } else {
            $$ = parser16Acfg(parser16lex).NewVarAction($2)
        }
        $$.SetLineno(tok16Lineno(parser16lex.(*parser16LexAdapter), $1))
    }
    | term PARSER16_TOK_ASSIGN fmla
    {
        xtracer.Trace("parser.p_action_assign ENTER (simpleact)")
        $$ = parser16Acfg(parser16lex).NewAssignAction($1, checkNonTemporal($3))
        $$.SetLineno(tok16Lineno(parser16lex.(*parser16LexAdapter), $2))
    }
    | termtuple PARSER16_TOK_ASSIGN callatom
    {
        xtracer.Trace("parser.p_action_termtuple_assign_fmla ENTER (simpleact)")
        callArgs := append([]Node{$3}, $1.Args()...)
        $$ = parser16Acfg(parser16lex).NewCallAction(callArgs...)
        $$.SetLineno(tok16Lineno(parser16lex.(*parser16LexAdapter), $2))
    }
    | term PARSER16_TOK_ASSIGN PARSER16_TOK_TIMES
    {
        xtracer.Trace("parser.p_action_term_assign_times ENTER (simpleact)")
        $$ = parser16Acfg(parser16lex).NewHavocAction($1)
        $$.SetLineno(tok16Lineno(parser16lex.(*parser16LexAdapter), $2))
    }
    | term     %prec PARSER16_TOK_SEMI
    {
        xtracer.Trace("parser.p_action_term ENTER (simpleact)")
        $$ = parser16Acfg(parser16lex).NewCallAction($1)
        $$.SetLineno($1.GetLineno())
    }
    ;

complexact:
    sequence
    {
        xtracer.Trace("parser.p_action_sequence ENTER (complexact)")
        $$ = $1
    }
    | PARSER16_TOK_IF somefmla sequence
    {
        xtracer.Trace("parser.p_action_if_fmla_lcb_action_rcb ENTER (complexact)")
        cond := checkNonTemporal($2)
        body := fixIfPart(cond, $3)
        a := parser16Acfg(parser16lex).NewIfAction(cond, body, nil)
        a.SetLineno(tok16Lineno(parser16lex.(*parser16LexAdapter), $1))
        $$ = a
    }
    | PARSER16_TOK_IF somefmla sequence PARSER16_TOK_ELSE action
    {
        xtracer.Trace("parser.p_action_if_fmla_lcb_action_rcb_else_LCB_action_RCB ENTER (complexact)")
        cond := checkNonTemporal($2)
        body := fixIfPart(cond, $3)
        a := parser16Acfg(parser16lex).NewIfAction(cond, body, $5)
        a.SetLineno(tok16Lineno(parser16lex.(*parser16LexAdapter), $1))
        $$ = a
    }
    | PARSER16_TOK_IF PARSER16_TOK_TIMES sequence PARSER16_TOK_ELSE action
    {
        xtracer.Trace("parser.p_action_if_times_lcb_action_rcb_else_LCB_action_RCB ENTER (complexact)")
        choice := parser16Acfg(parser16lex).NewChoiceAction($3, $5)
        choice.SetLineno(tok16Lineno(parser16lex.(*parser16LexAdapter), $1))
        $$ = choice
    }
    | PARSER16_TOK_WHILE somefmla invariants decreases sequence
    {
        xtracer.Trace("parser.p_action_while_somefmla_invariants_decreases_lcb_action_rcb ENTER (complexact)")
        cond := checkNonTemporal($2)
        body := fixIfPart($2, $5)
        args := []Node{cond, body}
        args = append(args, $3...)
        args = append(args, $4...)
        w := parser16Acfg(parser16lex).NewWhileAction(args...)
        w.SetLineno(tok16Lineno(parser16lex.(*parser16LexAdapter), $1))
        $$ = w
    }
    | PARSER16_TOK_LOCAL lparams sequence
    {
        xtracer.Trace("parser.p_action_local_params_lcb_action_rcb ENTER (complexact)")
        bounds := $2
        lsyms := make([]Node, len(bounds))
        subst := make(map[string]string)
        for i, s := range bounds {
            lsyms[i] = PrefixNode(s, "loc:")
            subst[NodeRep(s)] = NodeRep(lsyms[i])
        }
        action := SubstPrefixAtomsAst($3, subst, nil, nil, nil)
        args := append(lsyms, action)
        la := parser16Acfg(parser16lex).NewLocalAction("parser.local_action_rule", args...)
        la.SetLineno(tok16Lineno(parser16lex.(*parser16LexAdapter), $1))
        $$ = la
    }
    | PARSER16_TOK_LET eqns sequence
    {
        xtracer.Trace("parser.p_action_let_eqns_lcb_action_rcb ENTER (complexact)")
        args := append($2, $3)
        la := parser16Acfg(parser16lex).NewLetAction(args...)
        la.SetLineno(tok16Lineno(parser16lex.(*parser16LexAdapter), $1))
        $$ = la
    }
    ;

somefmla:
    fmla
    {
        xtracer.Trace("parser.p_somefmla_fmla ENTER (somefmla)")
        $$ = $1
    }
    | fmla PARSER16_TOK_ASSIGN fmla
    {
        xtracer.Trace("parser.p_somefmla_fmla_assign_fmla ENTER (somefmla)")
        lhs := $1
        lsym := PrefixNode(lhs, "loc:")
        if a, ok := lsym.(*Atom); ok {
            if orig, ok2 := lhs.(*Atom); ok2 {
                a.ASort = orig.ASort
            }
        }
        subst := map[string]string{NodeRep(lhs): NodeRep(lsym)}
        fmla := parser16Acfg(parser16lex).NewApp(parser16Acfg(parser16lex).NewSymbol("*>", nil), $3, lhs)
        fmla.SetLineno(tok16Lineno(parser16lex.(*parser16LexAdapter), $2))
        fmla2 := SubstPrefixAtomsAst(fmla, subst, nil, nil, nil)
        some := parser16Acfg(parser16lex).NewSome([]Node{lsym}, fmla2)
        some.SetLineno(tok16Lineno(parser16lex.(*parser16LexAdapter), $2))
        $$ = some
    }
    | PARSER16_TOK_SOME bounds fmla
    {
        xtracer.Trace("parser.p_somefmla_some_bounds_fmla ENTER (somefmla)")
        bounds := $2
        lsyms := make([]Node, len(bounds))
        subst := make(map[string]string)
        for i, s := range bounds {
            lsyms[i] = PrefixNode(s, "loc:")
            subst[NodeRep(s)] = NodeRep(lsyms[i])
        }
        fmla := SubstPrefixAtomsAst($3, subst, nil, nil, nil)
        some := parser16Acfg(parser16lex).NewSome(lsyms, fmla)
        some.SetLineno(tok16Lineno(parser16lex.(*parser16LexAdapter), $1))
        $$ = some
    }
    | PARSER16_TOK_SOME bounds fmla PARSER16_TOK_MINIMIZING term
    {
        xtracer.Trace("parser.p_somefmla_some_bounds_fmla_minimizing_term ENTER (somefmla)")
        bounds := $2
        lsyms := make([]Node, len(bounds))
        subst := make(map[string]string)
        for i, s := range bounds {
            lsyms[i] = PrefixNode(s, "loc:")
            subst[NodeRep(s)] = NodeRep(lsyms[i])
        }
        fmla := SubstPrefixAtomsAst($3, subst, nil, nil, nil)
        index := SubstPrefixAtomsAst($5, subst, nil, nil, nil)
        smin := parser16Acfg(parser16lex).NewSomeMin(lsyms, fmla, index)
        smin.SetLineno(tok16Lineno(parser16lex.(*parser16LexAdapter), $1))
        $$ = smin
    }
    | PARSER16_TOK_SOME bounds fmla PARSER16_TOK_MAXIMIZING term
    {
        xtracer.Trace("parser.p_somefmla_some_bounds_fmla_maximizing_term ENTER (somefmla)")
        bounds := $2
        lsyms := make([]Node, len(bounds))
        subst := make(map[string]string)
        for i, s := range bounds {
            lsyms[i] = PrefixNode(s, "loc:")
            subst[NodeRep(s)] = NodeRep(lsyms[i])
        }
        fmla := SubstPrefixAtomsAst($3, subst, nil, nil, nil)
        index := SubstPrefixAtomsAst($5, subst, nil, nil, nil)
        smax := parser16Acfg(parser16lex).NewSomeMax(lsyms, fmla, index)
        smax.SetLineno(tok16Lineno(parser16lex.(*parser16LexAdapter), $1))
        $$ = smax
    }
    ;

bounds:
    params PARSER16_TOK_DOT
    {
        xtracer.Trace("parser.p_bounds_params_dot ENTER (bounds)")
        $$ = $1
    }
    | PARSER16_TOK_LPAREN lparams PARSER16_TOK_RPAREN
    {
        xtracer.Trace("parser.p_bounds_lparen_lparams_rparen ENTER (bounds)")
        $$ = $2
    }
    ;

invariants:
    /* empty */
    {
        xtracer.Trace("parser.p_invariants ENTER (invariants)")
        $$ = nil
    }
    | invariants PARSER16_TOK_INVARIANT labeledfmla
    {
        xtracer.Trace("parser.p_invariant_invariant_fmla ENTER (invariants)")
        inv := checkNonTemporal($3)
        a := parser16Acfg(parser16lex).NewAssertAction(inv)
        a.SetLineno(tok16Lineno(parser16lex.(*parser16LexAdapter), $2))
        $$ = append($1, a)
    }
    | invariants PARSER16_TOK_INVARIANT labeledfmla PARSER16_TOK_PROOF proofstep
    {
        xtracer.Trace("parser.p_invariant_invariant_fmla_proof ENTER (invariants)")
        inv := checkNonTemporal($3)
        a := parser16Acfg(parser16lex).NewAssertAction(inv, $5)
        a.SetLineno(tok16Lineno(parser16lex.(*parser16LexAdapter), $2))
        $$ = append($1, a)
    }
    ;

decreases:
    /* empty */
    {
        xtracer.Trace("parser.p_decreases ENTER (decreases)")
        $$ = nil
    }
    | PARSER16_TOK_DECREASES fmla
    {
        xtracer.Trace("parser.p_decreases_decreases_fmla ENTER (decreases)")
        fmla := checkNonTemporal($2)
        rank := parser16Acfg(parser16lex).NewRanking(fmla)
        rank.SetLineno(tok16Lineno(parser16lex.(*parser16LexAdapter), $1))
        $$ = []Node{rank}
    }
    ;

eqn:
    SYMBOLx PARSER16_TOK_EQ SYMBOLx
    {
        xtracer.Trace("parser.p_eqn_SYMBOL_EQ_SYMBOL ENTER (eqn)")
        $$ = parser16Acfg(parser16lex).NewAtom("=", parser16Acfg(parser16lex).NewApp(parser16Acfg(parser16lex).NewSymbol($1, nil)), parser16Acfg(parser16lex).NewApp(parser16Acfg(parser16lex).NewSymbol($3, nil)))
    }
    ;

eqns:
    eqn
    {
        xtracer.Trace("parser.p_eqns_eqn ENTER (eqns)")
        $$ = []Node{$1}
    }
    | eqns PARSER16_TOK_COMMA eqn
    {
        xtracer.Trace("parser.p_eqns_eqns_comma_eqn ENTER (eqns)")
        $$ = append($1, $3)
    }
    ;

lit:
    atom
    {
        xtracer.Trace("parser.p_lit_atom ENTER (lit)")
        $$ = parser16Acfg(parser16lex).NewLiteral(1, $1)
        $$.SetLineno(nodeLineno($1))
    }
    | SYMBOLx PARSER16_TOK_EQ SYMBOLx
    {
        xtracer.Trace("parser.p_lit_term_eq_term ENTER (lit)")
        a := parser16Acfg(parser16lex).NewAtom("=", parser16Acfg(parser16lex).NewAtom($1), parser16Acfg(parser16lex).NewAtom($3))
        $$ = parser16Acfg(parser16lex).NewLiteral(1, a)
        $$.SetLineno(tok16Lineno(parser16lex.(*parser16LexAdapter), $2))
    }
    | SYMBOLx PARSER16_TOK_TILDAEQ SYMBOLx
    {
        xtracer.Trace("parser.p_lit_term_tildaeq_term ENTER (lit)")
        a := parser16Acfg(parser16lex).NewAtom("=", parser16Acfg(parser16lex).NewAtom($1), parser16Acfg(parser16lex).NewAtom($3))
        $$ = parser16Acfg(parser16lex).NewLiteral(0, a)
        $$.SetLineno(tok16Lineno(parser16lex.(*parser16LexAdapter), $2))
    }
    | PARSER16_TOK_TILDA lit
    {
        xtracer.Trace("parser.p_lit_tilda_atom ENTER (lit)")
        if lit, ok := $2.(*Literal); ok {
            $$ = parser16Acfg(parser16lex).NewLiteral(1 - lit.Polarity, lit.Atom)
        } else {
            $$ = parser16Acfg(parser16lex).NewLiteral(0, $2)
        }
        $$.SetLineno(tok16Lineno(parser16lex.(*parser16LexAdapter), $1))
    }
    ;

optargs:
    /* empty */
    {
        xtracer.Trace("parser.p_optargs ENTER (optargs)")
        $$ = nil
    }
    | PARSER16_TOK_LPAREN lparams PARSER16_TOK_RPAREN
    {
        xtracer.Trace("parser.p_optargs_params ENTER (optargs)")
        $$ = $2
    }
    ;

optreturns:
    /* empty */
    {
        xtracer.Trace("parser.p_optreturns ENTER (optreturns)")
        $$ = nil
    }
    | PARSER16_TOK_RETURNS PARSER16_TOK_LPAREN lparams PARSER16_TOK_RPAREN
    {
        xtracer.Trace("parser.p_optreturns_tsyms ENTER (optreturns)")
        $$ = $3
    }
    ;

optactualreturns:
    /* empty */
    {
        xtracer.Trace("parser.p_optactualreturns ENTER (optactualreturns)")
        $$ = nil
    }
    | callatoms PARSER16_TOK_ASSIGN
    {
        xtracer.Trace("parser.p_optactualreturns_callatoms_assign ENTER (optactualreturns)")
        $$ = $1
    }
    ;

lparam:
    SYMBOLx PARSER16_TOK_COLON atype
    {
        xtracer.Trace("parser.p_lparam_variable_colon_symbol ENTER (lparam)")
        a := parser16Acfg(parser16lex).NewApp(parser16Acfg(parser16lex).NewSymbol($1, nil))
        a.SetLineno(tok16Lineno(parser16lex.(*parser16LexAdapter), $2))
        a.ASort = $3
        $$ = a
    }
    ;

lparams:
    lparam
    {
        xtracer.Trace("parser.p_lparams_lparam ENTER (lparams)")
        $$ = []Node{$1}
    }
    | lparams PARSER16_TOK_COMMA lparam
    {
        xtracer.Trace("parser.p_lparams_lparams_comma_lparam ENTER (lparams)")
        $$ = append($1, $3)
    }
    ;

param:
    SYMBOLx PARSER16_TOK_COLON SYMBOLx
    {
        xtracer.Trace("parser.p_param_term_colon_symbol ENTER (param)")
        a := parser16Acfg(parser16lex).NewApp(parser16Acfg(parser16lex).NewSymbol($1, nil))
        a.SetLineno(tok16Lineno(parser16lex.(*parser16LexAdapter), $2))
        a.ASort = parser16Acfg(parser16lex).NewSymbol($3, nil)
        $$ = a
    }
    ;

params:
    param
    {
        xtracer.Trace("parser.p_params_param ENTER (params)")
        $$ = []Node{$1}
    }
    | params PARSER16_TOK_COMMA param
    {
        xtracer.Trace("parser.p_params_params_comma_param ENTER (params)")
        $$ = append($1, $3)
    }
    ;

optinit:
    /* empty */
    {
        xtracer.Trace("parser.p_optinit ENTER (optinit)")
        $$ = nil
    }
    | PARSER16_TOK_ASSIGN fmla
    {
        xtracer.Trace("parser.p_optinit_assign_fmla ENTER (optinit)")
        $$ = checkNonTemporal($2)
    }
    ;

termtuple:
    PARSER16_TOK_LPAREN term PARSER16_TOK_COMMA terms PARSER16_TOK_RPAREN
    {
        xtracer.Trace("parser.p_termtuple_lp_term_comma_terms_rp ENTER (termtuple)")
        args := append([]Node{$2}, $4...)
        t := parser16Acfg(parser16lex).NewTuple(args...)
        t.SetLineno(tok16Lineno(parser16lex.(*parser16LexAdapter), $1))
        $$ = t
    }
    ;

labeledfmla:
    fmla
    {
        xtracer.Trace("parser.p_labeledfmla_fmla ENTER (labeledfmla)")
        lf := parser16Acfg(parser16lex).NewLabeledFormula(nil, $1)
        if $1 != nil && $1.GetLineno().Line > 0 {
            lf.SetLineno($1.GetLineno())
        }
        $$ = lf
    }
    | labelname fmla
    {
        xtracer.Trace("parser.p_labeledfmla_label_fmla ENTER (labeledfmla)")
        lf := parser16Acfg(parser16lex).NewLabeledFormula(parser16Acfg(parser16lex).NewAtom($1.Val), $2)
        lf.SetLineno(tok16Lineno(parser16lex.(*parser16LexAdapter), $1))
        $$ = lf
    }
    ;

labelname:
    PARSER16_TOK_LB SYMBOLx PARSER16_TOK_RB
    {
        xtracer.Trace("parser.p_LABEL_LB_SYMBOL_RB ENTER (LABEL)")
        $$ = TokenInfo{Val: $2, Line: $1.Line}
    }
    ;

symdecl:
    constantdecl
    {
        xtracer.Trace("parser.p_symdecl_constantdecl ENTER (symdecl)")
        $$ = $1
    }
    ;

constantdecl:
    PARSER16_TOK_INDIV tterms
    {
        xtracer.Trace("parser.p_constantdecl_constant_tterms ENTER (constantdecl)")
        d := parser16Acfg(parser16lex).NewConstantDecl($2...)
        d.SetLineno(tok16Lineno(parser16lex.(*parser16LexAdapter), $1))
        $$ = d
    }
    | PARSER16_TOK_VAR tterms
    {
        xtracer.Trace("parser.p_constantdecl_var_tterms ENTER (constantdecl)")
        d := parser16Acfg(parser16lex).NewConstantDecl($2...)
        d.SetLineno(tok16Lineno(parser16lex.(*parser16LexAdapter), $1))
        $$ = d
    }
    ;

rel:
    aterm
    {
        xtracer.Trace("parser.p_rel_defnlhs ENTER (rel)")
        if a, ok := $1.(*Atom); ok {
            a.ASort = parser16Acfg(parser16lex).NewSymbol("bool", nil)
        }
        $$ = parser16Acfg(parser16lex).NewConstantDecl($1)
    }
    | defn
    {
        xtracer.Trace("parser.p_rel_defn ENTER (rel)")
        lf := parser17MkLF(parser16Acfg(parser16lex), $1)
        d := parser16Acfg(parser16lex).NewDerivedDecl(lf)
        $$ = d
    }
    ;

rels:
    rel
    {
        xtracer.Trace("parser.p_rels_rel ENTER (rels)")
        $$ = []Node{$1}
    }
    | rels PARSER16_TOK_COMMA rel
    {
        xtracer.Trace("parser.p_rels_rels_comma_rel ENTER (rels)")
        $$ = append($1, $3)
    }
    ;

fun:
    tterm
    {
        xtracer.Trace("parser.p_fun_defnlhs_colon_atype ENTER (fun)")
        d := parser16Acfg(parser16lex).NewConstantDecl($1)
        $$ = d
    }
    | tterm PARSER16_TOK_EQ fmla
    {
        xtracer.Trace("parser.p_fun_defn ENTER (fun)")
        df := parser16Acfg(parser16lex).NewDefinition(AppToAtom($1), checkNonTemporal($3))
        df.SetLineno(tok16Lineno(parser16lex.(*parser16LexAdapter), $2))
        lf := parser17MkLF(parser16Acfg(parser16lex), df)
        d := parser16Acfg(parser16lex).NewDerivedDecl(lf)
        $$ = d
    }
    ;

funs:
    fun
    {
        xtracer.Trace("parser.p_funs_fun ENTER (funs)")
        $$ = []Node{$1}
    }
    | funs PARSER16_TOK_COMMA fun
    {
        xtracer.Trace("parser.p_funs_funs_comma_fun ENTER (funs)")
        $$ = append($1, $3)
    }
    ;

defnrhs:
    fmla
    {
        xtracer.Trace("parser.p_defnrhs_fmla ENTER (defnrhs)")
        $$ = checkNonTemporal($1)
    }
    ;

defn:
    tterm PARSER16_TOK_EQ defnrhs
    {
        xtracer.Trace("parser.p_defn_atom_fmla ENTER (defn)")
        d := parser16Acfg(parser16lex).NewDefinition(AppToAtom($1), $3)
        d.SetLineno(tok16Lineno(parser16lex.(*parser16LexAdapter), $2))
        $$ = d
    }
    ;

defns:
    defn
    {
        xtracer.Trace("parser.p_defns_defn ENTER (defns)")
        $$ = []Node{$1}
    }
    | defns PARSER16_TOK_COMMA defn
    {
        xtracer.Trace("parser.p_defns_defns_comma_defn ENTER (defns)")
        $$ = append($1, $3)
    }
    ;

schdefnrhs:
    fmla
    {
        xtracer.Trace("parser.p_schdefnrhs_fmla ENTER (schdefnrhs)")
        $$ = checkNonTemporal($1)
    }
    | PARSER16_TOK_LCB schdecls schconc PARSER16_TOK_RCB
    {
        xtracer.Trace("parser.p_schdefnrhs_lcb_schdecls_rcb ENTER (schdefnrhs)")
        args := append($2, $3)
        $$ = parser16Acfg(parser16lex).NewSchemaBody(args...)
        $$.SetLineno(tok16Lineno(parser16lex.(*parser16LexAdapter), $1))
    }
    ;

schdecl:
    PARSER16_TOK_FUNCTION funs
    {
        xtracer.Trace("parser.p_schdecl_funcdecl ENTER (schdecl)")
        $$ = make([]Node, len($2))
        copy($$, $2)
    }
    | PARSER16_TOK_INDIV funs
    {
        xtracer.Trace("parser.p_schdecl_indivdecl ENTER (schdecl)")
        $$ = make([]Node, len($2))
        copy($$, $2)
    }
    | PARSER16_TOK_RELATION rels
    {
        xtracer.Trace("parser.p_schdecl_relation_rel ENTER (schdecl)")
        $$ = make([]Node, len($2))
        copy($$, $2)
    }
    | PARSER16_TOK_TYPE SYMBOLx
    {
        xtracer.Trace("parser.p_schdecl_typedecl ENTER (schdecl)")
        scnst := parser16Acfg(parser16lex).NewAtom($2)
        tdfn := parser16Acfg(parser16lex).NewTypeDef(scnst, parser16Acfg(parser16lex).NewUninterpretedSortAST())
        tdfn.SetLineno(tok16Lineno(parser16lex.(*parser16LexAdapter), $1))
        $$ = []Node{tdfn}
    }
    | PARSER16_TOK_PROPERTY labeledfmla
    {
        xtracer.Trace("parser.p_schdecl_propdecl ENTER (schdecl)")
        $$ = []Node{checkNonTemporal($2)}
    }
    | schdefnrhs
    {
        xtracer.Trace("parser.p_schdecl_theorem ENTER (schdecl)")
        lf := parser16Acfg(parser16lex).NewLabeledFormula(nil, $1)
        if $1 != nil && $1.GetLineno().Line > 0 {
            lf.SetLineno($1.GetLineno())
        }
        $$ = []Node{lf}
    }
    ;

schdecls:
    /* empty */
    {
        xtracer.Trace("parser.p_schdecls ENTER (schdecls)")
        $$ = nil
    }
    | schdecls schdecl
    {
        xtracer.Trace("parser.p_schdecls_schdecls_schdecl ENTER (schdecls)")
        $$ = append($1, $2...)
    }
    ;

schconc:
    PARSER16_TOK_DEFINITION defn
    {
        xtracer.Trace("parser.p_schconc_defdecl ENTER (schconc)")
        $$ = $2
    }
    | PARSER16_TOK_PROPERTY fmla
    {
        xtracer.Trace("parser.p_schconc_propdecl ENTER (schconc)")
        $$ = checkNonTemporal($2)
    }
    ;

schdefn:
    tterm PARSER16_TOK_EQ schdefnrhs
    {
        xtracer.Trace("parser.p_schdefn_atom_eq_fmla ENTER (schdefn)")
        $$ = parser16Acfg(parser16lex).NewDefinition(AppToAtom($1), $3)
        $$.SetLineno(tok16Lineno(parser16lex.(*parser16LexAdapter), $2))
    }
    ;

sort:
    PARSER16_TOK_LCB SYMBOLx PARSER16_TOK_RCB
    {
        xtracer.Trace("parser.p_sort_lcb_symbol_rcb ENTER (sort)")
        $$ = parser16Acfg(parser16lex).NewEnumeratedSort(parser16Acfg(parser16lex).NewAtom($2))
    }
    | PARSER16_TOK_LCB SYMBOLx PARSER16_TOK_COMMA names PARSER16_TOK_RCB
    {
        xtracer.Trace("parser.p_sort_lcb_names_rcb ENTER (sort)")
        elems := append([]Node{parser16Acfg(parser16lex).NewAtom($2)}, $4...)
        $$ = parser16Acfg(parser16lex).NewEnumeratedSort(elems...)
    }
    | PARSER16_TOK_LCB SYMBOLx PARSER16_TOK_DOTS SYMBOLx PARSER16_TOK_RCB
    {
        xtracer.Trace("parser.p_sort_lcb_symbol_dots_symbol_rcb ENTER (sort)")
        r := parser16Acfg(parser16lex).NewRange(parser16Acfg(parser16lex).NewAtom($2), parser16Acfg(parser16lex).NewAtom($4))
        r.SetLineno(tok16Lineno(parser16lex.(*parser16LexAdapter), $1))
        $$ = r
    }
    | PARSER16_TOK_STRUCT PARSER16_TOK_LCB tterms PARSER16_TOK_RCB
    {
        xtracer.Trace("parser.p_sort_struct_lcb_names_rcb ENTER (sort)")
        $$ = parser16Acfg(parser16lex).NewStructSort($3...)
    }
    | PARSER16_TOK_STRUCT PARSER16_TOK_LCB PARSER16_TOK_RCB
    {
        xtracer.Trace("parser.p_sort_struct_lcb_rcb ENTER (sort)")
        $$ = parser16Acfg(parser16lex).NewStructSort()
    }
    ;

names:
    SYMBOLx
    {
        xtracer.Trace("parser.p_names_symbol ENTER (names)")
        $$ = []Node{parser16Acfg(parser16lex).NewAtom($1)}
    }
    | names PARSER16_TOK_COMMA SYMBOLx
    {
        xtracer.Trace("parser.p_names_names_comma_symbol ENTER (names)")
        $$ = append($1, parser16Acfg(parser16lex).NewAtom($3))
    }
    ;

typesymbol:
    SYMBOLx
    {
        xtracer.Trace("parser.p_typesymbol_symbol ENTER (typesymbol)")
        $$ = parser16Acfg(parser16lex).NewAtom($1)
    }
    | PARSER16_TOK_THIS
    {
        xtracer.Trace("parser.p_typesymbol_this ENTER (typesymbol)")
        a := parser16Acfg(parser16lex).NewAtom("this")
        a.SetLineno(tok16Lineno(parser16lex.(*parser16LexAdapter), $1))
        $$ = a
    }
    ;

SYMBOLx:
    PARSER16_TOK_PRESYMBOL
    {
        $$ = $1.Val
    }
    ;

atype:
    SYMBOLx
    {
        $$ = parser16Acfg(parser16lex).NewSymbol($1, nil)
    }
    | atype PARSER16_TOK_DOT SYMBOLx
    {
        if _, ok := $1.(*This); ok {
            $$ = parser16Acfg(parser16lex).NewSymbol($3, nil)
        } else if sym, ok := $1.(*Symbol); ok {
            $$ = parser16Acfg(parser16lex).NewSymbol(sym.Rep + "." + $3, nil)
        } else {
            $$ = parser16Acfg(parser16lex).NewSymbol($3, nil)
        }
    }
    | PARSER16_TOK_THIS
    {
        $$ = parser16Acfg(parser16lex).NewThis()
    }
    ;

aterm:
    SYMBOLx
    {
        $$ = parser16Acfg(parser16lex).NewAtom($1)
    }
    | aterm PARSER16_TOK_LPAREN terms PARSER16_TOK_RPAREN
    {
        a := $1.(*Atom)
        a.Terms = append(a.Terms, $3...)
        $$ = a
    }
    | aterm PARSER16_TOK_DOT SYMBOLx
    {
        lhs := $1.(*Atom)
        $$ = parser16Acfg(parser16lex).NewAtom(lhs.Rep + "." + $3, lhs.Terms...)
    }
    ;

callatom:
    aterm
    {
        xtracer.Trace("parser.p_callatom_atom ENTER (callatom)")
        $$ = $1
    }
    | PARSER16_TOK_THIS
    {
        xtracer.Trace("parser.p_callatom_this ENTER (callatom)")
        $$ = parser16Acfg(parser16lex).NewAtom("this")
    }
    | callatom PARSER16_TOK_DOT callatom
    {
        xtracer.Trace("parser.p_callatom_callatom_dot_callatom ENTER (callatom)")
        $$ = ComposeAtoms($1.(*Atom), $3.(*Atom))
    }
    ;

callatoms:
    callatom
    {
        xtracer.Trace("parser.p_callatoms_callatom ENTER (callatoms)")
        $$ = []Node{$1}
    }
    | callatoms PARSER16_TOK_COMMA callatom
    {
        xtracer.Trace("parser.p_callatoms_callatoms_callatom ENTER (callatoms)")
        $$ = append($1, $3)
    }
    ;

atom:
    aterm
    {
        xtracer.Trace("parser.p_atom_tatom ENTER (atom)")
        $$ = $1
    }
    ;

var:
    PARSER16_TOK_VARIABLE
    { $$ = parser16Acfg(parser16lex).NewVariable($1.Val, "S") }
    | PARSER16_TOK_VARIABLE PARSER16_TOK_COLON atype
    { $$ = parser16Acfg(parser16lex).NewVariable($1.Val, parser17AtypeToString($3)) }
    ;

simplevar:
    PARSER16_TOK_VARIABLE
    { $$ = parser16Acfg(parser16lex).NewVariable($1.Val, "S") }
    | PARSER16_TOK_VARIABLE PARSER16_TOK_COLON SYMBOLx
    { $$ = parser16Acfg(parser16lex).NewVariable($1.Val, $3) }
    ;

vars:
    var
    { $$ = []Node{$1} }
    | vars PARSER16_TOK_COMMA var
    { $$ = append($1, $3) }
    ;

simplevars:
    simplevar
    { $$ = []Node{$1} }
    | simplevars PARSER16_TOK_COMMA simplevar
    { $$ = append($1, $3) }
    ;

terms:
    /* empty */
    { $$ = nil }
    | term
    { $$ = []Node{$1} }
    | terms PARSER16_TOK_COMMA term
    { $$ = append($1, $3) }
    ;

tterm:
    aterm
    {
        xtracer.Trace("parser.p_tterm_term ENTER (tterm)")
        $$ = $1
    }
    | aterm PARSER16_TOK_COLON atype
    {
        xtracer.Trace("parser.p_tterm_term_colon_symbol ENTER (tterm)")
        if a, ok := $1.(*Atom); ok {
            a.ASort = $3
        }
        $$ = $1
    }
    ;

tterms:
    tterm
    {
        xtracer.Trace("parser.p_tterms_tterm ENTER (tterms)")
        $$ = []Node{$1}
    }
    | tterms PARSER16_TOK_COMMA tterm
    {
        xtracer.Trace("parser.p_tterms_tterms_comma_tterm ENTER (tterms)")
        $$ = append($1, $3)
    }
    ;

term:
    aterm
    { $$ = $1 }
    | var
    { $$ = $1 }
    | PARSER16_TOK_OLD aterm
    { $$ = parser16Acfg(parser16lex).NewOld($2) }
    | PARSER16_TOK_LPAREN term PARSER16_TOK_RPAREN
    { $$ = $2 }
    | term PARSER16_TOK_PLUS term
    { $$ = parser16Acfg(parser16lex).NewAtom("+", $1, $3) }
    | term PARSER16_TOK_MINUS term
    { $$ = parser16Acfg(parser16lex).NewAtom("-", $1, $3) }
    | term PARSER16_TOK_TIMES term
    { $$ = parser16Acfg(parser16lex).NewAtom("*", $1, $3) }
    | term PARSER16_TOK_DIV term
    { $$ = parser16Acfg(parser16lex).NewAtom("/", $1, $3) }
    | term PARSER16_TOK_IF fmla PARSER16_TOK_ELSE term
    { $$ = parser16Acfg(parser16lex).NewIte($3, $1, $5) }
    ;

relop:
    PARSER16_TOK_EQ  { $$ = "=" }
    | PARSER16_TOK_LE  { $$ = "<=" }
    | PARSER16_TOK_LT  { $$ = "<" }
    | PARSER16_TOK_GE  { $$ = ">=" }
    | PARSER16_TOK_GT  { $$ = ">" }
    | PARSER16_TOK_PTO { $$ = "*>" }
    ;

assert_rhs:
    PARSER16_TOK_LCB requires modifies ensures PARSER16_TOK_RCB
    {
        xtracer.Trace("parser.p_assert_rhs_lcb_requires_modifies_ensures_rcb ENTER (assert_rhs)")
        $$ = parser16Acfg(parser16lex).NewRME($2, $3, $4)
    }
    |
    fmla
    {
        xtracer.Trace("parser.p_assert_rhs_fmla ENTER (assert_rhs)")
        $$ = checkNonTemporal($1)
    }
    ;

requires:
    /* empty */
    {
        xtracer.Trace("parser.p_requires ENTER (requires)")
        $$ = parser16Acfg(parser16lex).NewAnd()
    }
    | PARSER16_TOK_REQUIRES fmla
    {
        xtracer.Trace("parser.p_requires_requires_fmla ENTER (requires)")
        $$ = checkNonTemporal($2)
    }
    ;

modifies:
    /* empty */
    {
        xtracer.Trace("parser.p_modifies ENTER (modifies)")
        $$ = nil
    }
    | PARSER16_TOK_MODIFIES PARSER16_TOK_LCB PARSER16_TOK_RCB
    {
        xtracer.Trace("parser.p_modifies_modifies_lcb_rcb ENTER (modifies)")
        $$ = []Node{}
    }
    | PARSER16_TOK_MODIFIES PARSER16_TOK_TIMES
    {
        xtracer.Trace("parser.p_modifies_modofies_times ENTER (modifies)")
        $$ = nil
    }
    | PARSER16_TOK_MODIFIES atoms
    {
        xtracer.Trace("parser.p_modifies_modifies_atoms ENTER (modifies)")
        $$ = $2
    }
    ;

ensures:
    PARSER16_TOK_ENSURES fmla
    {
        xtracer.Trace("parser.p_ensures_ensures_fmla ENTER (ensures)")
        $$ = checkNonTemporal($2)
    }
    ;

atoms:
    atom
    {
        xtracer.Trace("parser.p_atoms_atom ENTER (atoms)")
        $$ = []Node{$1}
    }
    | atoms PARSER16_TOK_COMMA atom
    {
        xtracer.Trace("parser.p_atoms_atoms_atom ENTER (atoms)")
        $$ = append($1, $3)
    }
    ;

fmla:
    term
    { $$ = $1 }
    | term relop term
    { $$ = parser16Acfg(parser16lex).NewAtom($2, $1, $3) }
    | term PARSER16_TOK_TILDAEQ term
    { $$ = parser16Acfg(parser16lex).NewNot(parser16Acfg(parser16lex).NewAtom("=", $1, $3)) }
    | PARSER16_TOK_LPAREN fmla PARSER16_TOK_RPAREN
    { $$ = $2 }
    | PARSER16_TOK_TRUE
    { $$ = parser16Acfg(parser16lex).NewAnd() }
    | PARSER16_TOK_FALSE
    { $$ = parser16Acfg(parser16lex).NewOr() }
    | PARSER16_TOK_TILDA fmla
    { $$ = parser16Acfg(parser16lex).NewNot($2) }
    | fmla PARSER16_TOK_AND fmla
    { $$ = parser16Acfg(parser16lex).NewAnd($1, $3) }
    | fmla PARSER16_TOK_OR fmla
    { $$ = parser16Acfg(parser16lex).NewOr($1, $3) }
    | fmla PARSER16_TOK_ARROW fmla
    { $$ = parser16Acfg(parser16lex).NewImplies($1, $3) }
    | fmla PARSER16_TOK_IFF fmla
    { $$ = parser16Acfg(parser16lex).NewIff($1, $3) }
    | PARSER16_TOK_FORALL simplevars PARSER16_TOK_DOT fmla
    { $$ = parser16Acfg(parser16lex).NewForall($2, $4) }
    | PARSER16_TOK_EXISTS simplevars PARSER16_TOK_DOT fmla
    { $$ = parser16Acfg(parser16lex).NewExists($2, $4) }
    | PARSER16_TOK_GLOBALLY fmla
    { $$ = parser16Acfg(parser16lex).NewGlobally($2) }
    | PARSER16_TOK_EVENTUALLY fmla
    { $$ = parser16Acfg(parser16lex).NewEventually($2) }
    ;

%%
