package payments

const (
	USDT Coin = "USDT"
	USDC Coin = "USDC"
)

// Coin represents a coin
type Coin string

func (c Coin) IsStableCoin() bool {
	return c == USDT || c == USDC
}
