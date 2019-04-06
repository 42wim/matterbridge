package objects

// BuiltinModule is an importable module that's written in Go.
type BuiltinModule struct {
	Attrs map[string]Object
}

// Import returns an immutable map for the module.
func (m *BuiltinModule) Import(moduleName string) (interface{}, error) {
	return m.AsImmutableMap(moduleName), nil
}

// AsImmutableMap converts builtin module into an immutable map.
func (m *BuiltinModule) AsImmutableMap(moduleName string) *ImmutableMap {
	attrs := make(map[string]Object, len(m.Attrs))
	for k, v := range m.Attrs {
		attrs[k] = v.Copy()
	}

	attrs["__module_name__"] = &String{Value: moduleName}

	return &ImmutableMap{Value: attrs}
}
