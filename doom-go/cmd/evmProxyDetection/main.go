package main

import (
	"context"
	"encoding/hex"
	"errors"
	"log"
	"strconv"
	"strings"
	"sync"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

// proxy type
const (
	ProxyTypeEip1167        = "Eip1167"
	ProxyTypeEip1967Direct  = "Eip1967Direct"
	ProxyTypeEip1967Beacon  = "Eip1967Beacon"
	ProxyTypeEip1822        = "Eip1822"
	ProxyTypeEip897         = "Eip897"
	ProxyTypeOpenZeppelin   = "OpenZeppelin"
	ProxyTypeSafe           = "Safe"
	ProxyTypeComptroller    = "Comptroller"
	ProxyTypeEip3561Upgrade = "Eip3561Upgrade"
)

type EvmProxyDetectionResult struct {
	Target    string `json:"target"`
	Type      string `json:"type"`
	Immutable bool   `json:"boolean"`
}

// eip1167 -------------------------------------------------------------------------------------------------------------

const (
	EIP_1167_BYTECODE_PREFIX       = "363d3d373d3d3d363d"
	EIP_1167_BYTECODE_SUFFIX       = "57fd5bf3"
	SUFFIX_OFFSET_FROM_ADDRESS_END = 22
)

var (
	ErrNotEIP1167ByteCode = errors.New("not an EIP-1167 bytecode")
	ErrNotProxyToken      = errors.New("not proxy token")
)

func parse1167Bytecode(bytecodeHex string) (string, error) {
	if !strings.HasPrefix(bytecodeHex, EIP_1167_BYTECODE_PREFIX) {
		return "", ErrNotEIP1167ByteCode
	}
	// detect length of address (20 bytes non-optimized, 0 < N < 20 bytes for vanity addresses)
	prefixLength := int64(len(EIP_1167_BYTECODE_PREFIX))
	pushNHex := bytecodeHex[prefixLength : prefixLength+2]
	// push1 ... push20 use opcodes 0x60 ... 0x73
	pushNHexNum, err := strconv.ParseInt(string(pushNHex), 16, 64)
	if err != nil {
		return "", err
	}
	addressLength := pushNHexNum - 0x5f
	if addressLength < 1 || addressLength > 20 {
		return "", ErrNotEIP1167ByteCode
	}
	addressFromBytecode := bytecodeHex[prefixLength+2 : prefixLength+2+addressLength*2] // address length is in bytes, 2 hex chars make up 1 byte
	if !strings.HasPrefix(bytecodeHex[prefixLength+2+addressLength*2+SUFFIX_OFFSET_FROM_ADDRESS_END:], EIP_1167_BYTECODE_SUFFIX) {
		return "", ErrNotEIP1167ByteCode
	}
	var buf strings.Builder
	buf.WriteString("0x")
	// padStart is needed for vanity addresses
	buf.WriteString(strings.Repeat("0", 40))
	buf.WriteString(addressFromBytecode)
	return buf.String(), nil
}

const (
	EIP_1967_LOGIC_SLOT               = "0x360894a13ba1a3210667c828492db98dca3e2076cc3735a920a3ca505d382bbc" // obtained as bytes32(uint256(keccak256('eip1967.proxy.implementation')) - 1)
	EIP_1967_BEACON_SLOT              = "0xa3f0ad74e5423aebfd80d3ef4346578335a9a72aeaee59ff6cb3582b35133d50" // obtained as bytes32(uint256(keccak256('eip1967.proxy.beacon')) - 1)
	OPEN_ZEPPELIN_IMPLEMENTATION_SLOT = "0x7050c9e0f4ca769c69bd3a8ef740bc37934f8e2c036e5a723fd8ee048ed3f8c3" // obtained as keccak256("org.zeppelinos.proxy.implementation")
	EIP_1822_LOGIC_SLOT               = "0xc5f16f0fcc639fa48a6947836d9850f504798523bf8c9a3a87d5876cf622bcf7" // obtained as keccak256("PROXIABLE")
)

var (
	EIP_897_INTERFACE = []string{
		// bytes4(keccak256("implementation()")) padded to 32 bytes
		"5c60da1b00000000000000000000000000000000000000000000000000000000",
		// bytes4(keccak256("proxyType()")) padded to 32 bytes
		"4555d5c900000000000000000000000000000000000000000000000000000000",
	}
	EIP_1967_BEACON_METHODS = []string{
		// bytes4(keccak256("implementation()")) padded to 32 bytes
		"5c60da1b00000000000000000000000000000000000000000000000000000000",
		// bytes4(keccak256("childImplementation()")) padded to 32 bytes
		// some implementations use this over the standard method name so that the beacon contract is not detected as an EIP-897 proxy itself
		"da52571600000000000000000000000000000000000000000000000000000000",
	}
	SAFE_PROXY_INTERFACE = []string{
		// bytes4(keccak256("masterCopy()")) padded to 32 bytes
		"a619486e00000000000000000000000000000000000000000000000000000000",
	}
	COMPTROLLER_PROXY_INTERFACE = []string{
		// bytes4(keccak256("comptrollerImplementation()")) padded to 32 bytes
		"bb82aa5e00000000000000000000000000000000000000000000000000000000",
	}
)

var (
	zeroAddress       = "0x" + strings.Repeat("0", 40)
	ErrEmptyAddress   = errors.New("empty address")
	ErrInvalidAddress = errors.New("invalid address")
)

func readAddress(src string) (string, error) {
	if src == "0x" {
		return "", ErrInvalidAddress
	}
	if !strings.HasPrefix(src, "0x") {
		src = "0x" + src
	}
	if srcLen := len(src); srcLen == 66 {
		src = "0x" + src[srcLen-40:]
	}
	if src == zeroAddress {
		return "", ErrEmptyAddress
	}
	return src, nil
}

// EIP-1167 Minimal Proxy Contract
func EIP1167MinimalProxyContract(ctx context.Context, client *ethclient.Client, contractAddress common.Address) (*EvmProxyDetectionResult, error) {
	bytecode, err := client.CodeAt(ctx, contractAddress, nil)
	if err != nil {
		return nil, err
	}
	hexcode := hex.EncodeToString(bytecode)
	address, err := parse1167Bytecode(hexcode)
	if err != nil {
		return nil, err
	}
	target, err := readAddress(address)
	if err != nil {
		return nil, err
	}
	return &EvmProxyDetectionResult{Target: target, Type: ProxyTypeEip1167, Immutable: true}, nil
}

// EIP-1967 direct proxy
func EIP1967DirectProxy(ctx context.Context, client *ethclient.Client, contractAddress common.Address) (*EvmProxyDetectionResult, error) {
	storage, err := client.StorageAt(ctx, contractAddress, common.HexToHash(EIP_1967_LOGIC_SLOT), nil)
	if err != nil {
		return nil, err
	}
	address := hex.EncodeToString(storage)
	target, err := readAddress(address)
	if err != nil {
		return nil, err
	}
	return &EvmProxyDetectionResult{Target: target, Type: ProxyTypeEip1967Direct, Immutable: false}, nil
}

// EIP-1967 beacon proxy
func EIP1967BeaconProxy(ctx context.Context, client *ethclient.Client, contractAddress common.Address) (*EvmProxyDetectionResult, error) {
	storage, err := client.StorageAt(ctx, contractAddress, common.HexToHash(EIP_1967_BEACON_SLOT), nil)
	if err != nil {
		return nil, err
	}
	address := hex.EncodeToString(storage)
	beaconAddress, err := readAddress(address)
	if err != nil {
		return nil, err
	}
	toAddress := common.HexToAddress(beaconAddress)
	res, err := client.CallContract(ctx, ethereum.CallMsg{Data: common.Hex2Bytes(EIP_1967_BEACON_METHODS[0]), To: &toAddress}, nil)
	if err != nil {
		res, err = client.CallContract(ctx, ethereum.CallMsg{Data: common.Hex2Bytes(EIP_1967_BEACON_METHODS[1]), To: &toAddress}, nil)
		if err != nil {
			return nil, err
		}
	}
	target, err := readAddress(hex.EncodeToString(res))
	if err != nil {
		return nil, err
	}
	return &EvmProxyDetectionResult{Target: target, Type: ProxyTypeEip1967Beacon, Immutable: false}, nil
}

// OpenZeppelin proxy pattern
func OpenZeppelinProxyPattern(ctx context.Context, client *ethclient.Client, contractAddress common.Address) (*EvmProxyDetectionResult, error) {
	storage, err := client.StorageAt(ctx, contractAddress, common.HexToHash(OPEN_ZEPPELIN_IMPLEMENTATION_SLOT), nil)
	if err != nil {
		return nil, err
	}
	address := hex.EncodeToString(storage)
	target, err := readAddress(address)
	if err != nil {
		return nil, err
	}
	return &EvmProxyDetectionResult{Target: target, Type: ProxyTypeOpenZeppelin, Immutable: false}, nil
}

// EIP-1822 Universal Upgradeable Proxy Standard
func EIP1822UniversalUpgradeableProxyStandard(ctx context.Context, client *ethclient.Client, contractAddress common.Address) (*EvmProxyDetectionResult, error) {
	storage, err := client.StorageAt(ctx, contractAddress, common.HexToHash(EIP_1822_LOGIC_SLOT), nil)
	if err != nil {
		return nil, err
	}
	address := hex.EncodeToString(storage)
	target, err := readAddress(address)
	if err != nil {
		return nil, err
	}
	return &EvmProxyDetectionResult{Target: target, Type: ProxyTypeEip1822, Immutable: false}, nil
}

// EIP-897 DelegateProxy pattern
func EIP897DelegateProxyPattern(ctx context.Context, client *ethclient.Client, contractAddress common.Address) (*EvmProxyDetectionResult, error) {
	res, err := client.CallContract(ctx, ethereum.CallMsg{Data: common.Hex2Bytes(EIP_897_INTERFACE[0]), To: &contractAddress}, nil)
	if err != nil {
		return nil, err
	}
	target, err := readAddress(hex.EncodeToString(res))
	if err != nil {
		return nil, err
	}
	// proxyType === 1 means that the proxy is immutable
	var immutable bool
	res, _ = client.CallContract(ctx, ethereum.CallMsg{Data: common.Hex2Bytes(EIP_897_INTERFACE[1]), To: &contractAddress}, nil)
	if hex.EncodeToString(res) == "0000000000000000000000000000000000000000000000000000000000000001" {
		immutable = true
	}
	return &EvmProxyDetectionResult{Target: target, Type: ProxyTypeEip897, Immutable: immutable}, nil
}

// SafeProxy contract
func SafeProxyContract(ctx context.Context, client *ethclient.Client, contractAddress common.Address) (*EvmProxyDetectionResult, error) {
	res, err := client.CallContract(ctx, ethereum.CallMsg{Data: common.Hex2Bytes(SAFE_PROXY_INTERFACE[0]), To: &contractAddress}, nil)
	if err != nil {
		return nil, err
	}
	target, err := readAddress(hex.EncodeToString(res))
	if err != nil {
		return nil, err
	}
	return &EvmProxyDetectionResult{Target: target, Type: ProxyTypeSafe, Immutable: false}, nil
}

// Comptroller proxy
func ComptrollerProxy(ctx context.Context, client *ethclient.Client, contractAddress common.Address) (*EvmProxyDetectionResult, error) {
	res, err := client.CallContract(ctx, ethereum.CallMsg{Data: common.Hex2Bytes(COMPTROLLER_PROXY_INTERFACE[0]), To: &contractAddress}, nil)
	if err != nil {
		return nil, err
	}
	target, err := readAddress(hex.EncodeToString(res))
	if err != nil {
		return nil, err
	}
	return &EvmProxyDetectionResult{Target: target, Type: ProxyTypeComptroller, Immutable: false}, nil
}

func EvmProxyDetection(ctx context.Context, client *ethclient.Client, contractAddress common.Address) (result *EvmProxyDetectionResult, err error) {
	// 1. EIP-1167 Minimal Proxy Contract
	result, err = EIP1167MinimalProxyContract(ctx, client, contractAddress)
	if err == nil {
		return result, nil
	}
	// 2. EIP-1967 direct proxy
	result, err = EIP1967DirectProxy(ctx, client, contractAddress)
	if err == nil {
		return result, nil
	}
	// 3. EIP-1967 beacon proxy
	result, err = EIP1967BeaconProxy(ctx, client, contractAddress)
	if err == nil {
		return result, nil
	}
	// 4. OpenZeppelin proxy pattern
	result, err = OpenZeppelinProxyPattern(ctx, client, contractAddress)
	if err == nil {
		return result, nil
	}
	// 5. EIP-1822 Universal Upgradeable Proxy Standard
	result, err = EIP1822UniversalUpgradeableProxyStandard(ctx, client, contractAddress)
	if err == nil {
		return result, nil
	}
	// 6. EIP-897 DelegateProxy pattern
	result, err = EIP897DelegateProxyPattern(ctx, client, contractAddress)
	if err == nil {
		return result, nil
	}
	// 7. SafeProxy contract
	result, err = SafeProxyContract(ctx, client, contractAddress)
	if err == nil {
		return result, nil
	}
	// 8. Comptroller proxy
	result, err = ComptrollerProxy(ctx, client, contractAddress)
	if err == nil {
		return result, nil
	}
	return nil, err
}

func AnomalyNetwork(ctx context.Context, client *ethclient.Client, contractAddress common.Address) (result *EvmProxyDetectionResult, err error) {
	// go func() {
	// 	client.Close()
	// 	log.Println("ethclient is closed")
	// }()
	return EvmProxyDetection(ctx, client, contractAddress)
}

func ManyRequests(ctx context.Context, client *ethclient.Client, contractAddress common.Address) (result *EvmProxyDetectionResult, err error) {
	var wg sync.WaitGroup
	for i := 0; i < 500; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			result, err := EvmProxyDetection(ctx, client, contractAddress)
			if err != nil {
				log.Printf("request %d error: %v", i, err)
				return
			}
			log.Printf("request %d: %v", i, result)
		}(i)
	}
	wg.Wait()
	return nil, nil
}

func main() {
	log.SetFlags(log.Llongfile | log.LstdFlags)
	ctx := context.Background()
	client, err := ethclient.DialContext(ctx, "wss://mainnet.infura.io/ws/v3/ae1cb3ad3a4542e294f99a5f92be46c9")
	if err != nil {
		log.Println(err)
		return
	}
	defer client.Close()
	contractAddress := common.HexToAddress("0x42B24A95702b9986e82d421cC3568932790A48Ec")
	// result, err := AnomalyNetwork(ctx, client, contractAddress) // connect: network is unreachable
	// if err != nil {
	// 	log.Println(err)
	// 	return
	// }
	// result, err := ManyRequests(ctx, client, contractAddress) // project ID request rate exceeded
	// if err != nil {
	// 	log.Println(err)
	// 	return
	// }
	// log.Println(result)
	// log.Println("ok")
	// 1. EIP-1167 Minimal Proxy Contract
	_, err = EIP1167MinimalProxyContract(ctx, client, contractAddress)
	if err != nil {
		log.Println("EIP1167MinimalProxyContract error:", err)
	}
	// 2. EIP-1967 direct proxy
	_, err = EIP1967DirectProxy(ctx, client, contractAddress)
	if err != nil {
		log.Println("EIP1967DirectProxy error:", err)
	}
	// 3. EIP-1967 beacon proxy
	_, err = EIP1967BeaconProxy(ctx, client, contractAddress)
	if err != nil {
		log.Println("EIP1967BeaconProxy error:", err)
	}
	// 4. OpenZeppelin proxy pattern
	_, err = OpenZeppelinProxyPattern(ctx, client, contractAddress)
	if err != nil {
		log.Println("OpenZeppelinProxyPattern error:", err)
	}
	// 5. EIP-1822 Universal Upgradeable Proxy Standard
	_, err = EIP1822UniversalUpgradeableProxyStandard(ctx, client, contractAddress)
	if err != nil {
		log.Println("EIP1822UniversalUpgradeableProxyStandard error:", err)
	}
	// 6. EIP-897 DelegateProxy pattern
	_, err = EIP897DelegateProxyPattern(ctx, client, contractAddress)
	if err != nil {
		log.Println("EIP897DelegateProxyPattern error:", err)
	}
	// 7. SafeProxy contract
	_, err = SafeProxyContract(ctx, client, contractAddress)
	if err != nil {
		log.Println("SafeProxyContract error:", err)
	}
	// 8. Comptroller proxy
	_, err = ComptrollerProxy(ctx, client, contractAddress)
	if err != nil {
		log.Println("ComptrollerProxy error:", err)
	}
}
