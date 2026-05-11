# Diagnose and Fix TinyGo WASM nil pointer dereference in fmt.buffer.write
Created: 2026-05-10 21:35 UTC

## Context

`make tinynode-golden-2hr` crashes the TinyGo-compiled goivy WASM after ~26,493 xtraces
with a nil pointer dereference inside Go's fmt package:

```
panic: runtime error: nil pointer dereference
RuntimeError: unreachable
    at main.runtime.nilPanic
    at main.(*fmt.buffer).write
    at main.(*fmt.fmt).pad
    at main.(*fmt.fmt).fmtQ        ← %q verb for a string
    at main.(*fmt.pp).fmtString
    at main.(*fmt.pp).printArg
    at main.(*fmt.pp).doPrintf
    at main.fmt.Sprintf
```

Standard Go (`bigGo`) never crashes. TinyGo WASM always crashes.

---

## Root Cause: Fixed 64 KB Goroutine Stack, No Growth

### 1. TinyGo goroutine stack is fixed at 64 KB — forever

`targets/wasm.json` and `targets/wasip1.json` both contain:
```json
"default-stack-size": 65536
```
(`compileopts/target.go:266`: fallback is also `1024 * 64 = 65536`)

The TinyGo flag `-stack-size=N` (`main.go:1615`, `compileopts/config.go:220-224`)
overrides this, but **the current Makefile build does not pass it**.

### 2. TinyGo has zero stack growth capability

`grep` for `morestack`, `growStack`, `newstack`, `runtime.morestack` in
`compiler/` and `src/runtime/` returns **empty**. There is no equivalent to
Go's `runtime.morestack` in TinyGo. Goroutine stacks cannot grow.

### 3. The 64 KB buffer is shared between two competing regions

From `src/internal/task/task_asyncify.go:65-82`:
```go
stack := runtime_alloc(stackSize, nil)      // 64 KB heap block
s.canaryPtr = (*uintptr)(stack)
*s.canaryPtr = stackCanary
s.asyncifysp = unsafe.Add(stack, sizeof(uintptr))  // asyncify state → grows UP
s.csp        = unsafe.Add(stack, stackSize)        // C shadow stack → grows DOWN
```

The same 64 KB allocation holds **both**:
- The C shadow stack (local variables that need real addresses), growing **downward** from the top
- The asyncify save state (serialised frame locals on goroutine yield), growing **upward** from the bottom

The overflow check (`asyncifysp > csp`) fires only at `Resume()` time.
If the C stack overflows **past the bottom of the 64 KB block** during a
long uninterrupted computation, the check never fires and adjacent heap
memory is silently corrupted.

### 4. Standard Go is immune because it grows stacks dynamically

Standard Go goroutines start at 8 KB and grow on demand (up to 1 GB).
Deep ivy AST processing exhausts 64 KB but never 8 KB+ growing stacks.

### 5. How the nil pointer gets into fmt.buffer.write

The 64 KB block was allocated by `runtime_alloc` inside the GC heap.
Adjacent heap objects — including `fmt.pp` structs from `sync.Pool` —
sit just below the block's base address. When the C shadow stack
overflows past base address `A`, it writes garbage/zeros into memory
at `[A-N, A)`.

A `fmt.pp` struct that happens to occupy `[A-M, A-M+sizeof(pp))` can
have its `fmt.fmt.buf *buffer` pointer zeroed to `nil`.

The next call to `fmt.Printf` (via `xtracer.Trace`) retrieves a fresh
`pp` via `ppFree.Get()`, calls `p.fmt.init(&p.buf)` — **but the
overflow may corrupt the currently-active pp in the middle of a previous
fmt call**, after `init()` has already run for that call. That in-flight
`pp.fmt.buf` is set to nil by the overflow, so the very next
`f.buf.write(b)` call in `pad()` hits a nil receiver.

### 6. The Makefile already sets the WASM shadow stack — but not goroutine stacks

`Makefile:383`:
```
tinygo build -panic=print -gc=precise -no-debug \
  -ldflags="-extldflags='--initial-memory=4294967296 --stack-first -z stack-size=2097152'" \
  -o webvue/static/goivy-check-tinygo-js.wasm ./cmd/goivy_check_jswasm/
```

`-z stack-size=2097152` is a **wasm-ld linker flag** — it sets the WASM
linear-memory shadow stack for the *main* execution context. It does **not**
affect per-goroutine asyncify stack size. Those remain at 64 KB each.

---

## Fix

### Primary fix — add `-stack-size=2097152` to the tinygo build command

File: `Makefile` line ~383 (the `webvue/static/goivy-check-tinygo-js.wasm` recipe)

Change:
```makefile
GOOS=js GOARCH=wasm tinygo build -panic=print -gc=precise -no-debug \
  -ldflags="-extldflags='--initial-memory=4294967296 --stack-first -z stack-size=2097152'" \
  -o webvue/static/goivy-check-tinygo-js.wasm ./cmd/goivy_check_jswasm/
```

To:
```makefile
GOOS=js GOARCH=wasm tinygo build -panic=print -gc=precise -no-debug \
  -stack-size=2097152 \
  -ldflags="-extldflags='--initial-memory=4294967296 --stack-first -z stack-size=2097152'" \
  -o webvue/static/goivy-check-tinygo-js.wasm ./cmd/goivy_check_jswasm/
```

`-stack-size=2097152` sets each goroutine's asyncify buffer to **2 MB**
(32× the current 64 KB). `bytesize.Parse()` (`main.go:1616`) accepts
plain integer bytes; "2MB" or "2097152" both work.

The existing `-z stack-size=2097152` linker flag is unrelated (WASM
shadow stack for non-goroutine code) and should remain unchanged.

If 2 MB is still insufficient (possible for deeply nested ivy module
checks), try 4 MB (`4194304`) or 8 MB.

### Why this fixes it

`compileopts/config.go:220-224`:
```go
func (c *Config) StackSize() uint64 {
    if c.Options.StackSize != 0 {
        return c.Options.StackSize   // ← flag wins
    }
    return c.Target.DefaultStackSize // ← 65536 (current)
}
```

`builder/build.go:213`: `DefaultStackSize: config.StackSize()` feeds
this into the compiler, which emits `i32 2097152` instead of `i32 65536`
in all goroutine-start calls (confirmed in `compiler/testdata/goroutine-wasm-asyncify.ll`).

---

## Critical Files

| File | Purpose |
|------|---------|
| `Makefile:383` | TinyGo build command — **the one line to change** |
| `/Users/jaten/go/src/github.com/tinygo-org/tinygo/targets/wasm.json:13` | Confirms 64 KB default |
| `/Users/jaten/go/src/github.com/tinygo-org/tinygo/src/internal/task/task_asyncify.go:65-82` | The shared-buffer layout |
| `/Users/jaten/go/src/github.com/tinygo-org/tinygo/compileopts/config.go:218-224` | How `-stack-size` overrides the default |

---

## Verification

1. Rebuild the WASM: `make webvue/static/goivy-check-tinygo-js.wasm`
2. Run the golden test: `cd ~/ivy/goivy && make tinynode-golden-2hr`
3. Check that `log.tinynode.golden.2hr` contains no `panic: runtime error: nil pointer dereference`
   and that the xtrace count progresses well beyond i=26493.
4. Optionally pass `-print-stacks` to the tinygo build to confirm which
   goroutine function actually uses the most stack and tune the size precisely.
