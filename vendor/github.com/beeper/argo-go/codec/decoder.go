// Package codec provides the tools for encoding and decoding Argo data.
// This file focuses on the ArgoDecoder, which is responsible for parsing
// Argo binary messages and reconstructing them into Go data structures,
// typically an ordered map representing a GraphQL ExecutionResult.
package codec

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"strconv"

	"github.com/elliotchance/orderedmap/v3"
	"github.com/vektah/gqlparser/v2/ast"

	"github.com/beeper/argo-go/block"
	"github.com/beeper/argo-go/header"
	"github.com/beeper/argo-go/internal/util"
	"github.com/beeper/argo-go/label"
	"github.com/beeper/argo-go/pkg/buf"
	"github.com/beeper/argo-go/wire"
)

// anyBlockReader defines an interface for a generic block reader.
// Implementations of this interface are responsible for reading a specific
// type of data (e.g., string, varint) from a block within an Argo message.
// The Read method takes the parent buffer (which might be the core message buffer
// or a buffer for a specific record/array context) and returns the decoded data
// as an interface{} value, along with any error encountered.
type anyBlockReader interface {
	Read(parentBuf buf.Read) (interface{}, error)
}

// argoError represents a structured error encountered during Argo decoding.
// It includes the path within the data structure where the error occurred,
// a descriptive message, and the byte position in the input buffer.
type argoError struct {
	Path ast.Path
	Msg  string
	Pos  int64
}

// Error implements the error interface for argoError.
func (e *argoError) Error() string {
	return fmt.Sprintf("Argo decoding error at path %s (pos %d): %s", util.FormatPath(e.Path), e.Pos, e.Msg)
}

// newArgoError creates a new argoError with the given path, position, and formatted message.
func newArgoError(path ast.Path, pos int64, format string, args ...interface{}) error {
	return &argoError{Path: path, Pos: pos, Msg: fmt.Errorf(format, args...).Error()}
}

// ArgoDecoder decodes an Argo binary message into a Go data structure.
// It uses a MessageSlicer to access different parts of the Argo message (header, blocks, core)
// and maintains a map of block readers to efficiently decode block data.
type ArgoDecoder struct {
	slicer  *MessageSlicer
	readers map[wire.BlockKey]anyBlockReader // Caches block readers by their key.
}

// NewArgoDecoder creates and initializes a new ArgoDecoder.
// messageBuf should contain the entire Argo message to be decoded.
// It returns an error if the message slicer cannot be initialized (e.g., due to header read issues).
func NewArgoDecoder(messageBuf buf.Read) (*ArgoDecoder, error) {
	slicer, err := NewMessageSlicer(messageBuf)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize message slicer: %w", err)
	}
	return &ArgoDecoder{
		slicer:  slicer,
		readers: make(map[wire.BlockKey]anyBlockReader),
	}, nil
}

// ArgoToMap decodes the entire Argo message into an ordered map, which typically
// represents a GraphQL ExecutionResult. The wire.Type `wt` specifies the expected
// structure of the data. If the Argo message header indicates it is self-describing,
// the provided `wt` is overridden by `wire.Desc`.
func (ad *ArgoDecoder) ArgoToMap(wt wire.Type) (*orderedmap.OrderedMap[string, interface{}], error) {
	finalWt := wt

	if _, wantDesc := wt.(wire.DescType); wantDesc && ad.slicer.Header().GetFlag(header.HeaderSelfDescribingFlag) {
		finalWt = wire.Desc
	}

	if p, ok := ad.slicer.Core().(buf.BufPosition); ok {
		p.SetPosition(0)
	}
	result, err := ad.readArgo(ad.slicer.Core(), nil, finalWt, nil)
	if err != nil {
		return nil, err
	}
	if m, ok := result.(*orderedmap.OrderedMap[string, interface{}]); ok {
		return m, nil
	}
	return nil, fmt.Errorf("decoded result is not an ordered map, got %T", result)
}

func (ad *ArgoDecoder) readArgo(b buf.Read, currentPath ast.Path, wt wire.Type, currentBlock *wire.BlockType) (interface{}, error) {
	switch typedWt := wt.(type) {
	case wire.BlockType:
		trackVal := orderedmap.NewOrderedMap[string, interface{}]()
		trackVal.Set("key", typedWt.Key)
		trackVal.Set("dedupe", typedWt.Dedupe)
		return ad.readArgo(b, currentPath, typedWt.Of, &typedWt)

	case wire.NullableType:
		peekBytes, err := b.Peek(1)
		if err != nil {
			if err == io.EOF { // EOF here means the buffer ended before a nullable marker could be read.
				return nil, newArgoError(currentPath, b.Position(), "unexpected EOF while peeking for nullable type marker")
			}
			return nil, newArgoError(currentPath, b.Position(), "failed to peek for nullable type marker: %w", err)
		}
		peekLabelByte := peekBytes[0]

		if peekLabelByte == label.Null[0] { // Compare the first byte of the pre-encoded Null label.
			_, _ = b.ReadByte() // Consume the null label byte.
			return nil, nil
		}
		if peekLabelByte == label.Absent[0] { // Compare the first byte of the pre-encoded Absent label.
			_, _ = b.ReadByte()          // Consume the absent label byte.
			return wire.AbsentValue, nil // Return a special marker for absent, to be handled by RECORD logic.
		}
		if peekLabelByte == label.Error[0] { // Compare the first byte of the pre-encoded Error label.
			_, _ = b.ReadByte() // Consume the error label byte.

			lengthLabel, err := label.Read(b)
			if err != nil {
				return nil, newArgoError(currentPath, b.Position(), "failed to read error array length: %w", err)
			}
			length := int(lengthLabel.Value().Int64())
			if length < 0 {
				return nil, newArgoError(currentPath, b.Position(), "invalid negative error array length: %d", length)
			}

			// Argo Spec: Field errors propagate to the nearest nullable field.
			// The errors are then written. The `path` field in these errors is relative.
			// "implementations should make full path easily available to users."
			// The spec also says: "return null // simple for compatibility, but up to implementations what to do with inline errors"
			// We collect the errors and then return nil for the field value itself, as per spec.
			// The collected `errors` are not directly returned by this function but could be logged
			// or otherwise handled by the calling application if needed.
			errors := make([]interface{}, length)
			for i := 0; i < length; i++ {
				var errItem interface{}
				errPath := util.AddPathIndex(currentPath, i)
				if ad.slicer.Header().GetFlag(header.HeaderSelfDescribingErrorsFlag) {
					errItem, err = ad.readSelfDescribing(b, errPath)
				} else {
					errItem, err = ad.readArgo(b, errPath, wire.Error, nil) // currentBlock is nil for Error type
				}
				if err != nil {
					return nil, newArgoError(errPath, b.Position(), "failed to read error item %d: %w", i, err)
				}
				errors[i] = errItem
			}
			// The value of the field is null when there's an inline error.
			// The collected `errors` array here is for potential side-channel processing (e.g. logging)
			// but is not part of the main decoded result for this field.
			return nil, nil
		}

		if !wire.IsLabeled(typedWt.Of) {
			// For non-labeled types within a Nullable, expect a NonNullMarker (value 0) if not null/absent/error.
			marker, err := label.Read(b)
			if err != nil {
				return nil, newArgoError(currentPath, b.Position(), "failed to read non-null marker: %w", err)
			}
			if !label.NonNullMarker.Is(marker) {
				return nil, newArgoError(currentPath, b.Position(),
					"invalid non-null marker %s for %s. Path: %s, Pos: %d",
					marker.Value().String(), wire.Print(wt), util.FormatPath(currentPath), b.Position())
			}
		}
		// If labeled, no non-null marker is read here; the underlying type's label reading handles it.
		// Proceed to read the actual underlying type.
		return ad.readArgo(b, currentPath, typedWt.Of, currentBlock)

	case wire.RecordType:
		obj := orderedmap.NewOrderedMapWithCapacity[string, interface{}](len(typedWt.Fields))

		for _, field := range typedWt.Fields {
			fieldPath := util.AddPathName(currentPath, field.Name)

			var fieldValue interface{}
			var err error

			if field.Omittable {
				peekBytes, errPeek := b.Peek(1)
				if errPeek != nil {
					return nil, newArgoError(fieldPath, b.Position(), "failed to peek for omittable field %s: %w", field.Name, errPeek)
				}
				labelPeekByte := peekBytes[0]

				// Regarding Error labels for omittable fields:
				// The Argo spec states: "Nullable fields are the only valid location for Field errors".
				// If field.Of is Nullable, the wire.NullableType case in this function will handle any Error labels.
				// If field.Of is not Nullable, an Error label should not appear here according to the spec.
				// The reference JS implementation's check for `labelPeek == Label.Error[0]` in this context
				// might be for a slightly different interpretation or an older version of the spec.
				// This Go implementation adheres to "Error labels only on Nullable".

				if !wire.IsLabeled(field.Of) && labelPeekByte == label.NonNull[0] {
					_, _ = b.ReadByte() // Consume non-null marker.
					fieldValue, err = ad.readArgo(b, fieldPath, field.Of, currentBlock)
				} else if labelPeekByte == label.Absent[0] {
					_, _ = b.ReadByte() // Consume absent marker.
					fieldValue = wire.AbsentValue
				} else {
					// Neither NonNull (for unlabeled types) nor Absent.
					// Proceed to read normally:
					// - If field.Of is labeled, its own label will be read by the recursive call.
					// - If field.Of is unlabeled (and we didn't see a NonNull marker), it's a direct value.
					fieldValue, err = ad.readArgo(b, fieldPath, field.Of, currentBlock)
				}
			} else { // Not omittable.
				fieldValue, err = ad.readArgo(b, fieldPath, field.Of, currentBlock)
			}

			if err != nil {
				return nil, err // Error already contextualized by recursive call
			}

			if fieldValue != wire.AbsentValue {
				obj.Set(field.Name, fieldValue)
			}
		}
		return obj, nil

	case wire.ArrayType:
		lengthLabel, err := label.Read(b)
		if err != nil {
			return nil, newArgoError(currentPath, b.Position(), "failed to read array length: %w", err)
		}
		length := int(lengthLabel.Value().Int64())
		if length < 0 {
			return nil, newArgoError(currentPath, b.Position(), "invalid negative array length: %d", length)
		}

		arr := make([]interface{}, length) // This is a slice, not a map.
		for i := 0; i < length; i++ {
			itemPath := util.AddPathIndex(currentPath, i)
			item, err := ad.readArgo(b, itemPath, typedWt.Of, currentBlock)
			if err != nil {
				return nil, newArgoError(itemPath, b.Position(), "failed to read array item %d: %w", i, err)
			}
			if item == wire.AbsentValue {
				// JSON arrays don't have "absent" elements, they'd be null.
				// For now, treat absent as null in an array.
				arr[i] = nil
			} else {
				arr[i] = item
			}
		}
		return arr, nil

	case wire.BooleanType:
		l, err := label.Read(b)
		if err != nil {
			return nil, newArgoError(currentPath, b.Position(), "failed to read boolean label: %w", err)
		}
		if label.FalseMarker.Is(l) {
			return false, nil
		}
		if label.TrueMarker.Is(l) {
			return true, nil
		}
		return nil, newArgoError(currentPath, b.Position(), "invalid boolean label %s", l.Value().String())

	case wire.StringType, wire.BytesType, wire.VarintType, wire.Float64Type, wire.FixedType:
		if currentBlock == nil {
			return nil, newArgoError(currentPath, b.Position(), "programmer error: need block for %s", wire.Print(wt))
		}
		reader, err := ad.getBlockReader(*currentBlock, wt)
		if err != nil {
			return nil, newArgoError(currentPath, b.Position(), "failed to get block reader for %s (key %s): %w", wire.Print(wt), currentBlock.Key, err)
		}
		value, err := reader.Read(b) // `b` is the parent buffer (core or record/array context)
		if err != nil {
			return nil, newArgoError(currentPath, b.Position(), "block reader failed for %s (key %s): %w", wire.Print(wt), currentBlock.Key, err)
		}

		return value, nil

	case wire.DescType:
		return ad.readSelfDescribing(b, currentPath)

	case wire.PathType: // Path type is usually part of Error structure.
		// Decoding a raw PATH here would mean reading an ARRAY of VARINT.
		// This is essentially ArrayType{Of: Varint}
		return ad.readArgo(b, currentPath, wire.ArrayType{Of: wire.Varint}, currentBlock)

	default:
		return nil, newArgoError(currentPath, b.Position(), "unsupported wire type %T: %s", wt, wire.Print(wt))
	}
}

func (ad *ArgoDecoder) readSelfDescribing(b buf.Read, currentPath ast.Path) (interface{}, error) {
	typeMarkerLabel, err := label.Read(b)
	if err != nil {
		return nil, newArgoError(currentPath, b.Position(), "failed to read self-describing type marker: %w", err)
	}

	switch {
	case wire.SelfDescribingTypeMarkerNull.Is(typeMarkerLabel):
		return nil, nil
	case wire.SelfDescribingTypeMarkerFalse.Is(typeMarkerLabel):
		return false, nil
	case wire.SelfDescribingTypeMarkerTrue.Is(typeMarkerLabel):
		return true, nil
	case wire.SelfDescribingTypeMarkerObject.Is(typeMarkerLabel):
		lengthLabel, err := label.Read(b)
		if err != nil {
			return nil, newArgoError(currentPath, b.Position(), "failed to read self-describing object length: %w", err)
		}
		length := int(lengthLabel.Value().Int64())
		if length < 0 {
			return nil, newArgoError(currentPath, b.Position(), "invalid negative self-describing object length: %d", length)
		}
		obj := orderedmap.NewOrderedMapWithCapacity[string, interface{}](length)

		for i := 0; i < length; i++ {
			// Field name is a string, using the self-describing string block
			fieldNamePath := util.AddPathName(currentPath, strconv.Itoa(i)+"_key") // Path for the key itself

			// Construct the String BlockType for self-describing field names
			stringBlockKey := wire.BlockKey("String")
			stringElementType, ok := wire.SelfDescribingBlocks[stringBlockKey]
			if !ok {
				return nil, newArgoError(fieldNamePath, b.Position(), "self-describing string block key not found in map")
			}
			selfDescribingStringBlock := wire.NewBlockType(stringElementType, stringBlockKey, wire.MustDeduplicateByDefault(stringElementType))

			fieldNameVal, err := ad.readArgo(b, fieldNamePath, wire.String, &selfDescribingStringBlock)
			if err != nil {
				return nil, newArgoError(fieldNamePath, b.Position(), "failed to read self-describing object field name %d: %w", i, err)
			}
			fieldName, ok := fieldNameVal.(string)
			if !ok {
				return nil, newArgoError(fieldNamePath, b.Position(), "self-describing object field name %d is not a string", i)
			}

			// Field value is recursively self-describing
			fieldValuePath := util.AddPathName(currentPath, fieldName) // Path for the value
			value, err := ad.readSelfDescribing(b, fieldValuePath)
			if err != nil {
				return nil, newArgoError(fieldValuePath, b.Position(), "failed to read self-describing object field value for '%s': %w", fieldName, err)
			}
			obj.Set(fieldName, value)
		}
		return obj, nil

	case wire.SelfDescribingTypeMarkerList.Is(typeMarkerLabel):
		lengthLabel, err := label.Read(b)
		if err != nil {
			return nil, newArgoError(currentPath, b.Position(), "failed to read self-describing list length: %w", err)
		}
		length := int(lengthLabel.Value().Int64())
		if length < 0 {
			return nil, newArgoError(currentPath, b.Position(), "invalid negative self-describing list length: %d", length)
		}

		list := make([]interface{}, length) // This is a slice, not a map.
		for i := 0; i < length; i++ {
			itemPath := util.AddPathIndex(currentPath, i)
			item, err := ad.readSelfDescribing(b, itemPath)
			if err != nil {
				return nil, newArgoError(itemPath, b.Position(), "failed to read self-describing list item %d: %w", i, err)
			}
			list[i] = item
		}
		return list, nil

	case wire.SelfDescribingTypeMarkerString.Is(typeMarkerLabel):
		stringBlockKey := wire.BlockKey("String")
		stringElementType, ok := wire.SelfDescribingBlocks[stringBlockKey]
		if !ok {
			return nil, newArgoError(currentPath, b.Position(), "self-describing string block key not found in map")
		}
		selfDescribingStringBlock := wire.NewBlockType(stringElementType, stringBlockKey, wire.MustDeduplicateByDefault(stringElementType))
		return ad.readArgo(b, currentPath, wire.String, &selfDescribingStringBlock)
	case wire.SelfDescribingTypeMarkerBytes.Is(typeMarkerLabel):
		bytesBlockKey := wire.BlockKey("Bytes")
		bytesElementType, ok := wire.SelfDescribingBlocks[bytesBlockKey]
		if !ok {
			return nil, newArgoError(currentPath, b.Position(), "self-describing bytes block key not found in map")
		}
		selfDescribingBytesBlock := wire.NewBlockType(bytesElementType, bytesBlockKey, wire.MustDeduplicateByDefault(bytesElementType))
		return ad.readArgo(b, currentPath, wire.Bytes, &selfDescribingBytesBlock)
	case wire.SelfDescribingTypeMarkerInt.Is(typeMarkerLabel):
		varintBlockKey := wire.BlockKey("Int") // As defined in wire/wire.go init for VarintBlock
		varintElementType, ok := wire.SelfDescribingBlocks[varintBlockKey]
		if !ok {
			return nil, newArgoError(currentPath, b.Position(), "self-describing varint block key ('Int') not found in map")
		}
		selfDescribingVarintBlock := wire.NewBlockType(varintElementType, varintBlockKey, wire.MustDeduplicateByDefault(varintElementType))
		return ad.readArgo(b, currentPath, wire.Varint, &selfDescribingVarintBlock)
	case wire.SelfDescribingTypeMarkerFloat.Is(typeMarkerLabel):
		// The reference JS implementation uses "Float" as the key for self-describing float blocks.
		// Ensure wire.SelfDescribingBlocks is populated with this key if this path is taken.
		// Currently, wire.go's init for SelfDescribingBlocks does not include a "Float" block.
		// This might indicate a mismatch or an untested path.
		// For now, we assume "Float" is the intended key.
		floatBlockKey := wire.BlockKey("Float")
		floatElementType, ok := wire.SelfDescribingBlocks[floatBlockKey]
		if !ok {
			// Let's add a more specific error, and check wire.go's SelfDescribingBlocks initialization.
			return nil, newArgoError(currentPath, b.Position(), "self-describing float block key ('Float') not found in wire.SelfDescribingBlocks map. Check wire.go.")
		}
		selfDescribingFloatBlock := wire.NewBlockType(floatElementType, floatBlockKey, wire.MustDeduplicateByDefault(floatElementType))
		return ad.readArgo(b, currentPath, wire.Float64, &selfDescribingFloatBlock)
	default:
		return nil, newArgoError(currentPath, b.Position(), "invalid self-describing type marker: %s", typeMarkerLabel.Value().String())
	}
}

// getBlockReader retrieves or creates and caches a block reader for a given block definition.
func (ad *ArgoDecoder) getBlockReader(blockDef wire.BlockType, valueWireType wire.Type) (anyBlockReader, error) {
	if reader, found := ad.readers[blockDef.Key]; found {
		return reader, nil
	}
	reader, err := ad.makeBlockReader(valueWireType, blockDef.Dedupe, blockDef.Key)
	if err != nil {
		return nil, err
	}
	ad.readers[blockDef.Key] = reader
	return reader, nil
}

// and handles type-specific post-processing.
type genericBlockReaderWrapper struct {
	// coreRead is the underlying block-specific read function.
	// It's typically a method from a block.Reader implementation (e.g., block.LabelBlockReader.Read).
	coreRead func(parentBuf buf.Read) (interface{}, error)
	// blockDataBuffer is the buffer specific to this block's data.
	// It's used by some block readers that need to manage their own data consumption
	// (e.g., for null termination checks, though this is now handled internally by block readers).
	blockDataBuffer buf.Read
}

func (g *genericBlockReaderWrapper) Read(parentBuf buf.Read) (interface{}, error) {
	val, err := g.coreRead(parentBuf)
	if err != nil {
		return nil, err
	}
	// Null termination for strings is handled by the underlying block.Reader implementations
	// (e.g., LabelBlockReader) themselves, so no additional logic is needed here.
	return val, nil
}

// makeBlockReader creates a new block reader (wrapped by genericBlockReaderWrapper)
// based on the value's wire type, deduplication flag, and block key.
// It sources the block's data buffer from the MessageSlicer.
func (ad *ArgoDecoder) makeBlockReader(valueWireType wire.Type, dedupe bool, key wire.BlockKey) (anyBlockReader, error) {
	// Each new block type gets its own data buffer from the slicer,
	// unless HeaderInlineEverythingFlag is set, in which case all readers use the core buffer.
	blockDataForReader := ad.slicer.NextBlock()
	if blockDataForReader == nil {
		// This occurs if slicer.NextBlock() returns nil, meaning no more distinct blocks
		// are available from the message segments.
		if !ad.slicer.Header().GetFlag(header.HeaderInlineEverythingFlag) {
			return nil, fmt.Errorf("no more blocks available in slicer for key %s; schema may expect more blocks than provided in message", key)
		}
		// If HeaderInlineEverythingFlag is true, all "block" data is part of the core buffer.
		// The slicer.NextBlock() will return the coreBuffer in this mode.
		// If it's still nil here, it means slicer.Core() itself is nil, which is unexpected.
		// However, current slicer.NextBlock() logic should return slicer.Core() if the flag is set.
		// Re-assign for clarity if HeaderInlineEverythingFlag is set and NextBlock() somehow wasn't core.
		blockDataForReader = ad.slicer.Core()
		if blockDataForReader == nil {
			// This would be a critical issue with slicer initialization or logic.
			return nil, fmt.Errorf("internal error: core buffer is nil in inline-everything mode for key %s", key)
		}
	}

	var coreReadFunc func(parentBuf buf.Read) (interface{}, error)
	shouldReadNullTerminator := false

	switch t := valueWireType.(type) {
	case wire.StringType:
		shouldReadNullTerminator = ad.slicer.Header().GetFlag(header.HeaderNullTerminatedStringsFlag)
		fromBytes := func(b []byte) string { return string(b) }
		if dedupe {
			r := block.NewDeduplicatingLabelBlockReader[string](blockDataForReader, fromBytes, shouldReadNullTerminator)
			coreReadFunc = func(pbuf buf.Read) (interface{}, error) { return r.Read(pbuf) }
		} else {
			r := block.NewLabelBlockReader[string](blockDataForReader, fromBytes, shouldReadNullTerminator)
			coreReadFunc = func(pbuf buf.Read) (interface{}, error) { return r.Read(pbuf) }
		}
	case wire.BytesType:
		fromBytes := func(b []byte) []byte { return b } // No copy, direct use
		if dedupe {
			// BytesType never has null termination, so pass false
			r := block.NewDeduplicatingLabelBlockReader[[]byte](blockDataForReader, fromBytes, false)
			coreReadFunc = func(pbuf buf.Read) (interface{}, error) { return r.Read(pbuf) }
		} else {
			// BytesType never has null termination, so pass false
			r := block.NewLabelBlockReader[[]byte](blockDataForReader, fromBytes, false)
			coreReadFunc = func(pbuf buf.Read) (interface{}, error) { return r.Read(pbuf) }
		}
	case wire.VarintType:
		// Deduping VARINT not typically done this way via LabelBlockReader.
		if dedupe {
			return nil, fmt.Errorf("unimplemented: deduping VARINT with LabelBlockReader for key %s", key)
		}
		// UnlabeledVarIntBlockReader reads varint directly from its data buffer (blockDataForReader).
		// The parentBuf (core context) is not used by its Read method for label.
		r := block.NewUnlabeledVarIntBlockReader(blockDataForReader)
		coreReadFunc = func(pbuf buf.Read) (interface{}, error) { return r.Read(pbuf) }
	case wire.Float64Type:
		if dedupe {
			return nil, fmt.Errorf("unimplemented: deduping FLOAT64 for key %s", key)
		}
		// FixedSizeBlockReader for FLOAT64 reads from its data buffer. No label in parentBuf.
		fromBytes := func(b []byte) float64 { return math.Float64frombits(binary.LittleEndian.Uint64(b)) }
		r := block.NewFixedSizeBlockReader[float64](blockDataForReader, fromBytes, 8) // Float64 is 8 bytes
		coreReadFunc = func(pbuf buf.Read) (interface{}, error) { return r.Read(pbuf) }

	case wire.FixedType:
		if dedupe {
			return nil, fmt.Errorf("unimplemented: deduping FIXED for key %s", key)
		}
		// FixedSizeBlockReader for FIXED reads from its data buffer. No label in parentBuf.
		fromBytes := func(b []byte) []byte { return b }
		r := block.NewFixedSizeBlockReader[[]byte](blockDataForReader, fromBytes, t.Length)
		coreReadFunc = func(pbuf buf.Read) (interface{}, error) { return r.Read(pbuf) }
	default:
		return nil, fmt.Errorf("unsupported block value type %s for key %s", wire.Print(valueWireType), key)
	}

	return &genericBlockReaderWrapper{
		coreRead:        coreReadFunc,
		blockDataBuffer: blockDataForReader,
	}, nil
}

// MessageSlicer is responsible for parsing an Argo message into its constituent parts:
// the header, data blocks (if any), and the core data.
// It provides buffered views (buf.Read) into these parts without copying the underlying message data.
type MessageSlicer struct {
	hdr            *header.Header
	coreBuffer     buf.Read
	allSegments    [][]byte // Stores all byte slices: data blocks first, then the core data as the last segment.
	nextBlockIndex int      // Tracks the next data block to be vended by NextBlock().
}

// NewMessageSlicer creates a MessageSlicer from a buffer containing the entire Argo message.
// It reads the header and then parses out the block segments and the final core data segment
// based on length prefixes, unless the HeaderInlineEverythingFlag is set.
func NewMessageSlicer(fullMessageBuf buf.Read) (*MessageSlicer, error) {
	s := &MessageSlicer{}

	s.hdr = header.NewHeader()
	if err := s.hdr.Read(fullMessageBuf); err != nil {
		return nil, fmt.Errorf("failed to read header: %w", err)
	}

	if s.hdr.GetFlag(header.HeaderInlineEverythingFlag) {
		// The rest of the buffer is the core. There are no separate blocks.
		remainingBytes := make([]byte, fullMessageBuf.Len()-int(fullMessageBuf.Position()))
		_, err := io.ReadFull(fullMessageBuf, remainingBytes)
		if err != nil {
			return nil, fmt.Errorf("failed to read inline core data: %w", err)
		}
		s.allSegments = [][]byte{remainingBytes}
		s.coreBuffer = buf.NewBufReadonly(remainingBytes)

	} else {
		// Read all length-prefixed segments. The last one is the core.
		var segments [][]byte
		// fullMessageBuf is now positioned after the header.
		for fullMessageBuf.Position() < int64(fullMessageBuf.Len()) {
			lengthLabel, err := label.Read(fullMessageBuf)
			if err != nil {
				// If EOF and we expected more segments, or segments is empty, it's an error.
				if err == io.EOF && len(segments) > 0 { // EOF after reading some blocks, means core might be missing length
					return nil, fmt.Errorf("unexpected EOF after reading %d blocks, expecting core: %w", len(segments), err)
				}
				return nil, fmt.Errorf("failed to read segment length label: %w", err)
			}

			blockLengthVal := lengthLabel.Value().Int64()
			if blockLengthVal < 0 {
				return nil, fmt.Errorf("invalid negative segment length: %d", blockLengthVal)
			}
			blockLength := int(blockLengthVal)

			if fullMessageBuf.Position()+int64(blockLength) > int64(fullMessageBuf.Len()) {
				return nil, fmt.Errorf("segment length %d exceeds remaining buffer size %d", blockLength, int64(fullMessageBuf.Len())-fullMessageBuf.Position())
			}

			segmentBytes := make([]byte, blockLength)
			n, err := io.ReadFull(fullMessageBuf, segmentBytes)
			if err != nil {
				return nil, fmt.Errorf("failed to read segment data (expected %d, got %d): %w", blockLength, n, err)
			}
			segments = append(segments, segmentBytes)
		}

		if len(segments) == 0 {
			// This implies header was read, but no blocks and no core followed.
			// The spec implies at least a core (even if empty) prefixed by its length.
			return nil, fmt.Errorf("no blocks or core data found after header")
		}
		s.allSegments = segments
		s.coreBuffer = buf.NewBufReadonly(s.allSegments[len(s.allSegments)-1])
	}

	return s, nil
}

// Header returns the parsed message header.
func (s *MessageSlicer) Header() *header.Header {
	return s.hdr
}

// Core returns a read buffer for the core data part of the message.
// This buffer contains the main payload after all block definitions.
func (s *MessageSlicer) Core() buf.Read {
	return s.coreBuffer
}

// NextBlock returns a read buffer for the next data block in the message.
// If the HeaderInlineEverythingFlag is set in the header, this method
// will repeatedly return the coreBuffer, as all data is considered inline.
// Otherwise, it iterates through the pre-parsed data block segments.
// It returns nil if all data blocks have been vended or if in inline mode
// and no more distinct blocks were expected by the schema logic.
func (s *MessageSlicer) NextBlock() buf.Read {
	if s.hdr.GetFlag(header.HeaderInlineEverythingFlag) {
		// In inline mode, the "block" is the core itself.
		// The same coreBuffer instance is returned. Reads from this buffer
		// will advance its position, which is the expected behavior for inline scalar values.
		return s.coreBuffer
	}

	// Not inlineEverything: vend distinct block segments from allSegments.
	// allSegments contains data blocks followed by the core data as the last element.
	// We only vend actual data blocks here (i.e., segments before the final core segment).
	if s.nextBlockIndex < len(s.allSegments)-1 {
		blockData := s.allSegments[s.nextBlockIndex]
		s.nextBlockIndex++
		return buf.NewBufReadonly(blockData)
	}

	// All distinct data blocks have been vended.
	// Subsequent calls for new block types (if the schema expects more than were in the message)
	// will receive nil.
	return nil
}
