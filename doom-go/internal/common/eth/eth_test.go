package eth_test

import (
	"log"
	"testing"

	"github.com/nextsurfer/doom-go/internal/common/eth"
)

func TestFormatFloat64(t *testing.T) {
	log.Println(eth.FormatFloat64(12345.6789))
	log.Println(eth.FormatFloat64(0.000000000541457))
	log.Println(eth.FormatFloat64(-1681065128122680357))
	log.Println(eth.FormatFloat64(-0.000000000541457))
}
