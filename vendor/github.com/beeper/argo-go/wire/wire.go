// Package wire defines the fundamental Argo wire types, such as String, Varint, Record, Array, etc.,
// and provides utilities for working with these types. It forms the basis for how Argo data
// is structured and represented before being encoded or after being decoded.
package wire

import (
	"fmt"
	"math/big"
	"strings"

	"github.com/beeper/argo-go/label"
)

// BlockKey is a type alias for string, representing a key used to identify a block
// in an Argo stream. This key is often used for deduplication and referencing.
// For example, a block of strings might have the key "String".
type BlockKey string

// TypeKey is a string constant representing the kind of an Argo wire type.
// Each distinct wire type (e.g., StringType, ArrayType) has a unique TypeKey.
// This is used for type assertions and type-based dispatch.
type TypeKey string

const (
	// Primitive types

	// TypeKeyString represents the wire type for UTF-8 encoded strings.
	TypeKeyString TypeKey = "STRING"
	// TypeKeyBoolean represents the wire type for boolean values (true or false).
	TypeKeyBoolean TypeKey = "BOOLEAN"
	// TypeKeyVarint represents the wire type for variable-length integers.
	TypeKeyVarint TypeKey = "VARINT"
	// TypeKeyFloat64 represents the wire type for 64-bit floating-point numbers (IEEE 754).
	TypeKeyFloat64 TypeKey = "FLOAT64"
	// TypeKeyBytes represents the wire type for opaque byte arrays.
	TypeKeyBytes TypeKey = "BYTES"
	// TypeKeyPath represents the wire type for GraphQL paths, used for referring to specific locations within a data structure.
	TypeKeyPath TypeKey = "PATH"

	// Compound types

	// TypeKeyFixed represents the wire type for fixed-length byte arrays.
	TypeKeyFixed TypeKey = "FIXED"
	// TypeKeyBlock represents a block, which is a sequence of values of the same underlying type.
	// Blocks can optionally be deduplicated.
	TypeKeyBlock TypeKey = "BLOCK"
	// TypeKeyNullable represents a wire type that can hold a value of another type or be explicitly null.
	TypeKeyNullable TypeKey = "NULLABLE"
	// TypeKeyArray represents an ordered sequence of values, all of the same underlying type.
	TypeKeyArray TypeKey = "ARRAY"
	// TypeKeyRecord represents a structured collection of named fields, where each field has its own wire type.
	TypeKeyRecord TypeKey = "RECORD"
	// TypeKeyDesc represents a self-describing value, where the value itself carries its type information.
	TypeKeyDesc TypeKey = "DESC"
	// TypeKeyExtensions represents an extension
	TypeKeyExtensions TypeKey = "EXTENSIONS"
)

// AbsentValue is a sentinel value used to indicate that an omittable field in a RecordType
// was not present during decoding. Functions that decode record fields can return
// (AbsentValue, nil) to signal this absence without it being an error.
// It is a pointer to an empty struct to ensure it's a unique, comparable value.
var AbsentValue interface{} = &struct{}{}

// Type is the core interface implemented by all Argo wire types.
// It provides a way to get the TypeKey of the wire type and includes an unexported
// marker method (isWireType) to ensure that only types defined within this package
// can satisfy the interface. This helps maintain a closed set of known wire types.
type Type interface {
	GetTypeKey() TypeKey
	isWireType() // Unexported marker method to restrict implementations to this package.
}

// --- Primitive Types ---

// StringType represents the Argo wire type for UTF-8 encoded strings.
// It implements the Type interface.
// Use the global String instance for this type.
type StringType struct{}

func (StringType) GetTypeKey() TypeKey { return TypeKeyString }
func (StringType) isWireType()         {}

// BooleanType represents the Argo wire type for boolean values (true or false).
// It implements the Type interface.
// Use the global Boolean instance for this type.
type BooleanType struct{}

func (BooleanType) GetTypeKey() TypeKey { return TypeKeyBoolean }
func (BooleanType) isWireType()         {}

// VarintType represents the Argo wire type for variable-length integers.
// It implements the Type interface.
// Use the global Varint instance for this type.
type VarintType struct{}

func (VarintType) GetTypeKey() TypeKey { return TypeKeyVarint }
func (VarintType) isWireType()         {}

// Float64Type represents the Argo wire type for 64-bit floating-point numbers (IEEE 754).
// It implements the Type interface.
// Use the global Float64 instance for this type.
type Float64Type struct{}

func (Float64Type) GetTypeKey() TypeKey { return TypeKeyFloat64 }
func (Float64Type) isWireType()         {}

// BytesType represents the Argo wire type for opaque byte arrays.
// It implements the Type interface.
// Use the global Bytes instance for this type.
type BytesType struct{}

func (BytesType) GetTypeKey() TypeKey { return TypeKeyBytes }
func (BytesType) isWireType()         {}

// PathType represents the Argo wire type for Argo paths.
// Paths are used for referring to specific locations within a data structure.
// It implements the Type interface.
// Use the global Path instance for this type.
type PathType struct{}

func (PathType) GetTypeKey() TypeKey { return TypeKeyPath }
func (PathType) isWireType()         {}

// DescType represents the Argo wire type for self-describing values.
// A self-describing value carries its type information along with the data.
// It implements the Type interface.
// Use the global Desc instance for this type.
type DescType struct{}

func (DescType) GetTypeKey() TypeKey { return TypeKeyDesc }
func (DescType) isWireType()         {}

// ExtensionsType represents the Argo wire type for extensions.
// It implements the Type interface.
// Use the global Extensions instance for this type.
type ExtensionsType struct{}

func (ExtensionsType) GetTypeKey() TypeKey { return TypeKeyExtensions }
func (ExtensionsType) isWireType()         {}

// --- Global instances of primitive types ---

// Global pre-allocated instances of primitive wire types.
// These should be used instead of creating new zero-value structs of these types.
var (
	String     Type = StringType{}  // String is the global instance of StringType.
	Boolean    Type = BooleanType{} // Boolean is the global instance of BooleanType.
	Varint     Type = VarintType{}  // Varint is the global instance of VarintType.
	Float64    Type = Float64Type{} // Float64 is the global instance of Float64Type.
	Bytes      Type = BytesType{}   // Bytes is the global instance of BytesType.
	Path       Type = PathType{}    // Path is the global instance of PathType.
	Desc       Type = DescType{}    // Desc is the global instance of DescType.
	Extensions Type = ExtensionsType{}
)

// --- Compound Types ---

// FixedType represents the Argo wire type for fixed-length byte arrays.
// The Length field specifies the exact number of bytes this type represents.
// It implements the Type interface.
type FixedType struct {
	Length int
}

func (FixedType) GetTypeKey() TypeKey { return TypeKeyFixed }
func (FixedType) isWireType()         {}

// BlockType represents an Argo block, which is a sequence of values all of the same underlying type (Of).
// Blocks have a Key for identification and potential deduplication.
// The Dedupe flag indicates whether values within this block are subject to deduplication.
// It implements the Type interface.
type BlockType struct {
	Of     Type
	Key    BlockKey
	Dedupe bool
}

func (BlockType) GetTypeKey() TypeKey { return TypeKeyBlock }
func (BlockType) isWireType()         {}

// ArrayType represents an Argo array, which is an ordered sequence of values,
// all of the same underlying type (Of).
// It implements the Type interface.
type ArrayType struct {
	Of Type
}

func (ArrayType) GetTypeKey() TypeKey { return TypeKeyArray }
func (ArrayType) isWireType()         {}

// NullableType represents an Argo wire type that can either hold a value of another type (Of)
// or be explicitly null.
// It implements the Type interface.
type NullableType struct {
	Of Type
}

func (NullableType) GetTypeKey() TypeKey { return TypeKeyNullable }
func (NullableType) isWireType()         {}

// Field defines a single field within a RecordType.
// It has a Name, an underlying wire Type (Of), and an Omittable flag.
// If Omittable is true, the field may be absent from an encoded record.
type Field struct {
	Name      string
	Of        Type
	Omittable bool
}

// RecordType represents an Argo record, a structured collection of named Fields.
// Each field has its own wire type. The order of fields in the Fields slice is significant.
// It implements the Type interface.
type RecordType struct {
	Fields []Field
}

func (RecordType) GetTypeKey() TypeKey { return TypeKeyRecord }
func (RecordType) isWireType()         {}

// --- Helper functions for creating types ---

// NewBlockType is a constructor function that creates and returns a new BlockType.
// It initializes the BlockType with the specified underlying type (of),
// block key (key), and deduplication flag (dedupe).
func NewBlockType(of Type, key BlockKey, dedupe bool) BlockType {
	return BlockType{Of: of, Key: key, Dedupe: dedupe}
}

// NewNullableType is a constructor function that creates and returns a new NullableType.
// It initializes the NullableType with the specified underlying type (of)
// that can be made nullable.
func NewNullableType(of Type) NullableType {
	return NullableType{Of: of}
}

// DeduplicateByDefault determines whether a given wire type (t) should have
// its corresponding BlockType deduplicated by default.
// Primitive types like String and Bytes are typically deduplicated by default.
// Other types like numbers or booleans usually are not.
// This function returns an error for types like Array or Record, for which
// block-level deduplication is not directly applicable or meaningful in the same way.
func DeduplicateByDefault(t Type) (bool, error) {
	switch t.GetTypeKey() {
	case TypeKeyString, TypeKeyBytes:
		return true, nil
	case TypeKeyBoolean, TypeKeyVarint, TypeKeyFloat64, TypeKeyPath, TypeKeyFixed, TypeKeyDesc:
		return false, nil
	default:
		return false, fmt.Errorf("programmer error: DeduplicateByDefault does not make sense for type %s", t.GetTypeKey())
	}
}

// MustDeduplicateByDefault is a helper function that calls DeduplicateByDefault(t)
// and panics if an error occurs. This is intended for use during package initialization
// where an error from DeduplicateByDefault would indicate a programming error in defining
// default deduplication behavior for core types.
func MustDeduplicateByDefault(t Type) bool {
	val, err := DeduplicateByDefault(t)
	if err != nil {
		panic(fmt.Sprintf("initialization error for DeduplicateByDefault: %v", err))
	}
	return val
}

// --- Global instances of commonly used complex types ---

// Global pre-allocated instances of frequently used complex wire types.
// These are initialized in the init function.
var (
	// VarintBlock is a block of Varint values, keyed as "Int", with default deduplication.
	VarintBlock Type
	// Location is a RecordType representing a line and column, often used for error reporting.
	// Both fields use VarintBlock.
	Location Type
	// Error is a RecordType representing a structured error, potentially including a message,
	// locations, a path, and extensions.
	Error Type
)

// init initializes the global complex type instances like VarintBlock, Location, and Error.
// This ensures they are ready for use as soon as the package is imported.
func init() {
	VarintBlock = NewBlockType(Varint, "Int", MustDeduplicateByDefault(Varint))

	Location = RecordType{
		Fields: []Field{
			{Name: "line", Of: VarintBlock, Omittable: false},
			{Name: "column", Of: VarintBlock, Omittable: false},
		},
	}

	stringBlock := NewBlockType(String, "String", MustDeduplicateByDefault(String))
	locationArray := ArrayType{Of: Location}

	Error = RecordType{
		Fields: []Field{
			{Name: "message", Of: stringBlock, Omittable: false},
			{Name: "locations", Of: locationArray, Omittable: true},
			{Name: "path", Of: Path, Omittable: true},
			{Name: "extensions", Of: Desc, Omittable: true},
		},
	}
}

// --- Type guard functions ---

// IsString checks if the given Type is StringType. Returns true if it is, false otherwise.
func IsString(t Type) bool { return t.GetTypeKey() == TypeKeyString }

// IsBoolean checks if the given Type is BooleanType. Returns true if it is, false otherwise.
func IsBoolean(t Type) bool { return t.GetTypeKey() == TypeKeyBoolean }

// IsVarint checks if the given Type is VarintType. Returns true if it is, false otherwise.
func IsVarint(t Type) bool { return t.GetTypeKey() == TypeKeyVarint }

// IsFloat64 checks if the given Type is Float64Type. Returns true if it is, false otherwise.
func IsFloat64(t Type) bool { return t.GetTypeKey() == TypeKeyFloat64 }

// IsBytes checks if the given Type is BytesType. Returns true if it is, false otherwise.
func IsBytes(t Type) bool { return t.GetTypeKey() == TypeKeyBytes }

// IsPath checks if the given Type is PathType. Returns true if it is, false otherwise.
func IsPath(t Type) bool { return t.GetTypeKey() == TypeKeyPath }

// IsFixed checks if the given Type is FixedType. Returns true if it is, false otherwise.
func IsFixed(t Type) bool { return t.GetTypeKey() == TypeKeyFixed }

// IsDesc checks if the given Type is DescType. Returns true if it is, false otherwise.
func IsDesc(t Type) bool { return t.GetTypeKey() == TypeKeyDesc }

// IsBlock checks if the given Type is BlockType. Returns true if it is, false otherwise.
func IsBlock(t Type) bool { return t.GetTypeKey() == TypeKeyBlock }

// IsArray checks if the given Type is ArrayType. Returns true if it is, false otherwise.
func IsArray(t Type) bool { return t.GetTypeKey() == TypeKeyArray }

// IsNullable checks if the given Type is NullableType. Returns true if it is, false otherwise.
func IsNullable(t Type) bool { return t.GetTypeKey() == TypeKeyNullable }

// IsRecord checks if the given Type is RecordType. Returns true if it is, false otherwise.
func IsRecord(t Type) bool { return t.GetTypeKey() == TypeKeyRecord }

// IsLabeled checks if values of the given wire type (wt) are expected to start with a Label
// in the Argo binary encoding. This is true for types like Nullable, String, Boolean, Bytes, and Array.
// For a BlockType, it recursively checks if the underlying element type is labeled.
// It panics if it encounters a BlockType that doesn't conform to the expected structure
// (which indicates a programming error).
// Other types (e.g., Varint, Float64, Fixed, Path, Desc, Record) are not directly prefixed by a Label.
func IsLabeled(wt Type) bool {
	switch wt.GetTypeKey() {
	case TypeKeyNullable, TypeKeyString, TypeKeyBoolean, TypeKeyBytes, TypeKeyArray:
		return true
	case TypeKeyBlock:
		if bt, ok := wt.(BlockType); ok {
			return IsLabeled(bt.Of)
		}
		// Should not happen if type system is used correctly
		panic(fmt.Sprintf("IsLabeled: expected BlockType, got %T", wt))
	default:
		return false
	}
}

// Print generates a human-readable string representation of a wire type (wt).
// It formats the type structure with indentation for readability, useful for debugging
// or displaying type information.
// Example:
// RECORD (
//
//	"field1": STRING
//	"field2": NULLABLE (
//	  VARINT
//	)
//
// )
func Print(wt Type) string {
	return printRecursive(wt, 0)
}

// printRecursive is a helper for Print. It recursively builds the string representation
// of a wire type, using the 'indent' parameter to manage nesting levels for compound types
// like Record, Array, Block, and Nullable.
func printRecursive(wt Type, indent int) string {
	indentStr := func(plus int) string {
		return strings.Repeat(" ", indent+plus)
	}

	inner := func() string {
		switch t := wt.(type) {
		case StringType, VarintType, BooleanType, Float64Type, BytesType, PathType, DescType, ExtensionsType:
			return string(t.GetTypeKey())
		case NullableType:
			// The TS version `recurse(wt.of) + '?'` implies the recursed string includes its own indent.
			return printRecursive(t.Of, indent+1) + "?"
		case FixedType:
			return fmt.Sprintf("%s(%d)", t.GetTypeKey(), t.Length)
		case BlockType:
			// The TS version `recurse(wt.of) + (wt.dedupe ? '<' : '{') + wt.key + (wt.dedupe ? '>' : '}')`
			// implies the recursed string includes its own indent.
			brackets := "{}"
			if t.Dedupe {
				brackets = "<>"
			}
			return printRecursive(t.Of, indent+1) + string(brackets[0]) + string(t.Key) + string(brackets[1])
		case ArrayType:
			// The TS version `recurse(wt.of) + '[]'` implies the recursed string includes its own indent.
			return printRecursive(t.Of, indent+1) + "[]"
		case RecordType:
			var fieldStrings []string
			for _, field := range t.Fields {
				omittableMarker := ""
				if field.Omittable {
					omittableMarker = "?"
				}
				// TS: `${name}${omittable ? '?' : ''}: ${recurse(type).trimStart()}`
				// Here, trim the leading space from the recursive call to align field type info.
				fieldTypeStr := strings.TrimSpace(printRecursive(field.Of, indent+1))
				fieldStrings = append(fieldStrings,
					fmt.Sprintf("%s%s%s: %s", indentStr(1), field.Name, omittableMarker, fieldTypeStr),
				)
			}
			return "{\n" + strings.Join(fieldStrings, "\n") + "\n" + indentStr(0) + "}"
		default:
			panic(fmt.Sprintf("programmer error: printRecursive can't handle type %T with key %s", wt, wt.GetTypeKey()))
		}
	}
	return indentStr(0) + inner()
}

// PathToWirePath converts a human-readable path (a slice of strings and integers representing
// record field names and array indices) into a wire path (a slice of integers representing
// field/element indices) for a given wire type (wt).
// This is used to translate a user-friendly path into the compact numerical representation
// used in the Argo binary format (e.g., for error reporting or targeted data access).
// Returns an error if the path is invalid for the given wire type (e.g., a string field name
// used for an array, or an index out of bounds).
func PathToWirePath(wt Type, path []interface{}) ([]int, error) {
	if len(path) == 0 {
		return []int{}, nil
	}

	current := path[0]
	tail := path[1:]

	switch t := wt.(type) {
	case BlockType:
		return PathToWirePath(t.Of, path) // Pass full path, block doesn't consume a path element
	case NullableType:
		return PathToWirePath(t.Of, path) // Pass full path, nullable doesn't consume
	case ArrayType:
		arrayIdx, ok := current.(int)
		if !ok {
			return nil, fmt.Errorf("array index must be numeric, got: %v (type %T)", current, current)
		}
		if arrayIdx < 0 {
			return nil, fmt.Errorf("array index must be non-negative, got: %d", arrayIdx)
		}
		subPath, err := PathToWirePath(t.Of, tail)
		if err != nil {
			return nil, err
		}
		return append([]int{arrayIdx}, subPath...), nil
	case RecordType:
		fieldName, ok := current.(string)
		if !ok {
			return nil, fmt.Errorf("record field name must be a string, got: %v (type %T)", current, current)
		}
		fieldIndex := -1
		for i, f := range t.Fields {
			if f.Name == fieldName {
				fieldIndex = i
				break
			}
		}
		if fieldIndex == -1 {
			return nil, fmt.Errorf("encoding error: could not find record field: %s", fieldName)
		}
		field := t.Fields[fieldIndex]
		subPath, err := PathToWirePath(field.Of, tail)
		if err != nil {
			return nil, err
		}
		return append([]int{fieldIndex}, subPath...), nil
	case StringType, VarintType, BooleanType, Float64Type, BytesType, PathType, DescType, FixedType:
		if len(path) > 0 { // Path is not empty, but primitive type cannot be indexed further
			return nil, fmt.Errorf("encoding error: path %v attempts to index into primitive type %s", path, t.GetTypeKey())
		}
		return []int{}, nil
	default:
		panic(fmt.Sprintf("programmer error: PathToWirePath can't handle type %T with key %s", wt, wt.GetTypeKey()))
	}
}

// WirePathToPath converts a wire path (a slice of integers representing field/element indices)
// back into a human-readable path (a slice of strings for record field names and integers for
// array indices) for a given wire type (wt).
// This is the reverse of PathToWirePath and is useful for presenting internal Argo paths
// in a more understandable format.
// Returns an error if the wire path is invalid for the given wire type (e.g., an index
// is out of bounds for a record or array).
func WirePathToPath(wt Type, wirePath []int) ([]interface{}, error) {
	if len(wirePath) == 0 {
		return []interface{}{}, nil
	}

	currentIndex := wirePath[0]
	tailPath := wirePath[1:]

	switch t := wt.(type) {
	case BlockType:
		return WirePathToPath(t.Of, wirePath) // Pass full wirePath
	case NullableType:
		return WirePathToPath(t.Of, wirePath) // Pass full wirePath
	case ArrayType:
		// Array index is directly the current wirePath element
		subPath, err := WirePathToPath(t.Of, tailPath)
		if err != nil {
			return nil, err
		}
		return append([]interface{}{currentIndex}, subPath...), nil
	case RecordType:
		if currentIndex < 0 || currentIndex >= len(t.Fields) {
			return nil, fmt.Errorf("encoding error: could not find record field by index: %d (record has %d fields)", currentIndex, len(t.Fields))
		}
		field := t.Fields[currentIndex]
		subPath, err := WirePathToPath(field.Of, tailPath)
		if err != nil {
			return nil, err
		}
		return append([]interface{}{field.Name}, subPath...), nil
	case StringType, VarintType, BooleanType, Float64Type, BytesType, PathType, DescType, FixedType:
		if len(wirePath) > 0 { // wirePath is not empty, but primitive type cannot be indexed further
			return nil, fmt.Errorf("encoding error: wirePath %v attempts to index into primitive type %s", wirePath, t.GetTypeKey())
		}
		return []interface{}{}, nil
	default:
		panic(fmt.Sprintf("programmer error: WirePathToPath can't handle type %T with key %s", wt, wt.GetTypeKey()))
	}
}

// --- SelfDescribing Namespace ---

// SelfDescribingTypeMarkers are label.Label instances used to mark self-describing types.
var (
	SelfDescribingTypeMarkerNull   = label.NullMarker
	SelfDescribingTypeMarkerFalse  = label.FalseMarker
	SelfDescribingTypeMarkerTrue   = label.TrueMarker
	SelfDescribingTypeMarkerObject = label.NewFromInt64(2)
	SelfDescribingTypeMarkerList   = label.NewFromInt64(3)
	SelfDescribingTypeMarkerString = label.NewFromInt64(4)
	SelfDescribingTypeMarkerBytes  = label.NewFromInt64(5)
	SelfDescribingTypeMarkerInt    = label.NewFromInt64(6)
	SelfDescribingTypeMarkerFloat  = label.NewFromInt64(7)
)

// SelfDescribing is a collection of pre-encoded byte slices for self-describing type markers.
var (
	SelfDescribingNull   = label.Null  // From label package []byte
	SelfDescribingFalse  = label.False // From label package []byte
	SelfDescribingTrue   = label.True  // From label package []byte
	SelfDescribingObject = SelfDescribingTypeMarkerObject.Encode()
	SelfDescribingList   = SelfDescribingTypeMarkerList.Encode()
	SelfDescribingString = SelfDescribingTypeMarkerString.Encode()
	SelfDescribingBytes  = SelfDescribingTypeMarkerBytes.Encode()
	SelfDescribingInt    = SelfDescribingTypeMarkerInt.Encode()
	SelfDescribingFloat  = SelfDescribingTypeMarkerFloat.Encode()
)

// SelfDescribingBlocks is a map from BlockKey to the Type of the elements in that block.
// This is used when decoding self-describing values that are blocks, to know the type
// of the items within the block from its key.
// It's initialized in the second init function to ensure all base types are defined.
var SelfDescribingBlocks map[BlockKey]Type

// init (the second one in this file) initializes SelfDescribingBlocks.
// It populates the map with common block types that might be used in self-describing contexts.
// This is done in a separate init to avoid initialization cycles, ensuring that all base types
// (like String, Varint) and compound types derived from them (like stringBlock, VarintBlock)
// are fully defined before being added to this map.
func init() { // Second init for SelfDescribingBlocks to ensure base types are ready
	stringBlock := NewBlockType(String, "String", MustDeduplicateByDefault(String))
	bytesBlock := NewBlockType(Bytes, "Bytes", MustDeduplicateByDefault(Bytes))
	// Note: VarintBlock is already defined globally and initialized in the first init.
	// Create a Float64 block type for self-describing floats.
	float64Block := NewBlockType(Float64, "Float", MustDeduplicateByDefault(Float64))

	SelfDescribingBlocks = map[BlockKey]Type{
		stringBlock.Key:             stringBlock.Of,
		bytesBlock.Key:              bytesBlock.Of,
		VarintBlock.(BlockType).Key: VarintBlock.(BlockType).Of,
		float64Block.Key:            float64Block.Of,
	}
}

// bigInt is a small utility function to create a *big.Int from an int64 value.
// Primarily used for testing or when *big.Int literals are needed.
func bigInt(val int64) *big.Int {
	return big.NewInt(val)
}

// Ensure all Type implementations satisfy the Type interface.
var _ Type = StringType{}
var _ Type = BooleanType{}
var _ Type = VarintType{}
var _ Type = Float64Type{}
var _ Type = BytesType{}
var _ Type = PathType{}
var _ Type = FixedType{}
var _ Type = BlockType{}
var _ Type = ArrayType{}
var _ Type = NullableType{}
var _ Type = RecordType{}
var _ Type = DescType{}
