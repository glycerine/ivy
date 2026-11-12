package extra

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

func TestMaccOriginalDecompositions(t *testing.T) {
	// Little-endian packed tables from original C, coefficient widths 2..10.
	hash := sha256.New()
	for b := int32(2); b <= 10; b++ {
		p := maccGenerate(b, io.Discard)
		for _, v := range p {
			var data [4]byte
			binary.LittleEndian.PutUint32(data[:], v)
			hash.Write(data[:])
		}
	}
	if got := fmt.Sprintf("%x", hash.Sum(nil)); got != "98634aa12c603fb70ad87f852f3e92ffcf119cfd0faf90ad99a16bf32f0ec306" {
		t.Fatalf("C decomposition mismatch: %s", got)
	}
}

func TestMaccGeneratedNetworkArithmetic(t *testing.T) {
	wire := regexp.MustCompile(`wire \[\d+:0\] ([nN]\d+) = (.*);`)
	for b := int32(2); b <= 9; b++ {
		p := maccGenerate(b, io.Discard)
		bound := int32(1) << (b - 1)
		for n := -bound; n < bound; n++ {
			var out bytes.Buffer
			Macc_ConstMultGenMacc2(&out, p, n, b, 12)
			width := uint(b + 12)
			mask := uint64(1)<<width - 1
			for _, input := range []uint64{0, 1, 17, 2047, 4095, 0xfffff} {
				values := map[string]uint64{"i": input & mask}
				seen := map[string]bool{}
				read := func(name string) uint64 {
					v, ok := values[name]
					if !ok {
						t.Fatalf("coefficient %d: missing wire %s", n, name)
					}
					return v
				}
				for _, line := range strings.Split(out.String(), "\n") {
					m := wire.FindStringSubmatch(line)
					if m == nil {
						continue
					}
					if seen[m[1]] {
						t.Fatalf("duplicate declaration %s for %d", m[1], n)
					}
					seen[m[1]] = true
					fields := strings.Fields(m[2])
					var v uint64
					switch {
					case len(fields) == 1 && strings.HasPrefix(fields[0], "-"):
						v = -read(fields[0][1:])
					case len(fields) == 1:
						v = read(fields[0])
					case len(fields) == 3:
						a := read(fields[0])
						switch fields[1] {
						case "+":
							v = a + read(fields[2])
						case "-":
							v = a - read(fields[2])
						case "<<":
							shift, e := strconv.Atoi(fields[2])
							if e != nil {
								t.Fatal(e)
							}
							v = a << shift
						default:
							t.Fatal("operator")
						}
					default:
						t.Fatalf("unknown expression %q", m[2])
					}
					values[m[1]] = v & mask
				}
				if n != 0 {
					got := read(fmt.Sprintf("%c%d", maccSign(n), maccMagnitude(n)))
					want := uint64(int64(n)*int64(input&mask)) & mask
					if got != want {
						t.Fatalf("width%d n%d input%d got%d want%d\n%s", b, n, input, got, want, out.String())
					}
				}
			}
		}
	}
}
func TestMaccZeroAndNegativePower(t *testing.T) {
	p := maccGenerate(8, io.Discard)
	var out bytes.Buffer
	Macc_ConstMultGenOne_rec(&out, p, 0, 8, 16)
	if !strings.Contains(out.String(), "n0 = 24'h0;") {
		t.Fatal("zero recursion")
	}
	out.Reset()
	Macc_ConstMultGenMacc2(&out, p, -128, 8, 16)
	if !strings.Contains(out.String(), "N1 = -n1;") || !strings.Contains(out.String(), "N128 = N1 << 7;") {
		t.Fatal(out.String())
	}
	for mode := 0; mode < 3; mode++ {
		out.Reset()
		maccModule(&out, p, 0, 8, 16, mode)
		if strings.Contains(out.String(), "wire") {
			t.Fatal("zero should need no intermediate wires")
		}
	}
}
func TestMaccDimensions(t *testing.T) {
	for _, c := range [][3]int32{{0, 1, 8}, {0, 16, 8}, {128, 8, 16}, {0, 8, 0}, {0, 8, 1<<31 - 1}} {
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("accepted %v", c)
				}
			}()
			maccDimensions(c[0], c[1], c[2])
		}()
	}
}
