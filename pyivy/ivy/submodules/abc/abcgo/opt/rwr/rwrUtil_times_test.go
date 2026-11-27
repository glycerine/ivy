package rwr

import (
	"io"
	"os"
	"path/filepath"
	"regexp"
	"testing"
)

func captureRwrStdout(t *testing.T, f func()) string {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	stdout := os.Stdout
	os.Stdout = w
	done := make(chan string)
	go func() {
		b, _ := io.ReadAll(r)
		done <- string(b)
	}()
	defer func() { os.Stdout = stdout }()
	f()
	w.Close()
	return <-done
}

// rwrUtil.c reports with ABC_PRT( "Loading"/"Saving", ... ); the port
// printed "Loading: 1.2ms".
func TestRwrLoadSaveTimesLikeABC(t *testing.T) {
	path := filepath.Join(t.TempDir(), "library.bin")
	out := captureRwrStdout(t, func() {
		first := rwrFixtureManager()
		Rwr_ManLoadFromArray(first, 1)
		Rwr_ManWriteToFile(first, path)
		Rwr_ManLoadFromFile(rwrFixtureManager(), path)
	})
	want := regexp.MustCompile(`^The number of classes = [0-9]+\. Canonical nodes = [0-9]+\.
The number of nodes loaded = 1779\.  Loading = +[0-9]+\.[0-9][0-9] sec
The number of nodes saved = 1780\.   Saving = +[0-9]+\.[0-9][0-9] sec
The number of classes = [0-9]+\. Canonical nodes = [0-9]+\.
The number of nodes loaded = 1780\.   Loading = +[0-9]+\.[0-9][0-9] sec
$`)
	if !want.MatchString(out) {
		t.Fatalf("got:\n%s", out)
	}
}
