# cpp2go prototype

Source-preserving C/C++-to-Go draft tooling. This version recognizes a narrow
subset of C static-inline functions; it does not yet translate C++ or produce
verified executable Go automatically. Comments and original source are retained.

From `cpp2go/`:

```sh
go run ./cmd/cpp2go -out /tmp/new-draft ../src/sat/satoko/utils/heap.h
go test -v ./...
```

The output directory must not exist. Each input produces an untouched `.original`
file and a `.go.txt` draft. `run.json` includes source hashes, original byte spans,
per-function drafts, and unresolved review issues. Directives and unsupported
code remain in the draft. Generated spans refer to per-function draft strings;
no claim is made that a formatted Go file has been mapped.

The CLI may leave partial output on failure; use a new destination for the next
run. It never overwrites an existing output directory. There are no third-party
Go dependencies. The module declares Go 1.22; this experiment was tested with Go
1.27.3. Differential tests additionally require `cc` and the surrounding ABC tree.

The executable first experiment is in `experiments/satoko/repaired/`. It is a
recorded LLM repair, not automatically generated production output. The original
generation is frozen in `draft-v1/`; `repair.py` records the fixture-specific
repair procedure and is not part of the Go translator. Do not rerun that script
over later hand edits. `scope.json`, `profile.json`, and `repairs.json` describe
what was ported, the accepted semantic choices, and issue dispositions. Unselected
functions remain unsupported even though their drafts exist.

See [PLAN.md](PLAN.md) for milestones, measurements, conclusions, and next work.

Current drafts always start with a package declaration (the first source file’s directory name by default for CLI batches,
configurable with `-package` or `Options.Package`). Function-like macros become mechanical function drafts with `int` parameter
placeholders. Their replacement text is copied directly into the function body,
with only preprocessor line continuations removed. Empty for-loop bodies are made explicit as `{}`; no returns or other
semantics are invented. `Options.Macros` can supply reviewed signature types.
The second pass repairs types, C syntax, return values, and call-site semantics.

Outside copied macro bodies, whole-expression ternaries in returns, local initializers, and simple assignments
become nested `if/else`, preserving lazy branches. Embedded ternaries remain
flagged until expression lowering can safely introduce typed temporaries. C
truth-value conversions and other unsupported syntax still require review.
`draft-v10/` demonstrates the current rules; earlier drafts remain historical baselines.

The CLI resolves the package once for the entire output directory: `-package`
if supplied, otherwise the first input's source-directory name. For example,
`vec/vec_uint.h utils/heap.h` produces `package vec` in both files; use
`-package satoko` for a shared `satoko` package. The output directory's own name
does not affect this choice. `run.json` records the resolved package. A bare
filename uses the current source directory. Invalid inferred names require an
explicit override. Library batch callers can use `ResolvePackage` once and pass
that name through `Options.Package` for each input.

Adjacent typedef/struct pairs become Go type declarations, with ordinary scalar
and pointer fields reordered to `name type`. Comments and line breaks are kept.
Complex fields remain draft text with review issues. See `struct.go` and the
heap/vector declarations in `experiments/satoko/draft-v4/`.

Unbraced `if` and `else` bodies in ordinary function drafts receive braces using
token-based statement recognition. Nested control flow and comments are retained;
macro replacement text remains verbatim. Current examples are in `draft-v10/`.

Whole-file header include guards are omitted from drafts. Their original text
remains in the `.original` source copies; unrelated conditionals are retained.

Ternary-generated if/else blocks preserve the source statement's base indentation
and indent nested branches one level deeper without requiring gofmt on drafts.

Bodyless trailing for-headers in macro fragments and for-loops with `;` bodies
receive `{}`. This is the narrow exception to verbatim macro replacement copying.

Includes and unknown standalone top-level identifier markers are commented out in
drafts. Function-body identifiers and recognized declaration continuations remain
untouched. The original source copies retain all directives and marker text.
