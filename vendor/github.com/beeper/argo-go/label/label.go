// Package label defines the Label type and its associated operations for Argo.
// Labels are signed variable-length integers (encoded using ZigZag ULEB128)
// that serve multiple purposes in the Argo binary format:
//   - Encoding lengths of data segments (e.g., string length, array size).
//   - Representing special marker values like Null, Absent, or Error.
//   - Encoding backreferences to previously seen values for data deduplication.
//
// The package provides constants for common marker labels (Null, Absent, True, False, etc.),
// functions to create labels from integer values, methods to determine a label's kind
// (Length, Null, Backreference, etc.), and functions for encoding labels to and
// decoding labels from byte streams.
package label

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/beeper/argo-go/pkg/buf"
	"github.com/beeper/argo-go/pkg/varint"
)

// LabelKind categorizes a Label based on its numerical value and Argo semantics.
// Different kinds determine how the label (and potentially associated data) is interpreted.
type LabelKind int

const (
	// LabelKindNull (-1) indicates a null or missing value where the field itself is present.
	LabelKindNull LabelKind = iota
	// LabelKindAbsent (-2) indicates a field or array entry that is entirely absent or omitted.
	LabelKindAbsent
	// LabelKindError (-3) indicates that an error occurred at this position in the data structure.
	// The error details may follow or be found in a separate error list.
	LabelKindError
	// LabelKindBackreference (negative, < -3) indicates a reference to a previously encountered value.
	// The specific negative value is used to calculate an offset to the original value.
	LabelKindBackreference
	// LabelKindLength (non-negative, >= 0) indicates the length of a subsequent data value (e.g., string bytes, array elements).
	// It can also represent boolean true (1) or false/non-null marker (0) in specific contexts.
	LabelKindLength
)

// String returns a human-readable name for the LabelKind.
func (lk LabelKind) String() string {
	switch lk {
	case LabelKindNull:
		return "Null"
	case LabelKindAbsent:
		return "Absent"
	case LabelKindError:
		return "Error"
	case LabelKindBackreference:
		return "Backreference"
	case LabelKindLength:
		// This covers lengths, true (1), false (0), and non-null (0).
		return "Length/Boolean/Marker"
	default:
		return fmt.Sprintf("UnknownLabelKind(%d)", int(lk))
	}
}

// Label wraps a *big.Int to represent an Argo label.
// Its value determines its kind (length, null, backreference, etc.) and specific meaning.
// Labels are immutable once created; their underlying *big.Int should not be modified
// after creation, especially for the global Label constants.
type Label struct {
	value *big.Int
}

// New creates a new Label from the provided *big.Int.
// The caller should not modify `val` after passing it to New, especially if it's
// not a fresh *big.Int, to avoid unintended side effects on the Label.
// If `val` is nil, a Label with a *big.Int value of 0 is created.
func New(val *big.Int) Label {
	if val == nil {
		// Ensure value is never nil internally to prevent panics on method calls.
		// Default to 0 if nil is passed; callers should ideally avoid passing nil.
		return Label{value: big.NewInt(0)}
	}
	return Label{value: val}
}

// NewFromInt64 creates a new Label from an int64 value.
// This is a convenience constructor for creating labels from standard integer types.
func NewFromInt64(val int64) Label {
	return Label{value: big.NewInt(val)}
}

// Value returns the underlying *big.Int of the Label.
// Callers must not modify the returned *big.Int, as it may be shared by
// predefined Label constants (e.g., NullMarker, TrueMarker) or other Label instances.
// To perform arithmetic, create a new *big.Int (e.g., `new(big.Int).Set(l.Value())`).
func (l Label) Value() *big.Int {
	return l.value
}

// String returns the decimal string representation of the Label's underlying *big.Int value.
func (l Label) String() string {
	if l.value == nil {
		// This case should ideally not be reached if constructors always ensure a non-nil l.value.
		return "<uninitialized_label>"
	}
	return l.value.String()
}

// --- Internal *big.Int constants for special label values ---
// These are unexported and used to initialize the exported Label constants and for comparisons.
var (
	trueMarkerBigInt    *big.Int // Represents boolean True (value 1)
	falseMarkerBigInt   *big.Int // Represents boolean False (value 0), also used as NonNull marker
	nonNullMarkerBigInt *big.Int // Represents a NonNull marker (value 0), indicating presence for non-labeled types
	nullMarkerBigInt    *big.Int // Represents a Null value (value -1)
	absentMarkerBigInt  *big.Int // Represents an Absent field/item (value -2)
	errorMarkerBigInt   *big.Int // Represents an Error marker (value -3)

	// lowestReservedBigInt is the numerically largest (i.e., closest to zero) negative *big.Int
	// that has a special, non-backreference meaning. Values numerically smaller than this
	// (e.g., -4, -5, ...) are considered backreferences.
	lowestReservedBigInt *big.Int
)

// --- Exported Label constants ---
// These provide convenient access to common, predefined label values.
var (
	TrueMarker          Label // Label for boolean True (value 1).
	FalseMarker         Label // Label for boolean False (value 0).
	NonNullMarker       Label // Label for NonNull marker (value 0); alias for FalseMarker.
	NullMarker          Label // Label for Null (value -1).
	AbsentMarker        Label // Label for Absent (value -2).
	ErrorMarker         Label // Label for Error (value -3).
	LowestReservedValue Label // Label instance for the lowest reserved non-backreference value (-3).
)

// --- Exported pre-encoded byte slices for common Label values ---
// These are the ZigZag ULEB128 encoded forms of the marker labels.
var (
	Null    []byte // Encoded form of NullMarker (-1).
	Absent  []byte // Encoded form of AbsentMarker (-2).
	Error   []byte // Encoded form of ErrorMarker (-3).
	Zero    []byte // Encoded form of 0 (used for FalseMarker, NonNullMarker, and length 0).
	NonNull []byte // Encoded form of NonNullMarker (0); alias for Zero.
	False   []byte // Encoded form of FalseMarker (0); alias for Zero.
	True    []byte // Encoded form of TrueMarker (1).
)

// labelToOffsetFactor is used in the calculation of backreference offsets.
// It is derived from `lowestReservedBigInt` and ensures that backreference IDs
// map correctly to array/slice offsets.
// Specifically, offset = -NumericValue(backreference_label) + labelToOffsetFactor.
// If lowestReservedBigInt is -3, factor is -3 - 1 = -4.
var labelToOffsetFactor int64

func init() {
	// Initialize *big.Int marker values.
	trueMarkerBigInt = big.NewInt(1)
	falseMarkerBigInt = big.NewInt(0)
	nonNullMarkerBigInt = big.NewInt(0) // By spec, NonNull is 0, same as False.
	nullMarkerBigInt = big.NewInt(-1)
	absentMarkerBigInt = big.NewInt(-2)
	errorMarkerBigInt = big.NewInt(-3)
	// lowestReservedBigInt is set to errorMarkerBigInt (-3). Values less than this are backrefs.
	lowestReservedBigInt = new(big.Int).Set(errorMarkerBigInt) // Use a copy for safety, though *big.Ints from NewInt are new.

	// Initialize exported Label struct constants using the *big.Int markers.
	TrueMarker = New(trueMarkerBigInt)
	FalseMarker = New(falseMarkerBigInt)
	NonNullMarker = New(nonNullMarkerBigInt) // Effectively New(big.NewInt(0))
	NullMarker = New(nullMarkerBigInt)
	AbsentMarker = New(absentMarkerBigInt)
	ErrorMarker = New(errorMarkerBigInt)
	LowestReservedValue = New(lowestReservedBigInt) // Label for -3

	// Initialize pre-encoded byte slices for common labels.
	Null = varint.ZigZagEncode(nullMarkerBigInt)     // ZigZag(-1)
	Absent = varint.ZigZagEncode(absentMarkerBigInt) // ZigZag(-2)
	Error = varint.ZigZagEncode(errorMarkerBigInt)   // ZigZag(-3)
	Zero = varint.ZigZagEncode(big.NewInt(0))        // ZigZag(0)
	NonNull = Zero                                   // NonNull is 0.
	False = Zero                                     // False is 0.
	True = varint.ZigZagEncode(big.NewInt(1))        // ZigZag(1)

	// Calculate labelToOffsetFactor for backreference processing.
	// This factor adjusts the negated backreference label value to yield a 0-indexed offset.
	// Argo JS: labelToOffsetFactor = Number(LowestReservedValue) - 1
	// Since LowestReservedValue is -3, labelToOffsetFactor becomes -3 - 1 = -4.
	if !lowestReservedBigInt.IsInt64() {
		// This should not happen as -3 is well within int64 range.
		panic("label: internal error - lowestReservedBigInt cannot be represented as int64 during init")
	}
	labelToOffsetFactor = lowestReservedBigInt.Int64() - 1
}

// Kind determines and returns the semantic LabelKind of the Label (e.g., Null, Length, Backreference).
// The kind is derived from the label's numerical value according to Argo rules.
func (l Label) Kind() LabelKind {
	if l.value == nil { // Should ideally not occur due to constructors ensuring non-nil.
		// Consider this an internal error or undefined state.
		// Returning LabelKindError or panicking might be alternatives.
		// For now, assuming it implies an error state if reached.
		return LabelKindError
	}
	if l.value.Sign() >= 0 { // Non-negative values (0, 1, lengths) are LabelKindLength.
		return LabelKindLength
	}

	// Negative values: compare with specific marker *big.Int values.
	if l.value.Cmp(nullMarkerBigInt) == 0 { // -1
		return LabelKindNull
	}
	if l.value.Cmp(absentMarkerBigInt) == 0 { // -2
		return LabelKindAbsent
	}
	if l.value.Cmp(errorMarkerBigInt) == 0 { // -3
		return LabelKindError
	}
	// Any other negative value (i.e., value < -3) is a backreference.
	return LabelKindBackreference
}

// IsNull checks if the label represents a Null value (i.e., its value is -1).
func (l Label) IsNull() bool {
	if l.value == nil {
		return false // An uninitialized label is not NullMarker.
	}
	return l.value.Cmp(nullMarkerBigInt) == 0
}

// IsAbsent checks if the label represents an Absent value (i.e., its value is -2).
func (l Label) IsAbsent() bool {
	if l.value == nil {
		return false // An uninitialized label is not AbsentMarker.
	}
	return l.value.Cmp(absentMarkerBigInt) == 0
}

// IsError checks if the label represents an Error marker (i.e., its value is -3).
func (l Label) IsError() bool {
	if l.value == nil {
		return false // An uninitialized label is not ErrorMarker.
	}
	return l.value.Cmp(errorMarkerBigInt) == 0
}

// IsLength checks if the label represents a length or a non-negative marker (i.e., its value >= 0).
// This includes actual lengths, TrueMarker (1), FalseMarker (0), and NonNullMarker (0).
func (l Label) IsLength() bool {
	if l.value == nil {
		return false // An uninitialized label is not a length.
	}
	return l.value.Sign() >= 0
}

// IsBackref checks if the label represents a backreference.
// A backreference label has a value numerically less than `lowestReservedBigInt` (i.e., < -3).
func (l Label) IsBackref() bool {
	if l.value == nil {
		return false // An uninitialized label is not a backreference.
	}
	// A label is a backreference if its value is negative and not one of the special
	// negative markers (Null, Absent, Error). This is equivalent to being < lowestReservedBigInt.
	return l.value.Cmp(lowestReservedBigInt) < 0
}

// Encode converts the Label into its ZigZag ULEB128 encoded byte representation.
// For common marker labels (Null, Absent, Error), it returns their pre-encoded global byte slices.
// Otherwise, it performs ZigZag encoding on the label's *big.Int value.
func (l Label) Encode() []byte {
	if l.value == nil {
		// This handles the case of an uninitialized Label struct (e.g. var lbl Label).
		// Defaulting to encoding Null seems a reasonable, albeit potentially masking, behavior.
		// A stricter approach might panic or return an error.
		return Null
	}

	// Use the label's Kind to efficiently select pre-encoded values for markers.
	switch l.Kind() {
	case LabelKindNull:
		return Null // Use pre-encoded global constant.
	case LabelKindAbsent:
		return Absent // Use pre-encoded global constant.
	case LabelKindError:
		return Error // Use pre-encoded global constant.
	case LabelKindLength, LabelKindBackreference:
		// This path handles positive numbers (lengths), zero (False/NonNull/Length 0),
		// and true backreferences (negative values numerically smaller than errorMarkerBigInt).
		// If l.value was, for example, exactly nullMarkerBigInt, l.Kind() would be LabelKindNull,
		// correctly taking that path to use the pre-encoded `Null` slice.
		return varint.ZigZagEncode(l.value)
	default:
		// This case should ideally not be reached if Kind() is comprehensive and l.value is valid.
		// If an unknown or unexpected Kind were somehow produced, panicking might be appropriate
		// as it indicates an internal logic inconsistency.
		// For robustness in unexpected scenarios, falling back to direct encoding could be done,
		// but it implies an issue that should be investigated.
		// Assuming Kind() correctly categorizes all valid label values, this default is defensive.
		// Consider logging an internal error if this path is ever taken in a production system.
		// As a fallback, encode the raw value. This might occur if l.value was somehow
		// set to nil *after* construction and Kind() returned an unexpected default.
		fmt.Printf("label.Encode: warning - unexpected LabelKind, falling back to direct ZigZagEncode for value: %s\n", l.String())
		return varint.ZigZagEncode(l.value)
	}
}

// Read decodes a label from a buffer `b` that implements the `buf.Read` interface.
// It reads a ZigZag ULEB128 encoded number from the buffer at its current position,
// advances the buffer's position by the number of bytes read, and returns the decoded Label.
// Any errors encountered during varint decoding are propagated.
func Read(b buf.Read) (Label, error) {
	bufferBytes := b.Bytes()
	currentPosition := b.Position()

	if currentPosition < 0 {
		return Label{value: big.NewInt(0)}, errors.New("label: buffer position is negative, cannot read")
	}

	// Ensure currentPosition is int for slice indexing, though ZigZagDecode expects int offset.
	// This cast is safe if b.Position() is within typical buffer size ranges.
	// varint.ZigZagDecode will handle bounds checking against len(bufferBytes).
	decodedBigInt, numBytesRead, err := varint.ZigZagDecode(bufferBytes, int(currentPosition))
	if err != nil {
		// Return a zero-value Label on error, along with the error itself.
		return Label{value: big.NewInt(0)}, fmt.Errorf("label.Read: failed to decode varint: %w", err)
	}

	b.IncrementPosition(int64(numBytesRead))
	return New(decodedBigInt), nil
}

// ToOffset converts a backreference Label into a 0-indexed array/slice offset.
// The Label must represent a valid backreference (i.e., its Kind must be LabelKindBackreference).
// An error is returned if the Label is not a valid backreference (e.g., it's non-negative
// or one of the special markers like Null, Absent, Error), or if its value is too large
// to be represented as an int64 after transformation (highly unlikely for practical offsets).
// The formula used is: offset = -NumericValue(label) + labelToOffsetFactor.
// For example, if labelToOffsetFactor is -4:
//   - A label of -4 (smallest valid backref value) gives offset: -(-4) + (-4) = 4 - 4 = 0.
//   - A label of -5 gives offset: -(-5) + (-4) = 5 - 4 = 1.
func (l Label) ToOffset() (int64, error) {
	if l.value == nil {
		return 0, errors.New("label: ToOffset called on uninitialized Label")
	}
	// A label must be a true backreference to be converted to an offset.
	// This means its value must be < lowestReservedBigInt (-3).
	if l.Kind() != LabelKindBackreference {
		return 0, fmt.Errorf("label: cannot convert label of kind '%s' (value %s) to offset; must be a backreference", l.Kind(), l.String())
	}

	// Perform the calculation: offset = -labelValue + labelToOffsetFactor
	negatedLabelVal := new(big.Int).Neg(l.value)

	if !negatedLabelVal.IsInt64() {
		// This is extremely unlikely for a backreference offset but check for safety.
		return 0, fmt.Errorf("label: negated backreference value %s too large for int64 conversion in ToOffset", negatedLabelVal.String())
	}

	offset := negatedLabelVal.Int64() + labelToOffsetFactor
	if offset < 0 {
		// This would indicate an issue with labelToOffsetFactor or the input label value,
		// as valid backreferences should produce non-negative offsets.
		return 0, fmt.Errorf("label: calculated offset %d is negative for label %s; internal logic error or invalid backreference", offset, l.String())
	}
	return offset, nil
}

// Is checks if this label (l) has the same numerical value as another label (other).
// It compares their underlying *big.Int values.
// Handles cases where one or both labels might be uninitialized (value is nil),
// although constructors aim to prevent l.value from being nil.
func (l Label) Is(other Label) bool {
	val1 := l.Value() // Use Value() to respect potential nil internal value.
	val2 := other.Value()

	if val1 == nil && val2 == nil {
		// Both are uninitialized (e.g. zero value Label struct not passed through New*).
		// Or if New(nil) resulted in Label{value:nil} (though current New(nil) makes it 0).
		return true
	}
	if val1 == nil || val2 == nil {
		// One is initialized, the other is not (or became nil post-construction).
		return false
	}
	// Both have non-nil *big.Int values, compare them.
	return val1.Cmp(val2) == 0
}
