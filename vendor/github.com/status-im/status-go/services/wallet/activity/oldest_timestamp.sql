-- Get oldest timestamp query

WITH filter_conditions AS 
	(SELECT ? AS filterAllAddresses),
	filter_addresses(address) AS (
		SELECT * FROM (VALUES %s) WHERE (SELECT filterAllAddresses FROM filter_conditions) = 0
	)
SELECT
	transfers.tx_from_address AS from_address,
	transfers.tx_to_address AS to_address,
	transfers.timestamp AS timestamp
FROM transfers, filter_conditions
WHERE transfers.multi_transaction_id = 0
	AND (filterAllAddresses OR from_address IN filter_addresses OR to_address IN filter_addresses)

UNION ALL

SELECT
	pending_transactions.from_address AS from_address,
	pending_transactions.to_address AS to_address,
	pending_transactions.timestamp AS timestamp
FROM pending_transactions, filter_conditions
WHERE pending_transactions.multi_transaction_id = 0
	AND (filterAllAddresses OR from_address IN filter_addresses OR to_address IN filter_addresses)

UNION ALL

SELECT
	multi_transactions.from_address AS from_address,
	multi_transactions.to_address AS to_address,
	multi_transactions.timestamp AS timestamp
FROM multi_transactions, filter_conditions
WHERE filterAllAddresses OR from_address IN filter_addresses OR to_address IN filter_addresses
ORDER BY timestamp ASC
LIMIT 1