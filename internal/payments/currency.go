package payments

const (
	USDT Coin = "USDT" // USDT Coin
	USDC Coin = "USDC" // USDC Coin
)

// Coin represents a coin
type Coin string

// IsStableCoin returns true if the coin is a stable coin
func (c Coin) IsStableCoin() bool {
	return c == USDT || c == USDC
}
