package ext

import (
	"context"
	"time"

	"github.com/status-im/status-go/db"
)

// ContextKey is a type used for keys in ext Context.
type ContextKey struct {
	Name string
}

// NewContextKey returns new ContextKey instance.
func NewContextKey(name string) ContextKey {
	return ContextKey{Name: name}
}

var (
	historyDBKey       = NewContextKey("history_db")
	requestRegistryKey = NewContextKey("request_registry")
	timeKey            = NewContextKey("time")
)

// NewContext creates Context with all required fields.
func NewContext(ctx context.Context, source TimeSource, registry *RequestsRegistry, storage db.Storage) Context {
	ctx = context.WithValue(ctx, historyDBKey, db.NewHistoryStore(storage))
	ctx = context.WithValue(ctx, timeKey, source)
	ctx = context.WithValue(ctx, requestRegistryKey, registry)
	return Context{ctx}
}

// TimeSource is a type used for current time.
type TimeSource func() time.Time

// Context provides access to request-scoped values.
type Context struct {
	context.Context
}

// HistoryStore returns db.HistoryStore instance associated with this request.
func (c Context) HistoryStore() db.HistoryStore {
	return c.Value(historyDBKey).(db.HistoryStore)
}

// Time returns current time using time function associated with this request.
func (c Context) Time() time.Time {
	return c.Value(timeKey).(TimeSource)()
}

// RequestRegistry returns RequestRegistry that tracks each request life-span.
func (c Context) RequestRegistry() *RequestsRegistry {
	return c.Value(requestRegistryKey).(*RequestsRegistry)
}
