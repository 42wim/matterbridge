package stdlib

import (
	"fmt"

	"github.com/d5/tengo"
	"github.com/d5/tengo/objects"
)

// FuncAR transform a function of 'func()' signature
// into CallableFunc type.
func FuncAR(fn func()) objects.CallableFunc {
	return func(args ...objects.Object) (ret objects.Object, err error) {
		if len(args) != 0 {
			return nil, objects.ErrWrongNumArguments
		}

		fn()

		return objects.UndefinedValue, nil
	}
}

// FuncARI transform a function of 'func() int' signature
// into CallableFunc type.
func FuncARI(fn func() int) objects.CallableFunc {
	return func(args ...objects.Object) (ret objects.Object, err error) {
		if len(args) != 0 {
			return nil, objects.ErrWrongNumArguments
		}

		return &objects.Int{Value: int64(fn())}, nil
	}
}

// FuncARI64 transform a function of 'func() int64' signature
// into CallableFunc type.
func FuncARI64(fn func() int64) objects.CallableFunc {
	return func(args ...objects.Object) (ret objects.Object, err error) {
		if len(args) != 0 {
			return nil, objects.ErrWrongNumArguments
		}

		return &objects.Int{Value: fn()}, nil
	}
}

// FuncAI64RI64 transform a function of 'func(int64) int64' signature
// into CallableFunc type.
func FuncAI64RI64(fn func(int64) int64) objects.CallableFunc {
	return func(args ...objects.Object) (ret objects.Object, err error) {
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

		return &objects.Int{Value: fn(i1)}, nil
	}
}

// FuncAI64R transform a function of 'func(int64)' signature
// into CallableFunc type.
func FuncAI64R(fn func(int64)) objects.CallableFunc {
	return func(args ...objects.Object) (ret objects.Object, err error) {
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

		fn(i1)

		return objects.UndefinedValue, nil
	}
}

// FuncARB transform a function of 'func() bool' signature
// into CallableFunc type.
func FuncARB(fn func() bool) objects.CallableFunc {
	return func(args ...objects.Object) (ret objects.Object, err error) {
		if len(args) != 0 {
			return nil, objects.ErrWrongNumArguments
		}

		if fn() {
			return objects.TrueValue, nil
		}

		return objects.FalseValue, nil
	}
}

// FuncARE transform a function of 'func() error' signature
// into CallableFunc type.
func FuncARE(fn func() error) objects.CallableFunc {
	return func(args ...objects.Object) (ret objects.Object, err error) {
		if len(args) != 0 {
			return nil, objects.ErrWrongNumArguments
		}

		return wrapError(fn()), nil
	}
}

// FuncARS transform a function of 'func() string' signature
// into CallableFunc type.
func FuncARS(fn func() string) objects.CallableFunc {
	return func(args ...objects.Object) (ret objects.Object, err error) {
		if len(args) != 0 {
			return nil, objects.ErrWrongNumArguments
		}

		s := fn()

		if len(s) > tengo.MaxStringLen {
			return nil, objects.ErrStringLimit
		}

		return &objects.String{Value: s}, nil
	}
}

// FuncARSE transform a function of 'func() (string, error)' signature
// into CallableFunc type.
func FuncARSE(fn func() (string, error)) objects.CallableFunc {
	return func(args ...objects.Object) (ret objects.Object, err error) {
		if len(args) != 0 {
			return nil, objects.ErrWrongNumArguments
		}

		res, err := fn()
		if err != nil {
			return wrapError(err), nil
		}

		if len(res) > tengo.MaxStringLen {
			return nil, objects.ErrStringLimit
		}

		return &objects.String{Value: res}, nil
	}
}

// FuncARYE transform a function of 'func() ([]byte, error)' signature
// into CallableFunc type.
func FuncARYE(fn func() ([]byte, error)) objects.CallableFunc {
	return func(args ...objects.Object) (ret objects.Object, err error) {
		if len(args) != 0 {
			return nil, objects.ErrWrongNumArguments
		}

		res, err := fn()
		if err != nil {
			return wrapError(err), nil
		}

		if len(res) > tengo.MaxBytesLen {
			return nil, objects.ErrBytesLimit
		}

		return &objects.Bytes{Value: res}, nil
	}
}

// FuncARF transform a function of 'func() float64' signature
// into CallableFunc type.
func FuncARF(fn func() float64) objects.CallableFunc {
	return func(args ...objects.Object) (ret objects.Object, err error) {
		if len(args) != 0 {
			return nil, objects.ErrWrongNumArguments
		}

		return &objects.Float{Value: fn()}, nil
	}
}

// FuncARSs transform a function of 'func() []string' signature
// into CallableFunc type.
func FuncARSs(fn func() []string) objects.CallableFunc {
	return func(args ...objects.Object) (ret objects.Object, err error) {
		if len(args) != 0 {
			return nil, objects.ErrWrongNumArguments
		}

		arr := &objects.Array{}
		for _, elem := range fn() {
			if len(elem) > tengo.MaxStringLen {
				return nil, objects.ErrStringLimit
			}

			arr.Value = append(arr.Value, &objects.String{Value: elem})
		}

		return arr, nil
	}
}

// FuncARIsE transform a function of 'func() ([]int, error)' signature
// into CallableFunc type.
func FuncARIsE(fn func() ([]int, error)) objects.CallableFunc {
	return func(args ...objects.Object) (ret objects.Object, err error) {
		if len(args) != 0 {
			return nil, objects.ErrWrongNumArguments
		}

		res, err := fn()
		if err != nil {
			return wrapError(err), nil
		}

		arr := &objects.Array{}
		for _, v := range res {
			arr.Value = append(arr.Value, &objects.Int{Value: int64(v)})
		}

		return arr, nil
	}
}

// FuncAIRIs transform a function of 'func(int) []int' signature
// into CallableFunc type.
func FuncAIRIs(fn func(int) []int) objects.CallableFunc {
	return func(args ...objects.Object) (ret objects.Object, err error) {
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

		res := fn(i1)

		arr := &objects.Array{}
		for _, v := range res {
			arr.Value = append(arr.Value, &objects.Int{Value: int64(v)})
		}

		return arr, nil
	}
}

// FuncAFRF transform a function of 'func(float64) float64' signature
// into CallableFunc type.
func FuncAFRF(fn func(float64) float64) objects.CallableFunc {
	return func(args ...objects.Object) (ret objects.Object, err error) {
		if len(args) != 1 {
			return nil, objects.ErrWrongNumArguments
		}

		f1, ok := objects.ToFloat64(args[0])
		if !ok {
			return nil, objects.ErrInvalidArgumentType{
				Name:     "first",
				Expected: "float(compatible)",
				Found:    args[0].TypeName(),
			}
		}

		return &objects.Float{Value: fn(f1)}, nil
	}
}

// FuncAIR transform a function of 'func(int)' signature
// into CallableFunc type.
func FuncAIR(fn func(int)) objects.CallableFunc {
	return func(args ...objects.Object) (ret objects.Object, err error) {
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

		fn(i1)

		return objects.UndefinedValue, nil
	}
}

// FuncAIRF transform a function of 'func(int) float64' signature
// into CallableFunc type.
func FuncAIRF(fn func(int) float64) objects.CallableFunc {
	return func(args ...objects.Object) (ret objects.Object, err error) {
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

		return &objects.Float{Value: fn(i1)}, nil
	}
}

// FuncAFRI transform a function of 'func(float64) int' signature
// into CallableFunc type.
func FuncAFRI(fn func(float64) int) objects.CallableFunc {
	return func(args ...objects.Object) (ret objects.Object, err error) {
		if len(args) != 1 {
			return nil, objects.ErrWrongNumArguments
		}

		f1, ok := objects.ToFloat64(args[0])
		if !ok {
			return nil, objects.ErrInvalidArgumentType{
				Name:     "first",
				Expected: "float(compatible)",
				Found:    args[0].TypeName(),
			}
		}

		return &objects.Int{Value: int64(fn(f1))}, nil
	}
}

// FuncAFFRF transform a function of 'func(float64, float64) float64' signature
// into CallableFunc type.
func FuncAFFRF(fn func(float64, float64) float64) objects.CallableFunc {
	return func(args ...objects.Object) (ret objects.Object, err error) {
		if len(args) != 2 {
			return nil, objects.ErrWrongNumArguments
		}

		f1, ok := objects.ToFloat64(args[0])
		if !ok {
			return nil, objects.ErrInvalidArgumentType{
				Name:     "first",
				Expected: "float(compatible)",
				Found:    args[0].TypeName(),
			}
		}

		f2, ok := objects.ToFloat64(args[1])
		if !ok {
			return nil, objects.ErrInvalidArgumentType{
				Name:     "second",
				Expected: "float(compatible)",
				Found:    args[1].TypeName(),
			}
		}

		return &objects.Float{Value: fn(f1, f2)}, nil
	}
}

// FuncAIFRF transform a function of 'func(int, float64) float64' signature
// into CallableFunc type.
func FuncAIFRF(fn func(int, float64) float64) objects.CallableFunc {
	return func(args ...objects.Object) (ret objects.Object, err error) {
		if len(args) != 2 {
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

		f2, ok := objects.ToFloat64(args[1])
		if !ok {
			return nil, objects.ErrInvalidArgumentType{
				Name:     "second",
				Expected: "float(compatible)",
				Found:    args[1].TypeName(),
			}
		}

		return &objects.Float{Value: fn(i1, f2)}, nil
	}
}

// FuncAFIRF transform a function of 'func(float64, int) float64' signature
// into CallableFunc type.
func FuncAFIRF(fn func(float64, int) float64) objects.CallableFunc {
	return func(args ...objects.Object) (ret objects.Object, err error) {
		if len(args) != 2 {
			return nil, objects.ErrWrongNumArguments
		}

		f1, ok := objects.ToFloat64(args[0])
		if !ok {
			return nil, objects.ErrInvalidArgumentType{
				Name:     "first",
				Expected: "float(compatible)",
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

		return &objects.Float{Value: fn(f1, i2)}, nil
	}
}

// FuncAFIRB transform a function of 'func(float64, int) bool' signature
// into CallableFunc type.
func FuncAFIRB(fn func(float64, int) bool) objects.CallableFunc {
	return func(args ...objects.Object) (ret objects.Object, err error) {
		if len(args) != 2 {
			return nil, objects.ErrWrongNumArguments
		}

		f1, ok := objects.ToFloat64(args[0])
		if !ok {
			return nil, objects.ErrInvalidArgumentType{
				Name:     "first",
				Expected: "float(compatible)",
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

		if fn(f1, i2) {
			return objects.TrueValue, nil
		}

		return objects.FalseValue, nil
	}
}

// FuncAFRB transform a function of 'func(float64) bool' signature
// into CallableFunc type.
func FuncAFRB(fn func(float64) bool) objects.CallableFunc {
	return func(args ...objects.Object) (ret objects.Object, err error) {
		if len(args) != 1 {
			return nil, objects.ErrWrongNumArguments
		}

		f1, ok := objects.ToFloat64(args[0])
		if !ok {
			return nil, objects.ErrInvalidArgumentType{
				Name:     "first",
				Expected: "float(compatible)",
				Found:    args[0].TypeName(),
			}
		}

		if fn(f1) {
			return objects.TrueValue, nil
		}

		return objects.FalseValue, nil
	}
}

// FuncASRS transform a function of 'func(string) string' signature into CallableFunc type.
// User function will return 'true' if underlying native function returns nil.
func FuncASRS(fn func(string) string) objects.CallableFunc {
	return func(args ...objects.Object) (objects.Object, error) {
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

		s := fn(s1)

		if len(s) > tengo.MaxStringLen {
			return nil, objects.ErrStringLimit
		}

		return &objects.String{Value: s}, nil
	}
}

// FuncASRSs transform a function of 'func(string) []string' signature into CallableFunc type.
func FuncASRSs(fn func(string) []string) objects.CallableFunc {
	return func(args ...objects.Object) (objects.Object, error) {
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

		res := fn(s1)

		arr := &objects.Array{}
		for _, elem := range res {
			if len(elem) > tengo.MaxStringLen {
				return nil, objects.ErrStringLimit
			}

			arr.Value = append(arr.Value, &objects.String{Value: elem})
		}

		return arr, nil
	}
}

// FuncASRSE transform a function of 'func(string) (string, error)' signature into CallableFunc type.
// User function will return 'true' if underlying native function returns nil.
func FuncASRSE(fn func(string) (string, error)) objects.CallableFunc {
	return func(args ...objects.Object) (objects.Object, error) {
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

		res, err := fn(s1)
		if err != nil {
			return wrapError(err), nil
		}

		if len(res) > tengo.MaxStringLen {
			return nil, objects.ErrStringLimit
		}

		return &objects.String{Value: res}, nil
	}
}

// FuncASRE transform a function of 'func(string) error' signature into CallableFunc type.
// User function will return 'true' if underlying native function returns nil.
func FuncASRE(fn func(string) error) objects.CallableFunc {
	return func(args ...objects.Object) (objects.Object, error) {
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

		return wrapError(fn(s1)), nil
	}
}

// FuncASSRE transform a function of 'func(string, string) error' signature into CallableFunc type.
// User function will return 'true' if underlying native function returns nil.
func FuncASSRE(fn func(string, string) error) objects.CallableFunc {
	return func(args ...objects.Object) (objects.Object, error) {
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

		s2, ok := objects.ToString(args[1])
		if !ok {
			return nil, objects.ErrInvalidArgumentType{
				Name:     "second",
				Expected: "string(compatible)",
				Found:    args[1].TypeName(),
			}
		}

		return wrapError(fn(s1, s2)), nil
	}
}

// FuncASSRSs transform a function of 'func(string, string) []string' signature into CallableFunc type.
func FuncASSRSs(fn func(string, string) []string) objects.CallableFunc {
	return func(args ...objects.Object) (objects.Object, error) {
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

		s2, ok := objects.ToString(args[1])
		if !ok {
			return nil, objects.ErrInvalidArgumentType{
				Name:     "first",
				Expected: "string(compatible)",
				Found:    args[1].TypeName(),
			}
		}

		arr := &objects.Array{}
		for _, res := range fn(s1, s2) {
			if len(res) > tengo.MaxStringLen {
				return nil, objects.ErrStringLimit
			}

			arr.Value = append(arr.Value, &objects.String{Value: res})
		}

		return arr, nil
	}
}

// FuncASSIRSs transform a function of 'func(string, string, int) []string' signature into CallableFunc type.
func FuncASSIRSs(fn func(string, string, int) []string) objects.CallableFunc {
	return func(args ...objects.Object) (objects.Object, error) {
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

		s2, ok := objects.ToString(args[1])
		if !ok {
			return nil, objects.ErrInvalidArgumentType{
				Name:     "second",
				Expected: "string(compatible)",
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

		arr := &objects.Array{}
		for _, res := range fn(s1, s2, i3) {
			if len(res) > tengo.MaxStringLen {
				return nil, objects.ErrStringLimit
			}

			arr.Value = append(arr.Value, &objects.String{Value: res})
		}

		return arr, nil
	}
}

// FuncASSRI transform a function of 'func(string, string) int' signature into CallableFunc type.
func FuncASSRI(fn func(string, string) int) objects.CallableFunc {
	return func(args ...objects.Object) (objects.Object, error) {
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

		s2, ok := objects.ToString(args[1])
		if !ok {
			return nil, objects.ErrInvalidArgumentType{
				Name:     "second",
				Expected: "string(compatible)",
				Found:    args[0].TypeName(),
			}
		}

		return &objects.Int{Value: int64(fn(s1, s2))}, nil
	}
}

// FuncASSRS transform a function of 'func(string, string) string' signature into CallableFunc type.
func FuncASSRS(fn func(string, string) string) objects.CallableFunc {
	return func(args ...objects.Object) (objects.Object, error) {
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

		s2, ok := objects.ToString(args[1])
		if !ok {
			return nil, objects.ErrInvalidArgumentType{
				Name:     "second",
				Expected: "string(compatible)",
				Found:    args[1].TypeName(),
			}
		}

		s := fn(s1, s2)

		if len(s) > tengo.MaxStringLen {
			return nil, objects.ErrStringLimit
		}

		return &objects.String{Value: s}, nil
	}
}

// FuncASSRB transform a function of 'func(string, string) bool' signature into CallableFunc type.
func FuncASSRB(fn func(string, string) bool) objects.CallableFunc {
	return func(args ...objects.Object) (objects.Object, error) {
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

		s2, ok := objects.ToString(args[1])
		if !ok {
			return nil, objects.ErrInvalidArgumentType{
				Name:     "second",
				Expected: "string(compatible)",
				Found:    args[1].TypeName(),
			}
		}

		if fn(s1, s2) {
			return objects.TrueValue, nil
		}

		return objects.FalseValue, nil
	}
}

// FuncASsSRS transform a function of 'func([]string, string) string' signature into CallableFunc type.
func FuncASsSRS(fn func([]string, string) string) objects.CallableFunc {
	return func(args ...objects.Object) (objects.Object, error) {
		if len(args) != 2 {
			return nil, objects.ErrWrongNumArguments
		}

		var ss1 []string
		switch arg0 := args[0].(type) {
		case *objects.Array:
			for idx, a := range arg0.Value {
				as, ok := objects.ToString(a)
				if !ok {
					return nil, objects.ErrInvalidArgumentType{
						Name:     fmt.Sprintf("first[%d]", idx),
						Expected: "string(compatible)",
						Found:    a.TypeName(),
					}
				}
				ss1 = append(ss1, as)
			}
		case *objects.ImmutableArray:
			for idx, a := range arg0.Value {
				as, ok := objects.ToString(a)
				if !ok {
					return nil, objects.ErrInvalidArgumentType{
						Name:     fmt.Sprintf("first[%d]", idx),
						Expected: "string(compatible)",
						Found:    a.TypeName(),
					}
				}
				ss1 = append(ss1, as)
			}
		default:
			return nil, objects.ErrInvalidArgumentType{
				Name:     "first",
				Expected: "array",
				Found:    args[0].TypeName(),
			}
		}

		s2, ok := objects.ToString(args[1])
		if !ok {
			return nil, objects.ErrInvalidArgumentType{
				Name:     "second",
				Expected: "string(compatible)",
				Found:    args[1].TypeName(),
			}
		}

		s := fn(ss1, s2)
		if len(s) > tengo.MaxStringLen {
			return nil, objects.ErrStringLimit
		}

		return &objects.String{Value: s}, nil
	}
}

// FuncASI64RE transform a function of 'func(string, int64) error' signature
// into CallableFunc type.
func FuncASI64RE(fn func(string, int64) error) objects.CallableFunc {
	return func(args ...objects.Object) (ret objects.Object, err error) {
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

		return wrapError(fn(s1, i2)), nil
	}
}

// FuncAIIRE transform a function of 'func(int, int) error' signature
// into CallableFunc type.
func FuncAIIRE(fn func(int, int) error) objects.CallableFunc {
	return func(args ...objects.Object) (ret objects.Object, err error) {
		if len(args) != 2 {
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

		i2, ok := objects.ToInt(args[1])
		if !ok {
			return nil, objects.ErrInvalidArgumentType{
				Name:     "second",
				Expected: "int(compatible)",
				Found:    args[1].TypeName(),
			}
		}

		return wrapError(fn(i1, i2)), nil
	}
}

// FuncASIRS transform a function of 'func(string, int) string' signature
// into CallableFunc type.
func FuncASIRS(fn func(string, int) string) objects.CallableFunc {
	return func(args ...objects.Object) (ret objects.Object, err error) {
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

		i2, ok := objects.ToInt(args[1])
		if !ok {
			return nil, objects.ErrInvalidArgumentType{
				Name:     "second",
				Expected: "int(compatible)",
				Found:    args[1].TypeName(),
			}
		}

		s := fn(s1, i2)

		if len(s) > tengo.MaxStringLen {
			return nil, objects.ErrStringLimit
		}

		return &objects.String{Value: s}, nil
	}
}

// FuncASIIRE transform a function of 'func(string, int, int) error' signature
// into CallableFunc type.
func FuncASIIRE(fn func(string, int, int) error) objects.CallableFunc {
	return func(args ...objects.Object) (ret objects.Object, err error) {
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

		return wrapError(fn(s1, i2, i3)), nil
	}
}

// FuncAYRIE transform a function of 'func([]byte) (int, error)' signature
// into CallableFunc type.
func FuncAYRIE(fn func([]byte) (int, error)) objects.CallableFunc {
	return func(args ...objects.Object) (ret objects.Object, err error) {
		if len(args) != 1 {
			return nil, objects.ErrWrongNumArguments
		}

		y1, ok := objects.ToByteSlice(args[0])
		if !ok {
			return nil, objects.ErrInvalidArgumentType{
				Name:     "first",
				Expected: "bytes(compatible)",
				Found:    args[0].TypeName(),
			}
		}

		res, err := fn(y1)
		if err != nil {
			return wrapError(err), nil
		}

		return &objects.Int{Value: int64(res)}, nil
	}
}

// FuncAYRS transform a function of 'func([]byte) string' signature
// into CallableFunc type.
func FuncAYRS(fn func([]byte) string) objects.CallableFunc {
	return func(args ...objects.Object) (ret objects.Object, err error) {
		if len(args) != 1 {
			return nil, objects.ErrWrongNumArguments
		}

		y1, ok := objects.ToByteSlice(args[0])
		if !ok {
			return nil, objects.ErrInvalidArgumentType{
				Name:     "first",
				Expected: "bytes(compatible)",
				Found:    args[0].TypeName(),
			}
		}

		res := fn(y1)

		return &objects.String{Value: res}, nil
	}
}

// FuncASRIE transform a function of 'func(string) (int, error)' signature
// into CallableFunc type.
func FuncASRIE(fn func(string) (int, error)) objects.CallableFunc {
	return func(args ...objects.Object) (ret objects.Object, err error) {
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

		res, err := fn(s1)
		if err != nil {
			return wrapError(err), nil
		}

		return &objects.Int{Value: int64(res)}, nil
	}
}

// FuncASRYE transform a function of 'func(string) ([]byte, error)' signature
// into CallableFunc type.
func FuncASRYE(fn func(string) ([]byte, error)) objects.CallableFunc {
	return func(args ...objects.Object) (ret objects.Object, err error) {
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

		res, err := fn(s1)
		if err != nil {
			return wrapError(err), nil
		}

		if len(res) > tengo.MaxBytesLen {
			return nil, objects.ErrBytesLimit
		}

		return &objects.Bytes{Value: res}, nil
	}
}

// FuncAIRSsE transform a function of 'func(int) ([]string, error)' signature
// into CallableFunc type.
func FuncAIRSsE(fn func(int) ([]string, error)) objects.CallableFunc {
	return func(args ...objects.Object) (ret objects.Object, err error) {
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

		res, err := fn(i1)
		if err != nil {
			return wrapError(err), nil
		}

		arr := &objects.Array{}
		for _, r := range res {
			if len(r) > tengo.MaxStringLen {
				return nil, objects.ErrStringLimit
			}

			arr.Value = append(arr.Value, &objects.String{Value: r})
		}

		return arr, nil
	}
}

// FuncAIRS transform a function of 'func(int) string' signature
// into CallableFunc type.
func FuncAIRS(fn func(int) string) objects.CallableFunc {
	return func(args ...objects.Object) (ret objects.Object, err error) {
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

		s := fn(i1)

		if len(s) > tengo.MaxStringLen {
			return nil, objects.ErrStringLimit
		}

		return &objects.String{Value: s}, nil
	}
}
