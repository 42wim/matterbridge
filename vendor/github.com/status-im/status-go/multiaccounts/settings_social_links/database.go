package sociallinkssettings

import (
	"context"
	"database/sql"
	"errors"

	"github.com/status-im/status-go/protocol/identity"
)

const (
	MaxNumOfSocialLinks = 20
)

var (
	ErrNilSocialLinkProvided    = errors.New("social links, nil object provided")
	ErrOlderSocialLinksProvided = errors.New("older social links provided")
)

type SocialLinksSettings struct {
	db *sql.DB
}

func NewSocialLinksSettings(db *sql.DB) *SocialLinksSettings {
	return &SocialLinksSettings{
		db: db,
	}
}

func (s *SocialLinksSettings) getSocialLinksClock(tx *sql.Tx) (result uint64, err error) {
	query := "SELECT social_links FROM settings_sync_clock WHERE synthetic_id = 'id'"
	if tx == nil {
		err = s.db.QueryRow(query).Scan(&result)
	} else {
		err = tx.QueryRow(query).Scan(&result)
	}
	return result, err
}

func (s *SocialLinksSettings) getSocialLinks(tx *sql.Tx) (identity.SocialLinks, error) {
	var (
		rows *sql.Rows
		err  error
	)
	query := "SELECT text, url FROM profile_social_links ORDER BY position ASC"

	if tx == nil {
		rows, err = s.db.Query(query)
	} else {
		rows, err = tx.Query(query)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var socialLinks identity.SocialLinks
	for rows.Next() {
		socialLink := &identity.SocialLink{}
		err := rows.Scan(&socialLink.Text, &socialLink.URL)
		if err != nil {
			return nil, err
		}
		socialLinks = append(socialLinks, socialLink)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return socialLinks, nil
}

func (s *SocialLinksSettings) GetSocialLinks() (identity.SocialLinks, error) {
	return s.getSocialLinks(nil)
}

func (s *SocialLinksSettings) GetSocialLinksClock() (result uint64, err error) {
	return s.getSocialLinksClock(nil)
}

func (s *SocialLinksSettings) AddOrReplaceSocialLinksIfNewer(links identity.SocialLinks, clock uint64) error {
	tx, err := s.db.BeginTx(context.Background(), &sql.TxOptions{})
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

	dbClock, err := s.getSocialLinksClock(tx)
	if err != nil {
		return err
	}

	if dbClock > clock {
		return ErrOlderSocialLinksProvided
	}

	dbLinks, err := s.getSocialLinks(tx)
	if err != nil {
		return err
	}
	if len(dbLinks) > 0 {
		_, err = tx.Exec("DELETE from profile_social_links")
		if err != nil {
			return err
		}
	}

	stmt, err := tx.Prepare("INSERT INTO profile_social_links (text, url, position) VALUES (?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	for position, link := range links {
		if link == nil {
			return ErrNilSocialLinkProvided
		}
		_, err = stmt.Exec(
			link.Text,
			link.URL,
			position,
		)
		if err != nil {
			return err
		}
	}

	stmt, err = tx.Prepare("UPDATE settings_sync_clock SET social_links = ? WHERE synthetic_id = 'id'")
	if err != nil {
		return err
	}

	_, err = stmt.Exec(clock)
	return err
}
