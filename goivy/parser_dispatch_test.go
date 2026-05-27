package goivy

import "testing"

func TestParseFullFileVersionsRunInOneProcess(t *testing.T) {
	cases := []struct {
		name    string
		version Version
	}{
		{name: "ivy16", version: Version{1, 6}},
		{name: "ivy17", version: Version{1, 7}},
		{name: "ivy18", version: Version{1, 8}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := Parse("type t", tc.version, WithFilename(tc.name+".ivy"))
			if err != nil {
				t.Fatalf("Parse version %v: %v", tc.version, err)
			}
			if countDeclsOf[*TypeDecl](result.Decls) != 1 {
				t.Fatalf("Parse version %v TypeDecl count = %d, want 1", tc.version, countDeclsOf[*TypeDecl](result.Decls))
			}
		})
	}
}
