package utils

import (
	"math"
)

func CorrectPrecision(Amount, Price, LotSizeMinQty, MinNotional float64) (float64, error) {

	pow := 0.0
	for {
		pow++
		LotSizeMinQty *= 10
		if LotSizeMinQty == 1.0 {
			break
		}
	}
	amount := math.Ceil(Amount*math.Pow(10, pow)) / math.Pow(10, pow)
	if amount*Price < MinNotional {
		amount = MinNotional / Price
		amount = math.Ceil(amount*math.Pow(10, pow)) / math.Pow(10, pow)
	}
	return amount, nil
}
