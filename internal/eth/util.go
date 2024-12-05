package eth

import (
	"math/big"

	"github.com/ethereum/go-ethereum/params"
)

// ToWei converts an ETH amount (float64) tg wei (*big.Int).
func ToWei(amount float64) *big.Int {
	ethValue := new(big.Float).SetFloat64(amount)
	weiMultiplier := new(big.Float).SetInt64(params.Ether)
	weiValue := new(big.Float).Mul(ethValue, weiMultiplier)

	weiInt := new(big.Int)
	weiValue.Int(weiInt)

	return weiInt
}
