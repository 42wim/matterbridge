package codec

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"os"
	"reflect"
	"sort"

	"github.com/beeper/argo-go/block"
	"github.com/beeper/argo-go/header"
	"github.com/beeper/argo-go/internal/util"
	"github.com/beeper/argo-go/label"
	"github.com/beeper/argo-go/pkg/buf"
	"github.com/beeper/argo-go/pkg/varint"
	"github.com/beeper/argo-go/wire"
	"github.com/elliotchance/orderedmap/v3"
	"github.com/vektah/gqlparser/v2/ast"
)

// writerEntry stores an association between a block.AnyBlockWriter and the
// original wire.Type of the values it's intended to write. This is important for
// type-specific operations, such as determining if null termination is needed for string blocks.
type writerEntry struct {
	Writer            block.AnyBlockWriter // The type-erased block writer instance.
	OriginalValueType wire.Type            // The original wire type this writer was created for (e.g., wire.String, wire.Bytes).
}

// ArgoEncoder handles the conversion of Go data structures, typically representing
// GraphQL query results, into the Argo binary message format. It manages a core
// buffer for labels and inline data, a collection of block writers for scalar
// values that are not inlined, and an Argo header.
// The encoder supports various Argo features like deduplication, different block types,
// and self-describing values, controlled by header flags and wire type definitions.
type ArgoEncoder struct {
	coreBuf *buf.Buf // Buffer for the core message data, primarily labels and inlined scalar values.
	// writers stores block writers, keyed by their wire.BlockKey. Each entry contains the
	// AnyBlockWriter and the original wire.Type for which it was created.
	writers *orderedmap.OrderedMap[wire.BlockKey, writerEntry]
	header  *header.Header // The Argo header for the message being encoded.

	// Debug fields, used when ArgoEncoder.Debug is true.
	Debug bool // If true, enables tracking of encoding steps.
	// tracked stores a log of encoding operations when Debug is true.
	// Each entry is an ordered map representing a single tracked step.
	tracked []*orderedmap.OrderedMap[string, interface{}]
}

// NewArgoEncoder initializes and returns a new ArgoEncoder.
// It sets up the core buffer, the map for block writers, and a new Argo header.
func NewArgoEncoder() *ArgoEncoder {
	// Initial capacity for coreBuf; it will grow as needed.
	// The final message buffer is constructed in GetResult by combining the header,
	// block data (if not inlined), and the core buffer contents.
	coreBuffer := buf.NewBuf(1024) // Default initial capacity.
	hdr := header.NewHeader()

	return &ArgoEncoder{
		coreBuf: coreBuffer,
		writers: orderedmap.NewOrderedMap[wire.BlockKey, writerEntry](),
		header:  hdr,
		tracked: []*orderedmap.OrderedMap[string, interface{}]{}, // Initialize an empty slice for tracking.
	}
}

// Header returns the encoder's *header.Header instance, allowing the caller
// to set Argo header flags or other header properties before finalizing the message.
func (ae *ArgoEncoder) Header() *header.Header {
	return ae.header
}

// Track records an encoding step for debugging purposes if ae.Debug is true.
// It captures the GraphQL path, a descriptive message, the current buffer (if any),
// and the value being processed.
func (ae *ArgoEncoder) Track(path ast.Path, msg string, b buf.Write, value interface{}) {
	if ae.Debug {
		entry := orderedmap.NewOrderedMap[string, interface{}]()
		entry.Set("path", util.FormatPath(path))
		entry.Set("msg", msg)
		if b != nil { // Buffer might be nil for some tracking events (e.g., header bytes)
			entry.Set("pos", b.Position())
		} else {
			entry.Set("pos", -1) // Indicate no buffer position
		}

		// Avoid deep copying complex values or handle them carefully
		if s, ok := value.(string); ok && len(s) > 100 {
			entry.Set("value", s[:100]+"...")
		} else if b, ok := value.([]byte); ok && len(b) > 100 {
			entry.Set("value", fmt.Sprintf("bytes[%d]", len(b)))
		} else {
			entry.Set("value", value)
		}
		ae.tracked = append(ae.tracked, entry)
	}
}

// Log provides a more generic logging mechanism for debugging, used when ae.Debug is true.
// It records the current position in the core buffer and a message or detailed object.
func (ae *ArgoEncoder) Log(msg interface{}) {
	if ae.Debug {
		entry := orderedmap.NewOrderedMap[string, interface{}]()
		entry.Set("pos", ae.coreBuf.Position())

		if s, ok := msg.(string); ok {
			entry.Set("msg", s)
		} else if om, ok := msg.(*orderedmap.OrderedMap[string, interface{}]); ok {
			// If msg is an OrderedMap, merge its fields.
			for el := om.Front(); el != nil; el = el.Next() {
				entry.Set(el.Key, el.Value)
			}
		} else {
			entry.Set("detail", msg)
		}
		ae.tracked = append(ae.tracked, entry)
	}
}

// nullTerminator is a pre-allocated byte slice for null-terminating strings.
var nullTerminator = []byte{0x00}

// makeBlockWriter creates and returns a block.AnyBlockWriter suitable for the given wire.Type `t`
// and deduplication setting. It handles various scalar types by configuring appropriate
// ValueToBytesFunc and MakeLabelFunc for the underlying block.BlockWriter or
// block.DeduplicatingBlockWriter.
// For BytesType with deduplication, it uses a specialized bytesDeduplicatingAdapter.
func (ae *ArgoEncoder) makeBlockWriter(t wire.Type, dedupe bool) (block.AnyBlockWriter, error) {
	switch t.(type) {
	case wire.StringType:
		stringVTB := func(s string) ([]byte, error) { // ValueToBytesFunc for string
			return []byte(s), nil
		}
		if dedupe {
			// For strings, deduplication uses the string itself as the key.
			dbw := block.NewLengthOfBytesDeduplicatingBlockWriter[string](stringVTB)
			return block.NewAnyDeduplicatingBlockWriter(dbw), nil
		}
		// Non-deduplicating string writer also uses length-based labels.
		bw := block.NewLengthOfBytesBlockWriter[string](stringVTB)
		return block.NewAnyBlockWriter(bw), nil

	case wire.BytesType:
		bytesVTB := func(b []byte) ([]byte, error) { // ValueToBytesFunc for []byte
			return b, nil
		}
		if dedupe {
			// For BytesType with deduplication, the underlying DeduplicatingBlockWriter[string]
			// uses string(originalBytes) as the deduplication key.
			// The bytesDeduplicatingAdapter handles the conversion from []byte input (in its Write method)
			// to the string key expected by the core writer.
			dedupeKeyedVTB := func(sKey string) ([]byte, error) { // ValueToBytes for the string-keyed deduplicator
				return []byte(sKey), nil
			}
			dbw := block.NewLengthOfBytesDeduplicatingBlockWriter[string](dedupeKeyedVTB)
			return &bytesDeduplicatingAdapter{dbw}, nil // Specialized adapter for []byte with string-keyed dedupe.
		}
		// Non-deduplicating bytes writer.
		bw := block.NewLengthOfBytesBlockWriter[[]byte](bytesVTB)
		return block.NewAnyBlockWriter(bw), nil

	case wire.VarintType:
		if dedupe { // Deduplication for Varint is not standard/implemented.
			return nil, fmt.Errorf("unimplemented: deduping VARINT")
		}
		// Varint values are written without length labels by default (label is nil from NewAnyNoLabelBlockWriter).
		// The actual Varint encoding happens in this ValueToBytesFunc.
		varintVTB := func(v interface{}) ([]byte, error) {
			switch val := v.(type) {
			case int:
				return varint.ZigZagEncodeInt64(int64(val)), nil
			case int64:
				return varint.ZigZagEncodeInt64(val), nil
			case *big.Int:
				return varint.ZigZagEncode(val), nil
			case float64:
				// Check if float64 is a whole number
				if val == math.Trunc(val) {
					return varint.ZigZagEncodeInt64(int64(val)), nil
				}
				return nil, fmt.Errorf("float64 value %v is not a whole number for VarintType block", val)
			default:
				return nil, fmt.Errorf("expected int, int64, *big.Int or whole float64 for VarintType block, got %T for value %v", v, v)
			}
		}
		// Varints are not labeled with their length; their encoding is self-terminating.
		return block.NewAnyNoLabelBlockWriter(varintVTB), nil

	case wire.Float64Type:
		if dedupe { // Deduplication for Float64 is not standard/implemented.
			return nil, fmt.Errorf("unimplemented: deduping FLOAT64")
		}
		// Float64 values are written without length labels by default.
		floatVTB := func(v interface{}) ([]byte, error) {
			var f float64
			switch val := v.(type) {
			case float64:
				f = val
			case float32:
				f = float64(val)
			case int:
				f = float64(val)
			case int64:
				f = float64(val)
			default:
				return nil, fmt.Errorf("expected float64 compatible type for Float64Type block, got %T for value %v", v, v)
			}
			var b [8]byte // Float64 is 8 bytes.
			binary.LittleEndian.PutUint64(b[:], math.Float64bits(f))
			return b[:], nil
		}
		// Floats are fixed-width, so no length label is needed.
		return block.NewAnyNoLabelBlockWriter(floatVTB), nil

	case wire.FixedType:
		fixedType := t.(wire.FixedType)
		if dedupe { // Deduplication for FixedType is not standard/implemented.
			return nil, fmt.Errorf("unimplemented: deduping FIXED")
		}
		fixedVTB := func(v interface{}) ([]byte, error) {
			b, ok := v.([]byte)
			if !ok {
				return nil, fmt.Errorf("expected []byte for FixedType block, got %T for value %v", v, v)
			}
			if len(b) != fixedType.Length {
				return nil, fmt.Errorf("fixedType expected %d bytes, got %d for value %v", fixedType.Length, len(b), b)
			}
			return b, nil
		}
		// Fixed-length types do not need length labels.
		return block.NewAnyNoLabelBlockWriter(fixedVTB), nil

	default:
		return nil, fmt.Errorf("unsupported block writer type %s (underlying Go type: %T)", wire.Print(t), t)
	}
}

// bytesDeduplicatingAdapter is a specialized AnyBlockWriter adapter for BytesType when deduplication is enabled.
// It wraps a DeduplicatingBlockWriter[string] and handles the conversion of input []byte values
// to string keys for the underlying deduplicator.
type bytesDeduplicatingAdapter struct {
	coreWriter *block.DeduplicatingBlockWriter[string] // Underlying writer uses string keys for deduplication.
}

// Write converts the input value `v` (expected to be []byte or string) to a string key
// and passes it to the underlying string-keyed DeduplicatingBlockWriter.
func (a *bytesDeduplicatingAdapter) Write(v interface{}) (*label.Label, error) {
	bytesVal, ok := v.([]byte)
	if !ok {
		// Allow string input as well, interpreting it as bytes.
		if strVal, okStr := v.(string); okStr {
			bytesVal = []byte(strVal)
		} else {
			return nil, fmt.Errorf("bytesDeduplicatingAdapter expected []byte or string, got %T for value %v", v, v)
		}
	}
	// The coreWriter expects a string for deduplication; string(bytesVal) serves as the key.
	return a.coreWriter.Write(string(bytesVal))
}

// AllValuesAsBytes delegates to the underlying coreWriter.
func (a *bytesDeduplicatingAdapter) AllValuesAsBytes() [][]byte {
	return a.coreWriter.AllValuesAsBytes()
}

// WriteLastToBuf delegates to the underlying coreWriter.
func (a *bytesDeduplicatingAdapter) WriteLastToBuf(buf buf.Write) error {
	return a.coreWriter.WriteLastToBuf(buf)
}

// getWriter retrieves an existing block.AnyBlockWriter for the given blockDef.Key, or creates
// a new one if it doesn't exist. Created writers are stored in ae.writers for reuse.
// `valueWireType` is typically the `Of` type of the `blockDef` (e.g., wire.String for a block of strings).
func (ae *ArgoEncoder) getWriter(blockDef wire.BlockType, valueWireType wire.Type) (block.AnyBlockWriter, error) {
	if entry, found := ae.writers.Get(blockDef.Key); found {
		return entry.Writer, nil
	}
	// Create a new writer if one doesn't exist for this block key.
	writer, err := ae.makeBlockWriter(valueWireType, blockDef.Dedupe)
	if err != nil {
		return nil, fmt.Errorf("failed to make block writer for key '%s' (value type %s): %w", blockDef.Key, wire.Print(valueWireType), err)
	}
	ae.writers.Set(blockDef.Key, writerEntry{Writer: writer, OriginalValueType: valueWireType})
	return writer, nil
}

// Write is a core method for handling scalar values that are part of a block.
// It retrieves or creates the appropriate block writer for `blockDef`,
// writes the value `v` to this writer (which generates a label and stores bytes),
// and then writes the label (if any) to the encoder's coreBuf.
// If the `HeaderInlineEverythingFlag` is set, it also writes the value's bytes directly
// to the coreBuf for certain types of labels (e.g., length labels, non-null markers for unlabeled types).
func (ae *ArgoEncoder) Write(blockDef wire.BlockType, valueWireType wire.Type, v interface{}) (*label.Label, error) {
	writer, err := ae.getWriter(blockDef, valueWireType)
	if err != nil {
		return nil, err // Error from getWriter already has context.
	}

	lbl, err := writer.Write(v) // Value `v` is written to the chosen block writer.
	if err != nil {
		return nil, fmt.Errorf("block writer for key '%s' (value type %s) failed: %w", blockDef.Key, wire.Print(valueWireType), err)
	}

	// Part 1: Write the label to the core buffer.
	// This happens regardless of InlineEverything, as labels are part of the core stream.
	if lbl != nil {
		encodedLabel := lbl.Encode()
		if _, err := ae.coreBuf.Write(encodedLabel); err != nil {
			// Error context for label writing.
			mode := "non-inline"
			if ae.header.GetFlag(header.HeaderInlineEverythingFlag) {
				mode = "inline"
			}
			return nil, fmt.Errorf("failed to write label to core buffer (mode: %s): %w", mode, err)
		}
	}

	// Part 2: If InlineEverything is active, write the data bytes to coreBuf if appropriate.
	if ae.header.GetFlag(header.HeaderInlineEverythingFlag) {
		shouldWriteDataInline := false
		if lbl == nil {
			// For unlabeled types (e.g., FLOAT64, FIXED non-nullable), their data is always written inline.
			shouldWriteDataInline = true
		} else {
			// For labeled types, data is written inline only if the label isn't a "standalone" one
			// (i.e., if it implies data should follow, like a length or NonNull marker).
			// Backreferences, Null, Absent, Error labels do not have following data in the core stream.
			isStandaloneLabel := lbl.IsBackref() || lbl.IsNull() || lbl.IsAbsent() || lbl.IsError()
			if !isStandaloneLabel {
				shouldWriteDataInline = true
			}
		}

		if shouldWriteDataInline {
			// WriteLastToBuf gets the most recent value from the block writer and writes it to coreBuf.
			if err := writer.WriteLastToBuf(ae.coreBuf); err != nil {
				return nil, fmt.Errorf("failed to write value data to core buffer in inline mode: %w", err)
			}
		}
	}
	// If not InlineEverything, value bytes remain in their respective block writers (writer.valuesAsBytes)
	// and are assembled into the final message by GetResult().

	return lbl, nil
}

// ValueToArgoWithType is the primary entry point for encoding a Go data structure (typically from JSON-like input)
// into the Argo format based on a provided wire.Type schema.
// The `v` interface{} is expected to conform to the structure defined by `wt`.
// For example, if `wt` is a RecordType, `v` should be an *orderedmap.OrderedMap[string, interface{}].
// If `wt` is an ArrayType, `v` should be a slice or array.
// Debugging information, if enabled, is written to "tmp-gowritelog.json".
func (ae *ArgoEncoder) ValueToArgoWithType(v interface{}, wt wire.Type) error {
	// Start recursive encoding. currentPath is initially nil, currentBlock is initially nil.
	err := ae.writeArgo(nil, v, wt, nil)

	// If debugging is enabled, write the tracked encoding steps to a JSON file.
	if ae.Debug {
		jsony := make([]*util.OrderedMapJSON[string, any], len(ae.tracked))
		for i, obj := range ae.tracked {
			jsony[i] = util.NewOrderedMapJSON(obj)
		}
		trackedJSON, jsonErr := json.MarshalIndent(jsony, "", "  ")
		if jsonErr != nil {
			// Log marshalling error, but don't let it hide the main encoding error.
			fmt.Fprintf(os.Stderr, "Error marshalling debug tracking data: %v\n", jsonErr)
		} else {
			_ = os.WriteFile("tmp-gowritelog.json", trackedJSON, 0644) // Error is ignored for debug artifact.
		}
	}
	return err
}

// writeArgo is the main recursive workhorse for encoding. It traverses the input data `v`
// according to the structure defined by the wire.Type `wt`.
// `currentPath` tracks the path within the data structure for debugging.
// `currentBlock` points to the wire.BlockType definition if the current context is writing
// elements into a specific block (e.g., a block of strings or varints).
func (ae *ArgoEncoder) writeArgo(currentPath ast.Path, v interface{}, wt wire.Type, currentBlock *wire.BlockType) error {
	ae.Track(currentPath, "writeArgo type: "+string(wt.GetTypeKey()), ae.coreBuf, v)

	switch typedWt := wt.(type) {
	case wire.NullableType:
		if v == nil { // Handle explicit Go nil for a nullable type.
			ae.Track(currentPath, "null", ae.coreBuf, label.Null)
			_, err := ae.coreBuf.Write(label.Null)
			return err
		}

		// Handle inline errors: if `v` is an error or []error, write Argo error representation.
		// This needs to check various ways an error or slice of errors might be represented in `v`.
		var errorArray []error
		if errVal, ok := v.(error); ok {
			errorArray = []error{errVal} // Single error.
		} else if errSlice, ok := v.([]error); ok {
			errorArray = errSlice // Already a slice of errors.
		} else if interfaceSlice, ok := v.([]interface{}); ok {
			// Check if []interface{} contains only errors.
			canBeErrorArray := true
			for _, item := range interfaceSlice {
				if errVal, itemIsErr := item.(error); itemIsErr {
					errorArray = append(errorArray, errVal)
				} else {
					canBeErrorArray = false // Found a non-error item.
					break
				}
			}
			if !canBeErrorArray {
				errorArray = nil // Not a pure error array.
			}
		}

		if len(errorArray) > 0 {
			// Value is an error or slice of errors, write it in Argo error format.
			ae.Track(currentPath, "error value encountered", ae.coreBuf, v)
			if _, err := ae.coreBuf.Write(label.Error); err != nil { // Write the Error marker label.
				return err
			}
			// Write the number of errors as a length label.
			lenLabel := label.NewFromInt64(int64(len(errorArray)))
			if _, err := ae.coreBuf.Write(lenLabel.Encode()); err != nil {
				return err
			}
			// Write each error.
			for i, e := range errorArray {
				errPath := util.AddPathIndex(currentPath, i) // Path for this specific error in the array.
				if ae.header.GetFlag(header.HeaderSelfDescribingErrorsFlag) {
					// Use self-describing format for errors.
					if err := ae.writeSelfDescribing(errPath, e); err != nil {
						return fmt.Errorf("failed to write self-describing error item at index %d: %w", i, err)
					}
				} else {
					// Use structured Argo error format (defined by wire.Error type).
					if err := ae.writeGoError(errPath, e); err != nil {
						return fmt.Errorf("failed to write structured error item at index %d: %w", i, err)
					}
				}
			}
			return nil // Successfully wrote error(s).
		}

		// If not nil and not an error, it's a non-null instance of the underlying type.
		// If the underlying type is not intrinsically labeled (e.g. scalars in blocks often are),
		// a NonNull marker is needed here before writing the actual value.
		if !wire.IsLabeled(typedWt.Of) {
			ae.Track(currentPath, "non-null marker for nullable type", ae.coreBuf, label.NonNull)
			if _, err := ae.coreBuf.Write(label.NonNull); err != nil {
				return err
			}
		}
		// Continue writing with the underlying type.
		return ae.writeArgo(currentPath, v, typedWt.Of, currentBlock)

	case wire.BlockType:
		if currentBlock != nil {
			// This should not happen if wire types are structured correctly (no nested blocks).
			return fmt.Errorf("encoder error: already processing block '%s', cannot switch to new block '%s' at path %s. Wire type: %s",
				currentBlock.Key, typedWt.Key, util.FormatPath(currentPath), wire.Print(wt))
		}
		ae.Track(currentPath, "entering block with key", ae.coreBuf, typedWt.Key)
		// Recursively call writeArgo with the block's element type (`typedWt.Of`)
		// and pass `&typedWt` as the new `currentBlock` context.
		return ae.writeArgo(currentPath, v, typedWt.Of, &typedWt)

	case wire.RecordType:
		ae.Track(currentPath, "record with number of fields", ae.coreBuf, len(typedWt.Fields))
		// Expect v to be an *orderedmap.OrderedMap for records to maintain field order.
		om, ok := v.(*orderedmap.OrderedMap[string, interface{}])
		if !ok && v != nil { // If v is not nil, it must be the correct map type.
			return fmt.Errorf("type error: expected *orderedmap.OrderedMap[string, interface{}] for record, got %T at path %s", v, util.FormatPath(currentPath))
		}

		// Iterate through fields as defined in the wire.RecordType to ensure correct order and handling of all defined fields.
		for _, field := range typedWt.Fields {
			fieldPath := util.AddPathName(currentPath, field.Name)
			var fieldValue interface{}
			var fieldExists bool
			if om != nil { // If input map is nil (because parent was nil), all fields are treated as absent.
				fieldValue, fieldExists = om.Get(field.Name)
			} else {
				fieldValue = nil // Effectively absent.
				fieldExists = false
			}

			if fieldExists && fieldValue != nil && fieldValue != wire.AbsentValue {
				// Field is present and has a non-nil, non-absent value.
				// If omittable and its underlying type isn't self-labeling (like a BlockType often is),
				// write a NonNull marker to indicate its presence.
				if field.Omittable && !wire.IsLabeled(field.Of) {
					ae.Track(fieldPath, "omittable record field present, writing NonNull marker", ae.coreBuf, field.Name)
					if _, err := ae.coreBuf.Write(label.NonNull); err != nil {
						return err
					}
				}
				// Recursively write the field's value.
				if err := ae.writeArgo(fieldPath, fieldValue, field.Of, currentBlock); err != nil {
					return err
				}
			} else if field.Omittable && (!fieldExists || fieldValue == wire.AbsentValue) {
				// Field is omittable and is effectively absent (either not in map or explicit AbsentValue).
				ae.Track(fieldPath, "omittable record field absent, writing Absent marker", ae.coreBuf, field.Name)
				if _, err := ae.coreBuf.Write(label.Absent); err != nil {
					return err
				}
			} else if wire.IsNullable(field.Of) {
				// Field is not omittable (or omittable but present as nil) AND its type is nullable.
				// This handles cases where an explicit JSON null was provided for a nullable field,
				// or a non-omittable field is nil (which is only valid if its type is nullable).
				ae.Track(fieldPath, "record field is nil and type is nullable (or non-omittable field is nil), recursing", ae.coreBuf, field.Name)
				// Recursively call writeArgo. If fieldValue is nil, this will correctly write a Null label via the NullableType case.
				if err := ae.writeArgo(fieldPath, fieldValue, field.Of, currentBlock); err != nil {
					return err
				}
			} else if wire.IsBlock(field.Of) && wire.IsDesc(field.Of.(wire.BlockType).Of) {
				// Special case: field is a Block of SelfDescribing (DESC) type and is absent/nil.
				// SelfDescribing types can represent null, so we recurse to let DESC handle the nil.
				ae.Track(fieldPath, "record field is nil/absent but is Block<DESC>, recursing for self-describing null", ae.coreBuf, field.Name)
				if err := ae.writeArgo(fieldPath, nil, field.Of, currentBlock); err != nil {
					return err
				}
			} else {
				// Field is absent/nil, but it's not omittable, not nullable, and not Block<DESC>.
				// This is a data-schema mismatch.
				return fmt.Errorf("schema error: record field '%s' is absent/nil but its type (%s) is not omittable, not nullable, and not Block<DESC> at path %s", field.Name, wire.Print(field.Of), util.FormatPath(fieldPath))
			}
		}
		return nil

	case wire.ArrayType:
		reflectVal := reflect.ValueOf(v)
		if reflectVal.Kind() != reflect.Slice && reflectVal.Kind() != reflect.Array {
			// If `v` is nil for an ArrayType, it's an error because ArrayType itself is not nullable.
			// A nullable array would be NullableType{Of: ArrayType{...}}.
			if v == nil {
				return fmt.Errorf("type error: cannot encode Go nil as non-nullable Argo array at path %s. WireType: %s", util.FormatPath(currentPath), wire.Print(wt))
			}
			return fmt.Errorf("type error: expected slice or array for ArrayType, got %T at path %s", v, util.FormatPath(currentPath))
		}
		length := reflectVal.Len()
		ae.Track(currentPath, "array with length", ae.coreBuf, length)
		lenLabel := label.NewFromInt64(int64(length)) // Label for array length.
		if _, err := ae.coreBuf.Write(lenLabel.Encode()); err != nil {
			return err
		}
		// Recursively write each element of the array.
		for i := 0; i < length; i++ {
			itemPath := util.AddPathIndex(currentPath, i)
			itemValue := reflectVal.Index(i).Interface()
			if err := ae.writeArgo(itemPath, itemValue, typedWt.Of, currentBlock); err != nil {
				return err
			}
		}
		return nil

	case wire.BooleanType:
		bVal, ok := v.(bool)
		if !ok {
			return fmt.Errorf("expected bool for BooleanType, got %T at path %s", v, util.FormatPath(currentPath))
		}
		ae.Track(currentPath, "boolean", ae.coreBuf, bVal)
		if bVal {
			_, err := ae.coreBuf.Write(label.True)
			return err
		}
		_, err := ae.coreBuf.Write(label.False)
		return err

	case wire.StringType, wire.BytesType, wire.VarintType, wire.Float64Type, wire.FixedType:
		// Logic for types which may use blocks for data
		if currentBlock == nil {
			return fmt.Errorf("programmer error: need block for %s at path %s", wire.Print(wt), util.FormatPath(currentPath))
		}
		_, err := ae.Write(*currentBlock, wt, v)
		if err != nil {
			return err
		}
		ae.Track(currentPath, string(wt.GetTypeKey()), ae.coreBuf, v)
		return nil

	case wire.DescType:
		ae.Track(currentPath, "self-describing", ae.coreBuf, v)
		return ae.writeSelfDescribing(currentPath, v)

	case wire.PathType: // Argo spec: PATH values ... encoded exactly as an ARRAY of VARINT values.
		// This should be handled by the Error type definition, which includes a PATH field.
		// If we encounter a raw PathType, treat it as Array{Of: Varint}.
		pathSlice, ok := v.([]interface{})
		if !ok {
			if intSlice, isIntSlice := v.([]int); isIntSlice {
				pathSlice = make([]interface{}, len(intSlice))
				for i, v := range intSlice {
					pathSlice[i] = v
				}
				ok = true
			} else if int64Slice, isInt64Slice := v.([]int64); isInt64Slice {
				pathSlice = make([]interface{}, len(int64Slice))
				for i, v := range int64Slice {
					pathSlice[i] = v
				}
				ok = true
			} else {
				return fmt.Errorf("expected []interface{} or []int or []int64 for PathType, got %T at path %s", v, util.FormatPath(currentPath))
			}
		}
		// The actual transformation from GraphQL path to wire path (list of integers) happens
		// before this point if we are encoding a structured Error.
		// Here, we assume 'v' is already the list of integers for the wire.
		return ae.writeArgo(currentPath, pathSlice, wire.ArrayType{Of: wire.Varint}, currentBlock)

	default:
		return fmt.Errorf("unsupported wire type %T (%s) for encoding at path %s", wt, wire.Print(wt), util.FormatPath(currentPath))
	}
}

// writeSelfDescribing writes a Go value in Argo's self-describing format.
// This format uses specific leading bytes to indicate the type of the following data.
// It's used for errors when HeaderSelfDescribingErrorsFlag is set, or for fields of type wire.Desc.
func (ae *ArgoEncoder) writeSelfDescribing(currentPath ast.Path, v interface{}) error {
	ae.Track(currentPath, "writeSelfDescribing value", ae.coreBuf, v)
	if v == nil {
		_, err := ae.coreBuf.Write(wire.SelfDescribingNull) // Write the null marker.
		return err
	}

	// Optimized path for *orderedmap.OrderedMap (common for objects).
	if om, ok := v.(*orderedmap.OrderedMap[string, interface{}]); ok {
		if _, err := ae.coreBuf.Write(wire.SelfDescribingObject); err != nil { // Object marker.
			return err
		}
		numFields := om.Len()
		lenLabel := label.NewFromInt64(int64(numFields)) // Length of object (number of fields).
		if _, err := ae.coreBuf.Write(lenLabel.Encode()); err != nil {
			return err
		}

		// Iterate through fields of the ordered map.
		for el := om.Front(); el != nil; el = el.Next() {
			k := el.Key   // Field name
			v := el.Value // Field value
			fieldPath := util.AddPathName(currentPath, k)

			// Write field name (as a self-describing string).
			stringBlockKey := wire.BlockKey("String") // Standard key for self-describing strings.
			stringElementType, blockDefOk := wire.SelfDescribingBlocks[stringBlockKey]
			if !blockDefOk {
				return fmt.Errorf("internal error: self-describing string block key ('%s') not found in wire.SelfDescribingBlocks map for field name '%s'", stringBlockKey, k)
			}
			selfDescribingStringBlock := wire.NewBlockType(stringElementType, stringBlockKey, wire.MustDeduplicateByDefault(stringElementType))
			if _, err := ae.Write(selfDescribingStringBlock, wire.String, k); err != nil {
				return fmt.Errorf("failed to write self-describing object field name '%s': %w", k, err)
			}

			// Recursively write field value in self-describing format.
			if err := ae.writeSelfDescribing(fieldPath, v); err != nil {
				return fmt.Errorf("failed to write self-describing object field value for '%s': %w", k, err)
			}
		}
		return nil
	}

	// General path using reflection for other types.
	val := reflect.ValueOf(v)
	switch val.Kind() {
	case reflect.Map:
		// Handles native Go maps (e.g., map[string]interface{}).
		// For determinism, these are converted to *orderedmap.OrderedMap before recursive call.
		if val.Type().Key().Kind() != reflect.String {
			return fmt.Errorf("type error: cannot encode map with non-string keys in self-describing object at path %s (type: %T)", util.FormatPath(currentPath), v)
		}

		ae.Track(currentPath, "converting native map to OrderedMap for self-describing encoding", ae.coreBuf, v)
		tempOM := orderedmap.NewOrderedMap[string, interface{}]()
		var stringKeys []string
		for _, kVal := range val.MapKeys() {
			stringKeys = append(stringKeys, kVal.String())
		}
		sort.Strings(stringKeys) // Sort keys for deterministic order in the temporary OrderedMap.

		for _, sk := range stringKeys {
			mapValue := val.MapIndex(reflect.ValueOf(sk)).Interface()
			tempOM.Set(sk, mapValue)
		}
		return ae.writeSelfDescribing(currentPath, tempOM) // Recurse with the ordered map.

	case reflect.Slice, reflect.Array:
		// Handle []byte separately as SelfDescribingBytes.
		if byteSlice, isBytes := v.([]byte); isBytes {
			if _, err := ae.coreBuf.Write(wire.SelfDescribingBytes); err != nil { // Bytes marker.
				return err
			}
			bytesBlockKey := wire.BlockKey("Bytes") // Standard key for self-describing byte arrays.
			bytesElementType, blockDefOk := wire.SelfDescribingBlocks[bytesBlockKey]
			if !blockDefOk {
				return fmt.Errorf("internal error: self-describing bytes block key ('%s') not found in wire.SelfDescribingBlocks map", bytesBlockKey)
			}
			selfDescribingBytesBlock := wire.NewBlockType(bytesElementType, bytesBlockKey, wire.MustDeduplicateByDefault(bytesElementType))
			_, err := ae.Write(selfDescribingBytesBlock, wire.Bytes, byteSlice)
			return err
		}

		// Other slices/arrays are SelfDescribingList.
		if _, err := ae.coreBuf.Write(wire.SelfDescribingList); err != nil { // List marker.
			return err
		}
		length := val.Len()
		lenLabel := label.NewFromInt64(int64(length)) // Length of list.
		if _, err := ae.coreBuf.Write(lenLabel.Encode()); err != nil {
			return err
		}
		// Recursively write each list item in self-describing format.
		for i := 0; i < length; i++ {
			itemPath := util.AddPathIndex(currentPath, i)
			if err := ae.writeSelfDescribing(itemPath, val.Index(i).Interface()); err != nil {
				return fmt.Errorf("error writing self-describing list item at index %d (path %s): %w", i, util.FormatPath(itemPath), err)
			}
		}
		return nil

	case reflect.String:
		if _, err := ae.coreBuf.Write(wire.SelfDescribingString); err != nil { // String marker.
			return err
		}
		stringBlockKey := wire.BlockKey("String")
		stringElementType, blockDefOk := wire.SelfDescribingBlocks[stringBlockKey]
		if !blockDefOk {
			return fmt.Errorf("internal error: self-describing string block key ('%s') not found in wire.SelfDescribingBlocks map", stringBlockKey)
		}
		selfDescribingStringBlock := wire.NewBlockType(stringElementType, stringBlockKey, wire.MustDeduplicateByDefault(stringElementType))
		_, err := ae.Write(selfDescribingStringBlock, wire.String, v.(string))
		return err

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if _, err := ae.coreBuf.Write(wire.SelfDescribingInt); err != nil { // Integer marker.
			return err
		}
		varintBlockKey := wire.BlockKey("Int") // Standard key for self-describing integers.
		varintElementType, blockDefOk := wire.SelfDescribingBlocks[varintBlockKey]
		if !blockDefOk {
			return fmt.Errorf("internal error: self-describing varint block key ('%s') not found in wire.SelfDescribingBlocks map", varintBlockKey)
		}
		selfDescribingVarintBlock := wire.NewBlockType(varintElementType, varintBlockKey, wire.MustDeduplicateByDefault(varintElementType))
		_, err := ae.Write(selfDescribingVarintBlock, wire.Varint, val.Int()) // val.Int() converts various int types to int64 for varint encoder.
		return err

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		uVal := val.Uint()
		if _, err := ae.coreBuf.Write(wire.SelfDescribingInt); err != nil {
			return err
		}
		varintBlockKey := wire.BlockKey("Int")
		varintElementType, blockDefOk := wire.SelfDescribingBlocks[varintBlockKey]
		if !blockDefOk {
			return fmt.Errorf("internal error: self-describing varint block key ('%s') not found for uint value", varintBlockKey)
		}
		selfDescribingVarintBlock := wire.NewBlockType(varintElementType, varintBlockKey, wire.MustDeduplicateByDefault(varintElementType))
		if uVal <= math.MaxInt64 { // If fits in int64, use that directly for varint encoder.
			_, err := ae.Write(selfDescribingVarintBlock, wire.Varint, int64(uVal))
			return err
		}
		// Otherwise, use *big.Int for varint encoding.
		bigUVal := new(big.Int).SetUint64(uVal)
		_, err := ae.Write(selfDescribingVarintBlock, wire.Varint, bigUVal)
		return err

	case reflect.Float32, reflect.Float64:
		fVal := val.Float()
		if fVal == math.Trunc(fVal) && fVal >= float64(math.MinInt64) && fVal <= float64(math.MaxInt64) {
			// If float is a whole number and fits in int64, encode as SelfDescribingInt.
			if _, err := ae.coreBuf.Write(wire.SelfDescribingInt); err != nil {
				return err
			}
			varintBlockKey := wire.BlockKey("Int")
			varintElementType, blockDefOk := wire.SelfDescribingBlocks[varintBlockKey]
			if !blockDefOk {
				return fmt.Errorf("internal error: self-describing varint block key ('%s') not found for whole float", varintBlockKey)
			}
			selfDescribingVarintBlock := wire.NewBlockType(varintElementType, varintBlockKey, wire.MustDeduplicateByDefault(varintElementType))
			_, err := ae.Write(selfDescribingVarintBlock, wire.Varint, int64(fVal))
			return err
		}
		// Otherwise, encode as SelfDescribingFloat.
		if _, err := ae.coreBuf.Write(wire.SelfDescribingFloat); err != nil { // Float marker.
			return err
		}
		floatBlockKey := wire.BlockKey("Float") // Standard key for self-describing floats.
		floatElementType, blockDefOk := wire.SelfDescribingBlocks[floatBlockKey]
		if !blockDefOk {
			return fmt.Errorf("internal error: self-describing float block key ('%s') not found in wire.SelfDescribingBlocks map", floatBlockKey)
		}
		selfDescribingFloatBlock := wire.NewBlockType(floatElementType, floatBlockKey, wire.MustDeduplicateByDefault(floatElementType))
		_, err := ae.Write(selfDescribingFloatBlock, wire.Float64, fVal)
		return err

	case reflect.Bool:
		if v.(bool) {
			_, err := ae.coreBuf.Write(wire.SelfDescribingTrue) // True marker.
			return err
		}
		_, err := ae.coreBuf.Write(wire.SelfDescribingFalse) // False marker.
		return err

	case reflect.Ptr, reflect.Interface:
		if val.IsNil() { // Handle nil pointer or nil interface.
			_, err := ae.coreBuf.Write(wire.SelfDescribingNull)
			return err
		}
		// Dereference pointer/interface and recurse with the element.
		return ae.writeSelfDescribing(currentPath, val.Elem().Interface())

	default:
		// Handle *big.Int specifically, as it's a common type for large integers not caught by reflect.Int types.
		if bigIntValue, isBigInt := v.(*big.Int); isBigInt {
			if _, err := ae.coreBuf.Write(wire.SelfDescribingInt); err != nil {
				return err
			}
			varintBlockKey := wire.BlockKey("Int")
			varintElementType, blockDefOk := wire.SelfDescribingBlocks[varintBlockKey]
			if !blockDefOk {
				return fmt.Errorf("internal error: self-describing varint block key ('%s') not found for *big.Int", varintBlockKey)
			}
			selfDescribingVarintBlock := wire.NewBlockType(varintElementType, varintBlockKey, wire.MustDeduplicateByDefault(varintElementType))
			_, err := ae.Write(selfDescribingVarintBlock, wire.Varint, bigIntValue)
			return err
		}
		return fmt.Errorf("type error: cannot encode unsupported Go type %T (Kind: %s) in self-describing format at path %s", v, val.Kind(), util.FormatPath(currentPath))
	}
}

// ArgoErrorValue defines the structure for representing GraphQL errors in Argo format when not using self-describing errors.
// It includes standard GraphQL error fields. The `Path` is transformed into a list of integers/strings for Argo.
type ArgoErrorValue struct {
	Message    string                                      // The error message.
	Locations  []ArgoErrorLocation                         // Source locations of the error in the query.
	Path       []interface{}                               // Path to the field where the error occurred, as a list of string field names and integer indices.
	Extensions *orderedmap.OrderedMap[string, interface{}] // Custom error data.
}

// ArgoErrorLocation represents a single source location (line and column) for an error.
type ArgoErrorLocation struct {
	Line   int `json:"line"`   // 1-indexed line number.
	Column int `json:"column"` // 1-indexed column number.
}

// writeGoError converts a standard Go `error` into an ArgoErrorValue (represented as an *orderedmap.OrderedMap for deterministic field order)
// and then writes this map using the structured `wire.Error` type definition.
// This is invoked when the `HeaderSelfDescribingErrorsFlag` is false.
// `currentPath` is the GraphQL path to where the error label itself is being written.
func (ae *ArgoEncoder) writeGoError(currentPath ast.Path, goErr error) error {
	// Construct the ArgoErrorValue.
	argoErrVal := ArgoErrorValue{
		Message: goErr.Error(),
		// Locations and Path are typically not available from a generic Go error directly.
		// These would need to be populated if `goErr` is a more structured error type
		// that carries GraphQL-specific location/path information.
		// For now, we add the Go error type to extensions for some context.
		Extensions: orderedmap.NewOrderedMap[string, interface{}](),
	}
	argoErrVal.Extensions.Set("go_error_type", reflect.TypeOf(goErr).String())

	// Convert ArgoErrorValue to an *orderedmap.OrderedMap for encoding with wire.Error type.
	// The order of Set calls determines the field order in the Argo output if wire.Error is a RecordType.
	errorMap := orderedmap.NewOrderedMap[string, interface{}]()
	errorMap.Set("message", argoErrVal.Message)
	if argoErrVal.Locations != nil { // Only include if present.
		errorMap.Set("locations", argoErrVal.Locations)
	}
	if argoErrVal.Path != nil { // Only include if present.
		errorMap.Set("path", argoErrVal.Path)
	}
	if argoErrVal.Extensions != nil && argoErrVal.Extensions.Len() > 0 { // Only include if non-empty.
		errorMap.Set("extensions", argoErrVal.Extensions)
	}

	// Write the errorMap using the predefined wire.Error schema.
	// currentBlock is nil as errors are part of the core stream, not typically within other blocks.
	return ae.writeArgo(currentPath, errorMap, wire.Error, nil)
}

// GetResult finalizes the encoding process. It assembles the Argo header,
// data from all block writers (if not inlining everything), and the core buffer data
// into a single, final *buf.Buf containing the complete Argo message.
func (ae *ArgoEncoder) GetResult() (*buf.Buf, error) {
	headerBytes, err := ae.header.AsBytes() // Serialize the header to bytes.
	if err != nil {
		return nil, fmt.Errorf("failed to serialize Argo header: %w", err)
	}
	ae.Track(nil, "header bytes written", nil, headerBytes) // For debugging, length of header.

	shouldWriteBlocks := !ae.header.GetFlag(header.HeaderInlineEverythingFlag)
	totalDataBytesFromBlocks := 0   // Total size of content from all blocks.
	blockLengthLabelBytesTotal := 0 // Total size of all block length labels.

	// blockToWrite temporarily stores data for each block before final assembly.
	type blockToWrite struct {
		key             wire.BlockKey // The block's unique key.
		lengthLabelData []byte        // Encoded label for the total length of this block's content.
		contentBytes    [][]byte      // Slice of byte slices, each representing a value in the block.
	}
	var blocksToWrite []blockToWrite // List of blocks to be written, in order.

	if shouldWriteBlocks {
		// Iterate through writers in the order they were created (preserved by OrderedMap).
		// This ensures blocks are written in a deterministic order, matching reference implementations.
		for el := ae.writers.Front(); el != nil; el = el.Next() {
			key := el.Key
			entry := el.Value // writerEntry
			writer := entry.Writer
			originalValueType := entry.OriginalValueType // e.g. wire.String, wire.Bytes

			blockContentBytes := writer.AllValuesAsBytes() // Get all accumulated byte values for this block.
			currentBlockTotalBytes := 0

			isStringBlock := wire.IsString(originalValueType)

			processedBlockContentBytes := make([][]byte, 0, len(blockContentBytes))
			for _, valueBytes := range blockContentBytes {
				currentBlockTotalBytes += len(valueBytes)
				processedBlockContentBytes = append(processedBlockContentBytes, valueBytes)
				// If it's a string block and null termination is enabled, add terminator.
				if isStringBlock && ae.header.GetFlag(header.HeaderNullTerminatedStringsFlag) {
					processedBlockContentBytes = append(processedBlockContentBytes, nullTerminator)
					currentBlockTotalBytes += len(nullTerminator)
				}
			}

			lengthLabel := label.NewFromInt64(int64(currentBlockTotalBytes)) // Label for total length of this block.
			encodedLengthLabel := lengthLabel.Encode()

			blocksToWrite = append(blocksToWrite, blockToWrite{
				key:             key,
				lengthLabelData: encodedLengthLabel,
				contentBytes:    processedBlockContentBytes,
			})

			totalDataBytesFromBlocks += currentBlockTotalBytes
			blockLengthLabelBytesTotal += len(encodedLengthLabel)
		}
	}

	coreDataBytes := ae.coreBuf.Bytes() // Get all bytes from the core buffer (labels, inlined data).
	coreDataLength := len(coreDataBytes)
	var coreLengthLabelBytes []byte
	if shouldWriteBlocks { // If blocks are written, the core data also needs a length label.
		coreLengthLabel := label.NewFromInt64(int64(coreDataLength))
		coreLengthLabelBytes = coreLengthLabel.Encode()
	}

	// Calculate the total size of the final Argo message.
	finalSize := len(headerBytes)
	if shouldWriteBlocks {
		finalSize += blockLengthLabelBytesTotal // All block length labels.
		finalSize += totalDataBytesFromBlocks   // All block content.
		finalSize += len(coreLengthLabelBytes)  // Core data length label.
	}
	finalSize += coreDataLength // Core data itself.

	// Allocate the final buffer and write all parts in order.
	finalBuf := buf.NewBuf(finalSize)

	_, _ = finalBuf.Write(headerBytes) // 1. Header

	if shouldWriteBlocks {
		// 2. For each block: its length label, then its content.
		for _, btw := range blocksToWrite {
			_, _ = finalBuf.Write(btw.lengthLabelData)
			for _, valueData := range btw.contentBytes {
				_, _ = finalBuf.Write(valueData)
			}
		}
		// 3. Core data length label.
		_, _ = finalBuf.Write(coreLengthLabelBytes)
	}

	// 4. Core data.
	_, _ = finalBuf.Write(coreDataBytes)

	// Sanity check the final length.
	if finalBuf.Len() != finalSize {
		return nil, fmt.Errorf("internal encoder error: incorrect result length. Wrote %d, expected %d", finalBuf.Len(), finalSize)
	}

	return finalBuf, nil
}
