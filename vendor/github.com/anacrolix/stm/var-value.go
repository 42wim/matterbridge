package stm

type VarValue interface {
	Set(interface{}) VarValue
	Get() interface{}
	Changed(VarValue) bool
}

type version uint64

type versionedValue struct {
	value   interface{}
	version version
}

func (me versionedValue) Set(newValue interface{}) VarValue {
	return versionedValue{
		value:   newValue,
		version: me.version + 1,
	}
}

func (me versionedValue) Get() interface{} {
	return me.value
}

func (me versionedValue) Changed(other VarValue) bool {
	return me.version != other.(versionedValue).version
}

type customVarValue struct {
	value   interface{}
	changed func(interface{}, interface{}) bool
}

var _ VarValue = customVarValue{}

func (me customVarValue) Changed(other VarValue) bool {
	return me.changed(me.value, other.(customVarValue).value)
}

func (me customVarValue) Set(newValue interface{}) VarValue {
	return customVarValue{
		value:   newValue,
		changed: me.changed,
	}
}

func (me customVarValue) Get() interface{} {
	return me.value
}
