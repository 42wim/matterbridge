package stdlib

import (
	"os"
	"syscall"

	"github.com/d5/tengo/objects"
)

func makeOSProcessState(state *os.ProcessState) *objects.ImmutableMap {
	return &objects.ImmutableMap{
		Value: map[string]objects.Object{
			"exited":  &objects.UserFunction{Name: "exited", Value: FuncARB(state.Exited)},   //
			"pid":     &objects.UserFunction{Name: "pid", Value: FuncARI(state.Pid)},         //
			"string":  &objects.UserFunction{Name: "string", Value: FuncARS(state.String)},   //
			"success": &objects.UserFunction{Name: "success", Value: FuncARB(state.Success)}, //
		},
	}
}

func makeOSProcess(proc *os.Process) *objects.ImmutableMap {
	return &objects.ImmutableMap{
		Value: map[string]objects.Object{
			"kill":    &objects.UserFunction{Name: "kill", Value: FuncARE(proc.Kill)},       //
			"release": &objects.UserFunction{Name: "release", Value: FuncARE(proc.Release)}, //
			"signal": &objects.UserFunction{
				Name: "signal",
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

					return wrapError(proc.Signal(syscall.Signal(i1))), nil
				},
			},
			"wait": &objects.UserFunction{
				Name: "wait",
				Value: func(args ...objects.Object) (ret objects.Object, err error) {
					if len(args) != 0 {
						return nil, objects.ErrWrongNumArguments
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
