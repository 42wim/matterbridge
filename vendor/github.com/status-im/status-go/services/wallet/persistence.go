package wallet

import (
	"context"
	"database/sql"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

type Persistence struct {
	db *sql.DB
}

func NewPersistence(db *sql.DB) *Persistence {
	return &Persistence{db: db}
}

func (p *Persistence) SaveTokens(tokens map[common.Address][]Token) (err error) {
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

	for address, addressTokens := range tokens {
		for _, t := range addressTokens {
			for chainID, b := range t.BalancesPerChain {
				if b.HasError || b.Balance.Cmp(big.NewFloat(0)) == 0 {
					continue
				}
				_, err = tx.Exec(`INSERT INTO token_balances(user_address,token_name,token_symbol,token_address,token_decimals,token_description,token_url,balance,raw_balance,chain_id) VALUES (?,?,?,?,?,?,?,?,?,?)`, address.Hex(), t.Name, t.Symbol, b.Address.Hex(), t.Decimals, t.Description, t.AssetWebsiteURL, b.Balance.String(), b.RawBalance, chainID)
				if err != nil {
					return err
				}
			}

		}
	}

	return nil
}

func (p *Persistence) GetTokens() (map[common.Address][]Token, error) {
	rows, err := p.db.Query(`SELECT user_address, token_name, token_symbol, token_address, token_decimals, token_description, token_url, balance, raw_balance, chain_id FROM token_balances `)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	acc := make(map[common.Address]map[string]Token)

	for rows.Next() {
		var addressStr, balance, rawBalance, tokenAddress string
		token := Token{}
		var chainID uint64

		err := rows.Scan(&addressStr, &token.Name, &token.Symbol, &tokenAddress, &token.Decimals, &token.Description, &token.AssetWebsiteURL, &balance, &rawBalance, &chainID)
		if err != nil {
			return nil, err
		}

		address := common.HexToAddress(addressStr)

		if acc[address] == nil {
			acc[address] = make(map[string]Token)
		}

		if acc[address][token.Name].Name == "" {
			token.BalancesPerChain = make(map[uint64]ChainBalance)
			acc[address][token.Name] = token
		}

		tokenAcc := acc[address][token.Name]

		balanceFloat := new(big.Float)
		_, _, err = balanceFloat.Parse(balance, 10)
		if err != nil {
			return nil, err
		}

		tokenAcc.BalancesPerChain[chainID] = ChainBalance{
			RawBalance: rawBalance,
			Balance:    balanceFloat,
			Address:    common.HexToAddress(tokenAddress),
			ChainID:    chainID,
		}
	}

	result := make(map[common.Address][]Token)

	for address, tks := range acc {
		for _, t := range tks {
			result[address] = append(result[address], t)
		}
	}
	return result, nil
}
