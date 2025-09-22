package service

import (
	"context"
	"strconv"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	doom_console_api "github.com/nextsurfer/doom-console/api"
	"github.com/nextsurfer/doom-console/api/response"
	ethabi "github.com/nextsurfer/doom-console/internal/pkg/eth/abi"
	doom_console_mongo "github.com/nextsurfer/doom-console/internal/pkg/mongo"
	"github.com/nextsurfer/doom-console/internal/pkg/uniswap"
	gerror "github.com/nextsurfer/ground/pkg/err"
	"github.com/nextsurfer/ground/pkg/rpc"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

type DoomConsoleService struct {
	*EthService

	Logger  *zap.Logger
	MongoDB *mongo.Database
}

func NewDoomConsoleService(ctx context.Context, logger *zap.Logger, mongoDB *mongo.Database) (*DoomConsoleService, error) {
	DoomConsoleService := &DoomConsoleService{
		Logger:  logger,
		MongoDB: mongoDB,
	}
	DoomConsoleService.EthService = NewEthService(DoomConsoleService)
	return DoomConsoleService, nil
}

func (s *DoomConsoleService) ListReputableTokens(ctx context.Context, rpcCtx *rpc.Context, req *doom_console_api.ListReputableTokensRequest) (*doom_console_api.ListReputableTokensResponse_Data, *gerror.AppError) {
	var list []*doom_console_api.ListReputableTokensResponse_ReputableToken
	coll := s.MongoDB.Collection(doom_console_mongo.CollectionName_ReputableTokens)
	total, err := coll.CountDocuments(ctx, bson.D{})
	if err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	cursor, err := coll.Find(ctx, bson.D{}, options.Find().SetSkip(req.PageNumber*req.PageSize).SetLimit(req.PageSize))
	if err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	for cursor.Next(ctx) {
		var token doom_console_mongo.ReputableTokens
		if err := cursor.Decode(&token); err != nil {
			rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
			return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
		}
		var exchange string
		if len(token.Value.BinanceSymbols) > 0 {
			exchange = "Binance"
		} else if len(token.Value.OkxInstIds) > 0 {
			exchange = "OKX"
		}
		one := &doom_console_api.ListReputableTokensResponse_ReputableToken{
			Address:  token.Key,
			Type:     token.Value.Type,
			Symbol:   token.Value.Symbol,
			Name:     token.Value.Name,
			Decimals: uint32(token.Value.Decimals),
			Exchange: exchange,
		}
		list = append(list, one)
	}
	if err := cursor.Err(); err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	return &doom_console_api.ListReputableTokensResponse_Data{
		Total: total,
		List:  list,
	}, nil
}

func (s *DoomConsoleService) UniswapInfo(ctx context.Context, rpcCtx *rpc.Context) (*doom_console_api.UniswapInfoResponse_Data, *gerror.AppError) {
	client, err := ethclient.DialContext(ctx, "wss://mainnet.infura.io/ws/v3/ae1cb3ad3a4542e294f99a5f92be46c9")
	if err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	defer client.Close()
	coll := s.MongoDB.Collection(doom_console_mongo.CollectionName_UniswapTokens)
	/// v2 info
	v2FactoryABI, err := ethabi.GetABI(ethabi.UniswapV2FactoryABI)
	if err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	// total real time
	allPairsLength, err := uniswap.UniswapV2_Factory_AllPairsLength(ctx, client, *v2FactoryABI)
	if err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	// total in db
	v2Total, err := coll.CountDocuments(ctx, bson.D{{Key: "value.type", Value: doom_console_mongo.UniswapTokenTypeV2}})
	if err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	/// v3 info
	// timestamp
	var uniswapV3BlockNumber doom_console_mongo.UniswapV3BlockNumber
	if err := coll.FindOne(ctx, bson.D{{Key: "key", Value: doom_console_mongo.KeyUniswapV3BlockNumber}}).Decode(&uniswapV3BlockNumber); err != nil {
		if err != mongo.ErrNoDocuments {
			rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
			return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
		}
	}
	// total in db
	v3Total, err := coll.CountDocuments(ctx, bson.D{{Key: "value.type", Value: doom_console_mongo.UniswapTokenTypeV3}})
	if err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	return &doom_console_api.UniswapInfoResponse_Data{
		V2Info: &doom_console_api.UniswapInfoResponse_V2Info{
			TotalRealTime: allPairsLength,
			TotalInDB:     uint64(v2Total),
			DiffValue:     allPairsLength - uint64(v2Total),
		},
		V3Info: &doom_console_api.UniswapInfoResponse_V3Info{
			Timestamp: uniswapV3BlockNumber.Value.Timestamp,
			TotalInDB: uint64(v3Total),
		},
	}, nil
}

func (s *DoomConsoleService) Erc20TokensInfo(ctx context.Context, rpcCtx *rpc.Context) (*doom_console_api.Erc20TokensInfoResponse_Data, *gerror.AppError) {
	client, err := ethclient.DialContext(ctx, "wss://mainnet.infura.io/ws/v3/ae1cb3ad3a4542e294f99a5f92be46c9")
	if err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	defer client.Close()
	// header.Number
	var headerNumber uint64
	header, err := client.HeaderByNumber(ctx, nil)
	if err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	headerNumber = header.Number.Uint64()
	// toBlock number
	coll := s.MongoDB.Collection(doom_console_mongo.CollectionName_ERC20Tokens)
	recordTotal, err := coll.CountDocuments(ctx, bson.D{})
	if err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	var erc20TokenBlockNumber uint64
	var erc20TokenBlockNumberM doom_console_mongo.Erc20TokenBlockNumber
	if err := coll.FindOne(ctx, bson.D{{Key: "key", Value: "erc20TokenBlockNumber"}}).Decode(&erc20TokenBlockNumberM); err != nil {
		if err != mongo.ErrNoDocuments {
			rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
			return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
		}
	} else {
		erc20TokenBlockNumber, err = strconv.ParseUint(erc20TokenBlockNumberM.Value, 10, 64)
		if err != nil {
			rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
			return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
		}
		recordTotal -= 1
	}

	return &doom_console_api.Erc20TokensInfoResponse_Data{
		HeaderNumber:  headerNumber,
		ToBlockNumber: erc20TokenBlockNumber,
		NumberDiff:    headerNumber - erc20TokenBlockNumber,
		Days:          uint32((headerNumber - erc20TokenBlockNumber) / (24 * 60 * (60 / 12))),
		TotalTokens:   uint64(recordTotal),
	}, nil
}

func (s *DoomConsoleService) Erc20TokensQuery(ctx context.Context, rpcCtx *rpc.Context, req *doom_console_api.Erc20TokensQueryRequest) (*doom_console_api.Erc20TokensQueryResponse_Data, *gerror.AppError) {
	coll := s.MongoDB.Collection(doom_console_mongo.CollectionName_ERC20Tokens)
	var erc20Token doom_console_mongo.Erc20Tokens
	if err := coll.FindOne(ctx, bson.D{{Key: "key", Value: req.ContractAddress}}).Decode(&erc20Token); err != nil {
		if err != mongo.ErrNoDocuments {
			rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
			return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
		} else {
			rpcCtx.Logger.Error("bad request", zap.NamedError("appError", err))
			return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_UnsupportedToken")).WithCode(response.StatusCodeUnsupportedCryptocurrency)
		}
	}
	return &doom_console_api.Erc20TokensQueryResponse_Data{
		Address:  erc20Token.Key,
		Type:     erc20Token.Value.Type,
		Name:     erc20Token.Value.Name,
		Symbol:   erc20Token.Value.Symbol,
		Decimals: uint32(erc20Token.Value.Decimals),
		Priced:   erc20Token.Value.Priced,
		Checked:  erc20Token.Value.Checked,
	}, nil
}

func (s *DoomConsoleService) ListErrorTokens(ctx context.Context, rpcCtx *rpc.Context, req *doom_console_api.ListErrorTokensRequest) (*doom_console_api.ListErrorTokensResponse_Data, *gerror.AppError) {
	var list []*doom_console_api.ListErrorTokensResponse_ErrorToken
	coll := s.MongoDB.Collection(doom_console_mongo.CollectionName_ERC20Tokens)
	filter := bson.D{{Key: "value.type", Value: doom_console_mongo.TokenTypeError}}
	if req.Checked {
		filter = append(filter, bson.E{Key: "value.checked", Value: true})
	} else {
		filter = append(filter, bson.E{Key: "$or", Value: bson.A{bson.D{{Key: "value.checked", Value: bson.D{{Key: "$exists", Value: false}}}}, bson.D{{Key: "value.checked", Value: false}}}})
	}
	total, err := coll.CountDocuments(ctx, filter)
	if err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	cursor, err := coll.Find(ctx,
		filter,
		options.Find().SetSkip(req.PageNumber*req.PageSize).SetLimit(req.PageSize),
	)
	if err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	for cursor.Next(ctx) {
		var erc20Token doom_console_mongo.Erc20Tokens
		if err := cursor.Decode(&erc20Token); err != nil {
			rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
			return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
		}
		one := &doom_console_api.ListErrorTokensResponse_ErrorToken{
			Id:      erc20Token.ID.Hex(),
			Address: erc20Token.Key,
			Type:    erc20Token.Value.Type,
			Symbol:  erc20Token.Value.Symbol,
			Name:    erc20Token.Value.Name,
			Checked: erc20Token.Value.Checked,
		}
		list = append(list, one)
	}
	if err := cursor.Err(); err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	return &doom_console_api.ListErrorTokensResponse_Data{
		Total: total,
		List:  list,
	}, nil
}

func (s *DoomConsoleService) CheckErrorToken(ctx context.Context, rpcCtx *rpc.Context, req *doom_console_api.CheckErrorTokenRequest) *gerror.AppError {
	coll := s.MongoDB.Collection(doom_console_mongo.CollectionName_ERC20Tokens)
	recordID, err := primitive.ObjectIDFromHex(req.Id)
	if err != nil {
		rpcCtx.Logger.Error("bad request", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_WrongRequestParameters")).WithCode(response.StatusCodeWrongParameters)
	}
	var record doom_console_mongo.Erc20Tokens
	if err := coll.FindOne(ctx, bson.D{{Key: "_id", Value: recordID}}).Decode(&record); err != nil {
		if err != mongo.ErrNoDocuments {
			rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
			return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
		} else {
			rpcCtx.Logger.Error("bad request", zap.NamedError("appError", err))
			return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_BadRequest")).WithCode(response.StatusCodeBadRequest)
		}
	}
	if _, err := coll.UpdateOne(ctx, bson.D{{Key: "_id", Value: recordID}}, bson.D{{Key: "$set", Value: bson.D{{Key: "value.checked", Value: !record.Value.Checked}}}}); err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	return nil
}

func (s *DoomConsoleService) DetectErrorToken(ctx context.Context, rpcCtx *rpc.Context, req *doom_console_api.DetectErrorTokenRequest) (*doom_console_api.DetectErrorTokenResponse_Data, *gerror.AppError) {
	coll := s.MongoDB.Collection(doom_console_mongo.CollectionName_ERC20Tokens)
	recordID, err := primitive.ObjectIDFromHex(req.Id)
	if err != nil {
		rpcCtx.Logger.Error("bad request", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_WrongRequestParameters")).WithCode(response.StatusCodeWrongParameters)
	}
	var record doom_console_mongo.Erc20Tokens
	if err := coll.FindOne(ctx, bson.D{{Key: "_id", Value: recordID}}).Decode(&record); err != nil {
		if err != mongo.ErrNoDocuments {
			rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
			return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
		} else {
			rpcCtx.Logger.Error("bad request", zap.NamedError("appError", err))
			return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_BadRequest")).WithCode(response.StatusCodeBadRequest)
		}
	}
	client, err := ethclient.DialContext(ctx, "wss://mainnet.infura.io/ws/v3/ae1cb3ad3a4542e294f99a5f92be46c9")
	if err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	defer client.Close()
	// erc20ABI
	erc20ABI, err := ethabi.GetABI(ethabi.ERC20ABI)
	if err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	token, err := s.ERC20Detection(ctx, client, erc20ABI, common.HexToAddress(record.Key))
	if err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	var result string
	if token.Value.Type == doom_console_mongo.TokenTypeERC20 {
		result = rpcCtx.Localizer.Localize("Business_ErrorTokenDetectionERC20")
	} else {
		result = rpcCtx.Localizer.Localize("Business_ErrorTokenDetectionNotERC20")
	}
	return &doom_console_api.DetectErrorTokenResponse_Data{Result: result}, nil
}

func (s *DoomConsoleService) rpcServerDetection(ctx context.Context, rpcCtx *rpc.Context, rpcServer string) (*doom_console_api.RpcServerDetectionResponse_RpcServerState, *gerror.AppError) {
	res := doom_console_api.RpcServerDetectionResponse_RpcServerState{
		RpcServer: rpcServer,
	}
	client, err := ethclient.Dial(rpcServer)
	if err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	defer client.Close()
	// status
	res.Status = 1
	// blockNumber
	ts1 := time.Now()
	blockNumber, err := client.BlockNumber(ctx)
	if err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	res.BlockNumber = blockNumber
	// delay
	delay := time.Since(ts1) / time.Millisecond
	res.Delay = int32(delay)
	return &res, nil
}

func (s *DoomConsoleService) RpcServerDetection(ctx context.Context, rpcCtx *rpc.Context, req *doom_console_api.RpcServerDetectionRequest) (*doom_console_api.RpcServerDetectionResponse_Data, *gerror.AppError) {
	state, appError := s.rpcServerDetection(ctx, rpcCtx, req.ServerUrl)
	if appError != nil {
		return nil, appError
	}
	return &doom_console_api.RpcServerDetectionResponse_Data{
		State: state,
	}, nil
}
