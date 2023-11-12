package protocol

import (
	"context"

	"github.com/status-im/status-go/services/browsers"
)

func (m *Messenger) AddBrowser(ctx context.Context, browser browsers.Browser) error {
	return m.persistence.AddBrowser(browser)
}

func (m *Messenger) GetBrowsers(ctx context.Context) (browsers []*browsers.Browser, err error) {
	return m.persistence.GetBrowsers()
}

func (m *Messenger) DeleteBrowser(ctx context.Context, id string) error {
	return m.persistence.DeleteBrowser(id)
}
