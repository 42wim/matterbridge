package browsers

import (
	"context"
	"database/sql"

	"github.com/mat/besticon/besticon"

	"github.com/ethereum/go-ethereum/log"
)

// Database sql wrapper for operations with browser objects.
type Database struct {
	db *sql.DB
}

// Close closes database.
func (db Database) Close() error {
	return db.db.Close()
}

func NewDB(db *sql.DB) *Database {
	return &Database{db: db}
}

type BookmarksType string

type Bookmark struct {
	URL       string `json:"url"`
	Name      string `json:"name"`
	ImageURL  string `json:"imageUrl"`
	Removed   bool   `json:"removed"`
	Clock     uint64 `json:"-"` //used to sync
	DeletedAt uint64 `json:"deletedAt,omitempty"`
}
type Browser struct {
	ID           string   `json:"browser-id"`
	Name         string   `json:"name"`
	Timestamp    uint64   `json:"timestamp"`
	Dapp         bool     `json:"dapp?"`
	HistoryIndex int      `json:"history-index"`
	History      []string `json:"history,omitempty"`
}

func (db *Database) GetBookmarks() ([]*Bookmark, error) {
	rows, err := db.db.Query(`SELECT url, name, image_url, removed, deleted_at FROM bookmarks`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rst []*Bookmark
	for rows.Next() {
		bookmark := &Bookmark{}
		err := rows.Scan(&bookmark.URL, &bookmark.Name, &bookmark.ImageURL, &bookmark.Removed, &bookmark.DeletedAt)
		if err != nil {
			return nil, err
		}

		rst = append(rst, bookmark)
	}

	return rst, nil
}

func (db *Database) StoreBookmark(bookmark Bookmark) (Bookmark, error) {
	insert, err := db.db.Prepare("INSERT OR REPLACE INTO bookmarks (url, name, image_url, removed, clock, deleted_at) VALUES (?, ?, ?, ?, ?, ?)")

	if err != nil {
		return bookmark, err
	}

	// Get the right icon
	finder := besticon.IconFinder{}
	icons, iconError := finder.FetchIcons(bookmark.URL)

	if iconError == nil && len(icons) > 0 {
		icon := finder.IconInSizeRange(besticon.SizeRange{Min: 48, Perfect: 48, Max: 100})
		if icon != nil {
			bookmark.ImageURL = icon.URL
		} else {
			bookmark.ImageURL = icons[0].URL
		}
	} else {
		log.Error("error getting the bookmark icon", "iconError", iconError)
	}

	_, err = insert.Exec(bookmark.URL, bookmark.Name, bookmark.ImageURL, bookmark.Removed, bookmark.Clock, bookmark.DeletedAt)
	return bookmark, err
}

func (db *Database) StoreBookmarkWithoutFetchIcon(bookmark *Bookmark, tx *sql.Tx) (err error) {
	if tx == nil {
		tx, err = db.db.BeginTx(context.Background(), &sql.TxOptions{})
		if err != nil {
			return err
		}
		defer func() {
			if err == nil {
				err = tx.Commit()
				return
			}
			// don't shadow original error
			_ = tx.Rollback()
		}()
	}

	insert, err := tx.Prepare("INSERT OR REPLACE INTO bookmarks (url, name, image_url, removed, clock, deleted_at) VALUES (?, ?, ?, ?, ?, ?)")

	if err != nil {
		return err
	}

	defer insert.Close()

	_, err = insert.Exec(bookmark.URL, bookmark.Name, bookmark.ImageURL, bookmark.Removed, bookmark.Clock, bookmark.DeletedAt)
	return err
}

func (db *Database) StoreSyncBookmarks(bookmarks []*Bookmark) ([]*Bookmark, error) {
	tx, err := db.db.BeginTx(context.Background(), &sql.TxOptions{})
	if err != nil {
		return nil, err
	}

	defer func() {
		if err == nil {
			err = tx.Commit()
			return
		}
		// don't shadow original error
		_ = tx.Rollback()
	}()

	var storedBookmarks []*Bookmark
	for _, bookmark := range bookmarks {
		shouldSync, err := db.shouldSyncBookmark(bookmark, tx)
		if err != nil {
			return storedBookmarks, err
		}
		if shouldSync {
			err := db.StoreBookmarkWithoutFetchIcon(bookmark, tx)
			if err != nil {
				return storedBookmarks, err
			}
			storedBookmarks = append(storedBookmarks, bookmark)
		}
	}
	return storedBookmarks, nil
}

func (db *Database) shouldSyncBookmark(bookmark *Bookmark, tx *sql.Tx) (shouldSync bool, err error) {
	if tx == nil {
		tx, err = db.db.BeginTx(context.Background(), &sql.TxOptions{})
		if err != nil {
			return false, err
		}
		defer func() {
			if err == nil {
				err = tx.Commit()
				return
			}
			// don't shadow original error
			_ = tx.Rollback()
		}()
	}
	qr := tx.QueryRow(`SELECT 1 FROM bookmarks WHERE url = ? AND clock > ?`, bookmark.URL, bookmark.Clock)
	var result int
	err = qr.Scan(&result)
	switch err {
	case sql.ErrNoRows:
		// Query does not match, therefore synced_at value is not older than the new clock value or id was not found
		return true, nil
	case nil:
		// Error is nil, therefore query matched and synced_at is older than the new clock
		return false, nil
	default:
		// Error is not nil and is not sql.ErrNoRows, therefore pass out the error
		return false, err
	}
}

func (db *Database) UpdateBookmark(originalURL string, bookmark Bookmark) error {
	insert, err := db.db.Prepare("UPDATE bookmarks SET url = ?, name = ?, image_url = ?, removed = ?, clock = ?, deleted_at = ? WHERE url = ?")
	if err != nil {
		return err
	}
	_, err = insert.Exec(bookmark.URL, bookmark.Name, bookmark.ImageURL, bookmark.Removed, bookmark.Clock, bookmark.DeletedAt, originalURL)
	return err
}

func (db *Database) DeleteBookmark(url string) error {
	_, err := db.db.Exec(`DELETE FROM bookmarks WHERE url = ?`, url)
	return err
}

func (db *Database) RemoveBookmark(url string) error {
	_, err := db.db.Exec(`UPDATE bookmarks SET removed = 1 WHERE url = ?`, url)
	return err
}
