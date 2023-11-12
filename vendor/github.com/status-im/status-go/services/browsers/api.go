package browsers

import (
	"context"

	"github.com/ethereum/go-ethereum/log"
)

func NewAPI(db *Database) *API {
	return &API{db: db}
}

// API is class with methods available over RPC.
type API struct {
	db *Database
}

func (api *API) GetBookmarks(ctx context.Context) ([]*Bookmark, error) {
	log.Debug("call to get bookmarks")
	rst, err := api.db.GetBookmarks()
	log.Debug("result from database for bookmarks", "len", len(rst))
	return rst, err
}

func (api *API) StoreBookmark(ctx context.Context, bookmark Bookmark) (Bookmark, error) {
	log.Debug("call to create a bookmark")
	bookmarkResult, err := api.db.StoreBookmark(bookmark)
	log.Debug("result from database for creating a bookmark", "err", err)
	return bookmarkResult, err
}

func (api *API) UpdateBookmark(ctx context.Context, originalURL string, bookmark Bookmark) error {
	log.Debug("call to update a bookmark")
	err := api.db.UpdateBookmark(originalURL, bookmark)
	log.Debug("result from database for updating a bookmark", "err", err)
	return err
}

func (api *API) DeleteBookmark(ctx context.Context, url string) error {
	log.Debug("call to remove a bookmark")
	err := api.db.DeleteBookmark(url)
	log.Debug("result from database for remove a bookmark", "err", err)
	return err
}

func (api *API) RemoveBookmark(ctx context.Context, url string) error {
	log.Debug("call to remove a bookmark logically")
	err := api.db.RemoveBookmark(url)
	log.Debug("result from database for remove a bookmark logically", "err", err)
	return err
}
