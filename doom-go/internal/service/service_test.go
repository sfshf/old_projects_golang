package service_test

import (
	"log"
	"math"
	"testing"

	"github.com/holiman/uint256"
)

func TestUniswapV3(t *testing.T) {
	log.Printf("%.2f\n", PriceToTick(5000))  // 85176
	log.Printf("%.2f\n", TickToPrice(85176)) // 5000
	log.Printf("%.2f\n", PriceToTick(4545))  // 84222
	log.Printf("%.2f\n", TickToPrice(84222)) // 4545
	log.Printf("%.2f\n", PriceToTick(5500))  // 86129
	log.Printf("%.2f\n", TickToPrice(86129)) // 5500

	log.Printf("%.2f\n", TickToPrice(197250))
	log.Printf("%.2f\n", TickToPrice(197923))
	log.Printf("%.2f\n", SqrtpToPrice(uint256.MustFromDecimal("1572216576386672433546416844343924").Float64())) // 5500
}

func PriceToTick(p float64) float64 {
	return math.Floor(math.Log(p) / math.Log(1.0001))
}

func TickToPrice(t float64) float64 {
	return math.Ceil(math.Exp(math.Log(1.0001) * t))
}

func PriceToSqrtp(p float64) float64 {
	return math.Sqrt(p) * math.Exp2(96)
}

func SqrtpToPrice(s float64) float64 {
	return math.Pow(s/math.Exp2(96), 2)
}
