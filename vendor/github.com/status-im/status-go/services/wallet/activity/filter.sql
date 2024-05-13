-- Query includes duplicates, will return multiple rows for the same transaction if both to and from addresses are in the address list.
--
-- The switch for tr_type is used to de-conflict the source for the two entries for the same transaction
--
-- UNION ALL is used to avoid the overhead of DISTINCT given that we don't expect to have duplicate entries outside the sender and receiver addresses being in the list which is handled separately
--
-- Only status FailedAS, PendingAS and CompleteAS are returned. FinalizedAS requires correlation with blockchain current state. As an optimization we approximate it by using timestamp information; see startTimestamp and endTimestamp
--
-- ContractDeploymentAT is subtype of SendAT and MintAT is subtype of ReceiveAT. It means query must prevent returning MintAT when filtering by ReceiveAT or ContractDeploymentAT when filtering by SendAT. That required duplicated code in filter by type query, to maintain performance.
--
-- Token filtering has two parts
-- 1. Filtering by symbol (multi_transactions and pending_transactions tables) where the chain ID is ignored, basically the filter_networks will account for that
-- 2. Filtering by token identity (chain and address for transfers table) where the symbol is ignored and all the token identities must be provided
--
WITH filter_conditions AS (
	SELECT
		? AS startFilterDisabled,
		? AS startTimestamp,
		? AS endFilterDisabled,
		? AS endTimestamp,
		? AS filterActivityTypeAll,
		? AS filterActivityTypeSend,
		? AS filterActivityTypeReceive,
		? AS filterActivityTypeContractDeployment,
		? AS filterActivityTypeMint,
		? AS mTTypeSend,
		? AS fromTrType,
		? AS toTrType,
		? AS filterAllAddresses,
		? AS filterAllToAddresses,
		? AS filterAllActivityStatus,
		? AS filterStatusCompleted,
		? AS filterStatusFailed,
		? AS filterStatusFinalized,
		? AS filterStatusPending,
		? AS statusFailed,
		? AS statusCompleted,
		? AS statusFinalized,
		? AS statusPending,
		? AS includeAllTokenTypeAssets,
		? AS includeAllCollectibles,
		? AS includeAllNetworks,
		? AS pendingStatus,
		? AS nowTimestamp,
		? AS layer2FinalisationDuration,
		? AS layer1FinalisationDuration,
		X'0000000000000000000000000000000000000000' AS zeroAddress,
		'0x28c427b0611d99da5c4f7368abe57e86b045b483c4689ae93e90745802335b87' as communityMintEvent
),
-- This UNION between CTE and TEMP TABLE acts as an optimization. As soon as we drop one or use them interchangeably the performance drops significantly.
filter_addresses(address) AS (
	SELECT
		address
	FROM
		filter_addresses_table
	WHERE
		(
			SELECT
				filterAllAddresses
			FROM
				filter_conditions
		) != 0
	UNION
	ALL
	SELECT
		*
	FROM
		(
			VALUES
				%s
		)
	WHERE
		(
			SELECT
				filterAllAddresses
			FROM
				filter_conditions
		) = 0
),
filter_to_addresses(address) AS (
	VALUES
		%s
),
assets_token_codes(token_code) AS (
	VALUES
		%s
),
assets_erc20(chain_id, token_address) AS (
	VALUES
		%s
),
assets_erc721(chain_id, token_id, token_address) AS (
	VALUES
		%s
),
filter_networks(network_id) AS (
	VALUES
		%s
),
tr_status AS (
	SELECT
		multi_transaction_id,
		MIN(status) AS min_status,
		COUNT(*) AS count,
		network_id
	FROM
		transfers
	WHERE
		transfers.loaded == 1
		AND transfers.multi_transaction_id != 0
	GROUP BY
		transfers.multi_transaction_id
),
tr_network_ids AS (
	SELECT
		multi_transaction_id
	FROM
		transfers
	WHERE
		transfers.loaded == 1
		AND transfers.multi_transaction_id != 0
		AND network_id IN filter_networks
	GROUP BY
		transfers.multi_transaction_id
),
pending_status AS (
	SELECT
		multi_transaction_id,
		COUNT(*) AS count,
		network_id
	FROM
		pending_transactions,
		filter_conditions
	WHERE
		pending_transactions.multi_transaction_id != 0
		AND pending_transactions.status = pendingStatus
	GROUP BY
		pending_transactions.multi_transaction_id
),
pending_network_ids AS (
	SELECT
		multi_transaction_id
	FROM
		pending_transactions,
		filter_conditions
	WHERE
		pending_transactions.multi_transaction_id != 0
		AND pending_transactions.status = pendingStatus
		AND pending_transactions.network_id IN filter_networks
	GROUP BY
		pending_transactions.multi_transaction_id
),
layer2_networks(network_id) AS (
	VALUES
		%s
),
mint_methods(method_hash) AS (
	%s
)

SELECT
	transfers.hash AS transfer_hash,
	NULL AS pending_hash,
	transfers.network_id AS network_id,
	0 AS multi_tx_id,
	transfers.timestamp AS timestamp,
	NULL AS mt_type,
	CASE
		WHEN from_join.address IS NOT NULL AND to_join.address IS NULL THEN fromTrType
		WHEN to_join.address IS NOT NULL AND from_join.address IS NULL THEN toTrType
		WHEN from_join.address IS NOT NULL AND to_join.address IS NOT NULL THEN
			CASE
				WHEN transfers.address = transfers.tx_from_address THEN fromTrType
				WHEN transfers.address = transfers.tx_to_address THEN toTrType
				ELSE NULL
			END
		ELSE NULL
	END as tr_type,
	transfers.tx_from_address AS from_address,
	transfers.tx_to_address AS to_address,
	transfers.address AS owner_address,
	transfers.amount_padded128hex AS tr_amount,
	NULL AS ptr_amount,
	NULL AS mt_from_amount,
	NULL AS mt_to_amount,
	CASE
		WHEN transfers.status IS 1 THEN CASE
			WHEN transfers.timestamp > 0
			AND filter_conditions.nowTimestamp >= transfers.timestamp + (
				CASE
					WHEN transfers.network_id in layer2_networks THEN layer2FinalisationDuration
					ELSE layer1FinalisationDuration
				END
			) THEN statusFinalized
			ELSE statusCompleted
		END
		ELSE statusFailed
	END AS agg_status,
	1 AS agg_count,
	transfers.token_address AS token_address,
	CASE
		WHEN LENGTH(transfers.token_id) = 0 THEN X'00'
		ELSE transfers.token_id
	END AS tmp_token_id,
	NULL AS token_code,
	NULL AS from_token_code,
	NULL AS to_token_code,
	NULL AS out_network_id,
	NULL AS in_network_id,
	transfers.type AS type,
	transfers.contract_address AS contract_address,
	CASE
		WHEN transfers.tx_from_address = zeroAddress AND transfers.type = "erc20" THEN substr(json_extract(tx, '$.input'), 1, 10)
		ELSE NULL
	END AS method_hash,
	CASE
		WHEN transfers.tx_from_address = zeroAddress AND transfers.type = "erc20" THEN (SELECT 1 FROM json_each(transfers.receipt, '$.logs' ) WHERE json_extract( value, '$.topics[0]' ) = communityMintEvent)
		ELSE NULL
	END AS community_mint_event,
	CASE 
		WHEN transfers.type = 'erc20' THEN (SELECT community_id FROM tokens WHERE transfers.token_address = tokens.address AND transfers.network_id = tokens.network_id)
		WHEN transfers.type = 'erc721' OR transfers.type = 'erc1155' THEN (SELECT community_id FROM collectible_data_cache WHERE transfers.token_address = collectible_data_cache.contract_address AND transfers.network_id = collectible_data_cache.chain_id)
		ELSE NULL
	END AS community_id
FROM
	transfers
	CROSS JOIN filter_conditions
	LEFT JOIN filter_addresses from_join ON transfers.tx_from_address = from_join.address
	LEFT JOIN filter_addresses to_join ON transfers.tx_to_address = to_join.address
WHERE
	transfers.loaded == 1
	AND transfers.multi_transaction_id = 0
	AND (
		(
			startFilterDisabled
			OR transfers.timestamp >= startTimestamp
		)
		AND (
			endFilterDisabled
			OR transfers.timestamp <= endTimestamp
		)
	)
	AND (
		-- Check description at the top of the file why code below is duplicated
		filterActivityTypeAll
		OR (
			filterActivityTypeSend
			AND tr_type = fromTrType -- Check NOT ContractDeploymentAT
			AND NOT (
				transfers.tx_to_address IS NULL
				AND transfers.type = 'eth'
				AND transfers.contract_address IS NOT NULL
				AND transfers.contract_address != zeroAddress
			)
		)
		OR (
			filterActivityTypeReceive
			AND tr_type = toTrType -- Check NOT MintAT
			AND NOT (
				(
					transfers.tx_from_address IS NULL
					OR transfers.tx_from_address = zeroAddress
				)
				AND (
					transfers.type = 'erc721'
					OR (
						transfers.type = 'erc20'
						AND (
							(method_hash IS NOT NULL AND method_hash IN mint_methods)
							OR community_mint_event IS NOT NULL
						)
					)
				)
			)
		)
		OR (
			filterActivityTypeContractDeployment
			AND tr_type = fromTrType
			AND transfers.tx_to_address IS NULL
			AND transfers.type = 'eth'
			AND transfers.contract_address IS NOT NULL
			AND transfers.contract_address != zeroAddress
		)
		OR (
			filterActivityTypeMint
			AND tr_type = toTrType
			AND (
				transfers.tx_from_address IS NULL
				OR transfers.tx_from_address = zeroAddress
			)
			AND (
				transfers.type = 'erc721'
				OR (
					transfers.type = 'erc20'
					AND (
						(method_hash IS NOT NULL AND method_hash IN mint_methods)
						OR community_mint_event IS NOT NULL
					)
				)
			)
		)
	)
	AND (
		filterAllAddresses -- Every account address has an "owned" entry either as to or from
		OR (owner_address IN filter_addresses)
	)
	AND (
		filterAllToAddresses
		OR (transfers.tx_to_address IN filter_to_addresses)
	)
	AND (
		includeAllTokenTypeAssets
		OR (
			transfers.type = 'eth'
			AND ('ETH' IN assets_token_codes)
		)
		OR (
			transfers.type = 'erc20'
			AND (
				(
					transfers.network_id,
					transfers.token_address
				) IN assets_erc20
			)
		)
	)
	AND (
		includeAllCollectibles
		OR (
			transfers.type = "erc721"
			AND (
				(
					transfers.network_id,
					tmp_token_id,
					transfers.token_address
				) IN assets_erc721
			)
		)
	)
	AND (
		includeAllNetworks
		OR (transfers.network_id IN filter_networks)
	)
	AND (
		filterAllActivityStatus
		OR (
			filterStatusCompleted
			AND agg_status = statusCompleted
		)
		OR (
			filterStatusFinalized
			AND agg_status = statusFinalized
		)
		OR (
			filterStatusFailed
			AND agg_status = statusFailed
		)
	)
UNION
ALL
SELECT
	NULL AS transfer_hash,
	pending_transactions.hash AS pending_hash,
	pending_transactions.network_id AS network_id,
	0 AS multi_tx_id,
	pending_transactions.timestamp AS timestamp,
	NULL AS mt_type,
	CASE
		WHEN from_join.address IS NOT NULL AND to_join.address IS NULL THEN fromTrType
		WHEN to_join.address IS NOT NULL AND from_join.address IS NULL THEN toTrType
		WHEN from_join.address IS NOT NULL AND to_join.address IS NOT NULL THEN fromTrType
		ELSE NULL
	END as tr_type,
	pending_transactions.from_address AS from_address,
	pending_transactions.to_address AS to_address,
	NULL AS owner_address,
	NULL AS tr_amount,
	pending_transactions.value AS ptr_amount,
	NULL AS mt_from_amount,
	NULL AS mt_to_amount,
	statusPending AS agg_status,
	1 AS agg_count,
	NULL AS token_address,
	NULL AS tmp_token_id,
	pending_transactions.symbol AS token_code,
	NULL AS from_token_code,
	NULL AS to_token_code,
	NULL AS out_network_id,
	NULL AS in_network_id,
	pending_transactions.type AS type,
	NULL as contract_address,
	NULL AS method_hash,
	NULL AS community_mint_event,
	NULL AS community_id
FROM
	pending_transactions
	CROSS JOIN filter_conditions
	LEFT JOIN filter_addresses from_join ON pending_transactions.from_address = from_join.address
	LEFT JOIN filter_addresses to_join ON pending_transactions.to_address = to_join.address
WHERE
	pending_transactions.multi_transaction_id = 0
	AND pending_transactions.status = pendingStatus
	AND (
		filterAllActivityStatus
		OR filterStatusPending
	)
	AND includeAllCollectibles
	AND (
		(
			startFilterDisabled
			OR timestamp >= startTimestamp
		)
		AND (
			endFilterDisabled
			OR timestamp <= endTimestamp
		)
	)
	AND (
		filterActivityTypeAll
		OR filterActivityTypeSend
	)
	AND (
		filterAllAddresses
		OR tr_type NOT NULL
	)
	AND (
		filterAllToAddresses
		OR (pending_transactions.to_address IN filter_to_addresses)
	)
	AND (
		includeAllTokenTypeAssets
		OR (
			UPPER(pending_transactions.symbol) IN assets_token_codes
		)
	)
	AND (
		includeAllNetworks
		OR (
			pending_transactions.network_id IN filter_networks
		)
	)
UNION
ALL
SELECT
	NULL AS transfer_hash,
	NULL AS pending_hash,
	NULL AS network_id,
	multi_transactions.ROWID AS multi_tx_id,
	multi_transactions.timestamp AS timestamp,
	multi_transactions.type AS mt_type,
	NULL as tr_type,
	multi_transactions.from_address AS from_address,
	multi_transactions.to_address AS to_address,
	multi_transactions.from_address AS owner_address,
	NULL AS tr_amount,
	NULL AS ptr_amount,
	multi_transactions.from_amount AS mt_from_amount,
	multi_transactions.to_amount AS mt_to_amount,
	CASE
		WHEN tr_status.min_status = 1
		AND COALESCE(pending_status.count, 0) = 0 THEN CASE
			WHEN multi_transactions.timestamp > 0
			AND filter_conditions.nowTimestamp >= multi_transactions.timestamp + (
				CASE
					WHEN multi_transactions.from_network_id in layer2_networks
					OR multi_transactions.to_network_id in layer2_networks THEN layer2FinalisationDuration
					ELSE layer1FinalisationDuration
				END
			) THEN statusFinalized
			ELSE statusCompleted
		END
		WHEN tr_status.min_status = 0 THEN statusFailed
		ELSE statusPending
	END AS agg_status,
	COALESCE(tr_status.count, 0) + COALESCE(pending_status.count, 0) AS agg_count,
	NULL AS token_address,
	NULL AS tmp_token_id,
	NULL AS token_code,
	multi_transactions.from_asset AS from_token_code,
	multi_transactions.to_asset AS to_token_code,
	multi_transactions.from_network_id AS out_network_id,
	multi_transactions.to_network_id AS in_network_id,
	NULL AS type,
	NULL as contract_address,
	NULL AS method_hash,
	NULL AS community_mint_event,
	NULL AS community_id
FROM
	multi_transactions
	CROSS JOIN filter_conditions
	LEFT JOIN tr_status ON multi_transactions.ROWID = tr_status.multi_transaction_id
	LEFT JOIN pending_status ON multi_transactions.ROWID = pending_status.multi_transaction_id
WHERE
	(
		(
			startFilterDisabled
			OR multi_transactions.timestamp >= startTimestamp
		)
		AND (
			endFilterDisabled
			OR multi_transactions.timestamp <= endTimestamp
		)
	)
	AND includeAllCollectibles
	AND (
		filterActivityTypeAll
		OR (multi_transactions.type IN (%s))
	)
	AND (
		filterAllAddresses
		OR (
			-- Send multi-transaction types are exclusively for outbound transfers. The receiving end will have a corresponding entry as "owner_address" in the transfers table.
			mt_type = mTTypeSend
			AND owner_address IN filter_addresses
		)
		OR (
			mt_type != mTTypeSend
			AND (
				multi_transactions.from_address IN filter_addresses
				OR multi_transactions.to_address IN filter_addresses
			)
		)
	)
	AND (
		filterAllToAddresses
		OR (multi_transactions.to_address IN filter_to_addresses)
	)
	AND (
		includeAllTokenTypeAssets
		OR (
			multi_transactions.from_asset != ''
			AND (
				UPPER(multi_transactions.from_asset) IN assets_token_codes
			)
		)
		OR (
			multi_transactions.to_asset != ''
			AND (
				UPPER(multi_transactions.to_asset) IN assets_token_codes
			)
		)
	)
	AND (
		filterAllActivityStatus
		OR (
			filterStatusCompleted
			AND agg_status = statusCompleted
		)
		OR (
			filterStatusFinalized
			AND agg_status = statusFinalized
		)
		OR (
			filterStatusFailed
			AND agg_status = statusFailed
		)
		OR (
			filterStatusPending
			AND agg_status = statusPending
		)
	)
	AND (
		includeAllNetworks
		OR (
			multi_transactions.from_network_id IN filter_networks
		)
		OR (
			multi_transactions.to_network_id IN filter_networks
		)
		OR (
			COALESCE(multi_transactions.from_network_id, 0) = 0
			AND COALESCE(multi_transactions.to_network_id, 0) = 0
			AND (
				EXISTS (
					SELECT
						1
					FROM
						tr_network_ids
					WHERE
						multi_transactions.ROWID = tr_network_ids.multi_transaction_id
				)
				OR EXISTS (
					SELECT
						1
					FROM
						pending_network_ids
					WHERE
						multi_transactions.ROWID = pending_network_ids.multi_transaction_id
				)
			)
		)
	)
ORDER BY
	timestamp DESC
LIMIT
	? OFFSET ?
