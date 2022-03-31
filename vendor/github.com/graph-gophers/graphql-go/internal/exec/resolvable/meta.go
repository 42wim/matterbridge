package resolvable

import (
	"reflect"

	"github.com/graph-gophers/graphql-go/introspection"
	"github.com/graph-gophers/graphql-go/types"
)

// Meta defines the details of the metadata schema for introspection.
type Meta struct {
	FieldSchema   Field
	FieldType     Field
	FieldTypename Field
	Schema        *Object
	Type          *Object
}

func newMeta(s *types.Schema) *Meta {
	var err error
	b := newBuilder(s)

	metaSchema := s.Types["__Schema"].(*types.ObjectTypeDefinition)
	so, err := b.makeObjectExec(metaSchema.Name, metaSchema.Fields, nil, false, reflect.TypeOf(&introspection.Schema{}))
	if err != nil {
		panic(err)
	}

	metaType := s.Types["__Type"].(*types.ObjectTypeDefinition)
	t, err := b.makeObjectExec(metaType.Name, metaType.Fields, nil, false, reflect.TypeOf(&introspection.Type{}))
	if err != nil {
		panic(err)
	}

	if err := b.finish(); err != nil {
		panic(err)
	}

	fieldTypename := Field{
		FieldDefinition: types.FieldDefinition{
			Name: "__typename",
			Type: &types.NonNull{OfType: s.Types["String"]},
		},
		TraceLabel: "GraphQL field: __typename",
	}

	fieldSchema := Field{
		FieldDefinition: types.FieldDefinition{
			Name: "__schema",
			Type: s.Types["__Schema"],
		},
		TraceLabel: "GraphQL field: __schema",
	}

	fieldType := Field{
		FieldDefinition: types.FieldDefinition{
			Name: "__type",
			Type: s.Types["__Type"],
		},
		TraceLabel: "GraphQL field: __type",
	}

	return &Meta{
		FieldSchema:   fieldSchema,
		FieldTypename: fieldTypename,
		FieldType:     fieldType,
		Schema:        so,
		Type:          t,
	}
}
