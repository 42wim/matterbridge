package logutils

import (
	"fmt"
	"math"
	"sync"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/ethereum/go-ethereum/log"

	"github.com/status-im/status-go/protocol/zaputil"
)

type gethLoggerCore struct {
	zapcore.LevelEnabler
	fields []zapcore.Field
	logger log.Logger
}

func (c gethLoggerCore) With(fields []zapcore.Field) zapcore.Core {
	clone := c.clone()
	clone.fields = append(clone.fields, fields...)
	return clone
}

func (c gethLoggerCore) Check(ent zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if c.Enabled(ent.Level) {
		return ce.AddCore(ent, c)
	}
	return ce
}
func (c gethLoggerCore) Write(ent zapcore.Entry, fields []zapcore.Field) error {
	fields = append(c.fields[:], fields...)

	var args []interface{}
	for _, f := range fields {
		switch f.Type {
		case zapcore.ArrayMarshalerType,
			zapcore.ObjectMarshalerType,
			zapcore.BinaryType,
			zapcore.ByteStringType,
			zapcore.Complex128Type,
			zapcore.ReflectType,
			zapcore.StringerType,
			zapcore.ErrorType:
			args = append(args, f.Key, f.Interface)
		case zapcore.BoolType:
			args = append(args, f.Key, f.Integer == 1)
		case zapcore.DurationType:
			args = append(args, f.Key, time.Duration(f.Integer))
		case zapcore.Float64Type:
			args = append(args, f.Key, math.Float64frombits(uint64(f.Integer)))
		case zapcore.Float32Type:
			args = append(args, f.Key, math.Float32frombits(uint32(f.Integer)))
		case zapcore.Int64Type,
			zapcore.Int32Type,
			zapcore.Int16Type,
			zapcore.Int8Type,
			zapcore.Uint64Type,
			zapcore.Uint32Type,
			zapcore.Uint16Type,
			zapcore.Uint8Type:
			args = append(args, f.Key, f.Integer)
		case zapcore.UintptrType:
			args = append(args, f.Key, uintptr(f.Integer))
		case zapcore.StringType:
			args = append(args, f.Key, f.String)
		case zapcore.TimeType:
			if f.Interface != nil {
				args = append(args, f.Key, time.Unix(0, f.Integer).In(f.Interface.(*time.Location)))
			} else {
				// Fall back to UTC if location is nil.
				args = append(args, f.Key, time.Unix(0, f.Integer))
			}
		case zapcore.NamespaceType:
			args = append(args, "namespace", f.Key)
		case zapcore.SkipType:
			break
		default:
			panic(fmt.Sprintf("unknown field type: %v", f))
		}
	}

	// set callDepth to 3 for `Output` to skip the calls to zap.Logger
	// and get the correct caller in the log
	callDepth := 3
	switch ent.Level {
	case zapcore.DebugLevel:
		c.logger.Output(ent.Message, log.LvlDebug, callDepth, args...)
	case zapcore.InfoLevel:
		c.logger.Output(ent.Message, log.LvlInfo, callDepth, args...)
	case zapcore.WarnLevel:
		c.logger.Output(ent.Message, log.LvlWarn, callDepth, args...)
	case zapcore.ErrorLevel:
		c.logger.Output(ent.Message, log.LvlError, callDepth, args...)
	case zapcore.DPanicLevel, zapcore.PanicLevel, zapcore.FatalLevel:
		c.logger.Output(ent.Message, log.LvlCrit, callDepth, args...)
	}

	return nil
}

func (gethLoggerCore) Sync() error {
	return nil
}

func (c *gethLoggerCore) clone() *gethLoggerCore {
	return &gethLoggerCore{
		LevelEnabler: c.LevelEnabler,
		fields:       c.fields,
		logger:       c.logger,
	}
}

// NewZapAdapter returns a new zapcore.Core interface which forwards logs to log.Logger.
func NewZapAdapter(logger log.Logger, enab zapcore.LevelEnabler) zapcore.Core {
	return &gethLoggerCore{
		LevelEnabler: enab,
		logger:       logger,
	}
}

var registerOnce sync.Once

// NewZapLoggerWithAdapter returns a logger forwarding all logs with level info and above.
func NewZapLoggerWithAdapter(logger log.Logger) (*zap.Logger, error) {
	registerOnce.Do(func() {
		if err := zaputil.RegisterJSONHexEncoder(); err != nil {
			panic(err)
		}
	})

	cfg := zap.Config{
		Level:            zap.NewAtomicLevelAt(zapcore.DebugLevel),
		Development:      false,
		Sampling:         nil,
		Encoding:         "json-hex",
		EncoderConfig:    zap.NewProductionEncoderConfig(),
		OutputPaths:      []string{"stderr"},
		ErrorOutputPaths: []string{"stderr"},
	}
	adapter := zap.WrapCore(
		func(zapcore.Core) zapcore.Core {
			return NewZapAdapter(logger, cfg.Level)
		},
	)
	log.PrintOrigins(true)
	return cfg.Build(adapter)
}
