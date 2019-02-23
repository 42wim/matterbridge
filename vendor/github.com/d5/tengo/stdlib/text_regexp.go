package stdlib

import (
	"regexp"

	"github.com/d5/tengo/objects"
)

func makeTextRegexp(re *regexp.Regexp) *objects.ImmutableMap {
	return &objects.ImmutableMap{
		Value: map[string]objects.Object{
			// match(text) => bool
			"match": &objects.UserFunction{
				Value: func(args ...objects.Object) (ret objects.Object, err error) {
					if len(args) != 1 {
						err = objects.ErrWrongNumArguments
						return
					}

					s1, ok := objects.ToString(args[0])
					if !ok {
						err = objects.ErrInvalidArgumentType{
							Name:     "first",
							Expected: "string(compatible)",
							Found:    args[0].TypeName(),
						}
						return
					}

					if re.MatchString(s1) {
						ret = objects.TrueValue
					} else {
						ret = objects.FalseValue
					}

					return
				},
			},

			// find(text) 			=> array(array({text:,begin:,end:}))/undefined
			// find(text, maxCount) => array(array({text:,begin:,end:}))/undefined
			"find": &objects.UserFunction{
				Value: func(args ...objects.Object) (ret objects.Object, err error) {
					numArgs := len(args)
					if numArgs != 1 && numArgs != 2 {
						err = objects.ErrWrongNumArguments
						return
					}

					s1, ok := objects.ToString(args[0])
					if !ok {
						err = objects.ErrInvalidArgumentType{
							Name:     "first",
							Expected: "string(compatible)",
							Found:    args[0].TypeName(),
						}
						return
					}

					if numArgs == 1 {
						m := re.FindStringSubmatchIndex(s1)
						if m == nil {
							ret = objects.UndefinedValue
							return
						}

						arr := &objects.Array{}
						for i := 0; i < len(m); i += 2 {
							arr.Value = append(arr.Value, &objects.ImmutableMap{Value: map[string]objects.Object{
								"text":  &objects.String{Value: s1[m[i]:m[i+1]]},
								"begin": &objects.Int{Value: int64(m[i])},
								"end":   &objects.Int{Value: int64(m[i+1])},
							}})
						}

						ret = &objects.Array{Value: []objects.Object{arr}}

						return
					}

					i2, ok := objects.ToInt(args[1])
					if !ok {
						err = objects.ErrInvalidArgumentType{
							Name:     "second",
							Expected: "int(compatible)",
							Found:    args[1].TypeName(),
						}
						return
					}
					m := re.FindAllStringSubmatchIndex(s1, i2)
					if m == nil {
						ret = objects.UndefinedValue
						return
					}

					arr := &objects.Array{}
					for _, m := range m {
						subMatch := &objects.Array{}
						for i := 0; i < len(m); i += 2 {
							subMatch.Value = append(subMatch.Value, &objects.ImmutableMap{Value: map[string]objects.Object{
								"text":  &objects.String{Value: s1[m[i]:m[i+1]]},
								"begin": &objects.Int{Value: int64(m[i])},
								"end":   &objects.Int{Value: int64(m[i+1])},
							}})
						}

						arr.Value = append(arr.Value, subMatch)
					}

					ret = arr

					return
				},
			},

			// replace(src, repl) => string
			"replace": &objects.UserFunction{
				Value: func(args ...objects.Object) (ret objects.Object, err error) {
					if len(args) != 2 {
						err = objects.ErrWrongNumArguments
						return
					}

					s1, ok := objects.ToString(args[0])
					if !ok {
						err = objects.ErrInvalidArgumentType{
							Name:     "first",
							Expected: "string(compatible)",
							Found:    args[0].TypeName(),
						}
						return
					}

					s2, ok := objects.ToString(args[1])
					if !ok {
						err = objects.ErrInvalidArgumentType{
							Name:     "second",
							Expected: "string(compatible)",
							Found:    args[1].TypeName(),
						}
						return
					}

					ret = &objects.String{Value: re.ReplaceAllString(s1, s2)}

					return
				},
			},

			// split(text) 			 => array(string)
			// split(text, maxCount) => array(string)
			"split": &objects.UserFunction{
				Value: func(args ...objects.Object) (ret objects.Object, err error) {
					numArgs := len(args)
					if numArgs != 1 && numArgs != 2 {
						err = objects.ErrWrongNumArguments
						return
					}

					s1, ok := objects.ToString(args[0])
					if !ok {
						err = objects.ErrInvalidArgumentType{
							Name:     "first",
							Expected: "string(compatible)",
							Found:    args[0].TypeName(),
						}
						return
					}

					var i2 = -1
					if numArgs > 1 {
						i2, ok = objects.ToInt(args[1])
						if !ok {
							err = objects.ErrInvalidArgumentType{
								Name:     "second",
								Expected: "int(compatible)",
								Found:    args[1].TypeName(),
							}
							return
						}
					}

					arr := &objects.Array{}
					for _, s := range re.Split(s1, i2) {
						arr.Value = append(arr.Value, &objects.String{Value: s})
					}

					ret = arr

					return
				},
			},
		},
	}
}
