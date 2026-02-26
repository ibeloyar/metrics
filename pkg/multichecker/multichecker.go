package main

import (
	"strings"

	"github.com/gordonklaus/ineffassign/pkg/ineffassign"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"honnef.co/go/tools/staticcheck"

	gocritic "github.com/go-critic/go-critic/checkers/analyzer"
)

func main() {
	var checks []*analysis.Analyzer

	// All analyzers SA class (SA1000-SA9999)
	// S1003 - Replace call to strings.Index with strings.Contains
	// ST1000 -	Incorrect or missing package comment
	for _, a := range staticcheck.Analyzers {
		if strings.HasPrefix(a.Analyzer.Name, "SA") {
			checks = append(checks, a.Analyzer)
		}
		if a.Analyzer.Name == "S1003" {
			checks = append(checks, a.Analyzer)
		}
		if a.Analyzer.Name == "ST1000" {
			checks = append(checks, a.Analyzer)
		}
	}

	checks = append(checks,
		printf.Analyzer,
		shadow.Analyzer,
		structtag.Analyzer,
		ineffassign.Analyzer,
		gocritic.Analyzer,
		MainExitAnalyzer,
	)

	multichecker.Main(checks...)
}
