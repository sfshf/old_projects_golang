package mongo

import "go.mongodb.org/mongo-driver/bson/primitive"

const (
	CollectionName_UserERC20Tokens = "user_erc20_tokens"
)

type UserERC20Tokens struct {
	ID    primitive.ObjectID    `bson:"_id,omitempty"`
	Key   string                `bson:"key,omitempty"`
	Value UserERC20Tokens_Value `bson:"value,omitempty"`
}

type UserERC20Tokens_Value struct {
	ToBlock string                       `bson:"toBlock,omitempty"`
	Tokens  []UserERC20Tokens_ValueToken `bson:"tokens,omitempty"`
}

type UserERC20Tokens_ValueToken struct {
	Type     string `bson:"type,omitempty"`
	Balance  string `bson:"balance,omitempty"`
	Address  string `bson:"address,omitempty"`
	Name     string `bson:"name,omitempty"`
	Symbol   string `bson:"symbol,omitempty"`
	Decimals uint8  `bson:"decimals,omitempty"`
	Price    string `bson:"price,omitempty"`
	Value    string `bson:"value,omitempty"`
}

const (
	CollectionName_UniswapTokens = "uniswap_tokens"
	UniswapTokenTypeV2           = "Uniswap V2"
	UniswapTokenTypeV3           = "Uniswap V3"

	KeyUniswapV2Index       = "uniswapV2Index"
	KeyUniswapV3BlockNumber = "uniswapV3BlockNumber"
)

type UniswapV2Index struct {
	ID    primitive.ObjectID `bson:"_id,omitempty"`
	Key   string             `bson:"key,omitempty"`
	Value []int64            `bson:"value,omitempty"`
}

type UniswapV3BlockNumber struct {
	ID    primitive.ObjectID         `bson:"_id,omitempty"`
	Key   string                     `bson:"key,omitempty"`
	Value UniswapV3BlockNumber_Value `bson:"value,omitempty"`
}

type UniswapV3BlockNumber_Value struct {
	Timestamp   int64  `bson:"timestamp,omitempty"`
	BlockNumber string `bson:"blockNumber,omitempty"`
}

type UniswapTokens struct {
	ID    primitive.ObjectID  `bson:"_id,omitempty"`
	Key   string              `bson:"key,omitempty"`
	Index int64               `bson:"index,omitempty"`
	Value UniswapTokens_Value `bson:"value,omitempty"`
}

type UniswapTokens_Value struct {
	Type   string `bson:"type,omitempty"` // Uniswap V2; Uniswap V3
	Token0 string `bson:"token0,omitempty"`
	Token1 string `bson:"token1,omitempty"`
}

const (
	CollectionName_UserUniswapV3PoolTokens = "user_uniswapv3_pooltokens"
)

type UserUniswapV3PoolTokens struct {
	ID    primitive.ObjectID            `bson:"_id,omitempty"`
	Key   string                        `bson:"key,omitempty"`
	Value UserUniswapV3PoolTokens_Value `bson:"value,omitempty"`
}

type UserUniswapV3PoolTokens_Value struct {
	ToBlock string                               `bson:"toBlock,omitempty"`
	Tokens  []UserUniswapV3PoolTokens_ValueToken `bson:"tokens,omitempty"`
}

type UserUniswapV3PoolTokens_ValueToken struct {
	Address      string `bson:"address,omitempty"`
	PositionsKey string `bson:"positionsKey,omitempty"`
	Amount       uint64 `bson:"amount,omitempty"`
}

const (
	CollectionName_ReputableTokens = "reputable_tokens"
)

type ReputableTokens struct {
	ID    primitive.ObjectID    `bson:"_id,omitempty"`
	Key   string                `bson:"key,omitempty"`
	Value ReputableTokens_Value `bson:"value,omitempty"`
}

type ReputableTokens_Value struct {
	Type           string   `bson:"type,omitempty"`
	Name           string   `bson:"name,omitempty"`
	Symbol         string   `bson:"symbol,omitempty"`
	Decimals       uint8    `bson:"decimals,omitempty"`
	BinanceSymbols []string `bson:"binanceSymbols,omitempty"`
	OkxInstIds     []string `bson:"okxInstIds,omitempty"`
}

const (
	CollectionName_ERC20Tokens = "erc20_tokens"
	TokenTypeERC20             = "ERC20"
	TokenTypeNotERC20          = "NOT ERC20"
	TokenTypeError             = "ERROR"
)

type Erc20TokenBlockNumber struct {
	ID    primitive.ObjectID `bson:"_id,omitempty"`
	Key   string             `bson:"key,omitempty"`
	Value string             `bson:"value,omitempty"`
}

type Erc20Tokens struct {
	ID    primitive.ObjectID `bson:"_id,omitempty"`
	Key   string             `bson:"key,omitempty"`
	Value Erc20Tokens_Value  `bson:"value,omitempty"`
}

type Erc20Tokens_Value struct {
	Type     string `bson:"type,omitempty"`
	Name     string `bson:"name,omitempty"`
	Symbol   string `bson:"symbol,omitempty"`
	Decimals uint8  `bson:"decimals,omitempty"`
	Priced   bool   `bson:"priced,omitempty"`
	Checked  bool   `bson:"checked,omitempty"`
}

const (
	CollectionName_ABIs = "abis"
)

type ABIs struct {
	ID    primitive.ObjectID `bson:"_id,omitempty"`
	Key   string             `bson:"key,omitempty"`
	Value ABIs_Value         `bson:"value,omitempty"`
}

type ABIs_Value struct {
	ABI           string `bson:"ABI,omitempty"`
	IsProxy       bool   `bson:"isProxy,omitempty"`
	ProxyType     string `bson:"proxyType,omitempty"`
	TargetAddress string `bson:"targetAddress,omitempty"`
	Immutable     bool   `bson:"immutable,omitempty"`
}

const (
	CollectionName_UserTokenApprovals = "user_token_approvals"
)

type UserTokenApprovals struct {
	ID    primitive.ObjectID       `bson:"_id,omitempty"`
	Key   string                   `bson:"key,omitempty"`
	Value UserTokenApprovals_Value `bson:"value,omitempty"`
}

type UserTokenApprovals_Value struct {
	ToBlock   string                             `bson:"toBlock,omitempty"`
	Approvals []UserTokenApprovals_ValueApproval `bson:"approvals,omitempty"`
}

type UserTokenApprovals_ValueApproval struct {
	Address   string `bson:"address,omitempty"`
	Target    string `bson:"target,omitempty"`
	Allowance string `bson:"allowance,omitempty"`
}

const (
	CollectionName_ProxyTokens = "proxy_tokens"
)

type ProxyToken struct {
	ID    primitive.ObjectID `bson:"_id,omitempty"`
	Key   string             `bson:"key,omitempty"`
	Value ProxyToken_Value   `bson:"value,omitempty"`
}

type ProxyToken_Value struct {
	IsProxy     bool   `bson:"isProxy"`
	Type        string `bson:"type,omitempty"`
	Target      string `bson:"target,omitempty"`
	Immutable   bool   `bson:"immutable,omitempty"`
	BlockNumber uint64 `bson:"blockNumber,omitempty"`
}

const (
	CollectionName_AaveTokens = "aave_tokens"
	AaveTokenTypePool         = "POOL"
	AaveTokenTypeAToken       = "ATOKEN"
	AaveTokenTypeDebtToken    = "DEBT TOKEN"
)

type AaveToken struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"`
	Key         string             `bson:"key,omitempty"`
	Type        string             `bson:"type,omitempty"`
	PoolAddress string             `bson:"poolAddress,omitempty"`
}
