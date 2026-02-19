// Package dbutil provides a simple framework for in-process database
// migrations. You provide the SQL files and they are run to upgrade
// the database. A versions table is automatically created in the
// database to track which migrations have been applied. There is
// support for multiple migration pathways, for example v0->v2 versus
// v0->v1->v2, and the shorter one is prioritized if both are
// provided.
//
// Example usage from Go:
//
//	package main
//
//	import (
//		"context"
//		"database/sql"
//		"embed"
//
//		"go.mau.fi/util/dbutil"
//	)
//
//	//go:embed *.sql
//	var upgrades embed.FS
//
//	func mainE() error {
//		ctx := context.Background()
//		rawDB, err := sql.Open("sqlite3", "./hotdogs.db")
//		if err != nil {
//			return err
//		}
//		db, err := dbutil.NewWithDB(rawDB, "sqlite3")
//		if err != nil {
//			return err
//		}
//		table := dbutil.UpgradeTable{}
//		table.RegisterFS(upgrades)
//		err = db.Upgrade(ctx)
//		if err != nil {
//			return err
//		}
//		// db has been upgraded to latest version
//		return nil
//	}
//
// In dbutil, the database is understood to have a monotonic integer
// sequence of versions starting at v0, v1, v2, etc. By providing
// migrations you define a directed acyclic graph (DAG) that allows
// dbutil to find a path from the current recorded database version to
// the latest version available.
//
// Each SQL migration file has a mandatory comment header that
// identifies which database versions it upgrades between. For example
// this is a migration that upgrades from v0 to v2:
//
//	-- v0 -> v2: Do some things
//
// You can omit the first version for the common case of upgrading to
// a version from the previous version. For example this is a
// migration that upgrades from v1 to v2:
//
//	-- v2: Do fewer things
//
// By providing "v1" and "v2" migrations, a v0 database would be
// upgraded to v1 and then v2, while by providing an additional "v0 ->
// v2" migration a v0 database would be upgraded directly to v2 as it
// is a more direct path. With that migration provided the "v1"
// migration is no longer needed.
//
// By default, when running migrations, if a more recent database
// version is live than the current code knows about (for example,
// from running a previous version of the application), dbutil will
// error out. However, many database migrations are backwards
// compatible. You can therefore indicate this when writing a
// migration, and previous versions of the application will accept a
// database with that migration applied, even if they are unaware of
// its contents. For example, if the migration from v1 to v2 was
// backwards compatible, you could provide this migration:
//
//	-- v2 (compatible with v1+): Do fewer things
//
// When applying the migration, the compatibility level (v1) is saved
// to the versions table in the database, so that older versions of
// the application which only know about v1 will see that v2 of the
// database is still OK to use. If the compatibility level is not set,
// then it defaults to the same as the target version for the
// migration, which achieves the default behavior described in the
// previous paragraph.
//
// You can provide additional flags immediately following the header
// line. To disable wrapping the upgrade in a single transaction, put
// "transaction: off" on the second line.
//
//	-- v5: Upgrade without transaction
//	-- transaction: off
//	// do dangerous stuff
//
// Within migrations, there is special syntax that can be used to
// filter parts of the SQL to apply only with specific dialects. To
// limit the next line to one dialect:
//
//	-- only: postgres
//
// To limit the next N lines:
//
//	-- only: sqlite for next 123 lines
//
// To limit a block of code, fenced by another directive:
//
//	-- only: sqlite until "end only"
//	QUERY;
//	ANOTHER QUERY;
//	-- end only sqlite
//
// If the single-line limit is on the second line of the file, the
// whole file is limited to that dialect.
//
// If the filter ends with `(lines commented)`, then ALL lines chosen
// by the filter will be uncommented. The `--` comment prefix must be
// at the beginning of the line with no whitespace ahead of it.
package dbutil
