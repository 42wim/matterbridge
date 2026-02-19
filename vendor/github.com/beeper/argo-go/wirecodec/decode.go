package wirecodec

import (
	"encoding/binary"
	"fmt"
	"math"

	"github.com/beeper/argo-go/block"
	"github.com/beeper/argo-go/codec"
	"github.com/beeper/argo-go/header"
	"github.com/beeper/argo-go/label"
	"github.com/beeper/argo-go/pkg/buf"
	"github.com/beeper/argo-go/wire"
)

type anyBlockReader interface {
	Read(parent buf.Read) (interface{}, error)
}

type genericBlockReaderWrapper struct {
	coreRead        func(parentBuf buf.Read) (interface{}, error)
	blockDataBuffer buf.Read
}

func (g *genericBlockReaderWrapper) Read(parentBuf buf.Read) (interface{}, error) {
	return g.coreRead(parentBuf)
}

type Decoder struct {
	r       buf.Read
	slicer  *codec.MessageSlicer
	readers map[wire.BlockKey]anyBlockReader
}

func NewFromSlicer(s *codec.MessageSlicer) *Decoder {
	return &Decoder{
		r:       s.Core(),
		slicer:  s,
		readers: make(map[wire.BlockKey]anyBlockReader),
	}
}

func (d *Decoder) DecodeWireType() (wire.Type, error) {
	lbl, err := label.Read(d.r)
	if err != nil {
		return nil, err
	}

	switch {
	// ---------- scalars ----------
	case label.WireTypeMarkerString.Is(lbl):
		return wire.String, nil
	case label.WireTypeMarkerBoolean.Is(lbl):
		return wire.Boolean, nil
	case label.WireTypeMarkerVarint.Is(lbl):
		return wire.Varint, nil
	case label.WireTypeMarkerFloat64.Is(lbl):
		return wire.Float64, nil
	case label.WireTypeMarkerBytes.Is(lbl):
		return wire.Bytes, nil
	case label.WireTypeMarkerFixed.Is(lbl):
		lengthLbl, err := label.Read(d.r)
		if err != nil {
			return nil, err
		}
		n := int(lengthLbl.Value().Int64())
		if n < 0 {
			return nil, fmt.Errorf("wirecodec: negative FIXED length %d", n)
		}
		return wire.FixedType{Length: n}, nil

	// ---------- block ----------
	case label.WireTypeMarkerBlock.Is(lbl):
		// Only scalar/desc allowed for block element
		elem, err := d.decodeScalarWireType()
		if err != nil {
			return nil, err
		}
		key, err := d.readBlockString()
		if err != nil {
			return nil, err
		}
		dedupe, err := d.readBlockBool()
		if err != nil {
			return nil, err
		}
		return wire.NewBlockType(elem, wire.BlockKey(key), dedupe), nil

	// ---------- wrappers ----------
	case label.WireTypeMarkerNullable.Is(lbl):
		of, err := d.DecodeWireType()
		if err != nil {
			return nil, err
		}
		return wire.NewNullableType(of), nil

	case label.WireTypeMarkerArray.Is(lbl):
		of, err := d.DecodeWireType()
		if err != nil {
			return nil, err
		}
		return wire.ArrayType{Of: of}, nil

	// ---------- record ----------
	case label.WireTypeMarkerRecord.Is(lbl):
		lenLbl, err := label.Read(d.r)
		if err != nil {
			return nil, err
		}
		n := int(lenLbl.Value().Int64())
		if n < 0 {
			return nil, fmt.Errorf("wirecodec: negative record-field count %d", n)
		}
		return d.decodeRecord(n)

	// ---------- simple markers ----------
	case label.WireTypeMarkerDesc.Is(lbl):
		return wire.Desc, nil
	case label.WireTypeMarkerError.Is(lbl):
		return wire.Error, nil
	case label.WireTypeMarkerPath.Is(lbl):
		return wire.Path, nil
	case label.WireTypeMarkerExtensions.Is(lbl):
		return wire.Extensions, nil
	}

	return nil, fmt.Errorf("wirecodec: unknown wire-type label %s", lbl.Value().String())
}

func (d *Decoder) DecodeWireTypeStore() (map[string]wire.Type, error) {
	objLbl, err := label.Read(d.r)
	if err != nil {
		return nil, err
	}
	if !wire.SelfDescribingTypeMarkerObject.Is(objLbl) {
		return nil, fmt.Errorf("wirecodec: expected object marker, found %v", objLbl.Value())
	}

	cntLbl, err := label.Read(d.r)
	if err != nil {
		return nil, err
	}
	n := int(cntLbl.Value().Int64())
	if n < 0 {
		return nil, fmt.Errorf("wirecodec: negative type-store length %d", n)
	}

	out := make(map[string]wire.Type, n)
	for i := 0; i < n; i++ {
		name, err := d.readBlockString()
		if err != nil {
			return nil, fmt.Errorf("wirecodec: read store name: %w", err)
		}
		wt, err := d.DecodeWireType()
		if err != nil {
			return nil, fmt.Errorf("wirecodec: decode type %q: %w", name, err)
		}
		out[name] = wt
	}
	return out, nil
}

func (d *Decoder) decodeScalarWireType() (wire.Type, error) {
	lbl, err := label.Read(d.r)
	if err != nil {
		return nil, err
	}

	switch {
	case label.WireTypeMarkerString.Is(lbl):
		return wire.String, nil
	case label.WireTypeMarkerBoolean.Is(lbl):
		return wire.Boolean, nil
	case label.WireTypeMarkerVarint.Is(lbl):
		return wire.Varint, nil
	case label.WireTypeMarkerFloat64.Is(lbl):
		return wire.Float64, nil
	case label.WireTypeMarkerBytes.Is(lbl):
		return wire.Bytes, nil
	case label.WireTypeMarkerFixed.Is(lbl):
		lenLbl, err := label.Read(d.r)
		if err != nil {
			return nil, err
		}
		n := int(lenLbl.Value().Int64())
		if n < 0 {
			return nil, fmt.Errorf("wirecodec: negative FIXED length %d", n)
		}
		return wire.FixedType{Length: n}, nil
	case label.WireTypeMarkerDesc.Is(lbl):
		return wire.Desc, nil
	}
	return nil, fmt.Errorf("wirecodec: invalid scalar wire-type label %s", lbl.Value().String())
}

func (d *Decoder) decodeRecord(length int) (wire.Type, error) {
	if length < 0 {
		return nil, fmt.Errorf("wirecodec: negative record-field count %d", length)
	}
	fields := make([]wire.Field, 0, length)
	for i := 0; i < length; i++ {
		name, err := d.readBlockString()
		if err != nil {
			return nil, fmt.Errorf("record field %d name: %w", i, err)
		}
		ft, err := d.DecodeWireType()
		if err != nil {
			return nil, fmt.Errorf("record field %q type: %w", name, err)
		}
		omit, err := d.readBlockBool()
		if err != nil {
			return nil, fmt.Errorf("record field %q omittable: %w", name, err)
		}
		fields = append(fields, wire.Field{Name: name, Of: ft, Omittable: omit})
	}
	return wire.RecordType{Fields: fields}, nil
}

func (d *Decoder) readBlockString() (string, error) {
	key := wire.BlockKey("String")
	elem, ok := wire.SelfDescribingBlocks[key]
	if !ok {
		return "", fmt.Errorf("wirecodec: self-describing String block not found")
	}
	blk := wire.NewBlockType(elem, key, wire.MustDeduplicateByDefault(elem))

	r, err := d.getBlockReader(blk, wire.String)
	if err != nil {
		return "", fmt.Errorf("wirecodec: get String block reader: %w", err)
	}
	v, err := r.Read(d.r)
	if err != nil {
		return "", fmt.Errorf("wirecodec: read String from block: %w", err)
	}
	s, ok := v.(string)
	if !ok {
		return "", fmt.Errorf("wirecodec: String block reader returned %T", v)
	}
	return s, nil
}

func (d *Decoder) readBlockBool() (bool, error) {
	lbl, err := label.Read(d.r)
	if err != nil {
		return false, err
	}
	switch {
	case label.FalseMarker.Is(lbl):
		return false, nil
	case label.TrueMarker.Is(lbl):
		return true, nil
	default:
		return false, fmt.Errorf("wirecodec: expected boolean marker got %v", lbl.Value())
	}
}

func DecodeWireTypeStoreFile(msg []byte) (map[string]wire.Type, error) {
	slicer, err := codec.NewMessageSlicer(buf.NewBufReadonly(msg))
	if err != nil {
		return nil, err
	}
	return NewFromSlicer(slicer).DecodeWireTypeStore()
}

func (d *Decoder) getBlockReader(blockDef wire.BlockType, valueWireType wire.Type) (anyBlockReader, error) {
	if r, ok := d.readers[blockDef.Key]; ok {
		return r, nil
	}
	r, err := d.makeBlockReader(valueWireType, blockDef.Dedupe, blockDef.Key)
	if err != nil {
		return nil, err
	}
	d.readers[blockDef.Key] = r
	return r, nil
}

func (d *Decoder) makeBlockReader(valueWireType wire.Type, dedupe bool, key wire.BlockKey) (anyBlockReader, error) {
	blockData := d.slicer.NextBlock()
	if blockData == nil {
		if !d.slicer.Header().GetFlag(header.HeaderInlineEverythingFlag) {
			return nil, fmt.Errorf("wirecodec: no more block segments for key %s", key)
		}
		blockData = d.slicer.Core()
		if blockData == nil {
			return nil, fmt.Errorf("wirecodec: core buffer nil in inline-everything mode for %s", key)
		}
	}

	var coreRead func(parentBuf buf.Read) (interface{}, error)

	switch t := valueWireType.(type) {
	case wire.StringType:
		nt := d.slicer.Header().GetFlag(header.HeaderNullTerminatedStringsFlag)
		fromBytes := func(b []byte) string { return string(b) }
		if dedupe {
			r := block.NewDeduplicatingLabelBlockReader[string](blockData, fromBytes, nt)
			coreRead = func(p buf.Read) (interface{}, error) { return r.Read(p) }
		} else {
			r := block.NewLabelBlockReader[string](blockData, fromBytes, nt)
			coreRead = func(p buf.Read) (interface{}, error) { return r.Read(p) }
		}

	case wire.BytesType:
		fromBytes := func(b []byte) []byte { return b }
		if dedupe {
			r := block.NewDeduplicatingLabelBlockReader[[]byte](blockData, fromBytes, false)
			coreRead = func(p buf.Read) (interface{}, error) { return r.Read(p) }
		} else {
			r := block.NewLabelBlockReader[[]byte](blockData, fromBytes, false)
			coreRead = func(p buf.Read) (interface{}, error) { return r.Read(p) }
		}

	case wire.VarintType:
		if dedupe {
			return nil, fmt.Errorf("wirecodec: deduping VARINT via label blocks not supported for %s", key)
		}
		r := block.NewUnlabeledVarIntBlockReader(blockData)
		coreRead = func(p buf.Read) (interface{}, error) { return r.Read(p) }

	case wire.Float64Type:
		if dedupe {
			return nil, fmt.Errorf("wirecodec: deduping FLOAT64 not supported for %s", key)
		}
		fromBytes := func(b []byte) float64 { return math.Float64frombits(binary.LittleEndian.Uint64(b)) }
		r := block.NewFixedSizeBlockReader[float64](blockData, fromBytes, 8)
		coreRead = func(p buf.Read) (interface{}, error) { return r.Read(p) }

	case wire.FixedType:
		if dedupe {
			return nil, fmt.Errorf("wirecodec: deduping FIXED not supported for %s", key)
		}
		fromBytes := func(b []byte) []byte { return b }
		r := block.NewFixedSizeBlockReader[[]byte](blockData, fromBytes, t.Length)
		coreRead = func(p buf.Read) (interface{}, error) { return r.Read(p) }

	default:
		return nil, fmt.Errorf("wirecodec: unsupported block value type %s for key %s", wire.Print(valueWireType), key)
	}

	return &genericBlockReaderWrapper{coreRead: coreRead, blockDataBuffer: blockData}, nil
}
