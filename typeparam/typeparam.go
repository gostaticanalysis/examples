package typeparam

import (
	"fmt"
	"go/ast"
	"go/types"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

const doc = "typeparam is ..."

// Analyzer is ...
var Analyzer = &analysis.Analyzer{
	Name: "typeparam",
	Doc:  doc,
	Run:  run,
	Requires: []*analysis.Analyzer{
		inspect.Analyzer,
	},
}

func run(pass *analysis.Pass) (interface{}, error) {
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	ast.Print(pass.Fset, pass.Files[0])

	inspect.Preorder(nil, func(n ast.Node) {
		switch n := n.(type) {
		case *ast.FuncType:
			fmt.Println("FuncType")
			printTypeParams(pass, "  ", n.TypeParams)
		case *ast.TypeSpec:
			fmt.Println("TypeSpec")
			printTypeParams(pass, "  ", n.TypeParams)
			typ := pass.TypesInfo.TypeOf(n.Name)
			instyp, err := types.Instantiate(types.NewContext(), typ, []types.Type{types.Typ[types.Int]}, true)
			fmt.Printf("    Instantiate with int: %v %v\n", instyp, err)
		case *ast.CallExpr:
			fmt.Println("CallExpr")
			printInstance(pass, "  ", n.Fun)
		case *ast.IndexExpr:
			fmt.Println("IndexExpr")
			printInstance(pass, "  ", n.X)
		case *ast.IndexListExpr:
			fmt.Println("IndexListExpr")
			printInstance(pass, "  ", n.X)
		}
	})

	return nil, nil
}

func printTypeParams(pass *analysis.Pass, prefix string, typeParams *ast.FieldList) {
	if typeParams == nil {
		return
	}

	for _, tp := range typeParams.List {
		names := make([]string, len(tp.Names))
		for i := range tp.Names {
			names[i] = tp.Names[i].Name
		}
		typ := pass.TypesInfo.TypeOf(tp.Type)
		fmt.Printf("%s%s %s (%T)\n", prefix, strings.Join(names, ","), typ, typ)

		for i := range tp.Names {
			fmt.Printf("%s%s\n", prefix, tp.Names[i].Name)
			typ, _ := pass.TypesInfo.TypeOf(tp.Names[i]).(*types.TypeParam)
			fmt.Printf("%s  Constraint: %T\n", prefix, typ.Constraint())
			fmt.Printf("%s  %T underlying %T\n", prefix, typ, typ.Underlying())
			constraint, _ := typ.Constraint().(*types.Interface)
			fmt.Printf("%s  Comparable: %v\n", prefix, constraint.IsComparable())
			fmt.Printf("%s  Implicit: %v\n", prefix, constraint.IsImplicit())
			fmt.Printf("%s  IsMethodSet: %v\n", prefix, constraint.IsMethodSet())		
		}
	}
}

func printInstance(pass *analysis.Pass, prefix string, expr ast.Expr) {
	id, _ := expr.(*ast.Ident)
	ins, ok := pass.TypesInfo.Instances[id]
	if !ok {
		return
	}
	fmt.Println(prefix, ins)
}
