package peersyncing

import (
	"database/sql"

	"github.com/status-im/status-go/protocol/common"
)

type Config struct {
	SyncMessagePersistence SyncMessagePersistence
	Database               *sql.DB
	Timesource             common.TimeSource
}
