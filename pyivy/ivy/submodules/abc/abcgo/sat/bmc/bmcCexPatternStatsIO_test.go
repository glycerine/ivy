package bmc

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
)

// These APIs use ABC's fixed MNIST file name, so isolate their working directory.
func inMNISTDir(t *testing.T) string {
	t.Helper()
	old, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	if err = os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(old); err != nil {
			t.Error(err)
		}
	})
	return filepath.Join(dir, "train-images.idx3-ubyte")
}
func mnistFixture(count, rows, columns uint32, pixels []byte) []byte {
	b := make([]byte, 16, len(pixels)+16)
	binary.BigEndian.PutUint32(b, 2051)
	binary.BigEndian.PutUint32(b[4:], count)
	binary.BigEndian.PutUint32(b[8:], rows)
	binary.BigEndian.PutUint32(b[12:], columns)
	return append(b, pixels...)
}
func TestMNISTReadImagesBitOrderAndBounds(t *testing.T) {
	path := inMNISTDir(t)
	pixels := make([]byte, 2*784)
	pixels[0] = 0x81
	pixels[783] = 0x82
	pixels[784] = 0x46
	pixels[len(pixels)-1] = 0xff
	raw := mnistFixture(2, 28, 28, pixels)
	if err := os.WriteFile(path, raw, 0600); err != nil {
		t.Fatal(err)
	}
	for _, count := range []int32{0, 1, 2} {
		pats := Mnist_ReadImages_(count)
		if pats == nil || vec.Vec_WecSize(pats) != int(count) {
			t.Fatalf("count %d: %v", count, pats)
		}
		for image := 0; image < int(count); image++ {
			got := vec.Vec_IntArray(vec.Vec_WecEntry(pats, image))
			if len(got) != 784*8 {
				t.Fatalf("image length %d", len(got))
			}
			for pixel := 0; pixel < 784; pixel++ {
				for bit := 0; bit < 8; bit++ {
					want := int32(pixels[image*784+pixel] >> uint(bit) & 1)
					if got[pixel*8+bit] != want {
						t.Fatalf("image %d pixel %d bit %d: %d != %d", image, pixel, bit, got[pixel*8+bit], want)
					}
				}
			}
		}
	}
	if Mnist_ReadImages_(-1) != nil || Mnist_ReadImages_(3) != nil {
		t.Fatal("accepted invalid count")
	}
	for _, tc := range []struct {
		name  string
		data  []byte
		count int32
	}{
		{"short header", raw[:15], 1}, {"short pixels", raw[:len(raw)-1], 2}, {"wrong rows", mnistFixture(2, 27, 28, pixels), 1}, {"wrong columns", mnistFixture(2, 28, 29, pixels), 1},
		{"wrong magic", append([]byte{0, 0, 0, 0}, raw[4:]...), 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if err := os.WriteFile(path, tc.data, 0600); err != nil {
				t.Fatal(err)
			}
			if Mnist_ReadImages_(tc.count) != nil {
				t.Fatal("malformed MNIST accepted")
			}
		})
	}
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	if Mnist_ReadImages1_() != nil || Mnist_ReadImages_(1) != nil {
		t.Fatal("missing file accepted")
	}
}

func TestCareNodeStatisticsAggregate(t *testing.T) {
	g := gia.Gia_ManStart(12)
	a, b, c := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	x := gia.Gia_ManAppendAnd(g, a, b)
	gia.Gia_ManAppendCo(g, gia.Gia_ManAppendAnd(g, x, c))
	patterns := vec.Vec_WecStart(2)
	for _, v := range []int32{1, 1, 1} {
		vec.Vec_WecPush(patterns, 0, v)
	}
	for _, v := range []int32{1, 1, 0} {
		vec.Vec_WecPush(patterns, 1, v)
	}
	// Capture the reporting API's integer average and percentage, not just its helper.
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w
	Gia_ManCountCareBits(g, patterns)
	os.Stdout = old
	if err = w.Close(); err != nil {
		t.Fatal(err)
	}
	output := make([]byte, 256)
	n, err := r.Read(output)
	if err != nil {
		t.Fatal(err)
	}
	r.Close()
	if got := string(output[:n]); !strings.Contains(got, "Stats over 2 patterns: Average care-nodes = 1 ( 50.00 %)") {
		t.Fatalf("unexpected report %q", got)
	}
}
