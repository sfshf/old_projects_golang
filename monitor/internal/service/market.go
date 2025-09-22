package service

import (
	"context"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

type MarketService struct {
	*MonitorService

	BinanceWsConn       *websocket.Conn
	BinanceBTCPrices    *sync.Map // milli_ts->price
	CoinbaseWsConn      *websocket.Conn
	CoinbaseBTCPrices   *sync.Map // milli_ts->price
	CoinbaseLatestPrice string
}

func NewMarketService(ctx context.Context, monitorService *MonitorService) (*MarketService, error) {
	s := &MarketService{
		MonitorService:    monitorService,
		BinanceBTCPrices:  &sync.Map{},
		CoinbaseBTCPrices: &sync.Map{},
	}
	// subscribe to binance price channel
	binanceWsConn, err := s.SubscribeBinanceBtcPrice(ctx)
	if err != nil {
		return nil, err
	}
	s.BinanceWsConn = binanceWsConn
	// subscribe to coinbase price channel
	coinbaseWsConn, err := s.SubscribeCoinbaseBtcPrice(ctx)
	if err != nil {
		return nil, err
	}
	s.CoinbaseWsConn = coinbaseWsConn
	return s, nil
}

// binance wss -----------------------------------------------------------------------------------------

type BinanceRawPayload struct {
	EventType string `json:"e,omitempty"`
	EventTime int64  `json:"E,omitempty"`
	Symbol    string `json:"s,omitempty"`
	Price     string `json:"c,omitempty"`
	CloseTime int64  `json:"C,omitempty"`
}

func (s *MarketService) SubscribeBinanceBtcPrice(ctx context.Context) (*websocket.Conn, error) {
	url := "wss://data-stream.binance.vision/ws/btcusdt@ticker" // https://binance-docs.github.io/apidocs/spot/en/#individual-symbol-ticker-streams
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		s.Logger.Error("dial Binance websocket connection fail", zap.NamedError("appError", err))
		return nil, err
	}
	conn.SetPingHandler(nil) // important !!!
	cancelCtx, cancel := context.WithCancel(context.Background())
	// read goroutine
	go func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				s.Logger.Info("Binance websocket read goroutine has been canceled")
				return
			default:
				var rawPayload BinanceRawPayload
				if err := conn.ReadJSON(&rawPayload); err != nil {
					s.Logger.Error("Binance wss channel read error", zap.NamedError("appError", err))
					continue
				}
				// log.Printf("=================>binance price: %v\n", rawPayload)
				s.BinanceBTCPrices.Store(rawPayload.EventTime, rawPayload.Price)
			}
		}
	}(cancelCtx)
	// refresh ws connection goroutine
	go func() {
		defer conn.Close()
		defer cancel()               // cancel read goroutine
		<-time.After(time.Hour * 23) // due to 'A single connection to stream.binance.com is only valid for 24 hours'
		// refresh ws connection
		newConn, err := s.SubscribeBinanceBtcPrice(ctx)
		if err != nil {
			s.Logger.Error("SubscribeBinanceBtcPrice error", zap.NamedError("appError", err))
			return
		}
		s.BinanceWsConn = newConn
	}()
	return conn, nil
}

// coinbase wss -----------------------------------------------------------------------------------------

type CoinbaseWritePayload struct {
	Type       string   `json:"type"`
	ProductIDs []string `json:"product_ids"`
	Channel    string   `json:"channel"`
}

type CoinbaseReadPayload struct {
	Channel   string    `json:"channel,omitempty"`
	Timestamp time.Time `json:"timestamp,omitempty"`
	Events    []struct {
		Type    string `json:"type,omitempty"`
		Tickers []struct {
			Type      string `json:"type,omitempty"`
			ProductID string `json:"product_id,omitempty"`
			Price     string `json:"price,omitempty"`
		} `json:"tickers,omitempty"`
		CurrentTime      string      `json:"current_time,omitempty"`
		HeartbeatCounter interface{} `json:"heartbeat_counter,omitempty"`
	} `json:"events,omitempty"`
}

func (s *MarketService) SubscribeCoinbaseBtcPrice(ctx context.Context) (*websocket.Conn, error) {
	url := "wss://advanced-trade-ws.coinbase.com"
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		s.Logger.Error("dial Coinbase websocket connection fail", zap.NamedError("appError", err))
		return nil, err
	}
	cancelCtx, cancel := context.WithCancel(ctx)
	// 1. subscribe to ticker channel
	if err := conn.WriteJSON(&CoinbaseWritePayload{
		Type:       "subscribe",
		Channel:    "ticker",
		ProductIDs: []string{"BTC-USDT"},
	}); err != nil {
		s.Logger.Error("Coinbase subscribe to ticker channel fail", zap.NamedError("appError", err))
		cancel()
		conn.Close()
		return nil, err
	}
	// 1-1. subscribe to heartbeat channel
	if err := conn.WriteJSON(&CoinbaseWritePayload{
		Type:    "subscribe",
		Channel: "heartbeats",
	}); err != nil {
		s.Logger.Error("Coinbase subscribe to heartbeats channel fail", zap.NamedError("appError", err))
		cancel()
		conn.Close()
		return nil, err
	}
	// 2. read goroutine
	go func(ctx context.Context) {
		var heartbeat int64 = -1
		for {
			select {
			case <-ctx.Done():
				s.Logger.Info("Coinbase websocket read goroutine has been canceled")
				return
			default:
				var readPayload CoinbaseReadPayload
				if err := conn.ReadJSON(&readPayload); err != nil {
					if err == websocket.ErrCloseSent ||
						err == io.ErrUnexpectedEOF ||
						strings.Contains(err.Error(), "abnormal closure") ||
						strings.Contains(err.Error(), "unexpected EOF") {
						s.Logger.Error("Coinbase websocket connection has been closed (abnormal closure): unexpected EOF")
						// refresh ws connection goroutine
						go func() {
							conn.Close()
							cancel() // cancel read goroutine
							// refresh ws connection
							newConn, err := s.SubscribeCoinbaseBtcPrice(ctx)
							if err != nil {
								s.Logger.Error("SubscribeCoinbaseBtcPrice error", zap.NamedError("appError", err))
								return
							}
							s.CoinbaseWsConn = newConn
						}()
						return
					}
					s.Logger.Error("Coinbase wss channel read error", zap.NamedError("appError", err))
					continue
				}
				if readPayload.Channel == "ticker" && len(readPayload.Events) > 0 && len(readPayload.Events[0].Tickers) > 0 {
					s.CoinbaseBTCPrices.Store(readPayload.Timestamp.UnixMilli(), readPayload.Events[0].Tickers[0].Price)
					s.CoinbaseLatestPrice = readPayload.Events[0].Tickers[0].Price
				}
				if readPayload.Channel == "heartbeats" {
					heartbeat++
					if heartbeat%3600 == 0 {
						s.Logger.Info("Coinbase heartbeat received")
					}
				}
			}
		}
	}(cancelCtx)
	// refresh ws connection goroutine
	go func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				s.Logger.Info("Coinbase websocket write goroutine has been canceled")
				return
			case <-time.After(time.Hour * 23):
				defer conn.Close()
				defer cancel() // cancel read goroutine
				// unsubscribe
				if err := conn.WriteJSON(&CoinbaseWritePayload{
					Type:       "unsubscribe",
					Channel:    "ticker",
					ProductIDs: []string{"BTC-USDT"},
				}); err != nil {
					s.Logger.Error("Coinbase unsubscribe to channel fail", zap.NamedError("appError", err))
				}
				s.Logger.Info("Coinbase unsubscribe to channel success")
				// refresh ws connection
				newConn, err := s.SubscribeCoinbaseBtcPrice(ctx)
				if err != nil {
					s.Logger.Error("SubscribeCoinbaseBtcPrice error", zap.NamedError("appError", err))
					return
				}
				s.CoinbaseWsConn = newConn
			}
		}
	}(cancelCtx)
	return conn, nil
}
