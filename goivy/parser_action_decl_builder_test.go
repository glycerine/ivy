package goivy

import "testing"

func TestParserExportActionDeclaresActionAndExport(t *testing.T) {
	result, err := Parse("export action ping = { skip }", Version{1, 7})
	if err != nil {
		t.Fatalf("parse export action: %v", err)
	}

	if countDeclsOf[*ActionDecl](result.Decls) != 1 {
		t.Fatalf("export action should declare one ActionDecl, got %d", countDeclsOf[*ActionDecl](result.Decls))
	}
	if countDeclsOf[*ExportDecl](result.Decls) != 1 {
		t.Fatalf("export action should declare one ExportDecl, got %d", countDeclsOf[*ExportDecl](result.Decls))
	}
}

func TestParserImportActionDeclaresActionAndImport(t *testing.T) {
	result, err := Parse("import action ping = { skip }", Version{1, 7})
	if err != nil {
		t.Fatalf("parse import action: %v", err)
	}

	if countDeclsOf[*ActionDecl](result.Decls) != 1 {
		t.Fatalf("import action should declare one ActionDecl, got %d", countDeclsOf[*ActionDecl](result.Decls))
	}
	if countDeclsOf[*ImportDecl](result.Decls) != 1 {
		t.Fatalf("import action should declare one ImportDecl, got %d", countDeclsOf[*ImportDecl](result.Decls))
	}
}

func countDeclsOf[T Node](decls []Node) int {
	var count int
	for _, decl := range decls {
		if _, ok := decl.(T); ok {
			count++
		}
	}
	return count
}
