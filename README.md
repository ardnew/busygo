# busygo
Command-line interface to common functions in the Go standard library

# wat
Ever wanted to use a certain Go library function straight from the command-line? Didn't want to make an entire program just to expose it?

Similar to `busybox`, you can symlink a standard library function against `busygo`, and it will invoke that corresponding function with the given command-line arguments. For example, the `Abs` function from package `"path/filepath"`:

```
$ ln -s path/to/busygo abs
$ abs .
/my/current/dir
```

Or if you don't want symlinks, specify the package/function with the `-f` flag:

```
$ busygo -f abs .
/my/current/dir
$ busygo -f path.filepath.abs .
/my/current/dir
```

Available functions and their parameters can be seen with `-h` flag:
```
$ busygo -h
Usage of busygo:
  -f func
        invoke function named func

The following library functions are supported:
        package "path/filepath"
                func Abs(path string) (string, error)
                func Base(path string) string
                func Clean(path string) string
                func Dir(path string) string
                func EvalSymlinks(path string) (string, error)
                func Ext(path string) string
                func FromSlash(path string) string
                func Glob(pattern string) (matches []string, err error)
                func HasPrefix(p, prefix string) bool
                func IsAbs(path string) bool
                func Join(elem ...string) string
                func Match(pattern, name string) (matched bool, err error)
                func Rel(basepath, targpath string) (string, error)
                func Split(path string) (dir, file string)
                func SplitList(path string) []string
                func ToSlash(path string) string
                func VolumeName(path string) string
```

Or for a specific function:

```
$  abs -h
Usage of ("path/filepath") abs:
        abs path [...]
                returns: string, error
```
