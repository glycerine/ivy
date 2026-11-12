# Standard-library conversion guide

This is the working map for the token-based library conversion facility. It is
an implementation guide and a review queue, not a claim that arbitrary C and Go
library calls are interchangeable. The executable catalog is `LibraryRules()`
in `stdlib_catalog.go`; tests and this document should change with the rules.

## Sources and method

The survey uses the C-library synopsis in the WG21 paper
[Synopses for the C library](https://open-std.org/JTC1/SC22/WG21/docs/papers/2016/p0175r0.html),
the implementation-authored [OpenBSD stdio documentation](https://man.openbsd.org/stdio),
and the official Go package references linked below. Direct retrieval of some
Open Group and WG14 pages was unavailable in this environment; specific behavior
was checked against accessible implementation documentation and executable tests,
rather than claiming those unavailable pages were inspected.

Important behavior checks:

- `atoi` consumes an initial numeric portion. A plain `strconv.Atoi` call on the
  entire input would reject trailing text that C accepts. See
  [OpenBSD atoi](https://man.openbsd.org/atoi) and [Go strconv](https://pkg.go.dev/strconv).
- `fgets` has a byte bound, retains a newline, and writes a NUL terminator.
  See [IBM's ISO C fgets documentation](https://www.ibm.com/docs/en/zos/3.1.0?topic=functions-fgets-read-string-from-stream).
  A Scanner loop changes both the line ending and long-line behavior; a fresh
  buffered reader per call can lose prefetched bytes. See
  [Go bufio](https://pkg.go.dev/bufio).
- File operations expose errors separately in Go. A direct name replacement
  would change assignments and condition results. See [Go os](https://pkg.go.dev/os)
  and [Go io](https://pkg.go.dev/io).
- Memory byte counts and string byte lengths cannot be replaced with element or
  rune counts. Search functions returning C pointers cannot simply become Go
  index-returning functions. See [Go bytes](https://pkg.go.dev/bytes) and
  [Go strings](https://pkg.go.dev/strings).
- C character classification is locale-dependent; the current rule policy is
  ASCII/C locale. Unicode classification is an explicit future representation
  choice, not the default. See [Go unicode](https://pkg.go.dev/unicode).
- Math conversion uses [Go math](https://pkg.go.dev/math), preserving double
  values where possible. `fmin`/`fmax` need explicit NaN handling; C `errno`,
  floating-point exception state and long-double precision are separate work.

## How the facility works

`RewriteStandardLibrary(draft)` returns a rewritten draft, `[]LibraryChange`,
and an error. `AnalyzeWithOptions` invokes it after structural/type conversion
and before opaque preprocessor islands are restored. It also updates per-function
drafts. There is no Clang invocation, hidden compatibility runtime, or helper
module to install in the generated project.

Rules have a header, candidate Go equivalent, conversion mode and semantic note.
The engine finds balanced calls with tokens, recursively rewrites arguments,
then applies a rule. Nested calls and commas in nested expressions are handled
structurally. Ordinary comments and string literals are protected. A conversion
that would discard/duplicate a comment is rejected; its call remains for review.
Qualified calls and known source-defined or shadowed names are not treated as
libc calls. The analysis is intentionally conservative and does not perform C
name lookup, overload resolution, or include expansion.

Imports are requested by successful rules only, sorted deterministically,
deduplicated, and reused through existing aliases. A conflicting source name
gets a distinct import alias. Blank imports can be upgraded when the package
becomes needed. Failed/unsupported rules do not leave unused requested imports.

The prototype-43 policy forbids introducing anonymous functions or immediately
invoked closures. Direct Go calls are emitted even when their multiple return
values do not fit the surrounding C expression: downstream users and LLMs repair
compiler errors and semantic differences. For example, printf becomes fmt.Printf,
fopen becomes os.Open/OpenFile, and atoi becomes strconv.Atoi. Calls whose current
conversion requires an adapter are retained with a review record instead. The
older rule descriptions and catalog targets below describe semantic requirements;
they do not override this policy. No adapter is inserted to preserve C status codes.

`Result.Library` records every converted or retained recognized call, its rule
message, and a source-line excerpt plus line/column/byte location in the **pre-library Go draft**. Those
locations are not original C coordinates. Tree `run.json` records converted and
review counts per file; the directory CLI also prints their aggregate summary; per-file `_starting-point/*.json` retains full details.
"Converted" means a rule was applied, not that its assumptions were verified.
Existing original-source diagnostics and `-pre` reporting are independent.

## Representation policies

- Existing mappings remain: C int -> int32; long/long long conversion routines
  assume LP64 int64; char pointers -> string; char arrays -> byte arrays;
  FILE pointers -> *os.File. Pointer widths, signed char policy and external
  ABI choices remain explicit porting decisions.
- Standard streams become os.Stdin/os.Stdout/os.Stderr. File-open failures stay
  nil, and close/seek/filesystem failures preserve C-style integer status.
  Error objects, errno and ferror/feof state require a coordinated error policy.
- `fopen` accepts recognized literal standard modes, including update, append,
  binary spellings and exclusive-create spellings. Dynamic/extended modes are
  retained. Creation uses 0666 subject to umask. Platform text translation and
  platform-specific update/append positioning require review.
- `fgets` needs a known byte buffer. It preserves n-1, newline, NUL and a partial
  final read; no bytes are prefetched and discarded. Bounds outside the known
  buffer return nil; n=1 follows the documented draft policy of a zero-length
  successful result with a terminator. These edge policies do not legitimize
  invalid C callers. Raw pointers/string destinations remain unresolved.
- `fread`/`fwrite` only convert known byte buffers; returned counts are complete
  elements, not bytes. Error and EOF state are not hidden in a replacement FILE
  runtime. Non-byte structures remain untouched.
- `fflush` is a no-op only for the selected unbuffered file representation; its
  argument is still evaluated. It is never replaced with fsync/File.Sync.
  Buffered streams, flush-all and unknown streams remain review sites.
- Known byte buffers used as C strings are sliced at their first NUL before
  string operations. String-valued inputs assume the selected C-string model;
  embedded NUL in arbitrary Go strings remains a review obligation.
- memcpy/memmove/memset/memcmp require known byte buffers. Byte bounds and
  comparison sign are preserved. No unsafe conversion of arbitrary pointers or
  guessed struct layouts is emitted.
- C numeric predicates remain numeric (0/1), preserving callers that compare
  against zero. Generic C truth-value lowering remains a separate language pass.
- Literal printf formats normalize integer length modifiers and `%i`/`%u`.
  `%n`, hex-float and wide-character formats remain for review. Dynamic formats,
  precision/count type differences and pointer/string arguments are not proven
  equivalent. Used printf/fprintf results get a single count/-1 result; ignored
  results stay direct Go calls.
- getenv -> os.Getenv is an explicitly flagged draft conversion: absence becomes
  an empty string. Use LookupEnv where C tests for NULL. exit -> os.Exit does not
  recreate atexit handlers or deferred cleanup. Assertions use panic only for
  recognized boolean comparisons; NDEBUG and abort semantics remain review work.

## Header-by-header survey

This table covers the C11 header families. A family listed for review is not
silently replaced merely because a similarly named Go operation exists.

| C header/family | Go direction | Current decision |
| --- | --- | --- |
| assert.h | shared assert helper | Condition text supplied per call; repair C truth values downstream |
| complex.h | complex64/128, math/cmplx | Review representation, component widths, functions and constants together |
| ctype.h | ASCII checks or explicitly selected unicode policy | C-locale numeric predicates and case conversion |
| errno.h | explicit error values/errors.Is | Review; no fake global errno |
| fenv.h | math and explicit numeric policy | Review; rounding/exception environment has no direct standard-Go equivalent |
| float.h | math constants and numeric widths | Review target precision and limits |
| inttypes.h | fixed-width numbers, strconv/fmt | Normalize supported literal integer printf formats; parsing/endptr review |
| iso646.h | ordinary boolean/bitwise operators | Language-token work, not a library call rule |
| limits.h | fixed-width min/max policy | Review width and signedness; never assume host C long == Go int |
| locale.h | explicit locale configuration | Review; no process-wide libc locale emulation |
| math.h | math | Common double/float math, with semantic notes; output pointers/fenv/long double reviewed |
| setjmp.h | structured error/control-flow handling | Review; no mechanical panic/recover substitution |
| signal.h | os/signal, context | Review async-signal/control-flow semantics |
| stdalign.h | representation/layout choices | Review; not an imported runtime function |
| stdarg.h | variadic Go signatures and slices | Review va_list ABI and caller/callee together |
| stdatomic.h | sync/atomic | Review width, alignment, memory order and surrounding synchronization |
| stdbool.h | bool plus C truth-value conversion | Language/type pass; numeric predicate results retained |
| stddef.h | nil, size/index/difference types | Type/layout pass; sizeof/offsetof need representation decisions |
| stdint.h | int8…int64, uint8…uint64 | Type pass; exact-width/fast/least/intptr distinctions require policy |
| stdio.h | os, io, fmt, coordinated bufio | Implemented safe call shapes; retain buffered-state/scanning/position objects |
| stdlib.h | strconv, make/new/GC, os, sort/slices | Prefix integer conversions, abs, exit/getenv; allocation/shell/sort/RNG reviewed |
| stdnoreturn.h | explicit control-flow | Review; Go type checker does not infer noreturn behavior |
| string.h | strings, bytes, copy | Selected string/value and byte-buffer rules; pointer-returning/stateful APIs reviewed |
| tgmath.h | explicit numeric types/functions | Review type-generic dispatch |
| threads.h | goroutines, sync, channels | Review coordinated ownership, synchronization and lifecycle |
| time.h | time | Unix-second time queries and pointer stores; wall-nanosecond clock policy; calendar/struct tm operations reviewed |
| uchar.h | byte/rune conversions and encodings | Review code units, state, malformed input and locale |
| wchar.h | rune/string/encoding policy | Review wide-character width and stateful conversions |
| wctype.h | explicitly selected unicode/locale | Review; not a silent unicode substitution |

C23 additions, vendor extensions and the C++ standard library are not claimed as
covered by this C11-oriented survey. POSIX strdup is explicitly labeled as an
extension. POSIX pthread/system integration in starter.c remains separate work.

## Validation and remaining work

The runtime fixtures compile and execute generated Go, not just compare strings.
They cover prefix integers, sign/whitespace, argument single evaluation, nested
calls, import aliases, ASCII EOF handling, math, string comparisons, byte memory,
file modes, bounded/repeated line reads, append, EOF and file positioning.
A C-compiler differential test translates a real C consumer end-to-end and
compares output for empty, long, newline-only, unterminated-final and embedded-NUL
inputs. It skips only if no C compiler is installed; normal translation requires
no C compiler. The full ABC sweep checks analysis/fallback regressions and
metadata coverage, not compilation or semantic equivalence of all ABC.

Priority follow-ups are coordinated buffered I/O and error state; typed allocation
and ownership; strto* end pointers and atof's full grammar/locale; safe pointer
search-result lowering; qsort/bsearch comparators; scanner formats and bounded
output; and CPU-time/random-number reproducibility. None are hidden behind a
misleading direct call rename.

## Executable rule catalog

The following inventory is generated from LibraryRules(). `convert` means an
implemented rule with stated assumptions; `conditional` additionally requires
specific argument shapes/types; `review` leaves the call visible for repair.

| C name | Header | Mode | Go direction | Review notes |
| --- | --- | --- | --- | --- |
| strdup | POSIX string.h | conditional | string value | Known strings only; allocation failure and pointer identity are not retained. |
| assert | assert.h | convert | assert(condition, conditionText) | Shared package helper; repair C truth values downstream. NDEBUG/abort semantics differ. |
| isalnum | ctype.h | convert | ASCII byte/range checks | C locale only; preserve numeric predicate result and EOF for case conversion, do not silently use Unicode semantics. |
| isalpha | ctype.h | convert | ASCII byte/range checks | C locale only; preserve numeric predicate result and EOF for case conversion, do not silently use Unicode semantics. |
| isblank | ctype.h | convert | ASCII byte/range checks | C locale only; preserve numeric predicate result and EOF for case conversion, do not silently use Unicode semantics. |
| iscntrl | ctype.h | convert | ASCII byte/range checks | C locale only; preserve numeric predicate result and EOF for case conversion, do not silently use Unicode semantics. |
| isdigit | ctype.h | convert | ASCII byte/range checks | C locale only; preserve numeric predicate result and EOF for case conversion, do not silently use Unicode semantics. |
| isgraph | ctype.h | convert | ASCII byte/range checks | C locale only; preserve numeric predicate result and EOF for case conversion, do not silently use Unicode semantics. |
| islower | ctype.h | convert | ASCII byte/range checks | C locale only; preserve numeric predicate result and EOF for case conversion, do not silently use Unicode semantics. |
| isprint | ctype.h | convert | ASCII byte/range checks | C locale only; preserve numeric predicate result and EOF for case conversion, do not silently use Unicode semantics. |
| ispunct | ctype.h | convert | ASCII byte/range checks | C locale only; preserve numeric predicate result and EOF for case conversion, do not silently use Unicode semantics. |
| isspace | ctype.h | convert | ASCII byte/range checks | C locale only; preserve numeric predicate result and EOF for case conversion, do not silently use Unicode semantics. |
| isupper | ctype.h | convert | ASCII byte/range checks | C locale only; preserve numeric predicate result and EOF for case conversion, do not silently use Unicode semantics. |
| isxdigit | ctype.h | convert | ASCII byte/range checks | C locale only; preserve numeric predicate result and EOF for case conversion, do not silently use Unicode semantics. |
| tolower | ctype.h | convert | ASCII byte/range checks | C locale only; preserve numeric predicate result and EOF for case conversion, do not silently use Unicode semantics. |
| toupper | ctype.h | convert | ASCII byte/range checks | C locale only; preserve numeric predicate result and EOF for case conversion, do not silently use Unicode semantics. |
| localeconv | locale.h | review | explicit locale policy | No process-global libc locale equivalent in Go's standard library. |
| setlocale | locale.h | review | explicit locale policy | No process-global libc locale equivalent in Go's standard library. |
| acos | math.h | convert | math package | double draft; errno/fenv and exceptional-value details need review. |
| acosf | math.h | convert | float32(math operation) | Float inputs are widened for Go math, then rounded to float32; inspect exact rounding/exception requirements. |
| acosh | math.h | convert | math package | double draft; errno/fenv and exceptional-value details need review. |
| acoshf | math.h | convert | float32(math operation) | Float inputs are widened for Go math, then rounded to float32; inspect exact rounding/exception requirements. |
| acoshl | math.h | review | explicit long-double precision policy | Go has no native long double; do not silently narrow precision. |
| acosl | math.h | review | explicit long-double precision policy | Go has no native long double; do not silently narrow precision. |
| asin | math.h | convert | math package | double draft; errno/fenv and exceptional-value details need review. |
| asinf | math.h | convert | float32(math operation) | Float inputs are widened for Go math, then rounded to float32; inspect exact rounding/exception requirements. |
| asinh | math.h | convert | math package | double draft; errno/fenv and exceptional-value details need review. |
| asinhf | math.h | convert | float32(math operation) | Float inputs are widened for Go math, then rounded to float32; inspect exact rounding/exception requirements. |
| asinhl | math.h | review | explicit long-double precision policy | Go has no native long double; do not silently narrow precision. |
| asinl | math.h | review | explicit long-double precision policy | Go has no native long double; do not silently narrow precision. |
| atan | math.h | convert | math package | double draft; errno/fenv and exceptional-value details need review. |
| atan2 | math.h | convert | math package | double draft; errno/fenv and exceptional-value details need review. |
| atan2f | math.h | convert | float32(math operation) | Float inputs are widened for Go math, then rounded to float32; inspect exact rounding/exception requirements. |
| atan2l | math.h | review | explicit long-double precision policy | Go has no native long double; do not silently narrow precision. |
| atanf | math.h | convert | float32(math operation) | Float inputs are widened for Go math, then rounded to float32; inspect exact rounding/exception requirements. |
| atanh | math.h | convert | math package | double draft; errno/fenv and exceptional-value details need review. |
| atanhf | math.h | convert | float32(math operation) | Float inputs are widened for Go math, then rounded to float32; inspect exact rounding/exception requirements. |
| atanhl | math.h | review | explicit long-double precision policy | Go has no native long double; do not silently narrow precision. |
| atanl | math.h | review | explicit long-double precision policy | Go has no native long double; do not silently narrow precision. |
| cbrt | math.h | convert | math package | double draft; errno/fenv and exceptional-value details need review. |
| cbrtf | math.h | convert | float32(math operation) | Float inputs are widened for Go math, then rounded to float32; inspect exact rounding/exception requirements. |
| cbrtl | math.h | review | explicit long-double precision policy | Go has no native long double; do not silently narrow precision. |
| ceil | math.h | convert | math package | double draft; errno/fenv and exceptional-value details need review. |
| ceilf | math.h | convert | float32(math operation) | Float inputs are widened for Go math, then rounded to float32; inspect exact rounding/exception requirements. |
| ceill | math.h | review | explicit long-double precision policy | Go has no native long double; do not silently narrow precision. |
| copysign | math.h | convert | math package | double draft; errno/fenv and exceptional-value details need review. |
| copysignf | math.h | convert | float32(math operation) | Float inputs are widened for Go math, then rounded to float32; inspect exact rounding/exception requirements. |
| copysignl | math.h | review | explicit long-double precision policy | Go has no native long double; do not silently narrow precision. |
| cos | math.h | convert | math package | double draft; errno/fenv and exceptional-value details need review. |
| cosf | math.h | convert | float32(math operation) | Float inputs are widened for Go math, then rounded to float32; inspect exact rounding/exception requirements. |
| cosh | math.h | convert | math package | double draft; errno/fenv and exceptional-value details need review. |
| coshf | math.h | convert | float32(math operation) | Float inputs are widened for Go math, then rounded to float32; inspect exact rounding/exception requirements. |
| coshl | math.h | review | explicit long-double precision policy | Go has no native long double; do not silently narrow precision. |
| cosl | math.h | review | explicit long-double precision policy | Go has no native long double; do not silently narrow precision. |
| erf | math.h | convert | math package | double draft; errno/fenv and exceptional-value details need review. |
| erfc | math.h | convert | math package | double draft; errno/fenv and exceptional-value details need review. |
| erfcf | math.h | convert | float32(math operation) | Float inputs are widened for Go math, then rounded to float32; inspect exact rounding/exception requirements. |
| erfcl | math.h | review | explicit long-double precision policy | Go has no native long double; do not silently narrow precision. |
| erff | math.h | convert | float32(math operation) | Float inputs are widened for Go math, then rounded to float32; inspect exact rounding/exception requirements. |
| erfl | math.h | review | explicit long-double precision policy | Go has no native long double; do not silently narrow precision. |
| exp | math.h | convert | math package | double draft; errno/fenv and exceptional-value details need review. |
| exp2 | math.h | convert | math package | double draft; errno/fenv and exceptional-value details need review. |
| exp2f | math.h | convert | float32(math operation) | Float inputs are widened for Go math, then rounded to float32; inspect exact rounding/exception requirements. |
| exp2l | math.h | review | explicit long-double precision policy | Go has no native long double; do not silently narrow precision. |
| expf | math.h | convert | float32(math operation) | Float inputs are widened for Go math, then rounded to float32; inspect exact rounding/exception requirements. |
| expl | math.h | review | explicit long-double precision policy | Go has no native long double; do not silently narrow precision. |
| expm1 | math.h | convert | math package | double draft; errno/fenv and exceptional-value details need review. |
| expm1f | math.h | convert | float32(math operation) | Float inputs are widened for Go math, then rounded to float32; inspect exact rounding/exception requirements. |
| expm1l | math.h | review | explicit long-double precision policy | Go has no native long double; do not silently narrow precision. |
| fabs | math.h | convert | math package | double draft; errno/fenv and exceptional-value details need review. |
| fabsf | math.h | convert | float32(math operation) | Float inputs are widened for Go math, then rounded to float32; inspect exact rounding/exception requirements. |
| fabsl | math.h | review | explicit long-double precision policy | Go has no native long double; do not silently narrow precision. |
| fdim | math.h | convert | math package | double draft; errno/fenv and exceptional-value details need review. |
| fdimf | math.h | convert | float32(math operation) | Float inputs are widened for Go math, then rounded to float32; inspect exact rounding/exception requirements. |
| fdiml | math.h | review | explicit long-double precision policy | Go has no native long double; do not silently narrow precision. |
| floor | math.h | convert | math package | double draft; errno/fenv and exceptional-value details need review. |
| floorf | math.h | convert | float32(math operation) | Float inputs are widened for Go math, then rounded to float32; inspect exact rounding/exception requirements. |
| floorl | math.h | review | explicit long-double precision policy | Go has no native long double; do not silently narrow precision. |
| fmax | math.h | convert | math package | double draft; errno/fenv and exceptional-value details need review. |
| fmaxf | math.h | convert | float32(math operation) | Float inputs are widened for Go math, then rounded to float32; inspect exact rounding/exception requirements. |
| fmaxl | math.h | review | explicit long-double precision policy | Go has no native long double; do not silently narrow precision. |
| fmin | math.h | convert | math package | double draft; errno/fenv and exceptional-value details need review. |
| fminf | math.h | convert | float32(math operation) | Float inputs are widened for Go math, then rounded to float32; inspect exact rounding/exception requirements. |
| fminl | math.h | review | explicit long-double precision policy | Go has no native long double; do not silently narrow precision. |
| fmod | math.h | convert | math package | double draft; errno/fenv and exceptional-value details need review. |
| fmodf | math.h | convert | float32(math operation) | Float inputs are widened for Go math, then rounded to float32; inspect exact rounding/exception requirements. |
| fmodl | math.h | review | explicit long-double precision policy | Go has no native long double; do not silently narrow precision. |
| fpclassify | math.h | review | math with result/representation adaptation | Pointer outputs, integer exponents, predicates, rounding modes or NaN details need special rules. |
| frexp | math.h | review | math with result/representation adaptation | Pointer outputs, integer exponents, predicates, rounding modes or NaN details need special rules. |
| hypot | math.h | convert | math package | double draft; errno/fenv and exceptional-value details need review. |
| hypotf | math.h | convert | float32(math operation) | Float inputs are widened for Go math, then rounded to float32; inspect exact rounding/exception requirements. |
| hypotl | math.h | review | explicit long-double precision policy | Go has no native long double; do not silently narrow precision. |
| ilogb | math.h | review | math with result/representation adaptation | Pointer outputs, integer exponents, predicates, rounding modes or NaN details need special rules. |
| isfinite | math.h | review | math with result/representation adaptation | Pointer outputs, integer exponents, predicates, rounding modes or NaN details need special rules. |
| isinf | math.h | review | math with result/representation adaptation | Pointer outputs, integer exponents, predicates, rounding modes or NaN details need special rules. |
| isnan | math.h | review | math with result/representation adaptation | Pointer outputs, integer exponents, predicates, rounding modes or NaN details need special rules. |
| isnormal | math.h | review | math with result/representation adaptation | Pointer outputs, integer exponents, predicates, rounding modes or NaN details need special rules. |
| ldexp | math.h | review | math with result/representation adaptation | Pointer outputs, integer exponents, predicates, rounding modes or NaN details need special rules. |
| lgamma | math.h | convert | math package | double draft; errno/fenv and exceptional-value details need review. |
| lgammaf | math.h | convert | float32(math operation) | Float inputs are widened for Go math, then rounded to float32; inspect exact rounding/exception requirements. |
| lgammal | math.h | review | explicit long-double precision policy | Go has no native long double; do not silently narrow precision. |
| llrint | math.h | review | math with result/representation adaptation | Pointer outputs, integer exponents, predicates, rounding modes or NaN details need special rules. |
| llround | math.h | review | math with result/representation adaptation | Pointer outputs, integer exponents, predicates, rounding modes or NaN details need special rules. |
| log | math.h | convert | math package | double draft; errno/fenv and exceptional-value details need review. |
| log10 | math.h | convert | math package | double draft; errno/fenv and exceptional-value details need review. |
| log10f | math.h | convert | float32(math operation) | Float inputs are widened for Go math, then rounded to float32; inspect exact rounding/exception requirements. |
| log10l | math.h | review | explicit long-double precision policy | Go has no native long double; do not silently narrow precision. |
| log1p | math.h | convert | math package | double draft; errno/fenv and exceptional-value details need review. |
| log1pf | math.h | convert | float32(math operation) | Float inputs are widened for Go math, then rounded to float32; inspect exact rounding/exception requirements. |
| log1pl | math.h | review | explicit long-double precision policy | Go has no native long double; do not silently narrow precision. |
| log2 | math.h | convert | math package | double draft; errno/fenv and exceptional-value details need review. |
| log2f | math.h | convert | float32(math operation) | Float inputs are widened for Go math, then rounded to float32; inspect exact rounding/exception requirements. |
| log2l | math.h | review | explicit long-double precision policy | Go has no native long double; do not silently narrow precision. |
| logb | math.h | review | math with result/representation adaptation | Pointer outputs, integer exponents, predicates, rounding modes or NaN details need special rules. |
| logf | math.h | convert | float32(math operation) | Float inputs are widened for Go math, then rounded to float32; inspect exact rounding/exception requirements. |
| logl | math.h | review | explicit long-double precision policy | Go has no native long double; do not silently narrow precision. |
| lrint | math.h | review | math with result/representation adaptation | Pointer outputs, integer exponents, predicates, rounding modes or NaN details need special rules. |
| lround | math.h | review | math with result/representation adaptation | Pointer outputs, integer exponents, predicates, rounding modes or NaN details need special rules. |
| modf | math.h | review | math with result/representation adaptation | Pointer outputs, integer exponents, predicates, rounding modes or NaN details need special rules. |
| nan | math.h | review | math with result/representation adaptation | Pointer outputs, integer exponents, predicates, rounding modes or NaN details need special rules. |
| nearbyint | math.h | review | math with result/representation adaptation | Pointer outputs, integer exponents, predicates, rounding modes or NaN details need special rules. |
| nextafter | math.h | review | math with result/representation adaptation | Pointer outputs, integer exponents, predicates, rounding modes or NaN details need special rules. |
| nexttoward | math.h | review | math with result/representation adaptation | Pointer outputs, integer exponents, predicates, rounding modes or NaN details need special rules. |
| pow | math.h | convert | math package | double draft; errno/fenv and exceptional-value details need review. |
| powf | math.h | convert | float32(math operation) | Float inputs are widened for Go math, then rounded to float32; inspect exact rounding/exception requirements. |
| powl | math.h | review | explicit long-double precision policy | Go has no native long double; do not silently narrow precision. |
| remainder | math.h | convert | math package | double draft; errno/fenv and exceptional-value details need review. |
| remainderf | math.h | convert | float32(math operation) | Float inputs are widened for Go math, then rounded to float32; inspect exact rounding/exception requirements. |
| remainderl | math.h | review | explicit long-double precision policy | Go has no native long double; do not silently narrow precision. |
| remquo | math.h | review | math with result/representation adaptation | Pointer outputs, integer exponents, predicates, rounding modes or NaN details need special rules. |
| rint | math.h | review | math with result/representation adaptation | Pointer outputs, integer exponents, predicates, rounding modes or NaN details need special rules. |
| round | math.h | convert | math package | double draft; errno/fenv and exceptional-value details need review. |
| roundf | math.h | convert | float32(math operation) | Float inputs are widened for Go math, then rounded to float32; inspect exact rounding/exception requirements. |
| roundl | math.h | review | explicit long-double precision policy | Go has no native long double; do not silently narrow precision. |
| scalbln | math.h | review | math with result/representation adaptation | Pointer outputs, integer exponents, predicates, rounding modes or NaN details need special rules. |
| scalbn | math.h | review | math with result/representation adaptation | Pointer outputs, integer exponents, predicates, rounding modes or NaN details need special rules. |
| signbit | math.h | review | math with result/representation adaptation | Pointer outputs, integer exponents, predicates, rounding modes or NaN details need special rules. |
| sin | math.h | convert | math package | double draft; errno/fenv and exceptional-value details need review. |
| sinf | math.h | convert | float32(math operation) | Float inputs are widened for Go math, then rounded to float32; inspect exact rounding/exception requirements. |
| sinh | math.h | convert | math package | double draft; errno/fenv and exceptional-value details need review. |
| sinhf | math.h | convert | float32(math operation) | Float inputs are widened for Go math, then rounded to float32; inspect exact rounding/exception requirements. |
| sinhl | math.h | review | explicit long-double precision policy | Go has no native long double; do not silently narrow precision. |
| sinl | math.h | review | explicit long-double precision policy | Go has no native long double; do not silently narrow precision. |
| sqrt | math.h | convert | math package | double draft; errno/fenv and exceptional-value details need review. |
| sqrtf | math.h | convert | float32(math operation) | Float inputs are widened for Go math, then rounded to float32; inspect exact rounding/exception requirements. |
| sqrtl | math.h | review | explicit long-double precision policy | Go has no native long double; do not silently narrow precision. |
| tan | math.h | convert | math package | double draft; errno/fenv and exceptional-value details need review. |
| tanf | math.h | convert | float32(math operation) | Float inputs are widened for Go math, then rounded to float32; inspect exact rounding/exception requirements. |
| tanh | math.h | convert | math package | double draft; errno/fenv and exceptional-value details need review. |
| tanhf | math.h | convert | float32(math operation) | Float inputs are widened for Go math, then rounded to float32; inspect exact rounding/exception requirements. |
| tanhl | math.h | review | explicit long-double precision policy | Go has no native long double; do not silently narrow precision. |
| tanl | math.h | review | explicit long-double precision policy | Go has no native long double; do not silently narrow precision. |
| tgamma | math.h | convert | math package | double draft; errno/fenv and exceptional-value details need review. |
| tgammaf | math.h | convert | float32(math operation) | Float inputs are widened for Go math, then rounded to float32; inspect exact rounding/exception requirements. |
| tgammal | math.h | review | explicit long-double precision policy | Go has no native long double; do not silently narrow precision. |
| trunc | math.h | convert | math package | double draft; errno/fenv and exceptional-value details need review. |
| truncf | math.h | convert | float32(math operation) | Float inputs are widened for Go math, then rounded to float32; inspect exact rounding/exception requirements. |
| truncl | math.h | review | explicit long-double precision policy | Go has no native long double; do not silently narrow precision. |
| longjmp | setjmp.h/signal.h | review | structured errors / os/signal | Control-flow and asynchronous signal semantics require redesign. |
| raise | setjmp.h/signal.h | review | structured errors / os/signal | Control-flow and asynchronous signal semantics require redesign. |
| setjmp | setjmp.h/signal.h | review | structured errors / os/signal | Control-flow and asynchronous signal semantics require redesign. |
| signal | setjmp.h/signal.h | review | structured errors / os/signal | Control-flow and asynchronous signal semantics require redesign. |
| va_arg | stdarg.h | review | variadic parameters | C va_list ABI has no direct Go representation. |
| va_copy | stdarg.h | review | variadic parameters | C va_list ABI has no direct Go representation. |
| va_end | stdarg.h | review | variadic parameters | C va_list ABI has no direct Go representation. |
| va_start | stdarg.h | review | variadic parameters | C va_list ABI has no direct Go representation. |
| EOF | stdio.h | convert | os standard streams / -1 | Stream objects and the selected conventional EOF sentinel; source overrides are protected. |
| clearerr | stdio.h | review | fmt scanning / buffered I/O / explicit errors | Result counts, formats, bounds, state, positioning or lifetime require structural repair. |
| fclose | stdio.h | convert | File.Close / os.Remove / os.Rename | Preserve C success/failure integer status; errno is not reproduced. |
| feof | stdio.h | review | fmt scanning / buffered I/O / explicit errors | Result counts, formats, bounds, state, positioning or lifetime require structural repair. |
| ferror | stdio.h | review | fmt scanning / buffered I/O / explicit errors | Result counts, formats, bounds, state, positioning or lifetime require structural repair. |
| fflush | stdio.h | conditional | no-op for unbuffered os.File | Only known file identifiers/std streams; not File.Sync (different semantics). Retained otherwise. |
| fgetc | stdio.h | convert | File.Read / File.Write | Single-byte transfer, unsigned byte value or -1. EOF/error distinction still needs explicit handling. |
| fgetpos | stdio.h | review | fmt scanning / buffered I/O / explicit errors | Result counts, formats, bounds, state, positioning or lifetime require structural repair. |
| fgets | stdio.h | conditional | bounded File.Read loop | Known byte array/slice only; preserve n-1, newline, terminator and EOF. No read-ahead buffer is discarded. |
| fopen | stdio.h | conditional | os.Open / os.OpenFile | Literal standard modes; error maps to nil. Text/binary translation and errno need review. |
| fprintf | stdio.h | conditional | fmt.Printf / fmt.Fprintf / fmt.Sprintf | C format verbs, result arity and mutable destinations need review; existing draft rules retained. |
| fputc | stdio.h | convert | File.Read / File.Write | Single-byte transfer, unsigned byte value or -1. EOF/error distinction still needs explicit handling. |
| fputs | stdio.h | convert | fmt.Fprintln / io.WriteString | C nonnegative success / negative failure, not Go's two-result API. |
| fread | stdio.h | conditional | io.ReadFull / File.Write | Known byte buffers; byte count converted back to complete element count. |
| freopen | stdio.h | review | fmt scanning / buffered I/O / explicit errors | Result counts, formats, bounds, state, positioning or lifetime require structural repair. |
| fscanf | stdio.h | review | fmt scanning / buffered I/O / explicit errors | Result counts, formats, bounds, state, positioning or lifetime require structural repair. |
| fseek | stdio.h | conditional | File.Seek | Known SEEK constants; preserve status/position. Stream error flags are not tracked. |
| fsetpos | stdio.h | review | fmt scanning / buffered I/O / explicit errors | Result counts, formats, bounds, state, positioning or lifetime require structural repair. |
| ftell | stdio.h | conditional | File.Seek | Known SEEK constants; preserve status/position. Stream error flags are not tracked. |
| fwrite | stdio.h | conditional | io.ReadFull / File.Write | Known byte buffers; byte count converted back to complete element count. |
| getc | stdio.h | convert | File.Read / File.Write | Single-byte transfer, unsigned byte value or -1. EOF/error distinction still needs explicit handling. |
| getchar | stdio.h | convert | File.Read / File.Write | Single-byte transfer, unsigned byte value or -1. EOF/error distinction still needs explicit handling. |
| gets | stdio.h | review | fmt scanning / buffered I/O / explicit errors | Result counts, formats, bounds, state, positioning or lifetime require structural repair. |
| perror | stdio.h | review | fmt scanning / buffered I/O / explicit errors | Result counts, formats, bounds, state, positioning or lifetime require structural repair. |
| printf | stdio.h | conditional | fmt.Printf / fmt.Fprintf / fmt.Sprintf | C format verbs, result arity and mutable destinations need review; existing draft rules retained. |
| putc | stdio.h | convert | File.Read / File.Write | Single-byte transfer, unsigned byte value or -1. EOF/error distinction still needs explicit handling. |
| putchar | stdio.h | convert | File.Read / File.Write | Single-byte transfer, unsigned byte value or -1. EOF/error distinction still needs explicit handling. |
| puts | stdio.h | convert | fmt.Fprintln / io.WriteString | C nonnegative success / negative failure, not Go's two-result API. |
| remove | stdio.h | convert | File.Close / os.Remove / os.Rename | Preserve C success/failure integer status; errno is not reproduced. |
| rename | stdio.h | convert | File.Close / os.Remove / os.Rename | Preserve C success/failure integer status; errno is not reproduced. |
| rewind | stdio.h | conditional | File.Seek | Known SEEK constants; preserve status/position. Stream error flags are not tracked. |
| scanf | stdio.h | review | fmt scanning / buffered I/O / explicit errors | Result counts, formats, bounds, state, positioning or lifetime require structural repair. |
| setbuf | stdio.h | review | fmt scanning / buffered I/O / explicit errors | Result counts, formats, bounds, state, positioning or lifetime require structural repair. |
| setvbuf | stdio.h | review | fmt scanning / buffered I/O / explicit errors | Result counts, formats, bounds, state, positioning or lifetime require structural repair. |
| snprintf | stdio.h | review | fmt scanning / buffered I/O / explicit errors | Result counts, formats, bounds, state, positioning or lifetime require structural repair. |
| sprintf | stdio.h | conditional | fmt.Printf / fmt.Fprintf / fmt.Sprintf | C format verbs, result arity and mutable destinations need review; existing draft rules retained. |
| sscanf | stdio.h | review | fmt scanning / buffered I/O / explicit errors | Result counts, formats, bounds, state, positioning or lifetime require structural repair. |
| stderr | stdio.h | convert | os standard streams / -1 | Stream objects and the selected conventional EOF sentinel; source overrides are protected. |
| stdin | stdio.h | convert | os standard streams / -1 | Stream objects and the selected conventional EOF sentinel; source overrides are protected. |
| stdout | stdio.h | convert | os standard streams / -1 | Stream objects and the selected conventional EOF sentinel; source overrides are protected. |
| tmpfile | stdio.h | review | fmt scanning / buffered I/O / explicit errors | Result counts, formats, bounds, state, positioning or lifetime require structural repair. |
| tmpnam | stdio.h | review | fmt scanning / buffered I/O / explicit errors | Result counts, formats, bounds, state, positioning or lifetime require structural repair. |
| ungetc | stdio.h | review | fmt scanning / buffered I/O / explicit errors | Result counts, formats, bounds, state, positioning or lifetime require structural repair. |
| vfprintf | stdio.h | review | fmt scanning / buffered I/O / explicit errors | Result counts, formats, bounds, state, positioning or lifetime require structural repair. |
| vfscanf | stdio.h | review | fmt scanning / buffered I/O / explicit errors | Result counts, formats, bounds, state, positioning or lifetime require structural repair. |
| vprintf | stdio.h | review | fmt scanning / buffered I/O / explicit errors | Result counts, formats, bounds, state, positioning or lifetime require structural repair. |
| vscanf | stdio.h | review | fmt scanning / buffered I/O / explicit errors | Result counts, formats, bounds, state, positioning or lifetime require structural repair. |
| vsnprintf | stdio.h | review | fmt scanning / buffered I/O / explicit errors | Result counts, formats, bounds, state, positioning or lifetime require structural repair. |
| vsprintf | stdio.h | review | fmt scanning / buffered I/O / explicit errors | Result counts, formats, bounds, state, positioning or lifetime require structural repair. |
| vsscanf | stdio.h | review | fmt scanning / buffered I/O / explicit errors | Result counts, formats, bounds, state, positioning or lifetime require structural repair. |
| EXIT_FAILURE | stdlib.h | convert | 0 / 1 | Conventional process status constants; source overrides are protected. |
| EXIT_SUCCESS | stdlib.h | convert | 0 / 1 | Conventional process status constants; source overrides are protected. |
| _Exit | stdlib.h | convert | os.Exit | C atexit handlers and buffered flushing are not replicated. |
| abort | stdlib.h | review | math/rand / slices / sort / os/exec / encoding | Algorithms, ordering, shell semantics, signals, cleanup and locale are not interchangeable by name. |
| abs | stdlib.h | convert | typed sign check | Integer widths retained; minimum signed value remains an overflow case. |
| aligned_alloc | stdlib.h | review | make / new / GC | Ownership, element type, byte size, pointer arithmetic and failure policy must be established. |
| at_quick_exit | stdlib.h | review | math/rand / slices / sort / os/exec / encoding | Algorithms, ordering, shell semantics, signals, cleanup and locale are not interchangeable by name. |
| atexit | stdlib.h | review | math/rand / slices / sort / os/exec / encoding | Algorithms, ordering, shell semantics, signals, cleanup and locale are not interchangeable by name. |
| atof | stdlib.h | review | strconv.ParseFloat / ParseInt / ParseUint | Prefix consumption, endptr, locale, bases, overflow and errno need explicit decisions. |
| atoi | stdlib.h | convert | strconv.ParseInt with decimal-prefix scan | ASCII C-locale whitespace/sign/prefix; int32 or LP64 int64. C overflow is outside the defined domain. |
| atol | stdlib.h | convert | strconv.ParseInt with decimal-prefix scan | ASCII C-locale whitespace/sign/prefix; int32 or LP64 int64. C overflow is outside the defined domain. |
| atoll | stdlib.h | convert | strconv.ParseInt with decimal-prefix scan | ASCII C-locale whitespace/sign/prefix; int32 or LP64 int64. C overflow is outside the defined domain. |
| bsearch | stdlib.h | review | math/rand / slices / sort / os/exec / encoding | Algorithms, ordering, shell semantics, signals, cleanup and locale are not interchangeable by name. |
| calloc | stdlib.h | review | make / new / GC | Ownership, element type, byte size, pointer arithmetic and failure policy must be established. |
| div | stdlib.h | review | math/rand / slices / sort / os/exec / encoding | Algorithms, ordering, shell semantics, signals, cleanup and locale are not interchangeable by name. |
| exit | stdlib.h | convert | os.Exit | C atexit handlers and buffered flushing are not replicated. |
| free | stdlib.h | review | make / new / GC | Ownership, element type, byte size, pointer arithmetic and failure policy must be established. |
| getenv | stdlib.h | conditional | os.Getenv | Go strings cannot preserve NULL vs empty; converted with a review record. |
| labs | stdlib.h | convert | typed sign check | Integer widths retained; minimum signed value remains an overflow case. |
| ldiv | stdlib.h | review | math/rand / slices / sort / os/exec / encoding | Algorithms, ordering, shell semantics, signals, cleanup and locale are not interchangeable by name. |
| llabs | stdlib.h | convert | typed sign check | Integer widths retained; minimum signed value remains an overflow case. |
| lldiv | stdlib.h | review | math/rand / slices / sort / os/exec / encoding | Algorithms, ordering, shell semantics, signals, cleanup and locale are not interchangeable by name. |
| malloc | stdlib.h | review | make / new / GC | Ownership, element type, byte size, pointer arithmetic and failure policy must be established. |
| mblen | stdlib.h | review | math/rand / slices / sort / os/exec / encoding | Algorithms, ordering, shell semantics, signals, cleanup and locale are not interchangeable by name. |
| mbstowcs | stdlib.h | review | math/rand / slices / sort / os/exec / encoding | Algorithms, ordering, shell semantics, signals, cleanup and locale are not interchangeable by name. |
| mbtowc | stdlib.h | review | math/rand / slices / sort / os/exec / encoding | Algorithms, ordering, shell semantics, signals, cleanup and locale are not interchangeable by name. |
| qsort | stdlib.h | review | math/rand / slices / sort / os/exec / encoding | Algorithms, ordering, shell semantics, signals, cleanup and locale are not interchangeable by name. |
| quick_exit | stdlib.h | review | math/rand / slices / sort / os/exec / encoding | Algorithms, ordering, shell semantics, signals, cleanup and locale are not interchangeable by name. |
| rand | stdlib.h | review | math/rand / slices / sort / os/exec / encoding | Algorithms, ordering, shell semantics, signals, cleanup and locale are not interchangeable by name. |
| realloc | stdlib.h | review | make / new / GC | Ownership, element type, byte size, pointer arithmetic and failure policy must be established. |
| srand | stdlib.h | review | math/rand / slices / sort / os/exec / encoding | Algorithms, ordering, shell semantics, signals, cleanup and locale are not interchangeable by name. |
| strtod | stdlib.h | review | strconv.ParseFloat / ParseInt / ParseUint | Prefix consumption, endptr, locale, bases, overflow and errno need explicit decisions. |
| strtof | stdlib.h | review | strconv.ParseFloat / ParseInt / ParseUint | Prefix consumption, endptr, locale, bases, overflow and errno need explicit decisions. |
| strtol | stdlib.h | review | strconv.ParseFloat / ParseInt / ParseUint | Prefix consumption, endptr, locale, bases, overflow and errno need explicit decisions. |
| strtold | stdlib.h | review | strconv.ParseFloat / ParseInt / ParseUint | Prefix consumption, endptr, locale, bases, overflow and errno need explicit decisions. |
| strtoll | stdlib.h | review | strconv.ParseFloat / ParseInt / ParseUint | Prefix consumption, endptr, locale, bases, overflow and errno need explicit decisions. |
| strtoul | stdlib.h | review | strconv.ParseFloat / ParseInt / ParseUint | Prefix consumption, endptr, locale, bases, overflow and errno need explicit decisions. |
| strtoull | stdlib.h | review | strconv.ParseFloat / ParseInt / ParseUint | Prefix consumption, endptr, locale, bases, overflow and errno need explicit decisions. |
| system | stdlib.h | review | math/rand / slices / sort / os/exec / encoding | Algorithms, ordering, shell semantics, signals, cleanup and locale are not interchangeable by name. |
| wcstombs | stdlib.h | review | math/rand / slices / sort / os/exec / encoding | Algorithms, ordering, shell semantics, signals, cleanup and locale are not interchangeable by name. |
| wctomb | stdlib.h | review | math/rand / slices / sort / os/exec / encoding | Algorithms, ordering, shell semantics, signals, cleanup and locale are not interchangeable by name. |
| memchr | string.h | review | strings/bytes search, copy or explicit errors | Pointer results, NULL checks, padding, shared token state and locale require surrounding-code repair. |
| memcmp | string.h | conditional | copy / byte fill / bytes.Compare | Known byte buffers only; no guessing sizeof, raw pointers or element sizes. |
| memcpy | string.h | conditional | copy / byte fill / bytes.Compare | Known byte buffers only; no guessing sizeof, raw pointers or element sizes. |
| memmove | string.h | conditional | copy / byte fill / bytes.Compare | Known byte buffers only; no guessing sizeof, raw pointers or element sizes. |
| memset | string.h | conditional | copy / byte fill / bytes.Compare | Known byte buffers only; no guessing sizeof, raw pointers or element sizes. |
| strcat | string.h | conditional | string assignment / concatenation | Standalone calls into known string variables only; buffer/pointer return semantics otherwise retained. |
| strchr | string.h | review | strings/bytes search, copy or explicit errors | Pointer results, NULL checks, padding, shared token state and locale require surrounding-code repair. |
| strcmp | string.h | conditional | len / strings.Compare | Known byte buffers use the prefix before NUL; string defaults assume C-string representation. |
| strcoll | string.h | review | strings/bytes search, copy or explicit errors | Pointer results, NULL checks, padding, shared token state and locale require surrounding-code repair. |
| strcpy | string.h | conditional | string assignment / concatenation | Standalone calls into known string variables only; buffer/pointer return semantics otherwise retained. |
| strcspn | string.h | review | strings/bytes search, copy or explicit errors | Pointer results, NULL checks, padding, shared token state and locale require surrounding-code repair. |
| strerror | string.h | review | strings/bytes search, copy or explicit errors | Pointer results, NULL checks, padding, shared token state and locale require surrounding-code repair. |
| strlen | string.h | conditional | len / strings.Compare | Known byte buffers use the prefix before NUL; string defaults assume C-string representation. |
| strncat | string.h | review | strings/bytes search, copy or explicit errors | Pointer results, NULL checks, padding, shared token state and locale require surrounding-code repair. |
| strncmp | string.h | conditional | len / strings.Compare | Known byte buffers use the prefix before NUL; string defaults assume C-string representation. |
| strncpy | string.h | review | strings/bytes search, copy or explicit errors | Pointer results, NULL checks, padding, shared token state and locale require surrounding-code repair. |
| strpbrk | string.h | review | strings/bytes search, copy or explicit errors | Pointer results, NULL checks, padding, shared token state and locale require surrounding-code repair. |
| strrchr | string.h | review | strings/bytes search, copy or explicit errors | Pointer results, NULL checks, padding, shared token state and locale require surrounding-code repair. |
| strspn | string.h | review | strings/bytes search, copy or explicit errors | Pointer results, NULL checks, padding, shared token state and locale require surrounding-code repair. |
| strstr | string.h | review | strings/bytes search, copy or explicit errors | Pointer results, NULL checks, padding, shared token state and locale require surrounding-code repair. |
| strtok | string.h | review | strings/bytes search, copy or explicit errors | Pointer results, NULL checks, padding, shared token state and locale require surrounding-code repair. |
| strxfrm | string.h | review | strings/bytes search, copy or explicit errors | Pointer results, NULL checks, padding, shared token state and locale require surrounding-code repair. |
| atomic_compare_exchange_strong | threads.h/stdatomic.h | review | goroutines / sync / sync/atomic | Synchronization, memory order and data representation require a coordinated port. |
| atomic_fetch_add | threads.h/stdatomic.h | review | goroutines / sync / sync/atomic | Synchronization, memory order and data representation require a coordinated port. |
| atomic_load | threads.h/stdatomic.h | review | goroutines / sync / sync/atomic | Synchronization, memory order and data representation require a coordinated port. |
| atomic_store | threads.h/stdatomic.h | review | goroutines / sync / sync/atomic | Synchronization, memory order and data representation require a coordinated port. |
| call_once | threads.h/stdatomic.h | review | goroutines / sync / sync/atomic | Synchronization, memory order and data representation require a coordinated port. |
| cnd_signal | threads.h/stdatomic.h | review | goroutines / sync / sync/atomic | Synchronization, memory order and data representation require a coordinated port. |
| cnd_wait | threads.h/stdatomic.h | review | goroutines / sync / sync/atomic | Synchronization, memory order and data representation require a coordinated port. |
| mtx_init | threads.h/stdatomic.h | review | goroutines / sync / sync/atomic | Synchronization, memory order and data representation require a coordinated port. |
| mtx_lock | threads.h/stdatomic.h | review | goroutines / sync / sync/atomic | Synchronization, memory order and data representation require a coordinated port. |
| mtx_unlock | threads.h/stdatomic.h | review | goroutines / sync / sync/atomic | Synchronization, memory order and data representation require a coordinated port. |
| thrd_create | threads.h/stdatomic.h | review | goroutines / sync / sync/atomic | Synchronization, memory order and data representation require a coordinated port. |
| thrd_join | threads.h/stdatomic.h | review | goroutines / sync / sync/atomic | Synchronization, memory order and data representation require a coordinated port. |
| CLOCKS_PER_SEC | time.h | conditional | time.Now().UnixNano() / int64(time.Second) | Explicit wall-time draft policy replaces CPU ticks; review CPU-vs-wall and monotonic timing needs. |
| asctime | time.h | review | time package | CPU vs wall clock, epoch, time_t width, timezone and struct layouts must be chosen. |
| clock | time.h | conditional | time.Now().UnixNano() / int64(time.Second) | Explicit wall-time draft policy replaces CPU ticks; review CPU-vs-wall and monotonic timing needs. |
| ctime | time.h | review | time package | CPU vs wall clock, epoch, time_t width, timezone and struct layouts must be chosen. |
| difftime | time.h | convert | difference of int64 Unix seconds | Unsigned magnitude subtraction avoids signed overflow and time.Duration range limits. |
| gmtime | time.h | review | time package | CPU vs wall clock, epoch, time_t width, timezone and struct layouts must be chosen. |
| localtime | time.h | review | time package | CPU vs wall clock, epoch, time_t width, timezone and struct layouts must be chosen. |
| mktime | time.h | review | time package | CPU vs wall clock, epoch, time_t width, timezone and struct layouts must be chosen. |
| strftime | time.h | review | time package | CPU vs wall clock, epoch, time_t width, timezone and struct layouts must be chosen. |
| time | time.h | convert | time.Now().Unix() | Unix seconds; preserve optional output pointer. time_t defaults to int64. |
| timespec_get | time.h | review | time package | CPU vs wall clock, epoch, time_t width, timezone and struct layouts must be chosen. |


## Timing policy (prototype-42)

`time(NULL)` becomes `time.Now().Unix()`, retaining Unix seconds. `time_t`
defaults to int64. Under prototype-43, non-null output pointers and `difftime`
remain review sites because their former conversions required closures.

At the user's request, `clock()` now uses `time.Now().UnixNano()` and clock_t
becomes int64; CLOCKS_PER_SEC becomes int64(time.Second). This is an explicit
**wall-clock draft policy**, changing C's CPU-time meaning. Numeric subtraction
continues to work with consistent units, but wall-clock adjustments can affect
it. For elapsed timing, a later structural repair should retain time.Time values
and use time.Since. A source-defined CLOCKS_PER_SEC override blocks automatic
clock conversion pending unit review.

References: [Go time](https://pkg.go.dev/time),
[OpenBSD time](https://man.openbsd.org/time.3), and
[OpenBSD clock](https://man.openbsd.org/clock).
Calendar/struct tm routines remain review cases because field layout, timezone,
static-storage behavior and strftime layouts need coordinated conversion.


## Assertions (prototype-46)

C `assert(x)` becomes `assert(x, "x")`, with quoted/escaped condition text.
Nested library conversions can change the first argument; the message retains
its text from before the library pass. C truth values remain downstream work.
No inline panic blocks or closures are inserted into call sites.

Tree and file-batch output generate one helper in each directory that needs it,
using that directory's resolved package name, including an explicit -package.
After translated files are written, choose assert.go, assert2.go, assert3.go, ...
without overwriting any file. TreeReport.HelperFiles lists the generated paths.
Library users can inspect Result.NeedsAssertHelper() and call WriteAssertHelper
once per output directory after writing all drafts.

```go
func assert(x bool, message string) {
    if x {
        return
    }
    _, fileName, fileLine, ok := runtime.Caller(1)
    var where string
    if ok {
        where = fmt.Sprintf("%s:%d: ", path.Base(fileName), fileLine)
    }
    panic(where + "assertion failed, we do not see: " + message)
}
```

Assertions are always active; source-defined/shadowed assert names and opaque
preprocessor regions remain protected by the existing library-pass rules.

Prototype-48 uses runtime.Caller(1) to report the immediate assertion call site.
The filename is reduced with path.Base; unavailable caller information leaves
the prefix empty. Successful assertions return before collecting stack data.


## POSIX mutexes (prototype-50)

- pthread_mutex_t becomes sync.Mutex, including pointer parameters and fields.
- The declaration initializer PTHREAD_MUTEX_INITIALIZER is removed: Go's zero
  value is ready for use. Comments between initializer tokens are retained.
- pthread_mutex_lock(&mutex) becomes mutex.Lock().
- pthread_mutex_unlock(&mutex) becomes mutex.Unlock(). Pointer arguments also
  use method calls; grouped operands and member access are supported.
- The sync import is added only when a mutex type needs it; existing aliases and
  import-name collisions use the shared import planner.

No closures or status adapters are emitted. Integer-return checks are left for
compiler-guided repair because Lock and Unlock return no value. Initialization
attributes, destruction, trylock and timedlock remain review cases. Non-default
pthread mutex semantics and copying a Go mutex after use require review.


## Mechanical follow-ups (prototype-52)

Known numeric/pointer conditions now become comparisons to zero/nil. Direct
mutex-success assertions are reduced to the single Lock/Unlock operation; no
closure is generated. ABC_ALLOC/ABC_CALLOC of exactly one known type become
new(T); Go zero-initializes the object. Whole known struct clearing can become
*p = T{}. Other allocation, free/realloc and pthread lifecycle calls are recorded
for coordinated representation/ownership repair. Literal integer suffixes retain
their widths/signs through explicit typed conversions in expressions.
