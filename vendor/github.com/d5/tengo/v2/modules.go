package tengo

// Importable interface represents importable module instance.
type Importable interface {
	// Import should return either an Object or module source code ([]byte).
	Import(moduleName string) (interface{}, error)
}

// ModuleGetter enables implementing dynamic module loading.
type ModuleGetter interface {
	Get(name string) Importable
}

// ModuleMap represents a set of named modules. Use NewModuleMap to create a
// new module map.
type ModuleMap struct {
	m map[string]Importable
}

// NewModuleMap creates a new module map.
func NewModuleMap() *ModuleMap {
	return &ModuleMap{
		m: make(map[string]Importable),
	}
}

// Add adds an import module.
func (m *ModuleMap) Add(name string, module Importable) {
	m.m[name] = module
}

// AddBuiltinModule adds a builtin module.
func (m *ModuleMap) AddBuiltinModule(name string, attrs map[string]Object) {
	m.m[name] = &BuiltinModule{Attrs: attrs}
}

// AddSourceModule adds a source module.
func (m *ModuleMap) AddSourceModule(name string, src []byte) {
	m.m[name] = &SourceModule{Src: src}
}

// Remove removes a named module.
func (m *ModuleMap) Remove(name string) {
	delete(m.m, name)
}

// Get returns an import module identified by name. It returns if the name is
// not found.
func (m *ModuleMap) Get(name string) Importable {
	return m.m[name]
}

// GetBuiltinModule returns a builtin module identified by name. It returns
// if the name is not found or the module is not a builtin module.
func (m *ModuleMap) GetBuiltinModule(name string) *BuiltinModule {
	mod, _ := m.m[name].(*BuiltinModule)
	return mod
}

// GetSourceModule returns a source module identified by name. It returns if
// the name is not found or the module is not a source module.
func (m *ModuleMap) GetSourceModule(name string) *SourceModule {
	mod, _ := m.m[name].(*SourceModule)
	return mod
}

// Copy creates a copy of the module map.
func (m *ModuleMap) Copy() *ModuleMap {
	c := &ModuleMap{
		m: make(map[string]Importable),
	}
	for name, mod := range m.m {
		c.m[name] = mod
	}
	return c
}

// Len returns the number of named modules.
func (m *ModuleMap) Len() int {
	return len(m.m)
}

// AddMap adds named modules from another module map.
func (m *ModuleMap) AddMap(o *ModuleMap) {
	for name, mod := range o.m {
		m.m[name] = mod
	}
}

// SourceModule is an importable module that's written in Tengo.
type SourceModule struct {
	Src []byte
}

// Import returns a module source code.
func (m *SourceModule) Import(_ string) (interface{}, error) {
	return m.Src, nil
}
