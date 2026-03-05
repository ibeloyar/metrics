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

			if ident, okIdent := call.Fun.(*ast.Ident); okIdent && ident.Name == "panic" {
				if !isInMainFunc(file, call.Pos()) {
					pass.Reportf(call.Pos(), "panic() запрещен вне main() функции")
				}
				return true
			}

			sel, ok := call.Fun.(*ast.SelectorExpr)
			if !ok {
				return true
			}

			xIdent, ok := sel.X.(*ast.Ident)
			if !ok {
				return true
			}

			switch xIdent.Name {
			case "os":
				if sel.Sel.Name == "Exit" {
					if !isInMainFunc(file, call.Pos()) {
						pass.Reportf(call.Pos(), "os.Exit() запрещен вне main() функции")
					}
				}
			case "log":
				if sel.Sel.Name == "Fatal" || sel.Sel.Name == "Fatalf" || sel.Sel.Name == "Fatalln" ||
					sel.Sel.Name == "Panic" || sel.Sel.Name == "Panicf" || sel.Sel.Name == "Panicln" {
					if !isInMainFunc(file, call.Pos()) {
						pass.Reportf(call.Pos(), "log.%s() запрещен вне main() функции", sel.Sel.Name)
					}
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
