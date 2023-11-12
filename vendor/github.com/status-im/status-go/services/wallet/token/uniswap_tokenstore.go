package token

type uniswapStore struct {
}

func newUniswapStore() *uniswapStore {
	return &uniswapStore{}
}

func (s *uniswapStore) GetTokens() []*Token {
	for _, token := range uniswapTokens {
		token.TokenListID = "uniswap"
	}

	return uniswapTokens
}

func (s *uniswapStore) GetName() string {
	return "Uniswap Labs Default Token List"
}

func (s *uniswapStore) GetVersion() string {
	return "11.8.0"
}

func (s *uniswapStore) GetUpdatedAt() int64 {
	return 1697613003
}

func (s *uniswapStore) GetSource() string {
	return "https://gateway.ipfs.io/ipns/tokens.uniswap.org"
}
