package main

import (
	"go/ast"
	"go/token"

	"golang.org/x/tools/go/analysis"
)

var MainExitAnalyzer = &analysis.Analyzer{
	Name:     "mainexitanalyzer",
	Doc:      "Disables os.Exit() in the main() function of the main package",
	Run:      run,
	Requires: []*analysis.Analyzer{},
}

func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		if file.Name.Name != "main" {
			continue
		}

		ast.Inspect(file, func(node ast.Node) bool {
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
				if isInMainFunc(file, call.Pos()) {
					pass.Reportf(call.Pos(), "os.Exit() запрещен в main() функции")
				}
			}
			return true
		})
	}
	return nil, nil
}

// isInMainFunc - checks that we are in a main function without return result
func isInMainFunc(file *ast.File, pos token.Pos) bool {
	for _, decl := range file.Decls {
		funcDecl, ok := decl.(*ast.FuncDecl)
		if ok && funcDecl.Name.Name == "main" && funcDecl.Type.Results == nil {
			if pos >= funcDecl.Pos() && pos <= funcDecl.End() {
				return true
			}
		}
	}
	return false
}
