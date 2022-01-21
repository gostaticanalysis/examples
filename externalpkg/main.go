package main

import (
	"errors"
	"fmt"
	"go/build"
	"go/token"
	"go/types"
	"os"
	"os/exec"
	"strings"

	"github.com/otiai10/copy"
	"golang.org/x/tools/go/packages"
)

func main() {
	if len(os.Args) <= 2 {
		fmt.Fprintln(os.Stderr, "path and patterns must be specified")
		os.Exit(1)
	}

	path := os.Args[1]
	patterns := os.Args[2:]
	pkgs, err := Load(path, patterns...)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}

	fmt.Println(inDeps(path, pkgs))
}

func Load(path string, patterns ...string) ([]*packages.Package, error) {
	cfg := &packages.Config{
		Mode: packages.NeedTypes | packages.NeedSyntax | packages.NeedTypesInfo | packages.NeedModule,
	}

	pkgs, err := packages.Load(cfg, patterns...)
	if err != nil {
		return nil, err
	}

	if inDeps(path, pkgs) {
		return pkgs, nil
	}

	var mod *packages.Module
	for _, pkg := range pkgs {
		if mod != nil && mod != pkg.Module {
			return nil, errors.New("does not support multi modules")
		}
		mod = pkg.Module
	}

	if mod == nil {
		return nil, errors.New("There are not any modules. The target packages might be not managed by GoModules or the package is standard package.")
	}

	return loadByMod(path, mod, pkgs, patterns)
}

func inDeps(path string, pkgs []*packages.Package) bool {
	var in func(pkg *types.Package) bool
	done := make(map[*types.Package]bool)
	in = func(pkg *types.Package) bool {
		if done[pkg] {
			return false
		}
		done[pkg] = true

		if pkg.Path() == path {
			return true
		}

		for _, p := range pkg.Imports() {
			return in(p)
		}

		return false
	}

	for _, pkg := range pkgs {
		if in(pkg.Types) {
			return true
		}
	}

	return false
}

func loadByMod(path string, mod *packages.Module, orgPkgs []*packages.Package, patterns []string) ([]*packages.Package, error) {
	name := strings.ReplaceAll(mod.Path, "/", "-")
	tmp, err := os.MkdirTemp("", "external-"+name+"-*")
	if err != nil {
		return nil, err
	}

	if err := copy.Copy(mod.Dir, tmp); err != nil {
		return nil, err
	}	

	// import only
	// TODO: remove import only file from analyzed result
	for _, pkg := range orgPkgs {
		f, err := os.CreateTemp(tmp, "imports-*.go") 
		if err != nil {
			return nil, err
		}
		fmt.Fprintln(f, "package", pkg.Types.Name())
		fmt.Fprintf(f, "import _ %q\n", path)
		if err := f.Close(); err != nil {
			return nil, err
		}
	}

	// Execute go get command and add a dependency to go.mod
	cmd := exec.Command("go", "get", path)
	cmd.Dir = tmp
	if err := cmd.Run(); err != nil {
		return nil, err
	}

	cfg := &packages.Config{
		Mode: packages.NeedTypes | packages.NeedSyntax | packages.NeedTypesInfo | packages.NeedModule,
		Dir:  tmp,
	}

	pkgs, err := packages.Load(cfg, patterns...)
	if err != nil {
		return nil, err
	}

	// replace filepath from temp dir to original path
	for _, pkg := range pkgs {
		pkg.Module.Dir = mod.Dir
		fset := token.NewFileSet()
		var rerr error
		pkg.Fset.Iterate(func(f *token.File) bool {
			orgFile := strings.ReplaceAll(f.Name(), "$GOROOT", build.Default.GOROOT)
			content, err := os.ReadFile(orgFile)
			if err != nil {
				rerr = err
				return false
			}

			filename := f.Name()
			if strings.HasPrefix(filename, tmp) {
				filename = mod.Dir + filename[len(tmp):]
			}
			fset.AddFile(filename, f.Base(), f.Size()).SetLinesForContent(content)
			return true
		})
		if rerr != nil {
			return nil, rerr
		}
		pkg.Fset = fset
	}

	return pkgs, nil
}
