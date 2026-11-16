package gia

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"
	"time"
)

func TestMappingStatsGoOnlyPersistentAppend(t *testing.T) {
	giaMappingStatsState.Lock()
	oldName, oldSince := giaMappingStatsState.name, giaMappingStatsState.since
	giaMappingStatsState.name, giaMappingStatsState.since = "", time.Time{}
	giaMappingStatsState.Unlock()
	defer func() {
		giaMappingStatsState.Lock()
		giaMappingStatsState.name, giaMappingStatsState.since = oldName, oldSince
		giaMappingStatsState.Unlock()
	}()
	p := mappingReviewAverageGraph()
	defer Gia_ManStop(p)
	dir := t.TempDir()
	first, second := filepath.Join(dir, "first.txt"), filepath.Join(dir, "second.txt")
	hashReviewStdout(t, func() {
		p.SetIdentity("persistent_A", "")
		Gia_ManPrintMappingStats(p, first)
		Gia_ManPrintMappingStats(p, first)
		// The source history is shared across paths, not indexed by file name.
		Gia_ManPrintMappingStats(p, second)
		p.SetIdentity("persistent_B", "")
		Gia_ManPrintMappingStats(p, first)
		p.SetIdentity("persistent_A", "")
		Gia_ManPrintMappingStats(p, first)
	})
	elapsed := regexp.MustCompile(`[0-9]+\.[0-9]{2}`)
	for path, want := range map[string]string{
		first:  "\npersistent_A  2 2 2  2 2 <elapsed>\npersistent_B  2 2 2 \npersistent_A  2 2 2 ",
		second: " 2 2 <elapsed>",
	} {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if got := elapsed.ReplaceAllString(string(data), "<elapsed>"); got != want {
			t.Fatalf("file %s: got%q want%q", path, got, want)
		}
	}
	unmapped := Gia_ManStart(1)
	defer Gia_ManStop(unmapped)
	absent := filepath.Join(dir, "absent.txt")
	Gia_ManPrintMappingStats(unmapped, absent)
	if _, err := os.Stat(absent); !os.IsNotExist(err) {
		t.Fatal("unmapped report created a statistics file")
	}
}
