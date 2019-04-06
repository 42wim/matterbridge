package stdlib

import (
	"os"

	"github.com/d5/tengo/objects"
)

func makeOSFile(file *os.File) *objects.ImmutableMap {
	return &objects.ImmutableMap{
		Value: map[string]objects.Object{
			// chdir() => true/error
			"chdir": &objects.UserFunction{Name: "chdir", Value: FuncARE(file.Chdir)}, //
			// chown(uid int, gid int) => true/error
			"chown": &objects.UserFunction{Name: "chown", Value: FuncAIIRE(file.Chown)}, //
			// close() => error
			"close": &objects.UserFunction{Name: "close", Value: FuncARE(file.Close)}, //
			// name() => string
			"name": &objects.UserFunction{Name: "name", Value: FuncARS(file.Name)}, //
			// readdirnames(n int) => array(string)/error
			"readdirnames": &objects.UserFunction{Name: "readdirnames", Value: FuncAIRSsE(file.Readdirnames)}, //
			// sync() => error
			"sync": &objects.UserFunction{Name: "sync", Value: FuncARE(file.Sync)}, //
			// write(bytes) => int/error
			"write": &objects.UserFunction{Name: "write", Value: FuncAYRIE(file.Write)}, //
			// write(string) => int/error
			"write_string": &objects.UserFunction{Name: "write_string", Value: FuncASRIE(file.WriteString)}, //
			// read(bytes) => int/error
			"read": &objects.UserFunction{Name: "read", Value: FuncAYRIE(file.Read)}, //
			// chmod(mode int) => error
			"chmod": &objects.UserFunction{
				Name: "chmod",
				Value: func(args ...objects.Object) (ret objects.Object, err error) {
					if len(args) != 1 {
						return nil, objects.ErrWrongNumArguments
					}

					i1, ok := objects.ToInt64(args[0])
					if !ok {
						return nil, objects.ErrInvalidArgumentType{
							Name:     "first",
							Expected: "int(compatible)",
							Found:    args[0].TypeName(),
						}
					}

					return wrapError(file.Chmod(os.FileMode(i1))), nil
				},
			},
			// seek(offset int, whence int) => int/error
			"seek": &objects.UserFunction{
				Name: "seek",
				Value: func(args ...objects.Object) (ret objects.Object, err error) {
					if len(args) != 2 {
						return nil, objects.ErrWrongNumArguments
					}

					i1, ok := objects.ToInt64(args[0])
					if !ok {
						return nil, objects.ErrInvalidArgumentType{
							Name:     "first",
							Expected: "int(compatible)",
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

					res, err := file.Seek(i1, i2)
					if err != nil {
						return wrapError(err), nil
					}

					return &objects.Int{Value: res}, nil
				},
			},
			// stat() => imap(fileinfo)/error
			"stat": &objects.UserFunction{
				Name: "start",
				Value: func(args ...objects.Object) (ret objects.Object, err error) {
					if len(args) != 0 {
						return nil, objects.ErrWrongNumArguments
					}

					return osStat(&objects.String{Value: file.Name()})
				},
			},
		},
	}
}
