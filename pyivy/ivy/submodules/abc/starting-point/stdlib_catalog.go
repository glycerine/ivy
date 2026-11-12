package startingpoint

// LibraryRule documents a C-library family independently of its implementation.
// Mode is "convert", "conditional", or "review"; conditional rules need known
// argument shapes/types. LibraryRules returns a copy for tooling and coverage audits.
type LibraryRule struct {
	Name, Header, Mode, Go, Notes string
}

var libraryCatalog = buildLibraryCatalog()

func LibraryRules() []LibraryRule { return append([]LibraryRule(nil), libraryCatalog...) }
func buildLibraryCatalog() []LibraryRule {
	var out []LibraryRule
	add := func(names, header, mode, target, note string) {
		for _, name := range splitWords(names) {
			out = append(out, LibraryRule{name, header, mode, target, note})
		}
	}
	add("ABC_REALLOC ABC_FREE", "ABC allocation macro", "review", "new/make and Go ownership", "Pointer versus slice representation, aliases and lifetime require coordinated repair.")
	add("pthread_create pthread_join", "pthread.h", "review", "goroutines and synchronization", "Argument lifetime, joins, thread-local state and termination require coordinated repair.")
	add("ABC_CALLOC ABC_ALLOC", "ABC allocation macro", "conditional", "new(T)", "Only single-object allocations; Go zero-initializes the object. Array/ownership cases need representation review.")
	add("pthread_mutex_t PTHREAD_MUTEX_INITIALIZER", "pthread.h", "convert", "sync.Mutex zero value", "Zero-value mutex; review attributes and copying after use.")
	add("pthread_mutex_lock pthread_mutex_unlock", "pthread.h", "convert", "Mutex.Lock / Mutex.Unlock", "Go methods return no status; repair uses of the C integer result downstream.")
	add("pthread_mutex_init pthread_mutex_destroy pthread_mutex_trylock pthread_mutex_timedlock", "pthread.h", "review", "sync.Mutex", "Initialization attributes, destruction, try-lock status and timeouts require surrounding-code repair.")
	add("stdin stdout stderr EOF", "stdio.h", "convert", "os standard streams / -1", "Stream objects and the selected conventional EOF sentinel; source overrides are protected.")
	add("EXIT_SUCCESS EXIT_FAILURE", "stdlib.h", "convert", "0 / 1", "Conventional process status constants; source overrides are protected.")
	add("printf fprintf sprintf", "stdio.h", "conditional", "fmt.Printf / fmt.Fprintf / fmt.Sprintf", "C format verbs, result arity and mutable destinations need review; existing draft rules retained.")
	add("fopen", "stdio.h", "conditional", "os.Open / os.OpenFile", "Literal standard modes; error maps to nil. Text/binary translation and errno need review.")
	add("fclose remove rename", "stdio.h", "convert", "File.Close / os.Remove / os.Rename", "Preserve C success/failure integer status; errno is not reproduced.")
	add("fgets", "stdio.h", "conditional", "bounded File.Read loop", "Known byte array/slice only; preserve n-1, newline, terminator and EOF. No read-ahead buffer is discarded.")
	add("fgetc getc getchar fputc putc putchar", "stdio.h", "convert", "File.Read / File.Write", "Single-byte transfer, unsigned byte value or -1. EOF/error distinction still needs explicit handling.")
	add("puts fputs", "stdio.h", "convert", "fmt.Fprintln / io.WriteString", "C nonnegative success / negative failure, not Go's two-result API.")
	add("fseek ftell rewind", "stdio.h", "conditional", "File.Seek", "Known SEEK constants; preserve status/position. Stream error flags are not tracked.")
	add("fflush", "stdio.h", "conditional", "no-op for unbuffered os.File", "Only known file identifiers/std streams; not File.Sync (different semantics). Retained otherwise.")
	add("fread fwrite", "stdio.h", "conditional", "io.ReadFull / File.Write", "Known byte buffers; byte count converted back to complete element count.")
	add("scanf fscanf sscanf snprintf vprintf vfprintf vsprintf vsnprintf vscanf vfscanf vsscanf freopen setbuf setvbuf feof ferror clearerr fgetpos fsetpos ungetc perror tmpfile tmpnam gets", "stdio.h", "review", "fmt scanning / buffered I/O / explicit errors", "Result counts, formats, bounds, state, positioning or lifetime require structural repair.")
	add("atoi atol atoll", "stdlib.h", "convert", "strconv.ParseInt with decimal-prefix scan", "ASCII C-locale whitespace/sign/prefix; int32 or LP64 int64. C overflow is outside the defined domain.")
	add("atof strtod strtof strtold strtol strtoul strtoll strtoull", "stdlib.h", "review", "strconv.ParseFloat / ParseInt / ParseUint", "Prefix consumption, endptr, locale, bases, overflow and errno need explicit decisions.")
	add("abs labs llabs", "stdlib.h", "convert", "typed sign check", "Integer widths retained; minimum signed value remains an overflow case.")
	add("exit _Exit", "stdlib.h", "convert", "os.Exit", "C atexit handlers and buffered flushing are not replicated.")
	add("getenv", "stdlib.h", "conditional", "os.Getenv", "Go strings cannot preserve NULL vs empty; converted with a review record.")
	add("malloc calloc realloc free aligned_alloc", "stdlib.h", "review", "make / new / GC", "Ownership, element type, byte size, pointer arithmetic and failure policy must be established.")
	add("rand srand qsort bsearch system abort atexit at_quick_exit quick_exit div ldiv lldiv mblen mbtowc wctomb mbstowcs wcstombs", "stdlib.h", "review", "math/rand / slices / sort / os/exec / encoding", "Algorithms, ordering, shell semantics, signals, cleanup and locale are not interchangeable by name.")
	add("strlen strcmp strncmp", "string.h", "conditional", "len / strings.Compare", "Known byte buffers use the prefix before NUL; string defaults assume C-string representation.")
	add("strcpy strcat", "string.h", "conditional", "string assignment / concatenation", "Standalone calls into known string variables only; buffer/pointer return semantics otherwise retained.")
	add("memcpy memmove memset memcmp", "string.h", "conditional", "copy / byte fill / bytes.Compare", "Known byte buffers only; no guessing sizeof, raw pointers or element sizes.")
	add("strchr strrchr strstr strspn strcspn strpbrk strtok strncpy strncat strcoll strxfrm strerror memchr", "string.h", "review", "strings/bytes search, copy or explicit errors", "Pointer results, NULL checks, padding, shared token state and locale require surrounding-code repair.")
	add("strdup", "POSIX string.h", "conditional", "string value", "Known strings only; allocation failure and pointer identity are not retained.")
	add("isalnum isalpha isblank iscntrl isdigit isgraph islower isprint ispunct isspace isupper isxdigit tolower toupper", "ctype.h", "convert", "ASCII byte/range checks", "C locale only; preserve numeric predicate result and EOF for case conversion, do not silently use Unicode semantics.")
	mathNames := "acos asin atan atan2 cos sin tan acosh asinh atanh cosh sinh tanh exp exp2 expm1 log log10 log1p log2 sqrt cbrt pow hypot ceil floor trunc round fabs fmod remainder copysign fdim fmin fmax erf erfc tgamma lgamma"
	add(mathNames, "math.h", "convert", "math package", "double draft; errno/fenv and exceptional-value details need review.")
	for _, name := range splitWords(mathNames) {
		add(name+"f", "math.h", "convert", "float32(math operation)", "Float inputs are widened for Go math, then rounded to float32; inspect exact rounding/exception requirements.")
		add(name+"l", "math.h", "review", "explicit long-double precision policy", "Go has no native long double; do not silently narrow precision.")
	}
	add("frexp ldexp modf remquo scalbn scalbln ilogb logb nextafter nexttoward nan nearbyint rint lrint llrint lround llround fpclassify isfinite isinf isnan isnormal signbit", "math.h", "review", "math with result/representation adaptation", "Pointer outputs, integer exponents, predicates, rounding modes or NaN details need special rules.")
	add("assert", "assert.h", "convert", "assert(condition, conditionText)", "One assert helper per output package; repair C truth values downstream. NDEBUG/abort semantics differ.")
	add("time", "time.h", "convert", "time.Now().Unix()", "Unix seconds; preserve optional output pointer. time_t defaults to int64.")
	add("clock CLOCKS_PER_SEC", "time.h", "conditional", "time.Now().UnixNano() / int64(time.Second)", "Explicit wall-time draft policy replaces CPU ticks; review CPU-vs-wall and monotonic timing needs.")
	add("difftime", "time.h", "convert", "difference of int64 Unix seconds", "Unsigned magnitude subtraction avoids signed overflow and time.Duration range limits.")
	add("mktime localtime gmtime ctime asctime strftime timespec_get", "time.h", "review", "time package", "CPU vs wall clock, epoch, time_t width, timezone and struct layouts must be chosen.")
	add("setlocale localeconv", "locale.h", "review", "explicit locale policy", "No process-global libc locale equivalent in Go's standard library.")
	add("setjmp longjmp signal raise", "setjmp.h/signal.h", "review", "structured errors / os/signal", "Control-flow and asynchronous signal semantics require redesign.")
	add("va_start va_arg va_copy va_end", "stdarg.h", "review", "variadic parameters", "C va_list ABI has no direct Go representation.")
	add("thrd_create thrd_join mtx_init mtx_lock mtx_unlock cnd_wait cnd_signal call_once atomic_load atomic_store atomic_fetch_add atomic_compare_exchange_strong", "threads.h/stdatomic.h", "review", "goroutines / sync / sync/atomic", "Synchronization, memory order and data representation require a coordinated port.")
	return out
}
