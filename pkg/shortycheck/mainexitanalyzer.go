package shortycheck

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

// MainExitAnalyzer - анализатор, который проверяет наличие вызовов os.Exit в функции main.
var MainExitAnalyzer = &analysis.Analyzer{
	Name: "mainexitanalyzer",
	Doc:  "check for exits in main",
	Run:  run,
}

// run реализует логику анализатора.
func run(pass *analysis.Pass) (interface{}, error) {
	findExit := func(c *ast.CallExpr) {
		if s, ok := c.Fun.(*ast.SelectorExpr); ok {
			if p, ok := s.X.(*ast.Ident); ok && p.Name == "os" && s.Sel.Name == "Exit" {
				pass.Reportf(s.Pos(), "found Exit in main func")
			}
		}
	}

	for _, file := range pass.Files {
		// если пакет - не main, то дочерние элементы не смотрим
		if file.Name.Name != "main" {
			continue
		}
		// функцией ast.Inspect проходим по всем узлам AST
		ast.Inspect(file, func(node ast.Node) bool {
			switch o := node.(type) {
			case *ast.FuncDecl:
				// если функций - не main, то дочерние элементы не смотрим
				if o.Name.Name != "main" {
					return false
				}
			case *ast.CallExpr:
				findExit(o)
			}
			return true
		})
	}
	return nil, nil
}
