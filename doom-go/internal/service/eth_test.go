package service_test

import (
	"log"
	"testing"

	"github.com/holiman/uint256"
	"github.com/nextsurfer/doom-go/internal/common/eth"
)

func TestUint256(t *testing.T) {
	Uint256Max, _ := uint256.FromHex("0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff")
	num2, _ := uint256.FromDecimal("115792089237316195423570985008687907853269984665640564039457584007913129639935")
	log.Println(Uint256Max.String())
	log.Println(num2.String())
	log.Println(Uint256Max.Cmp(num2) == 0)
	log.Printf("%f\n", eth.Uint256ToFloat64(Uint256Max, 1))
}
