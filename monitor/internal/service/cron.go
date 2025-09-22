package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/nextsurfer/monitor/internal/common/simplehttp"
	monitor_mongo "github.com/nextsurfer/monitor/internal/mongo"
	"github.com/robfig/cron/v3"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.uber.org/zap"
)

type CronService struct {
	*MonitorService

	txIDs              []string
	cron               *cron.Cron
	cronEntries        []*CronEntryStatus
	lastHourDiffStates map[DiffState]time.Time
}

func NewCronService(ctx context.Context, MonitorService *MonitorService) (*CronService, error) {
	s := &CronService{
		MonitorService:     MonitorService,
		lastHourDiffStates: make(map[DiffState]time.Time),
	}
	// cron jobs
	crontab := cron.New(cron.WithSeconds())
	s.cron = crontab
	var cronEntries []*CronEntryStatus
	// cron jobs
	diffBetweenCoinbaseAndBinanceEntry, err := s.DiffBetweenCoinbaseAndBinanceEntry(ctx, crontab)
	if err != nil {
		return nil, err
	}
	txsMempoolEntry, err := s.USBtcTxsMempoolEntry(ctx, crontab)
	if err != nil {
		return nil, err
	}
	cronEntries = append(cronEntries, diffBetweenCoinbaseAndBinanceEntry, txsMempoolEntry)
	s.cronEntries = cronEntries
	// start cron jobs
	s.cron.Start()
	return s, nil
}

type CronEntryStatus struct {
	Name              string
	EntryID           cron.EntryID
	Started           bool
	StartedOrStopedAt time.Time
	ScheduleSpec      string
	LastExecError     string
}

// cron jobs -----------------------------------------------------------------------------------------

func (s *CronService) DiffBetweenCoinbaseAndBinanceEntry(ctx context.Context, crontab *cron.Cron) (*CronEntryStatus, error) {
	spec := "@every 5m"
	if s := os.Getenv("DIFF_COINBASE_BINANCE_CRON"); s != "" {
		spec = s
	}
	entryStatus := &CronEntryStatus{
		Name:              "DiffBetweenCoinbaseAndBinanceEntry",
		Started:           true,
		StartedOrStopedAt: time.Now(),
		ScheduleSpec:      spec,
	}
	entryID, err := crontab.AddFunc(entryStatus.ScheduleSpec, func() {
		s.Logger.Info("Cron Job -- DiffBetweenCoinbaseAndBinanceEntry")
		if err := s.handleDiffBetweenCoinbaseAndBinanceEntry(ctx); err != nil {
			entryStatus.LastExecError = err.Error()
			s.Logger.Error("cron error", zap.NamedError("appError", err))
		}
	})
	if err != nil {
		return nil, err
	}
	entryStatus.EntryID = entryID
	return entryStatus, nil
}

func (s *CronService) handleDiffBetweenCoinbaseAndBinanceEntry(ctx context.Context) error {
	now := time.Now()
	recordKey := "BTC/USDT"
	record := monitor_mongo.DiffCoinbaseBinance{
		Key:       recordKey,
		UpdatedAt: now.UnixMilli(),
	}
	ts1 := now.Add(-10 * time.Minute).UnixMilli()
	ts2 := now.Add(-5 * time.Minute).UnixMilli()
	// 1. coinbase BTC-USDT
	priceCB, err := s.Coinbase_BTCUSDT(ts1, ts2)
	if err != nil {
		return err
	}
	record.PriceCB = priceCB
	// 2. binance BTCUSDT
	priceBN, err := s.Binance_BTCUSDT(ts1, ts2)
	if err != nil {
		return err
	}
	record.PriceBN = priceBN
	// 3. compute
	priceDiff := priceCB - priceBN
	diffPercent := priceDiff * 100 / priceBN
	record.PriceDiff = priceDiff
	record.DiffPercent = diffPercent
	// 4. update record
	coll := s.MongoDB.Collection(monitor_mongo.CollectionName_DiffCoinbaseBinance)
	curDiffState := CompareDiff(diffPercent)
	ts, has := s.lastHourDiffStates[curDiffState]
	if !has || now.Sub(ts) >= 1*time.Hour {
		//post messages to telegram
		message := fmt.Sprintf("Coinbase价：%.8f；\nBinance价：%.8f；\n溢价：%.8f；\n溢价百分比：%.2f%%；\n", priceCB, priceBN, priceDiff, diffPercent)
		resp, err := simplehttp.PostJsonRequest(
			`https://api.telegram.org/bot1678806156:AAE8cWdlygrGCHWmHElQHNJ0ZjOv1IRQGeg/sendMessage`,
			map[string]string{
				"Accept":     "*/*",
				"Host":       "api.telegram.org",
				"User-Agent": `Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_6) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/14.0.2 Safari/605.1.15`,
			},
			struct {
				ChatID string `json:"chat_id"`
				Text   string `json:"text"`
			}{
				ChatID: "1417969737",
				Text:   message,
			},
			nil,
			nil,
		)
		if err != nil {
			record.ErrorMessage = fmt.Sprintf("!!! Telegram Request Error: %s", err)
		}
		if resp.StatusCode != http.StatusOK {
			record.ErrorMessage = fmt.Sprintf("!!! Telegram Request Error: %s", simplehttp.ErrResponseStatusCodeNotEqualTo200)
		}
	}
	s.lastHourDiffStates[curDiffState] = now
	// update record
	if _, err := coll.ReplaceOne(ctx, bson.D{{Key: "key", Value: recordKey}}, record, options.Replace().SetUpsert(true)); err != nil {
		return err
	}
	return nil
}

type DiffState int

const (
	DiffState_None DiffState = iota + 1
	DiffState_Negative
	DiffState_Positive
	DiffState_Substantial
)

func CompareDiff(diffPercent float64) DiffState {
	switch {
	case diffPercent < -0.02:
		return DiffState_Negative
	case diffPercent < 0.02:
		return DiffState_None
	case diffPercent < 0.08:
		return DiffState_Positive
	default:
		return DiffState_Substantial
	}
}

func (s *CronService) Binance_BTCUSDT(ts1, ts2 int64) (float64, error) {
	var total float64
	var count float64
	s.BinanceBTCPrices.Range(func(key, value any) bool {
		ts := key.(int64)
		if ts < ts1 {
			s.BinanceBTCPrices.Delete(key)
			return true
		}
		priceString := value.(string)
		if ts >= ts1 && ts < ts2 {
			price, err := strconv.ParseFloat(priceString, 64)
			if err != nil {
				s.Logger.Error("parse binance price string error", zap.NamedError("appError", err))
				return false
			}
			total += price
			count++
		}
		return true
	})
	if count > 0 {
		return total / count, nil
	}
	s.Logger.Error("no valid binance price data")
	return 0, nil
}

func (s *CronService) Coinbase_BTCUSDT(ts1, ts2 int64) (float64, error) {
	var total float64
	var count float64
	s.CoinbaseBTCPrices.Range(func(key, value any) bool {
		ts := key.(int64)
		if ts < ts1 {
			s.CoinbaseBTCPrices.Delete(key)
			return true
		}
		priceString := value.(string)
		if ts >= ts1 && ts < ts2 {
			price, err := strconv.ParseFloat(priceString, 64)
			if err != nil {
				s.Logger.Error("parse coinbase price string error", zap.NamedError("appError", err))
				return false
			}
			total += price
			count++
		}
		return true
	})
	if count > 0 {
		return total / count, nil
	} else {
		s.Logger.Warn("coinbase price did not updated for a while")
		coinbaseLatestPrice, _ := strconv.ParseFloat(s.CoinbaseLatestPrice, 64)
		if coinbaseLatestPrice > 0 {
			return coinbaseLatestPrice, nil
		}
	}
	s.Logger.Error("no valid coinbase price data")
	return 0, nil
}

func (s *CronService) USBtcTxsMempoolEntry(ctx context.Context, crontab *cron.Cron) (*CronEntryStatus, error) {
	spec := "@every 30s"
	if s := os.Getenv("TXS_MEMPOOL_CRON"); s != "" {
		spec = s
	}
	entryStatus := &CronEntryStatus{
		Name:              "TxsMempoolEntry",
		Started:           true,
		StartedOrStopedAt: time.Now(),
		ScheduleSpec:      spec,
	}
	entryID, err := crontab.AddFunc(entryStatus.ScheduleSpec, func() {
		s.Logger.Info("Cron Job -- TxsMempoolEntry")
		if err := s.handleUSBtcTxsMempoolEntry(ctx); err != nil {
			entryStatus.LastExecError = err.Error()
			s.Logger.Error("cron error", zap.NamedError("appError", err))
		}
	})
	if err != nil {
		return nil, err
	}
	entryStatus.EntryID = entryID
	return entryStatus, nil
}

func (s *CronService) handleTransaction(address string, transaction Mempool_GetAddressTransactionsMempool_Transaction) string {
	// check from/to the address
	var isFrom bool
	var fromValue int32
	for _, vin := range transaction.Vin { // from
		if vin.Prevout.ScriptpubkeyAddress == address {
			isFrom = true
		}
		fromValue += vin.Prevout.Value
	}
	var message string
	// var value int32
	if isFrom {
		// var changeValue int32
		// for _, vout := range transaction.Vout {
		// 	if vout.ScriptpubkeyAddress == address {
		// 		changeValue += vout.Value
		// 	}
		// }
		// value = fromValue - changeValue
		message = fmt.Sprintf("从账户地址<%s>转出%d聪BTC", address, fromValue-transaction.Fee)
	} else {
		// for _, vout := range transaction.Vout {
		// 	if vout.ScriptpubkeyAddress == address {
		// 		value += vout.Value
		// 	}
		// }
		message = fmt.Sprintf("向账户地址<%s>转入%d聪BTC", address, fromValue-transaction.Fee)
	}
	return message
}

func (s *CronService) handleUSBtcTxsMempoolEntry(ctx context.Context) error {
	addresses := []string{
		"bc1qngydl7hmgdtmuqjmtsyj3pcwszv0yn5mj6kz4c", // 美国政府持有的比特币
		"bc1qhse6nvlxauq8dfvas25gmhunveudphtlqe8fxz",
	}
	for _, address := range addresses {
		if err := s.handleUSBtcTxsMempool(ctx, address); err != nil {
			message := "BTC账号出账监控系统报错： 地址[" + address + "]，报错：" + err.Error() + "\n"
			resp, respError := simplehttp.PostJsonRequest(
				`https://api.telegram.org/bot1678806156:AAE8cWdlygrGCHWmHElQHNJ0ZjOv1IRQGeg/sendMessage`,
				map[string]string{
					"Accept":     "*/*",
					"Host":       "api.telegram.org",
					"User-Agent": `Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_6) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/14.0.2 Safari/605.1.15`,
				},
				struct {
					ChatID string `json:"chat_id"`
					Text   string `json:"text"`
				}{
					ChatID: "1417969737",
					Text:   message,
				},
				nil,
				nil,
			)
			if respError != nil {
				return fmt.Errorf("BTC账号出账监控系统报错：%s；发报报错：%s", err, respError)
			}
			if resp.StatusCode != http.StatusOK {
				return fmt.Errorf("BTC账号出账监控系统报错：%s；发报报错：%s", err, simplehttp.ErrResponseStatusCodeNotEqualTo200)
			}
		}
	}
	return nil
}

func (s *CronService) handleUSBtcTxsMempool(ctx context.Context, address string) error {
	record := monitor_mongo.BtcTxsMempools{
		Key:       address,
		UpdatedAt: time.Now().UnixMilli(),
	}
	// 1. mempool request
	rawData, err := Mempool_GetAddressTransactionsMempool(address)
	if err != nil {
		// type = 2: error
		record.Type = monitor_mongo.MempoolType_Error
		record.Data = err.Error()
	} else {
		// type = 1: ok
		record.Type = monitor_mongo.MempoolType_OK
		record.Data = rawData
	}
	// 3. if length of list gt 0
	var list []Mempool_GetAddressTransactionsMempool_Transaction
	if err := json.Unmarshal([]byte(rawData), &list); err != nil {
		record.Type = monitor_mongo.MempoolType_Error
		record.Data = fmt.Sprintf(record.Data+"!!! Mempool Request Error: %s", err)
	}
	if len(list) > 0 {
		for _, item := range list {
			var has bool
			for _, txid := range s.txIDs {
				if item.TxID == txid {
					has = true
					break
				}
			}
			if !has {
				s.txIDs = append(s.txIDs, item.TxID)
			} else {
				continue
			}
			// post messages to telegram
			message := `https://www.blockchain.com/explorer/transactions/btc/` + item.TxID + "\n" // transaction link
			message += s.handleTransaction(address, item) + "\n"
			resp, err := simplehttp.PostJsonRequest(
				`https://api.telegram.org/bot1678806156:AAE8cWdlygrGCHWmHElQHNJ0ZjOv1IRQGeg/sendMessage`,
				map[string]string{
					"Accept":     "*/*",
					"Host":       "api.telegram.org",
					"User-Agent": `Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_6) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/14.0.2 Safari/605.1.15`,
				},
				struct {
					ChatID string `json:"chat_id"`
					Text   string `json:"text"`
				}{
					ChatID: "1417969737",
					Text:   message,
				},
				nil,
				nil,
			)
			if err != nil {
				record.Type = monitor_mongo.MempoolType_Error
				record.Data = fmt.Sprintf(record.Data+"!!! Telegram Request Error [txid=%s]: %s", item.TxID, err)
			}
			if resp.StatusCode != http.StatusOK {
				record.Type = monitor_mongo.MempoolType_Error
				record.Data = fmt.Sprintf(record.Data+"!!! Telegram Request Error [txid=%s]: %s", item.TxID, simplehttp.ErrResponseStatusCodeNotEqualTo200)
			}
		}
	}
	// 2. insert a record
	coll := s.MongoDB.Collection(monitor_mongo.CollectionName_BtcTxsMempools)
	if _, err := coll.ReplaceOne(ctx, bson.D{{Key: "key", Value: address}}, record, options.Replace().SetUpsert(true)); err != nil {
		return err
	}
	return nil
}

type Mempool_GetAddressTransactionsMempool_Transaction struct {
	TxID     string                                       `json:"txid,omitempty"`
	Version  int32                                        `json:"version,omitempty"`
	Locktime int32                                        `json:"locktime,omitempty"`
	Vin      []Mempool_GetAddressTransactionsMempool_Vin  `json:"vin,omitempty"`
	Vout     []Mempool_GetAddressTransactionsMempool_Vout `json:"vout,omitempty"`
	Size     int32                                        `json:"size,omitempty"`
	Weight   int32                                        `json:"weight,omitempty"`
	Fee      int32                                        `json:"fee,omitempty"`
	Status   Mempool_GetAddressTransactionsMempool_Status `json:"status,omitempty"`
}

type Mempool_GetAddressTransactionsMempool_Vin struct {
	TxID         string                                            `json:"txid,omitempty"`
	Vout         int32                                             `json:"vout,omitempty"`
	Prevout      Mempool_GetAddressTransactionsMempool_Vin_Prevout `json:"prevout,omitempty"`
	Scriptsig    string                                            `json:"scriptsig,omitempty"`
	ScriptsigAsm string                                            `json:"scriptsig_asm,omitempty"`
	Witness      []string                                          `json:"witness,omitempty"`
	IsCoinbase   bool                                              `json:"is_coinbase,omitempty"`
	Sequence     int64                                             `json:"sequence,omitempty"`
}

type Mempool_GetAddressTransactionsMempool_Vin_Prevout struct {
	Scriptpubkey        string `json:"scriptpubkey,omitempty"`
	ScriptpubkeyAsm     string `json:"scriptpubkey_asm,omitempty"`
	ScriptpubkeyType    string `json:"scriptpubkey_type,omitempty"`
	ScriptpubkeyAddress string `json:"scriptpubkey_address,omitempty"`
	Value               int32  `json:"value,omitempty"`
}

type Mempool_GetAddressTransactionsMempool_Vout struct {
	Scriptpubkey        string `json:"scriptpubkey,omitempty"`
	ScriptpubkeyAsm     string `json:"scriptpubkey_asm,omitempty"`
	ScriptpubkeyType    string `json:"scriptpubkey_type,omitempty"`
	ScriptpubkeyAddress string `json:"scriptpubkey_address,omitempty"`
	Value               int32  `json:"value,omitempty"`
}

type Mempool_GetAddressTransactionsMempool_Status struct {
	Confirmed   bool   `json:"confirmed,omitempty"`
	BlockHeight int32  `json:"block_height,omitempty"`
	BlockHash   string `json:"block_hash,omitempty"`
	BlockTime   int64  `json:"block_time,omitempty"`
}

func Mempool_GetAddressTransactionsMempool(address string) (string, error) {
	url := "https://mempool.space/api/address/" + address + "/txs/mempool"
	resp, err := simplehttp.Get(url, nil, nil)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		return "", simplehttp.ErrResponseStatusCodeNotEqualTo200
	}
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(raw), nil
}
