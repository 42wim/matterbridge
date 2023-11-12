-- Query to retrive all recipients for selected addresses and networks
WITH filter_conditions AS (
	SELECT 
		? AS filterAllAddresses,
		? AS includeAllNetworks,
		? AS pendingStatus
),
filter_addresses(address) AS (
	VALUES 
		%s
),
filter_networks(network_id) AS (
	VALUES
		%s
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
)
SELECT
	to_address,
	MIN(timestamp) AS min_timestamp
FROM (
	SELECT
		transfers.tx_to_address as to_address,
		MIN(transfers.timestamp) AS timestamp
	FROM
		transfers, filter_conditions
	WHERE
		transfers.multi_transaction_id = 0 AND transfers.tx_to_address NOT NULL
		AND (filterAllAddresses OR transfers.address IN filter_addresses)
		AND (includeAllNetworks OR transfers.network_id IN filter_networks)
	GROUP BY
		transfers.tx_to_address

	UNION

	SELECT
		pending_transactions.to_address AS to_address,
		MIN(pending_transactions.timestamp) AS timestamp
	FROM
		pending_transactions, filter_conditions
	WHERE
		pending_transactions.multi_transaction_id = 0 AND pending_transactions.to_address NOT NULL
		AND (filterAllAddresses OR pending_transactions.from_address IN filter_addresses)
		AND (includeAllNetworks OR pending_transactions.network_id IN filter_networks)
	GROUP BY
		pending_transactions.to_address

	UNION

	SELECT
		multi_transactions.to_address AS to_address,
		MIN(multi_transactions.timestamp) AS timestamp
	FROM
		multi_transactions, filter_conditions
	WHERE
		(filterAllAddresses OR multi_transactions.from_address IN filter_addresses)
		AND (
			includeAllNetworks
			OR (multi_transactions.from_network_id IN filter_networks)
			OR (multi_transactions.to_network_id IN filter_networks)
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
	GROUP BY
		multi_transactions.to_address
) AS combined_result
GROUP BY
	to_address
ORDER BY
	min_timestamp DESC
LIMIT ? OFFSET ?;