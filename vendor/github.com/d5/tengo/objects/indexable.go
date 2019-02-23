package objects

// Indexable is an object that can take an index and return an object.
type Indexable interface {
	// IndexGet should take an index Object and return a result Object or an error.
	// If error is returned, the runtime will treat it as a run-time error and ignore returned value.
	// If nil is returned as value, it will be converted to Undefined value by the runtime.
	IndexGet(index Object) (value Object, err error)
}
