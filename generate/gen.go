// busygo exposes functions from the Go standard library through a single
// standalone executable for individual use directly from the command line.
package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"unicode"
)

// May be overridden using -pkg flags. Provide the -pkg flag multiple times to
// specify multiple packages (e.g. "-pkg 'the/first' -pkg 'the/second' ...").

// defaultPkg defines which Go standard library packages to generate supporting
// interfaces for if the -pkg flag is not provided.
var defaultPkg = PackageList{
	"path/filepath",
}

type variadic int

const (
	none variadic = iota
	array
	ellipses
)

func (v variadic) String() string {
	switch v {
	case array:
		return "array"
	case ellipses:
		return "ellipses"
	}
	return "none"
}

type argument struct {
	list variadic
	kind string
}

type returns struct {
	list variadic
	kind string
}

type function struct {
	name string
	args []argument
	rets []returns
}

type funcVisitor struct {
	funcs []function
}

func (v *funcVisitor) Visit(n ast.Node) ast.Visitor {
	if nil == n {
		return nil
	}
	switch o := n.(type) {
	case *ast.FuncDecl:
		if o.Name.IsExported() {

			cf := function{
				name: o.Name.Name,
				args: []argument{},
				rets: []returns{},
			}

			if nil != o.Type.Params {
				for _, f := range o.Type.Params.List {
					switch t := f.Type.(type) {
					case *ast.Ident:
						cf.args = append(cf.args, argument{list: none, kind: t.Name})
					case *ast.ArrayType:
						switch e := t.Elt.(type) {
						case *ast.Ident:
							cf.args = append(cf.args, argument{list: array, kind: e.Name})
						}
					case *ast.Ellipsis:
						switch e := t.Elt.(type) {
						case *ast.Ident:
							cf.args = append(cf.args, argument{list: ellipses, kind: e.Name})
						}
					}
				}
			}

			if nil != o.Type.Results {
				for _, f := range o.Type.Results.List {
					fmt.Printf("%q: %#v\n", o.Name.Name, f)
					switch t := f.Type.(type) {
					case *ast.Ident:
						cf.rets = append(cf.rets, returns{list: none, kind: t.Name})
					case *ast.ArrayType:
						switch e := t.Elt.(type) {
						case *ast.Ident:
							cf.rets = append(cf.rets, returns{list: array, kind: e.Name})
						}
					}
				}
			}

			v.funcs = append(v.funcs, cf)
		}
	}
	return v
}

func main() {

	var (
		argRoot string
		argPkg  PackageList
	)

	flag.StringVar(&argRoot, "root", runtime.GOROOT(), "path to GOROOT (must contain src)")
	flag.Var(&argPkg, "pkg", "generate interfaces for functions from package `path`. may be specified multiple times.")
	flag.Parse()

	if len(argPkg) == 0 {
		argPkg = append(argPkg, defaultPkg...)
	}

	if err := argPkg.parse(argRoot, "src"); err != nil {
		log.Fatalf("error: %+v", err)
	}
}

// fileExists returns whether or not a file exists, and if it exists whether or
// not it is a directory.
func fileExists(path string) (exists, isDir bool) {

	stat, err := os.Stat(path)
	exists = err == nil || !os.IsNotExist(err)
	isDir = exists && stat.IsDir()
	return
}

// PackageList represents packages to parse and generate interfaces for.
type PackageList []string

// String constructs a descriptive representation of a PackageList.
func (p *PackageList) String() string {

	q := []string{}
	for _, s := range *p {
		q = append(q, fmt.Sprintf("%q", s))
	}
	return fmt.Sprintf("[%s]", strings.Join(q, ", "))
}

// Set implements the flag.Value interface to parse packages from -pkg flags.
func (p *PackageList) Set(value string) error {

	// basic sanity tests, not spec-correct
	validPackageRune := func(c rune) bool {
		return unicode.IsLetter(c) || unicode.IsDigit(c) || c == '/'
	}
	validPackage := func(s string) bool {
		for _, c := range s {
			if !validPackageRune(c) {
				return false
			}
		}
		return true
	}

	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("(empty)")
	} else if !validPackage(value) {
		return fmt.Errorf("package name: %q", value)
	}

	for _, k := range *p {
		if k == value {
			return fmt.Errorf("duplicate name: %q", value)
		}
	}

	*p = append(*p, value)

	return nil
}

// withPrefix prepends each path in the package list with the given prefix
// strings, returning the resulting slice of strings.
func (p *PackageList) withPrefix(prefix ...string) []string {

	q := []string{}
	x := filepath.Join(prefix...)
	for _, s := range *p {
		q = append(q, filepath.Join(x, s))
	}
	return q
}

func (p *PackageList) parse(prefix ...string) error {

	path := p.withPrefix(prefix...)
	for _, dir := range path {
		if _, isDir := fileExists(dir); !isDir {
			return fmt.Errorf("invalid package source directory: %q", dir)
		}
	}

	const testFileSuffix = "_test.go"

	fset := token.NewFileSet()

	for _, dir := range path {

		var v = funcVisitor{
			funcs: []function{},
		}

		pkg, err := parser.ParseDir(fset, dir,
			func(info os.FileInfo) bool {
				return !strings.HasSuffix(info.Name(), testFileSuffix)
			}, 0)
		if nil != err {
			return fmt.Errorf("failed to parse: %q: %+v\n", dir, err)
		}

		for _, pkgNode := range pkg {
			ast.Walk(&v, pkgNode)
		}

		for _, f := range v.funcs {
			fmt.Printf("%+v:\n", f)
		}
	}

	return nil
}
