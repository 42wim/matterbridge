package stdlib

import (
	"math/rand"

	"github.com/d5/tengo/objects"
)

var randModule = map[string]objects.Object{
	"int":        &objects.UserFunction{Name: "int", Value: FuncARI64(rand.Int63)},
	"float":      &objects.UserFunction{Name: "float", Value: FuncARF(rand.Float64)},
	"intn":       &objects.UserFunction{Name: "intn", Value: FuncAI64RI64(rand.Int63n)},
	"exp_float":  &objects.UserFunction{Name: "exp_float", Value: FuncARF(rand.ExpFloat64)},
	"norm_float": &objects.UserFunction{Name: "norm_float", Value: FuncARF(rand.NormFloat64)},
	"perm":       &objects.UserFunction{Name: "perm", Value: FuncAIRIs(rand.Perm)},
	"seed":       &objects.UserFunction{Name: "seed", Value: FuncAI64R(rand.Seed)},
	"read": &objects.UserFunction{
		Name: "read",
		Value: func(args ...objects.Object) (ret objects.Object, err error) {
			if len(args) != 1 {
				return nil, objects.ErrWrongNumArguments
			}

			y1, ok := args[0].(*objects.Bytes)
			if !ok {
				return nil, objects.ErrInvalidArgumentType{
					Name:     "first",
					Expected: "bytes",
					Found:    args[0].TypeName(),
				}
			}

			res, err := rand.Read(y1.Value)
			if err != nil {
				ret = wrapError(err)
				return
			}

			return &objects.Int{Value: int64(res)}, nil
		},
	},
	"rand": &objects.UserFunction{
		Name: "rand",
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

			src := rand.NewSource(i1)

			return randRand(rand.New(src)), nil
		},
	},
}

func randRand(r *rand.Rand) *objects.ImmutableMap {
	return &objects.ImmutableMap{
		Value: map[string]objects.Object{
			"int":        &objects.UserFunction{Name: "int", Value: FuncARI64(r.Int63)},
			"float":      &objects.UserFunction{Name: "float", Value: FuncARF(r.Float64)},
			"intn":       &objects.UserFunction{Name: "intn", Value: FuncAI64RI64(r.Int63n)},
			"exp_float":  &objects.UserFunction{Name: "exp_float", Value: FuncARF(r.ExpFloat64)},
			"norm_float": &objects.UserFunction{Name: "norm_float", Value: FuncARF(r.NormFloat64)},
			"perm":       &objects.UserFunction{Name: "perm", Value: FuncAIRIs(r.Perm)},
			"seed":       &objects.UserFunction{Name: "seed", Value: FuncAI64R(r.Seed)},
			"read": &objects.UserFunction{
				Name: "read",
				Value: func(args ...objects.Object) (ret objects.Object, err error) {
					if len(args) != 1 {
						return nil, objects.ErrWrongNumArguments
					}

					y1, ok := args[0].(*objects.Bytes)
					if !ok {
						return nil, objects.ErrInvalidArgumentType{
							Name:     "first",
							Expected: "bytes",
							Found:    args[0].TypeName(),
						}
					}

					res, err := r.Read(y1.Value)
					if err != nil {
						ret = wrapError(err)
						return
					}

					return &objects.Int{Value: int64(res)}, nil
				},
			},
		},
	}
}
