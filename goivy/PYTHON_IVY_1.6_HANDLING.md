# Python Ivy 1.6 Handling

This note records how the Python Ivy implementation supports `#lang ivy1.6`
versus `#lang ivy1.7` and later. It is intended as a design reference for
future `goivy` work.

The current partial `goivy` support for Ivy 1.6 is intentionally out of scope
here. Treat the Python implementation as the source of truth; existing Go 1.6
artifacts are known incomplete and should not drive the design.

Generated parser tables such as `ivy_parsetab.py`, `ivy_formulatab.py`,
`ivy_termtab.py`, `ivy_dafny_parsetab.py`, and other `ivy_*tab.py` files are
also out of scope. The meaningful architecture lives in the handwritten Python
files listed below.

## Source Inventory

The compact source inventory for Python Ivy 1.6 versus 1.7 handling is:

- `pyivy/ivy/ivy/ivy_utils.py`: `564-583`, `599-609`
- `pyivy/ivy/ivy/ivy_lexer.py`: `253-304`
- `pyivy/ivy/ivy/ivy_compiler.py`: `1788-1799`, `1997-2004`,
  `2405-2409`, `2561-2566`, `2609-2631`, `2643-2652`
- `pyivy/ivy/ivy/ivy_logic_parser.py`: `83-117`, `157-175`,
  `475-555`, `776-783`
- `pyivy/ivy/ivy/ivy_parser.py`: `22-40`, `520-640`, `839-875`,
  `1079-1089`, `1243-1348`, `1372-1379`, `1654-1698`, `1731-1739`,
  `2196-2216`, `2324-2331`, `2351-2362`, `2431-2438`, `2602-2608`,
  `2799-2864`, `3183-3190`, `3245-3268`, `3666-3678`
- `pyivy/ivy/ivy/ivy_ast.py`: `683-687`, `1116-1128`
- `pyivy/ivy/ivy/ivy_actions.py`: `406-412`
- `pyivy/ivy/ivy/ivy_isolate.py`: `226-231`, `470-472`, `803-840`,
  `1165-1169`, `1506-1508`, `1764-1776`, `1994-1999`
- `pyivy/ivy/ivy/ivy_theory.py`: `13-97`, `154-161`
- `pyivy/ivy/ivy/ivy_to_cpp.py`: `2989-2992`, `4570-4571`,
  `4608-4609`, `4628-4631`, `4753-4758`
- include library directories: `pyivy/ivy/ivy/include/1.6`,
  `pyivy/ivy/ivy/include/1.7`, `pyivy/ivy/ivy/include/1.8`

## Architectural Overview

Python Ivy does not maintain a separate full source file grammar for Ivy 1.6.
It also does not simply parse everything as Ivy 1.7 and patch the result after
the fact.

Instead, Python uses one version-aware front end:

1. `ivy_compiler.read_module` reads the first line of the file, extracts
   `#lang ivy...`, and calls `ivy_utils.set_string_version`.
2. `ivy_utils` stores global language-version state and derives global flags
   from it.
3. If the language version changes, `read_module` clears PLY rule functions
   from `ivy_logic_parser` and `ivy_parser`, then reloads both modules.
4. At import time, `ivy_logic_parser.py` and `ivy_parser.py` execute top-level
   `if iu.get_numeric_version() <= [1,6]` conditionals. Those conditionals
   decide which grammar production functions exist.
5. During parsing, `ivy_parser.parse` wraps PLY in `ivy_lexer.LexerVersion`,
   which edits the reserved-word table for the active language version.
6. Downstream AST, compiler, isolate, theory, and C++ generation code also
   branches on `<=1.6` or `>=1.7`.

The key design point is that Python makes Ivy 1.6 a coordinated language mode:
lexing, grammar shape, AST declarations, compiler semantics, isolate behavior,
built-in theories, include libraries, and code generation all participate.

## Version Selection And Global State

`pyivy/ivy/ivy/ivy_utils.py`

- `564-569`: default language version is the latest Python default (`1.7`);
  version-derived globals are initialized.
- `572-583`: `set_string_version` records the string version and derives:
  - `ivy_compose_character`, `:` for `<=1.1`, otherwise `.`
  - `symbol_chars_parser`
  - `ivy_have_polymorphism`, false for `<=1.2`
  - `ivy_use_polymorphic_macros`, false for `<=1.5`
  - `ivy_forbid_ghost_init`, false for `<=1.6`, true from `1.7`
- `588-595`: string/numeric version conversion and `version_le`.
- `599-609`: `get_std_include_dir` chooses the smallest include directory
  version `d` such that `current_version <= d`; `ivy1.6` selects `include/1.6`
  when available.

`pyivy/ivy/ivy/ivy_compiler.py`

- `1788-1799`: `get_file_version` reads a file header and returns the numeric
  version from `#lang ivy...`.
- `2609-2631`: `read_module` reads the header, sets the global string version,
  rejects nested include files whose version differs from the including file,
  clears PLY rules, reloads `ivy_logic_parser` and `ivy_parser`, installs the
  importer, and parses.
- `2643-2652`: `import_module` falls back to `iu.get_std_include_dir`, so
  standard libraries are version-selected.

## Lexer Handling

`pyivy/ivy/ivy/ivy_lexer.py`

- `253-304`: `LexerVersion` is a context manager that replaces the global
  reserved-word table while parsing.
- `263-266`: for `<=1.0`, removes `state` and `local`.
- `267-275`: for `<=1.1`, removes `returns`, `mixin`, `before`, `after`,
  `isolate`, `with`, `export`, `delegate`, `import`, and `include`; otherwise
  removes old `state`, `set`, `null`, and `match`.
- `276-281`: for `<=1.4`, removes many later keywords, including `function`,
  `class`, `object`, `method`, `some`, `property`, `while`, `invariant`,
  `definition`, `ghost`, `this`, `var`, `scenario`, `proof`, `named`, and
  `fresh`.
- `282-286`: for `<=1.5`, removes `variant`, `of`, `globally`, `eventually`,
  and `temporal`.
- `287-290`: for `<=1.6`, removes 1.7-era keywords: `decreases`,
  `specification`, `implementation`, `require`, `ensure`, `around`,
  `parameter`, `apply`, `theorem`, `showgoals`, `spoil`, `explicit`, `thunk`,
  `isa`, `autoinstance`, `constructor`, `tactic`, `finite`, `unfold`, and
  `forget`.
- `291-298`: for `<=1.7`, removes post-1.7 keywords such as `global`,
  `common`, `debug`, `field`, `for`, `process`, `subclass`, `template`,
  `whenfirst`, `whenlast`, `whennext`, `whenprev`, `unprovable`, and
  `trigger`; for versions after 1.7, removes old `requires` and `ensures`.

This means several 1.6 versus 1.7 differences are lexical, not just AST
rewrites. For example, `apply`, `tactic`, `theorem`, `require`, and `ensure`
are not keywords in Ivy 1.6.

## Parser Reload Architecture

`pyivy/ivy/ivy/ivy_parser.py`

- `3666-3678`: `parse` resets parser globals for top-level parses, reads the
  current numeric version, runs PLY under `LexerVersion(vernum)`, expands
  autoinstances only for top-level parses, then raises accumulated parse
  errors.

`pyivy/ivy/ivy/ivy_compiler.py`

- `2620-2629`: when `#lang` changes the version, Python clears `p_*` rules
  from `ivy_logic_parser` and `ivy_parser` and reloads the modules. The
  top-level `if version <= 1.6` statements in those modules then define a
  different active grammar.

This is the closest Python equivalent to a separate grammar: the source file is
shared, but the active PLY grammar is rebuilt under the selected version.

## Logic Parser Differences

`pyivy/ivy/ivy/ivy_logic_parser.py`

- `83-117`: for `<=1.6`, applications are parsed through `aterm`, with
  repeated calls mutating/extending the same application node; dotted names
  compose `aterm DOT SYMBOL`. For later versions, `appelem` is separate and
  supports `SYMBOL` and `SYMBOL(terms)`.
- `157-175`: for `<=1.6`, `term : aterm`; old terms bind as `OLD aterm`.
  The later `term DOT aterm` rule is guarded away here, so the 1.6 shape is
  narrower.
- `475-555`: for `<=1.6`, formulas are structurally distinct from terms:
  comparisons, `~=`, parentheses, `true`, `false`, negation, conjunction,
  disjunction, implication, iff, and quantifiers are all formula productions.
  The comment at `473` is explicit: before version 1.7, formulas cannot be
  terms.
- `547-555`: 1.6 quantifier productions are `FORALL simplevars DOT fmla` and
  `EXISTS simplevars DOT fmla`.
- `776-783`: `term ISA atype` exists only after 1.6.

Implication: a faithful Go design cannot treat formula parsing as a trivial
1.7 parse plus postprocessing. Some 1.6 differences affect grammar and
precedence.

## Main Ivy Parser Differences

`pyivy/ivy/ivy/ivy_parser.py`

- `22-40`: the 1.7+ precedence table is only installed when the version is
  later than 1.6.
- `520-526`: `addlabel` does not synthesize labels for `<=1.6`. From 1.7,
  unlabeled formulas receive generated labels.
- `528-563`: 1.6 axiom syntax is `top opttemporal AXIOM labeledfmla`; 1.7
  adds `optexplicit`, `lgprop`, labels, and schema-body-capable propositions.
- `576-589`: property syntax exists across versions but interacts with
  version-specific `addlabel`, `optexplicit`, and proof/skolem handling.
- `591-597`: `conjecture` remains a top-level declaration.
- `609-640`: from 1.7, `invariant` replaces `conjecture` for invariant-style
  declarations and adds `unprovable invariant`.
- `839-859`: schema declarations use old `PROPERTY labeledfmla` in 1.6; 1.7
  supports `optexplicit PROPERTY lgprop` and `THEOREM lgprop`.
- `866-875`: schema conclusions use old `PROPERTY fmla` in 1.6; 1.7 uses the
  richer labeled property form.
- `1079-1089`: `CONSTRUCTOR tterms` exists only after 1.6.
- `1243-1250`: old proof step syntax in 1.6 is just `SYMBOL`, producing a
  schema instantiation.
- `1251-1348`: 1.7 adds proof steps for `APPLY`, global `ASSUME`,
  `INSTANTIATE`, labeled instantiation, `SHOWGOALS`, `DEFERGOAL`, `SPOIL`,
  `TACTIC`, proof-level `PROPERTY`, proof-level `FUNCTION`, proof-level
  `THEOREM`, and labeled proof groups.
- `1372-1379`: 1.6 supports `SYMBOL WITH matches` schema instantiation.
- `1380-1385` and following: 1.7 introduces renaming syntax such as
  `VARIABLE / VARIABLE` and related proof renaming infrastructure.
- `1654-1661`: 1.6 definition declaration syntax is
  `DEFINITION defns optproof`.
- `1662-1698`: 1.7 definition syntax adds `optexplicit`, optional labels,
  generic definition bodies, and definition schemas.
- `1731-1739`: top-level `init labeledfmla` exists only in 1.6; it is banned
  as of 1.7.
- `2196-2216`: `around` mixins exist only after 1.6.
- `2324-2331`: top-level `private callatom` is present in 1.6 under this rule.
- `2351-2362`: `specification`, `implementation`, and newer private attribute
  parsing exist only after 1.6.
- `2431-2438`: top-level old `ASSERT SYMBOL -> assert_rhs` exists only in 1.6.
- `2602-2608`: scenario mixin names are deterministic in 1.6 (`[before]`);
  later versions append a fresh counter (`[beforeN]`).
- `2799-2810`: 1.6 action assertions are `ASSERT labeledfmla` and
  `ENSURES labeledfmla`.
- `2811-2864`: 1.7 action assertions add `unprovable`, proof-bearing
  `assert`, `require`, and `ensure`.
- `3183-3190`: caret/key parameters (`^x:T`) exist only after 1.6.
- `3245-3252`: `thunk` actions exist only after 1.6.
- `3256-3268`: debug action arguments exist only after 1.6.

## AST And Action Semantics

`pyivy/ivy/ivy/ivy_ast.py`

- `683-687`: `LabeledDecl.defines` returns no definitions for `<=1.6`. From
  1.7, labeled declarations define their labels.
- `1116-1128`: `InterpretDecl.defines` includes the interpret-label definition
  only after 1.6, but always includes non-numeric range endpoints.

`pyivy/ivy/ivy/ivy_actions.py`

- `406-412`: before 1.7, `EnsuresAction` is always verified. From 1.7 it uses
  the `AssertAction.assert_to_assume` behavior.

## Compiler Differences

`pyivy/ivy/ivy/ivy_compiler.py`

- `1997-2004`: action interference checking has a `>=1.7` branch for checking
  action side effects against axioms and definitions.
- `2405-2409`: `create_conj_actions` exits immediately for `<=1.6`; object
  invariant preservation is a 1.7 feature.
- `2561-2566`: from 1.7, if there is no explicit global isolate, the compiler
  creates the implicit `this = this` isolate.

## Isolate Differences

`pyivy/ivy/ivy/ivy_isolate.py`

- `226-231`: `vstartswith_eq_some` bypasses implementation-map handling for
  `<=1.6`.
- `470-472`: isolate parameter stripping clears `mod.params` for `<=1.6`.
- `803-805`: property proof classification uses the original 1.6 algorithm
  for `<=1.6`.
- `806-821`: later versions reset privates, compute implementation visibility,
  and account for other isolates' verified objects.
- `837-840`: derived definitions are keyed differently in `<=1.6` versus later
  versions.
- `1165-1169`: conjecture filtering uses `keep_ax` for `<=1.6`; later versions
  filter by verified/present isolate names.
- `1506-1508`: interference term checking is enabled only when enforcing axioms
  and running `>=1.7`.
- `1764-1769`: from 1.7, if no isolate is specified and there is exactly one
  isolate, that isolate is selected automatically.
- `1774-1776`: present conjectures are applied only for named isolates in
  `>=1.7`.
- `1994-1999`: present conjecture bracketing is applied only for named isolates
  in `>=1.7`.

## Built-In Theories And Includes

`pyivy/ivy/ivy/ivy_theory.py`

- `13-53`: `theories_1_6`, especially the old `ind` schema written with
  anonymous premise blocks rather than theorem blocks.
- `54-94`: `theories_1_7`, where `ind` uses `theorem [base]` and
  `theorem [step]`.
- `96-97`: `theories()` selects `theories_1_7` for `>=1.7`, otherwise
  `theories_1_6`.
- `154-161`: `get_theory_schemata` emits theory schemata only for versions
  where `"1.6" <= current_version`. It maps range sorts, `nat`, and `bv[...]`
  to the `int` schemas.

Include directories:

- `pyivy/ivy/ivy/include/1.6`
- `pyivy/ivy/ivy/include/1.7`
- `pyivy/ivy/ivy/include/1.8`

`ivy_utils.get_std_include_dir` chooses the smallest include directory that is
greater than or equal to the active language version. In practice, `#lang
ivy1.6` uses the `1.6` standard library when present.

## C++ Generation Differences

`pyivy/ivy/ivy/ivy_to_cpp.py`

- `2989-2992`: inconsistent initial conditions include "Initial condition
  and/or axioms are inconsistent" in 1.6, but only "Axioms are inconsistent"
  later.
- `4570-4571`: for `<=1.6`, code generation sets `interpret_all_sorts`.
- `4608-4609`: if there is no isolate, later versions compile `this`; 1.6
  leaves the isolate as `None`.
- `4628-4631`: REPL isolate conversion through `ExtractDef` is gated to
  later versions.
- `4753-4758`: REPL/test parameter description emission is gated to versions
  after 1.6.

## Major Semantic Deltas

- Ivy 1.6 has old proof syntax and schema instantiation syntax. Ivy 1.7 adds
  `apply`, `tactic`, `theorem`, proof groups, and renaming syntax.
- Ivy 1.6 keeps `conjecture` as the invariant-like top-level declaration.
  Ivy 1.7 introduces object/invariant semantics around `invariant` and
  `unprovable invariant`.
- Ivy 1.6 allows top-level `init`; Ivy 1.7 bans it.
- Ivy 1.6 has old `ensures` action behavior. Ivy 1.7 adds `require`,
  `ensure`, `unprovable`, and proof-bearing assertion forms.
- Ivy 1.6 formula parsing is structurally different from Ivy 1.7's
  term/formula unification.
- Ivy 1.7 adds automatic/global isolate behavior, present conjecture
  bracketing, and object-invariant preservation.
- Built-in theory schemata and standard include directories are selected by
  language version.

## Design Implications For Go

The Python design argues against a simplistic "parse with 1.7 and special-case
afterward" approach. Some differences are lexical, some are grammar-level, some
are AST definition semantics, and some are compiler/isolate/codegen behavior.

The Python design also does not require a wholly separate handwritten module
grammar per version. The source is shared, but the active grammar is rebuilt
under the selected version. A Go design should preserve that spirit: one
coherent front end whose behavior is version-aware, with explicit tests that
old syntax is accepted only in the intended language mode and newer syntax does
not leak backward.
