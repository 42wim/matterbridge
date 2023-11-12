package mailserver

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/lib/pq"

	// Import postgres driver
	_ "github.com/lib/pq"
	"github.com/status-im/migrate/v4"
	"github.com/status-im/migrate/v4/database/postgres"
	bindata "github.com/status-im/migrate/v4/source/go_bindata"

	"github.com/status-im/status-go/mailserver/migrations"

	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rlp"

	"github.com/status-im/status-go/eth-node/types"
	waku "github.com/status-im/status-go/waku/common"
)

type PostgresDB struct {
	db   *sql.DB
	name string
	done chan struct{}
}

func NewPostgresDB(uri string) (*PostgresDB, error) {
	db, err := sql.Open("postgres", uri)
	if err != nil {
		return nil, err
	}

	instance := &PostgresDB{
		db:   db,
		done: make(chan struct{}),
	}
	if err := instance.setup(); err != nil {
		return nil, err
	}

	// name is used for metrics labels
	if name, err := instance.getDBName(uri); err == nil {
		instance.name = name
	}

	// initialize the metric value
	instance.updateArchivedEnvelopesCount()
	// checking count on every insert is inefficient
	go func() {
		for {
			select {
			case <-instance.done:
				return
			case <-time.After(time.Second * envelopeCountCheckInterval):
				instance.updateArchivedEnvelopesCount()
			}
		}
	}()
	return instance, nil
}

type postgresIterator struct {
	*sql.Rows
}

func (i *PostgresDB) getDBName(uri string) (string, error) {
	query := "SELECT current_database()"
	var dbName string
	return dbName, i.db.QueryRow(query).Scan(&dbName)
}

func (i *PostgresDB) envelopesCount() (int, error) {
	query := "SELECT count(*) FROM envelopes"
	var count int
	return count, i.db.QueryRow(query).Scan(&count)
}

func (i *PostgresDB) updateArchivedEnvelopesCount() {
	if count, err := i.envelopesCount(); err != nil {
		log.Warn("db query for envelopes count failed", "err", err)
	} else {
		archivedEnvelopesGauge.WithLabelValues(i.name).Set(float64(count))
	}
}

func (i *postgresIterator) DBKey() (*DBKey, error) {
	var value []byte
	var id []byte
	if err := i.Scan(&id, &value); err != nil {
		return nil, err
	}
	return &DBKey{raw: id}, nil
}

func (i *postgresIterator) Error() error {
	return i.Err()
}

func (i *postgresIterator) Release() error {
	return i.Close()
}

func (i *postgresIterator) GetEnvelopeByBloomFilter(bloom []byte) ([]byte, error) {
	var value []byte
	var id []byte
	if err := i.Scan(&id, &value); err != nil {
		return nil, err
	}

	return value, nil
}

func (i *postgresIterator) GetEnvelopeByTopicsMap(topics map[types.TopicType]bool) ([]byte, error) {
	var value []byte
	var id []byte
	if err := i.Scan(&id, &value); err != nil {
		return nil, err
	}

	return value, nil
}

func (i *PostgresDB) BuildIterator(query CursorQuery) (Iterator, error) {
	var args []interface{}

	stmtString := "SELECT id, data FROM envelopes"

	var historyRange string
	if len(query.cursor) > 0 {
		args = append(args, query.start, query.cursor)
		// If we have a cursor, we don't want to include that envelope in the result set
		stmtString += " " + "WHERE id >= $1 AND id < $2"
		historyRange = "partial" //nolint: goconst
	} else {
		args = append(args, query.start, query.end)
		stmtString += " " + "WHERE id >= $1 AND id <= $2"
		historyRange = "full" //nolint: goconst
	}

	var filterRange string
	if len(query.topics) > 0 {
		args = append(args, pq.Array(query.topics))
		stmtString += " " + "AND topic = any($3)"
		filterRange = "partial" //nolint: goconst
	} else {
		stmtString += " " + fmt.Sprintf("AND bloom & b'%s'::bit(512) = bloom", toBitString(query.bloom))
		filterRange = "full" //nolint: goconst
	}

	// Positional argument depends on the fact whether the query uses topics or bloom filter.
	// If topic is used, the list of topics is passed as an argument to the query.
	// If bloom filter is used, it is included into the query statement.
	args = append(args, query.limit)
	stmtString += " " + fmt.Sprintf("ORDER BY ID DESC LIMIT $%d", len(args))

	stmt, err := i.db.Prepare(stmtString)
	if err != nil {
		return nil, err
	}

	envelopeQueriesCounter.WithLabelValues(filterRange, historyRange).Inc()
	rows, err := stmt.Query(args...)
	if err != nil {
		return nil, err
	}

	return &postgresIterator{rows}, nil
}

func (i *PostgresDB) setup() error {
	resources := bindata.Resource(
		migrations.AssetNames(),
		migrations.Asset,
	)

	source, err := bindata.WithInstance(resources)
	if err != nil {
		return err
	}

	driver, err := postgres.WithInstance(i.db, &postgres.Config{})
	if err != nil {
		return err
	}

	m, err := migrate.NewWithInstance(
		"go-bindata",
		source,
		"postgres",
		driver)
	if err != nil {
		return err
	}

	if err = m.Up(); err != migrate.ErrNoChange {
		return err
	}

	return nil
}

func (i *PostgresDB) Close() error {
	select {
	case <-i.done:
	default:
		close(i.done)
	}
	return i.db.Close()
}

func (i *PostgresDB) GetEnvelope(key *DBKey) ([]byte, error) {
	statement := `SELECT data FROM envelopes WHERE id = $1`

	stmt, err := i.db.Prepare(statement)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	var envelope []byte

	if err = stmt.QueryRow(key.Bytes()).Scan(&envelope); err != nil {
		return nil, err
	}

	return envelope, nil
}

func (i *PostgresDB) Prune(t time.Time, batch int) (int, error) {
	var zero types.Hash
	var emptyTopic types.TopicType
	kl := NewDBKey(0, emptyTopic, zero)
	ku := NewDBKey(uint32(t.Unix()), emptyTopic, zero)
	statement := "DELETE FROM envelopes WHERE id BETWEEN $1 AND $2"

	stmt, err := i.db.Prepare(statement)
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	result, err := stmt.Exec(kl.Bytes(), ku.Bytes())
	if err != nil {
		return 0, err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}
	return int(rows), nil
}

func (i *PostgresDB) SaveEnvelope(env types.Envelope) error {
	topic := env.Topic()
	key := NewDBKey(env.Expiry()-env.TTL(), topic, env.Hash())
	rawEnvelope, err := rlp.EncodeToBytes(env.Unwrap())
	if err != nil {
		log.Error(fmt.Sprintf("rlp.EncodeToBytes failed: %s", err))
		archivedErrorsCounter.WithLabelValues(i.name).Inc()
		return err
	}
	if rawEnvelope == nil {
		archivedErrorsCounter.WithLabelValues(i.name).Inc()
		return errors.New("failed to encode envelope to bytes")
	}

	statement := "INSERT INTO envelopes (id, data, topic, bloom) VALUES ($1, $2, $3, B'"
	statement += toBitString(env.Bloom())
	statement += "'::bit(512)) ON CONFLICT (id) DO NOTHING;"
	stmt, err := i.db.Prepare(statement)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(
		key.Bytes(),
		rawEnvelope,
		topicToByte(topic),
	)

	if err != nil {
		archivedErrorsCounter.WithLabelValues(i.name).Inc()
		return err
	}

	archivedEnvelopesGauge.WithLabelValues(i.name).Inc()
	archivedEnvelopeSizeMeter.WithLabelValues(i.name).Observe(
		float64(waku.EnvelopeHeaderLength + env.Size()))

	return nil
}

func topicToByte(t types.TopicType) []byte {
	return []byte{t[0], t[1], t[2], t[3]}
}

func toBitString(bloom []byte) string {
	val := ""
	for _, n := range bloom {
		val += fmt.Sprintf("%08b", n)
	}
	return val
}
