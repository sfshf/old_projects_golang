package mongo

import "go.mongodb.org/mongo-driver/v2/bson"

const (
	CollectionName_BtcTxsMempools = "btc_txs_mempools"
	MempoolType_OK                = "OK"
	MempoolType_Error             = "ERROR"
)

type BtcTxsMempools struct {
	Key       string `bson:"key,omitempty"`
	Type      string `bson:"type,omitempty"`
	Data      string `bson:"data,omitempty"`
	UpdatedAt int64  `bson:"updatedAt,omitempty"`
}

const (
	CollectionName_DiffCoinbaseBinance = "diff_coinbase_binance"
)

type DiffCoinbaseBinance struct {
	Key          string  `bson:"key,omitempty"`
	PriceCB      float64 `bson:"priceCB,omitempty"`
	PriceBN      float64 `bson:"priceBN,omitempty"`
	PriceDiff    float64 `bson:"priceDiff,omitempty"`
	DiffPercent  float64 `bson:"diffPercent,omitempty"`
	ErrorMessage string  `bson:"errorMessage,omitempty"`
	UpdatedAt    int64   `bson:"updatedAt,omitempty"`
}

const (
	CollectionName_MessageNotificationConfig = "message_notification_config"
)

type MessageNotificationConfig struct {
	ID           bson.ObjectID `bson:"_id,omitempty"`
	UseTelegram  bool          `bson:"useTelegram"`
	UseWxpusher  bool          `bson:"useWxpusher"`
	WxpusherUIDs []string      `bson:"wxpusherUIDs"`
	UseEmail     bool          `bson:"useEmail"`
	Emails       []string      `bson:"emails"`
}

const (
	CollectionName_MessageNotificationLog = "message_notification_log"
)

const (
	Ltelegram = 1 << iota
	Lwxpusher
	Lemail
	Ltw  = Ltelegram | Lwxpusher
	Lte  = Ltelegram | Lemail
	Lwe  = Lwxpusher | Lemail
	Lall = Ltelegram | Lwxpusher | Lemail
)

type MessageNotificationLog struct {
	ID        bson.ObjectID `bson:"_id,omitempty"`
	Mode      int32         `bson:"mode"`
	Message   string        `bson:"message"`
	CreatedAt int64         `bson:"createdAt"`
}
