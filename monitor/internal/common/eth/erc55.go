package eth

import "github.com/ethereum/go-ethereum/common"

func MixedcaseAddress(hexAddr string) string {
	return common.AddressEIP55(common.HexToAddress(hexAddr)).String()
}
