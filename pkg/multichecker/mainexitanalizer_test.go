package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"
)

func TestAnalyzerRun(t *testing.T) {
	tests := []struct {
		name    string
		code    string
		wantErr bool
	}{
		{
			name: "os.Exit in main",
			code: `package main
import "os"
func main() {
    os.Exit(1)
}`,
			wantErr: true,
		},
		{
			name: "os.Exit not in main",
			code: `package main
import "os"
func foo() { os.Exit(1) }
func main() {}`,
			wantErr: false,
		},
		{
			name: "main with return",
			code: `package main
import "os"
func main() int { os.Exit(1); return 0 }`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, "main.go", tt.code, 0)
			if err != nil {
				t.Fatal(err)
			}

			for _, f := range []*ast.File{file} {
				if f.Name.Name != "main" {
					continue
				}
				ast.Inspect(f, func(node ast.Node) bool {
					call, ok := node.(*ast.CallExpr)
					if !ok {
						return true
					}
					sel, ok := call.Fun.(*ast.SelectorExpr)
					if !ok {
						return true
					}
					ident, ok := sel.X.(*ast.Ident)
					if !ok || ident.Name != "os" {
						return true
					}
					if sel.Sel.Name == "Exit" {
						if isInMainFunc(f, call.Pos()) {
							t.Log("найден os.Exit в main")
						}
					}
					return true
				})
			}
		})
	}
}
