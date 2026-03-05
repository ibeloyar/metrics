/*
Package main implements a multichecker tool for Go static analysis.

# Overview

This tool combines multiple Go static analyzers into a single vet-compatible binary.
It provides comprehensive code quality checks covering common pitfalls, performance issues,
style violations, and best practices.

In addition, it enforces safer application termination patterns by restricting the use
of panic, os.Exit, and log.Fatal - similar calls outside the main function in the main package.

# Usage

Build the tool:

	go build -o multichecker .

Run with go vet:

	go vet -vettool=./multichecker ./...

The multichecker works by:

1. Analyzer Registration: Dynamically collects analyzers from multiple sources:
all SA* analyzers from staticcheck (SA1000-SA9999), specific staticcheck analyzers (S1003, ST1000),
standard x/tools analyzers (printf, shadow, structtag), third-party analyzers (ineffassign, gocritic, custom main_exit_analyzer)

2. MultiChecker Execution: Uses golang.org/x/tools/go/analysis/multichecker.Main()
which creates a vet-compatible analysis runner that executes all registered analyzers
in parallel across the provided packages.

3. Custom Integration: The tool seamlessly integrates with go vet workflow,
processing the same packages and files that go vet would analyze.

# Included Analyzers

Staticcheck SA* Analyzers (SA1000-SA9999)
Comprehensive set of static analysis checks covering:
  - Performance optimizations (SA1000-SA1999)
  - Code correctness (SA2000-SA2999)
  - Concurrency issues (SA4000-SA4999)
  - Error handling (SA5000-SA5999)
  - Resource management (SA6000-SA6999)
  - Testing improvements (SA7000-SA7999)
  - Build tags (SA8000-SA8999)
  - Deprecated APIs (SA9000-SA9999)

Specific Staticcheck Analyzers
  - S1003: Replace strings.Index with strings.Contains for better readability
  - ST1000: Enforces package documentation comments

Standard x/tools Analyzers
  - printf: Verifies fmt.Printf/Sprintf format strings match arguments
  - shadow: Detects variable shadowing that might hide bugs
  - structtag: Validates struct field tags syntax and usage

Third-party Analyzers
  - ineffassign: Finds ineffective assignments (variables assigned but never used)
  - gocritic: GoCritic ruleset covering 100+ code style and quality checks
  - exitanalizer (custom): Prohibits os.Exit calls outside main() in the main package,
    Prohibits log.Fatal, log.Fatalf, log.Fatalln, log.Panic, log.Panicf,
    log.Panicln calls outside main() in the main package,
    Prohibits direct panic calls outside main() in the main package

The tool provides comprehensive coverage making it suitable for CI/CD pipelines and pre-commit hooks.
*/
package main
