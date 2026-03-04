package main

import (
	"bytes"
	"go/ast"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func ident(name string) *ast.Ident {
	return &ast.Ident{Name: name}
}

func TestGenPath(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"models/user.go", "models/user.gen.go"},
		{"pkg/model.go", "pkg/model.gen.go"},
		{"singlefile.go", "singlefile.gen.go"},
		{"dir/with/sub/file.go", "dir/with/sub/file.gen.go"},
		{"nofile", "nofile.gen.go"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := genPath(tt.input)
			if result != tt.expected {
				t.Errorf("genPath(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestHasGenerateResetComment(t *testing.T) {
	tests := []struct {
		name     string
		doc      *ast.CommentGroup
		expected bool
	}{
		{
			name:     "has_generate_reset",
			doc:      commentGroup("// generate:reset"),
			expected: true,
		},
		{
			name:     "has_generate_reset_with_space",
			doc:      commentGroup("  // generate:reset  "),
			expected: true,
		},
		{
			name:     "no_comment",
			doc:      nil,
			expected: false,
		},
		{
			name:     "different_comment",
			doc:      commentGroup("// go:generate something"),
			expected: false,
		},
		{
			name:     "multiple_comments_positive",
			doc:      commentGroup("// other", "// generate:reset"),
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			genDecl := &ast.GenDecl{Doc: tt.doc}
			result := hasGenerateResetComment(genDecl)
			if result != tt.expected {
				t.Errorf("hasGenerateResetComment() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestGenerateIdentStarExprReset(t *testing.T) {
	tests := []struct {
		name     string
		elemType ast.Expr
		field    string
		expected string
	}{
		{
			name:     "string_pointer",
			elemType: ident("string"),
			field:    "Name",
			expected: `if r.Name != nil {
*r.Name = ""
}
`,
		},
		{
			name:     "int_pointer",
			elemType: ident("int64"),
			field:    "Count",
			expected: `if r.Count != nil {
*r.Count = 0
}
`,
		},
		{
			name:     "float_pointer",
			elemType: ident("float64"),
			field:    "Value",
			expected: `if r.Value != nil {
*r.Value = 0.0
}
`,
		},
		{
			name:     "bool_pointer",
			elemType: ident("bool"),
			field:    "Active",
			expected: `if r.Active != nil {
*r.Active = false
}
`,
		},
		{
			name:     "custom_type",
			elemType: ident("UserData"),
			field:    "Data",
			expected: `if r.Data != nil && hasReset(r.Data) {
r.Data.Reset()
}
`,
		},
		{
			name:     "unknown_type",
			elemType: &ast.SelectorExpr{X: ident("time"), Sel: ident("Time")},
			field:    "Created",
			expected: `r.Created = nil
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generateIdentStarExprReset(tt.elemType, tt.field)
			if result != tt.expected {
				t.Errorf("generateIdentStarExprReset() =\n%q\nwant\n%q", result, tt.expected)
			}
		})
	}
}

func commentGroup(comments ...string) *ast.CommentGroup {
	list := make([]*ast.Comment, len(comments))
	for i, c := range comments {
		list[i] = &ast.Comment{Text: c}
	}
	return &ast.CommentGroup{List: list}
}

func TestGenerateResetFileTemplate_StringField(t *testing.T) {
	node := &ast.File{Name: ast.NewIdent("main")}

	typeSpec := &ast.TypeSpec{
		Name: ast.NewIdent("User"),
		Type: &ast.StructType{
			Fields: &ast.FieldList{
				List: []*ast.Field{{
					Names: []*ast.Ident{ast.NewIdent("Name")},
					Type:  ast.NewIdent("string"),
				}},
			},
		},
	}

	structs := []*StructToGenerate{{
		typeSpec:   typeSpec,
		structType: typeSpec.Type.(*ast.StructType),
	}}

	result, err := generateResetFileTemplate(structs, node)
	assert.NoError(t, err)

	assert.Contains(t, string(result), "package main")
	assert.Contains(t, string(result), `func (r *User) Reset()`)
	assert.Contains(t, string(result), `r.Name = ""`)
}

func TestGenerateResetFileTemplate_PointerField(t *testing.T) {
	node := &ast.File{Name: ast.NewIdent("main")}

	typeSpec := &ast.TypeSpec{
		Name: ast.NewIdent("Config"),
		Type: &ast.StructType{
			Fields: &ast.FieldList{
				List: []*ast.Field{{
					Names: []*ast.Ident{ast.NewIdent("Value")},
					Type:  &ast.StarExpr{X: ast.NewIdent("int64")},
				}},
			},
		},
	}

	structs := []*StructToGenerate{{
		typeSpec:   typeSpec,
		structType: typeSpec.Type.(*ast.StructType),
	}}

	result, err := generateResetFileTemplate(structs, node)
	assert.NoError(t, err)

	assert.Contains(t, string(result), `if r.Value != nil`)
	assert.Contains(t, string(result), `*r.Value = 0`)
}

func createTempGoFile(t *testing.T, content string) string {
	t.Helper()

	tmpfile, err := os.CreateTemp("", "*.go")
	if err != nil {
		t.Fatalf("create temp file: %v", err)
	}

	if _, err := tmpfile.WriteString(content); err != nil {
		tmpfile.Close()
		os.Remove(tmpfile.Name())
		t.Fatalf("write to temp file: %v", err)
	}
	tmpfile.Close()

	return tmpfile.Name()
}

func TestScanFile_WithGenerateComment(t *testing.T) {
	tmpFile := createTempGoFile(t, `package model

// generate:reset
type User struct {
    Name *string
}`)

	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr)

	scanFile(tmpFile)

	genFile := strings.Replace(tmpFile, ".go", ".gen.go", 1)
	assert.FileExists(t, genFile)

	data, _ := os.ReadFile(genFile)
	assert.Contains(t, string(data), "func (r *User) Reset()")

	os.Remove(tmpFile)
	os.Remove(genFile)
}
