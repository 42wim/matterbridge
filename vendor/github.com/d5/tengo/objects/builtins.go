package objects

// NamedBuiltinFunc is a named builtin function.
type NamedBuiltinFunc struct {
	Name string
	Func CallableFunc
}

// Builtins contains all default builtin functions.
var Builtins = []NamedBuiltinFunc{
	{
		Name: "print",
		Func: builtinPrint,
	},
	{
		Name: "printf",
		Func: builtinPrintf,
	},
	{
		Name: "sprintf",
		Func: builtinSprintf,
	},
	{
		Name: "len",
		Func: builtinLen,
	},
	{
		Name: "copy",
		Func: builtinCopy,
	},
	{
		Name: "append",
		Func: builtinAppend,
	},
	{
		Name: "string",
		Func: builtinString,
	},
	{
		Name: "int",
		Func: builtinInt,
	},
	{
		Name: "bool",
		Func: builtinBool,
	},
	{
		Name: "float",
		Func: builtinFloat,
	},
	{
		Name: "char",
		Func: builtinChar,
	},
	{
		Name: "bytes",
		Func: builtinBytes,
	},
	{
		Name: "time",
		Func: builtinTime,
	},
	{
		Name: "is_int",
		Func: builtinIsInt,
	},
	{
		Name: "is_float",
		Func: builtinIsFloat,
	},
	{
		Name: "is_string",
		Func: builtinIsString,
	},
	{
		Name: "is_bool",
		Func: builtinIsBool,
	},
	{
		Name: "is_char",
		Func: builtinIsChar,
	},
	{
		Name: "is_bytes",
		Func: builtinIsBytes,
	},
	{
		Name: "is_array",
		Func: builtinIsArray,
	},
	{
		Name: "is_immutable_array",
		Func: builtinIsImmutableArray,
	},
	{
		Name: "is_map",
		Func: builtinIsMap,
	},
	{
		Name: "is_immutable_map",
		Func: builtinIsImmutableMap,
	},
	{
		Name: "is_time",
		Func: builtinIsTime,
	},
	{
		Name: "is_error",
		Func: builtinIsError,
	},
	{
		Name: "is_undefined",
		Func: builtinIsUndefined,
	},
	{
		Name: "is_function",
		Func: builtinIsFunction,
	},
	{
		Name: "is_callable",
		Func: builtinIsCallable,
	},
	{
		Name: "to_json",
		Func: builtinToJSON,
	},
	{
		Name: "from_json",
		Func: builtinFromJSON,
	},
	{
		Name: "type_name",
		Func: builtinTypeName,
	},
}
