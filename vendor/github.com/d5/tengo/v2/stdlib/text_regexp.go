package stdlib

import (
	"regexp"

	"github.com/d5/tengo/v2"
)

func makeTextRegexp(re *regexp.Regexp) *tengo.ImmutableMap {
	return &tengo.ImmutableMap{
		Value: map[string]tengo.Object{
			// match(text) => bool
			"match": &tengo.UserFunction{
				Value: func(args ...tengo.Object) (
					ret tengo.Object,
					err error,
				) {
					if len(args) != 1 {
						err = tengo.ErrWrongNumArguments
						return
					}

					s1, ok := tengo.ToString(args[0])
					if !ok {
						err = tengo.ErrInvalidArgumentType{
							Name:     "first",
							Expected: "string(compatible)",
							Found:    args[0].TypeName(),
						}
						return
					}

					if re.MatchString(s1) {
						ret = tengo.TrueValue
					} else {
						ret = tengo.FalseValue
					}

					return
				},
			},

			// find(text) 			=> array(array({text:,begin:,end:}))/undefined
			// find(text, maxCount) => array(array({text:,begin:,end:}))/undefined
			"find": &tengo.UserFunction{
				Value: func(args ...tengo.Object) (
					ret tengo.Object,
					err error,
				) {
					numArgs := len(args)
					if numArgs != 1 && numArgs != 2 {
						err = tengo.ErrWrongNumArguments
						return
					}

					s1, ok := tengo.ToString(args[0])
					if !ok {
						err = tengo.ErrInvalidArgumentType{
							Name:     "first",
							Expected: "string(compatible)",
							Found:    args[0].TypeName(),
						}
						return
					}

					if numArgs == 1 {
						m := re.FindStringSubmatchIndex(s1)
						if m == nil {
							ret = tengo.UndefinedValue
							return
						}

						arr := &tengo.Array{}
						for i := 0; i < len(m); i += 2 {
							arr.Value = append(arr.Value,
								&tengo.ImmutableMap{
									Value: map[string]tengo.Object{
										"text": &tengo.String{
											Value: s1[m[i]:m[i+1]],
										},
										"begin": &tengo.Int{
											Value: int64(m[i]),
										},
										"end": &tengo.Int{
											Value: int64(m[i+1]),
										},
									}})
						}

						ret = &tengo.Array{Value: []tengo.Object{arr}}

						return
					}

					i2, ok := tengo.ToInt(args[1])
					if !ok {
						err = tengo.ErrInvalidArgumentType{
							Name:     "second",
							Expected: "int(compatible)",
							Found:    args[1].TypeName(),
						}
						return
					}
					m := re.FindAllStringSubmatchIndex(s1, i2)
					if m == nil {
						ret = tengo.UndefinedValue
						return
					}

					arr := &tengo.Array{}
					for _, m := range m {
						subMatch := &tengo.Array{}
						for i := 0; i < len(m); i += 2 {
							subMatch.Value = append(subMatch.Value,
								&tengo.ImmutableMap{
									Value: map[string]tengo.Object{
										"text": &tengo.String{
											Value: s1[m[i]:m[i+1]],
										},
										"begin": &tengo.Int{
											Value: int64(m[i]),
										},
										"end": &tengo.Int{
											Value: int64(m[i+1]),
										},
									}})
						}

						arr.Value = append(arr.Value, subMatch)
					}

					ret = arr

					return
				},
			},

			// replace(src, repl) => string
			"replace": &tengo.UserFunction{
				Value: func(args ...tengo.Object) (
					ret tengo.Object,
					err error,
				) {
					if len(args) != 2 {
						err = tengo.ErrWrongNumArguments
						return
					}

					s1, ok := tengo.ToString(args[0])
					if !ok {
						err = tengo.ErrInvalidArgumentType{
							Name:     "first",
							Expected: "string(compatible)",
							Found:    args[0].TypeName(),
						}
						return
					}

					s2, ok := tengo.ToString(args[1])
					if !ok {
						err = tengo.ErrInvalidArgumentType{
							Name:     "second",
							Expected: "string(compatible)",
							Found:    args[1].TypeName(),
						}
						return
					}

					s, ok := doTextRegexpReplace(re, s1, s2)
					if !ok {
						return nil, tengo.ErrStringLimit
					}

					ret = &tengo.String{Value: s}

					return
				},
			},

			// split(text) 			 => array(string)
			// split(text, maxCount) => array(string)
			"split": &tengo.UserFunction{
				Value: func(args ...tengo.Object) (
					ret tengo.Object,
					err error,
				) {
					numArgs := len(args)
					if numArgs != 1 && numArgs != 2 {
						err = tengo.ErrWrongNumArguments
						return
					}

					s1, ok := tengo.ToString(args[0])
					if !ok {
						err = tengo.ErrInvalidArgumentType{
							Name:     "first",
							Expected: "string(compatible)",
							Found:    args[0].TypeName(),
						}
						return
					}

					var i2 = -1
					if numArgs > 1 {
						i2, ok = tengo.ToInt(args[1])
						if !ok {
							err = tengo.ErrInvalidArgumentType{
								Name:     "second",
								Expected: "int(compatible)",
								Found:    args[1].TypeName(),
							}
							return
						}
					}

					arr := &tengo.Array{}
					for _, s := range re.Split(s1, i2) {
						arr.Value = append(arr.Value,
							&tengo.String{Value: s})
					}

					ret = arr

					return
				},
			},
		},
	}
}

// Size-limit checking implementation of regexp.ReplaceAllString.
func doTextRegexpReplace(re *regexp.Regexp, src, repl string) (string, bool) {
	idx := 0
	out := ""
	for _, m := range re.FindAllStringSubmatchIndex(src, -1) {
		var exp []byte
		exp = re.ExpandString(exp, repl, src, m)
		if len(out)+m[0]-idx+len(exp) > tengo.MaxStringLen {
			return "", false
		}
		out += src[idx:m[0]] + string(exp)
		idx = m[1]
	}
	if idx < len(src) {
		if len(out)+len(src)-idx > tengo.MaxStringLen {
			return "", false
		}
		out += src[idx:]
	}
	return out, true
}
