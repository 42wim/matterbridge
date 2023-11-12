package verification

import (
	"database/sql"
	"errors"
	"time"
)

var (
	ErrVerificationRequestNotFound = errors.New("verification request not found")
)

type Persistence struct {
	db *sql.DB
}

func NewPersistence(db *sql.DB) *Persistence {
	return &Persistence{
		db: db,
	}
}

type RequestStatus int

const (
	RequestStatusUNKNOWN RequestStatus = iota
	RequestStatusPENDING
	RequestStatusACCEPTED
	RequestStatusDECLINED
	RequestStatusCANCELED
	RequestStatusTRUSTED
	RequestStatusUNTRUSTWORTHY
)

type TrustStatus int

const (
	TrustStatusUNKNOWN TrustStatus = iota
	TrustStatusTRUSTED
	TrustStatusUNTRUSTWORTHY
)

type Request struct {
	ID            string        `json:"id"`
	From          string        `json:"from"`
	To            string        `json:"to"`
	Challenge     string        `json:"challenge"`
	Response      string        `json:"response"`
	RequestedAt   uint64        `json:"requested_at"`
	RequestStatus RequestStatus `json:"verification_status"`
	RepliedAt     uint64        `json:"replied_at"`
}

func (p *Persistence) GetVerificationRequests() ([]Request, error) {
	rows, err := p.db.Query("SELECT id, from_user, to_user, challenge, response, requested_at, verification_status, replied_at FROM verification_requests_individual")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []Request
	for rows.Next() {
		var vr Request
		err = rows.Scan(&vr.ID, &vr.From, &vr.To, &vr.Challenge, &vr.Response, &vr.RequestedAt, &vr.RepliedAt, &vr.RequestStatus)
		if err != nil {
			return nil, err
		}
		result = append(result, vr)
	}
	return result, nil
}

func (p *Persistence) GetVerificationRequest(id string) (*Request, error) {
	var vr Request
	err := p.db.QueryRow(`SELECT id, from_user, to_user, challenge, response, requested_at, verification_status, replied_at FROM verification_requests_individual WHERE id = ?`, id).Scan(
		&vr.ID,
		&vr.From,
		&vr.To,
		&vr.Challenge,
		&vr.Response,
		&vr.RequestedAt,
		&vr.RequestStatus,
		&vr.RepliedAt,
	)

	switch err {
	case sql.ErrNoRows:
		return nil, nil
	case nil:
		return &vr, nil
	default:
		return nil, err
	}
}

func (p *Persistence) GetReceivedVerificationRequests(myPublicKey string) ([]*Request, error) {
	response := make([]*Request, 0)

	query := `SELECT id, from_user, to_user, challenge, response, requested_at, verification_status, replied_at FROM verification_requests_individual WHERE to_user = ?`
	rows, err := p.db.Query(query, myPublicKey)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	for rows.Next() {
		var vr Request

		err := rows.Scan(
			&vr.ID,
			&vr.From,
			&vr.To,
			&vr.Challenge,
			&vr.Response,
			&vr.RequestedAt,
			&vr.RequestStatus,
			&vr.RepliedAt,
		)
		if err != nil {
			return nil, err
		}

		response = append(response, &vr)
	}

	return response, nil
}

func (p *Persistence) GetLatestVerificationRequestSentTo(contactID string) (*Request, error) {
	var vr Request
	err := p.db.QueryRow(`SELECT id, from_user, to_user, challenge, response, requested_at, verification_status, replied_at FROM verification_requests_individual WHERE to_user = ? ORDER BY requested_at DESC`, contactID).Scan(
		&vr.ID,
		&vr.From,
		&vr.To,
		&vr.Challenge,
		&vr.Response,
		&vr.RequestedAt,
		&vr.RequestStatus,
		&vr.RepliedAt,
	)

	switch err {
	case sql.ErrNoRows:
		return nil, nil
	case nil:
		return &vr, nil
	default:
		return nil, err
	}
}

func (p *Persistence) GetLatestVerificationRequestFrom(contactID string) (*Request, error) {
	var vr Request
	err := p.db.QueryRow(`SELECT id, from_user, to_user, challenge, response, requested_at, verification_status, replied_at FROM verification_requests_individual WHERE from_user = ? ORDER BY requested_at DESC`, contactID).Scan(
		&vr.ID,
		&vr.From,
		&vr.To,
		&vr.Challenge,
		&vr.Response,
		&vr.RequestedAt,
		&vr.RequestStatus,
		&vr.RepliedAt,
	)

	switch err {
	case sql.ErrNoRows:
		return nil, nil
	case nil:
		return &vr, nil
	default:
		return nil, err
	}
}

func (p *Persistence) SaveVerificationRequest(vr *Request) error {
	if vr == nil {
		return errors.New("invalid verification request provided")
	}
	_, err := p.db.Exec(`INSERT INTO verification_requests_individual (id, from_user, to_user, challenge, response, requested_at, verification_status, replied_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`, vr.ID, vr.From, vr.To, vr.Challenge, vr.Response, vr.RequestedAt, vr.RequestStatus, vr.RepliedAt)
	return err
}

func (p *Persistence) AcceptContactVerificationRequest(id string, response string) error {
	result, err := p.db.Exec("UPDATE verification_requests_individual SET response = ?, replied_at = ?, verification_status = ? WHERE id = ?", response, time.Now().Unix(), RequestStatusACCEPTED, id)
	if err != nil {
		return err
	}

	numRows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if numRows == 0 {
		return ErrVerificationRequestNotFound
	}

	return nil
}

func (p *Persistence) DeclineContactVerificationRequest(id string) error {
	result, err := p.db.Exec("UPDATE verification_requests_individual SET response = '', replied_at = ?, verification_status = ? WHERE id = ?", time.Now().Unix(), RequestStatusDECLINED, id)
	if err != nil {
		return err
	}

	numRows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if numRows == 0 {
		return ErrVerificationRequestNotFound
	}

	return nil
}

func (p *Persistence) SetTrustStatus(contactID string, trust TrustStatus, updatedAt uint64) error {
	_, err := p.db.Exec(`INSERT INTO trusted_users (id, trust_status, updated_at) VALUES (?, ?, ?)`, contactID, trust, updatedAt)
	return err
}

func (p *Persistence) UpsertTrustStatus(contactID string, trust TrustStatus, updatedAt uint64) (shouldSync bool, err error) {
	var t uint64
	err = p.db.QueryRow(`SELECT updated_at FROM trusted_users WHERE id = ?`, contactID).Scan(&t)

	if err == sql.ErrNoRows {
		return true, p.SetTrustStatus(contactID, trust, updatedAt)
	}

	if err == nil && updatedAt > t {
		_, err := p.db.Exec("UPDATE trusted_users SET trust_status = ?, updated_at = ? WHERE id = ?", trust, updatedAt, contactID)
		if err != nil {
			return true, err
		}
	}

	return false, err
}

func (p *Persistence) GetTrustStatus(contactID string) (TrustStatus, error) {
	var t TrustStatus
	err := p.db.QueryRow(`SELECT trust_status FROM trusted_users WHERE id = ?`, contactID).Scan(&t)

	switch err {
	case sql.ErrNoRows:
		return TrustStatusUNKNOWN, nil
	case nil:
		return t, nil
	default:
		return TrustStatusUNKNOWN, err
	}
}

func (p *Persistence) GetAllTrustStatus() (map[string]TrustStatus, error) {
	result := make(map[string]TrustStatus)
	rows, err := p.db.Query("SELECT id, trust_status FROM trusted_users")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var id string
		var ts TrustStatus
		err = rows.Scan(&id, &ts)
		if err != nil {
			return nil, err
		}

		result[id] = ts
	}

	return result, nil
}
