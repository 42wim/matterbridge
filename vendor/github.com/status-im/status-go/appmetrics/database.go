package appmetrics

import (
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/xeipuuv/gojsonschema"
)

type AppMetricEventType string

// Value is `json.RawMessage` so we can send any json shape, including strings
// Validation is handled using JSON schemas defined in validators.go, instead of Golang structs
type AppMetric struct {
	ID         int                `json:"-"`
	MessageID  string             `json:"message_id"`
	Event      AppMetricEventType `json:"event"`
	Value      json.RawMessage    `json:"value"`
	AppVersion string             `json:"app_version"`
	OS         string             `json:"os"`
	SessionID  string             `json:"session_id"`
	CreatedAt  time.Time          `json:"created_at"`
	Processed  bool               `json:"processed"`
	ReceivedAt time.Time          `json:"received_at"`
}

type AppMetricValidationError struct {
	Metric AppMetric
	Errors []gojsonschema.ResultError
}

type Page struct {
	AppMetrics []AppMetric
	TotalCount int
}

const (
	// status-mobile navigation events
	NavigateTo         AppMetricEventType = "navigate-to"
	ScreensOnWillFocus AppMetricEventType = "screens/on-will-focus"
)

// EventSchemaMap Every event should have a schema attached
var EventSchemaMap = map[AppMetricEventType]interface{}{
	NavigateTo:         NavigateToCofxSchema,
	ScreensOnWillFocus: NavigateToCofxSchema,
}

func NewDB(db *sql.DB) *Database {
	return &Database{db: db}
}

// Database sql wrapper for operations with browser objects.
type Database struct {
	db *sql.DB
}

// Close closes database.
func (db Database) Close() error {
	return db.db.Close()
}

func jsonschemaErrorsToError(validationErrors []AppMetricValidationError) error {
	var fieldErrors []string

	for _, appMetricValidationError := range validationErrors {
		metric := appMetricValidationError.Metric
		errors := appMetricValidationError.Errors

		var errorDesc string = "Error in event: " + string(metric.Event) + " - "
		for _, e := range errors {
			errorDesc = errorDesc + "value." + e.Context().String() + ":" + e.Description()
		}
		fieldErrors = append(fieldErrors, errorDesc)
	}

	return errors.New(strings.Join(fieldErrors[:], "/ "))
}

func (db *Database) ValidateAppMetrics(appMetrics []AppMetric) (err error) {
	var calculatedErrors []AppMetricValidationError
	for _, metric := range appMetrics {
		schema := EventSchemaMap[metric.Event]

		if schema == nil {
			return errors.New("No schema defined for: " + string(metric.Event))
		}

		schemaLoader := gojsonschema.NewGoLoader(schema)
		valLoader := gojsonschema.NewStringLoader(string(metric.Value))
		res, err := gojsonschema.Validate(schemaLoader, valLoader)

		if err != nil {
			return err
		}

		// validate all metrics and save errors
		if !res.Valid() {
			calculatedErrors = append(calculatedErrors, AppMetricValidationError{metric, res.Errors()})
		}
	}

	if len(calculatedErrors) > 0 {
		return jsonschemaErrorsToError(calculatedErrors)
	}
	return
}

func (db *Database) SaveAppMetrics(appMetrics []AppMetric, sessionID string) (err error) {
	var (
		tx     *sql.Tx
		insert *sql.Stmt
	)

	// make sure that the shape of the metric is same as expected
	err = db.ValidateAppMetrics(appMetrics)
	if err != nil {
		return err
	}

	// start txn
	tx, err = db.db.Begin()
	if err != nil {
		return err
	}

	defer func() {
		if err == nil {
			err = tx.Commit()
			return
		}
		_ = tx.Rollback()
	}()

	insert, err = tx.Prepare("INSERT INTO app_metrics (event, value, app_version, operating_system, session_id, processed) VALUES (?, ?, ?, ?, ?, ?)")
	if err != nil {
		return err
	}

	for _, metric := range appMetrics {
		_, err = insert.Exec(metric.Event, metric.Value, metric.AppVersion, metric.OS, sessionID, metric.Processed)
		if err != nil {
			return
		}
	}
	return
}

func (db *Database) GetAppMetrics(limit int, offset int) (page Page, err error) {
	countErr := db.db.QueryRow("SELECT count(*) FROM app_metrics").Scan(&page.TotalCount)
	if countErr != nil {
		return page, countErr
	}

	rows, err := db.db.Query("SELECT id, event, value, app_version, operating_system, session_id, created_at, processed FROM app_metrics LIMIT ? OFFSET ?", limit, offset)
	if err != nil {
		return page, err
	}
	defer rows.Close()

	page.AppMetrics, err = db.getFromRows(rows)
	return page, err
}

func (db *Database) getFromRows(rows *sql.Rows) (appMetrics []AppMetric, err error) {
	var metrics []AppMetric

	for rows.Next() {
		metric := AppMetric{}
		err = rows.Scan(
			&metric.ID,
			&metric.Event,
			&metric.Value,
			&metric.AppVersion,
			&metric.OS,
			&metric.SessionID,
			&metric.CreatedAt,
			&metric.Processed,
		)
		if err != nil {
			return metrics, err
		}
		metrics = append(metrics, metric)
	}
	return metrics, nil
}

func (db *Database) GetUnprocessed() ([]AppMetric, error) {
	rows, err := db.db.Query("SELECT id, event, value, app_version, operating_system, session_id, created_at, processed FROM app_metrics WHERE processed IS ? ORDER BY session_id ASC, created_at ASC", false)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return db.getFromRows(rows)
}

func (db *Database) GetUnprocessedGroupedBySession() (map[string][]AppMetric, error) {
	uam, err := db.GetUnprocessed()
	if err != nil {
		return nil, err
	}

	out := map[string][]AppMetric{}
	for _, am := range uam {
		out[am.SessionID] = append(out[am.SessionID], am)
	}

	return out, nil
}

func (db *Database) SetToProcessedByIDs(ids []int) (err error) {
	var (
		tx     *sql.Tx
		update *sql.Stmt
	)

	// start txn
	tx, err = db.db.Begin()
	if err != nil {
		return err
	}

	defer func() {
		if err == nil {
			err = tx.Commit()
			return
		}
		_ = tx.Rollback()
	}()

	// Generate prepared statement IN list
	in := "("
	for i := 0; i < len(ids); i++ {
		in += "?,"
	}
	in = in[:len(in)-1] + ")"

	update, err = tx.Prepare("UPDATE app_metrics SET processed = 1 WHERE id IN " + in) // nolint: gosec
	if err != nil {
		return err
	}

	// Convert the ids into Stmt.Exec compatible variadic
	args := make([]interface{}, 0, len(ids))
	for _, id := range ids {
		args = append(args, id)
	}

	_, err = update.Exec(args...)
	if err != nil {
		return
	}
	return
}

func (db *Database) SetToProcessed(appMetrics []AppMetric) (err error) {
	ids := GetAppMetricsIDs(appMetrics)
	return db.SetToProcessedByIDs(ids)
}

func (db *Database) GetMessagesOlderThan(date *time.Time) ([]AppMetric, error) {
	rows, err := db.db.Query("SELECT id, event, value, app_version, operating_system, session_id, created_at, processed FROM app_metrics WHERE created_at < ?", date)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return db.getFromRows(rows)
}

func (db *Database) DeleteOlderThan(date *time.Time) (err error) {
	var (
		tx *sql.Tx
		d  *sql.Stmt
	)

	// start txn
	tx, err = db.db.Begin()
	if err != nil {
		return err
	}

	defer func() {
		if err == nil {
			err = tx.Commit()
			return
		}
		_ = tx.Rollback()
	}()

	d, err = tx.Prepare("DELETE FROM app_metrics WHERE created_at < ?")
	if err != nil {
		return err
	}

	_, err = d.Exec(date)
	if err != nil {
		return
	}
	return
}

func GetAppMetricsIDs(appMetrics []AppMetric) []int {
	var ids []int

	for _, am := range appMetrics {
		ids = append(ids, am.ID)
	}

	return ids
}
