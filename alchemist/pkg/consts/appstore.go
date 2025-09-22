package consts

// designated environment values
const (
	DESIGNATED_ENVIRONMENT_EMPTY   = ""
	DESIGNATED_ENVIRONMENT_SANDBOX = "sandbox"
	DESIGNATED_ENVIRONMENT_PROD    = "production"
)

const (
	DESIGNATED_ENVIRONMENT_NUM_EMPTY   = 0
	DESIGNATED_ENVIRONMENT_NUM_SANDBOX = 1
	DESIGNATED_ENVIRONMENT_NUM_PROD    = 2
)

func Environment(et int32) string {
	switch et {
	case DESIGNATED_ENVIRONMENT_NUM_EMPTY:
		return DESIGNATED_ENVIRONMENT_EMPTY
	case DESIGNATED_ENVIRONMENT_NUM_SANDBOX:
		return DESIGNATED_ENVIRONMENT_SANDBOX
	case DESIGNATED_ENVIRONMENT_NUM_PROD:
		return DESIGNATED_ENVIRONMENT_PROD
	default:
		return "unknown"
	}
}

func EnvironmentNum(es string) int32 {
	switch es {
	case DESIGNATED_ENVIRONMENT_EMPTY, "0":
		return DESIGNATED_ENVIRONMENT_NUM_EMPTY
	case DESIGNATED_ENVIRONMENT_SANDBOX, "1":
		return DESIGNATED_ENVIRONMENT_NUM_SANDBOX
	case DESIGNATED_ENVIRONMENT_PROD, "2":
		return DESIGNATED_ENVIRONMENT_NUM_PROD
	default:
		return -1
	}
}

// notificationType --> https://www.notion.so/7ebbd95ef97b4ff1a40077340adc46f6?pvs=4#8293af38cc1d43d088e6ffbf0b60a66c
const (
	DID_CHANGE_RENEWAL_STATUS = "DID_CHANGE_RENEWAL_STATUS"
	DID_FAIL_TO_RENEW         = "DID_FAIL_TO_RENEW"
	DID_RENEW                 = "DID_RENEW" // 更新 subscription_state
	EXPIRED                   = "EXPIRED"   // 更新 subscription_state
	GRACE_PERIOD_EXPIRED      = "GRACE_PERIOD_EXPIRED"
	OFFER_REDEEMED            = "OFFER_REDEEMED"
	REFUND                    = "REFUND"
	REFUND_DECLINED           = "REFUND_DECLINED"
	REFUND_REVERSED           = "REFUND_REVERSED"
	SUBSCRIBED                = "SUBSCRIBED" // 更新 subscription_state
	TEST                      = "TEST"       // 测试用， 不做处理

	CONSUMPTION_REQUEST             = "CONSUMPTION_REQUEST"
	DID_CHANGE_RENEWAL_PREF         = "DID_CHANGE_RENEWAL_PREF"
	NOTIFICATIONTYPE_PRICE_INCREASE = "PRICE_INCREASE"
	RENEWAL_EXTENDED                = "RENEWAL_EXTENDED"
	RENEWAL_EXTENSION               = "RENEWAL_EXTENSION"
	REVOKE                          = "REVOKE"
)

// subtype --> https://developer.apple.com/documentation/appstoreservernotifications/subtype
const (
	ACCEPTED               = "ACCEPTED"
	AUTO_RENEW_DISABLED    = "AUTO_RENEW_DISABLED"
	AUTO_RENEW_ENABLED     = "AUTO_RENEW_ENABLED"
	BILLING_RECOVERY       = "BILLING_RECOVERY"
	BILLING_RETRY          = "BILLING_RETRY"
	DOWNGRADE              = "DOWNGRADE"
	FAILURE                = "FAILURE"
	GRACE_PERIOD           = "GRACE_PERIOD"
	INITIAL_BUY            = "INITIAL_BUY"
	PENDING                = "PENDING"
	SUBTYPE_PRICE_INCREASE = "PRICE_INCREASE"
	PRODUCT_NOT_FOR_SALE   = "PRODUCT_NOT_FOR_SALE"
	RESUBSCRIBE            = "RESUBSCRIBE"
	SUMMARY                = "SUMMARY"
	UPGRADE                = "UPGRADE"
	VOLUNTARY              = "VOLUNTARY"
)

// offerType
const (
	OfferTypeIntroductory = 1
	OfferTypePromotional  = 2
	OfferTypeOfferCode    = 3

	OfferTypeDescriptionIntroductory = "introductory"
	OfferTypeDescriptionPromotional  = "promotional"
	OfferTypeDescriptionOfferCode    = "offerCode"
)

func OfferType(ot int32) string {
	switch ot {
	case OfferTypeIntroductory:
		return OfferTypeDescriptionIntroductory
	case OfferTypePromotional:
		return OfferTypeDescriptionPromotional
	case OfferTypeOfferCode:
		return OfferTypeDescriptionOfferCode
	default:
		return "unknown"
	}
}
