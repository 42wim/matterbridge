package objects

// Builtins contains all default builtin functions.
// Use GetBuiltinFunctions instead of accessing Builtins directly.
var Builtins = []*BuiltinFunction{
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
		Name:  "is_iterable",
		Value: builtinIsIterable,
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
		Name:  "type_name",
		Value: builtinTypeName,
	},
	{
		Name:  "format",
		Value: builtinFormat,
	},
}
