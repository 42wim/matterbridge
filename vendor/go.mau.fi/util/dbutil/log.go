package dbutil

import (
	"context"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

type DatabaseLogger interface {
	QueryTiming(ctx context.Context, method, query string, args []any, nrows int, duration time.Duration, err error)
	WarnUnsupportedVersion(current, compat, latest int)
	PrepareUpgrade(current, compat, latest int)
	DoUpgrade(from, to int, message string, txn TxnMode)
	// Deprecated: legacy warning method, return errors instead
	Warn(msg string, args ...any)
}

type noopLogger struct{}

var NoopLogger DatabaseLogger = &noopLogger{}

func (n noopLogger) WarnUnsupportedVersion(_, _, _ int)      {}
func (n noopLogger) PrepareUpgrade(_, _, _ int)              {}
func (n noopLogger) DoUpgrade(_, _ int, _ string, _ TxnMode) {}
func (n noopLogger) Warn(msg string, args ...any)            {}

func (n noopLogger) QueryTiming(_ context.Context, _, _ string, _ []any, _ int, _ time.Duration, _ error) {
}

type zeroLogger struct {
	l *zerolog.Logger
	ZeroLogSettings
}

type ZeroLogSettings struct {
	CallerSkipFrame int
	Caller          bool

	// TraceLogAllQueries specifies whether or not all queries should be logged
	// at the TRACE level.
	TraceLogAllQueries bool
}

func ZeroLogger(log zerolog.Logger, cfg ...ZeroLogSettings) DatabaseLogger {
	return ZeroLoggerPtr(&log, cfg...)
}

func ZeroLoggerPtr(log *zerolog.Logger, cfg ...ZeroLogSettings) DatabaseLogger {
	wrapped := &zeroLogger{l: log}
	if len(cfg) > 0 {
		wrapped.ZeroLogSettings = cfg[0]
	} else {
		wrapped.ZeroLogSettings = ZeroLogSettings{
			CallerSkipFrame: 2, // Skip LoggingExecable.ExecContext and zeroLogger.QueryTiming
			Caller:          true,
		}
	}
	return wrapped
}

func (z zeroLogger) WarnUnsupportedVersion(current, compat, latest int) {
	z.l.Warn().
		Int("current_version", current).
		Int("oldest_compatible_version", compat).
		Int("latest_known_version", latest).
		Msg("Unsupported database schema version, continuing anyway")
}

func (z zeroLogger) PrepareUpgrade(current, compat, latest int) {
	evt := z.l.Info().
		Int("current_version", current).
		Int("oldest_compatible_version", compat).
		Int("latest_known_version", latest)
	if current >= latest {
		evt.Msg("Database is up to date")
	} else {
		evt.Msg("Preparing to update database schema")
	}
}

func (z zeroLogger) DoUpgrade(from, to int, message string, txn TxnMode) {
	z.l.Info().
		Int("from", from).
		Int("to", to).
		Str("txn_mode", string(txn)).
		Str("description", message).
		Msg("Upgrading database")
}

var whitespaceRegex = regexp.MustCompile(`\s+`)

var GlobalSafeQueryLog bool

func (z zeroLogger) QueryTiming(ctx context.Context, method, query string, args []any, nrows int, duration time.Duration, err error) {
	log := zerolog.Ctx(ctx)
	if log.GetLevel() == zerolog.Disabled || log == zerolog.DefaultContextLogger {
		log = z.l
	}
	if (!z.TraceLogAllQueries || log.GetLevel() != zerolog.TraceLevel) && !GlobalSafeQueryLog && duration < 1*time.Second {
		return
	}
	if nrows > -1 {
		rowLog := log.With().Int("rows", nrows).Logger()
		log = &rowLog
	}
	query = strings.TrimSpace(whitespaceRegex.ReplaceAllLiteralString(query, " "))
	callerSkipFrame := z.CallerSkipFrame
	if GlobalSafeQueryLog || duration > 1*time.Second {
		for ; callerSkipFrame < 10; callerSkipFrame++ {
			_, filename, _, _ := runtime.Caller(callerSkipFrame)
			if !strings.Contains(filename, "/dbutil/") {
				break
			}
		}
	}
	if GlobalSafeQueryLog {
		log.Debug().
			Err(err).
			Int64("duration_µs", duration.Microseconds()).
			Str("method", method).
			Str("query", query).
			Caller(callerSkipFrame).
			Msg("Query")
	} else {
		log.Trace().
			Err(err).
			Int64("duration_µs", duration.Microseconds()).
			Str("method", method).
			Str("query", query).
			Interface("query_args", args).
			Msg("Query")
	}
	if duration >= 1*time.Second {
		evt := log.Warn().
			Float64("duration_seconds", duration.Seconds()).
			AnErr("result_error", err).
			Str("method", method).
			Str("query", query)
		if z.Caller {
			evt = evt.Caller(callerSkipFrame)
		}
		evt.Msg("Query took long")
	}
}

func (z zeroLogger) Warn(msg string, args ...any) {
	z.l.Warn().Msgf(msg, args...) // zerolog-allow-msgf
}
