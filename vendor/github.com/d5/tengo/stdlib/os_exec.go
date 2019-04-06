package stdlib

import (
	"os/exec"

	"github.com/d5/tengo/objects"
)

func makeOSExecCommand(cmd *exec.Cmd) *objects.ImmutableMap {
	return &objects.ImmutableMap{
		Value: map[string]objects.Object{
			// combined_output() => bytes/error
			"combined_output": &objects.UserFunction{Name: "combined_output", Value: FuncARYE(cmd.CombinedOutput)}, //
			// output() => bytes/error
			"output": &objects.UserFunction{Name: "output", Value: FuncARYE(cmd.Output)}, //
			// run() => error
			"run": &objects.UserFunction{Name: "run", Value: FuncARE(cmd.Run)}, //
			// start() => error
			"start": &objects.UserFunction{Name: "start", Value: FuncARE(cmd.Start)}, //
			// wait() => error
			"wait": &objects.UserFunction{Name: "wait", Value: FuncARE(cmd.Wait)}, //
			// set_path(path string)
			"set_path": &objects.UserFunction{
				Name: "set_path",
				Value: func(args ...objects.Object) (ret objects.Object, err error) {
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

					cmd.Path = s1

					return objects.UndefinedValue, nil
				},
			},
			// set_dir(dir string)
			"set_dir": &objects.UserFunction{
				Name: "set_dir",
				Value: func(args ...objects.Object) (ret objects.Object, err error) {
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

					cmd.Dir = s1

					return objects.UndefinedValue, nil
				},
			},
			// set_env(env array(string))
			"set_env": &objects.UserFunction{
				Name: "set_env",
				Value: func(args ...objects.Object) (objects.Object, error) {
					if len(args) != 1 {
						return nil, objects.ErrWrongNumArguments
					}

					var env []string
					var err error
					switch arg0 := args[0].(type) {
					case *objects.Array:
						env, err = stringArray(arg0.Value, "first")
						if err != nil {
							return nil, err
						}
					case *objects.ImmutableArray:
						env, err = stringArray(arg0.Value, "first")
						if err != nil {
							return nil, err
						}
					default:
						return nil, objects.ErrInvalidArgumentType{
							Name:     "first",
							Expected: "array",
							Found:    arg0.TypeName(),
						}
					}

					cmd.Env = env

					return objects.UndefinedValue, nil
				},
			},
			// process() => imap(process)
			"process": &objects.UserFunction{
				Name: "process",
				Value: func(args ...objects.Object) (ret objects.Object, err error) {
					if len(args) != 0 {
						return nil, objects.ErrWrongNumArguments
					}

					return makeOSProcess(cmd.Process), nil
				},
			},
		},
	}
}
