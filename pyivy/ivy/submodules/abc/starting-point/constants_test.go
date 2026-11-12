package startingpoint

import (
	"strings"
	"testing"
)

func TestIntegerDefines(t *testing.T) {
	source := `#define NUMBER1  3716960521u
#define HEX 0xDeADbeefULL // keep suffix text: ULL
#define OCT 077UL
#define BIN 0b101u
#define NEG -42L
#define ZERO 0
#define HUGE 18446744073709551615ULL
#define NOTE /* comment */ 23u
`
	r := checkedDraft(t, source, Options{Package: "p"})
	for _, want := range []string{"const NUMBER1 uint32 =  3716960521", "const HEX uint64 = 0xDeADbeef // keep suffix text: ULL", "const OCT uint64 = 077", "const BIN uint32 = 0b101", "const NEG int64 = -42", "const HUGE uint64 = 18446744073709551615", "const NOTE uint32 = /* comment */ 23"} {
		if !strings.Contains(r.Draft, want) {
			t.Fatal(r.Draft)
		}
	}
}

func TestNonLiteralDefinesUnchanged(t *testing.T) {
	source := "#define F(x) x\n#define EMPTY\n#define EXPR (1 << 4)\n#define INVALID 09u\n// #define FAKE 4u\n"
	got, err := integerDefines(source)
	if err != nil {
		t.Fatal(err)
	}
	if got != source {
		t.Fatal(got)
	}
}

func TestGeneratedConstNotStripped(t *testing.T) {
	r := checkedDraft(t, `static inline unsigned f() {
#define LOCAL 7u
    return LOCAL;
}`, Options{Package: "p"})
	if !strings.Contains(r.Functions[0].Draft, "const LOCAL uint32 = 7") {
		t.Fatal(r.Functions[0].Draft)
	}
}

func TestIntegerSuffixWidths(t *testing.T) {
	source := `#define SMALL 4294967295u
#define BIG 4294967296U
#define LONG 1L
#define LONG_LONG 1LL
#define ULONG 1UL
#define ULLONG 1ULL
#define NEGATIVE -1u
#define NEGATIVE64 -1ULL
#define HEX_SIGNED 0x7fffffffffffffffL
#define HEX_UNSIGNED 0xffffffffffffffffL
`
	r := checkedDraft(t, source, Options{Package: "p"})
	for _, want := range []string{"const SMALL uint32 = 4294967295", "const BIG uint64 = 4294967296", "const LONG int64 = 1", "const LONG_LONG int64 = 1", "const ULONG uint64 = 1", "const ULLONG uint64 = 1", "const NEGATIVE uint32 = 4294967295", "const NEGATIVE64 uint64 = 18446744073709551615", "const HEX_SIGNED int64 = 0x7fffffffffffffff", "const HEX_UNSIGNED uint64 = 0xffffffffffffffff"} {
		if !strings.Contains(r.Draft, want) {
			t.Fatal(r.Draft)
		}
	}
}
