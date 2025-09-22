package mongo

import (
	"go.mongodb.org/mongo-driver/v2/bson"
)

const (
	CollectionName_AppApiTestcases = "app_api_testcases"
)

type AppApiTestcases struct {
	App          string        `bson:"app,omitempty"`
	ApiTestCases []ApiTestcase `bson:"apiTestCases,omitempty"`
}

type ApiTestcase struct {
	Name string `bson:"name,omitempty"`
	Path string `bson:"path,omitempty"`
	Body string `bson:"body,omitempty"`
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
	CollectionName_RegistrationCaptcha = "registration_captcha"
)

type RegistrationCaptcha struct {
	ID            bson.ObjectID  `bson:"_id,omitempty"`
	EmailCaptchas []EmailCaptcha `bson:"emailCaptchas,omitempty"`
}

type EmailCaptcha struct {
	Email     string `bson:"email,omitempty"`
	Captcha   string `bson:"captcha,omitempty"`
	CreatedAt int64  `bson:"createdAt,omitempty"`
}

const (
	CollectionName_LoginCaptcha = "login_captcha"
)

type LoginCaptcha struct {
	ID            bson.ObjectID  `bson:"_id,omitempty"`
	EmailCaptchas []EmailCaptcha `bson:"emailCaptchas,omitempty"`
}

const (
	CollectionName_UploadApp = "upload_app"
)

type UploadApp struct {
	ID          bson.ObjectID `bson:"_id,omitempty"`
	AppName     string        `bson:"appName"`
	AppVersions []AppVersion  `bson:"appVersions,omitempty"`
}

type AppVersion struct {
	Version  int    `bson:"version"`
	Download string `bson:"download"`
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
