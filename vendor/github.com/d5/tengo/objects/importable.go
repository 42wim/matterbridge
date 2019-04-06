package objects

// Importable interface represents importable module instance.
type Importable interface {
	// Import should return either an Object or module source code ([]byte).
	Import(moduleName string) (interface{}, error)
}
