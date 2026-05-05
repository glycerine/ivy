package goivy

import (
	"strings"
	"testing"
)

func TestPrettyBraceIndent(t *testing.T) {
	s := "if cond { x = 1; y = 2 }"
	result := Pretty(s, 0)
	lines := strings.Split(result, "\n")
	// Should have indented content inside braces
	foundIndented := false
	for _, line := range lines {
		if strings.HasPrefix(line, "    ") {
			foundIndented = true
			break
		}
	}
	if !foundIndented {
		t.Errorf("Pretty should indent content inside braces, got:\n%s", result)
	}
}

func TestPrettyTruncation(t *testing.T) {
	s := "a\nb\nc\nd\ne\nf"
	result := Pretty(s, 3)
	lines := strings.Split(result, "\n")
	if len(lines) != 3 { // 2 lines + "..."
		t.Errorf("Pretty truncation: got %d lines, want 3:\n%s", len(lines), result)
	}
	if lines[len(lines)-1] != "..." {
		t.Errorf("last line should be '...', got %q", lines[len(lines)-1])
	}
}

func TestPrettyNoTruncation(t *testing.T) {
	s := "a\nb"
	result := Pretty(s, 0)
	if result != s {
		t.Errorf("Pretty with maxLines=0 should not truncate, got %q", result)
	}
}

func TestPrettyClosingBraces(t *testing.T) {
	// Unclosed brace should add closing at end
	s := "a {"
	result := Pretty(s, 0)
	if !strings.HasSuffix(result, "}") {
		t.Errorf("Pretty should append closing brace, got %q", result)
	}
}
