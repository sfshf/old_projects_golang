package service

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"math/big"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/go-redis/redis/v8"
	"github.com/holiman/uint256"
	doom_api "github.com/nextsurfer/doom-go/api"
	"github.com/nextsurfer/doom-go/api/response"
	"github.com/nextsurfer/doom-go/internal/common/config"
	"github.com/nextsurfer/doom-go/internal/common/eth"
	ethabi "github.com/nextsurfer/doom-go/internal/common/eth/abi"
	"github.com/nextsurfer/doom-go/internal/common/simplehttp"
	. "github.com/nextsurfer/doom-go/internal/model"
	gerror "github.com/nextsurfer/ground/pkg/err"
	"github.com/nextsurfer/ground/pkg/rpc"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"
)

type EthService struct {
	*DoomService
}

func NewEthService(DoomService *DoomService) *EthService {
	return &EthService{
		DoomService: DoomService,
	}
}

type UserToken struct {
	Type         string `json:"type"`
	Balance      string `json:"balance"`
	BalanceValue string `json:"balanceValue"`
	Address      string `json:"address"`
	Name         string `json:"name"`
	Symbol       string `json:"symbol"`
	Decimals     uint8  `json:"decimals"`
	Price        string `json:"price"`
	Value        string `json:"value"`
}

type UpsertUserErc20TokensResult struct {
	ToBlock int64
	Tokens  []*UserToken
}

func (s *EthService) handleUserErc20Tokens(ctx context.Context, rpcCtx *rpc.Context, client *ethclient.Client, erc20ABI abi.ABI, userERC20Tokens *UserERC20Tokens, result *FilterTransferLogsResult, recordBalances map[string]*uint256.Int, needPrice bool) ([]*UserToken, error) {
	var once sync.Once
	var res []*UserToken
	for key, val := range result.ToBalances {
		fromMapBalance, has := result.FromBalances[key]
		// handle balance
		if has {
			val = val.Sub(val, fromMapBalance)
		}
		recordBalance, has := recordBalances[key]
		if has {
			val = val.Add(val, recordBalance)
		}
		token, err := s.GenerateEthToken(ctx, client, erc20ABI, key, val, &once)
		if err != nil {
			return nil, err
		}
		if token != nil {
			res = append(res, token)
		}
		delete(recordBalances, key) // 去除已处理项
	}
	// 清空剩余项
	for key := range recordBalances {
		for _, item := range userERC20Tokens.Value.Tokens {
			if item.Address == key {
				res = append(res, &UserToken{
					Address:  item.Address,  // address
					Balance:  item.Balance,  // original balance
					Type:     item.Type,     // type
					Name:     item.Name,     // name
					Symbol:   item.Symbol,   // symbol
					Decimals: item.Decimals, // decimals
				})
			}
		}
	}
	if needPrice {
		// handle prices
		if err := s.handleUserErc20TokenPrices(ctx, rpcCtx, res); err != nil {
			return nil, err
		}
	}
	return res, nil
}

func (s *EthService) HandleTransferLogsResult(ctx context.Context, rpcCtx *rpc.Context, mongoColl *mongo.Collection, userAddress string, userTokens []*UserToken, result *FilterTransferLogsResult) (*UpsertUserErc20TokensResult, *gerror.AppError) {
	logger := s.Logger
	if rpcCtx != nil {
		logger = rpcCtx.Logger
	}
	record := UserERC20Tokens{
		Key: userAddress,
	}
	res := &UpsertUserErc20TokensResult{}
	if result.ToBlock > 0 {
		record.Value.ToBlock = strconv.FormatInt(result.ToBlock, 10)
		res.ToBlock = result.ToBlock
	}
	tokensLen := len(userTokens)
	recordTokens := make([]UserERC20Tokens_ValueToken, 0, tokensLen) // 回写到user_erc20_tokens表的token列表
	resTokens := make([]*UserToken, 0, tokensLen)                    // 本地请求要返回的token列表
	for _, token := range userTokens {
		recordTokens = append(recordTokens, UserERC20Tokens_ValueToken{
			Type:     token.Type,
			Address:  token.Address,
			Balance:  token.Balance,
			Name:     token.Name,
			Symbol:   token.Symbol,
			Decimals: token.Decimals,
			Price:    token.Price,
			Value:    token.Value,
		})
		if token.Type != TokenTypeNotERC20 {
			resTokens = append(resTokens, token)
		}
	}
	record.Value.Tokens = recordTokens
	res.Tokens = resTokens
	// last, upsert the record, and return res
	ts := time.Now()
	_, err := mongoColl.ReplaceOne(ctx, bson.D{{Key: "key", Value: userAddress}}, record, options.Replace().SetUpsert(true))
	StatisticMongoCall(ctx, ts)
	if err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	return res, nil
}

func (s *EthService) handleUserErc20TokenPrices(ctx context.Context, rpcCtx *rpc.Context, userTokens []*UserToken) error {
	for _, userToken := range userTokens {
		if userToken.Type == TokenTypeNotERC20 {
			continue
		}
		if userToken.Symbol != "" {
			price, _ := s.GetLatestSpotPrice(ctx, rpcCtx, &doom_api.GetLatestSpotPriceRequest{
				Symbol:   userToken.Symbol,
				BaseCoin: "USDT",
			})
			if price != nil {
				userToken.Price = price.Price
				priceValue, err := strconv.ParseFloat(userToken.Price, 64)
				if err != nil {
					return err
				}
				balanceValue := eth.Uint256ToFloat64(uint256.MustFromDecimal(userToken.Balance), userToken.Decimals)
				userToken.BalanceValue = eth.FormatFloat64(balanceValue)       // computed balance
				userToken.Value = eth.FormatFloat64(priceValue * balanceValue) // value
			}
		}
	}
	return nil
}

func (s *EthService) HandleTransferLogs(ctx context.Context, rpcCtx *rpc.Context, client *ethclient.Client, erc20ABI abi.ABI, mongoColl *mongo.Collection, userAddress string, userERC20Tokens *UserERC20Tokens, needPrice bool) (*UpsertUserErc20TokensResult, *gerror.AppError) {
	logger := s.Logger
	if rpcCtx != nil {
		logger = rpcCtx.Logger
	}
	// fromBlock, toBlock
	var fromBlock *big.Int
	if userERC20Tokens.Key != "" && userERC20Tokens.Value.ToBlock != "" {
		toBlockValue, err := strconv.ParseInt(userERC20Tokens.Value.ToBlock, 10, 64)
		if err != nil {
			logger.Error("internal error", zap.NamedError("appError", err))
			return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
		}
		fromBlock = big.NewInt(toBlockValue + 1)
	}
	ts := time.Now()
	header, err := client.HeaderByNumber(ctx, nil)
	StatisticWeb3Call(ctx, ts)
	if err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	result := s.FilterTransferLogs(ctx, rpcCtx, client, crypto.Keccak256Hash([]byte("Transfer(address,address,uint256)")), common.HexToHash(userAddress), fromBlock, header.Number, regexp.MustCompile(`.*0x([[:xdigit:]]+).*0x([[:xdigit:]]+).*`))
	if result.Err != nil {
		logger.Error("internal error", zap.NamedError("appError", result.Err))
		return nil, gerror.NewError(result.Err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	// third, handle result
	// 上一次搜索到的token的余额集，用于计算token在本次处理中的最新余额值
	recordBalances := make(map[string]*uint256.Int, len(userERC20Tokens.Value.Tokens))
	for _, item := range userERC20Tokens.Value.Tokens {
		balanceValue, err := uint256.FromDecimal(item.Balance)
		if err != nil {
			logger.Error("internal error", zap.NamedError("appError", err))
			return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
		}
		if item.Type != TokenTypeNotERC20 {
			recordBalances[item.Address] = balanceValue
		}
	}
	userTokens, err := s.handleUserErc20Tokens(ctx, rpcCtx, client, erc20ABI, userERC20Tokens, &result, recordBalances, needPrice)
	if err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	return s.HandleTransferLogsResult(ctx, rpcCtx, mongoColl, userAddress, userTokens, &result)
}

func (s *EthService) UpsertUserErc20Tokens(ctx context.Context, rpcCtx *rpc.Context, client *ethclient.Client, userAddress string, needPrice bool) (*UpsertUserErc20TokensResult, *gerror.AppError) {
	logger := s.Logger
	if rpcCtx != nil {
		logger = rpcCtx.Logger
	}
	userAddress = eth.MixedcaseAddress(userAddress)
	// erc20ABI
	erc20ABI, err := ethabi.GetABI(ethabi.ERC20ABI)
	if err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	// first, get record from mongo db
	mongoColl := s.MongoDB.Collection(CollectionName_UserERC20Tokens)
	var userERC20Tokens UserERC20Tokens
	ts := time.Now()
	err = mongoColl.FindOne(ctx, bson.D{{Key: "key", Value: userAddress}}).Decode(&userERC20Tokens)
	StatisticMongoCall(ctx, ts)
	if err != nil {
		if err != mongo.ErrNoDocuments {
			logger.Error("internal error", zap.NamedError("appError", err))
			return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
		}
	}
	return s.HandleTransferLogs(ctx, rpcCtx, client, *erc20ABI, mongoColl, userAddress, &userERC20Tokens, needPrice)
}

type HandleEthChainApprovalLogResult struct {
	ToBlock   int64
	Approvals []*ApprovalInfo
}

func (s *EthService) handleApprovalLogs(ctx context.Context, rpcCtx *rpc.Context, userAddress string, mongoColl *mongo.Collection, lastApprovals map[string]*ApprovalInfo, result *HandleApprovalLogsResult) (*HandleEthChainApprovalLogResult, *gerror.AppError) {
	logger := s.Logger
	if rpcCtx != nil {
		logger = rpcCtx.Logger
	}
	record := UserTokenApprovals{
		Key: userAddress,
	}
	if result.ToBlock > 0 {
		record.Value.ToBlock = strconv.FormatInt(result.ToBlock, 10)
	}
	// refresh approvals
	var approvals []UserTokenApprovals_ValueApproval
	var resultApprovals []*ApprovalInfo
	zero := uint256.NewInt(0)
	for key, val := range result.Approvals {
		_, has := lastApprovals[key]
		if has {
			delete(lastApprovals, key)
		}
		if val.Allowance.Cmp(zero) != 0 {
			approvals = append(approvals, UserTokenApprovals_ValueApproval{
				Address:   val.Address,
				Target:    val.Target,
				Allowance: val.Allowance.String(),
			})
			resultApprovals = append(resultApprovals, val)
		}
	}
	for key, val := range lastApprovals {
		result.Approvals[key] = val
		approvals = append(approvals, UserTokenApprovals_ValueApproval{
			Address:   val.Address,
			Target:    val.Target,
			Allowance: val.Allowance.String(),
		})
		resultApprovals = append(resultApprovals, val)
	}
	record.Value.Approvals = approvals
	// third, upsert the record
	if _, err := mongoColl.ReplaceOne(ctx, bson.D{{Key: "key", Value: userAddress}}, record, options.Replace().SetUpsert(true)); err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	return &HandleEthChainApprovalLogResult{
		ToBlock:   result.ToBlock,
		Approvals: resultApprovals,
	}, nil
}

func (s *EthService) HandleEthChain_ApprovalLog(ctx context.Context, rpcCtx *rpc.Context, client *ethclient.Client, userAddress string) (*HandleEthChainApprovalLogResult, *gerror.AppError) {
	logger := s.Logger
	if rpcCtx != nil {
		logger = rpcCtx.Logger
	}
	// fromBlock, toBlock error regexp
	// first, get record from mongo db
	mongoColl := s.MongoDB.Collection(CollectionName_UserTokenApprovals)
	var userTokenApprovals UserTokenApprovals
	if err := mongoColl.FindOne(ctx, bson.D{{Key: "key", Value: userAddress}}).Decode(&userTokenApprovals); err != nil {
		if err != mongo.ErrNoDocuments {
			logger.Error("internal error", zap.NamedError("appError", err))
			return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
		}
	}
	lastApprovals := make(map[string]*ApprovalInfo)
	if userTokenApprovals.Key != "" {
		for _, approval := range userTokenApprovals.Value.Approvals {
			allowance, err := uint256.FromDecimal(approval.Allowance)
			if err != nil {
				logger.Error("internal error", zap.NamedError("appError", err))
				return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
			}
			lastApprovals[approval.Address+approval.Target] = &ApprovalInfo{
				Address:   approval.Address,
				Target:    approval.Target,
				Allowance: allowance,
			}
		}
	}
	var fromBlock int64
	if userTokenApprovals.Key != "" && userTokenApprovals.Value.ToBlock != "" {
		toBlockValue, err := strconv.ParseInt(userTokenApprovals.Value.ToBlock, 10, 64)
		if err != nil {
			logger.Error("internal error", zap.NamedError("appError", err))
			return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
		}
		fromBlock = toBlockValue + 1
	}
	// second, handle Approval event logs
	result := s.HandleApprovalLogs(ctx, rpcCtx, client, crypto.Keccak256Hash([]byte("Approval(address,address,uint256)")), common.HexToHash(userAddress), big.NewInt(fromBlock), regexp.MustCompile(`.*0x([[:xdigit:]]+).*0x([[:xdigit:]]+).*`))
	if result.Err != nil {
		logger.Error("internal error", zap.NamedError("appError", result.Err))
		return nil, gerror.NewError(result.Err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	return s.handleApprovalLogs(ctx, rpcCtx, userAddress, mongoColl, lastApprovals, &result)
}

const (
	LogLimit uint64 = 50000
)

var (
	ErrLogLimit            = errors.New("log limit")
	ErrInvalidBlockSection = errors.New("invalid block section")
)

type FilterTransferLogsResult struct {
	FromBalances map[string]*uint256.Int
	ToBalances   map[string]*uint256.Int
	ToBlock      int64
	Err          error
}

func transferLogHandle(tokenMap map[string]*uint256.Int) func(vLog types.Log) {
	return func(vLog types.Log) {
		key := eth.MixedcaseAddress(vLog.Address.Hex())
		totalValue, has := tokenMap[key]
		if !has {
			totalValue = uint256.NewInt(0)
		}
		val := new(uint256.Int)
		val.SetFromHex("0x" + hex.EncodeToString(vLog.Data))
		totalValue = totalValue.Add(totalValue, val)
		tokenMap[key] = totalValue
	}
}

func (s *EthService) FilterTransferLogs(ctx context.Context, rpcCtx *rpc.Context, client *ethclient.Client, sigHash, userAddress common.Hash, fromBlock, toBlock *big.Int, errRegexp *regexp.Regexp) (result FilterTransferLogsResult) {
	fromBalances := make(map[string]*uint256.Int) // transfer from user
	toBalances := make(map[string]*uint256.Int)   // transfer to user
	var count uint64
	lastFromBlock := fromBlock
	// from
	fromTopics := [][]common.Hash{{sigHash}, {userAddress}, {}} // fromTopics
	for {
		ts := time.Now()
		logs, query, err := s.FilterLogs(ctx, client, fromBlock, toBlock, fromTopics, errRegexp, "")
		StatisticWeb3Call(ctx, ts)
		if err != nil {
			result.Err = err
			return
		}
		count += s.HandleLogs(logs, sigHash, transferLogHandle(fromBalances))
		if count >= LogLimit {
			fromBlock = lastFromBlock
			toBlock = query.ToBlock
			break
		}
		if query.ToBlock.Cmp(toBlock) == 0 {
			fromBlock = lastFromBlock
			break
		} else {
			fromBlock = big.NewInt(query.ToBlock.Int64() + 1) // next turn
		}
	}
	// to
	toTopics := [][]common.Hash{{sigHash}, {}, {userAddress}} // toTopics
	for {
		ts := time.Now()
		logs, query, err := s.FilterLogs(ctx, client, fromBlock, toBlock, toTopics, errRegexp, "")
		if err != nil {
			result.Err = err
			return
		}
		StatisticWeb3Call(ctx, ts)
		s.HandleLogs(logs, sigHash, transferLogHandle(toBalances))
		if query.ToBlock.Cmp(toBlock) == 0 {
			break
		} else {
			fromBlock = big.NewInt(query.ToBlock.Int64() + 1) // next turn
		}
	}
	result.FromBalances = fromBalances
	result.ToBalances = toBalances
	result.ToBlock = toBlock.Int64()
	return
}

type HandleApprovalLogsResult struct {
	Approvals map[string]*ApprovalInfo
	ToBlock   int64
	Err       error
}

type ApprovalInfo struct {
	Address   string
	Target    string
	Allowance *uint256.Int
}

func (s *EthService) HandleApprovalLogs(ctx context.Context, rpcCtx *rpc.Context, client *ethclient.Client, sigHash, userAddress common.Hash, lastFromBlock *big.Int, errRegexp *regexp.Regexp) (result HandleApprovalLogsResult) {
	var err error
	// fromBlock, toBlock
	header, err := client.HeaderByNumber(ctx, nil)
	if err != nil {
		return
	}
	fromBlock := lastFromBlock
	toBlock := header.Number
	// topics
	topics := [][]common.Hash{{sigHash}, {userAddress}, {}}
	var count uint64
	var logs []types.Log
	var query ethereum.FilterQuery
	approvals := make(map[string]*ApprovalInfo)
	for {
		logs, query, err = s.FilterLogs(ctx, client, fromBlock, toBlock, topics, errRegexp, "")
		if err != nil {
			result.Err = err
			return
		}
		s.HandleLogs(logs, sigHash, func(vLog types.Log) {
			address := eth.MixedcaseAddress(vLog.Address.Hex())
			target := eth.MixedcaseAddress("0x" + vLog.Topics[2].Hex()[26:])
			val := new(uint256.Int)
			val.SetFromHex("0x" + hex.EncodeToString(vLog.Data))
			approvals[address+target] = &ApprovalInfo{
				Address:   address,
				Target:    target,
				Allowance: val,
			}
		})
		count++
		if count > 3 {
			toBlock = query.ToBlock
			break
		}
		// next turn
		if query.ToBlock == nil || query.ToBlock.Cmp(toBlock) == 0 {
			break
		} else {
			fromBlock = big.NewInt(query.ToBlock.Int64() + 1)
		}
	}
	result.Approvals = approvals
	result.ToBlock = header.Number.Int64()
	if toBlock != nil && toBlock.Cmp(header.Number) < 0 {
		result.ToBlock = toBlock.Int64()
	}
	return
}

func (s *EthService) FilterLogs(ctx context.Context, client *ethclient.Client, fromBlock, toBlock *big.Int, topics [][]common.Hash, errRegexp *regexp.Regexp, contractAddress string) ([]types.Log, ethereum.FilterQuery, error) {
	var addresses []common.Address
	if contractAddress != "" {
		addresses = append(addresses, common.HexToAddress(contractAddress))
	}
	query := ethereum.FilterQuery{
		FromBlock: fromBlock,
		ToBlock:   toBlock,
		Topics:    topics,
		Addresses: addresses,
	}
	logs, err := client.FilterLogs(ctx, query)
	if err != nil {
		matches := errRegexp.FindStringSubmatch(err.Error())
		if len(matches) == 3 {
			from, err := strconv.ParseInt(matches[1], 16, 64)
			if err != nil {
				return nil, ethereum.FilterQuery{}, err
			}
			to, err := strconv.ParseInt(matches[2], 16, 64)
			if err != nil {
				return nil, ethereum.FilterQuery{}, err
			}
			return s.FilterLogs(ctx, client, big.NewInt(from), big.NewInt(to-1), topics, errRegexp, "")
		} else {
			return nil, ethereum.FilterQuery{}, err
		}
	}
	return logs, query, nil
}

func (s *EthService) HandleLogs(logs []types.Log, sigHash common.Hash, handle func(vLog types.Log)) uint64 {
	var count uint64
	for _, vLog := range logs {
		switch vLog.Topics[0].Hex() {
		case sigHash.Hex():
			if vLog.Removed {
				continue
			}
			handle(vLog)
			count++
		}
	}
	return count
}

const (
	RedisPrefix_ERC20Token = "ERC20Token::"
)

func (s *EthService) getERC20TokenModel(ctx context.Context, client *ethclient.Client, erc20ABI abi.ABI, contractAddress string, once *sync.Once) (*Erc20Tokens, error) {
	// first, from redis
	data, err := s.RedisClient.Get(ctx, RedisPrefix_ERC20Token+contractAddress).Result()
	if err != nil {
		if err != redis.Nil {
			return nil, err
		} else {
			// second, from mongo
			ts := time.Now()
			mongoColl := s.MongoDB.Collection(CollectionName_ERC20Tokens)
			var erc20Token Erc20Tokens
			err := mongoColl.FindOne(ctx, bson.D{{Key: "key", Value: contractAddress}}).Decode(&erc20Token)
			StatisticMongoCall(ctx, ts)
			if err != nil {
				if err != mongo.ErrNoDocuments {
					return nil, err
				} else {
					erc20Token.Value.Type = TokenTypeError
					once.Do(func() {
						// erc20 detection
						result, err := s.ERC20Detection(ctx, client, &erc20ABI, common.HexToAddress(contractAddress))
						if err != nil {
							return
						}
						erc20Token = *result
					})
				}
			}
			// cache, if is is found
			if erc20Token.Key != "" {
				data, err := json.Marshal(erc20Token)
				if err != nil {
					return nil, err
				}
				if err := s.RedisClient.Set(ctx, RedisPrefix_ERC20Token+contractAddress, string(data), time.Hour*24*3).Err(); err != nil {
					return nil, err
				}
			}
			return &erc20Token, nil
		}
	}
	var erc20Token Erc20Tokens
	if err := json.Unmarshal([]byte(data), &erc20Token); err != nil {
		return nil, err
	}
	return &erc20Token, nil
}

func (s *EthService) GenerateEthToken(ctx context.Context, client *ethclient.Client, erc20ABI abi.ABI, contractAddress string, balanceValue *uint256.Int, once *sync.Once) (*UserToken, error) {
	res := &UserToken{
		Address: contractAddress,       // address
		Balance: balanceValue.String(), // original balance
	}
	// first, from top100
	if token := config.InReputableTokens(contractAddress); token != nil {
		res.Type = token.Type         // type
		res.Name = token.Name         // name
		res.Symbol = token.Symbol     // symbol
		res.Decimals = token.Decimals // decimals
	} else {
		// second, from mongo
		erc20Token, err := s.getERC20TokenModel(ctx, client, erc20ABI, contractAddress, once)
		if err != nil {
			return nil, err
		}
		res.Type = erc20Token.Value.Type         // type
		res.Name = erc20Token.Value.Name         // name
		res.Symbol = erc20Token.Value.Symbol     // symbol
		res.Decimals = erc20Token.Value.Decimals // decimals
	}
	return res, nil
}

type Etherscan_GetABIResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Result  string `json:"result"`
}

var (
	ErrEtherscan_GetABINotOk = errors.New("response data status not ok")
)

func Etherscan_GetABI(contractAddress string) (*Etherscan_GetABIResponse, error) {
	url := "https://api.etherscan.io/api?module=contract&action=getabi&address=" + contractAddress + "&apikey=KBIF7HCBUFREP3WA9Y4JHJ28BR7KMCKCEQ"
	var respData Etherscan_GetABIResponse
	resp, err := simplehttp.Get(url, nil, &respData)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, simplehttp.ErrResponseStatusCodeNotEqualTo200
	}
	if respData.Status != "1" {
		return nil, simplehttp.ErrResponseDataStatusNotOK
	}
	return &respData, nil
}

func (s *EthService) getABIModel(ctx context.Context, rpcCtx *rpc.Context, req *doom_api.GetABIRequest, coll *mongo.Collection, ethServerAddress string) (*ABIs, *gerror.AppError) {
	logger := s.Logger
	if rpcCtx != nil {
		logger = rpcCtx.Logger
	}
	var abiModel ABIs
	if err := coll.FindOne(ctx, bson.D{{Key: "key", Value: req.Address}}).Decode(&abiModel); err != nil {
		if err != mongo.ErrNoDocuments {
			logger.Error("internal error", zap.NamedError("appError", err))
			return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
		} else {
			// second, fetch from etherscan api
			respData, err := Etherscan_GetABI(req.Address)
			if err != nil {
				logger.Error("bad request", zap.NamedError("appError", err))
				return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_WrongRequestParameters")).WithCode(response.StatusCodeBadRequest)
			}
			abiValue := ABIs_Value{
				ABI: respData.Result,
			}
			// check whether it is proxy
			client, err := ethclient.DialContext(ctx, ethServerAddress)
			if err != nil {
				logger.Error("bad request", zap.NamedError("appError", err))
				return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_WrongRequestParameters")).WithCode(response.StatusCodeBadRequest)
			}
			result, err := s.EvmProxyDetection(ctx, client, common.HexToAddress(req.Address))
			if err != nil {
				logger.Error("internal error", zap.NamedError("appError", err))
				return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
			}
			if result != nil {
				abiValue.IsProxy = true
				abiValue.ProxyType = result.Type
				abiValue.TargetAddress = eth.MixedcaseAddress(result.Target)
				abiValue.Immutable = result.Immutable
			}
			client.Close()
			// upsert the abi
			abiModel = ABIs{
				Key:   req.Address,
				Value: abiValue,
			}
			if _, err := coll.ReplaceOne(ctx, bson.D{{Key: "key", Value: req.Address}}, abiModel, options.Replace().SetUpsert(true)); err != nil {
				logger.Error("internal error", zap.NamedError("appError", err))
				return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
			}
		}
	}
	return &abiModel, nil
}

func (s *EthService) GetABI(ctx context.Context, rpcCtx *rpc.Context, req *doom_api.GetABIRequest) (*doom_api.GetABIResponse_Data, *gerror.AppError) {
	logger := s.Logger
	if rpcCtx != nil {
		logger = rpcCtx.Logger
	}
	// validate request basically
	if appError := s.ValidateRequest(ctx, rpcCtx, req); appError != nil {
		return nil, appError
	}
	chain, err := config.ChainByName(req.Chain)
	if err != nil {
		logger.Error("bad request", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_WrongRequestParameters")).WithCode(response.StatusCodeBadRequest)
	}
	var res doom_api.GetABIResponse_Data
	// first, fetch from mongo
	coll := s.MongoDB.Collection(CollectionName_ABIs)
	abiModel, appError := s.getABIModel(ctx, rpcCtx, req, coll, chain.Address)
	if appError != nil {
		return nil, appError
	}
	res.ABI = abiModel.Value.ABI
	res.IsProxy = abiModel.Value.IsProxy
	// check whether it is proxy
	if abiModel.Value.IsProxy {
		if abiModel.Value.ProxyType != "" {
			res.ProxyType = structpb.NewStringValue(abiModel.Value.ProxyType)
		}
		if abiModel.Value.TargetAddress != "" {
			targetAddress := eth.MixedcaseAddress(abiModel.Value.TargetAddress)
			res.TargetAddress = structpb.NewStringValue(targetAddress)
			// fetch target abi
			var targetModel ABIs
			if err = coll.FindOne(ctx, bson.D{{Key: "key", Value: targetAddress}}).Decode(&targetModel); err != nil {
				if err != mongo.ErrNoDocuments {
					logger.Error("internal error", zap.NamedError("appError", err))
					return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
				} else {
					// fetch from etherscan api
					respData, err := Etherscan_GetABI(targetAddress)
					if err == nil {
						// upsert the target abi
						targetModel = ABIs{
							Key: targetAddress,
							Value: ABIs_Value{
								ABI: respData.Result,
							},
						}
						if _, err := coll.ReplaceOne(ctx, bson.D{{Key: "key", Value: targetAddress}}, targetModel, options.Replace().SetUpsert(true)); err != nil {
							logger.Error("internal error", zap.NamedError("appError", err))
							return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
						}
					}
				}
			}
			if targetModel.Key != "" {
				res.ProxyABI = structpb.NewStringValue(targetModel.Value.ABI)
			}
		}
	}
	return &res, nil
}

func (s *EthService) GetTokenApprovals(ctx context.Context, rpcCtx *rpc.Context, req *doom_api.GetTokenApprovalsRequest) (*doom_api.GetTokenApprovalsResponse_Data, *gerror.AppError) {
	logger := s.Logger
	if rpcCtx != nil {
		logger = rpcCtx.Logger
	}
	// validate request basically
	if appError := s.ValidateRequest(ctx, rpcCtx, req); appError != nil {
		return nil, appError
	}
	chain, err := config.ChainByName(req.Chain)
	if err != nil {
		logger.Error("bad request", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_WrongRequestParameters")).WithCode(response.StatusCodeBadRequest)
	}
	// 1. tokens
	var approvals []*doom_api.GetTokenApprovalsResponse_Approval
	var unknownApprovals []*doom_api.GetTokenApprovalsResponse_UnknownApproval
	client, err := ethclient.DialContext(ctx, chain.Address)
	if err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	defer client.Close()
	handleEthChainResult, appError := s.HandleEthChain_ApprovalLog(ctx, rpcCtx, client, req.Address)
	if appError != nil {
		return nil, appError
	}
	coll := s.MongoDB.Collection(CollectionName_ERC20Tokens)
	for _, item := range handleEthChainResult.Approvals {
		// fetch token info
		var erc20Token Erc20Tokens
		if err := coll.FindOne(ctx, bson.D{{Key: "key", Value: item.Address}}).Decode(&erc20Token); err != nil {
			if err != mongo.ErrNoDocuments {
				logger.Error("internal error", zap.NamedError("appError", err))
				return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
			}
		}
		var unlimited bool
		// check uint256 max value and unlimited
		if item.Allowance.Cmp(eth.Uint256Max) == 0 {
			unlimited = true
		}
		allowance := structpb.NewStringValue(item.Allowance.String())
		if erc20Token.Value.Decimals > 0 && !unlimited {
			allowance = structpb.NewStringValue(eth.FormatFloat64(eth.Uint256ToFloat64(item.Allowance, erc20Token.Value.Decimals)))
		}
		if erc20Token.Value.Type == TokenTypeERC20 {
			approvals = append(approvals, &doom_api.GetTokenApprovalsResponse_Approval{
				Address:   item.Address,
				Target:    item.Target,
				Allowance: allowance,
				Unlimited: unlimited,
				Name:      erc20Token.Value.Name,
				Symbol:    erc20Token.Value.Symbol,
			})
		} else {
			unknownApprovals = append(unknownApprovals, &doom_api.GetTokenApprovalsResponse_UnknownApproval{
				Address:   item.Address,
				Target:    item.Target,
				Allowance: allowance,
				Unlimited: unlimited,
			})
		}
	}
	return &doom_api.GetTokenApprovalsResponse_Data{Approvals: approvals, UnknownApprovals: unknownApprovals}, nil
}

func (s *EthService) ERC20Detection(ctx context.Context, client *ethclient.Client, erc20ABI *abi.ABI, contractAddress common.Address) (*Erc20Tokens, error) {
	address := common.AddressEIP55(contractAddress).String()
	// first, fetch from mongo
	var erc20Token Erc20Tokens
	coll := s.MongoDB.Collection(CollectionName_ERC20Tokens)
	if err := coll.FindOne(ctx, bson.D{{Key: "key", Value: address}}).Decode(&erc20Token); err != nil {
		if err != mongo.ErrNoDocuments {
			return nil, err
		} else {
			erc20Token.Key = address
			erc20Token.Value.Type = TokenTypeError
			// second, fetch from web3
			// 1. erc20 detection directly
			bytecode, err := client.CodeAt(ctx, contractAddress, nil)
			if err != nil {
				return nil, err
			} else {
				if hexcode := hex.EncodeToString(bytecode); strings.Contains(hexcode, "a9059cbb") && strings.Contains(hexcode, "18160ddd") {
					// it is a erc20 token
					erc20Token.Value.Type = TokenTypeERC20
					name, err := eth.ERC20_Name(ctx, client, *erc20ABI, address)
					if err != nil {
						erc20Token.Value.Type = TokenTypeError
					} else {
						erc20Token.Value.Name = name
					}
					if erc20Token.Value.Type == TokenTypeERC20 {
						symbol, err := eth.ERC20_Symbol(ctx, client, *erc20ABI, address)
						if err != nil {
							erc20Token.Value.Type = TokenTypeError
						} else {
							erc20Token.Value.Symbol = strings.ToUpper(symbol)
						}
					}
					if erc20Token.Value.Type == TokenTypeERC20 {
						decimals, err := eth.ERC20_Decimals(ctx, client, *erc20ABI, address)
						if err != nil {
							erc20Token.Value.Type = TokenTypeError
						} else {
							erc20Token.Value.Decimals = decimals
						}
					}
				} else {
					erc20Token.Value.Type = TokenTypeNotERC20
				}
			}
			// 2. evm proxy detection, if it is not a direct ERC20
			if erc20Token.Value.Type != TokenTypeERC20 {
				result, err := s.EvmProxyDetection(ctx, client, contractAddress)
				if err != nil {
					return nil, err
				}
				if result != nil {
					// it is a proxy
					targetAddress := eth.MixedcaseAddress(result.Target)
					// fetch target address model from mongo
					var targetToken Erc20Tokens
					time.Sleep(500 * time.Millisecond) // !!! important for reducing 'Too Many Requests' error
					// get infos of target address from web3
					bytecode, err := client.CodeAt(ctx, common.HexToAddress(targetAddress), nil)
					if err != nil {
						return nil, err
					} else {
						if hexcode := hex.EncodeToString(bytecode); strings.Contains(hexcode, "a9059cbb") && strings.Contains(hexcode, "18160ddd") {
							// it is a erc20 token
							targetToken.Value.Type = TokenTypeERC20
							erc20ABI, err := ethabi.GetABI(ethabi.ERC20ABI)
							if err != nil {
								return nil, err
							}
							name, err := eth.ERC20_Name(ctx, client, *erc20ABI, address)
							if err != nil {
								targetToken.Value.Type = TokenTypeError
							} else {
								targetToken.Value.Name = name
							}
							if targetToken.Value.Type == TokenTypeERC20 {
								symbol, err := eth.ERC20_Symbol(ctx, client, *erc20ABI, address)
								if err != nil {
									targetToken.Value.Type = TokenTypeError
								} else {
									targetToken.Value.Symbol = strings.ToUpper(symbol)
								}
							}
							if targetToken.Value.Type == TokenTypeERC20 {
								decimals, err := eth.ERC20_Decimals(ctx, client, *erc20ABI, address)
								if err != nil {
									targetToken.Value.Type = TokenTypeError
								} else {
									targetToken.Value.Decimals = decimals
								}
							}
						} else {
							targetToken.Value.Type = TokenTypeNotERC20
						}
					}
					// update erc20Token from targetToken
					erc20Token.Value.Type = targetToken.Value.Type
					if erc20Token.Value.Name == "" {
						erc20Token.Value.Name = targetToken.Value.Name
					}
					if erc20Token.Value.Symbol == "" {
						erc20Token.Value.Symbol = targetToken.Value.Symbol
					}
					if erc20Token.Value.Decimals == 0 {
						erc20Token.Value.Decimals = targetToken.Value.Decimals
					}
				}
			}
			// upsert the erc20 record
			if _, err = coll.ReplaceOne(ctx, bson.D{{Key: "key", Value: address}}, erc20Token, options.Replace().SetUpsert(true)); err != nil {
				return nil, err
			}
		}
	}
	return &erc20Token, nil
}

// evm proxy detection
// types -------------------------------------------------------------------------------------------------------------

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

// EvmProxyDetection -------------------------------------------------------------------------------------------------------------

func (s *EthService) EvmProxyDetection(ctx context.Context, client *ethclient.Client, contractAddress common.Address) (result *EvmProxyDetectionResult, err error) {
	address := common.AddressEIP55(contractAddress).String()
	// first, fetch from mongo
	var proxyToken ProxyToken
	coll := s.MongoDB.Collection(CollectionName_ProxyTokens)
	defer func() {
		if err != nil {
			if strings.Contains(err.Error(), "network is unreachable") ||
				strings.Contains(err.Error(), "rate exceeded") {
				return
			}
		}
		if proxyToken.Key == "" {
			proxyToken.Key = address
			if err == nil {
				proxyToken.Value.IsProxy = true
				proxyToken.Value.Type = result.Type
				proxyToken.Value.Target = result.Target
				proxyToken.Value.Immutable = result.Immutable
			} else {
				err = nil // don't export, because it is not an internal error
			}
			if blockNumber, e := client.BlockNumber(ctx); e == nil {
				proxyToken.Value.BlockNumber = blockNumber
			}
			_, err = coll.ReplaceOne(ctx, bson.D{{Key: "key", Value: address}}, proxyToken, options.Replace().SetUpsert(true))
		}
	}()
	if err = coll.FindOne(ctx, bson.D{{Key: "key", Value: address}}).Decode(&proxyToken); err != nil {
		if err != mongo.ErrNoDocuments {
			return nil, err
		} else {
			// second, fetch from web3
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
	}
	if proxyToken.Value.IsProxy {
		result = &EvmProxyDetectionResult{
			Target:    proxyToken.Value.Target,
			Type:      proxyToken.Value.Type,
			Immutable: proxyToken.Value.Immutable,
		}
	}
	return
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
