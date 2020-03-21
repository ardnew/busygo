package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const baseName = "busygo"

var (
	argf      string
	invokedAs string
)

type function func([]string) ([]string, error)
type functionProp struct {
	name string
	argp string
	arep bool
	retp string
	fref function
}

var (
	functions = []struct {
		pkg  []string
		prop []functionProp
	}{{
		[]string{"path", "filepath"},
		[]functionProp{
			{
				name: "Abs",
				argp: "path string",
				arep: true,
				retp: "string, error",
				fref: filepath_Abs,
			},
			{
				name: "Base",
				argp: "path string",
				arep: false,
				retp: "string",
				fref: filepath_Base,
			},
			{
				name: "Clean",
				argp: "path string",
				arep: false,
				retp: "string",
				fref: filepath_Clean,
			},
			{
				name: "Dir",
				argp: "path string",
				arep: false,
				retp: "string",
				fref: filepath_Dir,
			},
			{
				name: "EvalSymlinks",
				argp: "path string",
				arep: false,
				retp: "string, error",
				fref: filepath_EvalSymlinks,
			},
			{
				name: "Ext",
				argp: "path string",
				arep: false,
				retp: "string",
				fref: filepath_Ext,
			},
			{
				name: "FromSlash",
				argp: "path string",
				arep: false,
				retp: "string",
				fref: filepath_FromSlash,
			},
			{
				name: "Glob",
				argp: "pattern string",
				arep: false,
				retp: "matches []string, err error",
				fref: filepath_Glob,
			},
			{
				name: "HasPrefix",
				argp: "p, prefix string",
				arep: false,
				retp: "bool",
				fref: filepath_HasPrefix,
			},
			{
				name: "IsAbs",
				argp: "path string",
				arep: false,
				retp: "bool",
				fref: filepath_IsAbs,
			},
			{
				name: "Join",
				argp: "elem ...string",
				arep: false,
				retp: "string",
				fref: filepath_Join,
			},
			{
				name: "Match",
				argp: "pattern, name string",
				arep: false,
				retp: "matched bool, err error",
				fref: filepath_Match,
			},
			{
				name: "Rel",
				argp: "basepath, targpath string",
				arep: false,
				retp: "string, error",
				fref: filepath_Rel,
			},
			{
				name: "Split",
				argp: "path string",
				arep: false,
				retp: "dir, file string",
				fref: filepath_Split,
			},
			{
				name: "SplitList",
				argp: "path string",
				arep: false,
				retp: "[]string",
				fref: filepath_SplitList,
			},
			{
				name: "ToSlash",
				argp: "path string",
				arep: false,
				retp: "string",
				fref: filepath_ToSlash,
			},
			{
				name: "VolumeName",
				argp: "path string",
				arep: false,
				retp: "string",
				fref: filepath_VolumeName,
			},
		}},
	}
)

func printUsage(as string) {

	pkg, prop, ok := resolve(as)

	if as == baseName || !ok {
		fmt.Printf("Usage of %s:\n", as)
		flag.PrintDefaults()
		fmt.Printf("\nThe following library functions are supported:\n")
		for _, f := range functions {
			fmt.Printf("\tpackage %q\n", strings.Join(f.pkg, "/"))
			for _, p := range f.prop {
				fmt.Printf("\t\t%s\n", p.String())
			}
		}
	} else {

		name := strings.ToLower(prop.name)
		fmt.Printf("Usage of (%q) %s:\n", pkg, name)
		fmt.Printf("\t%s %s\n", name, strings.Join(prop.args(), " "))
		if len(prop.returns()) > 0 {
			fmt.Printf("\t\treturns: %s\n", strings.Join(prop.returns(), ", "))
		}
	}
}

func main() {

	invokedAs = filepath.Base(os.Args[0])

	if invokedAs == baseName {
		flag.StringVar(&argf, "f", "", "invoke function named `func`")
	}
	flag.CommandLine.Usage = func() { printUsage(invokedAs) }
	flag.Parse()

	fn, fa := invokedAs, os.Args[1:]
	if "" != argf {
		fn, fa = argf, flag.Args()
	}

	_, prop, ok := resolve(fn)
	if !ok {
		printUsage(baseName)
	}
	res, err := prop.fref(fa)
	if nil != err {
		fmt.Fprintf(os.Stderr, "%s\n", err)
	}
	if nil != res {
		if len(res) > 0 {
			fmt.Fprintf(os.Stdout, "%s\n", strings.Join(res, " "))
		}
	}
}

func (p functionProp) String() string {
	retp := p.retp
	if strings.ContainsRune(p.retp, ',') {
		retp = fmt.Sprintf("(%s)", retp)
	}
	return fmt.Sprintf("func %s(%s) %s", p.name, p.argp, retp)
}

func formatList(s string) (string, string) {
	if strings.HasPrefix(s, "[]") || strings.HasPrefix(s, "...") {
		return strings.TrimLeft(s, "[]."), "..."
	}
	return s, ""
}

func formatArg(s string) []string {
	f := strings.Fields(s)
	a := []string{}
	var n, e string
	switch len(f) {
	case 1:
		n, e = formatList(f[0])
	case 2:
		n = f[0]
		_, e = formatList(f[1])
	}
	a = append(a, n)
	if e != "" {
		a = append(a, e)
	}
	return a
}

func (p *functionProp) args() []string {
	arg := []string{}
	for _, a := range strings.Split(p.argp, ", ") {
		f := formatArg(a)
		if len(f) > 0 {
			arg = append(arg, f...)
		}
	}
	if p.arep {
		arg = append(arg, "[...]")
	}
	return arg
}

func (p *functionProp) returns() []string {
	ret := []string{}
	for _, r := range strings.Split(p.retp, ", ") {
		f := formatArg(r)
		if len(f) > 0 {
			ret = append(ret, f...)
		}
	}
	return ret
}

func resolve(s string) (string, *functionProp, bool) {
	var name, pkg string
	path := strings.Split(s, ".")
	switch len(path) {
	case 0:
		return "", nil, false
	case 1:
		name = path[0]
	default:
		nameIndex := len(path) - 1
		name, pkg = path[nameIndex], strings.Join(path[:nameIndex], "/")
	}
	name = strings.ToLower(name)
	pkg = strings.ToLower(pkg)
	for _, f := range functions {
		fpkg := strings.ToLower(strings.Join(f.pkg, "/"))
		if "" == pkg || strings.HasSuffix(fpkg, pkg) {
			for _, p := range f.prop {
				if strings.ToLower(p.name) == name {
					return fpkg, &p, true
				}
			}
		}
	}
	return "", nil, false
}

func filepath_Abs(in []string) ([]string, error) {
	ret := []string{}
	for _, p := range in {
		abs, err := filepath.Abs(p)
		if nil != err {
			return nil, err
		} else {
			ret = append(ret, abs)
		}
	}
	return ret, nil
}

func filepath_Base(in []string) ([]string, error) {
	return []string{}, nil
}

func filepath_Clean(in []string) ([]string, error) {
	return []string{}, nil
}

func filepath_Dir(in []string) ([]string, error) {
	return []string{}, nil
}

func filepath_EvalSymlinks(in []string) ([]string, error) {
	return []string{}, nil
}

func filepath_Ext(in []string) ([]string, error) {
	return []string{}, nil
}

func filepath_FromSlash(in []string) ([]string, error) {
	return []string{}, nil
}

func filepath_Glob(in []string) ([]string, error) {
	return []string{}, nil
}

func filepath_HasPrefix(in []string) ([]string, error) {
	return []string{}, nil
}

func filepath_IsAbs(in []string) ([]string, error) {
	return []string{}, nil
}

func filepath_Join(in []string) ([]string, error) {
	return []string{}, nil
}

func filepath_Match(in []string) ([]string, error) {
	return []string{}, nil
}

func filepath_Rel(in []string) ([]string, error) {
	return []string{}, nil
}

func filepath_Split(in []string) ([]string, error) {
	return []string{}, nil
}

func filepath_SplitList(in []string) ([]string, error) {
	return []string{}, nil
}

func filepath_ToSlash(in []string) ([]string, error) {
	return []string{}, nil
}

func filepath_VolumeName(in []string) ([]string, error) {
	return []string{}, nil
}
