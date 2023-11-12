package protocol

import (
	"context"
	"time"

	"github.com/status-im/status-go/services/browsers"
)

func (m *Messenger) AddBookmark(ctx context.Context, bookmark browsers.Bookmark) error {
	bookmark.Removed = false
	bookmark.DeletedAt = 0
	bmr, err := m.persistence.AddBookmark(bookmark)
	if err != nil {
		return err
	}
	return m.SyncBookmark(ctx, &bmr, m.dispatchMessage)
}

func (m *Messenger) RemoveBookmark(ctx context.Context, url string) error {
	deletedAt := time.Now().Unix()
	err := m.persistence.RemoveBookmark(url, uint64(deletedAt))
	if err != nil {
		return err
	}

	bmr, err := m.persistence.GetBookmarkByURL(url)
	if err != nil {
		return err
	}
	return m.SyncBookmark(ctx, bmr, m.dispatchMessage)
}

func (m *Messenger) UpdateBookmark(ctx context.Context, oldURL string, bookmark browsers.Bookmark) error {
	err := m.persistence.UpdateBookmark(oldURL, bookmark)
	if err != nil {
		return err
	}
	return m.SyncBookmark(ctx, &bookmark, m.dispatchMessage)
}

func (m *Messenger) GarbageCollectRemovedBookmarks() error {
	return m.persistence.DeleteSoftRemovedBookmarks(uint64(time.Now().AddDate(0, 0, -30).Unix()))
}
