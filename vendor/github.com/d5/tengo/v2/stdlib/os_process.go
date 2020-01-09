package stdlib

import (
	"os"
	"syscall"

	"github.com/d5/tengo/v2"
)

func makeOSProcessState(state *os.ProcessState) *tengo.ImmutableMap {
	return &tengo.ImmutableMap{
		Value: map[string]tengo.Object{
			"exited": &tengo.UserFunction{
				Name:  "exited",
				Value: FuncARB(state.Exited),
			},
			"pid": &tengo.UserFunction{
				Name:  "pid",
				Value: FuncARI(state.Pid),
			},
			"string": &tengo.UserFunction{
				Name:  "string",
				Value: FuncARS(state.String),
			},
			"success": &tengo.UserFunction{
				Name:  "success",
				Value: FuncARB(state.Success),
			},
		},
	}
}

func makeOSProcess(proc *os.Process) *tengo.ImmutableMap {
	return &tengo.ImmutableMap{
		Value: map[string]tengo.Object{
			"kill": &tengo.UserFunction{
				Name:  "kill",
				Value: FuncARE(proc.Kill),
			},
			"release": &tengo.UserFunction{
				Name:  "release",
				Value: FuncARE(proc.Release),
			},
			"signal": &tengo.UserFunction{
				Name: "signal",
				Value: func(args ...tengo.Object) (tengo.Object, error) {
					if len(args) != 1 {
						return nil, tengo.ErrWrongNumArguments
					}
					i1, ok := tengo.ToInt64(args[0])
					if !ok {
						return nil, tengo.ErrInvalidArgumentType{
							Name:     "first",
							Expected: "int(compatible)",
							Found:    args[0].TypeName(),
						}
					}
					return wrapError(proc.Signal(syscall.Signal(i1))), nil
				},
			},
			"wait": &tengo.UserFunction{
				Name: "wait",
				Value: func(args ...tengo.Object) (tengo.Object, error) {
					if len(args) != 0 {
						return nil, tengo.ErrWrongNumArguments
					}
					state, err := proc.Wait()
					if err != nil {
						return wrapError(err), nil
					}
					return makeOSProcessState(state), nil
				},
			},
		},
	}
}
