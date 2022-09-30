// This linter checks the call os.Exit function from main and warns if such call will be found.
// It can be called from multichecker.Main(...*analysis.Analyzer) of golang.org/x/tools/go/analysis/multichecker
//  multichecker.Main(linter.ExitInMainAnalyzer)

package linter

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

var ExitInMainAnalyzer = &analysis.Analyzer{
	Name:     "exitinmain",
	Doc:      "check for os.Exit call in func main in package main",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	isMainPkg := func(x *ast.File) bool {
		return x.Name.Name == "main"
	}

	isMainFunc := func(x *ast.FuncDecl) bool {
		return x.Name.Name == "main"
	}

	isOsExit := func(x *ast.SelectorExpr, isMain bool) bool {
		if !isMain || x.X == nil {
			return false
		}
		ident, ok := x.X.(*ast.Ident)
		if !ok {
			return false
		}
		if ident.Name == "os" && x.Sel.Name == "Exit" {
			pass.Reportf(ident.NamePos, "os.Exit called in main func in main package")
			return true
		}
		return false
	}

	i := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	nodeFilter := []ast.Node{
		(*ast.File)(nil),
		(*ast.FuncDecl)(nil),
		(*ast.SelectorExpr)(nil),
	}
	mainInspecting := false
	i.Preorder(nodeFilter, func(n ast.Node) {
		switch x := n.(type) {
		case *ast.File:
			if !isMainPkg(x) { // если пакет на main - выходим
				return
			}
		case *ast.FuncDecl: // определение функции
			f := isMainFunc(x)
			if mainInspecting && !f { // если до этого инспектировали main, а теперь нет - можно заканчивать
				mainInspecting = false
				return
			}
			mainInspecting = f
		case *ast.SelectorExpr:
			if isOsExit(x, mainInspecting) {
				return
			}
		}
	})

	return nil, nil
}
