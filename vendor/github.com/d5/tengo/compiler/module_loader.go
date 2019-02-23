package compiler

// ModuleLoader should take a module name and return the module data.
type ModuleLoader func(moduleName string) ([]byte, error)
