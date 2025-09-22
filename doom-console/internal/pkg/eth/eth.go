package eth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/big"
	"reflect"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/holiman/uint256"
	"github.com/klaytn/klaytn/common/hexutil"
)

var (
	ErrUnexpectType = errors.New("unexpected type")
)

func CallConstantFunction(ctx context.Context, client *ethclient.Client, myabi abi.ABI, address string, functionName string, params ...interface{}) ([]interface{}, error) {
	if address == "" {
		return nil, errors.New("no contract address specified")
	}
	fn := myabi.Methods[functionName]
	goParams, err := ConvertArguments(fn.Inputs, params)
	if err != nil {
		return nil, err
	}
	input, err := myabi.Pack(functionName, goParams...)
	if err != nil {
		return nil, fmt.Errorf("failed to pack values: %v", err)
	}
	toAddress := common.HexToAddress(address)
	res, err := client.CallContract(ctx, ethereum.CallMsg{Data: input, To: &toAddress}, nil)
	if err != nil {
		return nil, err
	}
	vals, err := fn.Outputs.UnpackValues(res)
	if err != nil {
		return nil, fmt.Errorf("failed to unpack values from %s: %v", hexutil.Encode(res), err)
	}
	return convertOutputParams(vals), nil
}

func ConvertArguments(args abi.Arguments, params []interface{}) ([]interface{}, error) {
	if len(args) != len(params) {
		return nil, fmt.Errorf("mismatched argument (%d) and parameter (%d) counts", len(args), len(params))
	}
	var convertedParams []interface{}
	for i, input := range args {
		param, err := ConvertArgument(input.Type, params[i])
		if err != nil {
			return nil, err
		}
		convertedParams = append(convertedParams, param)
	}
	return convertedParams, nil
}

func ConvertArgument(abiType abi.Type, param interface{}) (interface{}, error) {
	size := abiType.Size
	switch abiType.T {
	case abi.StringTy:
	case abi.BoolTy:
		if s, ok := param.(string); ok {
			val, err := strconv.ParseBool(s)
			if err != nil {
				return nil, fmt.Errorf("failed to parse bool %q: %v", s, err)
			}
			return val, nil
		}
	case abi.UintTy, abi.IntTy:
		if j, ok := param.(json.Number); ok {
			param = string(j)
		}
		if s, ok := param.(string); ok {
			val, ok := new(big.Int).SetString(s, 0)
			if !ok {
				return nil, fmt.Errorf("failed to parse big.Int: %s", s)
			}
			return ConvertInt(abiType.T == abi.IntTy, size, val)
		} else if i, ok := param.(*big.Int); ok {
			return ConvertInt(abiType.T == abi.IntTy, size, i)
		}
		v := reflect.ValueOf(param)
		switch v.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			i := new(big.Int).SetInt64(v.Int())
			return ConvertInt(abiType.T == abi.IntTy, size, i)
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			i := new(big.Int).SetUint64(v.Uint())
			return ConvertInt(abiType.T == abi.IntTy, size, i)
		case reflect.Float64, reflect.Float32:
			return nil, fmt.Errorf("floating point numbers are not valid in web3 - please use an integer or string instead (including big.Int and json.Number)")
		}
	case abi.AddressTy:
		if s, ok := param.(string); ok {
			if !common.IsHexAddress(s) {
				return nil, fmt.Errorf("invalid hex address: %s", s)
			}
			return common.HexToAddress(s), nil
		}
	case abi.SliceTy, abi.ArrayTy:
		s, ok := param.(string)
		if !ok {
			return nil, fmt.Errorf("invalid array: %s", s)
		}
		s = strings.TrimPrefix(s, "[")
		s = strings.TrimSuffix(s, "]")
		inputArray := strings.Split(s, ",")
		switch abiType.Elem.T {
		case abi.AddressTy:
			arrayParams := make([]common.Address, len(inputArray))
			for i, elem := range inputArray {
				converted, err := ConvertArgument(*abiType.Elem, elem)
				if err != nil {
					return nil, err
				}
				arrayParams[i] = converted.(common.Address)
			}
			return arrayParams, nil
		case abi.StringTy:
			arrayParams := make([]string, len(inputArray))
			for i, elem := range inputArray {
				converted, err := ConvertArgument(*abiType.Elem, elem)
				if err != nil {
					return nil, err
				}
				arrayParams[i] = converted.(string)
			}
			return arrayParams, nil
		case abi.BoolTy:
			arrayParams := make([]bool, len(inputArray))
			for i, elem := range inputArray {
				converted, err := ConvertArgument(*abiType.Elem, elem)
				if err != nil {
					return nil, err
				}
				arrayParams[i] = converted.(bool)
			}
			return arrayParams, nil
		default:
			arrayParams := make([]int, len(inputArray))
			for i, elem := range inputArray {
				converted, err := ConvertArgument(*abiType.Elem, elem)
				if err != nil {
					return nil, err
				}
				arrayParams[i] = converted.(int)
			}
			return arrayParams, nil
		}
	case abi.BytesTy:
		if s, ok := param.(string); ok {
			val, err := hexutil.Decode(s)
			if err != nil {
				return nil, fmt.Errorf("failed to parse bytes %q: %v", s, err)
			}
			return val, nil
		}
	case abi.HashTy:
		if s, ok := param.(string); ok {
			val, err := hexutil.Decode(s)
			if err != nil {
				return nil, fmt.Errorf("failed to parse hash %q: %v", s, err)
			}
			if len(val) != common.HashLength {
				return nil, fmt.Errorf("invalid hash length %d:hash must be 32 bytes", len(val))
			}
			return common.BytesToHash(val), nil
		}
	case abi.FixedBytesTy:
		switch {
		case size == 32:
			if s, ok := param.(string); ok {
				val, err := hexutil.Decode(s)
				if err != nil {
					return nil, fmt.Errorf("failed to parse hash %q: %v", s, err)
				}
				if len(val) != common.HashLength {
					return nil, fmt.Errorf("invalid hash length %d:hash must be 32 bytes", len(val))
				}
				return common.BytesToHash(val), nil
			}
		default:
			if s, ok := param.(string); ok {
				fmt.Println(s)
				val, err := hexutil.Decode(s)
				if err != nil {
					return nil, fmt.Errorf("failed to parse hash %q: %v", s, err)
				}
				if len(val) != size {
					return nil, fmt.Errorf("invalid byte array length %d: size is %d bytes", len(val), size)
				}
				arrayT := reflect.ArrayOf(size, reflect.TypeOf(byte(0)))
				array := reflect.New(arrayT).Elem()
				reflect.Copy(array, reflect.ValueOf(val))
				return array.Interface(), nil
			}
		}
	default:
		return nil, fmt.Errorf("unsupported input type %v", abiType)
	}
	return param, nil
}

func ConvertInt(signed bool, size int, i *big.Int) (interface{}, error) {
	if signed {
		switch {
		case size > 64:
			return i, nil
		case size > 32:
			if !i.IsInt64() {
				return nil, fmt.Errorf("integer overflows int64: %s", i)
			}
			return i.Int64(), nil
		case size > 16:
			if !i.IsInt64() || i.Int64() > math.MaxInt32 {
				return nil, fmt.Errorf("integer overflows int32: %s", i)
			}
			return int32(i.Int64()), nil
		case size > 8:
			if !i.IsInt64() || i.Int64() > math.MaxInt16 {
				return nil, fmt.Errorf("integer overflows int16: %s", i)
			}
			return int16(i.Int64()), nil
		default:
			if !i.IsInt64() || i.Int64() > math.MaxInt8 {
				return nil, fmt.Errorf("integer overflows int8: %s", i)
			}
			return int8(i.Int64()), nil
		}
	} else {
		switch {
		case size > 64:
			if i.Sign() == -1 {
				return nil, fmt.Errorf("negative value in unsigned field: %s", i)
			}
			return i, nil
		case size > 32:
			if !i.IsUint64() {
				return nil, fmt.Errorf("integer overflows uint64: %s", i)
			}
			return i.Uint64(), nil
		case size > 16:
			if !i.IsUint64() || i.Uint64() > math.MaxUint32 {
				return nil, fmt.Errorf("integer overflows uint32: %s", i)
			}
			return uint32(i.Uint64()), nil
		case size > 8:
			if !i.IsUint64() || i.Uint64() > math.MaxUint16 {
				return nil, fmt.Errorf("integer overflows uint16: %s", i)
			}
			return uint16(i.Uint64()), nil
		default:
			if !i.IsUint64() || i.Uint64() > math.MaxUint8 {
				return nil, fmt.Errorf("integer overflows uint8: %s", i)
			}
			return uint8(i.Uint64()), nil
		}
	}
}

func convertOutputParams(params []interface{}) []interface{} {
	for i := range params {
		p := params[i]
		if h, ok := p.(common.Hash); ok {
			params[i] = h
		} else if a, ok := p.(common.Address); ok {
			params[i] = a
		} else if b, ok := p.(hexutil.Bytes); ok {
			params[i] = b
		} else if v := reflect.ValueOf(p); v.Kind() == reflect.Array {
			if t := v.Type(); t.Elem().Kind() == reflect.Uint8 {
				b := make([]byte, t.Len())
				bv := reflect.ValueOf(b)
				// Copy since we can't t.Slice() unaddressable arrays.
				for i := 0; i < t.Len(); i++ {
					bv.Index(i).Set(v.Index(i))
				}
				params[i] = hexutil.Bytes(b)
			}
		}
	}
	return params
}

var (
	Uint256Max, _ = uint256.FromHex("0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff")
	Uint256Zero   = uint256.NewInt(0)
)

func Uint256ToFloat64(src *uint256.Int, decimals uint8) float64 {
	ten := uint256.NewInt(10)
	for ; decimals > 0; decimals-- {
		src = src.Div(src, ten)
	}
	return src.Float64()
}

// 保留两位小数/保留三位有效数字
func FormatFloat64(value float64) string {
	if value >= 0 {
		if value >= 1 {
			return strconv.FormatFloat(value, 'f', 2, 64)
		} else {
			estr := strconv.FormatFloat(value, 'e', 2, 64)
			val, err := strconv.ParseFloat(estr, 64)
			if err != nil {
				return estr
			}
			return strconv.FormatFloat(val, 'f', -1, 64)
		}
	} else {
		if value <= -1 {
			return strconv.FormatFloat(value, 'f', 2, 64)
		} else {
			estr := strconv.FormatFloat(value, 'e', 2, 64)
			val, err := strconv.ParseFloat(estr, 64)
			if err != nil {
				return estr
			}
			return strconv.FormatFloat(val, 'f', -1, 64)
		}
	}
}

func GetEthClientWorkers() ([]*ethclient.Client, func(), error) {
	serverAddresses := []string{
		"https://eth.llamarpc.com",
		"https://rpc.ankr.com/eth",
		"https://ethereum.blockpi.network/v1/rpc/public",
		"https://eth.drpc.org",
		"https://rpc.payload.de",
		"https://eth-pokt.nodies.app",
		"https://rpc.graffiti.farm",
	}
	var workers []*ethclient.Client
	for _, serverAddress := range serverAddresses {
		worker, err := ethclient.Dial(serverAddress)
		if err != nil {
			return nil, nil, err
		}
		workers = append(workers, worker)
	}
	return workers, func() {
		for _, worker := range workers {
			worker.Close()
		}
	}, nil
}
