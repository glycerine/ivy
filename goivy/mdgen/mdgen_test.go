package mdgen

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// helper: convert lines preserving newlines
func mkLines(ss ...string) []string {
	out := make([]string, len(ss))
	for i, s := range ss {
		out[i] = s + "\n"
	}
	return out
}

func TestEmptyInput(t *testing.T) {
	// Only a header line -> no output
	var buf bytes.Buffer
	ConvertLines(&buf, mkLines("header"))
	if buf.Len() != 0 {
		t.Errorf("expected empty output, got %q", buf.String())
	}
}

func TestHeaderSkipped(t *testing.T) {
	// First line is always skipped
	var buf bytes.Buffer
	ConvertLines(&buf, mkLines("header", "# hello"))
	got := buf.String()
	if strings.Contains(got, "header") {
		t.Error("header line should be skipped")
	}
	if !strings.Contains(got, "hello") {
		t.Error("comment should appear")
	}
}

func TestCommentToProse(t *testing.T) {
	var buf bytes.Buffer
	ConvertLines(&buf, mkLines("hdr", "# This is a comment"))
	got := buf.String()
	if !strings.Contains(got, "This is a comment") {
		t.Errorf("expected prose, got %q", got)
	}
	if strings.Contains(got, "```") {
		t.Error("pure comment output should not have code fences")
	}
}

func TestCodeToFencedBlock(t *testing.T) {
	var buf bytes.Buffer
	ConvertLines(&buf, mkLines("hdr", "action foo = {}"))
	got := buf.String()
	if !strings.Contains(got, "```") {
		t.Error("code should be in fenced block")
	}
	if !strings.Contains(got, "action foo = {}") {
		t.Error("code line missing")
	}
}

func TestMixedCommentAndCode(t *testing.T) {
	lines := mkLines(
		"hdr",
		"# intro",
		"type node",
		"# another comment",
		"relation link(X:node, Y:node)",
	)
	var buf bytes.Buffer
	ConvertLines(&buf, lines)
	got := buf.String()
	// Should have opening and closing fences
	if strings.Count(got, "```") < 2 {
		t.Errorf("expected at least 2 fence markers, got %d in:\n%s",
			strings.Count(got, "```"), got)
	}
}

func TestDashCommentIsCode(t *testing.T) {
	// #- lines are treated as code, not comments
	var buf bytes.Buffer
	ConvertLines(&buf, mkLines("hdr", "#- this is code"))
	got := buf.String()
	if !strings.Contains(got, "```") {
		t.Error("#- should be treated as code")
	}
}

func TestBlankLinesPreserved(t *testing.T) {
	lines := mkLines("hdr", "", "# comment")
	var buf bytes.Buffer
	ConvertLines(&buf, lines)
	got := buf.String()
	if !strings.HasPrefix(got, "\n") {
		t.Errorf("blank line should be preserved, got %q", got)
	}
}

func TestClosingFenceAtEnd(t *testing.T) {
	lines := mkLines("hdr", "some code")
	var buf bytes.Buffer
	ConvertLines(&buf, lines)
	got := buf.String()
	if !strings.HasSuffix(got, "```\n") {
		t.Errorf("should end with closing fence, got %q", got)
	}
}

func TestNoClosingFenceForCommentEnd(t *testing.T) {
	lines := mkLines("hdr", "# just a comment")
	var buf bytes.Buffer
	ConvertLines(&buf, lines)
	got := buf.String()
	if strings.Contains(got, "```") {
		t.Errorf("should not have fences for comment-only output, got %q", got)
	}
}

func TestConvertString(t *testing.T) {
	input := "header\n# hello world\naction foo = {}\n"
	got := ConvertString(input)
	if !strings.Contains(got, "hello world") {
		t.Error("missing comment prose")
	}
	if !strings.Contains(got, "action foo") {
		t.Error("missing code")
	}
}

func TestConvertFileRoundTrip(t *testing.T) {
	dir := t.TempDir()
	inPath := filepath.Join(dir, "test.ivy")
	outPath := filepath.Join(dir, "test.md")

	content := "header\n# My doc\ntype node\n"
	os.WriteFile(inPath, []byte(content), 0644)

	err := ConvertFile(inPath, "")
	if err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatal(err)
	}
	got := string(data)
	if !strings.Contains(got, "My doc") {
		t.Error("missing comment in output file")
	}
}

func TestConvertFileNotIvy(t *testing.T) {
	err := ConvertFile("foo.txt", "")
	if err == nil {
		t.Error("expected error for non-.ivy file")
	}
}

func TestConvertFileMissing(t *testing.T) {
	err := ConvertFile("/nonexistent/path/test.ivy", "")
	if err == nil {
		t.Error("expected error for missing file")
	}
}

func TestMultipleHashInComment(t *testing.T) {
	// "## heading" -> the text after first # should be "# heading"
	var buf bytes.Buffer
	ConvertLines(&buf, mkLines("hdr", "## heading"))
	got := buf.String()
	if !strings.Contains(got, "# heading") {
		t.Errorf("multi-hash should preserve extra hashes, got %q", got)
	}
}

func TestTransitionCommentToCodeToComment(t *testing.T) {
	lines := mkLines("hdr", "# start", "code1", "code2", "# end")
	var buf bytes.Buffer
	ConvertLines(&buf, lines)
	got := buf.String()
	// Should have: start prose, ```, code, ```, end prose
	if strings.Count(got, "```") != 2 {
		t.Errorf("expected exactly 2 fence markers, got %d in:\n%s",
			strings.Count(got, "```"), got)
	}
}

func FuzzConvertString(f *testing.F) {
	f.Add("header\n# comment\ncode line\n")
	f.Add("header\n\n# multi\n# line\ncode\n#- dash\n")
	f.Add("")
	f.Add("single line")

	f.Fuzz(func(t *testing.T, input string) {
		// Should not panic on any input
		result := ConvertString(input)
		// Every opening ``` should have a matching close
		count := strings.Count(result, "```")
		if count%2 != 0 {
			t.Errorf("unbalanced fences (%d) for input %q", count, input)
		}
	})
}
