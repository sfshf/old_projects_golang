package eth_test

import (
	"log"
	"testing"

	"github.com/nextsurfer/doom-go/internal/common/eth"
)

func TestErc55Address(t *testing.T) {
	log.Println(eth.MixedcaseAddress("0x3ddfa8ec3052539b6c9549f12cea2c295cff5296"))
	log.Println(eth.MixedcaseAddress("0x0c10bf8fcb7bf5412187a595ab97a3609160b5c6"))
	log.Println(eth.MixedcaseAddress("0xdac17f958d2ee523a2206206994597c13d831ec7"))
	log.Println(eth.MixedcaseAddress("dac17f958d2ee523a2206206994597c13d831ec7"))
	log.Println(eth.MixedcaseAddress("0x1F98431c8aD98523631AE4a59f267346ea31F984"))
	log.Println(eth.MixedcaseAddress("0x4d5f47fa6a74757f35c14fd3a6ef8e3c9bc514e8"))
}
