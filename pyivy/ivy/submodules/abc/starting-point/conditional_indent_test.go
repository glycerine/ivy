package startingpoint

import "testing"

func TestConditionalIndentation(t *testing.T) {
	cases := []struct{ source, want string }{
		{"{\n    p.data = p.cap ? satoko_alloc(unsigned, p.cap ) : nil;\n}",
			"{\n    if (p.cap) {\n        p.data = satoko_alloc(unsigned, p.cap );\n    } else {\n        p.data = nil;\n    }\n}"},
		{"{\n\treturn a ? b : c ? d : e;\n}",
			"{\n\tif (a) {\n\t\treturn b;\n\t} else {\n\t\tif (c) {\n\t\t\treturn d;\n\t\t} else {\n\t\t\treturn e;\n\t\t}\n\t}\n}"},
		{"{\n    var x int32 = a ? 1 : 2;\n}",
			"{\n    var x int32;\n    if (a) {\n        x = 1;\n    } else {\n        x = 2;\n    }\n}"},
	}
	for _, tc := range cases {
		got, err := lowerConditionals(tc.source)
		if err != nil {
			t.Fatal(err)
		}
		if got != tc.want {
			t.Fatalf("got:\n%s\nwant:\n%s", got, tc.want)
		}
	}
}
