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
%token <tok>  PARSER16_TOK_COMMA PARSER16_TOK_SEMI PARSER16_TOK_COLON PARSER16_TOK_DOT
%token <tok>  PARSER16_TOK_PLUS PARSER16_TOK_MINUS PARSER16_TOK_TIMES PARSER16_TOK_DIV
%token <tok>  PARSER16_TOK_EQ PARSER16_TOK_TILDAEQ PARSER16_TOK_TILDA PARSER16_TOK_LE PARSER16_TOK_LT PARSER16_TOK_GE PARSER16_TOK_GT
%token <tok>  PARSER16_TOK_AND PARSER16_TOK_OR PARSER16_TOK_ARROW PARSER16_TOK_IFF
%token <tok>  PARSER16_TOK_PTO PARSER16_TOK_DOLLAR PARSER16_TOK_ASSIGN
%token <tok>  PARSER16_TOK_FORALL PARSER16_TOK_EXISTS
%token <tok>  PARSER16_TOK_TRUE PARSER16_TOK_FALSE
%token <tok>  PARSER16_TOK_OLD PARSER16_TOK_THIS
%token <tok>  PARSER16_TOK_IF PARSER16_TOK_ELSE
%token <tok>  PARSER16_TOK_GLOBALLY PARSER16_TOK_EVENTUALLY
%token <tok>  PARSER16_TOK_DOTDOTDOT
%token <tok>  PARSER16_TOK_TYPE PARSER16_TOK_INDIV PARSER16_TOK_VAR PARSER16_TOK_FUNCTION PARSER16_TOK_RELATION PARSER16_TOK_DERIVED
%token <tok>  PARSER16_TOK_AXIOM PARSER16_TOK_PROPERTY PARSER16_TOK_CONJECTURE PARSER16_TOK_SCHEMA PARSER16_TOK_ASSERT PARSER16_TOK_DEFINITION PARSER16_TOK_PROOF PARSER16_TOK_WITH PARSER16_TOK_INIT
%token <tok>  PARSER16_TOK_OBJECT
%token <tok>  PARSER16_TOK_INCLUDE PARSER16_TOK_ACTION PARSER16_TOK_CALL
%token <tok>  PARSER16_TOK_IMPORT PARSER16_TOK_EXPORT PARSER16_TOK_PRIVATE
%token <tok>  PARSER16_TOK_MACRO PARSER16_TOK_ALIAS PARSER16_TOK_PROGRESS PARSER16_TOK_RELY PARSER16_TOK_MIXORD
%token <tok>  PARSER16_TOK_NAMED

%type <accum> top
%type <node>  labeledfmla fmla term aterm var simplevar atype tterm typesymbol symdecl constantdecl rel
%type <node>  fun defn defnrhs optproof proofstep match opttemporal optactiondef sequence simpleact callatom atom optimpex assert_rhs
%type <node>  optskolem schdefn schdefnrhs schconc objsym objectend
%type <nodes> terms vars simplevars tterms rels funs defns matches optargs actseq schdecl schdecls objectargs
%type <bval>  optdotdotdot
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
    | top PARSER16_TOK_OBJECT objsym objectargs PARSER16_TOK_EQ PARSER16_TOK_LCB optdotdotdot top PARSER16_TOK_RCB objectend
    {
        xtracer.Trace("parser.p_top_object_symbol_eq_lcb_top_rcb ENTER (top)")
        $$ = $1
        objAccum := $8
        pref := $3.(*Atom)
        createObject(parser16Acfg(parser16lex), $$, pref, $4, objAccum, nodeLineno($3), $7)
        parser16lex.(*parser16LexAdapter).accum = $$
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
    | top optimpex PARSER16_TOK_ACTION SYMBOLx optargs optactiondef
    {
        xtracer.Trace("parser.p_top_optimpex_action_symbol_optargs_eq_action ENTER (top)")
        $$ = $1
        parser17DeclareAction(parser16Acfg(parser16lex), $$, $2, false, $4, $5, nil, $6, tok16Lineno(parser16lex.(*parser16LexAdapter), $3))
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
    | PARSER16_TOK_GLOBALLY
    {
        xtracer.Trace("parser.p_opttemporal_temporal ENTER (opttemporal)")
        $$ = parser16Acfg(parser16lex).NewAtom("globally")
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
    PARSER16_TOK_LCB actseq PARSER16_TOK_RCB
    {
        xtracer.Trace("parser.p_sequence_lcb_actseq_rcb ENTER (sequence)")
        $$ = parser17MakeSequence(parser16Acfg(parser16lex), $2)
        $$.SetLineno(tok16Lineno(parser16lex.(*parser16LexAdapter), $1))
    }
    ;

actseq:
    /* empty */
    {
        xtracer.Trace("parser.p_actseq ENTER (actseq)")
        $$ = nil
    }
    | simpleact
    {
        xtracer.Trace("parser.p_actseq_action ENTER (actseq)")
        $$ = []Node{$1}
    }
    | actseq PARSER16_TOK_SEMI simpleact
    {
        xtracer.Trace("parser.p_actseq_actseq_semi_action ENTER (actseq)")
        $$ = append($1, $3)
    }
    ;

simpleact:
    PARSER16_TOK_CALL callatom
    {
        xtracer.Trace("parser.p_action_call_callatom ENTER (simpleact)")
        $$ = parser16Acfg(parser16lex).NewCallAction($2)
        $$.SetLineno(tok16Lineno(parser16lex.(*parser16LexAdapter), $1))
    }
    | term PARSER16_TOK_ASSIGN fmla
    {
        xtracer.Trace("parser.p_action_assign ENTER (simpleact)")
        $$ = parser16Acfg(parser16lex).NewAssignAction($1, checkNonTemporal($3))
        $$.SetLineno(tok16Lineno(parser16lex.(*parser16LexAdapter), $2))
    }
    ;

optargs:
    /* empty */
    {
        xtracer.Trace("parser.p_optargs ENTER (optargs)")
        $$ = nil
    }
    | PARSER16_TOK_LPAREN tterms PARSER16_TOK_RPAREN
    {
        xtracer.Trace("parser.p_optargs_params ENTER (optargs)")
        $$ = $2
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
    fmla
    {
        xtracer.Trace("parser.p_assert_rhs_fmla ENTER (assert_rhs)")
        $$ = checkNonTemporal($1)
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
