package abc

// Stand-alone permanent-oracle integrity checks.

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type standaloneOracleManifest struct {
	SchemaVersion int               `json:"schema_version"`
	OracleKind    string            `json:"oracle_kind"`
	Fixtures      map[string]string `json:"fixtures_sha256"`
	Policy        []string          `json:"policy"`
}

func TestStandaloneOracleDatasets(t *testing.T) {
	allowedTools := map[string]bool{
		"tools/source_manifest.py": true,
	}
	manifestData, err := os.ReadFile("test_vectors/oracle/reference_build.json")
	if err != nil {
		t.Fatal(err)
	}
	var manifest standaloneOracleManifest
	if err = json.Unmarshal(manifestData, &manifest); err != nil {
		t.Fatal(err)
	}
	if manifest.SchemaVersion != 2 || manifest.OracleKind != "standalone_go_and_immutable_datasets" || len(manifest.Policy) == 0 || len(manifest.Fixtures) == 0 {
		t.Fatal("invalid standalone oracle manifest", manifest.SchemaVersion, manifest.OracleKind, len(manifest.Policy), len(manifest.Fixtures))
	}
	for relative, want := range manifest.Fixtures {
		path := filepath.Join("test_vectors", filepath.FromSlash(relative))
		data, err := os.ReadFile(path)
		if err != nil {
			t.Errorf("manifest dataset %q: %v", relative, err)
			continue
		}
		actual := sha256.Sum256(data)
		if hex.EncodeToString(actual[:]) != want {
			t.Errorf("oracle dataset hash mismatch for %q", relative)
		}
	}
	err = filepath.WalkDir("test_vectors/oracle", func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || path == "test_vectors/oracle/reference_build.json" {
			return nil
		}
		relative, err := filepath.Rel("test_vectors", path)
		if err != nil {
			return err
		}
		relative = filepath.ToSlash(relative)
		want, ok := manifest.Fixtures[relative]
		if !ok {
			t.Errorf("untracked permanent oracle dataset %q", relative)
			return nil
		}
		_ = want // Hashes are checked uniformly above, including non-oracle vectors.
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	err = filepath.WalkDir(".", func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			if entry.Name() == "testdata" {
				t.Errorf("permanent vectors must use test_vectors, not %s", path)
			}
			return nil
		}
		if entry.Type()&os.ModeSymlink != 0 {
			t.Errorf("stand-alone test inputs and tools must not escape the Go module through a symlink: %s", path)
			return nil
		}
		switch strings.ToLower(filepath.Ext(path)) {
		case ".c", ".cc", ".cpp", ".cxx", ".h", ".hh", ".hpp", ".hxx":
			t.Errorf("native oracle source is forbidden in the Go module: %s", path)
		}
		if strings.HasSuffix(path, "_test.go") {
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			for _, token := range []string{
				"os/" + "exec",
				"exec" + ".Command(",
				"syscall" + ".Exec(",
				"//go:" + "generate",
				"import " + "\"C\"",
			} {
				if strings.Contains(string(data), token) {
					t.Errorf("test must not build or invoke an external oracle (%q): %s", token, path)
				}
			}
		}
		slashPath := filepath.ToSlash(path)
		if strings.HasPrefix(slashPath, "tools/") && !allowedTools[slashPath] {
			t.Errorf("pkg/tools contains an unapproved file that could restore an external oracle dependency: %s", path)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestStandaloneAigerRoundTrip(t *testing.T) {
	vectors := [][]byte{
		[]byte("aig 0 0 0 2 0\n0\n1\n"),
		append([]byte("aig 3 2 0 1 1\n6\n"), 1, 3),
	}
	benchmark, err := os.ReadFile("test_vectors/aiger/i10.aig")
	if err != nil {
		t.Fatal(err)
	}
	vectors = append(vectors, benchmark)
	for index, raw := range vectors {
		first, err := ReadAiger(raw, AigerLimits{})
		if err != nil {
			t.Fatal(index, err)
		}
		encoded, err := first.MarshalBinary()
		if err != nil {
			t.Fatal(index, err)
		}
		second, err := ReadAiger(encoded, AigerLimits{})
		if err != nil {
			t.Fatal(index, err)
		}
		again, err := second.MarshalBinary()
		if err != nil || string(encoded) != string(again) {
			t.Fatal(index, "noncanonical Go-only AIGER round trip", err)
		}
	}
}
