package stdlib

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/d5/tengo"
	"github.com/d5/tengo/objects"
)

var osModule = map[string]objects.Object{
	"o_rdonly":            &objects.Int{Value: int64(os.O_RDONLY)},
	"o_wronly":            &objects.Int{Value: int64(os.O_WRONLY)},
	"o_rdwr":              &objects.Int{Value: int64(os.O_RDWR)},
	"o_append":            &objects.Int{Value: int64(os.O_APPEND)},
	"o_create":            &objects.Int{Value: int64(os.O_CREATE)},
	"o_excl":              &objects.Int{Value: int64(os.O_EXCL)},
	"o_sync":              &objects.Int{Value: int64(os.O_SYNC)},
	"o_trunc":             &objects.Int{Value: int64(os.O_TRUNC)},
	"mode_dir":            &objects.Int{Value: int64(os.ModeDir)},
	"mode_append":         &objects.Int{Value: int64(os.ModeAppend)},
	"mode_exclusive":      &objects.Int{Value: int64(os.ModeExclusive)},
	"mode_temporary":      &objects.Int{Value: int64(os.ModeTemporary)},
	"mode_symlink":        &objects.Int{Value: int64(os.ModeSymlink)},
	"mode_device":         &objects.Int{Value: int64(os.ModeDevice)},
	"mode_named_pipe":     &objects.Int{Value: int64(os.ModeNamedPipe)},
	"mode_socket":         &objects.Int{Value: int64(os.ModeSocket)},
	"mode_setuid":         &objects.Int{Value: int64(os.ModeSetuid)},
	"mode_setgui":         &objects.Int{Value: int64(os.ModeSetgid)},
	"mode_char_device":    &objects.Int{Value: int64(os.ModeCharDevice)},
	"mode_sticky":         &objects.Int{Value: int64(os.ModeSticky)},
	"mode_type":           &objects.Int{Value: int64(os.ModeType)},
	"mode_perm":           &objects.Int{Value: int64(os.ModePerm)},
	"path_separator":      &objects.Char{Value: os.PathSeparator},
	"path_list_separator": &objects.Char{Value: os.PathListSeparator},
	"dev_null":            &objects.String{Value: os.DevNull},
	"seek_set":            &objects.Int{Value: int64(io.SeekStart)},
	"seek_cur":            &objects.Int{Value: int64(io.SeekCurrent)},
	"seek_end":            &objects.Int{Value: int64(io.SeekEnd)},
	"args":                &objects.UserFunction{Name: "args", Value: osArgs},                             // args() => array(string)
	"chdir":               &objects.UserFunction{Name: "chdir", Value: FuncASRE(os.Chdir)},                // chdir(dir string) => error
	"chmod":               osFuncASFmRE("chmod", os.Chmod),                                                // chmod(name string, mode int) => error
	"chown":               &objects.UserFunction{Name: "chown", Value: FuncASIIRE(os.Chown)},              // chown(name string, uid int, gid int) => error
	"clearenv":            &objects.UserFunction{Name: "clearenv", Value: FuncAR(os.Clearenv)},            // clearenv()
	"environ":             &objects.UserFunction{Name: "environ", Value: FuncARSs(os.Environ)},            // environ() => array(string)
	"exit":                &objects.UserFunction{Name: "exit", Value: FuncAIR(os.Exit)},                   // exit(code int)
	"expand_env":          &objects.UserFunction{Name: "expand_env", Value: osExpandEnv},                  // expand_env(s string) => string
	"getegid":             &objects.UserFunction{Name: "getegid", Value: FuncARI(os.Getegid)},             // getegid() => int
	"getenv":              &objects.UserFunction{Name: "getenv", Value: FuncASRS(os.Getenv)},              // getenv(s string) => string
	"geteuid":             &objects.UserFunction{Name: "geteuid", Value: FuncARI(os.Geteuid)},             // geteuid() => int
	"getgid":              &objects.UserFunction{Name: "getgid", Value: FuncARI(os.Getgid)},               // getgid() => int
	"getgroups":           &objects.UserFunction{Name: "getgroups", Value: FuncARIsE(os.Getgroups)},       // getgroups() => array(string)/error
	"getpagesize":         &objects.UserFunction{Name: "getpagesize", Value: FuncARI(os.Getpagesize)},     // getpagesize() => int
	"getpid":              &objects.UserFunction{Name: "getpid", Value: FuncARI(os.Getpid)},               // getpid() => int
	"getppid":             &objects.UserFunction{Name: "getppid", Value: FuncARI(os.Getppid)},             // getppid() => int
	"getuid":              &objects.UserFunction{Name: "getuid", Value: FuncARI(os.Getuid)},               // getuid() => int
	"getwd":               &objects.UserFunction{Name: "getwd", Value: FuncARSE(os.Getwd)},                // getwd() => string/error
	"hostname":            &objects.UserFunction{Name: "hostname", Value: FuncARSE(os.Hostname)},          // hostname() => string/error
	"lchown":              &objects.UserFunction{Name: "lchown", Value: FuncASIIRE(os.Lchown)},            // lchown(name string, uid int, gid int) => error
	"link":                &objects.UserFunction{Name: "link", Value: FuncASSRE(os.Link)},                 // link(oldname string, newname string) => error
	"lookup_env":          &objects.UserFunction{Name: "lookup_env", Value: osLookupEnv},                  // lookup_env(key string) => string/false
	"mkdir":               osFuncASFmRE("mkdir", os.Mkdir),                                                // mkdir(name string, perm int) => error
	"mkdir_all":           osFuncASFmRE("mkdir_all", os.MkdirAll),                                         // mkdir_all(name string, perm int) => error
	"readlink":            &objects.UserFunction{Name: "readlink", Value: FuncASRSE(os.Readlink)},         // readlink(name string) => string/error
	"remove":              &objects.UserFunction{Name: "remove", Value: FuncASRE(os.Remove)},              // remove(name string) => error
	"remove_all":          &objects.UserFunction{Name: "remove_all", Value: FuncASRE(os.RemoveAll)},       // remove_all(name string) => error
	"rename":              &objects.UserFunction{Name: "rename", Value: FuncASSRE(os.Rename)},             // rename(oldpath string, newpath string) => error
	"setenv":              &objects.UserFunction{Name: "setenv", Value: FuncASSRE(os.Setenv)},             // setenv(key string, value string) => error
	"symlink":             &objects.UserFunction{Name: "symlink", Value: FuncASSRE(os.Symlink)},           // symlink(oldname string newname string) => error
	"temp_dir":            &objects.UserFunction{Name: "temp_dir", Value: FuncARS(os.TempDir)},            // temp_dir() => string
	"truncate":            &objects.UserFunction{Name: "truncate", Value: FuncASI64RE(os.Truncate)},       // truncate(name string, size int) => error
	"unsetenv":            &objects.UserFunction{Name: "unsetenv", Value: FuncASRE(os.Unsetenv)},          // unsetenv(key string) => error
	"create":              &objects.UserFunction{Name: "create", Value: osCreate},                         // create(name string) => imap(file)/error
	"open":                &objects.UserFunction{Name: "open", Value: osOpen},                             // open(name string) => imap(file)/error
	"open_file":           &objects.UserFunction{Name: "open_file", Value: osOpenFile},                    // open_file(name string, flag int, perm int) => imap(file)/error
	"find_process":        &objects.UserFunction{Name: "find_process", Value: osFindProcess},              // find_process(pid int) => imap(process)/error
	"start_process":       &objects.UserFunction{Name: "start_process", Value: osStartProcess},            // start_process(name string, argv array(string), dir string, env array(string)) => imap(process)/error
	"exec_look_path":      &objects.UserFunction{Name: "exec_look_path", Value: FuncASRSE(exec.LookPath)}, // exec_look_path(file) => string/error
	"exec":                &objects.UserFunction{Name: "exec", Value: osExec},                             // exec(name, args...) => command
	"stat":                &objects.UserFunction{Name: "stat", Value: osStat},                             // stat(name) => imap(fileinfo)/error
	"read_file":           &objects.UserFunction{Name: "read_file", Value: osReadFile},                    // readfile(name) => array(byte)/error
}

func osReadFile(args ...objects.Object) (ret objects.Object, err error) {
	if len(args) != 1 {
		return nil, objects.ErrWrongNumArguments
	}

	fname, ok := objects.ToString(args[0])
	if !ok {
		return nil, objects.ErrInvalidArgumentType{
			Name:     "first",
			Expected: "string(compatible)",
			Found:    args[0].TypeName(),
		}
	}

	bytes, err := ioutil.ReadFile(fname)
	if err != nil {
		return wrapError(err), nil
	}

	if len(bytes) > tengo.MaxBytesLen {
		return nil, objects.ErrBytesLimit
	}

	return &objects.Bytes{Value: bytes}, nil
}

func osStat(args ...objects.Object) (ret objects.Object, err error) {
	if len(args) != 1 {
		return nil, objects.ErrWrongNumArguments
	}

	fname, ok := objects.ToString(args[0])
	if !ok {
		return nil, objects.ErrInvalidArgumentType{
			Name:     "first",
			Expected: "string(compatible)",
			Found:    args[0].TypeName(),
		}
	}

	stat, err := os.Stat(fname)
	if err != nil {
		return wrapError(err), nil
	}

	fstat := &objects.ImmutableMap{
		Value: map[string]objects.Object{
			"name":  &objects.String{Value: stat.Name()},
			"mtime": &objects.Time{Value: stat.ModTime()},
			"size":  &objects.Int{Value: stat.Size()},
			"mode":  &objects.Int{Value: int64(stat.Mode())},
		},
	}

	if stat.IsDir() {
		fstat.Value["directory"] = objects.TrueValue
	} else {
		fstat.Value["directory"] = objects.FalseValue
	}

	return fstat, nil
}

func osCreate(args ...objects.Object) (objects.Object, error) {
	if len(args) != 1 {
		return nil, objects.ErrWrongNumArguments
	}

	s1, ok := objects.ToString(args[0])
	if !ok {
		return nil, objects.ErrInvalidArgumentType{
			Name:     "first",
			Expected: "string(compatible)",
			Found:    args[0].TypeName(),
		}
	}

	res, err := os.Create(s1)
	if err != nil {
		return wrapError(err), nil
	}

	return makeOSFile(res), nil
}

func osOpen(args ...objects.Object) (objects.Object, error) {
	if len(args) != 1 {
		return nil, objects.ErrWrongNumArguments
	}

	s1, ok := objects.ToString(args[0])
	if !ok {
		return nil, objects.ErrInvalidArgumentType{
			Name:     "first",
			Expected: "string(compatible)",
			Found:    args[0].TypeName(),
		}
	}

	res, err := os.Open(s1)
	if err != nil {
		return wrapError(err), nil
	}

	return makeOSFile(res), nil
}

func osOpenFile(args ...objects.Object) (objects.Object, error) {
	if len(args) != 3 {
		return nil, objects.ErrWrongNumArguments
	}

	s1, ok := objects.ToString(args[0])
	if !ok {
		return nil, objects.ErrInvalidArgumentType{
			Name:     "first",
			Expected: "string(compatible)",
			Found:    args[0].TypeName(),
		}
	}

	i2, ok := objects.ToInt(args[1])
	if !ok {
		return nil, objects.ErrInvalidArgumentType{
			Name:     "second",
			Expected: "int(compatible)",
			Found:    args[1].TypeName(),
		}
	}

	i3, ok := objects.ToInt(args[2])
	if !ok {
		return nil, objects.ErrInvalidArgumentType{
			Name:     "third",
			Expected: "int(compatible)",
			Found:    args[2].TypeName(),
		}
	}

	res, err := os.OpenFile(s1, i2, os.FileMode(i3))
	if err != nil {
		return wrapError(err), nil
	}

	return makeOSFile(res), nil
}

func osArgs(args ...objects.Object) (objects.Object, error) {
	if len(args) != 0 {
		return nil, objects.ErrWrongNumArguments
	}

	arr := &objects.Array{}
	for _, osArg := range os.Args {
		if len(osArg) > tengo.MaxStringLen {
			return nil, objects.ErrStringLimit
		}

		arr.Value = append(arr.Value, &objects.String{Value: osArg})
	}

	return arr, nil
}

func osFuncASFmRE(name string, fn func(string, os.FileMode) error) *objects.UserFunction {
	return &objects.UserFunction{
		Name: name,
		Value: func(args ...objects.Object) (objects.Object, error) {
			if len(args) != 2 {
				return nil, objects.ErrWrongNumArguments
			}

			s1, ok := objects.ToString(args[0])
			if !ok {
				return nil, objects.ErrInvalidArgumentType{
					Name:     "first",
					Expected: "string(compatible)",
					Found:    args[0].TypeName(),
				}
			}
			i2, ok := objects.ToInt64(args[1])
			if !ok {
				return nil, objects.ErrInvalidArgumentType{
					Name:     "second",
					Expected: "int(compatible)",
					Found:    args[1].TypeName(),
				}
			}

			return wrapError(fn(s1, os.FileMode(i2))), nil
		},
	}
}

func osLookupEnv(args ...objects.Object) (objects.Object, error) {
	if len(args) != 1 {
		return nil, objects.ErrWrongNumArguments
	}

	s1, ok := objects.ToString(args[0])
	if !ok {
		return nil, objects.ErrInvalidArgumentType{
			Name:     "first",
			Expected: "string(compatible)",
			Found:    args[0].TypeName(),
		}
	}

	res, ok := os.LookupEnv(s1)
	if !ok {
		return objects.FalseValue, nil
	}

	if len(res) > tengo.MaxStringLen {
		return nil, objects.ErrStringLimit
	}

	return &objects.String{Value: res}, nil
}

func osExpandEnv(args ...objects.Object) (objects.Object, error) {
	if len(args) != 1 {
		return nil, objects.ErrWrongNumArguments
	}

	s1, ok := objects.ToString(args[0])
	if !ok {
		return nil, objects.ErrInvalidArgumentType{
			Name:     "first",
			Expected: "string(compatible)",
			Found:    args[0].TypeName(),
		}
	}

	var vlen int
	var failed bool
	s := os.Expand(s1, func(k string) string {
		if failed {
			return ""
		}

		v := os.Getenv(k)

		// this does not count the other texts that are not being replaced
		// but the code checks the final length at the end
		vlen += len(v)
		if vlen > tengo.MaxStringLen {
			failed = true
			return ""
		}

		return v
	})

	if failed || len(s) > tengo.MaxStringLen {
		return nil, objects.ErrStringLimit
	}

	return &objects.String{Value: s}, nil
}

func osExec(args ...objects.Object) (objects.Object, error) {
	if len(args) == 0 {
		return nil, objects.ErrWrongNumArguments
	}

	name, ok := objects.ToString(args[0])
	if !ok {
		return nil, objects.ErrInvalidArgumentType{
			Name:     "first",
			Expected: "string(compatible)",
			Found:    args[0].TypeName(),
		}
	}

	var execArgs []string
	for idx, arg := range args[1:] {
		execArg, ok := objects.ToString(arg)
		if !ok {
			return nil, objects.ErrInvalidArgumentType{
				Name:     fmt.Sprintf("args[%d]", idx),
				Expected: "string(compatible)",
				Found:    args[1+idx].TypeName(),
			}
		}

		execArgs = append(execArgs, execArg)
	}

	return makeOSExecCommand(exec.Command(name, execArgs...)), nil
}

func osFindProcess(args ...objects.Object) (objects.Object, error) {
	if len(args) != 1 {
		return nil, objects.ErrWrongNumArguments
	}

	i1, ok := objects.ToInt(args[0])
	if !ok {
		return nil, objects.ErrInvalidArgumentType{
			Name:     "first",
			Expected: "int(compatible)",
			Found:    args[0].TypeName(),
		}
	}

	proc, err := os.FindProcess(i1)
	if err != nil {
		return wrapError(err), nil
	}

	return makeOSProcess(proc), nil
}

func osStartProcess(args ...objects.Object) (objects.Object, error) {
	if len(args) != 4 {
		return nil, objects.ErrWrongNumArguments
	}

	name, ok := objects.ToString(args[0])
	if !ok {
		return nil, objects.ErrInvalidArgumentType{
			Name:     "first",
			Expected: "string(compatible)",
			Found:    args[0].TypeName(),
		}
	}

	var argv []string
	var err error
	switch arg1 := args[1].(type) {
	case *objects.Array:
		argv, err = stringArray(arg1.Value, "second")
		if err != nil {
			return nil, err
		}
	case *objects.ImmutableArray:
		argv, err = stringArray(arg1.Value, "second")
		if err != nil {
			return nil, err
		}
	default:
		return nil, objects.ErrInvalidArgumentType{
			Name:     "second",
			Expected: "array",
			Found:    arg1.TypeName(),
		}
	}

	dir, ok := objects.ToString(args[2])
	if !ok {
		return nil, objects.ErrInvalidArgumentType{
			Name:     "third",
			Expected: "string(compatible)",
			Found:    args[2].TypeName(),
		}
	}

	var env []string
	switch arg3 := args[3].(type) {
	case *objects.Array:
		env, err = stringArray(arg3.Value, "fourth")
		if err != nil {
			return nil, err
		}
	case *objects.ImmutableArray:
		env, err = stringArray(arg3.Value, "fourth")
		if err != nil {
			return nil, err
		}
	default:
		return nil, objects.ErrInvalidArgumentType{
			Name:     "fourth",
			Expected: "array",
			Found:    arg3.TypeName(),
		}
	}

	proc, err := os.StartProcess(name, argv, &os.ProcAttr{
		Dir: dir,
		Env: env,
	})
	if err != nil {
		return wrapError(err), nil
	}

	return makeOSProcess(proc), nil
}

func stringArray(arr []objects.Object, argName string) ([]string, error) {
	var sarr []string
	for idx, elem := range arr {
		str, ok := elem.(*objects.String)
		if !ok {
			return nil, objects.ErrInvalidArgumentType{
				Name:     fmt.Sprintf("%s[%d]", argName, idx),
				Expected: "string",
				Found:    elem.TypeName(),
			}
		}

		sarr = append(sarr, str.Value)
	}

	return sarr, nil
}
