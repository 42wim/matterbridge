// Package block provides BlockWriter and DeduplicatingBlockWriter types for
// preparing Argo data blocks. These writers manage the conversion of values
// to their byte representations and the generation of corresponding labels,
// which are essential for the Argo binary format.
package block

import (
	"fmt"
	"math/big"
	"reflect" // For robust nil checking in DeduplicatingBlockWriter

	"github.com/beeper/argo-go/label"
	"github.com/beeper/argo-go/pkg/buf"
	"github.com/elliotchance/orderedmap/v3" // Added for deterministic maps
)

// MakeLabelFunc defines a function signature for creating a Label for a given input value
// and its byte representation. It returns a pointer to a label, or nil if no
// label should be generated. It can also return an error if label creation fails.
type MakeLabelFunc[In any] func(value In, data []byte) (l *label.Label, err error)

// ValueToBytesFunc defines a function signature for converting an input value to its byte representation.
// It returns the byte data or an error if the conversion fails.
type ValueToBytesFunc[In any] func(value In) (data []byte, err error)

// BlockWriter is responsible for converting input values of type `In` into their byte
// representations and generating corresponding labels using provided functions.
// It accumulates the byte representations internally. The actual writing of labels
// and byte data to output streams is typically handled by the caller.
type BlockWriter[In any] struct {
	makeLabelFunc    MakeLabelFunc[In]
	valueToBytesFunc ValueToBytesFunc[In]
	valuesAsBytes    [][]byte // valuesAsBytes stores the byte representations of all values processed by the Write method.
}

// NewBlockWriter creates and returns a new BlockWriter configured with the
// provided makeLabel and valueToBytes functions.
// If either function is nil, the Write method will subsequently fail.
func NewBlockWriter[In any](
	makeLabel MakeLabelFunc[In],
	valueToBytes ValueToBytesFunc[In],
) *BlockWriter[In] {
	// Note: Write methods will check for nil makeLabelFunc or valueToBytesFunc.
	return &BlockWriter[In]{
		makeLabelFunc:    makeLabel,
		valueToBytesFunc: valueToBytes,
		valuesAsBytes:    make([][]byte, 0),
	}
}

// NewLengthOfBytesBlockWriter creates a BlockWriter that generates labels
// representing the length (in bytes) of each value's binary representation.
func NewLengthOfBytesBlockWriter[In any](valueToBytes ValueToBytesFunc[In]) *BlockWriter[In] {
	makeLabel := func(v In, data []byte) (*label.Label, error) {
		l := label.NewFromInt64(int64(len(data)))
		return &l, nil
	}
	return NewBlockWriter(makeLabel, valueToBytes)
}

// NewNoLabelBlockWriter creates a BlockWriter that does not generate any labels.
// Its makeLabelFunc will always return (nil, nil).
func NewNoLabelBlockWriter[In comparable](valueToBytes ValueToBytesFunc[In]) *BlockWriter[In] {
	makeLabel := func(v In, data []byte) (*label.Label, error) {
		return nil, nil
	}
	return NewBlockWriter(makeLabel, valueToBytes)
}

// AfterNewWrite is a hook method called by Write after a new value's byte
// representation has been successfully generated and stored.
// Its default implementation is a no-op.
func (bw *BlockWriter[In]) AfterNewWrite() {
	// Default no-op
}

// Write converts the given value `v` to its byte representation using valueToBytesFunc,
// stores these bytes internally, and then generates a label using makeLabelFunc.
// It returns the generated label (or nil if no label is to be generated) and any
// error encountered during the process. This method calls AfterNewWrite after
// successfully storing the bytes.
func (bw *BlockWriter[In]) Write(v In) (*label.Label, error) {
	if bw.valueToBytesFunc == nil {
		return nil, fmt.Errorf("BlockWriter.Write: valueToBytesFunc is nil")
	}
	bytes, err := bw.valueToBytesFunc(v)
	if err != nil {
		return nil, fmt.Errorf("BlockWriter.Write: valueToBytesFunc failed: %w", err)
	}

	bw.valuesAsBytes = append(bw.valuesAsBytes, bytes)
	bw.AfterNewWrite() // Call hook after new bytes are stored.

	if bw.makeLabelFunc == nil {
		return nil, fmt.Errorf("BlockWriter.Write: makeLabelFunc is nil")
	}
	l, err := bw.makeLabelFunc(v, bytes)
	if err != nil {
		return nil, fmt.Errorf("BlockWriter.Write: makeLabelFunc failed: %w", err)
	}
	return l, nil
}

// WriteLastToBuf writes the byte representation of the most recently processed value
// to the provided buf.Write buffer. This method is primarily intended for scenarios
// like "noBlocks" mode where values are written directly. It does not remove the value
// from its internal store. For block construction, AllValuesAsBytes is typically used.
func (bw *BlockWriter[In]) WriteLastToBuf(buf buf.Write) error {
	if len(bw.valuesAsBytes) == 0 {
		return fmt.Errorf("BlockWriter.WriteLastToBuf: called on empty BlockWriter")
	}

	lastValueIndex := len(bw.valuesAsBytes) - 1
	lastValueBytes := bw.valuesAsBytes[lastValueIndex]
	// The value is not popped from bw.valuesAsBytes, to support modes like InlineEverything.

	// buf.Write(nil) is typically a no-op or writes 0 bytes.
	_, err := buf.Write(lastValueBytes)
	if err != nil {
		// Note: The original value remains in valuesAsBytes even if this write fails.
		// This behavior is simpler than attempting to revert internal state on write failure.
		return fmt.Errorf("BlockWriter.WriteLastToBuf: buf.Write failed: %w", err)
	}
	return nil
}

// AllValuesAsBytes returns a new slice containing all byte arrays accumulated by the BlockWriter.
// This is typically used when constructing a final value block from all processed items.
// The returned slice is a copy of the slice header, but the underlying byte arrays are shared.
func (bw *BlockWriter[In]) AllValuesAsBytes() [][]byte {
	vals := make([][]byte, len(bw.valuesAsBytes))
	copy(vals, bw.valuesAsBytes)
	return vals
}

// DeduplicatingBlockWriter embeds BlockWriter and extends its functionality to support
// value deduplication. When a value is processed, if it has been seen before, a
// backreference label is returned. For new values, a unique ID is assigned (and used
// for future backreferences), and then a label is generated using its labelForNew function.
// The input type `In` must be comparable to be used as a key for tracking seen values.
type DeduplicatingBlockWriter[In comparable] struct {
	// Embeds BlockWriter. The embedded `makeLabelFunc` is not used by DeduplicatingBlockWriter's `Write` method.
	// `valueToBytesFunc`, `valuesAsBytes`, and `AfterNewWrite` are inherited and used.
	BlockWriter[In]

	seen        *orderedmap.OrderedMap[In, label.Label] // Stores seen values and their assigned backreference labels.
	lastIDValue *big.Int                                // Stores the numeric value of the last assigned backreference ID.
	labelForNew MakeLabelFunc[In]                       // Function to generate labels for new, non-backreferenced items.
}

// NewDeduplicatingBlockWriter creates and returns a new DeduplicatingBlockWriter.
// It requires a `labelForNew` function (to generate labels for unique items) and
// a `valueToBytes` function (to convert values to bytes).
// Backreference IDs are initialized based on label.LowestReservedValue.
// If either function is nil, the Write method will subsequently fail.
func NewDeduplicatingBlockWriter[In comparable](
	labelForNew MakeLabelFunc[In],
	valueToBytes ValueToBytesFunc[In],
) *DeduplicatingBlockWriter[In] {
	// Note: Write methods will check for nil labelForNew or valueToBytesFunc.
	initialIDVal := new(big.Int).Set(label.LowestReservedValue.Value())

	return &DeduplicatingBlockWriter[In]{
		BlockWriter: BlockWriter[In]{
			valueToBytesFunc: valueToBytes,
			valuesAsBytes:    make([][]byte, 0),
			// The embedded makeLabelFunc is not used by DeduplicatingBlockWriter's Write method;
			// it could be set to nil, as DeduplicatingBlockWriter uses its own labelForNew.
			makeLabelFunc: nil,
		},
		seen:        orderedmap.NewOrderedMap[In, label.Label](),
		lastIDValue: initialIDVal,
		labelForNew: labelForNew,
	}
}

// NewLengthOfBytesDeduplicatingBlockWriter creates a DeduplicatingBlockWriter
// where labels for new (non-duplicate) values are generated based on the length
// of their byte representation.
func NewLengthOfBytesDeduplicatingBlockWriter[In comparable](
	valueToBytes ValueToBytesFunc[In],
) *DeduplicatingBlockWriter[In] {
	labelForNew := func(v In, data []byte) (*label.Label, error) {
		l := label.NewFromInt64(int64(len(data)))
		return &l, nil
	}
	return NewDeduplicatingBlockWriter[In](labelForNew, valueToBytes)
}

// nextID generates and returns the next sequential backreference ID.
// IDs are generated by decrementing from label.LowestReservedValue.
func (dbw *DeduplicatingBlockWriter[In]) nextID() label.Label {
	one := big.NewInt(1)
	dbw.lastIDValue.Sub(dbw.lastIDValue, one)
	idCopy := new(big.Int).Set(dbw.lastIDValue) // Use a copy for the new Label.
	return label.New(idCopy)
}

// labelForValue determines the appropriate label for a given value `v` *before* its byte conversion.
// It handles three cases:
//  1. If `v` is nil (checked robustly using reflection), it returns label.NullMarker.
//  2. If `v` has been seen before, it returns the stored backreference label.
//  3. If `v` is new and not nil, it assigns a new backreference ID, stores it with `v` in the seen map,
//     and returns (nil, nil) to signal that the main Write method should proceed with byte conversion
//     and new-item label generation via `labelForNew`.
func (dbw *DeduplicatingBlockWriter[In]) labelForValue(v In) (*label.Label, error) {
	// Check if 'v' is a nil pointer or nil interface using reflection for robustness.
	valOfV := reflect.ValueOf(v)
	isConsideredNil := false
	switch valOfV.Kind() {
	case reflect.Ptr, reflect.Interface, reflect.Map, reflect.Slice, reflect.Chan, reflect.Func:
		isConsideredNil = valOfV.IsNil()
	}

	if isConsideredNil {
		lm := label.NullMarker // label.NullMarker is a value, return its address.
		return &lm, nil
	}

	// Check if the non-nil value has been seen before.
	if savedLabel, found := dbw.seen.Get(v); found {
		return &savedLabel, nil // Return pointer to the copy of the stored label from Get.
	}

	// Value is new and not nil. Assign a new ID and store it.
	newID := dbw.nextID()
	dbw.seen.Set(v, newID)
	return nil, nil // Indicates it's new; Write method will continue processing.
}

// Write processes the value `v` for the DeduplicatingBlockWriter.
// It first calls `labelForValue` to check if `v` is nil or a duplicate.
//   - If `labelForValue` returns a non-nil label (NullMarker or backreference), Write returns that label immediately.
//   - If `labelForValue` returns `(nil, nil)` (indicating `v` is new, non-nil, and has been assigned a new
//     backreference ID in the `seen` map), Write then converts `v` to bytes, stores the bytes (appending to
//     the embedded BlockWriter's `valuesAsBytes`), calls `AfterNewWrite`, and finally generates a label for this
//     new item using `dbw.labelForNew`.
//
// This method effectively overrides the Write method of the embedded BlockWriter.
func (dbw *DeduplicatingBlockWriter[In]) Write(v In) (*label.Label, error) {
	// Determine if 'v' is nil, a backreference, or new.
	// labelForValue handles nil check, seen map lookup, and registers new items in seen map.
	existingLabel, err := dbw.labelForValue(v)
	if err != nil {
		// This error typically indicates an internal issue within labelForValue.
		return nil, fmt.Errorf("DeduplicatingBlockWriter.Write: labelForValue failed: %w", err)
	}

	if existingLabel != nil {
		// If existingLabel is not nil, it's either Label.NullMarker or a backreference.
		return existingLabel, nil
	}

	// If existingLabel is nil, 'v' is a new, non-nil value.
	// It has already been added to dbw.seen with a new ID by labelForValue.
	// Now, convert to bytes, store them, and generate the "new item" label.
	if dbw.valueToBytesFunc == nil {
		return nil, fmt.Errorf("DeduplicatingBlockWriter.Write: valueToBytesFunc is nil (from embedded BlockWriter)")
	}
	bytes, err := dbw.valueToBytesFunc(v)
	if err != nil {
		// If conversion to bytes fails, the ID was already assigned in `seen`.
		// This order (ID assignment before byte conversion) is consistent with the reference
		// implementation and is maintained here.
		return nil, fmt.Errorf("DeduplicatingBlockWriter.Write: valueToBytesFunc failed: %w", err)
	}

	// Use the inherited valuesAsBytes and AfterNewWrite
	dbw.valuesAsBytes = append(dbw.valuesAsBytes, bytes)
	dbw.AfterNewWrite() // Calls the AfterNewWrite method of the embedded BlockWriter.

	if dbw.labelForNew == nil {
		return nil, fmt.Errorf("DeduplicatingBlockWriter.Write: labelForNew is nil")
	}
	finalLabel, err := dbw.labelForNew(v, bytes)
	if err != nil {
		return nil, fmt.Errorf("DeduplicatingBlockWriter.Write: labelForNew failed: %w", err)
	}
	// If `labelForNew` returns `(nil, nil)`, that's the equivalent of returning no label for the new item.
	return finalLabel, nil
}

// ToDeduplicating on a DeduplicatingBlockWriter returns the receiver `dbw` itself,
// as it's already a deduplicating writer.
func (dbw *DeduplicatingBlockWriter[In]) ToDeduplicating() *DeduplicatingBlockWriter[In] {
	return dbw
}

// ToDeduplicatingWriter is a package-level utility function that converts a given
// BlockWriter[In] to a DeduplicatingBlockWriter[In]. The type parameter `In` for
// the input BlockWriter must be `comparable`.
// It initializes the new DeduplicatingBlockWriter using the `makeLabelFunc` (as `labelForNew`)
// and `valueToBytesFunc` from the original BlockWriter.
func ToDeduplicatingWriter[In comparable](
	bw *BlockWriter[In], // The input BlockWriter's type parameter must be comparable.
) *DeduplicatingBlockWriter[In] {
	// This conversion is type-safe because the Go compiler ensures that 'In'
	// for 'bw' satisfies 'comparable' at the call site of this function.
	// Thus, 'In' can be used for DeduplicatingBlockWriter's type parameter.
	return NewDeduplicatingBlockWriter[In](
		bw.makeLabelFunc, // Use original BlockWriter's makeLabel as labelForNew.
		bw.valueToBytesFunc,
	)
}

// AnyBlockWriter defines an interface for block writers that accept `interface{}` (any) values.
// This allows for type-erased handling of block writing operations, useful in contexts
// where the specific type of values being written is not known at compile time or varies.
// Implementations typically wrap a generic BlockWriter[T] or DeduplicatingBlockWriter[T]
// and perform a type assertion before delegating to the underlying writer.
type AnyBlockWriter interface {
	// Write processes a value `v` of any type. It attempts to assert `v` to the
	// concrete type expected by the underlying block writer. If successful, it delegates
	// to the underlying writer's Write method. It returns the generated label and any error.
	Write(v interface{}) (*label.Label, error)
	// AllValuesAsBytes returns all accumulated byte arrays from the underlying writer.
	AllValuesAsBytes() [][]byte
	// WriteLastToBuf writes the byte representation of the most recently processed value
	// to the provided buffer, by delegating to the underlying writer.
	WriteLastToBuf(buf buf.Write) error
}

// --- Non-Deduplicating Adapter ---

// nonDeduplicatingAdapter adapts a generic BlockWriter[T] to the AnyBlockWriter interface.
// The type parameter T can be any type. This adapter allows a type-specific BlockWriter
// to be used in contexts requiring an AnyBlockWriter.
type nonDeduplicatingAdapter[T any] struct {
	coreWriter *BlockWriter[T]
}

// Write attempts to assert the input `v` to type T and then calls the underlying coreWriter.Write.
// If the type assertion fails, it returns a type mismatch error.
func (a *nonDeduplicatingAdapter[T]) Write(v interface{}) (*label.Label, error) {
	val, ok := v.(T)
	if !ok {
		// This error occurs if T is a concrete type and v is not assignable to it (e.g., T is string, v is int).
		// If T is interface{}, this assertion will always succeed.
		return nil, fmt.Errorf("type mismatch for block writer: value of type %T cannot be asserted to the writer's target type. Value: %v", v, v)
	}
	return a.coreWriter.Write(val)
}

// AllValuesAsBytes delegates to the underlying coreWriter.
func (a *nonDeduplicatingAdapter[T]) AllValuesAsBytes() [][]byte {
	return a.coreWriter.AllValuesAsBytes()
}

// WriteLastToBuf delegates to the underlying coreWriter.
func (a *nonDeduplicatingAdapter[T]) WriteLastToBuf(buf buf.Write) error {
	return a.coreWriter.WriteLastToBuf(buf)
}

// NewAnyBlockWriter creates an AnyBlockWriter by wrapping a given generic BlockWriter[T].
// This allows the specifically typed BlockWriter to be used through the type-erased AnyBlockWriter interface.
func NewAnyBlockWriter[T any](bw *BlockWriter[T]) AnyBlockWriter {
	return &nonDeduplicatingAdapter[T]{coreWriter: bw}
}

// --- Deduplicating Adapter ---

// deduplicatingAdapter adapts a generic DeduplicatingBlockWriter[T] to the AnyBlockWriter interface.
// The type parameter T must be comparable for the underlying DeduplicatingBlockWriter.
// This adapter allows a type-specific, comparable DeduplicatingBlockWriter to be used as an AnyBlockWriter.
type deduplicatingAdapter[T comparable] struct {
	coreWriter *DeduplicatingBlockWriter[T]
}

// Write attempts to assert the input `v` to type T (which must be comparable) and then
// calls the underlying coreWriter.Write. If the type assertion fails, it returns a type mismatch error.
func (a *deduplicatingAdapter[T]) Write(v interface{}) (*label.Label, error) {
	val, ok := v.(T)
	if !ok {
		var zeroT T // Used to get a string representation of type T for the error message.
		return nil, fmt.Errorf("type mismatch for deduplicating block writer: expected %T, got %T for value %v", zeroT, v, v)
	}
	return a.coreWriter.Write(val)
}

// AllValuesAsBytes delegates to the underlying coreWriter.
func (a *deduplicatingAdapter[T]) AllValuesAsBytes() [][]byte {
	return a.coreWriter.AllValuesAsBytes()
}

// WriteLastToBuf delegates to the underlying coreWriter.
func (a *deduplicatingAdapter[T]) WriteLastToBuf(buf buf.Write) error {
	return a.coreWriter.WriteLastToBuf(buf)
}

// NewAnyDeduplicatingBlockWriter creates an AnyBlockWriter by wrapping a given generic DeduplicatingBlockWriter[T].
// The type T must be comparable. This allows the specifically typed DeduplicatingBlockWriter
// to be used through the type-erased AnyBlockWriter interface.
func NewAnyDeduplicatingBlockWriter[T comparable](dbw *DeduplicatingBlockWriter[T]) AnyBlockWriter {
	return &deduplicatingAdapter[T]{coreWriter: dbw}
}

// --- Specialized Constructors returning AnyBlockWriter ---

// NewAnyNoLabelBlockWriter creates a BlockWriter specifically for `interface{}` values (T is interface{})
// that generates no labels, and then wraps it as an AnyBlockWriter.
// This is a convenience constructor for a common use case.
func NewAnyNoLabelBlockWriter(valueToBytes ValueToBytesFunc[interface{}]) AnyBlockWriter {
	// The underlying BlockWriter is instantiated with T as interface{}.
	bw := NewNoLabelBlockWriter[interface{}](valueToBytes)
	// This is then wrapped by the nonDeduplicatingAdapter, which also expects T as interface{}.
	return NewAnyBlockWriter[interface{}](bw)
}

// NewAnyLengthOfBytesBlockWriter creates a BlockWriter specifically for `interface{}` values (T is interface{})
// that generates labels based on byte length, and then wraps it as an AnyBlockWriter.
// This is a convenience constructor for another common use case.
func NewAnyLengthOfBytesBlockWriter(valueToBytes ValueToBytesFunc[interface{}]) AnyBlockWriter {
	// The underlying BlockWriter is instantiated with T as interface{}.
	bw := NewLengthOfBytesBlockWriter[interface{}](valueToBytes)
	// This is then wrapped by the nonDeduplicatingAdapter, which also expects T as interface{}.
	return NewAnyBlockWriter[interface{}](bw)
}
