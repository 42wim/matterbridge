package objects

// Builtins contains all default builtin functions.
// Use GetBuiltinFunctions instead of accessing Builtins directly.
var Builtins = []BuiltinFunction{
	{
		Name:  "print",
		Value: builtinPrint,
	},
	{
		Name:  "printf",
		Value: builtinPrintf,
	},
	{
		Name:  "sprintf",
		Value: builtinSprintf,
	},
	{
		Name:  "len",
		Value: builtinLen,
	},
	{
		Name:  "copy",
		Value: builtinCopy,
	},
	{
		Name:  "append",
		Value: builtinAppend,
	},
	{
		Name:  "string",
		Value: builtinString,
	},
	{
		Name:  "int",
		Value: builtinInt,
	},
	{
		Name:  "bool",
		Value: builtinBool,
	},
	{
		Name:  "float",
		Value: builtinFloat,
	},
	{
		Name:  "char",
		Value: builtinChar,
	},
	{
		Name:  "bytes",
		Value: builtinBytes,
	},
	{
		Name:  "time",
		Value: builtinTime,
	},
	{
		Name:  "is_int",
		Value: builtinIsInt,
	},
	{
		Name:  "is_float",
		Value: builtinIsFloat,
	},
	{
		Name:  "is_string",
		Value: builtinIsString,
	},
	{
		Name:  "is_bool",
		Value: builtinIsBool,
	},
	{
		Name:  "is_char",
		Value: builtinIsChar,
	},
	{
		Name:  "is_bytes",
		Value: builtinIsBytes,
	},
	{
		Name:  "is_array",
		Value: builtinIsArray,
	},
	{
		Name:  "is_immutable_array",
		Value: builtinIsImmutableArray,
	},
	{
		Name:  "is_map",
		Value: builtinIsMap,
	},
	{
		Name:  "is_immutable_map",
		Value: builtinIsImmutableMap,
	},
	{
		Name:  "is_time",
		Value: builtinIsTime,
	},
	{
		Name:  "is_error",
		Value: builtinIsError,
	},
	{
		Name:  "is_undefined",
		Value: builtinIsUndefined,
	},
	{
		Name:  "is_function",
		Value: builtinIsFunction,
	},
	{
		Name:  "is_callable",
		Value: builtinIsCallable,
	},
	{
		Name:  "to_json",
		Value: builtinToJSON,
	},
	{
		Name:  "from_json",
		Value: builtinFromJSON,
	},
	{
		Name:  "type_name",
		Value: builtinTypeName,
	},
}

// AllBuiltinFunctionNames returns a list of all default builtin function names.
func AllBuiltinFunctionNames() []string {
	var names []string
	for _, bf := range Builtins {
		names = append(names, bf.Name)
	}
	return names
}

// GetBuiltinFunctions returns a slice of builtin function objects.
// GetBuiltinFunctions removes the duplicate names, and, the returned builtin functions
// are not guaranteed to be in the same order as names.
func GetBuiltinFunctions(names ...string) []*BuiltinFunction {
	include := make(map[string]bool)
	for _, name := range names {
		include[name] = true
	}

	var builtinFuncs []*BuiltinFunction
	for _, bf := range Builtins {
		if include[bf.Name] {
			bf := bf
			builtinFuncs = append(builtinFuncs, &bf)
		}
	}

	return builtinFuncs
}

// GetAllBuiltinFunctions returns all builtin functions.
func GetAllBuiltinFunctions() []*BuiltinFunction {
	return GetBuiltinFunctions(AllBuiltinFunctionNames()...)
}
