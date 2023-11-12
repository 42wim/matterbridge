-- Query for getting collectibles from transfers
-- It can be filtered by owner addresses and networks
WITH filter_conditions AS (
		SELECT 
			? AS filterAllAddresses,
			? AS includeAllNetworks
	),
	owner_addresses(address) AS (
		VALUES 
			%s
	),
	filter_networks(network_id) AS (
		VALUES
			%s
	)
SELECT network_id, token_address, token_id 
FROM 
	transfers, filter_conditions
WHERE 
token_id IS NOT NULL
AND token_address IS NOT NULL 
AND (
	filterAllAddresses 
	OR tx_from_address IN owner_addresses 
	OR tx_to_address IN owner_addresses
) 
AND (
	includeAllNetworks 
	OR network_id IN filter_networks
)
GROUP BY
	token_id, token_address
LIMIT ? OFFSET ?