package communities

import (
	"context"
	"crypto/ecdsa"
	"database/sql"
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"

	"github.com/golang/protobuf/proto"

	"github.com/status-im/status-go/eth-node/crypto"
	"github.com/status-im/status-go/eth-node/types"
	"github.com/status-im/status-go/protocol/common"
	"github.com/status-im/status-go/protocol/common/shard"
	"github.com/status-im/status-go/protocol/communities/token"
	"github.com/status-im/status-go/protocol/protobuf"
	"github.com/status-im/status-go/services/wallet/bigint"
)

type Persistence struct {
	db *sql.DB

	recordBundleToCommunity func(*CommunityRecordBundle) (*Community, error)
}

var ErrOldRequestToJoin = errors.New("old request to join")
var ErrOldRequestToLeave = errors.New("old request to leave")
var ErrOldShardInfo = errors.New("old shard info")

type CommunityRecord struct {
	id           []byte
	privateKey   []byte
	controlNode  []byte
	description  []byte
	joined       bool
	joinedAt     int64
	verified     bool
	spectated    bool
	muted        bool
	mutedTill    time.Time
	shardCluster *uint
	shardIndex   *uint
	lastOpenedAt int64
}

type EventsRecord struct {
	id             []byte
	rawEvents      []byte
	rawDescription []byte
}

type RequestToJoinRecord struct {
	id          []byte
	publicKey   string
	clock       int
	ensName     string
	chatID      string
	communityID []byte
	state       int
}

type CommunityRecordBundle struct {
	community      *CommunityRecord
	events         *EventsRecord
	requestToJoin  *RequestToJoinRecord
	installationID *string
}

const OR = " OR "
const communitiesBaseQuery = `
	SELECT
		c.id, c.private_key, c.control_node, c.description, c.joined, c.joined_at, c.last_opened_at, c.spectated, c.verified, c.muted, c.muted_till,
		csd.shard_cluster, csd.shard_index,
		r.id, r.public_key, r.clock, r.ens_name, r.chat_id, r.state,
		ae.raw_events, ae.raw_description,
		ccn.installation_id
	FROM communities_communities c
	LEFT JOIN communities_shards csd ON c.id = csd.community_id
	LEFT JOIN communities_requests_to_join r ON c.id = r.community_id AND r.public_key = ?
	LEFT JOIN communities_events ae ON c.id = ae.id
	LEFT JOIN communities_control_node ccn ON c.id = ccn.community_id`

func scanCommunity(scanner func(dest ...any) error) (*CommunityRecordBundle, error) {
	r := &CommunityRecordBundle{
		community:      &CommunityRecord{},
		events:         nil,
		requestToJoin:  nil,
		installationID: nil,
	}

	var mutedTill sql.NullTime
	var cluster, index sql.NullInt64

	var requestToJoinID []byte
	var requestToJoinPublicKey, requestToJoinENSName, requestToJoinChatID sql.NullString
	var requestToJoinClock, requestToJoinState sql.NullInt64

	var events, eventsDescription []byte

	var installationID sql.NullString

	err := scanner(
		&r.community.id,
		&r.community.privateKey,
		&r.community.controlNode,
		&r.community.description,
		&r.community.joined,
		&r.community.joinedAt,
		&r.community.lastOpenedAt,
		&r.community.spectated,
		&r.community.verified,
		&r.community.muted,
		&mutedTill,
		&cluster,
		&index,

		&requestToJoinID,
		&requestToJoinPublicKey,
		&requestToJoinClock,
		&requestToJoinENSName,
		&requestToJoinChatID,
		&requestToJoinState,

		&events,
		&eventsDescription,

		&installationID,
	)
	if err != nil {
		return nil, err
	}

	if mutedTill.Valid {
		r.community.mutedTill = mutedTill.Time
	}
	if cluster.Valid {
		clusterValue := uint(cluster.Int64)
		r.community.shardCluster = &clusterValue
	}
	if index.Valid {
		shardIndexValue := uint(index.Int64)
		r.community.shardIndex = &shardIndexValue
	}

	if requestToJoinID != nil {
		r.requestToJoin = &RequestToJoinRecord{
			id:          requestToJoinID,
			publicKey:   requestToJoinPublicKey.String,
			clock:       int(requestToJoinClock.Int64),
			ensName:     requestToJoinENSName.String,
			chatID:      requestToJoinChatID.String,
			communityID: r.community.id,
			state:       int(requestToJoinState.Int64),
		}
	}

	if events != nil {
		r.events = &EventsRecord{
			id:             r.community.id,
			rawEvents:      events,
			rawDescription: eventsDescription,
		}
	}

	if installationID.Valid {
		r.installationID = &installationID.String
	}

	return r, nil
}

func (p *Persistence) saveCommunity(r *CommunityRecord) error {
	_, err := p.db.Exec(`
        INSERT INTO communities_communities (
            id, private_key, control_node, description,
            joined, joined_at, spectated, verified, muted, muted_till, last_opened_at
        ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		r.id, r.privateKey, r.controlNode, r.description,
		r.joined, r.joinedAt, r.spectated, r.verified, r.muted, r.mutedTill, r.lastOpenedAt)
	return err
}

func (p *Persistence) SaveCommunity(community *Community) error {
	record, err := communityToRecord(community)
	if err != nil {
		return err
	}
	return p.saveCommunity(record)
}

func (p *Persistence) DeleteCommunityEvents(id types.HexBytes) error {
	_, err := p.db.Exec(`DELETE FROM communities_events WHERE id = ?;`, id)
	return err
}

func (p *Persistence) saveCommunityEvents(r *EventsRecord) error {
	_, err := p.db.Exec(`
		INSERT INTO communities_events (
			id, raw_events, raw_description
		) VALUES (?, ?, ?);`,
		r.id, r.rawEvents, r.rawDescription)
	return err
}

func (p *Persistence) SaveCommunityEvents(community *Community) error {
	record, err := communityToEventsRecord(community)
	if err != nil {
		return err
	}
	return p.saveCommunityEvents(record)
}

func (p *Persistence) DeleteCommunity(id types.HexBytes) error {
	_, err := p.db.Exec(`DELETE FROM communities_communities WHERE id = ?;
						 DELETE FROM communities_events WHERE id = ?;
						 DELETE FROM communities_shards WHERE community_id = ?`, id, id, id)
	return err
}

func (p *Persistence) ShouldHandleSyncCommunitySettings(settings *protobuf.SyncCommunitySettings) (bool, error) {

	qr := p.db.QueryRow(`SELECT * FROM communities_settings WHERE community_id = ? AND clock > ?`, settings.CommunityId, settings.Clock)
	_, err := p.scanRowToStruct(qr.Scan)
	switch err {
	case sql.ErrNoRows:
		// Query does not match, therefore clock value is not older than the new clock value or id was not found
		return true, nil
	case nil:
		// Error is nil, therefore query matched and clock is older than the new clock
		return false, nil
	default:
		// Error is not nil and is not sql.ErrNoRows, therefore pass out the error
		return false, err
	}
}

func (p *Persistence) ShouldHandleSyncCommunity(community *protobuf.SyncInstallationCommunity) (bool, error) {
	// TODO see if there is a way to make this more elegant
	// When the test for this function fails because the table has changed we should update sync functionality
	qr := p.db.QueryRow(`SELECT id, private_key, description, joined, joined_at, verified, spectated, muted, muted_till, synced_at, last_opened_at FROM communities_communities WHERE id = ? AND synced_at > ?`, community.Id, community.Clock)
	_, err := p.scanRowToStruct(qr.Scan)

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

func (p *Persistence) queryCommunities(memberIdentity *ecdsa.PublicKey, query string) (response []*Community, err error) {
	rows, err := p.db.Query(query, common.PubkeyToHex(memberIdentity))
	if err != nil {
		return nil, err
	}

	return p.rowsToCommunities(rows)
}

func (p *Persistence) AllCommunities(memberIdentity *ecdsa.PublicKey) ([]*Community, error) {
	return p.queryCommunities(memberIdentity, communitiesBaseQuery)
}

func (p *Persistence) JoinedCommunities(memberIdentity *ecdsa.PublicKey) ([]*Community, error) {
	query := communitiesBaseQuery + ` WHERE c.joined`
	return p.queryCommunities(memberIdentity, query)
}

func (p *Persistence) UpdateLastOpenedAt(communityID types.HexBytes, timestamp int64) error {
	_, err := p.db.Exec(`UPDATE communities_communities SET last_opened_at = ? WHERE id = ?`, timestamp, communityID)
	return err
}

func (p *Persistence) SpectatedCommunities(memberIdentity *ecdsa.PublicKey) ([]*Community, error) {
	query := communitiesBaseQuery + ` WHERE c.spectated`
	return p.queryCommunities(memberIdentity, query)
}

func (p *Persistence) rowsToCommunityRecords(rows *sql.Rows) (result []*CommunityRecordBundle, err error) {
	defer func() {
		if err != nil {
			// Don't shadow original error
			_ = rows.Close()
			return

		}
		err = rows.Close()
	}()

	for rows.Next() {
		r, err := scanCommunity(rows.Scan)
		if err != nil {
			return nil, err
		}
		result = append(result, r)
	}

	return result, nil
}

func (p *Persistence) rowsToCommunities(rows *sql.Rows) (comms []*Community, err error) {
	records, err := p.rowsToCommunityRecords(rows)
	if err != nil {
		return nil, err
	}

	for _, record := range records {
		org, err := p.recordBundleToCommunity(record)
		if err != nil {
			return nil, err
		}

		comms = append(comms, org)
	}

	return comms, nil
}

func (p *Persistence) JoinedAndPendingCommunitiesWithRequests(memberIdentity *ecdsa.PublicKey) (comms []*Community, err error) {
	query := communitiesBaseQuery + ` WHERE c.Joined OR r.state = ?`

	rows, err := p.db.Query(query, common.PubkeyToHex(memberIdentity), RequestToJoinStatePending)
	if err != nil {
		return nil, err
	}

	return p.rowsToCommunities(rows)
}

func (p *Persistence) DeletedCommunities(memberIdentity *ecdsa.PublicKey) (comms []*Community, err error) {
	query := communitiesBaseQuery + ` WHERE NOT c.Joined AND (r.community_id IS NULL or r.state != ?)`

	rows, err := p.db.Query(query, common.PubkeyToHex(memberIdentity), RequestToJoinStatePending)
	if err != nil {
		return nil, err
	}

	return p.rowsToCommunities(rows)
}

func (p *Persistence) CommunitiesWithPrivateKey(memberIdentity *ecdsa.PublicKey) ([]*Community, error) {
	query := communitiesBaseQuery + ` WHERE c.private_key IS NOT NULL`
	return p.queryCommunities(memberIdentity, query)
}

func (p *Persistence) getByID(id []byte, memberIdentity *ecdsa.PublicKey) (*CommunityRecordBundle, error) {
	r, err := scanCommunity(p.db.QueryRow(communitiesBaseQuery+` WHERE c.id = ?`, common.PubkeyToHex(memberIdentity), id).Scan)
	return r, err
}

func (p *Persistence) GetByID(memberIdentity *ecdsa.PublicKey, id []byte) (*Community, error) {
	r, err := p.getByID(id, memberIdentity)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return p.recordBundleToCommunity(r)
}

func (p *Persistence) SaveRequestToJoin(request *RequestToJoin) (err error) {
	tx, err := p.db.BeginTx(context.Background(), &sql.TxOptions{})
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

	var clock uint64
	// Fetch any existing request to join
	err = tx.QueryRow(`SELECT clock FROM communities_requests_to_join WHERE state = ? AND public_key = ? AND community_id = ?`, RequestToJoinStatePending, request.PublicKey, request.CommunityID).Scan(&clock)
	if err != nil && err != sql.ErrNoRows {
		return err
	}

	// This is already processed
	if clock >= request.Clock {
		return ErrOldRequestToJoin
	}

	_, err = tx.Exec(`INSERT OR REPLACE INTO communities_requests_to_join(id,public_key,clock,ens_name,chat_id,community_id,state) VALUES (?, ?, ?, ?, ?, ?, ?)`, request.ID, request.PublicKey, request.Clock, request.ENSName, request.ChatID, request.CommunityID, request.State)
	return err
}

func (p *Persistence) SaveRequestToJoinRevealedAddresses(requestID types.HexBytes, revealedAccounts []*protobuf.RevealedAccount) (err error) {
	tx, err := p.db.BeginTx(context.Background(), &sql.TxOptions{})
	if err != nil {
		return
	}
	defer func() {
		if err == nil {
			err = tx.Commit()
			return
		}
		// don't shadow original error
		_ = tx.Rollback()
	}()

	query := `INSERT OR REPLACE INTO communities_requests_to_join_revealed_addresses (request_id, address, chain_ids, is_airdrop_address, signature) VALUES (?, ?, ?, ?, ?)`
	stmt, err := tx.Prepare(query)
	if err != nil {
		return
	}
	defer stmt.Close()
	for _, account := range revealedAccounts {

		var chainIDs []string
		for _, ID := range account.ChainIds {
			chainIDs = append(chainIDs, strconv.Itoa(int(ID)))
		}

		_, err = stmt.Exec(
			requestID,
			account.Address,
			strings.Join(chainIDs, ","),
			account.IsAirdropAddress,
			account.Signature,
		)
		if err != nil {
			return
		}
	}
	return
}

func (p *Persistence) SaveCheckChannelPermissionResponse(communityID string, chatID string, response *CheckChannelPermissionsResponse) error {
	tx, err := p.db.BeginTx(context.Background(), &sql.TxOptions{})
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

	viewOnlyPermissionIDs := make([]string, 0)
	viewAndPostPermissionIDs := make([]string, 0)

	for permissionID := range response.ViewOnlyPermissions.Permissions {
		viewOnlyPermissionIDs = append(viewOnlyPermissionIDs, permissionID)
	}
	for permissionID := range response.ViewAndPostPermissions.Permissions {
		viewAndPostPermissionIDs = append(viewAndPostPermissionIDs, permissionID)
	}

	_, err = tx.Exec(`INSERT INTO communities_check_channel_permission_responses (community_id,chat_id,view_only_permissions_satisfied,view_and_post_permissions_satisfied, view_only_permission_ids, view_and_post_permission_ids) VALUES (?, ?, ?, ?, ?, ?)`, communityID, chatID, response.ViewOnlyPermissions.Satisfied, response.ViewAndPostPermissions.Satisfied, strings.Join(viewOnlyPermissionIDs[:], ","), strings.Join(viewAndPostPermissionIDs[:], ","))
	if err != nil {
		return err
	}

	saveCriteriaResults := func(permissions map[string]*PermissionTokenCriteriaResult) error {
		for permissionID, criteriaResult := range permissions {

			criteria := make([]string, 0)
			for _, val := range criteriaResult.Criteria {
				criteria = append(criteria, strconv.FormatBool(val))
			}

			_, err = tx.Exec(`INSERT INTO communities_permission_token_criteria_results (permission_id,community_id, chat_id, criteria) VALUES (?, ?, ?, ?)`, permissionID, communityID, chatID, strings.Join(criteria[:], ","))
			if err != nil {
				return err
			}
		}
		return nil
	}

	err = saveCriteriaResults(response.ViewOnlyPermissions.Permissions)
	if err != nil {
		return err
	}
	return saveCriteriaResults(response.ViewAndPostPermissions.Permissions)
}

func (p *Persistence) GetCheckChannelPermissionResponses(communityID string) (map[string]*CheckChannelPermissionsResponse, error) {

	rows, err := p.db.Query(`SELECT chat_id, view_only_permissions_satisfied, view_and_post_permissions_satisfied, view_only_permission_ids, view_and_post_permission_ids FROM communities_check_channel_permission_responses WHERE community_id = ?`, communityID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	checkChannelPermissionResponses := make(map[string]*CheckChannelPermissionsResponse, 0)

	for rows.Next() {

		permissionResponse := &CheckChannelPermissionsResponse{
			ViewOnlyPermissions: &CheckChannelViewOnlyPermissionsResult{
				Satisfied:   false,
				Permissions: make(map[string]*PermissionTokenCriteriaResult),
			},
			ViewAndPostPermissions: &CheckChannelViewAndPostPermissionsResult{
				Satisfied:   false,
				Permissions: make(map[string]*PermissionTokenCriteriaResult),
			},
		}

		var chatID string
		var viewOnlyPermissionIDsString string
		var viewAndPostPermissionIDsString string

		err := rows.Scan(&chatID, &permissionResponse.ViewOnlyPermissions.Satisfied, &permissionResponse.ViewAndPostPermissions.Satisfied, &viewOnlyPermissionIDsString, &viewAndPostPermissionIDsString)
		if err != nil {
			return nil, err
		}

		for _, permissionID := range strings.Split(viewOnlyPermissionIDsString, ",") {
			if permissionID != "" {
				permissionResponse.ViewOnlyPermissions.Permissions[permissionID] = &PermissionTokenCriteriaResult{Criteria: make([]bool, 0)}
			}
		}
		for _, permissionID := range strings.Split(viewAndPostPermissionIDsString, ",") {
			if permissionID != "" {
				permissionResponse.ViewAndPostPermissions.Permissions[permissionID] = &PermissionTokenCriteriaResult{Criteria: make([]bool, 0)}
			}
		}
		checkChannelPermissionResponses[chatID] = permissionResponse
	}

	addCriteriaResult := func(channelResponses map[string]*CheckChannelPermissionsResponse, permissions map[string]*PermissionTokenCriteriaResult, chatID string, viewOnly bool) error {
		for permissionID := range permissions {
			criteria, err := p.GetPermissionTokenCriteriaResult(permissionID, communityID, chatID)
			if err != nil {
				return err
			}
			if viewOnly {
				channelResponses[chatID].ViewOnlyPermissions.Permissions[permissionID] = criteria
			} else {
				channelResponses[chatID].ViewAndPostPermissions.Permissions[permissionID] = criteria
			}
		}
		return nil
	}

	for chatID, response := range checkChannelPermissionResponses {
		err := addCriteriaResult(checkChannelPermissionResponses, response.ViewOnlyPermissions.Permissions, chatID, true)
		if err != nil {
			return nil, err
		}
		err = addCriteriaResult(checkChannelPermissionResponses, response.ViewAndPostPermissions.Permissions, chatID, false)
		if err != nil {
			return nil, err
		}
	}
	return checkChannelPermissionResponses, nil
}

func (p *Persistence) GetPermissionTokenCriteriaResult(permissionID string, communityID string, chatID string) (*PermissionTokenCriteriaResult, error) {
	tx, err := p.db.BeginTx(context.Background(), &sql.TxOptions{})
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

	criteriaString := ""
	err = tx.QueryRow(`SELECT criteria FROM communities_permission_token_criteria_results WHERE permission_id = ? AND community_id = ? AND chat_id = ?`, permissionID, communityID, chatID).Scan(&criteriaString)
	if err != nil {
		return nil, err
	}

	criteria := make([]bool, 0)
	for _, r := range strings.Split(criteriaString, ",") {
		val, err := strconv.ParseBool(r)
		if err != nil {
			return nil, err
		}
		criteria = append(criteria, val)
	}

	return &PermissionTokenCriteriaResult{Criteria: criteria}, nil
}

func (p *Persistence) RemoveRequestToJoinRevealedAddresses(requestID []byte) error {
	_, err := p.db.Exec(`DELETE FROM communities_requests_to_join_revealed_addresses WHERE request_id = ?`, requestID)
	return err
}

func (p *Persistence) GetRequestToJoinRevealedAddresses(requestID []byte) ([]*protobuf.RevealedAccount, error) {
	revealedAccounts := make([]*protobuf.RevealedAccount, 0)
	rows, err := p.db.Query(`SELECT address, chain_ids, is_airdrop_address, signature FROM communities_requests_to_join_revealed_addresses WHERE request_id = ?`, requestID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var address sql.NullString
		var chainIDsStr sql.NullString
		var isAirdropAddress sql.NullBool
		var signature sql.RawBytes
		err := rows.Scan(&address, &chainIDsStr, &isAirdropAddress, &signature)
		if err != nil {
			return nil, err
		}

		revealedAccount, err := toRevealedAccount(address, chainIDsStr, isAirdropAddress, signature)
		if err != nil {
			return nil, err
		}

		if revealedAccount == nil {
			return nil, errors.New("invalid RequestToJoin RevealedAddresses data")
		}
		revealedAccounts = append(revealedAccounts, revealedAccount)
	}
	return revealedAccounts, nil
}

func (p *Persistence) SaveRequestToLeave(request *RequestToLeave) error {
	tx, err := p.db.BeginTx(context.Background(), &sql.TxOptions{})
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

	var clock uint64
	// Fetch any existing request to leave
	err = tx.QueryRow(`SELECT clock FROM communities_requests_to_leave WHERE public_key = ? AND community_id = ?`, request.PublicKey, request.CommunityID).Scan(&clock)
	if err != nil && err != sql.ErrNoRows {
		return err
	}

	// This is already processed
	if clock >= request.Clock {
		return ErrOldRequestToLeave
	}

	_, err = tx.Exec(`INSERT INTO communities_requests_to_leave(id,public_key,clock,community_id) VALUES (?, ?, ?, ?)`, request.ID, request.PublicKey, request.Clock, request.CommunityID)
	return err
}

func (p *Persistence) CanceledRequestsToJoinForUser(pk string) ([]*RequestToJoin, error) {
	var requests []*RequestToJoin
	rows, err := p.db.Query(`SELECT id,public_key,clock,ens_name,chat_id,community_id,state FROM communities_requests_to_join WHERE state = ? AND public_key = ?`, RequestToJoinStateCanceled, pk)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		request := &RequestToJoin{}
		err := rows.Scan(&request.ID, &request.PublicKey, &request.Clock, &request.ENSName, &request.ChatID, &request.CommunityID, &request.State)
		if err != nil {
			return nil, err
		}
		requests = append(requests, request)
	}
	return requests, nil
}

func (p *Persistence) CanceledRequestToJoinForUserForCommunityID(pk string, communityID []byte) (*RequestToJoin, error) {
	row := p.db.QueryRow(`SELECT id,public_key,clock,ens_name,chat_id,community_id,state
	FROM
	communities_requests_to_join
	WHERE
	state = ? AND public_key = ? AND community_id = ?`,
		RequestToJoinStateCanceled, pk, communityID)

	request := &RequestToJoin{}

	err := row.Scan(&request.ID, &request.PublicKey, &request.Clock, &request.ENSName, &request.ChatID, &request.CommunityID, &request.State)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return request, nil
}

func (p *Persistence) RequestsToJoinForUserByState(pk string, state RequestToJoinState) ([]*RequestToJoin, error) {
	var requests []*RequestToJoin
	rows, err := p.db.Query(`SELECT id,public_key,clock,ens_name,chat_id,community_id,state FROM communities_requests_to_join WHERE state = ? AND public_key = ?`, state, pk)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		request := &RequestToJoin{}
		err := rows.Scan(&request.ID, &request.PublicKey, &request.Clock, &request.ENSName, &request.ChatID, &request.CommunityID, &request.State)
		if err != nil {
			return nil, err
		}
		requests = append(requests, request)
	}
	return requests, nil
}

func (p *Persistence) HasPendingRequestsToJoinForUserAndCommunity(userPk string, communityID []byte) (bool, error) {
	var count int
	err := p.db.QueryRow(`SELECT count(1) FROM communities_requests_to_join WHERE state = ? AND public_key = ? AND community_id = ?`, RequestToJoinStatePending, userPk, communityID).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (p *Persistence) RequestsToJoinForCommunityWithState(id []byte, state RequestToJoinState) ([]*RequestToJoin, error) {
	var requests []*RequestToJoin
	rows, err := p.db.Query(`SELECT id,public_key,clock,ens_name,chat_id,community_id,state FROM communities_requests_to_join WHERE state = ? AND community_id = ?`, state, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		request := &RequestToJoin{}
		err := rows.Scan(&request.ID, &request.PublicKey, &request.Clock, &request.ENSName, &request.ChatID, &request.CommunityID, &request.State)
		if err != nil {
			return nil, err
		}
		requests = append(requests, request)
	}
	return requests, nil
}

func (p *Persistence) PendingRequestsToJoin() ([]*RequestToJoin, error) {
	var requests []*RequestToJoin
	rows, err := p.db.Query(`SELECT id,public_key,clock,ens_name,chat_id,community_id,state FROM communities_requests_to_join WHERE state = ?`, RequestToJoinStatePending)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		request := &RequestToJoin{}
		err := rows.Scan(&request.ID, &request.PublicKey, &request.Clock, &request.ENSName, &request.ChatID, &request.CommunityID, &request.State)
		if err != nil {
			return nil, err
		}
		requests = append(requests, request)
	}
	return requests, nil
}

func (p *Persistence) PendingRequestsToJoinForCommunity(id []byte) ([]*RequestToJoin, error) {
	return p.RequestsToJoinForCommunityWithState(id, RequestToJoinStatePending)
}

func (p *Persistence) DeclinedRequestsToJoinForCommunity(id []byte) ([]*RequestToJoin, error) {
	return p.RequestsToJoinForCommunityWithState(id, RequestToJoinStateDeclined)
}

func (p *Persistence) CanceledRequestsToJoinForCommunity(id []byte) ([]*RequestToJoin, error) {
	return p.RequestsToJoinForCommunityWithState(id, RequestToJoinStateCanceled)
}

func (p *Persistence) AcceptedRequestsToJoinForCommunity(id []byte) ([]*RequestToJoin, error) {
	return p.RequestsToJoinForCommunityWithState(id, RequestToJoinStateAccepted)
}

func (p *Persistence) AcceptedPendingRequestsToJoinForCommunity(id []byte) ([]*RequestToJoin, error) {
	return p.RequestsToJoinForCommunityWithState(id, RequestToJoinStateAcceptedPending)
}

func (p *Persistence) DeclinedPendingRequestsToJoinForCommunity(id []byte) ([]*RequestToJoin, error) {
	return p.RequestsToJoinForCommunityWithState(id, RequestToJoinStateDeclinedPending)
}

func (p *Persistence) RequestsToJoinForCommunityAwaitingAddresses(id []byte) ([]*RequestToJoin, error) {
	return p.RequestsToJoinForCommunityWithState(id, RequestToJoinStateAwaitingAddresses)
}

func (p *Persistence) SetRequestToJoinState(pk string, communityID []byte, state RequestToJoinState) error {
	_, err := p.db.Exec(`UPDATE communities_requests_to_join SET state = ? WHERE community_id = ? AND public_key = ?`, state, communityID, pk)
	return err
}

func (p *Persistence) DeletePendingRequestToJoin(id []byte) error {
	_, err := p.db.Exec(`DELETE FROM communities_requests_to_join WHERE id = ?`, id)
	return err
}

// UpdateClockInRequestToJoin method is used for testing
func (p *Persistence) UpdateClockInRequestToJoin(id []byte, clock uint64) error {
	_, err := p.db.Exec(`UPDATE communities_requests_to_join SET clock = ? WHERE id = ?`, clock, id)
	return err
}

func (p *Persistence) SetMuted(communityID []byte, muted bool) error {
	_, err := p.db.Exec(`UPDATE communities_communities SET muted = ? WHERE id = ?`, muted, communityID)
	return err
}

func (p *Persistence) MuteCommunityTill(communityID []byte, mutedTill time.Time) error {
	mutedTillFormatted := mutedTill.Format(time.RFC3339)
	_, err := p.db.Exec(`UPDATE communities_communities SET muted_till = ? WHERE id = ?`, mutedTillFormatted, communityID)
	return err
}

func (p *Persistence) GetRequestToJoin(id []byte) (*RequestToJoin, error) {
	request := &RequestToJoin{}
	err := p.db.QueryRow(`SELECT id,public_key,clock,ens_name,chat_id,community_id,state FROM communities_requests_to_join WHERE id = ?`, id).Scan(&request.ID, &request.PublicKey, &request.Clock, &request.ENSName, &request.ChatID, &request.CommunityID, &request.State)
	if err != nil {
		return nil, err
	}

	return request, nil
}

func (p *Persistence) GetRequestToJoinByPkAndCommunityID(pk string, communityID []byte) (*RequestToJoin, error) {
	request := &RequestToJoin{}
	err := p.db.QueryRow(`SELECT id,public_key,clock,ens_name,chat_id,community_id,state FROM communities_requests_to_join WHERE public_key = ? AND community_id = ?`, pk, communityID).Scan(&request.ID, &request.PublicKey, &request.Clock, &request.ENSName, &request.ChatID, &request.CommunityID, &request.State)
	if err != nil {
		return nil, err
	}

	return request, nil
}

func (p *Persistence) GetRequestToJoinIDByPkAndCommunityID(pk string, communityID []byte) ([]byte, error) {
	var id []byte
	err := p.db.QueryRow(`SELECT id FROM communities_requests_to_join WHERE community_id = ? AND public_key = ?`, communityID, pk).Scan(&id)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	return id, nil
}

func (p *Persistence) GetRequestToJoinByPk(pk string, communityID []byte, state RequestToJoinState) (*RequestToJoin, error) {
	request := &RequestToJoin{}
	err := p.db.QueryRow(`SELECT id,public_key,clock,ens_name,chat_id,community_id,state FROM communities_requests_to_join WHERE public_key = ? AND community_id = ? AND state = ?`, pk, communityID, state).Scan(&request.ID, &request.PublicKey, &request.Clock, &request.ENSName, &request.ChatID, &request.CommunityID, &request.State)
	if err != nil {
		return nil, err
	}

	return request, nil
}

func (p *Persistence) SetSyncClock(id []byte, clock uint64) error {
	_, err := p.db.Exec(`UPDATE communities_communities SET synced_at = ? WHERE id = ? AND synced_at < ?`, clock, id, clock)
	return err
}

func (p *Persistence) SetPrivateKey(id []byte, privKey *ecdsa.PrivateKey) error {
	_, err := p.db.Exec(`UPDATE communities_communities SET private_key = ? WHERE id = ?`, crypto.FromECDSA(privKey), id)
	return err
}

func (p *Persistence) SaveWakuMessages(messages []*types.Message) (err error) {
	tx, err := p.db.BeginTx(context.Background(), &sql.TxOptions{})
	if err != nil {
		return
	}
	defer func() {
		if err == nil {
			err = tx.Commit()
			return
		}
		// don't shadow original error
		_ = tx.Rollback()
	}()
	query := `INSERT OR REPLACE INTO waku_messages (sig, timestamp, topic, payload, padding, hash, third_party_id) VALUES (?, ?, ?, ?, ?, ?, ?)`
	stmt, err := tx.Prepare(query)
	if err != nil {
		return
	}
	defer stmt.Close()
	for _, msg := range messages {
		_, err = stmt.Exec(
			msg.Sig,
			msg.Timestamp,
			msg.Topic.String(),
			msg.Payload,
			msg.Padding,
			types.Bytes2Hex(msg.Hash),
			msg.ThirdPartyID,
		)
		if err != nil {
			return
		}
	}
	return
}

func (p *Persistence) SaveWakuMessage(message *types.Message) error {
	_, err := p.db.Exec(`INSERT OR REPLACE INTO waku_messages (sig, timestamp, topic, payload, padding, hash, third_party_id) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		message.Sig,
		message.Timestamp,
		message.Topic.String(),
		message.Payload,
		message.Padding,
		types.Bytes2Hex(message.Hash),
		message.ThirdPartyID,
	)
	return err
}

func wakuMessageTimestampQuery(topics []types.TopicType) string {
	query := " FROM waku_messages WHERE "
	for i, topic := range topics {
		query += `topic = "` + topic.String() + `"`
		if i < len(topics)-1 {
			query += OR
		}
	}
	return query
}

func (p *Persistence) GetOldestWakuMessageTimestamp(topics []types.TopicType) (uint64, error) {
	var timestamp sql.NullInt64
	query := "SELECT MIN(timestamp)"
	query += wakuMessageTimestampQuery(topics)
	err := p.db.QueryRow(query).Scan(&timestamp)
	return uint64(timestamp.Int64), err
}

func (p *Persistence) GetLatestWakuMessageTimestamp(topics []types.TopicType) (uint64, error) {
	var timestamp sql.NullInt64
	query := "SELECT MAX(timestamp)"
	query += wakuMessageTimestampQuery(topics)
	err := p.db.QueryRow(query).Scan(&timestamp)
	return uint64(timestamp.Int64), err
}

func (p *Persistence) GetWakuMessagesByFilterTopic(topics []types.TopicType, from uint64, to uint64) ([]types.Message, error) {

	query := "SELECT sig, timestamp, topic, payload, padding, hash, third_party_id FROM waku_messages WHERE timestamp >= " + fmt.Sprint(from) + " AND timestamp < " + fmt.Sprint(to) + " AND ("

	for i, topic := range topics {
		query += `topic = "` + topic.String() + `"`
		if i < len(topics)-1 {
			query += OR
		}
	}
	query += ")"

	rows, err := p.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	messages := []types.Message{}

	for rows.Next() {
		msg := types.Message{}
		var topicStr string
		var hashStr string
		err := rows.Scan(&msg.Sig, &msg.Timestamp, &topicStr, &msg.Payload, &msg.Padding, &hashStr, &msg.ThirdPartyID)
		if err != nil {
			return nil, err
		}
		msg.Topic = types.StringToTopic(topicStr)
		msg.Hash = types.Hex2Bytes(hashStr)
		messages = append(messages, msg)
	}

	return messages, nil
}

func (p *Persistence) HasCommunityArchiveInfo(communityID types.HexBytes) (exists bool, err error) {
	err = p.db.QueryRow(`SELECT EXISTS(SELECT 1 FROM communities_archive_info WHERE community_id = ?)`, communityID.String()).Scan(&exists)
	return exists, err
}

func (p *Persistence) GetLastSeenMagnetlink(communityID types.HexBytes) (string, error) {
	var magnetlinkURI string
	err := p.db.QueryRow(`SELECT last_magnetlink_uri FROM communities_archive_info WHERE community_id = ?`, communityID.String()).Scan(&magnetlinkURI)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return magnetlinkURI, err
}

func (p *Persistence) GetMagnetlinkMessageClock(communityID types.HexBytes) (uint64, error) {
	var magnetlinkClock uint64
	err := p.db.QueryRow(`SELECT magnetlink_clock FROM communities_archive_info WHERE community_id = ?`, communityID.String()).Scan(&magnetlinkClock)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	return magnetlinkClock, err
}

func (p *Persistence) SaveCommunityArchiveInfo(communityID types.HexBytes, clock uint64, lastArchiveEndDate uint64) error {
	_, err := p.db.Exec(`INSERT INTO communities_archive_info (magnetlink_clock, last_message_archive_end_date, community_id) VALUES (?, ?, ?)`,
		clock,
		lastArchiveEndDate,
		communityID.String())
	return err
}

func (p *Persistence) UpdateMagnetlinkMessageClock(communityID types.HexBytes, clock uint64) error {
	_, err := p.db.Exec(`UPDATE communities_archive_info SET
    magnetlink_clock = ?
    WHERE community_id = ?`,
		clock,
		communityID.String())
	return err
}

func (p *Persistence) UpdateLastSeenMagnetlink(communityID types.HexBytes, magnetlinkURI string) error {
	_, err := p.db.Exec(`UPDATE communities_archive_info SET
    last_magnetlink_uri = ?
    WHERE community_id = ?`,
		magnetlinkURI,
		communityID.String())
	return err
}

func (p *Persistence) SaveLastMessageArchiveEndDate(communityID types.HexBytes, endDate uint64) error {
	_, err := p.db.Exec(`INSERT INTO communities_archive_info (last_message_archive_end_date, community_id) VALUES (?, ?)`,
		endDate,
		communityID.String())
	return err
}

func (p *Persistence) UpdateLastMessageArchiveEndDate(communityID types.HexBytes, endDate uint64) error {
	_, err := p.db.Exec(`UPDATE communities_archive_info SET
    last_message_archive_end_date = ?
    WHERE community_id = ?`,
		endDate,
		communityID.String())
	return err
}

func (p *Persistence) GetLastMessageArchiveEndDate(communityID types.HexBytes) (uint64, error) {

	var lastMessageArchiveEndDate uint64
	err := p.db.QueryRow(`SELECT last_message_archive_end_date FROM communities_archive_info WHERE community_id = ?`, communityID.String()).Scan(&lastMessageArchiveEndDate)
	if err == sql.ErrNoRows {
		return 0, nil
	} else if err != nil {
		return 0, err
	}
	return lastMessageArchiveEndDate, nil
}

func (p *Persistence) GetMessageArchiveIDsToImport(communityID types.HexBytes) ([]string, error) {
	rows, err := p.db.Query("SELECT hash FROM community_message_archive_hashes WHERE community_id = ? AND NOT(imported)", communityID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	ids := []string{}
	for rows.Next() {
		id := ""
		err := rows.Scan(&id)
		if err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, err
}

func (p *Persistence) GetDownloadedMessageArchiveIDs(communityID types.HexBytes) ([]string, error) {
	rows, err := p.db.Query("SELECT hash FROM community_message_archive_hashes WHERE community_id = ?", communityID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	ids := []string{}
	for rows.Next() {
		id := ""
		err := rows.Scan(&id)
		if err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, err
}

func (p *Persistence) SetMessageArchiveIDImported(communityID types.HexBytes, hash string, imported bool) error {
	_, err := p.db.Exec(`UPDATE community_message_archive_hashes SET imported = ? WHERE hash = ? AND community_id = ?`, imported, hash, communityID.String())
	return err
}

func (p *Persistence) HasMessageArchiveID(communityID types.HexBytes, hash string) (exists bool, err error) {
	err = p.db.QueryRow(`SELECT EXISTS (SELECT 1 FROM community_message_archive_hashes WHERE community_id = ? AND hash = ?)`,
		communityID.String(),
		hash,
	).Scan(&exists)
	return exists, err
}

func (p *Persistence) SaveMessageArchiveID(communityID types.HexBytes, hash string) error {
	_, err := p.db.Exec(`INSERT INTO community_message_archive_hashes (community_id, hash) VALUES (?, ?)`,
		communityID.String(),
		hash,
	)
	return err
}

func (p *Persistence) GetCommunitiesSettings() ([]CommunitySettings, error) {
	rows, err := p.db.Query("SELECT community_id, message_archive_seeding_enabled, message_archive_fetching_enabled, clock FROM communities_settings")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	communitiesSettings := []CommunitySettings{}

	for rows.Next() {
		settings := CommunitySettings{}
		err := rows.Scan(&settings.CommunityID, &settings.HistoryArchiveSupportEnabled, &settings.HistoryArchiveSupportEnabled, &settings.Clock)
		if err != nil {
			return nil, err
		}
		communitiesSettings = append(communitiesSettings, settings)
	}
	return communitiesSettings, err
}

func (p *Persistence) CommunitySettingsExist(communityID types.HexBytes) (bool, error) {
	var count int
	err := p.db.QueryRow(`SELECT count(1) FROM communities_settings WHERE community_id = ?`, communityID.String()).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (p *Persistence) GetCommunitySettingsByID(communityID types.HexBytes) (*CommunitySettings, error) {
	settings := CommunitySettings{}
	err := p.db.QueryRow(`SELECT community_id, message_archive_seeding_enabled, message_archive_fetching_enabled, clock FROM communities_settings WHERE community_id = ?`, communityID.String()).Scan(&settings.CommunityID, &settings.HistoryArchiveSupportEnabled, &settings.HistoryArchiveSupportEnabled, &settings.Clock)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	return &settings, nil
}

func (p *Persistence) DeleteCommunitySettings(communityID types.HexBytes) error {
	_, err := p.db.Exec("DELETE FROM communities_settings WHERE community_id = ?", communityID.String())
	return err
}

func (p *Persistence) SaveCommunitySettings(communitySettings CommunitySettings) error {
	_, err := p.db.Exec(`INSERT INTO communities_settings (
    community_id,
    message_archive_seeding_enabled,
    message_archive_fetching_enabled,
    clock
  ) VALUES (?, ?, ?, ?)`,
		communitySettings.CommunityID,
		communitySettings.HistoryArchiveSupportEnabled,
		communitySettings.HistoryArchiveSupportEnabled,
		communitySettings.Clock,
	)
	return err
}

func (p *Persistence) UpdateCommunitySettings(communitySettings CommunitySettings) error {
	_, err := p.db.Exec(`UPDATE communities_settings SET
    message_archive_seeding_enabled = ?,
    message_archive_fetching_enabled = ?,
    clock = ?
    WHERE community_id = ?`,
		communitySettings.HistoryArchiveSupportEnabled,
		communitySettings.HistoryArchiveSupportEnabled,
		communitySettings.Clock,
		communitySettings.CommunityID,
	)
	return err
}

func (p *Persistence) GetCommunityChatIDs(communityID types.HexBytes) ([]string, error) {
	rows, err := p.db.Query(`SELECT id FROM chats WHERE community_id = ?`, communityID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ids := []string{}
	for rows.Next() {
		id := ""
		err := rows.Scan(&id)
		if err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}

func (p *Persistence) GetAllCommunityTokens() ([]*token.CommunityToken, error) {
	rows, err := p.db.Query(`SELECT community_id, address, type, name, symbol, description, supply_str,
	infinite_supply, transferable, remote_self_destruct, chain_id, deploy_state, image_base64, decimals,
	deployer, privileges_level FROM community_tokens`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return p.getCommunityTokensInternal(rows)
}

func (p *Persistence) GetCommunityTokens(communityID string) ([]*token.CommunityToken, error) {
	rows, err := p.db.Query(`SELECT community_id, address, type, name, symbol, description, supply_str,
	infinite_supply, transferable, remote_self_destruct, chain_id, deploy_state, image_base64, decimals,
	deployer, privileges_level
	FROM community_tokens WHERE community_id = ?`, communityID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return p.getCommunityTokensInternal(rows)
}

func (p *Persistence) GetCommunityToken(communityID string, chainID int, address string) (*token.CommunityToken, error) {
	token := token.CommunityToken{}
	var supplyStr string
	err := p.db.QueryRow(`SELECT community_id, address, type, name, symbol, description, supply_str, infinite_supply,
		transferable, remote_self_destruct, chain_id, deploy_state, image_base64, decimals, deployer, privileges_level
		FROM community_tokens WHERE community_id = ? AND chain_id = ? AND address = ?`, communityID, chainID, address).Scan(&token.CommunityID, &token.Address, &token.TokenType, &token.Name,
		&token.Symbol, &token.Description, &supplyStr, &token.InfiniteSupply, &token.Transferable,
		&token.RemoteSelfDestruct, &token.ChainID, &token.DeployState, &token.Base64Image, &token.Decimals,
		&token.Deployer, &token.PrivilegesLevel)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	supplyBigInt, ok := new(big.Int).SetString(supplyStr, 10)
	if ok {
		token.Supply = &bigint.BigInt{Int: supplyBigInt}
	} else {
		token.Supply = &bigint.BigInt{Int: big.NewInt(0)}
	}
	return &token, nil
}

func (p *Persistence) getCommunityTokensInternal(rows *sql.Rows) ([]*token.CommunityToken, error) {
	tokens := []*token.CommunityToken{}

	for rows.Next() {
		token := token.CommunityToken{}
		var supplyStr string
		err := rows.Scan(&token.CommunityID, &token.Address, &token.TokenType, &token.Name,
			&token.Symbol, &token.Description, &supplyStr, &token.InfiniteSupply, &token.Transferable,
			&token.RemoteSelfDestruct, &token.ChainID, &token.DeployState, &token.Base64Image, &token.Decimals,
			&token.Deployer, &token.PrivilegesLevel)
		if err != nil {
			return nil, err
		}
		supplyBigInt, ok := new(big.Int).SetString(supplyStr, 10)
		if ok {
			token.Supply = &bigint.BigInt{Int: supplyBigInt}
		} else {
			token.Supply = &bigint.BigInt{Int: big.NewInt(0)}
		}

		tokens = append(tokens, &token)
	}
	return tokens, nil
}

func (p *Persistence) HasCommunityToken(communityID string, address string, chainID int) (bool, error) {
	var count int
	err := p.db.QueryRow(`SELECT count(1) FROM community_tokens WHERE community_id = ? AND address = ? AND chain_id = ?`, communityID, address, chainID).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (p *Persistence) AddCommunityToken(token *token.CommunityToken) error {
	_, err := p.db.Exec(`INSERT INTO community_tokens (community_id, address, type, name, symbol, description, supply_str,
		infinite_supply, transferable, remote_self_destruct, chain_id, deploy_state, image_base64, decimals, deployer, privileges_level)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, token.CommunityID, token.Address, token.TokenType, token.Name,
		token.Symbol, token.Description, token.Supply.String(), token.InfiniteSupply, token.Transferable, token.RemoteSelfDestruct,
		token.ChainID, token.DeployState, token.Base64Image, token.Decimals, token.Deployer, token.PrivilegesLevel)
	return err
}

func (p *Persistence) UpdateCommunityTokenState(chainID int, contractAddress string, deployState token.DeployState) error {
	_, err := p.db.Exec(`UPDATE community_tokens SET deploy_state = ? WHERE address = ? AND chain_id = ?`, deployState, contractAddress, chainID)
	return err
}

func (p *Persistence) UpdateCommunityTokenAddress(chainID int, oldContractAddress string, newContractAddress string) error {
	_, err := p.db.Exec(`UPDATE community_tokens SET address = ? WHERE address = ? AND chain_id = ?`, newContractAddress, oldContractAddress, chainID)
	return err
}

func (p *Persistence) UpdateCommunityTokenSupply(chainID int, contractAddress string, supply *bigint.BigInt) error {
	_, err := p.db.Exec(`UPDATE community_tokens SET supply_str = ? WHERE address = ? AND chain_id = ?`, supply.String(), contractAddress, chainID)
	return err
}

func (p *Persistence) RemoveCommunityToken(chainID int, contractAddress string) error {
	_, err := p.db.Exec(`DELETE FROM community_tokens WHERE chain_id = ? AND address = ?`, chainID, contractAddress)
	return err
}

func decodeWrappedCommunityDescription(wrappedDescriptionBytes []byte) (*protobuf.CommunityDescription, error) {
	metadata := &protobuf.ApplicationMetadataMessage{}

	err := proto.Unmarshal(wrappedDescriptionBytes, metadata)
	if err != nil {
		return nil, err
	}

	description := &protobuf.CommunityDescription{}

	err = proto.Unmarshal(metadata.Payload, description)
	if err != nil {
		return nil, err
	}

	return description, nil
}

func decodeEventsData(eventsBytes []byte, eventsDescriptionBytes []byte) (*EventsData, error) {
	if len(eventsDescriptionBytes) == 0 {
		return nil, nil
	}
	var events []CommunityEvent
	if eventsBytes != nil {
		var err error
		events, err = communityEventsFromJSONEncodedBytes(eventsBytes)
		if err != nil {
			return nil, err
		}
	}

	return &EventsData{
		EventsBaseCommunityDescription: eventsDescriptionBytes,
		Events:                         events,
	}, nil
}

func (p *Persistence) GetCommunityRequestsToJoinWithRevealedAddresses(communityID []byte) ([]*RequestToJoin, error) {
	requests := []*RequestToJoin{}
	rows, err := p.db.Query(`
	SELECT r.id, r.public_key, r.clock, r.ens_name, r.chat_id, r.state, r.community_id,
		a.address, a.chain_ids, a.is_airdrop_address, a.signature
	FROM communities_requests_to_join r
	LEFT JOIN communities_requests_to_join_revealed_addresses a ON r.id = a.request_id
	WHERE r.community_id = ? AND r.state != ?`, communityID, RequestToJoinStateAwaitingAddresses)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	prevRequest := &RequestToJoin{}
	for rows.Next() {
		request := &RequestToJoin{}
		var address sql.NullString
		var chainIDsStr sql.NullString
		var isAirdropAddress sql.NullBool
		var signature sql.RawBytes

		err = rows.Scan(&request.ID, &request.PublicKey, &request.Clock, &request.ENSName, &request.ChatID, &request.State, &request.CommunityID,
			&address, &chainIDsStr, &isAirdropAddress, &signature)
		if err != nil {
			return nil, err
		}

		revealedAccount, err := toRevealedAccount(address, chainIDsStr, isAirdropAddress, signature)
		if err != nil {
			return nil, err
		}

		if types.EncodeHex(prevRequest.ID) == types.EncodeHex(request.ID) {
			if revealedAccount != nil {
				prevRequest.RevealedAccounts = append(prevRequest.RevealedAccounts, revealedAccount)
			}
		} else {
			if revealedAccount != nil {
				request.RevealedAccounts = []*protobuf.RevealedAccount{
					revealedAccount,
				}
			}
			requests = append(requests, request)
			prevRequest = request
		}
	}

	return requests, nil
}

func toRevealedAccount(rawAddress sql.NullString, rawChainIDsStr sql.NullString, isAirdropAddress sql.NullBool, rawSignature sql.RawBytes) (*protobuf.RevealedAccount, error) {
	if !rawAddress.Valid {
		return nil, nil
	}

	address := rawAddress.String
	if address == "" {
		return nil, nil
	}

	chainIDsStr := ""
	if rawChainIDsStr.Valid {
		chainIDsStr = rawChainIDsStr.String
	}

	chainIDs := make([]uint64, 0)
	for _, chainIDstr := range strings.Split(chainIDsStr, ",") {
		if chainIDstr != "" {
			chainID, err := strconv.Atoi(chainIDstr)
			if err != nil {
				return nil, err
			}
			chainIDs = append(chainIDs, uint64(chainID))
		}
	}

	revealedAccount := &protobuf.RevealedAccount{
		Address:          address,
		ChainIds:         chainIDs,
		IsAirdropAddress: false,
		Signature:        rawSignature,
	}
	if isAirdropAddress.Valid {
		revealedAccount.IsAirdropAddress = isAirdropAddress.Bool
	}
	return revealedAccount, nil
}

type communityToValidate struct {
	id         []byte
	clock      uint64
	payload    []byte
	validateAt uint64
	signer     []byte
}

func (p *Persistence) SaveCommunityToValidate(c communityToValidate) error {
	_, err := p.db.Exec(`
		INSERT INTO communities_validate_signer (id, clock, payload, validate_at, signer) VALUES (?, ?, ?, ?, ?)`, c.id, c.clock, c.payload, c.validateAt, c.signer)
	return err
}

func (p *Persistence) getCommunitiesToValidateCount() (int, error) {

	var count int
	qr := p.db.QueryRow(`SELECT COUNT(1) FROM communities_validate_signer`)
	err := qr.Scan(&count)

	return count, err

}

func (p *Persistence) getCommunitiesToValidate() (map[string][]communityToValidate, error) {
	communitiesToValidate := make(map[string][]communityToValidate)
	rows, err := p.db.Query(`SELECT id, clock, payload, signer FROM communities_validate_signer WHERE validate_at <= ? ORDER BY clock DESC`, time.Now().UnixNano())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		communityToValidate := communityToValidate{}
		err := rows.Scan(&communityToValidate.id, &communityToValidate.clock, &communityToValidate.payload, &communityToValidate.signer)
		if err != nil {
			return nil, err
		}
		communitiesToValidate[types.EncodeHex(communityToValidate.id)] = append(communitiesToValidate[types.EncodeHex(communityToValidate.id)], communityToValidate)
	}

	return communitiesToValidate, nil

}

func (p *Persistence) getCommunityToValidateByID(communityID types.HexBytes) ([]communityToValidate, error) {
	communityToValidateArray := []communityToValidate{}
	rows, err := p.db.Query(`SELECT id, clock, payload, signer FROM communities_validate_signer WHERE id = ? AND validate_at <= ? ORDER BY clock DESC`, communityID, time.Now().UnixNano())

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		communityToValidate := communityToValidate{}
		err := rows.Scan(&communityToValidate.id, &communityToValidate.clock, &communityToValidate.payload, &communityToValidate.signer)
		if err != nil {
			return nil, err
		}

		communityToValidateArray = append(communityToValidateArray, communityToValidate)
	}

	return communityToValidateArray, nil

}

func (p *Persistence) DeleteCommunitiesToValidateByCommunityID(communityID []byte) error {
	_, err := p.db.Exec(`DELETE FROM communities_validate_signer WHERE id = ?`, communityID)
	return err
}

func (p *Persistence) DeleteCommunityToValidate(communityID []byte, clock uint64) error {
	_, err := p.db.Exec(`DELETE FROM communities_validate_signer WHERE id = ? AND clock = ?`, communityID, clock)
	return err
}

func (p *Persistence) GetSyncControlNode(communityID types.HexBytes) (*protobuf.SyncCommunityControlNode, error) {
	result := &protobuf.SyncCommunityControlNode{}

	err := p.db.QueryRow(`
        SELECT clock, installation_id
        FROM communities_control_node
        WHERE community_id = ?
    `, communityID).Scan(&result.Clock, &result.InstallationId)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return result, nil
}

func (p *Persistence) SaveSyncControlNode(communityID types.HexBytes, clock uint64, installationID string) error {
	_, err := p.db.Exec(
		`INSERT INTO communities_control_node (
			community_id,
			clock,
			installation_id
		) VALUES (?, ?, ?)`,
		communityID,
		clock,
		installationID,
	)
	return err
}

func (p *Persistence) GetCommunityRequestToJoinWithRevealedAddresses(pubKey string, communityID []byte) (*RequestToJoin, error) {
	requestToJoin, err := p.GetRequestToJoinByPkAndCommunityID(pubKey, communityID)
	if err != nil {
		return nil, err
	}

	revealedAccounts, err := p.GetRequestToJoinRevealedAddresses(requestToJoin.ID)
	if err != nil {
		return nil, err
	}

	requestToJoin.RevealedAccounts = revealedAccounts

	return requestToJoin, nil
}

func (p *Persistence) SaveRequestsToJoin(requests []*RequestToJoin) (err error) {
	tx, err := p.db.BeginTx(context.Background(), &sql.TxOptions{})
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			// Rollback the transaction on error
			_ = tx.Rollback()
		}
	}()

	stmt, err := tx.Prepare(`INSERT OR REPLACE INTO communities_requests_to_join(id,public_key,clock,ens_name,chat_id,community_id,state)
		VALUES (?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, request := range requests {
		var clock uint64
		// Fetch any existing request to join
		err = tx.QueryRow(`SELECT clock FROM communities_requests_to_join WHERE public_key = ? AND community_id = ?`, request.PublicKey, request.CommunityID).Scan(&clock)
		if err != nil && err != sql.ErrNoRows {
			return err
		}

		if clock >= request.Clock {
			return ErrOldRequestToJoin
		}

		_, err = stmt.Exec(request.ID, request.PublicKey, request.Clock, request.ENSName, request.ChatID, request.CommunityID, request.State)
		if err != nil {
			return err
		}
	}

	err = tx.Commit()
	return err
}

func (p *Persistence) GetCuratedCommunities() (*CuratedCommunities, error) {
	rows, err := p.db.Query("SELECT community_id, featured FROM curated_communities")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := &CuratedCommunities{
		ContractCommunities:         []string{},
		ContractFeaturedCommunities: []string{},
	}
	for rows.Next() {
		var communityID string
		var featured bool
		if err := rows.Scan(&communityID, &featured); err != nil {
			return nil, err
		}
		result.ContractCommunities = append(result.ContractCommunities, communityID)
		if featured {
			result.ContractFeaturedCommunities = append(result.ContractFeaturedCommunities, communityID)
		}
	}

	return result, nil
}

func (p *Persistence) SetCuratedCommunities(communities *CuratedCommunities) error {
	tx, err := p.db.BeginTx(context.Background(), &sql.TxOptions{})
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

	// Clear the existing communities
	if _, err = tx.Exec("DELETE FROM curated_communities"); err != nil {
		return err
	}

	stmt, err := tx.Prepare("INSERT INTO curated_communities (community_id, featured) VALUES (?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	featuredMap := make(map[string]bool)
	for _, community := range communities.ContractFeaturedCommunities {
		featuredMap[community] = true
	}

	for _, community := range communities.ContractCommunities {
		_, err := stmt.Exec(community, featuredMap[community])
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *Persistence) AllNonApprovedCommunitiesRequestsToJoin() ([]*RequestToJoin, error) {
	nonApprovedRequestsToJoin := []*RequestToJoin{}
	rows, err := p.db.Query(`SELECT id,public_key,clock,ens_name,chat_id,community_id,state FROM communities_requests_to_join WHERE state != ?`, RequestToJoinStateAccepted)

	if err == sql.ErrNoRows {
		return nonApprovedRequestsToJoin, nil
	} else if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		request := &RequestToJoin{}
		err := rows.Scan(&request.ID, &request.PublicKey, &request.Clock, &request.ENSName, &request.ChatID, &request.CommunityID, &request.State)
		if err != nil {
			return nil, err
		}
		nonApprovedRequestsToJoin = append(nonApprovedRequestsToJoin, request)
	}
	return nonApprovedRequestsToJoin, nil
}

func (p *Persistence) RemoveAllCommunityRequestsToJoinWithRevealedAddressesExceptPublicKey(pk string, communityID []byte) error {
	_, err := p.db.Exec(`
	DELETE FROM communities_requests_to_join_revealed_addresses
		WHERE request_id IN (SELECT id FROM communities_requests_to_join WHERE community_id = ? AND public_key != ?);
	DELETE FROM communities_requests_to_join
		WHERE community_id = ? AND public_key != ?;`, communityID, pk, communityID, pk)
	return err
}

func (p *Persistence) SaveCommunityShard(communityID types.HexBytes, shard *shard.Shard, clock uint64) error {
	var cluster, index *uint16

	if shard != nil {
		cluster = &shard.Cluster
		index = &shard.Index
	}

	result, err := p.db.Exec(`
		INSERT INTO communities_shards (community_id, shard_cluster, shard_index, clock)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(community_id)
		DO UPDATE SET
			shard_cluster = CASE WHEN excluded.clock > communities_shards.clock THEN excluded.shard_cluster ELSE communities_shards.shard_cluster END,
			shard_index = CASE WHEN excluded.clock > communities_shards.clock THEN excluded.shard_index ELSE communities_shards.shard_index END,
			clock = CASE WHEN excluded.clock > communities_shards.clock THEN excluded.clock ELSE communities_shards.clock END
		WHERE excluded.clock > communities_shards.clock OR communities_shards.community_id IS NULL`,
		communityID, cluster, index, clock)

	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrOldShardInfo
	}
	return nil
}

// if data will not be found, will return sql.ErrNoRows. Must be handled on the caller side
func (p *Persistence) GetCommunityShard(communityID types.HexBytes) (*shard.Shard, error) {
	var cluster sql.NullInt64
	var index sql.NullInt64
	err := p.db.QueryRow(`SELECT shard_cluster, shard_index FROM communities_shards WHERE community_id = ?`,
		communityID).Scan(&cluster, &index)

	if err != nil {
		return nil, err
	}

	if !cluster.Valid || !index.Valid {
		return nil, nil
	}

	return &shard.Shard{
		Cluster: uint16(cluster.Int64),
		Index:   uint16(index.Int64),
	}, nil
}

func (p *Persistence) DeleteCommunityShard(communityID types.HexBytes) error {
	_, err := p.db.Exec(`DELETE FROM communities_shards WHERE community_id = ?`, communityID)
	return err
}
