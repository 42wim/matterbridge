package persistenceutil

import (
	"github.com/pkg/errors"
	"strings"

	nodemigrations "github.com/status-im/mvds/node/migrations"
	peersmigrations "github.com/status-im/mvds/peers/migrations"
	statemigrations "github.com/status-im/mvds/state/migrations"
	storemigrations "github.com/status-im/mvds/store/migrations"
)

type getter func(string) ([]byte, error)

type Migration struct {
	Names  []string
	Getter func(name string) ([]byte, error)
}

func prepareMigrations(migrations []Migration) ([]string, getter, error) {
	var allNames []string
	nameToGetter := make(map[string]getter)

	for _, m := range migrations {
		for _, name := range m.Names {
			if !validateName(name) {
				continue
			}

			if _, ok := nameToGetter[name]; ok {
				return nil, nil, errors.Errorf("migration with name %s already exists", name)
			}
			allNames = append(allNames, name)
			nameToGetter[name] = m.Getter
		}
	}

	return allNames, func(name string) ([]byte, error) {
		getter, ok := nameToGetter[name]
		if !ok {
			return nil, errors.Errorf("no migration for name %s", name)
		}
		return getter(name)
	}, nil
}

// DefaultMigrations is a collection of all mvds components migrations.
var DefaultMigrations = []Migration{
	{
		Names:  nodemigrations.AssetNames(),
		Getter: nodemigrations.Asset,
	},
	{
		Names:  peersmigrations.AssetNames(),
		Getter: peersmigrations.Asset,
	},
	{
		Names:  statemigrations.AssetNames(),
		Getter: statemigrations.Asset,
	},
	{
		Names:  storemigrations.AssetNames(),
		Getter: storemigrations.Asset,
	},
}

// validateName verifies that only *.sql files are taken into consideration.
func validateName(name string) bool {
	return strings.HasSuffix(name, ".sql")
}
