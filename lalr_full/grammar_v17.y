// grammar_v17.y — goyacc LALR(1) grammar for FULL Ivy file parsing, version 1.7+.
// Mechanically translated from Python PLY grammar in ivy_parser.py and ivy_logic_parser.py.
//
// This is the FULL grammar: it parses complete Ivy files (top-level declarations),
// not just formulas/terms like lalr_logicparser.
//
// Precedence table (copied exactly from Python ivy_parser.py, v1.7+):
//   SEMI < GLOBALLY/EVENTUALLY < ARROW/IFF < OR < AND < TILDA
//   < EQ/LE/LT/GE/GT/PTO < TILDAEQ < IF < ELSE < COLON
//   < PLUS/MINUS < TIMES/DIV < DOLLAR < OLD < DOT

%{
package lalr_full

import (
	"fmt"
	"github.com/glycerine/goivy/ast"
)

// labelCounter is a package-level counter for generating unique label/mixer names.
var lalrLabelCounter int

// newLabel generates a unique label with the given prefix, matching Python newlabel().
func newLabel(pref string) *ast.Atom {
	lalrLabelCounter++
	return ast.NewAtom(fmt.Sprintf("%s%d", pref, lalrLabelCounter))
}

// addLabel adds a label to a LabeledFormula if it doesn't have one.
// Matches Python addlabel() (ivy_parser.py:443-448).
func addLabel(lf *ast.LabeledFormula, pref string) *ast.LabeledFormula {
	if lf.Label != nil {
		return lf
	}
	res := ast.NewLabeledFormula(newLabel(pref), lf.Formula)
	res.Lineno = lf.Lineno
	return res
}

// mkLF wraps a node in a LabeledFormula with no label.
func mkLF(x ast.Node) *ast.LabeledFormula {
	lf := ast.NewLabeledFormula(nil, x)
	return lf
}

%}

// The union type for semantic values.
%union {
	node     ast.Node
	nodes    []ast.Node
	str      string
	bval     bool
	accum    *ivyAccum
}

// Terminal tokens — identifiers and literals
%token <str>  TOK_PRESYMBOL TOK_VARIABLE
%token <str>  TOK_LABEL
%token <str>  TOK_NATIVEQUOTE

// Terminal tokens — punctuation
%token        TOK_LPAREN TOK_RPAREN TOK_LB TOK_RB TOK_LCB TOK_RCB
%token        TOK_COMMA TOK_SEMI TOK_COLON TOK_DOT
%token        TOK_DOTS TOK_DOTDOTDOT

// Terminal tokens — operators
%token        TOK_PLUS TOK_MINUS TOK_TIMES TOK_DIV
%token        TOK_EQ TOK_TILDAEQ TOK_TILDA TOK_LE TOK_LT TOK_GE TOK_GT
%token        TOK_AND TOK_OR TOK_ARROW TOK_IFF
%token        TOK_PTO TOK_DOLLAR TOK_CARET
%token        TOK_ASSIGN

// Terminal tokens — keywords (logic)
%token        TOK_FORALL TOK_EXISTS
%token        TOK_TRUE TOK_FALSE
%token        TOK_OLD TOK_THIS TOK_ISA
%token        TOK_IF TOK_ELSE
%token        TOK_GLOBALLY TOK_EVENTUALLY
%token        TOK_WHENNEXT TOK_WHENPREV TOK_WHENFIRST TOK_WHENLAST

// Terminal tokens — keywords (actions)
%token        TOK_ASSUME TOK_ASSERT TOK_REQUIRE TOK_ENSURE
%token        TOK_VAR TOK_LOCAL TOK_LET TOK_CALL
%token        TOK_WHILE TOK_FOR TOK_IN TOK_INVARIANT TOK_DECREASES
%token        TOK_RETURNS
%token        TOK_SOME TOK_MINIMIZING TOK_MAXIMIZING
%token        TOK_DEBUG TOK_THUNK TOK_UNPROVABLE TOK_PROOF
%token        TOK_INSTANTIATE

// Terminal tokens — keywords (declarations)
%token        TOK_RELATION TOK_INDIV TOK_FUNCTION TOK_DERIVED
%token        TOK_AXIOM TOK_CONJECTURE TOK_SCHEMA TOK_THEOREM
%token        TOK_PROPERTY TOK_DEFINITION
%token        TOK_TYPE TOK_STRUCT
%token        TOK_MODULE TOK_OBJECT TOK_CLASS TOK_SUBCLASS
%token        TOK_ACTION TOK_METHOD
%token        TOK_BEFORE TOK_AFTER TOK_AROUND TOK_MIXIN TOK_IMPLEMENT
%token        TOK_ISOLATE TOK_EXTRACT TOK_TRUSTED
%token        TOK_EXPORT TOK_IMPORT TOK_DELEGATE TOK_USING TOK_INCLUDE
%token        TOK_INTERPRET TOK_MACRO TOK_ALIAS TOK_ATTRIBUTE
%token        TOK_VARIANT TOK_OF
%token        TOK_SCENARIO
%token        TOK_PROGRESS TOK_RELY TOK_MIXORD
%token        TOK_CONCEPT TOK_STATE TOK_UPDATE TOK_FROM
%token        TOK_PARAMS TOK_MODIFIES TOK_ENSURES TOK_REQUIRES
%token        TOK_INIT TOK_ENTRY TOK_SET TOK_NULL TOK_MATCH
%token        TOK_FRESH TOK_NAMED
%token        TOK_TEMPORAL TOK_EXPLICIT
%token        TOK_SPECIFICATION TOK_IMPLEMENTATION TOK_PRIVATE
%token        TOK_GLOBAL TOK_COMMON
%token        TOK_GHOST TOK_FINITE
%token        TOK_PARAMETER
%token        TOK_DESTRUCTOR TOK_CONSTRUCTOR TOK_FIELD
%token        TOK_AUTOINSTANCE
%token        TOK_VAR_KW  // not used yet, placeholder

// Proof/tactic tokens
%token        TOK_TACTIC TOK_TRIGGER
%token        TOK_SHOWGOALS TOK_DEFERGOAL TOK_SPOIL
%token        TOK_UNFOLD TOK_FORGET
%token        TOK_APPLY
%token        TOK_WITH
%token        TOK_METHOD_KW TOK_NULL_KW TOK_SET_KW

// Nonterminal types — formula/term
%type <node>  term fmla appelem var simplevar atype
%type <nodes> terms vars simplevars
%type <str>   SYMBOLx SYMsubscr labelname

// Nonterminal types — top-level
%type <accum> top

// Nonterminal types — declarations
%type <node>  labeledfmla lgprop gprop
%type <node>  opttemporal optunprovable optexplicit optlabel optskolem
%type <node>  optproof
%type <node>  defn defnlhs defnrhs typeddefn gdefn schdefn schdefnrhs schconc
%type <node>  defarg somevarfmla
%type <nodes> defns defargs
%type <nodes> schdecl schdecls
%type <node>  symdecl constantdecl parameter paramval
%type <node>  tapp tterm
%type <nodes> tterms targs tsyms
%type <node>  tatom
%type <nodes> tatoms
%type <node>  rel fun
%type <nodes> rels funs
%type <node>  sort typesymbol
%type <bval>  optfinite optghost
%type <nodes> names
%type <str>   dotsym
%type <node>  optin optelse

// Nonterminal types — actions
%type <node>  action simpleact complexact sequence topseq
%type <nodes> actseq
%type <node>  lparam
%type <nodes> lparams
%type <node>  optactiondef
%type <bval>  actmeth
%type <node>  optimpex
%type <nodes> optargs optreturns optactualreturns
%type <node>  optsemi

// Nonterminal types — callatom
%type <node>  atom callatom
%type <nodes> callatoms

// Nonterminal types — module/object
%type <node>  modcat opteq objsym
%type <bval>  optdotdotdot opttrusted
%type <nodes> objectargs
%type <node>  param
%type <nodes> params
%type <nodes> optwith

// Nonterminal types — instantiate
%type <node>  inst modinst pname
%type <nodes> insts pnames

// Nonterminal types — interpret/misc
%type <node>  oper attributeval
%type <nodes> moresymbols

// Nonterminal types — specimpl
%type <str>   specimpl

// Nonterminal types — relop/infix
%type <str>   relop infix

// Nonterminal types — app
%type <node>  app
%type <nodes> apps

// Nonterminal types — lit
%type <node>  lit

// Nonterminal types — atoms
%type <nodes> atoms

// Nonterminal types — scenario
%type <node>  sceninit scenariomixin scentrans
%type <nodes> scentranss places

// Nonterminal types — proof/tactic
%type <node>  proofstep proofseq proofgroup optproofgroup opttacticwith
%type <node>  tacticwithlistchoice tacticwithelem
%type <nodes> tacticwithlist pflets
%type <node>  pflet

// Nonterminal types — somefmla/while extensions
%type <node>  somefmla
%type <nodes> bounds invariants decreases

// Nonterminal types — match/renaming
%type <node>  match renamingitem renaming optrenaming
%type <nodes> matches renaminglist

// Nonterminal types — update
%type <node>  requires ensures modifies
%type <nodes> upaxes
%type <node>  upax

// Nonterminal types — concept space
%type <node>  expr exprterm cdefn
%type <nodes> cdefns prod sum

// Nonterminal types — state
%type <node>  state_expr assert_rhs

// Nonterminal types — delegate
%type <node>  optdelegee

// Nonterminal types — eqn/let
%type <node>  eqn
%type <nodes> eqns

// Nonterminal types — termtuple
%type <node>  termtuple

// Nonterminal types — debug
%type <node>  debugarg
%type <nodes> debugargs optdebugargs

// Nonterminal types — loc
%type <str>   loc

// Nonterminal types — symbols list
%type <nodes> symbols

// Precedence declarations — copied exactly from Python v1.7+ precedence table.
%left         TOK_SEMI
%left         TOK_GLOBALLY TOK_EVENTUALLY
%left         TOK_ARROW TOK_IFF
%left         TOK_OR
%left         TOK_AND
%left         TOK_TILDA
%left         TOK_EQ TOK_LE TOK_LT TOK_GE TOK_GT TOK_PTO TOK_ISA
%left         TOK_TILDAEQ
%left         TOK_IF
%left         TOK_ELSE
%left         TOK_COLON
%left         TOK_PLUS TOK_MINUS
%left         TOK_TIMES TOK_DIV
%left         TOK_DOLLAR
%left         TOK_OLD
%left         TOK_DOT
%right        TOK_ASSIGN

%start        top

%%

// ============================================================
// top: The Ivy file top-level. Matches Python p_top (ivy_parser.py:361).
// The accumulator (ivyAccum) collects declarations as parsing proceeds.
// ============================================================

top:
    /* empty */
    {
        $$ = newIvyAccum()
        v17lex.(*v17LexAdapter).accum = $$
    }
    | top TOK_USING SYMBOLx
    {
        $$ = $1
        // Python: importer(p[3]) and merge decls — deferred to post-parse
    }
    | top TOK_INCLUDE SYMBOLx
    {
        $$ = $1
        // Python: importer and merge — deferred to post-parse
    }
    // --- Axiom (v1.7+): top optexplicit opttemporal AXIOM lgprop ---
    | top optexplicit opttemporal TOK_AXIOM lgprop
    {
        $$ = $1
        lf := addLabel($5.(*ast.LabeledFormula), "axiom")
        if $3 != nil { // temporal
            t := true
            lf.Temporal = &t
        }
        if $2 != nil { // explicit
            lf.Explicit = true
        }
        d := ast.NewAxiomDecl(lf)
        $$.declare(d)
    }
    // --- Property (v1.7+): top optexplicit opttemporal PROPERTY labeledfmla optskolem optproof ---
    | top optexplicit opttemporal TOK_PROPERTY labeledfmla optskolem optproof
    {
        $$ = $1
        lf := addLabel($5.(*ast.LabeledFormula), "prop")
        if $3 != nil {
            t := true
            lf.Temporal = &t
        }
        if $2 != nil {
            lf.Explicit = true
        }
        d := ast.NewPropertyDecl(lf)
        $$.declare(d)
        if $6 != nil {
            $$.declare(ast.NewNamedDecl($6))
        }
        if $7 != nil {
            $$.declare(ast.NewProofDecl($7))
        }
    }
    // --- Conjecture ---
    | top TOK_CONJECTURE labeledfmla
    {
        $$ = $1
        lf := addLabel($3.(*ast.LabeledFormula), "conj")
        d := ast.NewConjectureDecl(lf)
        $$.declare(d)
    }
    // --- Invariant (v1.7+): top optexplicit INVARIANT labeledfmla optproof ---
    | top optexplicit TOK_INVARIANT labeledfmla optproof
    {
        $$ = $1
        lf := addLabel($4.(*ast.LabeledFormula), "invar")
        lf.Unprovable = false
        if $2 != nil {
            lf.Explicit = true
        }
        d := ast.NewConjectureDecl(lf)
        $$.declare(d)
        if $5 != nil {
            $$.declare(ast.NewProofDecl($5))
        }
    }
    // --- Unprovable Invariant ---
    | top TOK_UNPROVABLE TOK_INVARIANT labeledfmla optproof
    {
        $$ = $1
        lf := addLabel($4.(*ast.LabeledFormula), "invar")
        lf.Unprovable = true
        lf.Explicit = true
        d := ast.NewConjectureDecl(lf)
        // Python: only declare if check_unprovable — we always declare for now
        $$.declare(d)
        if $5 != nil {
            $$.declare(ast.NewProofDecl($5))
        }
    }
    // --- Module ---
    | top TOK_MODULE SYMBOLx modcat atom optwith TOK_EQ TOK_LCB top TOK_RCB
    {
        $$ = $1
        modAccum := $9
        body := ast.NewSequence(modAccum.decls...)
        d := ast.NewDefinition(ast.AppToAtom(ast.NewAtom($3)), body)
        $$.declare(ast.NewModuleDecl(d))
    }
    // --- Object ---
    | top TOK_OBJECT SYMBOLx objectargs TOK_EQ TOK_LCB optdotdotdot top TOK_RCB
    {
        $$ = $1
        objAccum := $8
        pref := ast.NewAtom($3)
        
        objDecl := ast.NewObjectDecl(pref)
        $$.declare(objDecl)
        for _, d := range objAccum.decls {
            $$.declare(d)
        }
    }
    // --- Class ---
    | top TOK_CLASS SYMBOLx objectargs TOK_EQ TOK_LCB optdotdotdot top TOK_RCB
    {
        $$ = $1
        objAccum := $8
        pref := ast.NewAtom($3)
        // Declare type
        scnst := &ast.This{}
        tdfn := &ast.TypeDef{Name: ast.NewAtom("this"), Value: ast.NewUninterpretedSortAST()}
        td := ast.NewTypeDecl(tdfn)
        _ = scnst
        
        objDecl := ast.NewObjectDecl(pref)
        $$.declare(objDecl)
        $$.declare(td)
        for _, d := range objAccum.decls {
            $$.declare(d)
        }
    }
    // --- Subclass ---
    | top TOK_SUBCLASS SYMBOLx TOK_OF atype TOK_EQ TOK_LCB optdotdotdot top TOK_RCB
    {
        $$ = $1
        objAccum := $9
        pref := ast.NewAtom($3)
        
        objDecl := ast.NewObjectDecl(pref)
        $$.declare(objDecl)
        for _, d := range objAccum.decls {
            $$.declare(d)
        }
    }
    // --- Definition (v1.7+): top optexplicit DEFINITION optlabel gdefn optproof ---
    | top optexplicit TOK_DEFINITION optlabel gdefn optproof
    {
        $$ = $1
        lf := ast.NewLabeledFormula($4, $5)
        lf = addLabel(lf, "def")
        dd := ast.NewDefinitionDecl(lf)
        $$.declare(dd)
        if $6 != nil {
            $$.declare(ast.NewProofDecl($6))
        }
    }
    // --- Schema ---
    | top TOK_SCHEMA schdefn
    {
        $$ = $1
        sch := &ast.Schema{Defn: $3}
        sd := ast.NewSchemaDecl(sch)
        $$.declare(sd)
    }
    // --- Theorem with schdefn ---
    | top TOK_THEOREM schdefn optproof
    {
        $$ = $1
        sch := &ast.Schema{Defn: $3}
        td := ast.NewTheoremDecl(sch)
        $$.declare(td)
        if $4 != nil {
            $$.declare(ast.NewProofDecl($4))
        }
    }
    // --- Theorem with LABEL schdefnrhs ---
    | top TOK_THEOREM labelname schdefnrhs optproof
    {
        $$ = $1
        label := ast.NewAtom($3)
        df := ast.NewDefinition(label, $4)
        sch := &ast.Schema{Defn: df}
        td := ast.NewTheoremDecl(sch)
        $$.declare(td)
        if $5 != nil {
            $$.declare(ast.NewProofDecl($5))
        }
    }
    // --- Proof LABEL proofstep ---
    | top TOK_PROOF labelname proofstep
    {
        $$ = $1
        label := ast.NewAtom($3)
        lf := ast.NewLabeledFormula(label, $4)
        $$.declare(ast.NewProofDecl(lf))
    }
    // --- Instantiate ---
    | top TOK_INSTANTIATE insts
    {
        $$ = $1
        d := ast.NewInstantiateDecl($3...)
        $$.declare(d)
    }
    // --- Autoinstance ---
    | top TOK_AUTOINSTANCE insts
    {
        $$ = $1
        d := ast.NewAutoInstanceDecl($3...)
        $$.declare(d)
    }
    // --- symdecl ---
    | top symdecl
    {
        $$ = $1
        $$.declare($2)
    }
    // --- Relation ---
    | top TOK_RELATION rels
    {
        $$ = $1
        for _, d := range $3 {
            $$.declare(d)
        }
    }
    // --- Function ---
    | top TOK_FUNCTION funs
    {
        $$ = $1
        for _, d := range $3 {
            $$.declare(d)
        }
    }
    // --- Derived ---
    | top TOK_DERIVED defns
    {
        $$ = $1
        args := make([]ast.Node, len($3))
        for i, x := range $3 {
            args[i] = addLabel(mkLF(x), "def")
        }
        dd := ast.NewDerivedDecl(args...)
        $$.declare(dd)
    }
    // --- Type (uninterpreted) ---
    | top optfinite optghost TOK_TYPE typesymbol
    {
        $$ = $1
        scnst := ast.NewAtom($5.(*ast.Atom).Rep)
        tdfn := &ast.TypeDef{Name: scnst, Value: ast.NewUninterpretedSortAST()}
        if $2 { tdfn.Finite = true }
        td := ast.NewTypeDecl(tdfn)
        $$.declare(td)
    }
    // --- Type with sort ---
    | top optfinite optghost TOK_TYPE typesymbol TOK_EQ sort
    {
        $$ = $1
        scnst := ast.NewAtom($5.(*ast.Atom).Rep)
        tdfn := &ast.TypeDef{Name: scnst, Value: $7}
        if $2 { tdfn.Finite = true }
        td := ast.NewTypeDecl(tdfn)
        $$.declare(td)
    }
    // --- Progress ---
    | top TOK_PROGRESS defns
    {
        $$ = $1
        pd := ast.NewProgressDecl($3...)
        $$.declare(pd)
    }
    // --- Rely ---
    | top TOK_RELY atom TOK_ARROW atom
    {
        $$ = $1
        imp := &ast.Implies{T1: $3, T2: $5}
        rd := ast.NewRelyDecl(imp)
        $$.declare(rd)
    }
    | top TOK_RELY atom
    {
        $$ = $1
        rd := ast.NewRelyDecl($3)
        $$.declare(rd)
    }
    // --- Mixord ---
    | top TOK_MIXORD callatom TOK_ARROW callatom
    {
        $$ = $1
        imp := &ast.Implies{T1: $3, T2: $5}
        md := ast.NewMixOrdDecl(imp)
        $$.declare(md)
    }
    // --- Concept ---
    | top TOK_CONCEPT cdefns
    {
        $$ = $1
        cd := ast.NewConceptDecl($3...)
        $$.declare(cd)
    }
    // --- Update ---
    | top TOK_UPDATE apps TOK_FROM apps upaxes
    {
        $$ = $1
        // Simplified: store as raw nodes
        _ = $3
        _ = $5
        _ = $6
    }
    // --- Macro ---
    | top TOK_MACRO atom TOK_EQ sequence
    {
        $$ = $1
        d := ast.NewDefinition(ast.AppToAtom($3), $5)
        md := ast.NewMacroDecl(d)
        $$.declare(md)
    }
    // --- Action (v1.7+): top optimpex actmeth SYMBOL optargs optreturns optactiondef ---
    | top optimpex actmeth SYMBOLx optargs optreturns optactiondef
    {
        $$ = $1
        theAtom := ast.NewAtom($4)
        actdef := &ast.ActionDef{
            Name:    theAtom,
            Body:    $7,
            FormalParams: $5,
            FormalReturns: $6,
        }
        decl := ast.NewActionDecl(actdef)
        $$.declare(decl)
        // If export/import was specified
        if $2 != nil {
            $$.declare($2)
        }
    }
    // --- Mixin before ---
    | top TOK_MIXIN callatom TOK_BEFORE callatom
    {
        $$ = $1
        m := &ast.MixinBeforeDef{MixerNode: $3, MixeeNode: $5}
        md := ast.NewMixinDecl(m)
        $$.declare(md)
    }
    // --- Mixin after ---
    | top TOK_MIXIN callatom TOK_AFTER callatom
    {
        $$ = $1
        m := &ast.MixinAfterDef{MixerNode: $3, MixeeNode: $5}
        md := ast.NewMixinDecl(m)
        $$.declare(md)
    }
    // --- Before ---
    | top TOK_BEFORE atype optargs optreturns sequence
    {
        $$ = $1
        atom := ast.NewAtom($3.(*ast.Symbol).Rep)
        lalrLabelCounter++
        mixer := ast.NewAtom(fmt.Sprintf("%s[before%d]", atom.Rep, lalrLabelCounter))
        df := &ast.ActionDef{Name: mixer, Body: $6, FormalParams: $4, FormalReturns: $5}
        decl := ast.NewActionDecl(df)
        $$.declare(decl)
        m := &ast.MixinBeforeDef{MixerNode: mixer, MixeeNode: atom}
        md := ast.NewMixinDecl(m)
        $$.declare(md)
    }
    // --- After ---
    | top TOK_AFTER atype optargs optreturns topseq
    {
        $$ = $1
        atom := ast.NewAtom($3.(*ast.Symbol).Rep)
        lalrLabelCounter++
        mixer := ast.NewAtom(fmt.Sprintf("%s[after%d]", atom.Rep, lalrLabelCounter))
        df := &ast.ActionDef{Name: mixer, Body: $6, FormalParams: $4, FormalReturns: $5}
        decl := ast.NewActionDecl(df)
        $$.declare(decl)
        m := &ast.MixinAfterDef{MixerNode: mixer, MixeeNode: atom}
        md := ast.NewMixinDecl(m)
        $$.declare(md)
    }
    // --- Around ---
    | top TOK_AROUND atype optargs optreturns TOK_LCB actseq optsemi TOK_DOTDOTDOT actseq optsemi TOK_RCB
    {
        $$ = $1
        atom := ast.NewAtom($3.(*ast.Symbol).Rep)
        before := lalrMakeSequence($7)
        after := lalrMakeSequence($10)
        // before mixin
        lalrLabelCounter++
        bmixer := ast.NewAtom(fmt.Sprintf("%s[before%d]", atom.Rep, lalrLabelCounter))
        bdf := &ast.ActionDef{Name: bmixer, Body: before, FormalParams: $4, FormalReturns: $5}
        bdecl := ast.NewActionDecl(bdf)
        $$.declare(bdecl)
        bm := &ast.MixinBeforeDef{MixerNode: bmixer, MixeeNode: atom}
        bmd := ast.NewMixinDecl(bm)
        $$.declare(bmd)
        // after mixin
        lalrLabelCounter++
        amixer := ast.NewAtom(fmt.Sprintf("%s[after%d]", atom.Rep, lalrLabelCounter))
        adf := &ast.ActionDef{Name: amixer, Body: after, FormalParams: $4, FormalReturns: $5}
        adecl := ast.NewActionDecl(adf)
        $$.declare(adecl)
        am := &ast.MixinAfterDef{MixerNode: amixer, MixeeNode: atom}
        amd := ast.NewMixinDecl(am)
        $$.declare(amd)
    }
    // --- After init ---
    | top TOK_AFTER TOK_INIT optargs topseq
    {
        $$ = $1
        atom := ast.NewAtom("init")
        lalrLabelCounter++
        mixer := ast.NewAtom(fmt.Sprintf("init[after%d]", lalrLabelCounter))
        df := &ast.ActionDef{Name: mixer, Body: $5, FormalParams: $4}
        decl := ast.NewActionDecl(df)
        $$.declare(decl)
        m := &ast.MixinAfterDef{MixerNode: mixer, MixeeNode: atom}
        md := ast.NewMixinDecl(m)
        $$.declare(md)
    }
    // --- Implement ---
    | top TOK_IMPLEMENT atype optargs optreturns topseq
    {
        $$ = $1
        atom := ast.NewAtom($3.(*ast.Symbol).Rep)
        lalrLabelCounter++
        mixer := ast.NewAtom(fmt.Sprintf("%s[implement%d]", atom.Rep, lalrLabelCounter))
        df := &ast.ActionDef{Name: mixer, Body: $6, FormalParams: $4, FormalReturns: $5}
        decl := ast.NewActionDecl(df)
        $$.declare(decl)
        m := &ast.MixinImplementDef{MixerNode: mixer, MixeeNode: atom}
        md := ast.NewMixinDecl(m)
        $$.declare(md)
    }
    // --- Implement type ---
    | top TOK_IMPLEMENT TOK_TYPE SYMBOLx TOK_WITH SYMBOLx
    {
        $$ = $1
        a1 := ast.NewAtom($4)
        a2 := ast.NewAtom($6)
        impl := &ast.ImplementTypeDef{Elems: []ast.Node{a1, a2}}
        d := ast.NewImplementTypeDecl(mkLF(impl))
        $$.declare(d)
    }
    // --- Isolate ---
    | top opttrusted TOK_ISOLATE SYMBOLx optargs TOK_EQ callatoms
    {
        $$ = $1
        idef := &ast.IsolateDef{Elems: append([]ast.Node{ast.NewAtom($4)}, $7...)}
        id := ast.NewIsolateDecl(idef)
        $$.declare(id)
    }
    // --- Isolate with WITH ---
    | top opttrusted TOK_ISOLATE SYMBOLx optargs TOK_EQ callatoms TOK_WITH callatoms
    {
        $$ = $1
        idef := &ast.IsolateDef{Elems: append(append([]ast.Node{ast.NewAtom($4)}, $7...), $9...)}
        id := ast.NewIsolateDecl(idef)
        $$.declare(id)
    }
    // --- Isolate with body ---
    | top opttrusted TOK_ISOLATE SYMBOLx optargs TOK_EQ TOK_LCB top TOK_RCB optwith
    {
        $$ = $1
        objAccum := $8
        pref := ast.NewAtom($4)
        
        objDecl := ast.NewObjectDecl(pref)
        $$.declare(objDecl)
        for _, d := range objAccum.decls {
            $$.declare(d)
        }
        args := []ast.Node{ast.NewAtom($4), ast.NewAtom($4)}
        args = append(args, $10...)
        idef := &ast.IsolateDef{Elems: args}
        id := ast.NewIsolateDecl(idef)
        $$.declare(id)
    }
    // --- Extract with body ---
    | top TOK_EXTRACT SYMBOLx objectargs TOK_EQ TOK_LCB top TOK_RCB optwith
    {
        $$ = $1
        objAccum := $7
        pref := ast.NewAtom($3)
        
        objDecl := ast.NewObjectDecl(pref)
        $$.declare(objDecl)
        for _, d := range objAccum.decls {
            $$.declare(d)
        }
    }
    // --- Extract without body ---
    | top TOK_EXTRACT SYMBOLx objectargs TOK_EQ callatoms
    {
        $$ = $1
        // store extract def
    }
    // --- Export ---
    | top TOK_EXPORT callatom
    {
        $$ = $1
        ed := ast.NewExportDecl(&ast.ExportDef{ExportedNode: $3, ScopeNode: ast.NewAtom("")})
        $$.declare(ed)
    }
    // --- Import ---
    | top TOK_IMPORT callatom
    {
        $$ = $1
        id := ast.NewImportDecl(&ast.ImportDef{Imported: $3, Scope: ast.NewAtom("")})
        $$.declare(id)
    }
    // --- Delegate ---
    | top TOK_DELEGATE callatoms optdelegee
    {
        $$ = $1
        args := make([]ast.Node, len($3))
        for i, s := range $3 {
            if $4 != nil {
                args[i] = &ast.DelegateDef{Elems: []ast.Node{s, $4}}
            } else {
                args[i] = &ast.DelegateDef{Elems: []ast.Node{s}}
            }
        }
        dd := ast.NewDelegateDecl(args...)
        $$.declare(dd)
    }
    // --- Interpret ---
    | top TOK_INTERPRET oper TOK_ARROW oper
    {
        $$ = $1
        imp := &ast.Implies{T1: $3, T2: $5}
        lf := addLabel(mkLF(imp), "interp")
        d := ast.NewInterpretDecl(lf)
        $$.declare(d)
    }
    // --- Interpret with range ---
    | top TOK_INTERPRET oper TOK_ARROW TOK_LCB term TOK_DOTS term TOK_RCB
    {
        $$ = $1
        rng := &ast.Range{Lo: $6, Hi: $8}
        imp := &ast.Implies{T1: $3, T2: rng}
        lf := addLabel(mkLF(imp), "interp")
        d := ast.NewInterpretDecl(lf)
        $$.declare(d)
    }
    // --- Interpret with enum ---
    | top TOK_INTERPRET oper TOK_ARROW TOK_LCB SYMBOLx moresymbols TOK_RCB
    {
        $$ = $1
        names := append([]string{$6}, func() []string {
            var r []string
            for _, n := range $7 {
                if a, ok := n.(*ast.Atom); ok {
                    r = append(r, a.Rep)
                }
            }
            return r
        }()...)
        atoms := make([]ast.Node, len(names))
        for i, n := range names {
            atoms[i] = ast.NewAtom(n)
        }
        es := &ast.EnumeratedSort{Elems: atoms}
        imp := &ast.Implies{T1: $3, T2: es}
        lf := addLabel(mkLF(imp), "interp")
        d := ast.NewInterpretDecl(lf)
        $$.declare(d)
    }
    // --- Alias ---
    | top TOK_ALIAS SYMBOLx TOK_EQ callatom
    {
        $$ = $1
        d := ast.NewAliasDecl(ast.NewDefinition(ast.NewAtom($3), $5))
        $$.declare(d)
    }
    // --- Attribute ---
    | top TOK_ATTRIBUTE callatom TOK_EQ attributeval
    {
        $$ = $1
        adef := ast.NewAttributeDef($3, $5)
        d := ast.NewAttributeDecl(adef)
        $$.declare(d)
    }
    // --- Variant ---
    | top TOK_VARIANT typesymbol TOK_OF atype
    {
        $$ = $1
        scnst := ast.NewAtom($3.(*ast.Atom).Rep)
        tdfn := &ast.TypeDef{Name: scnst, Value: ast.NewUninterpretedSortAST()}
        td := ast.NewTypeDecl(tdfn)
        $$.declare(td)
        vdfn := &ast.VariantDef{Name: scnst, VSort: $5}
        vd := ast.NewVariantDecl(vdfn)
        $$.declare(vd)
    }
    // --- Variant with sort ---
    | top TOK_VARIANT typesymbol TOK_OF atype TOK_EQ sort
    {
        $$ = $1
        scnst := ast.NewAtom($3.(*ast.Atom).Rep)
        tdfn := &ast.TypeDef{Name: scnst, Value: $7}
        td := ast.NewTypeDecl(tdfn)
        $$.declare(td)
        vdfn := &ast.VariantDef{Name: scnst, VSort: $5}
        vd := ast.NewVariantDecl(vdfn)
        $$.declare(vd)
    }
    // --- Nativequote ---
    | top TOK_NATIVEQUOTE
    {
        $$ = $1
        // Parse native quote text — simplified for now
    }
    // --- Scenario ---
    | top TOK_SCENARIO TOK_LCB sceninit TOK_SEMI scentranss TOK_RCB
    {
        $$ = $1
        elems := append([]ast.Node{$4}, $6...)
        sdef := &ast.ScenarioDef{Elems: elems}
        sd := ast.NewScenarioDecl(sdef)
        $$.declare(sd)
    }
    // --- Spec/Impl blocks ---
    | top specimpl TOK_LCB top TOK_RCB
    {
        $$ = $1
        innerAccum := $4
        for _, d := range innerAccum.decls {
            $$.declare(d)
        }
    }
    // --- State ---
    | top TOK_STATE SYMBOLx TOK_EQ state_expr
    {
        $$ = $1
        sd := &ast.StateDef{Name: $3, State: $5}
        _ = sd
    }
    // --- Assert (v1.6 only, kept for compat) ---
    | top TOK_ASSERT SYMBOLx TOK_ARROW assert_rhs
    {
        $$ = $1
    }
    ;

// ============================================================
// --- SYMBOL handling ---
// ============================================================

SYMBOLx:
    TOK_PRESYMBOL
    {
        $$ = $1
    }
    ;

SYMsubscr:
    SYMBOLx
    {
        $$ = $1
    }
    | TOK_THIS
    {
        $$ = "this"
    }
    | SYMsubscr TOK_DOT SYMBOLx
    {
        $$ = $1 + "." + $3
    }
    ;

// ============================================================
// --- atype (sort names) ---
// ============================================================

atype:
    SYMBOLx
    {
        $$ = &ast.Symbol{Rep: $1}
    }
    | atype TOK_DOT SYMBOLx
    {
        if _, ok := $1.(*ast.This); ok {
            $$ = &ast.Symbol{Rep: $3}
        } else if sym, ok := $1.(*ast.Symbol); ok {
            $$ = &ast.Symbol{Rep: sym.Rep + "." + $3}
        } else {
            $$ = &ast.Symbol{Rep: $3}
        }
    }
    | TOK_THIS
    {
        $$ = &ast.This{}
    }
    ;

// ============================================================
// --- appelem ---
// ============================================================

appelem:
    SYMBOLx
    {
        $$ = &ast.Atom{Rep: $1}
    }
    | SYMBOLx TOK_LPAREN terms TOK_RPAREN
    {
        $$ = &ast.Atom{Rep: $1, Terms: $3}
    }
    ;

// ============================================================
// --- Variables ---
// ============================================================

var:
    TOK_VARIABLE
    {
        $$ = &ast.Variable{Rep: $1}
    }
    | TOK_VARIABLE TOK_COLON atype
    {
        v := &ast.Variable{Rep: $1}
        v.VSort = $3
        $$ = v
    }
    ;

simplevar:
    TOK_VARIABLE
    {
        $$ = &ast.Variable{Rep: $1}
    }
    | TOK_VARIABLE TOK_COLON SYMBOLx
    {
        v := &ast.Variable{Rep: $1}
        v.VSort = &ast.Symbol{Rep: $3}
        $$ = v
    }
    ;

vars:
    var
    {
        $$ = []ast.Node{$1}
    }
    | vars TOK_COMMA var
    {
        $$ = append($1, $3)
    }
    ;

simplevars:
    simplevar
    {
        $$ = []ast.Node{$1}
    }
    | simplevars TOK_COMMA simplevar
    {
        $$ = append($1, $3)
    }
    ;

// ============================================================
// --- Terms list ---
// ============================================================

terms:
    /* empty */
    {
        $$ = nil
    }
    | term
    {
        $$ = []ast.Node{$1}
    }
    | terms TOK_COMMA term
    {
        $$ = append($1, $3)
    }
    ;

// ============================================================
// --- Terms (v1.7+: unified with formulas) ---
// ============================================================

term:
    appelem
    {
        $$ = $1
    }
    | var
    {
        $$ = $1
    }
    | TOK_OLD appelem
    {
        $$ = &ast.Old{Term: $2}
    }
    | term TOK_DOT appelem
    {
        switch lhs := $1.(type) {
        case *ast.Atom:
            rhs := $3.(*ast.Atom)
            newRep := lhs.Rep + "." + rhs.Rep
            newTerms := make([]ast.Node, 0, len(lhs.Terms)+len(rhs.Terms))
            newTerms = append(newTerms, lhs.Terms...)
            newTerms = append(newTerms, rhs.Terms...)
            $$ = &ast.Atom{Rep: newRep, Terms: newTerms}
        case *ast.Old:
            if inner, ok := lhs.Term.(*ast.Atom); ok {
                rhs := $3.(*ast.Atom)
                newRep := inner.Rep + "." + rhs.Rep
                newTerms := make([]ast.Node, 0, len(inner.Terms)+len(rhs.Terms))
                newTerms = append(newTerms, inner.Terms...)
                newTerms = append(newTerms, rhs.Terms...)
                lhs.Term = &ast.Atom{Rep: newRep, Terms: newTerms}
                $$ = lhs
            } else {
                $$ = &ast.MethodCall{Obj: $1, Method: $3}
            }
        default:
            $$ = &ast.MethodCall{Obj: $1, Method: $3}
        }
    }
    | TOK_LPAREN term TOK_RPAREN
    {
        $$ = $2
    }
    // --- Arithmetic ---
    | term TOK_PLUS term
    {
        $$ = ast.NewApp(ast.NewSymbol("+", nil), $1, $3)
    }
    | term TOK_MINUS term
    {
        $$ = ast.NewApp(ast.NewSymbol("-", nil), $1, $3)
    }
    | term TOK_TIMES term
    {
        $$ = ast.NewApp(ast.NewSymbol("*", nil), $1, $3)
    }
    | term TOK_DIV term
    {
        $$ = ast.NewApp(ast.NewSymbol("/", nil), $1, $3)
    }
    // --- If/else ---
    | term TOK_IF fmla TOK_ELSE term
    {
        $$ = &ast.Ite{Cond: $3, Then: $1, Else: $5}
    }
    // --- Comparison ---
    | term TOK_EQ term
    {
        $$ = &ast.Atom{Rep: "=", Terms: []ast.Node{$1, $3}}
    }
    | term TOK_LE term
    {
        $$ = &ast.Atom{Rep: "<=", Terms: []ast.Node{$1, $3}}
    }
    | term TOK_LT term
    {
        $$ = &ast.Atom{Rep: "<", Terms: []ast.Node{$1, $3}}
    }
    | term TOK_GE term
    {
        $$ = &ast.Atom{Rep: ">=", Terms: []ast.Node{$1, $3}}
    }
    | term TOK_GT term
    {
        $$ = &ast.Atom{Rep: ">", Terms: []ast.Node{$1, $3}}
    }
    | term TOK_PTO term
    {
        $$ = ast.NewApp(ast.NewSymbol("*>", nil), $1, $3)
    }
    | term TOK_TILDAEQ term
    {
        $$ = &ast.Not{Body: &ast.Atom{Rep: "=", Terms: []ast.Node{$1, $3}}}
    }
    // --- Boolean ---
    | TOK_TRUE
    {
        $$ = &ast.And{}
    }
    | TOK_FALSE
    {
        $$ = &ast.Or{}
    }
    | TOK_TILDA term
    {
        $$ = &ast.Not{Body: $2}
    }
    | term TOK_AND term
    {
        $$ = &ast.And{Terms: []ast.Node{$1, $3}}
    }
    | term TOK_OR term
    {
        $$ = &ast.Or{Terms: []ast.Node{$1, $3}}
    }
    | term TOK_ARROW term
    {
        $$ = &ast.Implies{T1: $1, T2: $3}
    }
    | term TOK_IFF term
    {
        $$ = &ast.Iff{T1: $1, T2: $3}
    }
    // --- Quantifiers ---
    | TOK_FORALL simplevars TOK_DOT term    %prec TOK_SEMI
    {
        $$ = &ast.Forall{Bounds: $2, Body: $4}
    }
    | TOK_EXISTS simplevars TOK_DOT term    %prec TOK_SEMI
    {
        $$ = &ast.Exists{Bounds: $2, Body: $4}
    }
    | TOK_FORALL TOK_LPAREN vars TOK_RPAREN term
    {
        $$ = &ast.Forall{Bounds: $3, Body: $5}
    }
    | TOK_EXISTS TOK_LPAREN vars TOK_RPAREN term
    {
        $$ = &ast.Exists{Bounds: $3, Body: $5}
    }
    // --- Temporal ---
    | TOK_GLOBALLY term
    {
        $$ = &ast.Globally{Body: $2}
    }
    | TOK_EVENTUALLY term
    {
        $$ = &ast.Eventually{Body: $2}
    }
    | term TOK_WHENNEXT term
    {
        $$ = &ast.WhenOperator{Name: "next", T1: $1, T2: $3}
    }
    | term TOK_WHENPREV term
    {
        $$ = &ast.WhenOperator{Name: "prev", T1: $1, T2: $3}
    }
    | term TOK_WHENFIRST term
    {
        $$ = &ast.WhenOperator{Name: "first", T1: $1, T2: $3}
    }
    | term TOK_WHENLAST term
    {
        $$ = &ast.WhenOperator{Name: "last", T1: $1, T2: $3}
    }
    // --- ISA ---
    | term TOK_ISA atype
    {
        $$ = &ast.Isa{Terms: []ast.Node{$1, $3}}
    }
    // --- Sort annotation ---
    | term TOK_COLON atype
    {
        if v, ok := $1.(*ast.Variable); ok {
            v.VSort = $3
        }
        $$ = $1
    }
    // --- Named binders ---
    | TOK_LPAREN TOK_DOLLAR SYMBOLx simplevars TOK_DOT fmla TOK_RPAREN TOK_LPAREN terms TOK_RPAREN
    {
        binder := &ast.NamedBinder{Name: $3, Bounds: $4, Body: $6}
        $$ = &ast.Atom{Rep: "", Terms: append([]ast.Node{binder}, $9...)}
    }
    | TOK_DOLLAR SYMBOLx TOK_DOT fmla     %prec TOK_SEMI
    {
        $$ = &ast.NamedBinder{Name: $2, Body: $4}
    }
    | TOK_DOLLAR SYMBOLx TOK_DOLLAR fmla   %prec TOK_SEMI
    {
        $$ = &ast.NamedBinder{Name: $2, Body: $4}
    }
    ;

// ============================================================
// --- fmla ---
// ============================================================

fmla:
    term
    {
        $$ = $1
    }
    ;

// ============================================================
// --- labeledfmla ---
// ============================================================

labeledfmla:
    fmla
    {
        $$ = ast.NewLabeledFormula(nil, $1)
    }
    | labelname fmla
    {
        $$ = ast.NewLabeledFormula(ast.NewAtom($1), $2)
    }
    ;

// labelname matches Python's LABEL : LB SYMBOL RB (ivy_logic_parser.py:23-25).
// The Python lexer produces LABEL as a terminal, but our lexer produces
// separate LB, SYMBOL, RB tokens, so we combine them in the grammar.
labelname:
    TOK_LB SYMBOLx TOK_RB
    {
        $$ = "[" + $2 + "]"
    }
    | TOK_LABEL
    {
        $$ = $1
    }
    ;

// ============================================================
// --- lgprop / gprop (v1.7+) ---
// ============================================================

gprop:
    fmla
    {
        $$ = $1
    }
    | schdefnrhs
    {
        $$ = $1
    }
    ;

lgprop:
    optlabel gprop
    {
        lf := ast.NewLabeledFormula($1, $2)
        $$ = lf
    }
    ;

// ============================================================
// --- Optional markers ---
// ============================================================

opttemporal:
    /* empty */
    {
        $$ = nil
    }
    | TOK_TEMPORAL
    {
        $$ = &ast.And{} // non-nil marker
    }
    ;

optunprovable:
    /* empty */
    {
        $$ = nil
    }
    | TOK_UNPROVABLE
    {
        $$ = &ast.And{} // non-nil marker
    }
    ;

optexplicit:
    /* empty */
    {
        $$ = nil
    }
    | TOK_EXPLICIT
    {
        $$ = &ast.And{} // non-nil marker
    }
    ;

optlabel:
    /* empty */
    {
        $$ = nil
    }
    | labelname
    {
        $$ = ast.NewAtom($1)
    }
    ;

optskolem:
    /* empty */
    {
        $$ = nil
    }
    | TOK_NAMED defnlhs
    {
        $$ = $2
    }
    ;

optproof:
    /* empty */
    {
        $$ = nil
    }
    | TOK_PROOF proofstep
    {
        $$ = $2
    }
    | TOK_PROOF labelname proofstep
    {
        label := ast.NewAtom($2)
        $$ = ast.NewLabeledFormula(label, $3)
    }
    ;

// optproofgroup2 removed — use optproofgroup instead

optsemi:
    /* empty */
    {
        $$ = nil
    }
    | TOK_SEMI
    {
        $$ = nil
    }
    ;

// ============================================================
// --- Definitions ---
// ============================================================

dotsym:
    SYMBOLx
    {
        $$ = $1
    }
    | dotsym TOK_DOT SYMBOLx
    {
        $$ = $1 + "." + $3
    }
    ;

defnlhs:
    dotsym
    {
        $$ = ast.NewAtom($1)
    }
    | dotsym TOK_LPAREN defargs TOK_RPAREN
    {
        a := ast.NewAtom($1)
        a.Terms = $3
        $$ = a
    }
    | TOK_LPAREN defarg relop defarg TOK_RPAREN
    {
        $$ = ast.NewAtom($3, $2, $4)
    }
    | TOK_LPAREN defarg infix defarg TOK_RPAREN
    {
        $$ = ast.NewApp(ast.NewSymbol($3, nil), $2, $4)
    }
    ;

defarg:
    lparam
    {
        $$ = $1
    }
    | var
    {
        $$ = $1
    }
    ;

defargs:
    defarg
    {
        $$ = []ast.Node{$1}
    }
    | defargs TOK_COMMA defarg
    {
        $$ = append($1, $3)
    }
    ;

typeddefn:
    defnlhs
    {
        $$ = $1
    }
    | defnlhs TOK_COLON atype
    {
        // set sort on the defnlhs
        if a, ok := $1.(*ast.Atom); ok {
            a.ASort = $3
        } else if app, ok := $1.(*ast.App); ok {
            app.ASort = $3
        }
        $$ = $1
    }
    ;

defnrhs:
    fmla
    {
        $$ = $1
    }
    | somevarfmla
    {
        $$ = $1
    }
    | TOK_NATIVEQUOTE
    {
        $$ = &ast.NativeExpr{}
    }
    ;

defn:
    typeddefn TOK_EQ defnrhs
    {
        $$ = ast.NewDefinition(ast.AppToAtom($1), $3)
    }
    ;

defns:
    defn
    {
        $$ = []ast.Node{$1}
    }
    | defns TOK_COMMA defn
    {
        $$ = append($1, $3)
    }
    ;

gdefn:
    defn
    {
        $$ = $1
    }
    | TOK_LCB defn TOK_RCB
    {
        d := $2.(*ast.Definition)
        $$ = &ast.DefinitionSchema{Definition: *d}
    }
    ;

somevarfmla:
    TOK_SOME simplevar TOK_DOT fmla optin optelse
    {
        se := &ast.SomeExpr{Param: $2, Fmla: $4}
        if $5 != nil { se.IfValue = $5 }
        if $6 != nil { se.ElseVal = $6 }
        $$ = se
    }
    ;

optin:
    /* empty */
    {
        $$ = nil
    }
    | TOK_IN fmla
    {
        $$ = $2
    }
    ;

optelse:
    /* empty */
    {
        $$ = nil
    }
    | TOK_ELSE fmla
    {
        $$ = $2
    }
    ;

// ============================================================
// --- Schema ---
// ============================================================

schdefnrhs:
    fmla
    {
        $$ = $1
    }
    | TOK_LCB schdecls schconc TOK_RCB
    {
        args := append($2, $3)
        $$ = ast.NewSchemaBody(args...)
    }
    ;

schdecl:
    TOK_FUNCTION funs
    {
        $$ = make([]ast.Node, len($2))
        copy($$, $2)
    }
    | TOK_FRESH TOK_FUNCTION funs
    {
        $$ = make([]ast.Node, len($3))
        copy($$, $3)
    }
    | TOK_INDIV funs
    {
        $$ = make([]ast.Node, len($2))
        copy($$, $2)
    }
    | TOK_FRESH TOK_INDIV funs
    {
        $$ = make([]ast.Node, len($3))
        copy($$, $3)
    }
    | TOK_RELATION rels
    {
        $$ = make([]ast.Node, len($2))
        copy($$, $2)
    }
    | TOK_FRESH TOK_RELATION rels
    {
        $$ = make([]ast.Node, len($3))
        copy($$, $3)
    }
    | TOK_TYPE SYMBOLx
    {
        scnst := ast.NewAtom($2)
        tdfn := &ast.TypeDef{Name: scnst, Value: ast.NewUninterpretedSortAST()}
        $$ = []ast.Node{tdfn}
    }
    | optexplicit TOK_PROPERTY lgprop
    {
        lf := addLabel($3.(*ast.LabeledFormula), "prop")
        if $1 != nil {
            lf.Explicit = true
        }
        $$ = []ast.Node{lf}
    }
    | TOK_THEOREM lgprop
    {
        lf := addLabel($2.(*ast.LabeledFormula), "prop")
        $$ = []ast.Node{lf}
    }
    | schdefnrhs
    {
        lf := ast.NewLabeledFormula(nil, $1)
        lf = addLabel(lf, "sch")
        $$ = []ast.Node{lf}
    }
    ;

schdecls:
    /* empty */
    {
        $$ = nil
    }
    | schdecls schdecl
    {
        $$ = append($1, $2...)
    }
    ;

schconc:
    TOK_DEFINITION defn
    {
        $$ = $2
    }
    | optexplicit TOK_PROPERTY lgprop
    {
        lf := $3.(*ast.LabeledFormula)
        $$ = lf.Formula
    }
    ;

schdefn:
    defnlhs TOK_EQ schdefnrhs
    {
        $$ = ast.NewDefinition(ast.AppToAtom($1), $3)
    }
    ;

// ============================================================
// --- Symbol/Type/Relation/Function Declarations ---
// ============================================================

symdecl:
    constantdecl
    {
        $$ = $1
    }
    | TOK_DESTRUCTOR tterms
    {
        d := ast.NewDestructorDecl($2...)
        $$ = d
    }
    | TOK_FIELD tterms
    {
        d := ast.NewDestructorDecl($2...)
        $$ = d
    }
    | TOK_CONSTRUCTOR tterms
    {
        d := ast.NewConstructorDecl($2...)
        $$ = d
    }
    ;

constantdecl:
    TOK_INDIV tterms
    {
        d := ast.NewConstantDecl($2...)
        $$ = d
    }
    | TOK_VAR tterms
    {
        d := ast.NewConstantDecl($2...)
        $$ = d
    }
    | TOK_PARAMETER parameter
    {
        $$ = $2
    }
    ;

parameter:
    tterm
    {
        d := ast.NewParameterDecl($1)
        $$ = d
    }
    | tterm TOK_EQ paramval
    {
        df := ast.NewDefinition($1, $3)
        d := ast.NewParameterDecl(df)
        $$ = d
    }
    ;

paramval:
    TOK_TRUE
    {
        $$ = ast.NewAtom("true")
    }
    | TOK_FALSE
    {
        $$ = ast.NewAtom("false")
    }
    | SYMBOLx
    {
        $$ = ast.NewApp(ast.NewSymbol($1, nil))
    }
    ;

tapp:
    SYMBOLx
    {
        $$ = ast.NewApp(ast.NewSymbol($1, nil))
    }
    | SYMBOLx targs
    {
        args := make([]ast.Node, len($2))
        copy(args, $2)
        $$ = ast.NewApp(ast.NewSymbol($1, nil), args...)
    }
    | TOK_LPAREN var infix var TOK_RPAREN
    {
        $$ = ast.NewApp(ast.NewSymbol($3, nil), $2, $4)
    }
    ;

tterm:
    tapp
    {
        $$ = $1
    }
    | tapp TOK_COLON atype
    {
        if app, ok := $1.(*ast.App); ok {
            app.ASort = $3
        }
        $$ = $1
    }
    | SYMBOLx TOK_COLON atype
    {
        a := &ast.Atom{Rep: $1}
        a.ASort = $3
        $$ = a
    }
    | TOK_CARET SYMBOLx TOK_COLON atype
    {
        a := &ast.Atom{Rep: $2}
        a.ASort = $4
        $$ = a
    }
    ;

tterms:
    tterm
    {
        $$ = []ast.Node{$1}
    }
    | tterms TOK_COMMA tterm
    {
        $$ = append($1, $3)
    }
    ;

targs:
    TOK_LPAREN TOK_RPAREN
    {
        $$ = nil
    }
    | TOK_LPAREN tsyms TOK_RPAREN
    {
        $$ = $2
    }
    ;

tsyms:
    var
    {
        $$ = []ast.Node{$1}
    }
    | tsyms TOK_COMMA var
    {
        $$ = append($1, $3)
    }
    ;

tatom:
    SYMBOLx
    {
        $$ = ast.NewAtom($1)
    }
    | SYMBOLx targs
    {
        $$ = &ast.Atom{Rep: $1, Terms: $2}
    }
    | TOK_LPAREN var relop var TOK_RPAREN
    {
        $$ = ast.NewAtom($3, $2, $4)
    }
    ;

tatoms:
    tatom
    {
        $$ = []ast.Node{$1}
    }
    | tatoms TOK_COMMA tatom
    {
        $$ = append($1, $3)
    }
    ;

rel:
    defnlhs
    {
        // relation declaration (sort = bool)
        d := ast.NewConstantDecl($1)
        $$ = d
    }
    | defn
    {
        lf := addLabel(mkLF($1), "def")
        d := ast.NewDerivedDecl(lf)
        $$ = d
    }
    ;

rels:
    rel
    {
        $$ = []ast.Node{$1}
    }
    | rels TOK_COMMA rel
    {
        $$ = append($1, $3)
    }
    ;

fun:
    typeddefn
    {
        d := ast.NewConstantDecl($1)
        $$ = d
    }
    | typeddefn TOK_EQ defnrhs
    {
        df := ast.NewDefinition(ast.AppToAtom($1), $3)
        lf := addLabel(mkLF(df), "def")
        d := ast.NewDerivedDecl(lf)
        $$ = d
    }
    ;

funs:
    fun
    {
        $$ = []ast.Node{$1}
    }
    | funs TOK_COMMA fun
    {
        $$ = append($1, $3)
    }
    ;

// --- Type-related ---

typesymbol:
    SYMBOLx
    {
        $$ = ast.NewAtom($1)
    }
    | TOK_THIS
    {
        $$ = ast.NewAtom("this")
    }
    ;

optfinite:
    /* empty */
    {
        $$ = false
    }
    | TOK_FINITE
    {
        $$ = true
    }
    ;

optghost:
    /* empty */
    {
        $$ = false
    }
    | TOK_GHOST
    {
        $$ = true
    }
    ;

sort:
    TOK_LCB SYMBOLx TOK_RCB
    {
        $$ = &ast.EnumeratedSort{Elems: []ast.Node{ast.NewAtom($2)}}
    }
    | TOK_LCB SYMBOLx TOK_COMMA names TOK_RCB
    {
        vals := []ast.Node{ast.NewAtom($2)}
        for _, n := range $4 {
            vals = append(vals, n)
        }
        $$ = &ast.EnumeratedSort{Elems: vals}
    }
    | TOK_LCB SYMBOLx TOK_DOTS SYMBOLx TOK_RCB
    {
        $$ = &ast.Range{Lo: ast.NewAtom($2), Hi: ast.NewAtom($4)}
    }
    | TOK_STRUCT TOK_LCB tterms TOK_RCB
    {
        $$ = &ast.StructSort{Fields: $3}
    }
    | TOK_STRUCT TOK_LCB TOK_RCB
    {
        $$ = &ast.StructSort{}
    }
    ;

names:
    SYMBOLx
    {
        $$ = []ast.Node{ast.NewAtom($1)}
    }
    | names TOK_COMMA SYMBOLx
    {
        $$ = append($1, ast.NewAtom($3))
    }
    ;

// ============================================================
// --- relop / infix ---
// ============================================================

relop:
    TOK_EQ     { $$ = "=" }
    | TOK_LE   { $$ = "<=" }
    | TOK_LT   { $$ = "<" }
    | TOK_GE   { $$ = ">=" }
    | TOK_GT   { $$ = ">" }
    | TOK_PTO  { $$ = "*>" }
    ;

infix:
    TOK_PLUS   { $$ = "+" }
    | TOK_MINUS { $$ = "-" }
    | TOK_TIMES { $$ = "*" }
    | TOK_DIV   { $$ = "/" }
    ;

// ============================================================
// --- atom / atoms / app / apps / lit ---
// ============================================================

atom:
    SYMBOLx
    {
        $$ = ast.NewAtom($1)
    }
    | SYMBOLx TOK_LPAREN terms TOK_RPAREN
    {
        $$ = &ast.Atom{Rep: $1, Terms: $3}
    }
    ;

atoms:
    atom
    {
        $$ = []ast.Node{$1}
    }
    | atoms TOK_COMMA atom
    {
        $$ = append($1, $3)
    }
    ;

app:
    SYMBOLx
    {
        $$ = ast.NewApp(ast.NewSymbol($1, nil))
    }
    | SYMBOLx TOK_LPAREN terms TOK_RPAREN
    {
        $$ = ast.NewApp(ast.NewSymbol($1, nil), $3...)
    }
    | term infix term
    {
        $$ = ast.NewApp(ast.NewSymbol($2, nil), $1, $3)
    }
    ;

apps:
    app
    {
        $$ = []ast.Node{$1}
    }
    | apps TOK_COMMA app
    {
        $$ = append($1, $3)
    }
    ;

lit:
    atom
    {
        $$ = $1
    }
    | TOK_TILDA lit
    {
        $$ = &ast.Not{Body: $2}
    }
    ;

// ============================================================
// --- callatom / callatoms ---
// ============================================================

callatom:
    atom
    {
        $$ = $1
    }
    | TOK_THIS
    {
        $$ = ast.NewAtom("this")
    }
    | TOK_METHOD
    {
        $$ = ast.NewAtom("method")
    }
    | callatom TOK_DOT callatom
    {
        lhs := $1.(*ast.Atom)
        rhs := $3.(*ast.Atom)
        $$ = ast.ComposeAtoms(lhs, rhs)
    }
    ;

callatoms:
    callatom
    {
        $$ = []ast.Node{$1}
    }
    | callatoms TOK_COMMA callatom
    {
        $$ = append($1, $3)
    }
    ;

// ============================================================
// --- Module/Object helpers ---
// ============================================================

modcat:
    /* empty */
    {
        $$ = nil
    }
    | TOK_OBJECT
    {
        $$ = ast.NewAtom("object")
    }
    | TOK_ISOLATE
    {
        $$ = ast.NewAtom("isolate")
    }
    ;

opteq:
    /* empty */
    {
        $$ = nil
    }
    | TOK_EQ
    {
        $$ = nil
    }
    ;

optdotdotdot:
    /* empty */
    {
        $$ = false
    }
    | TOK_DOTDOTDOT
    {
        $$ = true
    }
    ;

objectargs:
    optargs
    {
        $$ = $1
    }
    ;

objsym:
    SYMBOLx
    {
        $$ = ast.NewAtom($1)
    }
    ;

opttrusted:
    /* empty */
    {
        $$ = false
    }
    | TOK_TRUSTED
    {
        $$ = true
    }
    ;

optargs:
    /* empty */
    {
        $$ = nil
    }
    | TOK_LPAREN lparams TOK_RPAREN
    {
        $$ = $2
    }
    ;

optreturns:
    /* empty */
    {
        $$ = nil
    }
    | TOK_RETURNS TOK_LPAREN lparams TOK_RPAREN
    {
        $$ = $3
    }
    ;

optactualreturns:
    /* empty */
    {
        $$ = nil
    }
    | callatoms TOK_ASSIGN
    {
        $$ = $1
    }
    ;

param:
    SYMBOLx TOK_COLON SYMBOLx
    {
        a := ast.NewApp(ast.NewSymbol($1, nil))
        a.ASort = &ast.Symbol{Rep: $3}
        $$ = a
    }
    ;

params:
    param
    {
        $$ = []ast.Node{$1}
    }
    | params TOK_COMMA param
    {
        $$ = append($1, $3)
    }
    ;

optwith:
    /* empty */
    {
        $$ = nil
    }
    | TOK_WITH callatoms
    {
        $$ = $2
    }
    ;

// ============================================================
// --- lparam / lparams ---
// ============================================================

lparam:
    SYMBOLx TOK_COLON atype
    {
        a := &ast.Atom{Rep: $1}
        a.ASort = $3
        $$ = a
    }
    | TOK_CARET SYMBOLx TOK_COLON atype
    {
        a := &ast.Atom{Rep: $2}
        a.ASort = $4
        $$ = a
    }
    ;

lparams:
    lparam
    {
        $$ = []ast.Node{$1}
    }
    | lparams TOK_COMMA lparam
    {
        $$ = append($1, $3)
    }
    ;

// ============================================================
// --- Action top-level helpers ---
// ============================================================

optactiondef:
    /* empty */
    {
        $$ = &ast.Sequence{}
    }
    | TOK_EQ topseq
    {
        $$ = $2
    }
    | TOK_EQ TOK_TIMES
    {
        $$ = &ast.CrashAction{}
    }
    ;

topseq:
    sequence
    {
        $$ = $1
    }
    | TOK_LCB TOK_NATIVEQUOTE TOK_RCB
    {
        // TODO: add ast.NativeAction (Python ivy_actions.py:1155)
        $$ = ast.NewAtom("native")
    }
    ;

optimpex:
    /* empty */
    {
        $$ = nil
    }
    | TOK_EXPORT
    {
        $$ = nil // marker handled in top rule
    }
    | TOK_IMPORT
    {
        $$ = nil // marker handled in top rule
    }
    ;

actmeth:
    TOK_ACTION
    {
        $$ = false
    }
    | TOK_METHOD
    {
        $$ = true
    }
    ;

// ============================================================
// --- specimpl ---
// ============================================================

specimpl:
    TOK_SPECIFICATION
    {
        $$ = "spec"
    }
    | TOK_IMPLEMENTATION
    {
        $$ = "impl"
    }
    | TOK_PRIVATE
    {
        $$ = "private"
    }
    | TOK_GLOBAL
    {
        $$ = "global"
    }
    | TOK_COMMON
    {
        $$ = "common"
    }
    ;

// ============================================================
// --- Instantiate ---
// ============================================================

insts:
    inst
    {
        $$ = []ast.Node{$1}
    }
    | insts TOK_COMMA inst
    {
        $$ = append($1, $3)
    }
    ;

inst:
    modinst
    {
        $$ = &ast.Instantiation{Name: nil, Sort: ast.AppToAtom($1)}
    }
    | modinst TOK_COLON modinst
    {
        $$ = &ast.Instantiation{Name: ast.AppToAtom($1), Sort: ast.AppToAtom($3)}
    }
    ;

modinst:
    dotsym
    {
        $$ = ast.NewAtom($1)
    }
    | dotsym TOK_LPAREN pnames TOK_RPAREN
    {
        a := ast.NewAtom($1)
        a.Terms = $3
        $$ = a
    }
    ;

pname:
    atype
    {
        $$ = ast.NewApp(ast.NewSymbol($1.(*ast.Symbol).Rep, nil))
    }
    | var
    {
        $$ = $1
    }
    | infix
    {
        $$ = ast.NewApp(ast.NewSymbol($1, nil))
    }
    | relop
    {
        $$ = ast.NewApp(ast.NewSymbol($1, nil))
    }
    | TOK_THIS
    {
        $$ = ast.NewApp(ast.NewSymbol("this", nil))
    }
    | TOK_TRUE
    {
        $$ = ast.NewAtom("true")
    }
    | TOK_FALSE
    {
        $$ = ast.NewAtom("false")
    }
    ;

pnames:
    /* empty */
    {
        $$ = nil
    }
    | pname
    {
        $$ = []ast.Node{$1}
    }
    | pnames TOK_COMMA pname
    {
        $$ = append($1, $3)
    }
    ;

// ============================================================
// --- Interpret/Alias/Attribute helpers ---
// ============================================================

oper:
    atype
    {
        $$ = $1
    }
    | relop
    {
        $$ = ast.NewAtom($1)
    }
    | infix
    {
        $$ = ast.NewAtom($1)
    }
    | TOK_NATIVEQUOTE
    {
        $$ = &ast.NativeType{}
    }
    ;

attributeval:
    callatom
    {
        $$ = $1
    }
    | TOK_TRUE
    {
        $$ = ast.NewAtom("true")
    }
    | TOK_FALSE
    {
        $$ = ast.NewAtom("false")
    }
    ;

moresymbols:
    /* empty */
    {
        $$ = nil
    }
    | moresymbols TOK_COMMA SYMBOLx
    {
        $$ = append($1, ast.NewAtom($3))
    }
    ;

// ============================================================
// --- Delegate ---
// ============================================================

optdelegee:
    /* empty */
    {
        $$ = nil
    }
    | TOK_ARROW callatom
    {
        $$ = $2
    }
    ;

// ============================================================
// --- sequence / actseq / action ---
// ============================================================

sequence:
    TOK_LCB TOK_RCB
    {
        $$ = &ast.Sequence{}
    }
    | TOK_LCB actseq TOK_RCB
    {
        $$ = lalrMakeSequence($2)
    }
    | TOK_LCB actseq TOK_SEMI TOK_RCB
    {
        $$ = lalrMakeSequence($2)
    }
    ;

actseq:
    action
    {
        $$ = []ast.Node{$1}
    }
    | actseq TOK_SEMI action
    {
        $$ = append($1, $3)
    }
    | actseq complexact
    {
        $$ = append($1, $2)
    }
    ;

action:
    simpleact
    {
        $$ = $1
    }
    | complexact
    {
        $$ = $1
    }
    ;

// ============================================================
// --- Simple actions ---
// ============================================================

simpleact:
    TOK_ASSUME labeledfmla
    {
        $$ = ast.NewAtom("assume", $2)
    }
    | optunprovable TOK_ASSERT labeledfmla
    {
        $$ = ast.NewAtom("assert", $3)
    }
    | optunprovable TOK_ASSERT labeledfmla TOK_PROOF proofstep
    {
        $$ = ast.NewAtom("assert", $3, $5)
    }
    | optunprovable TOK_REQUIRE labeledfmla
    {
        $$ = ast.NewAtom("require", $3)
    }
    | optunprovable TOK_REQUIRE labeledfmla TOK_PROOF proofstep
    {
        $$ = ast.NewAtom("require", $3, $5)
    }
    | optunprovable TOK_ENSURE labeledfmla
    {
        $$ = ast.NewAtom("ensure", $3)
    }
    | optunprovable TOK_ENSURE labeledfmla TOK_PROOF proofstep
    {
        $$ = ast.NewAtom("ensure", $3, $5)
    }
    | term TOK_ASSIGN fmla
    {
        $$ = ast.NewAtom(":=", $1, $3)
    }
    | termtuple TOK_ASSIGN callatom
    {
        $$ = ast.NewAtom("call", $3, $1)
    }
    | term TOK_ASSIGN TOK_TIMES
    {
        $$ = ast.NewAtom("havoc", $1)
    }
    | TOK_VAR tterm
    {
        $$ = ast.NewAtom("var", $2)
    }
    | TOK_VAR tterm TOK_ASSIGN fmla
    {
        $$ = ast.NewAtom("var", $2, $4)
    }
    | TOK_CALL optactualreturns callatom
    {
        args := append([]ast.Node{$3}, $2...)
        $$ = ast.NewAtom("call", args...)
    }
    | TOK_CALL callatom
    {
        $$ = ast.NewAtom("call", $2)
    }
    | TOK_SET lit
    {
        $$ = ast.NewAtom("set", $2)
    }
    | TOK_INSTANTIATE callatom
    {
        $$ = ast.NewAtom("instantiate", $2)
    }
    | TOK_UNPROVABLE simpleact
    {
        $$ = &ast.Sequence{} // no-op
    }
    | TOK_DEBUG SYMBOLx optdebugargs
    {
        args := append([]ast.Node{ast.NewAtom($2)}, $3...)
        $$ = ast.NewAtom("debug", args...)
    }
    | term     %prec TOK_SEMI
    {
        $$ = $1
    }
    ;

termtuple:
    TOK_LPAREN term TOK_COMMA terms TOK_RPAREN
    {
        args := append([]ast.Node{$2}, $4...)
        $$ = &ast.Tuple{Elems: args}
    }
    ;

debugarg:
    SYMBOLx TOK_EQ fmla
    {
        $$ = ast.NewDefinition(ast.NewApp(ast.NewSymbol($1, nil)), $3)
    }
    ;

debugargs:
    debugarg
    {
        $$ = []ast.Node{$1}
    }
    | debugargs TOK_COMMA debugarg
    {
        $$ = append($1, $3)
    }
    ;

optdebugargs:
    /* empty */
    {
        $$ = nil
    }
    | TOK_WITH debugargs
    {
        $$ = $2
    }
    ;

// ============================================================
// --- Complex actions ---
// ============================================================

complexact:
    sequence
    {
        $$ = $1
    }
    | TOK_IF somefmla sequence
    {
        $$ = ast.NewIte($2, $3, &ast.Sequence{})
    }
    | TOK_IF somefmla sequence TOK_ELSE action
    {
        $$ = ast.NewIte($2, $3, $5)
    }
    | TOK_IF TOK_TIMES sequence TOK_ELSE action
    {
        $$ = ast.NewIte(ast.NewSymbol("*", nil), $3, $5)
    }
    | TOK_WHILE somefmla invariants decreases sequence
    {
        args := []ast.Node{$2, $5}
        args = append(args, $3...)
        args = append(args, $4...)
        $$ = ast.NewAtom("while", args...)
    }
    | TOK_FOR tterm TOK_COMMA tterm TOK_IN fmla invariants decreases sequence
    {
        $$ = ast.NewAtom("for", $2, $4, $6, $9)
    }
    | TOK_LOCAL lparams sequence
    {
        args := append($2, $3)
        $$ = ast.NewAtom("local", args...)
    }
    | TOK_LET eqns sequence
    {
        args := append($2, $3)
        $$ = ast.NewAtom("let", args...)
    }
    | TOK_THUNK labelname SYMBOLx optargs TOK_COLON atype TOK_ASSIGN sequence
    {
        $$ = ast.NewAtom("thunk", ast.NewAtom($2), ast.NewAtom($3), $8)
    }
    ;

// --- somefmla ---

somefmla:
    fmla
    {
        $$ = $1
    }
    | fmla TOK_ASSIGN fmla
    {
        $$ = ast.NewAtom("some_assign", $1, $3)
    }
    | TOK_SOME bounds fmla
    {
        args := append($2, $3)
        $$ = ast.NewAtom("some", args...)
    }
    | TOK_SOME bounds fmla TOK_MINIMIZING term
    {
        args := append($2, $3, $5)
        $$ = ast.NewAtom("some_min", args...)
    }
    | TOK_SOME bounds fmla TOK_MAXIMIZING term
    {
        args := append($2, $3, $5)
        $$ = ast.NewAtom("some_max", args...)
    }
    ;

bounds:
    params TOK_DOT
    {
        $$ = $1
    }
    | TOK_LPAREN lparams TOK_RPAREN
    {
        $$ = $2
    }
    ;

invariants:
    /* empty */
    {
        $$ = nil
    }
    | invariants TOK_INVARIANT labeledfmla
    {
        $$ = append($1, $3)
    }
    | invariants TOK_INVARIANT labeledfmla TOK_PROOF proofstep
    {
        $$ = append($1, $3, $5)
    }
    ;

decreases:
    /* empty */
    {
        $$ = nil
    }
    | TOK_DECREASES fmla
    {
        $$ = []ast.Node{$2}
    }
    ;

// --- eqn / eqns ---

eqn:
    SYMBOLx TOK_EQ SYMBOLx
    {
        $$ = ast.NewDefinition(ast.NewApp(ast.NewSymbol($1, nil)), ast.NewApp(ast.NewSymbol($3, nil)))
    }
    ;

eqns:
    eqn
    {
        $$ = []ast.Node{$1}
    }
    | eqns TOK_COMMA eqn
    {
        $$ = append($1, $3)
    }
    ;

// ============================================================
// --- Scenario ---
// ============================================================

sceninit:
    TOK_ARROW places
    {
        $$ = &ast.PlaceList{Elems: $2}
    }
    ;

places:
    SYMBOLx
    {
        $$ = []ast.Node{ast.NewAtom($1)}
    }
    | places TOK_COMMA SYMBOLx
    {
        $$ = append($1, ast.NewAtom($3))
    }
    ;

scentranss:
    /* empty */
    {
        $$ = nil
    }
    | scentranss scentrans
    {
        $$ = append($1, $2)
    }
    ;

scentrans:
    places TOK_ARROW places TOK_COLON scenariomixin
    {
        from := &ast.PlaceList{Elems: $1}
        to := &ast.PlaceList{Elems: $3}
        $$ = &ast.ScenarioTransition{From: from, To: to, Action: $5}
    }
    | places TOK_COLON scenariomixin
    {
        from := &ast.PlaceList{Elems: $1}
        $$ = &ast.ScenarioTransition{From: from, To: &ast.PlaceList{}, Action: $3}
    }
    ;

scenariomixin:
    TOK_BEFORE atype optargs optreturns sequence
    {
        atom := ast.NewAtom($2.(*ast.Symbol).Rep)
        lalrLabelCounter++
        mixerName := fmt.Sprintf("%s[before%d]", atom.Rep, lalrLabelCounter)
        mixer := ast.NewAtom(mixerName)
        adef := &ast.ActionDef{Name: atom, Body: $5, FormalParams: $3, FormalReturns: $4}
        $$ = &ast.ScenarioBeforeMixin{Mixer: mixer, Def: adef}
    }
    | TOK_AFTER atype optargs optreturns sequence
    {
        atom := ast.NewAtom($2.(*ast.Symbol).Rep)
        lalrLabelCounter++
        mixerName := fmt.Sprintf("%s[after%d]", atom.Rep, lalrLabelCounter)
        mixer := ast.NewAtom(mixerName)
        adef := &ast.ActionDef{Name: atom, Body: $5, FormalParams: $3, FormalReturns: $4}
        $$ = &ast.ScenarioAfterMixin{Mixer: mixer, Def: adef}
    }
    ;

// ============================================================
// --- Proof / tactic ---
// ============================================================

pflet:
    var TOK_EQ fmla
    {
        $$ = ast.NewDefinition($1, $3)
    }
    ;

pflets:
    pflet
    {
        $$ = []ast.Node{$1}
    }
    | pflets TOK_COMMA pflet
    {
        $$ = append($1, $3)
    }
    ;

tacticwithelem:
    TOK_INVARIANT labeledfmla
    {
        $$ = $2
    }
    | TOK_DEFINITION atype TOK_EQ fmla
    {
        $$ = ast.NewDefinition($2, $4)
    }
    | TOK_TRIGGER atype TOK_WITH terms
    {
        $$ = &ast.Trigger{Terms: append([]ast.Node{$2}, $4...)}
    }
    ;

tacticwithlist:
    tacticwithelem
    {
        $$ = []ast.Node{$1}
    }
    | tacticwithlist tacticwithelem
    {
        $$ = append($1, $2)
    }
    ;

tacticwithlistchoice:
    tacticwithlist
    {
        $$ = &ast.TacticWith{Elems: $1}
    }
    | pflets
    {
        $$ = &ast.TacticLets{Lets: $1}
    }
    ;

opttacticwith:
    /* empty */
    {
        $$ = &ast.TacticWith{}
    }
    | TOK_WITH tacticwithlistchoice
    {
        $$ = $2
    }
    | TOK_WITH TOK_LCB tacticwithlist TOK_RCB
    {
        $$ = &ast.TacticWith{Elems: $3}
    }
    ;

proofgroup:
    TOK_LCB proofseq TOK_RCB
    {
        $$ = $2
    }
    | TOK_LCB TOK_RCB
    {
        $$ = &ast.NullTactic{}
    }
    ;

optproofgroup:
    /* empty */
    {
        $$ = &ast.NoneAST{}
    }
    | proofgroup
    {
        $$ = $1
    }
    ;

proofseq:
    proofstep
    {
        $$ = $1
    }
    | proofseq TOK_SEMI proofstep
    {
        $$ = &ast.ComposeTactics{Tactics: []ast.Node{$1, $3}}
    }
    | proofseq proofstep
    {
        $$ = &ast.ComposeTactics{Tactics: []ast.Node{$1, $2}}
    }
    ;

// --- match/renaming ---

match:
    defn
    {
        $$ = $1
    }
    | var TOK_EQ fmla
    {
        $$ = ast.NewDefinition($1, $3)
    }
    ;

matches:
    match
    {
        $$ = []ast.Node{$1}
    }
    | matches TOK_COMMA match
    {
        $$ = append($1, $3)
    }
    ;

renamingitem:
    TOK_VARIABLE TOK_DIV TOK_VARIABLE
    {
        $$ = ast.NewDefinition(&ast.Variable{Rep: $3}, &ast.Variable{Rep: $1})
    }
    | SYMBOLx TOK_DIV SYMBOLx
    {
        $$ = ast.NewDefinition(ast.NewAtom($3), ast.NewAtom($1))
    }
    ;

renaminglist:
    renamingitem
    {
        $$ = []ast.Node{$1}
    }
    | renaminglist TOK_COMMA renamingitem
    {
        $$ = append($1, $3)
    }
    ;

optrenaming:
    /* empty */
    {
        $$ = &ast.Renaming{}
    }
    | renaming
    {
        $$ = $1
    }
    ;

renaming:
    TOK_LT renaminglist TOK_GT
    {
        r := &ast.Renaming{Elems: $2}
        $$ = r
    }
    ;

// --- proofstep ---

proofstep:
    TOK_APPLY atype optrenaming
    {
        $$ = &ast.SchemaInstantiation{SchemaName: $2, Ren: $3}
    }
    | TOK_APPLY atype optrenaming TOK_WITH matches
    {
        $$ = &ast.SchemaInstantiation{SchemaName: $2, Ren: $3, Matches: $5}
    }
    | TOK_ASSUME atype optrenaming
    {
        $$ = &ast.AssumeTactic{SchemaName: $2, Ren: $3}
    }
    | TOK_ASSUME atype optrenaming TOK_WITH matches
    {
        $$ = &ast.AssumeTactic{SchemaName: $2, Ren: $3, Matches: $5}
    }
    | TOK_INSTANTIATE atype optrenaming
    {
        $$ = &ast.AssumeTactic{SchemaName: $2, Ren: $3}
    }
    | TOK_INSTANTIATE labelname atype optrenaming
    {
        $$ = &ast.AssumeTactic{SchemaName: $3, Ren: $4}
    }
    | TOK_INSTANTIATE atype optrenaming TOK_WITH matches
    {
        $$ = &ast.AssumeTactic{SchemaName: $2, Ren: $3, Matches: $5}
    }
    | TOK_INSTANTIATE TOK_WITH pflets
    {
        $$ = &ast.WitnessTactic{Witnesses: $3}
    }
    | TOK_SHOWGOALS
    {
        $$ = &ast.ShowGoalsTactic{}
    }
    | TOK_DEFERGOAL
    {
        $$ = &ast.DeferGoalTactic{}
    }
    | TOK_SPOIL atype
    {
        $$ = &ast.SpoilTactic{Target: $2}
    }
    | TOK_TACTIC atype opttacticwith optproofgroup
    {
        $$ = &ast.TacticTactic{TName: $2, Body: $3, Proof: $4}
    }
    | opttemporal TOK_PROPERTY labeledfmla optskolem optproofgroup
    {
        lf := addLabel($3.(*ast.LabeledFormula), "prop")
        if $1 != nil {
            t := true
            lf.Temporal = &t
        }
        name := $4
        if name == nil { name = &ast.NoneAST{} }
        proof := $5
        $$ = &ast.PropertyTactic{Prop: lf, PName: name, Proof: proof}
    }
    | TOK_FUNCTION funs
    {
        $$ = &ast.FunctionTactic{Elems: $2}
    }
    | TOK_THEOREM lgprop optproofgroup
    {
        lf := addLabel($2.(*ast.LabeledFormula), "thm")
        $$ = &ast.PropertyTactic{Prop: lf, PName: &ast.NoneAST{}, Proof: $3}
    }
    | TOK_PROOF labelname proofgroup
    {
        $$ = &ast.ProofTactic{TLabel: ast.NewAtom($2), Proof: $3}
    }
    | TOK_LET pflets
    {
        $$ = &ast.LetTactic{Defs: $2}
    }
    | TOK_IF fmla proofgroup TOK_ELSE proofgroup
    {
        $$ = &ast.IfTactic{Cond: $2, Then: $3, Else: $5}
    }
    | TOK_UNFOLD atype TOK_WITH callatoms
    {
        args := make([]ast.Node, len($4))
        copy(args, $4)
        $$ = &ast.UnfoldTactic{Premise: $2, UnfSpecs: args}
    }
    | TOK_UNFOLD TOK_WITH callatoms
    {
        $$ = &ast.UnfoldTactic{Premise: &ast.NoneAST{}, UnfSpecs: $3}
    }
    | TOK_FORGET callatoms
    {
        $$ = &ast.ForgetTactic{Names: $2}
    }
    | proofgroup
    {
        $$ = $1
    }
    ;

// ============================================================
// --- Update / State / Concept (less common) ---
// ============================================================

requires:
    /* empty */
    {
        $$ = &ast.And{}
    }
    | TOK_REQUIRES fmla
    {
        $$ = $2
    }
    ;

ensures:
    TOK_ENSURES fmla
    {
        $$ = $2
    }
    ;

modifies:
    /* empty */
    {
        $$ = nil
    }
    | TOK_MODIFIES TOK_LCB TOK_RCB
    {
        $$ = &ast.And{} // empty modifies
    }
    | TOK_MODIFIES TOK_TIMES
    {
        $$ = nil
    }
    | TOK_MODIFIES atoms
    {
        $$ = &ast.And{Terms: $2}
    }
    ;

upaxes:
    /* empty */
    {
        $$ = nil
    }
    | upaxes upax
    {
        $$ = append($1, $2)
    }
    ;

upax:
    TOK_PARAMS tterms TOK_IN action TOK_ARROW requires ensures
    {
        $$ = ast.NewAtom("upax", $6, $7)
    }
    ;

assert_rhs:
    TOK_LCB requires modifies ensures TOK_RCB
    {
        $$ = ast.NewAtom("rme", $2, $4)
    }
    | fmla
    {
        $$ = $1
    }
    ;

state_expr:
    TOK_TRUE
    {
        $$ = &ast.And{}
    }
    | TOK_FALSE
    {
        $$ = &ast.Or{}
    }
    | SYMBOLx
    {
        $$ = ast.NewAtom($1)
    }
    | SYMBOLx TOK_LPAREN state_expr TOK_RPAREN
    {
        $$ = ast.NewAtom($1, $3)
    }
    | state_expr TOK_OR state_expr
    {
        $$ = &ast.Or{Terms: []ast.Node{$1, $3}}
    }
    | TOK_LCB requires modifies ensures TOK_RCB
    {
        $$ = ast.NewAtom("rme", $2, $4)
    }
    | TOK_ENTRY
    {
        $$ = ast.NewAtom("entry")
    }
    ;

// --- Concept space ---

cdefn:
    atom TOK_EQ expr
    {
        $$ = ast.NewDefinition(ast.AppToAtom($1), $3)
    }
    ;

cdefns:
    cdefn
    {
        $$ = []ast.Node{$1}
    }
    | cdefns TOK_COMMA cdefn
    {
        $$ = append($1, $3)
    }
    ;

expr:
    TOK_LCB fmla TOK_RCB
    {
        $$ = $2
    }
    | exprterm
    {
        $$ = $1
    }
    | exprterm relop exprterm
    {
        $$ = ast.NewAtom($2, $1, $3)
    }
    | exprterm TOK_TILDAEQ exprterm
    {
        $$ = &ast.Not{Body: ast.NewAtom("=", $1, $3)}
    }
    | TOK_TILDA expr
    {
        $$ = &ast.Not{Body: $2}
    }
    | TOK_LPAREN expr TOK_RPAREN
    {
        $$ = $2
    }
    | prod
    {
        $$ = ast.NewAtom("product", $1...)
    }
    | sum
    {
        $$ = ast.NewAtom("sum", $1...)
    }
    ;

exprterm:
    appelem
    {
        $$ = $1
    }
    | var
    {
        $$ = $1
    }
    ;

prod:
    expr TOK_TIMES expr
    {
        $$ = []ast.Node{$1, $3}
    }
    | prod TOK_TIMES expr
    {
        $$ = append($1, $3)
    }
    ;

sum:
    expr TOK_PLUS expr
    {
        $$ = []ast.Node{$1, $3}
    }
    | sum TOK_PLUS expr
    {
        $$ = append($1, $3)
    }
    ;

// --- loc (for old v1.1 compat) ---

loc:
    /* empty */
    {
        $$ = ""
    }
    | SYMBOLx
    {
        $$ = $1
    }
    ;

// --- symbols list ---

symbols:
    SYMBOLx
    {
        $$ = []ast.Node{ast.NewAtom($1)}
    }
    | symbols TOK_COMMA SYMBOLx
    {
        $$ = append($1, ast.NewAtom($3))
    }
    ;

%%

// lalrMakeSequence wraps a list of action nodes into a single sequence node.
func lalrMakeSequence(stmts []ast.Node) ast.Node {
	if len(stmts) == 0 {
		return &ast.Sequence{}
	}
	if len(stmts) == 1 {
		return stmts[0]
	}
	return &ast.Sequence{Stmts: stmts}
}
