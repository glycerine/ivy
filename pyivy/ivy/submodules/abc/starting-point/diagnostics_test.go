package startingpoint

import (
	"errors"
	"testing"
)

func TestOriginalSourceCoordinates(t *testing.T) {
	for _, tc := range []struct {
		source               string
		line, column, offset int
	}{
		{"// α\r\n\t}", 2, 2, 8},
		{"int x;\n/* unfinished", 2, 1, 7},
		{"/* prefix */\n{\n", 2, 1, 13},
		{"#define A /* spanning\n unfinished", 1, 11, 10},
	} {
		_, err := Analyze("p/file.c", tc.source)
		var diagnostic *SourceError
		if !errors.As(err, &diagnostic) {
			t.Fatalf("expected source diagnostic: %v", err)
		}
		want := SourceLocation{Line: tc.line, Column: tc.column, ByteOffset: tc.offset}
		if diagnostic.Location != want {
			t.Fatalf("got %+v want %+v (%q)", diagnostic.Location, want, tc.source)
		}
	}
}
