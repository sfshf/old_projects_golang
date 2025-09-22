package service

import (
	"container/list"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	jsoniter "github.com/json-iterator/go"
	"github.com/juju/ratelimit"
	doom_api "github.com/nextsurfer/doom-go/api"
	"github.com/nextsurfer/doom-go/api/response"
	"github.com/nextsurfer/doom-go/internal/common/config"
	"github.com/nextsurfer/doom-go/internal/common/simplehttp"
	gerror "github.com/nextsurfer/ground/pkg/err"
	"github.com/nextsurfer/ground/pkg/rpc"
	"go.uber.org/zap"
)

type MarketService struct {
	*DoomService

	BinanceApiMu                    sync.Mutex
	BinanceApiTotalWeight           int
	BinanceApiRequestWeightList     *list.List
	OkxApiHistoryMarkPriceRatelimit *ratelimit.Bucket
	OkxWsConn                       *websocket.Conn
	BinanceWsConn                   *websocket.Conn
	BinanceSymbols                  []string
	SpotPrices                      map[string]*sync.Map // baseCoin->{symbol->price}
}

func NewMarketService(DoomService *DoomService, simple bool) (*MarketService, error) {
	s := &MarketService{
		DoomService:                     DoomService,
		BinanceApiRequestWeightList:     list.New(),
		OkxApiHistoryMarkPriceRatelimit: ratelimit.NewBucket(2*time.Second, 10),
	}
	if !simple {
		// 1. binance symbols
		var binanceSymbols []string
		// deduplicate symbols. !!! important !!!
		for _, token := range config.DefaultReputableTokens {
			for _, binanceSymbol := range token.BinanceSymbol {
				var has bool
				for _, symbol := range binanceSymbols {
					if binanceSymbol == symbol {
						has = true
						break
					}
				}
				if !has {
					binanceSymbols = append(binanceSymbols, binanceSymbol)
				}
			}
		}
		s.BinanceSymbols = binanceSymbols
		// 2. spot prices map
		spotPrices := make(map[string]*sync.Map, 2)
		spotPrices[BaseCoinUSDT] = &sync.Map{}
		spotPrices[BaseCoinUSD] = &sync.Map{}
		for _, token := range config.DefaultReputableTokens {
			if len(token.BinanceSymbol) > 0 {
				for _, binanceSymbol := range token.BinanceSymbol {
					if strings.HasSuffix(binanceSymbol, BaseCoinUSDT) {
						spotPrices[BaseCoinUSDT].Store(token.Symbol, "unknown")
					} else if strings.HasSuffix(binanceSymbol, BaseCoinUSD) {
						spotPrices[BaseCoinUSD].Store(token.Symbol, "unknown")
					}
				}
			}
			if len(token.OkxInstIds) > 0 {
				for _, okxInstId := range token.OkxInstIds {
					if strings.HasSuffix(okxInstId, BaseCoinUSDT) {
						spotPrices[BaseCoinUSDT].Store(token.Symbol, "unknown")
					} else if strings.HasSuffix(okxInstId, BaseCoinUSD) {
						spotPrices[BaseCoinUSD].Store(token.Symbol, "unknown")
					}
				}
			}
		}
		s.SpotPrices = spotPrices
		// 3. binance prices
		binanceWsConn, err := s.SubscribeBinancePrices()
		if err != nil {
			return nil, err
		}
		s.BinanceWsConn = binanceWsConn
		// 4. okx prices
		okxWsConn, err := s.SubscribeOKXPrices()
		if err != nil {
			return nil, err
		}
		s.OkxWsConn = okxWsConn
	}
	return s, nil
}

type BinanceApiRequestWeight struct {
	Timestamp int64
	Weight    int
}

func (s *MarketService) CheckBinanceRateLimit(rpcCtx *rpc.Context, wight int) *gerror.AppError {
	logger := s.Logger
	if rpcCtx != nil {
		logger = rpcCtx.Logger
	}
	s.BinanceApiMu.Lock()
	defer s.BinanceApiMu.Unlock()
	now := time.Now()
	for e := s.BinanceApiRequestWeightList.Front(); e != nil; e = e.Next() {
		reqWeight := e.Value.(BinanceApiRequestWeight)
		if now.Sub(time.UnixMilli(reqWeight.Timestamp)) > time.Minute*1 {
			s.BinanceApiRequestWeightList.Remove(e)
			s.BinanceApiTotalWeight -= reqWeight.Weight
		}
	}
	if s.BinanceApiTotalWeight > 45 {
		err := errors.New("request rate limit")
		logger.Error("bad request", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_TooManyRequests")).WithCode(response.StatusCodeTooManyRequests)
	}
	s.BinanceApiRequestWeightList.PushBack(BinanceApiRequestWeight{
		Timestamp: now.UnixMilli(),
		Weight:    wight,
	})
	s.BinanceApiTotalWeight += wight
	return nil
}

// binance prices -----------------------------------------------------------------------------------------

type StreamEvent struct {
	Stream string                      `json:"stream,omitempty"`
	Data   AggregateTradeStreamPayload `json:"data,omitempty"`
}

type AggregateTradeStreamPayload struct {
	Symbol string `json:"s,omitempty"`
	Price  string `json:"p,omitempty"`
}

func (s *MarketService) SubscribeBinancePrices() (*websocket.Conn, error) {
	// 1. stream name
	var streamNames []string
	for _, binanceSymbol := range s.BinanceSymbols {
		if !slices.Contains(config.ExcludedBinanceSymbols, binanceSymbol) {
			streamNames = append(streamNames, strings.ToLower(binanceSymbol)+"@aggTrade")
		}
	}
	url := "wss://data-stream.binance.vision/stream?streams=" + strings.Join(streamNames, "/")
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		s.Logger.Error("dial Binance websocket connection fail", zap.NamedError("appError", err))
		return nil, err
	}
	conn.SetPingHandler(nil) // important !!!
	ctx, cancel := context.WithCancel(context.Background())
	// read goroutine
	go func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				s.Logger.Info("Binance websocket read goroutine has been canceled")
				return
			default:
				var streamEvent StreamEvent
				if err := conn.ReadJSON(&streamEvent); err != nil {
					s.Logger.Error("Binance wss channel read error", zap.NamedError("appError", err))
					continue
				}
				if strings.HasSuffix(streamEvent.Data.Symbol, BaseCoinUSDT) {
					s.SpotPrices[BaseCoinUSDT].Store(strings.TrimSuffix(streamEvent.Data.Symbol, BaseCoinUSDT), streamEvent.Data.Price)
				} else if strings.HasSuffix(streamEvent.Data.Symbol, BaseCoinUSD) {
					s.SpotPrices[BaseCoinUSD].Store(strings.TrimSuffix(streamEvent.Data.Symbol, BaseCoinUSD), streamEvent.Data.Price)
				}
			}
		}
	}(ctx)
	// refresh ws connection goroutine
	go func() {
		defer conn.Close()
		defer cancel()               // cancel read goroutine
		<-time.After(time.Hour * 23) // due to 'A single connection to stream.binance.com is only valid for 24 hours'
		// refresh ws connection
		newConn, err := s.SubscribeBinancePrices()
		if err != nil {
			s.Logger.Error("SubscribeBinancePrices error", zap.NamedError("appError", err))
			return
		}
		s.BinanceWsConn = newConn
	}()
	return conn, nil
}

// okx prices -----------------------------------------------------------------------------------------

type OkxSubscribeRequest struct {
	Op   string            `json:"op,omitempty"`
	Args []OkxSubscribeArg `json:"args,omitempty"`
}

type OkxSubscribeArg struct {
	Channel string `json:"channel,omitempty"`
	InstId  string `json:"instId,omitempty"`
}

type OkxSubscribeResponse struct {
	Event  string          `json:"event,omitempty"`
	Code   string          `json:"code,omitempty"`
	Msg    string          `json:"msg,omitempty"`
	Arg    OkxSubscribeArg `json:"arg,omitempty"`
	ConnId string          `json:"connId,omitempty"`
}

type MarkPriceMessage struct {
	Arg  OkxSubscribeArg `json:"arg,omitempty"`
	Data []MarkPriceData `json:"data,omitempty"`
}

type MarkPriceData struct {
	InstType string `json:"instType,omitempty"`
	InstId   string `json:"instId,omitempty"`
	MarkPx   string `json:"markPx,omitempty"`
	Ts       string `json:"ts,omitempty"`
}

func (s *MarketService) SubscribeOKXPrices() (*websocket.Conn, error) {
	conn, _, err := websocket.DefaultDialer.Dial("wss://ws.okx.com:8443/ws/v5/public", nil)
	if err != nil {
		s.Logger.Error("dial OKX websocket connection fail", zap.NamedError("appError", err))
		return nil, err
	}
	// 1. subscribe to Mark price channel
	var args []OkxSubscribeArg
	for _, token := range config.DefaultReputableTokens {
		for _, instId := range token.OkxInstIds {
			if strings.HasSuffix(instId, "USDT") || strings.HasSuffix(instId, "USD") {
				if !slices.Contains(config.ExcludedOkxInstIds, instId) {
					args = append(args, OkxSubscribeArg{
						Channel: "mark-price",
						InstId:  instId,
					})
				}
			}
		}
	}
	subscribeRequest := OkxSubscribeRequest{
		Op:   "subscribe",
		Args: args,
	}
	if err := conn.WriteJSON(subscribeRequest); err != nil {
		s.Logger.Error("write OKX subscribe message fail", zap.NamedError("appError", err))
		return nil, err
	}
	var subscribeResponse OkxSubscribeResponse
	if err := conn.ReadJSON(&subscribeResponse); err != nil {
		s.Logger.Error("read OKX subscribe message fail", zap.NamedError("appError", err))
		return nil, err
	}
	if subscribeResponse.Event == "subscribe" {
		s.Logger.Info("subscribe OKX mark price channel success")
	} else if subscribeResponse.Event == "error" {
		return nil, fmt.Errorf("subscribe OKX mark price channel fail: %v", subscribeResponse.Msg)
	}
	ctx, cancel := context.WithCancel(context.Background())
	// read goroutine
	go func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				s.Logger.Info("OKX websocket read goroutine has been canceled")
				return
			default:
				_, message, err := conn.ReadMessage()
				if err != nil {
					s.Logger.Error("OKX wss channel read error", zap.NamedError("appError", err))
					return
				}
				if string(message) == "pong" { // keep connection healthy
					continue
				} else if event := jsoniter.Get(message, "event").ToString(); event == "unsubscribe" || event == "error" {
					var unsubscribeResponse OkxSubscribeResponse
					if err := json.Unmarshal(message, &unsubscribeResponse); err != nil {
						s.Logger.Error("OKX unmarshal SubscribeResponse error", zap.NamedError("appError", err))
						return
					}
					if unsubscribeResponse.Event == "unsubscribe" {
						s.Logger.Info("unsubscribe OKX mark price channel success")
					} else if unsubscribeResponse.Event == "error" {
						s.Logger.Error("unsubscribe OKX mark price channel fail: " + unsubscribeResponse.Msg)
					}
					return
				}
				var markPriceMessage MarkPriceMessage
				if err := json.Unmarshal(message, &markPriceMessage); err != nil {
					s.Logger.Error("OKX unmarshal MarkPriceMessage error", zap.NamedError("appError", err))
					return
				}
				if len(markPriceMessage.Data) == 0 {
					continue
				}
				switch markPriceMessage.Arg.InstId {
				case "OKB-USDT":
					s.SpotPrices[BaseCoinUSDT].Store("OKB", markPriceMessage.Data[0].MarkPx)
				case "CRO-USDT":
					s.SpotPrices[BaseCoinUSDT].Store("CRO", markPriceMessage.Data[0].MarkPx)
				default:
					if strings.HasSuffix(markPriceMessage.Arg.InstId, "USDT") {
						s.SpotPrices[BaseCoinUSDT].Store(strings.TrimSuffix(strings.ReplaceAll(markPriceMessage.Arg.InstId, "-", ""), "USDT"), markPriceMessage.Data[0].MarkPx)
					} else if strings.HasSuffix(markPriceMessage.Arg.InstId, "USD") {
						s.SpotPrices[BaseCoinUSD].Store(strings.TrimSuffix(strings.ReplaceAll(markPriceMessage.Arg.InstId, "-", ""), "USD"), markPriceMessage.Data[0].MarkPx)
					}
				}
			}
		}
	}(ctx)
	// write goroutine
	go func(ctx context.Context) {
		ticker := time.NewTicker(time.Second * 25) // keep connection healthy
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				err := conn.WriteMessage(websocket.TextMessage, []byte("ping"))
				if err != nil {
					s.Logger.Error("OKX write ping fail", zap.NamedError("appError", err))
					return
				}
			case <-ctx.Done():
				s.Logger.Info("OKX websocket write goroutine has been canceled")
				return
			}
		}
	}(ctx)
	// refresh ws connection goroutine
	go func() {
		defer conn.Close()
		defer cancel() // cancel read goroutine
		<-time.After(time.Hour * 23)
		// unsubscribe market stream
		unsubscribeRequest := OkxSubscribeRequest{
			Op:   "unsubscribe",
			Args: args,
		}
		if err := conn.WriteJSON(unsubscribeRequest); err != nil {
			s.Logger.Error("write OKX unsubscribe message fail", zap.NamedError("appError", err))
		}
		// refresh ws connection
		newConn, err := s.SubscribeOKXPrices()
		if err != nil {
			s.Logger.Error("SubscribeOKXPrices error", zap.NamedError("appError", err))
			return
		}
		s.OkxWsConn = newConn
	}()
	return conn, nil
}

const (
	BaseCoinUSDT = "USDT"
	BaseCoinUSD  = "USD"
)

var (
	ErrUnsupportedToken = errors.New("unsupported token")
)

func (s *MarketService) GetLatestSpotPrice(ctx context.Context, rpcCtx *rpc.Context, req *doom_api.GetLatestSpotPriceRequest) (*doom_api.GetLatestSpotPriceResponse_Data, *gerror.AppError) {
	logger := s.Logger
	if rpcCtx != nil {
		logger = rpcCtx.Logger
	}
	// validate request basically
	if appError := s.ValidateRequest(ctx, rpcCtx, req); appError != nil {
		return nil, appError
	}
	symbol := strings.ToUpper(req.Symbol)
	var price string
	if symbol == "USDT" || symbol == "USD" {
		price = "1.00"
	} else {
		if symbol == "WETH" {
			symbol = "ETH"
		}
		spotPrice, has := s.SpotPrices[req.BaseCoin].Load(symbol)
		if !has {
			logger.Error("bad request", zap.NamedError("appError", ErrUnsupportedToken))
			return nil, gerror.NewError(ErrUnsupportedToken).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_WrongRequestParameters")).WithCode(response.StatusCodeUnsupportedCryptocurrency)
		}
		price = spotPrice.(string)
		if price == "unknown" {
			logger.Error("bad request", zap.NamedError("appError", ErrUnsupportedToken))
			return nil, gerror.NewError(ErrUnsupportedToken).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_UnknownSpotPrice")).WithCode(response.StatusCodeUnknownSpotPrice)
		}
	}
	return &doom_api.GetLatestSpotPriceResponse_Data{
		Price: price,
	}, nil
}

func (s *MarketService) GetLatestSpotPrices(ctx context.Context, rpcCtx *rpc.Context, req *doom_api.GetLatestSpotPricesRequest) (*doom_api.GetLatestSpotPricesResponse_Data, *gerror.AppError) {
	logger := s.Logger
	if rpcCtx != nil {
		logger = rpcCtx.Logger
	}
	// validate request basically
	if appError := s.ValidateRequest(ctx, rpcCtx, req); appError != nil {
		return nil, appError
	}
	res := make([]string, len(req.Symbols))
	for idx, symbol := range req.Symbols {
		var price string
		var spotPrice any
		var has bool
		if strings.EqualFold(symbol, "USDT") || strings.EqualFold(symbol, "USD") {
			price = "1.00"
		} else if strings.EqualFold(symbol, "WETH") {
			spotPrice, has = s.SpotPrices[req.BaseCoin].Load("ETH")
		} else {
			spotPrice, has = s.SpotPrices[req.BaseCoin].Load(strings.ToUpper(symbol))
		}
		if !has {
			logger.Error("bad request", zap.NamedError("appError", ErrUnsupportedToken))
			return nil, gerror.NewError(ErrUnsupportedToken).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_WrongRequestParameters")).WithCode(response.StatusCodeUnsupportedCryptocurrency)
		}
		price = spotPrice.(string)
		if price == "unknown" {
			logger.Error("bad request", zap.NamedError("appError", ErrUnsupportedToken))
			return nil, gerror.NewError(ErrUnsupportedToken).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_UnknownSpotPrice")).WithCode(response.StatusCodeUnknownSpotPrice)
		}
		res[idx] = price
	}
	return &doom_api.GetLatestSpotPricesResponse_Data{List: res}, nil
}

// GetTokens 是支持时间范围内查询，也就是用币安定期查询的token列表
// 另一种情况，是支持实时价格查询， 在GetTokens的基础上加了两个特殊的，USDT和WETH 。实时价格中， usdt一直是1，weth的价格是eth的价格。
func (s *MarketService) GetTokens(ctx context.Context, rpcCtx *rpc.Context, req *doom_api.GetTokensRequest) (*doom_api.GetTokensResponse_Data, *gerror.AppError) {
	// validate request basically
	if appError := s.ValidateRequest(ctx, rpcCtx, req); appError != nil {
		return nil, appError
	}
	var res []*doom_api.GetTokensResponse_Token
	for _, item := range config.DefaultReputableTokens {
		symbol := strings.ToUpper(item.Symbol)
		if slices.Contains(config.ExcludedReputableTokens, symbol) {
			continue
		}
		res = append(res, &doom_api.GetTokensResponse_Token{
			Symbol: symbol,
			Name:   item.Name,
		})
	}
	return &doom_api.GetTokensResponse_Data{List: res}, nil
}

func (s *MarketService) getUIKlines_Binance(rpcCtx *rpc.Context, beginTime int64, endTime int64, baseCoin string, symbol string, interval string, token *config.ReputableToken) ([][]interface{}, *gerror.AppError) {
	logger := s.Logger
	if rpcCtx != nil {
		logger = rpcCtx.Logger
	}
	// check rate limit
	if appError := s.CheckBinanceRateLimit(rpcCtx, 2); appError != nil {
		return nil, appError
	}
	// check binance symbol
	binanceSymbol := strings.ToUpper(symbol + baseCoin)
	var valid bool
	for _, item := range token.BinanceSymbol {
		if strings.EqualFold(item, binanceSymbol) {
			valid = true
			break
		}
	}
	if !valid {
		err := fmt.Errorf("unsupported token: %v", symbol)
		logger.Error("bad request", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_WrongRequestParameters")).WithCode(response.StatusCodeUnsupportedCryptocurrency)
	}
	// fetch prices
	var prices [][]interface{}
	url := fmt.Sprintf("https://api.binance.com/api/v3/uiKlines?symbol=%s&interval=%s&startTime=%d&endTime=%d&limit=1000", binanceSymbol, interval, beginTime, endTime)
	resp, err := simplehttp.Get(url, nil, &prices)
	if err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	if resp.StatusCode != http.StatusOK {
		err = simplehttp.ErrResponseStatusCodeNotEqualTo200
		logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	return prices, nil
}

type OkxUIKlinesResponse struct {
	Code string          `json:"code"`
	Msg  string          `json:"msg"`
	Data [][]interface{} `json:"data"`
}

func (s *MarketService) getUIKlines_Okx(rpcCtx *rpc.Context, endTime int64, baseCoin string, symbol string, interval string, token *config.ReputableToken) ([][]interface{}, *gerror.AppError) {
	logger := s.Logger
	if rpcCtx != nil {
		logger = rpcCtx.Logger
	}
	// check rate limit
	if res := s.OkxApiHistoryMarkPriceRatelimit.TakeAvailable(1); res == 0 {
		err := errors.New("request rate limit")
		logger.Error("bad request", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_TooManyRequests")).WithCode(response.StatusCodeTooManyRequests)

	}
	// check binance symbol
	instId := strings.ToUpper(symbol + "-" + baseCoin)
	var valid bool
	for _, item := range token.OkxInstIds {
		if strings.EqualFold(item, instId) {
			valid = true
			break
		}
	}
	if !valid {
		err := fmt.Errorf("unsupported token: %v", symbol)
		logger.Error("bad request", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_WrongRequestParameters")).WithCode(response.StatusCodeUnsupportedCryptocurrency)
	}
	switch interval {
	case "5m":
		interval = "5m"
	case "1h":
		interval = "1H"
	case "1d":
		interval = "1D"
	}
	// fetch prices
	var respData OkxUIKlinesResponse
	url := fmt.Sprintf("https://www.okx.com/api/v5/market/mark-price-candles?instId=%s&bar=%s&after=%d", instId, interval, endTime)
	resp, err := simplehttp.Get(url, nil, &respData)
	if err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	if resp.StatusCode != http.StatusOK {
		logger.Error("internal error", zap.NamedError("appError", simplehttp.ErrResponseStatusCodeNotEqualTo200))
		return nil, gerror.NewError(simplehttp.ErrResponseStatusCodeNotEqualTo200).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	if respData.Code != "0" {
		logger.Error("internal error", zap.NamedError("appError", simplehttp.ErrResponseDataStatusNotOK))
		return nil, gerror.NewError(simplehttp.ErrResponseDataStatusNotOK).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	return respData.Data, nil
}

func (s *MarketService) GetUIKlines(ctx context.Context, rpcCtx *rpc.Context, req *doom_api.GetUIKlinesRequest) (*doom_api.GetUIKlinesResponse_Data, *gerror.AppError) {
	logger := s.Logger
	if rpcCtx != nil {
		logger = rpcCtx.Logger
	}
	// validate request basically
	if appError := s.ValidateRequest(ctx, rpcCtx, req); appError != nil {
		return nil, appError
	}
	// check duration
	var intervalValue time.Duration
	switch req.Interval {
	case "5m":
		intervalValue = 5 * time.Minute
	case "1h":
		intervalValue = 1 * time.Hour
	case "1d":
		intervalValue = 24 * time.Hour
	}
	dataLength := time.Millisecond * time.Duration(req.EndTime-req.BeginTime) / intervalValue
	if dataLength > 1000 {
		err := errors.New("time duration too large, or interval too short")
		logger.Error("bad request", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_WrongRequestParameters")).WithCode(response.StatusCodeWrongParameters)
	}
	// split to binance and okx
	token := config.InReputableTokens(req.Symbol)
	if token == nil {
		logger.Error("bad request", zap.NamedError("appError", ErrUnsupportedToken))
		return nil, gerror.NewError(ErrUnsupportedToken).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_WrongRequestParameters")).WithCode(response.StatusCodeUnsupportedCryptocurrency)
	}
	var list []*doom_api.GetUIKlinesResponse_Price
	if len(token.BinanceSymbol) > 0 {
		// getUIKlines_Binance
		prices, appError := s.getUIKlines_Binance(rpcCtx, req.BeginTime, req.EndTime, req.BaseCoin, req.Symbol, req.Interval, token)
		if appError != nil {
			return nil, appError
		}
		for _, item := range prices {
			list = append(list, &doom_api.GetUIKlinesResponse_Price{
				OpenPrice:  item[1].(string),         // Open price
				ClosePrice: item[4].(string),         // Close price
				HighPrice:  item[2].(string),         // High price
				LowPrice:   item[3].(string),         // Low price
				OpenTime:   int64(item[0].(float64)), // Kline open time
				CloseTime:  int64(item[6].(float64)), // Kline close time
			})
		}
	} else if len(token.OkxInstIds) > 0 {
		// getUIKlines_Binance
		prices, appError := s.getUIKlines_Okx(rpcCtx, req.EndTime, req.BaseCoin, req.Symbol, req.Interval, token)
		if appError != nil {
			return nil, appError
		}
		beginTime := time.UnixMilli(req.BeginTime)
		for _, item := range prices {
			ts := item[0].(string) // 开始时间，Unix时间戳的毫秒数格式，如 1597026383085
			tsVal, _ := strconv.Atoi(ts)
			// confirm := item[5].(string)  // K线状态 0 代表 K 线未完结，1 代表 K 线已完结。
			closeTimeMilli := int64(tsVal + 3599999) // closeTime
			if time.UnixMilli(int64(tsVal)).Before(beginTime) {
				continue
			}
			list = append(list, &doom_api.GetUIKlinesResponse_Price{
				OpenPrice:  item[1].(string), // 开盘价格
				ClosePrice: item[4].(string), // 收盘价格
				HighPrice:  item[2].(string), // 最高价格
				LowPrice:   item[3].(string), // 最低价格
				OpenTime:   int64(tsVal),     // 开始时间
				CloseTime:  closeTimeMilli,
			})
		}
	}
	return &doom_api.GetUIKlinesResponse_Data{
		List: list,
	}, nil
}

type Etherscan_GetGasFeeResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Result  struct {
		LastBlock       string `json:"LastBlock"`
		SafeGasPrice    string `json:"SafeGasPrice"`
		ProposeGasPrice string `json:"ProposeGasPrice"`
		FastGasPrice    string `json:"FastGasPrice"`
		SuggestBaseFee  string `json:"suggestBaseFee"`
		GasUsedRatio    string `json:"gasUsedRatio"`
	} `json:"result"`
}

func Etherscan_GetGasFee() (*Etherscan_GetGasFeeResponse, error) {
	url := "https://api.etherscan.io/api?module=gastracker&action=gasoracle&apikey=KBIF7HCBUFREP3WA9Y4JHJ28BR7KMCKCEQ"
	var respData Etherscan_GetGasFeeResponse
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

var (
	CachePrefixMarketGasFee = "Market::GasFee"
)

func (s *MarketService) GetGasFee(ctx context.Context, rpcCtx *rpc.Context, req *doom_api.GetGasFeeRequest) (*doom_api.GetGasFeeResponse_Data, *gerror.AppError) {
	logger := s.Logger
	if rpcCtx != nil {
		logger = rpcCtx.Logger
	}
	// validate request basically
	if appError := s.ValidateRequest(ctx, rpcCtx, req); appError != nil {
		return nil, appError
	}
	_, err := config.ChainByName(req.Chain)
	if err != nil {
		logger.Error("bad request", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_WrongRequestParameters")).WithCode(response.StatusCodeBadRequest)
	}
	var respData *Etherscan_GetGasFeeResponse
	// get from local cache
	data, has := s.ExpirableCache.Get(CachePrefixMarketGasFee)
	if !has {
		respData, err = Etherscan_GetGasFee()
		if err != nil {
			logger.Error("bad request", zap.NamedError("appError", err))
			return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_WrongRequestParameters")).WithCode(response.StatusCodeBadRequest)
		}
		// set local cache
		s.ExpirableCache.Set(CachePrefixMarketGasFee, respData, 1*time.Second)
	} else {
		respData = data.(*Etherscan_GetGasFeeResponse)
	}
	return &doom_api.GetGasFeeResponse_Data{
		BaseFee:        respData.Result.SuggestBaseFee,
		SlowGasPrice:   respData.Result.SafeGasPrice,
		NormalGasPrice: respData.Result.ProposeGasPrice,
		FastGasPrice:   respData.Result.FastGasPrice,
	}, nil
}

type Etherscan_GetEstimationOfConfirmationTimeResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Result  string `json:"result"`
}

func Etherscan_GetEstimationOfConfirmationTime(gasPrice string) (*Etherscan_GetEstimationOfConfirmationTimeResponse, error) {
	url := "https://api.etherscan.io/api?module=gastracker&action=gasestimate&gasprice=" + gasPrice + "&apikey=KBIF7HCBUFREP3WA9Y4JHJ28BR7KMCKCEQ"
	var respData Etherscan_GetEstimationOfConfirmationTimeResponse
	resp, err := simplehttp.Get(url, nil, &respData)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, simplehttp.ErrResponseStatusCodeNotEqualTo200
	}
	if respData.Status != "1" {
		return nil, errors.New(respData.Message + "\n" + respData.Result)
	}
	return &respData, nil
}

var (
	CachePrefixMarketGetEstimationOfConfirmationTime = "Market::GetEstimationOfConfirmationTime"
)

func (s *MarketService) GetEstimationOfConfirmationTime(ctx context.Context, rpcCtx *rpc.Context, req *doom_api.GetEstimationOfConfirmationTimeRequest) (*doom_api.GetEstimationOfConfirmationTimeResponse_Data, *gerror.AppError) {
	logger := s.Logger
	if rpcCtx != nil {
		logger = rpcCtx.Logger
	}
	// validate request basically
	if appError := s.ValidateRequest(ctx, rpcCtx, req); appError != nil {
		return nil, appError
	}
	_, err := config.ChainByName(req.Chain)
	if err != nil {
		logger.Error("bad request", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_WrongRequestParameters")).WithCode(response.StatusCodeBadRequest)
	}
	var respData *Etherscan_GetEstimationOfConfirmationTimeResponse
	// get from local cache
	data, has := s.ExpirableCache.Get(CachePrefixMarketGetEstimationOfConfirmationTime + "::" + req.GasPrice)
	if !has {
		respData, err = Etherscan_GetEstimationOfConfirmationTime(req.GasPrice)
		if err != nil {
			logger.Error("bad request", zap.NamedError("appError", err))
			return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_WrongRequestParameters")).WithCode(response.StatusCodeBadRequest)
		}
		// set local cache
		s.ExpirableCache.Set(CachePrefixMarketGetEstimationOfConfirmationTime+"::"+req.GasPrice, respData, 2*time.Second)
	} else {
		respData = data.(*Etherscan_GetEstimationOfConfirmationTimeResponse)
	}
	return &doom_api.GetEstimationOfConfirmationTimeResponse_Data{
		EstimatedSeconds: respData.Result,
	}, nil
}
