/*
 staticlint implements set of static checks.

 Following checks are included:
	1. All checks from golang.org/x/tools/go/analysis/passes
	2. All SA checks from https://staticcheck.io/docs/checks/
	3. ST1019 check from https://staticcheck.io/docs/checks/#ST1019
	4. Check for database query in loops https://github.com/masibw/goone
	5. Check wrapping errors https://github.com/fatih/errwrap
	6. Check for calling os.Exit in main func of main package


*/

package main

import (
	"go/ast"

	"github.com/fatih/errwrap/errwrap"
	"github.com/masibw/goone"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/asmdecl"
	"golang.org/x/tools/go/analysis/passes/assign"
	"golang.org/x/tools/go/analysis/passes/atomic"
	"golang.org/x/tools/go/analysis/passes/atomicalign"
	"golang.org/x/tools/go/analysis/passes/bools"
	"golang.org/x/tools/go/analysis/passes/buildssa"
	"golang.org/x/tools/go/analysis/passes/buildtag"
	"golang.org/x/tools/go/analysis/passes/cgocall"
	"golang.org/x/tools/go/analysis/passes/composite"
	"golang.org/x/tools/go/analysis/passes/copylock"
	"golang.org/x/tools/go/analysis/passes/ctrlflow"
	"golang.org/x/tools/go/analysis/passes/deepequalerrors"
	"golang.org/x/tools/go/analysis/passes/errorsas"
	"golang.org/x/tools/go/analysis/passes/fieldalignment"
	"golang.org/x/tools/go/analysis/passes/findcall"
	"golang.org/x/tools/go/analysis/passes/framepointer"
	"golang.org/x/tools/go/analysis/passes/httpresponse"
	"golang.org/x/tools/go/analysis/passes/ifaceassert"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/analysis/passes/loopclosure"
	"golang.org/x/tools/go/analysis/passes/lostcancel"
	"golang.org/x/tools/go/analysis/passes/nilfunc"
	"golang.org/x/tools/go/analysis/passes/nilness"
	"golang.org/x/tools/go/analysis/passes/pkgfact"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/reflectvaluecompare"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/shift"
	"golang.org/x/tools/go/analysis/passes/sigchanyzer"
	"golang.org/x/tools/go/analysis/passes/sortslice"
	"golang.org/x/tools/go/analysis/passes/stdmethods"
	"golang.org/x/tools/go/analysis/passes/stringintconv"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"golang.org/x/tools/go/analysis/passes/tests"
	"golang.org/x/tools/go/analysis/passes/unmarshal"
	"golang.org/x/tools/go/analysis/passes/unreachable"
	"golang.org/x/tools/go/analysis/passes/unsafeptr"
	"golang.org/x/tools/go/analysis/passes/unusedresult"
	"golang.org/x/tools/go/analysis/passes/unusedwrite"
	"golang.org/x/tools/go/ast/inspector"
	"honnef.co/go/tools/staticcheck"
	"honnef.co/go/tools/stylecheck"
)

func main() {
	mychecks := make([]*analysis.Analyzer, 0, 245)
	for _, v := range staticcheck.Analyzers {
		mychecks = append(mychecks, v.Analyzer)
	}
	for _, v := range stylecheck.Analyzers {
		if v.Analyzer.Name == "ST1019" {
			mychecks = append(mychecks, v.Analyzer)
		}
	}
	mychecks = appendPasses(mychecks)
	mychecks = append(mychecks, errwrap.Analyzer)
	mychecks = append(mychecks, goone.Analyzer)
	mychecks = append(mychecks, ExitInMainAnalyzer)
	multichecker.Main(mychecks...)
}

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

// func run(pass *analysis.Pass) (interface{}, error) {
// 	isMainPkg := func(x *ast.File) bool {
// 		return x.Name.Name == "main"
// 	}
//
// 	isMainFunc := func(x *ast.FuncDecl) bool {
// 		return x.Name.Name == "main"
// 	}
//
// 	isOsExit := func(x *ast.SelectorExpr, isMain bool) bool {
// 		if !isMain || x.X == nil {
// 			return false
// 		}
// 		ident, ok := x.X.(*ast.Ident)
// 		if !ok {
// 			return false
// 		}
// 		if ident.Name == "os" && x.Sel.Name == "Exit" {
// 			pass.Reportf(x.Sel.NamePos, "os.Exit called in main func in main package")
// 			return true
// 		}
// 		return false
// 	}
//
// 	for _, file := range pass.Files {
// 		mainInspecting := false
// 		// функцией ast.Inspect проходим по всем узлам AST
// 		ast.Inspect(file, func(node ast.Node) bool {
// 			switch x := node.(type) {
// 			case *ast.File: // package
// 				if !isMainPkg(x) { // если пакет на main - выходим
// 					return true
// 				}
// 			case *ast.FuncDecl: // определение функции
// 				f := isMainFunc(x)
// 				if mainInspecting && !f { // если до этого инспектировали main, а теперь нет - можно заканчивать
// 					return true
// 				}
// 				mainInspecting = f
// 			case *ast.SelectorExpr: // вызова функции
// 				if isOsExit(x, mainInspecting) {
// 					return true
// 				}
// 			}
// 			return true
// 		})
// 	}
// 	return nil, nil
// }

func appendPasses(in []*analysis.Analyzer) []*analysis.Analyzer {
	passes := []*analysis.Analyzer{
		asmdecl.Analyzer,
		assign.Analyzer,
		atomic.Analyzer,
		atomicalign.Analyzer,
		bools.Analyzer,
		buildssa.Analyzer,
		buildtag.Analyzer,
		cgocall.Analyzer,
		composite.Analyzer,
		copylock.Analyzer,
		ctrlflow.Analyzer,
		deepequalerrors.Analyzer,
		errorsas.Analyzer,
		fieldalignment.Analyzer,
		findcall.Analyzer,
		framepointer.Analyzer,
		httpresponse.Analyzer,
		ifaceassert.Analyzer,
		inspect.Analyzer,
		loopclosure.Analyzer,
		lostcancel.Analyzer,
		nilfunc.Analyzer,
		nilness.Analyzer,
		pkgfact.Analyzer,
		printf.Analyzer,
		reflectvaluecompare.Analyzer,
		shadow.Analyzer,
		shift.Analyzer,
		sigchanyzer.Analyzer,
		sortslice.Analyzer,
		stdmethods.Analyzer,
		stringintconv.Analyzer,
		structtag.Analyzer,
		tests.Analyzer,
		unmarshal.Analyzer,
		unreachable.Analyzer,
		unsafeptr.Analyzer,
		unusedresult.Analyzer,
		unusedwrite.Analyzer,
	}
	return append(in, passes...)
}
