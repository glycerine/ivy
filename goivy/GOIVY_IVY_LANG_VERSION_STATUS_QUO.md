# Goivy Ivy Language Version Status Quo

This note records how the current Go implementation handles Ivy language
versions `1.6`, `1.7`, and `1.8`. It is descriptive, not prescriptive: it
documents the code as it exists today so future design work can distinguish
working version plumbing from partial or misleading compatibility artifacts.

Repo-relative paths are used throughout. Generated goyacc `.go` files are not
treated as source of truth when the corresponding `.y` grammar is available.

## Executive Summary

Goivy has several real language-version mechanisms:

- `#lang ivyN.N` is read by the compiler entry points.
- `IvyUtilsConfig` stores a string language version and derives Python-style
  global flags from it.
- The lexer removes reserved words according to the requested version.
- Standard include directories are selected by version.
- The standalone logic/formula parser dispatches among v1.2, v1.6, and v1.7+
  grammars.
- Some compiler, action, isolate, theory, and code generator paths branch on
  `<=1.6`, `>=1.7`, or `>1.7`.

The full Ivy-file parser now dispatches by version for the first major split:
`<=1.6` uses a dedicated thin v1.6 full-file grammar, while `>=1.7` continues
to use the hardened v1.7+ grammar. The v1.6 grammar is intentionally initial
and conservative: it covers the core source forms needed to enter the shared
post-parse pipeline, but it is not yet a complete port of every Python Ivy
1.6 production. There is still no separate full-file v1.8 grammar.

The implementation also has inconsistent defaults: `IvyUtilsConfig` defaults
to language version `1.8`, the public logic parser defaults to `Version{1,7}`,
in-memory source parsing defaults to `1.7`, the web UI inserts `#lang ivy1.7`
when a source has no header, and `IsolateConfig` defaults to `1.7`.

## Source Inventory

Core version state and file loading:

- `goivy/ivyutils_config.go`: `74-125`, `127-129`
- `goivy/ivyutils_names.go`: `71-99`, `118-168`, `170-208`, `247-292`
- `goivy/compiler_ivyinit.go`: `48-116`, `119-201`, `204-242`
- `goivy/compiler_phase6.go`: `2001-2019`, `2804-2879`
- `goivy/compiler_stdlib.go`: `27-47`, `74-113`, `116-141`

Lexer and parsers:

- `goivy/lexer.go`: `9-217`
- `goivy/parser_lalr_parser.go`: `7-22`, `24-56`, `84-88`, `159-163`
- `goivy/parser_parse_config.go`: `1-96`
- `goivy/parser_v16_lalr_parser.go`: `7-32`, `35-176`
- `goivy/parser_v16_builders.go`: `5-55`
- `goivy/parser_grammar_v16.y`: `1-5`, `67-199`, `201-256`,
  `258-341`, `357-407`, `409-493`, `495-709`
- `goivy/parser_grammar_v17.y`: `1-10`, `70-82`, `774-860`,
  `981-1039`, `1694-1702`, `1901-2253`, `2303-2358`, `3621-3658`,
  `4014-4093`, `4936-5110`, `5168-5185`
- `goivy/logicparser.go`: `9-52`
- `goivy/lalr_logicparser_lalr_parser.go`: `10-23`, `25-38`, `50-57`
- `goivy/v16_lalr_v16.go`: `7-18`, `28-35`
- `goivy/v16_logicgrammar_v16.y`: `1-8`, `57-74`, `113-242`
- `goivy/v12_lalr_v12.go`: `7-18`, `28-35`
- `goivy/lalr_logicparser_grammar_v17.y`: `1-13`, `200-215`,
  `463-501`

Downstream semantic branches:

- `goivy/ivylogic.go`: `402-418`
- `goivy/actions_transforms.go`: `52-77`
- `goivy/compiler_ivy_compile.go`: `186-195`, `1524-1529`,
  `1704-1713`
- `goivy/compiler_phase6.go`: `2609-2647`, `2661-2692`
- `goivy/theory.go`: `109-209`, `252-268`
- `goivy/module_config.go`: `200-221`, `224-244`
- `goivy/isolate_iter.go`: `39-50`
- `goivy/isolate_strip.go`: `888-892`
- `goivy/isolate_helpers.go`: `613-635`
- `goivy/isolate.go`: `794-800`, `1088-1093`, `1614-1622`
- `goivy/isolate_create.go`: `208-225`, `469-476`

Front doors and generators:

- `goivy/isolist.go`: `51-78`
- `goivy/webui/webui_session.go`: `144-150`
- `goivy/ivy2cpp/compile.go`: `62-75`, `326-345`
- `goivy/ivy2go/compile.go`: `67-75`, `226-247`

Include directories currently present in the repository:

- `pyivy/ivy/ivy/include/1.6`
- `pyivy/ivy/ivy/include/1.7`
- `pyivy/ivy/ivy/include/1.8`
- `ivy-lang-examples/ivy/include/1.6`
- `ivy-lang-examples/ivy/include/1.7`
- `ivy-lang-examples/ivy/include/1.8`

## Version State

`goivy/ivyutils_config.go`

- `74-93`: `NewIvyUtilsConfig` defaults `LanguageVersion` and
  `LatestLanguageVersion` to `1.8`, then calls `SetStringVersion`.
- `111-125`: `SetStringVersion` trims the string version, clears the cached
  include directory, computes the numeric version, and derives:
  - `ComposeCharacter`, `:` for `<=1.1`, otherwise `.`
  - `SymbolCharsParser`
  - `HavePolymorphism`, false for `<=1.2`
  - `UsePolymorphicMacros`, false for `<=1.5`
  - `ForbidGhostInit`, false for `<=1.6`
- `127-129`: `GetNumericVersion` re-parses the current string version.

`goivy/ivyutils_names.go`

- `71-99`: `SetStringVersionOn` is a thin wrapper, and `versionLESlice`
  implements Python-like list comparison for numeric version slices.
- `247-292`: `VersionLE` and `StringVersionToNumericVersion` provide the
  string/numeric helpers used throughout the package.

Important status quo: this is per-config state, not a package-global Python
style variable. That is good for concurrency, but every subsystem has to share
the same `Config` for version behavior to agree.

## File Loading

`goivy/compiler_ivyinit.go`

- `48-134`: `ReadModule` reads the first line, requires `#lang ivy...`,
  strips it from the body while preserving line numbers, calls
  `setConfigLanguageVersion` when a non-empty version suffix is present,
  rejects nested includes whose version differs from the including file, then
  calls `Parse` with the parsed version.
- `136-156`: `ReadModuleFromString` parses an in-memory source by calling
  `parseIvySource`, then propagates the parsed version into config.
- `158-218`: `ReadModuleFromNamedString` mirrors `ReadModule` for in-memory
  sources with a filename and propagates the parsed version into config.
- `221-259`: imports first try a preloaded standard library and then fall back
  to `cfg.IuCfg.GetStdIncludeDir()`.

`goivy/compiler_phase6.go`

- `2001-2019`: `GetFileVersion` extracts the `#lang ivy...` suffix from a
  file header.
- `2804-2836`: `IvyFromString` parses a source header into a `Version`, parses
  with that version, then propagates the parsed version into the fresh
  module config before compilation.
- `2838-2857`: `parseIvyVersion` defaults an empty or malformed version to
  `Version{1,7}`.
- `2860-2879`: `parseIvySource` strips a `#lang ivy...` header and defaults
  to `Version{1,7}` if there is no suffix.

Status quo implication: the main file-backed and in-memory front doors now keep
the parsed language version attached to the session config. Lower-level direct
`Parse` callers still need to coordinate parse and compile config themselves.

## Include Selection

`goivy/ivyutils_names.go`

- `118-168`: `GetStdIncludeDir` scans the include base directory and picks the
  smallest version directory `d` such that `current_version <= d`.
- `170-208`: `getIncludeBaseDir` searches `GOIVY_INCLUDE`, executable-relative
  include directories, source-tree `ivy-lang-examples/ivy/include`, and
  current-directory include paths.

`goivy/compiler_stdlib.go`

- `27-47`: `PreloadStandardLibrary` loads a versioned include tree into a
  config and pins `IncludeBaseDir`.
- `74-113`: `readStandardLibrary` loads all versioned `.ivy` files into memory.
- `116-141`: `includeSource` uses the same smallest-directory-at-least-current
  rule for preloaded includes.

For current repository directories, `1.6`, `1.7`, and `1.8` all have include
trees. Therefore a normal `#lang ivy1.6`, `#lang ivy1.7`, or `#lang ivy1.8`
file can select the matching include directory, assuming the include base path
is found.

## Lexer Handling

`goivy/lexer.go`

- `9-131`: `allReserved` contains the full keyword table, including 1.7-era
  and post-1.7 keywords.
- `133-134`: `Version` is a two-int `[major, minor]` value.
- `147-154`: `NewLexer` builds a reserved-word table for the requested version.
- `158-217`: `buildReserved` deletes keywords by version:
  - `<=1.0`: removes `state`, `local`
  - `<=1.1`: removes `returns`, `mixin`, `before`, `after`, `isolate`,
    `with`, `export`, `delegate`, `import`, `include`; otherwise removes old
    `state`, `set`, `null`, `match`
  - `<=1.4`: removes object/class/function/action-era keywords such as
    `function`, `class`, `object`, `method`, `property`, `while`,
    `invariant`, `definition`, `ghost`, `var`, `scenario`, `proof`, `named`,
    and `fresh`
  - `<=1.5`: removes `variant`, `of`, `globally`, `eventually`, `temporal`
  - `<=1.6`: removes `decreases`, `specification`, `implementation`,
    `require`, `ensure`, `around`, `parameter`, `apply`, `theorem`,
    `showgoals`, `spoil`, `explicit`, `thunk`, `isa`, `autoinstance`,
    `constructor`, `tactic`, `finite`, `unfold`, and `forget`
  - `<=1.7`: removes post-1.7 keywords including `global`, `common`,
    `debug`, `field`, `for`, `process`, `subclass`, `template`,
    `whenfirst`, `whenlast`, `whennext`, `whenprev`, `unprovable`, and
    `trigger`
  - `>1.7`: removes old plural `requires` and `ensures`

This is one of the strongest parts of current version handling. It closely
matches the Python `LexerVersion` idea, but it feeds a fixed full-file parser.

## Full Ivy-File Parser

`goivy/parser_lalr_parser.go`

- `7-22`: `Parse` dispatches `<=1.6` to `ParseFullV16` and everything else to
  `ParseV17`.
- `24-56`: `ParseV17` runs the v1.7+ parser and expands autoinstances for
  top-level parses.
- `64-264`: the v1.7 parser adapter still passes the requested `Version` to
  `NewLexer`, so version affects tokenization.

`goivy/parser_v16_lalr_parser.go`

- `7-32`: `ParseFullV16` mirrors the v1.7 parser entry point while using the
  v1.6 generated parser.
- `35-176`: the v1.6 parser adapter maps the version-aware lexer tokens into
  the generated v1.6 grammar token set.

`goivy/parser_grammar_v16.y`

- `1-5`: documents the grammar as a thin full-file v1.6 syntax adapter.
- `67-155`: top-level declarations for include, type, individual/relation,
  axiom, property, conjecture, init, action, import, and export.
- `157-240`: simple temporal markers and action bodies.
- `256-338`: v1.6 labels, constants, relations, and type symbols.
- `347-531`: v1.6-style separate term/formula grammar.

`goivy/parser_grammar_v17.y`

- `1-10`: the grammar announces itself as the full Ivy file parser for
  version `1.7+`.
- `70-82`: `parser17AddLabel` always synthesizes labels for unlabeled
  formulas. There is no active `<=1.6` branch that leaves labels unsynthesized.
- `774-860`: top-level axiom/property/conjecture/invariant parsing follows
  the v1.7+ shape: `lgprop`, `optexplicit`, proof support, `invariant`, and
  `unprovable invariant` are in the grammar.
- `981-1039`: v1.7+ definition and theorem grammar is present.
- `1694-1702`: an old v1.6 top-level `assert SYMBOL -> assert_rhs` rule is
  present as a compatibility-looking production, but its action deliberately
  does nothing.
- `1901-2253`: formulas and terms are unified in the v1.7+ style.
- `2303-2358`: `lgprop`, `gprop`, `opttemporal`, and `optunprovable` are
  v1.7+ style support productions.
- `3621-3658`: `specification`, `implementation`, `private`, `global`, and
  `common` attributes are parsed here. `global` and `common` only become
  keywords in versions after `1.7` because of lexer gating.
- `4014-4093`: action assertions include `require`, `ensure`, `unprovable`,
  and proof-bearing forms.
- `4936-5110`: proof steps include `apply`, `assume`, `spoil`, `tactic`,
  proof-level `property`, proof-level `function`, proof-level `theorem`, and
  proof groups.
- `5168-5185`: old plural update-clause syntax `requires` and `ensures`
  remains in the grammar, but the lexer removes those keywords for versions
  after `1.7`.

Status quo implication: `#lang ivy1.6` changes which words are keywords, but it
does not rebuild the full grammar into Python's v1.6 grammar. Old 1.6 proof
syntax, old schema instantiation syntax, top-level `init`, and old formula
structure are therefore not faithfully represented by the full module parser.

## Logic Parser

`goivy/logicparser.go`

- `9-11`: standalone formula/term parsing defaults to `Version{1,7}`.
- `15-25`: `ParseFormula` and `ParseTerm` both call `parseString`.
- `47-52`: `parseString` calls `ParseLogic`.

`goivy/lalr_logicparser_lalr_parser.go`

- `10-23`: `ParseLogic` dispatches by version:
  - `<=1.2`: `ParseV12`
  - `1.3` through `1.6`: `ParseV16`
  - otherwise: `ParseLogicV17`
- `25-38`: `ParseLogicV17` uses the v1.7+ logic grammar.
- `50-57`: the v1.7 logic parser also uses `NewLexer(input, version)`.

`goivy/v16_lalr_v16.go`

- `7-18`: `ParseV16` runs the v1.3-v1.6 LALR logic grammar.
- `28-35`: the v1.6 logic parser uses the version-aware lexer.

`goivy/v16_logicgrammar_v16.y`

- `1-8`: documents the key v1.6 differences: separate formulas and terms,
  comparison operators only in formula rules, different precedence, and no
  v1.7 term/formula unification.
- `57-74`: installs v1.3-v1.6 precedence.
- `113-131`: v1.6 `aterm` handles repeated application and dot composition.
- `170-191`: v1.6 `term` excludes boolean and comparison operators.
- `202-242`: v1.6 `fmla` separately handles comparisons, boolean operators,
  quantifiers, and temporal operators.

`goivy/lalr_logicparser_grammar_v17.y`

- `1-13`: identifies the v1.7+ formula/term/action grammar.
- `200-215`: v1.7+ `appelem` production.
- `463-501`: v1.7+ `isa`, sort annotation, named binders, and `fmla: term`.

Status quo implication: v1.6 expression parsing exists and is meaningful, but
that does not imply v1.6 source-file parsing exists.

## Downstream Version Branches

`goivy/ivylogic.go`

- `402-418`: default sort creation is blocked after `1.2`, using
  `sig.IuCfg.LanguageVersion`.

`goivy/actions_transforms.go`

- `52-77`: `EnsuresAction.assert_to_assume` branches at `<=1.6`. In v1.6 it
  recurses without class-converting; after 1.6 it behaves like assert-to-assume
  for the `ensure` kind.

`goivy/compiler_ivy_compile.go`

- `186-195`: after `1.6`, compilation creates the implicit `this = this`
  isolate when it is absent.
- `1524-1529`: action interference checking is gated to `>=1.7`.
- `1704-1713`: `CreateConjActions` returns immediately for `<=1.6`.

`goivy/theory.go`

- `109-177`: contains separate theory schema strings for pre-1.7 and 1.7+.
- `179-186`: `Theories` returns v1.7 schemata for `>=1.7`, otherwise v1.6
  schemata.
- `188-209`: `GetTheorySchemata` emits schemata only for versions at least
  `1.6`, mapping range sorts, `nat`, and `bv[...]` to the integer schemas.

`goivy/compiler_phase6.go`

- `2609-2647`: `CompileTheory` reads `mod.Cfg.IuCfg.GetStringVersion()` and
  uses it to select the schema text passed to `IvyCompileTheoryFromString`.
- `2661-2692`: `CompileTheories` repeats the same version-based schema
  selection for every interpreted sort in the module.

`goivy/module_config.go`

- `211-221`: `NewIsolateConfig` defaults `IvyVersion` to `1.7`.
- `224-244`: `NewConfig` creates both `IuCfg` and `IsolateCfg`, but does not
  wire `IsolateCfg.IvyVersion` to `IuCfg.LanguageVersion`.

`goivy/isolate_*`

- `goivy/isolate_iter.go:39-50`: `VStartsWithEqSome` branches at `<=1.6`,
  but it reads `mod.Cfg.IsolateCfg.IvyVersion`.
- `goivy/isolate_strip.go:888-892`: `StripIsolate` clears module parameters
  for `<=1.6`, using `mod.Cfg.IuCfg.GetStringVersion()`.
- `goivy/isolate_helpers.go:613-635`: property proof classification uses the
  original v1.6 algorithm when `IsolateCfg.IvyVersion <= 1.6`.
- `goivy/isolate.go:794-800`: conjecture filtering uses the v1.6 path when
  `IsolateCfg.IvyVersion <= 1.6`.
- `goivy/isolate.go:1088-1093`: polymorphic macro normalization is controlled
  by `IsolateCfg.IvyVersion > 1.5`.
- `goivy/isolate.go:1614-1622`: interference term checking is enabled when
  enforcing axioms and `IsolateCfg.IvyVersion >= 1.7`.
- `goivy/isolate_create.go:208-225`: automatic isolate selection and present
  conjecture setup are gated to `IsolateCfg.IvyVersion >= 1.7`.
- `goivy/isolate_create.go:469-476`: present-conjecture bracket actions are
  gated the same way.

Status quo implication: isolate behavior has many version branches. The main
source-loading paths now synchronize `IsolateCfg.IvyVersion` with `#lang
ivy...`, but lower-level tests and helper-constructed modules can still diverge
unless they set both version fields.

## Front Doors

`goivy/isolist.go`

- `51-75`: `IvyVersionSupported` reads only the first line and returns true
  for versions `>=1.6`.

`goivy/webui/webui_session.go`

- `144-150`: uploaded/source text without a `#lang ivy...` header is given
  `#lang ivy1.7`.

Status quo implication: the web-facing default is still 1.7 even though
`NewIvyUtilsConfig` defaults to 1.8.

## Code Generators

`goivy/ivy2cpp/compile.go`

- `62-75`: for `target=repl`, non-extract isolates named on the command line
  are rewritten to extract isolates for `>=1.7`; `target=test` compiles with
  invariants only for versions after `1.7`; isolate creation is automatic for
  `>=1.7`.
- `326-345`: selected-isolate defaults and helper functions compare
  `mod.Cfg.IuCfg.GetStringVersion()`.

`goivy/ivy2go/compile.go`

- `67-75`: mirrors the `ivy2cpp` version branches for repl/test/isolate
  behavior.
- `226-247`: mirrors the selected-isolate and comparison helpers.

Status quo implication: generators mostly assume the module's `IuCfg` carries
the authoritative language version. If the module was built through a path that
did not set `IuCfg` from the source header, generator behavior can diverge from
the parsed version.

## Status By Version

### Ivy 1.6

Implemented or partially implemented:

- `#lang ivy1.6` is recognized by `ReadModule` and `ReadModuleFromNamedString`.
- `SetStringVersion` derives the correct `ForbidGhostInit=false` state.
- Source loading now propagates the parsed language version into both
  `IuCfg.LanguageVersion` and `IsolateCfg.IvyVersion`.
- The lexer removes 1.7-era keywords for `Version{1,6}`.
- Include selection can choose `include/1.6`.
- The standalone logic parser has a v1.6 grammar with separate `term` and
  `fmla` categories.
- Full-file parsing dispatches `Version{1,6}` to `parser_grammar_v16.y`, whose
  initial thin grammar supports include, type, individual, var, relation,
  function, axiom, property with old bare schema proofs, conjecture, old
  top-level assert, definition, top-level init, simple import/export action
  forms, standalone import/export/private call atoms, simple call actions,
  assignment actions, and the v1.6 term/formula split.
- The generated v1.6 full-file parser is patched with the same PLY-style
  lookahead behavior as the v1.7 parser for better xtrace ordering.
- Some compiler/action/theory/isolate code has v1.6 branches.
- `IvyVersionSupported` now reports `#lang ivy1.6` files as supported.

Important gaps:

- The v1.6 full-file grammar is not yet complete.
- Old v1.6 proof syntax and schema-instantiation syntax are not active full
  module grammar.
- The old v1.6 top-level assert production has not been restored in the v1.6
  grammar yet.
- Many larger module/object/mixin/scenario/proof declarations still need to be
  added to `parser_grammar_v16.y` as thin syntax rules over shared builders.

### Ivy 1.7

Implemented or strongest path:

- The full Ivy-file grammar is explicitly v1.7+.
- The logic parser default is `Version{1,7}`.
- In-memory parse helpers default to `1.7`.
- The web UI injects `#lang ivy1.7` for headerless sources.
- Compiler branches for implicit `this` isolate, action interference, and
  object-invariant preservation target `>=1.7`.
- Code generators use `>=1.7` for isolate/repl behavior.

Important gaps:

- `NewIvyUtilsConfig` defaults to `1.8`, so code paths that do not read a
  source header are not uniformly 1.7.

### Ivy 1.8

Implemented or partially implemented:

- `NewIvyUtilsConfig` defaults to `1.8`.
- Include selection can choose `include/1.8`.
- The lexer enables post-1.7 keywords such as `global`, `common`, `debug`,
  `field`, `for`, `process`, `subclass`, `whenfirst`, `whenlast`,
  `whennext`, `whenprev`, `unprovable`, and `trigger`.
- The v1.7+ full parser has productions for several post-1.7 constructs,
  including `subclass`, `global`, `common`, `debug`, `field`, `trigger`, and
  `unprovable`.
- Code generators treat `1.8` as after `1.7`, enabling test-target invariant
  compilation.

Important gaps:

- There is no distinct full-file v1.8 grammar or parser dispatch.
- Most downstream version branches distinguish only `<=1.6`, `>=1.7`, and
  `>1.7`; there is little explicit 1.8 semantic handling.
- The old plural `requires` and `ensures` tokens remain in the grammar but are
  not reserved words after `1.7`, which is probably intentional but means the
  grammar file alone is not enough to understand v1.8 syntax.

## Current High-Risk Inconsistencies

1. The initial v1.6 full-file grammar is intentionally incomplete.

   `Parse` now dispatches `<=1.6` to `ParseFullV16`, but that parser currently
   covers only the first set of core v1.6 declarations and simple actions.

2. Defaults disagree.

   `IvyUtilsConfig` defaults to `1.8`, `DefaultVersion` and `parseIvyVersion`
   default to `1.7`, `IsolateConfig` defaults to `1.7`, and the web UI inserts
   `#lang ivy1.7`.

3. Some in-memory paths still depend on the caller using the version-aware
   loader.

   `IvyFromString` and `ReadModuleFromString` now propagate the parsed version
   into config. Any other direct `Parse` callers must still thread the same
   version into their downstream compile config explicitly.

4. Isolate version state is separate and mostly unsynchronized.

   Source loading now syncs `IuCfg.LanguageVersion` and `IsolateCfg.IvyVersion`,
   but the two fields are still separate and can diverge in hand-constructed
   test modules or lower-level helper calls.

5. The v1.6 logic parser can give a false impression of full v1.6 support.

   `ParseFormula`, `ParseTerm`, and the initial full-file parser have real
   v1.6 behavior, but the full-file grammar is not complete yet.

## Design Takeaway

Current goivy is best described as:

- v1.7-oriented for the mature full source parser,
- initially v1.6-capable for full source parsing through a thin parser16
  grammar,
- version-aware for lexing and include selection,
- partially v1.6-aware in standalone logic parsing and downstream branches,
- partially v1.8-aware through lexer keywords, include directories, and some
  v1.7+ grammar productions,
- not yet a complete Python-style language mode system.

The largest architectural gap relative to Python is not the absence of a
`Version` type or include selection; those exist. The remaining gap is that
the new v1.6 grammar must be broadened while continuing to reuse the shared
post-parse framework, and downstream semantic state still needs one
authoritative language-version owner.
