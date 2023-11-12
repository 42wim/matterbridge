package mailserver

import (
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/log"
)

const (
	dbCleanerBatchSize = 1000
	dbCleanerPeriod    = time.Hour
)

// dbCleaner removes old messages from a db.
type dbCleaner struct {
	sync.RWMutex

	db        DB
	batchSize int
	retention time.Duration

	period time.Duration
	cancel chan struct{}
}

// newDBCleaner returns a new cleaner for db.
func newDBCleaner(db DB, retention time.Duration) *dbCleaner {
	return &dbCleaner{
		db:        db,
		retention: retention,

		batchSize: dbCleanerBatchSize,
		period:    dbCleanerPeriod,
	}
}

// Start starts a loop that cleans up old messages.
func (c *dbCleaner) Start() {
	log.Info("Starting cleaning envelopes", "period", c.period, "retention", c.retention)

	cancel := make(chan struct{})

	c.Lock()
	c.cancel = cancel
	c.Unlock()

	go c.schedule(c.period, cancel)
}

// Stops stops the cleaning loop.
func (c *dbCleaner) Stop() {
	c.Lock()
	defer c.Unlock()

	if c.cancel == nil {
		return
	}
	close(c.cancel)
	c.cancel = nil
}

func (c *dbCleaner) schedule(period time.Duration, cancel <-chan struct{}) {
	t := time.NewTicker(period)
	defer t.Stop()

	for {
		select {
		case <-t.C:
			count, err := c.PruneEntriesOlderThan(time.Now().Add(-c.retention))
			if err != nil {
				log.Error("failed to prune data", "err", err)
			}
			log.Info("Prunned some some messages successfully", "count", count)
		case <-cancel:
			return
		}
	}
}

// PruneEntriesOlderThan removes messages sent between lower and upper timestamps
// and returns how many have been removed.
func (c *dbCleaner) PruneEntriesOlderThan(t time.Time) (int, error) {
	return c.db.Prune(t, c.batchSize)
}
