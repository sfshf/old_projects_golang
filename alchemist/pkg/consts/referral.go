package consts

import "github.com/nextsurfer/ground/pkg/localize"

const (
	ReferralPointNewUser    = 3
	ReferralPointNonNewUser = 1
)

const (
	ReferralLogTypeGain    = 1
	ReferralLogTypeConsume = 2
)

const (
	ReferralLogReasonNewUser          = 1 // "referral new user"
	ReferralLogReasonNewUserFirstTime = 2 //  "referral new user first time"
	ReferralLogReasonNewUserBilled    = 3 // "referral new user billed"
	ReferralLogReasonFreeTrial        = 4 // "free trial"
	ReferralLogReasonRedeemReward     = 5 // "redeem reward"
	ReferralLogReasonConvertPoints    = 6 // "convert points"
)

func ReferralLogReason(localizer *localize.Localizer, reason int32) string {
	switch reason {
	case ReferralLogReasonNewUser:
		return localizer.Localize("ReferralLogReason_NewUser")
	case ReferralLogReasonNewUserFirstTime:
		return localizer.Localize("ReferralLogReason_NewUserFirstTime")
	case ReferralLogReasonNewUserBilled:
		return localizer.Localize("ReferralLogReason_NewUserBilled")
	case ReferralLogReasonFreeTrial:
		return localizer.Localize("ReferralLogReason_FreeTrial")
	case ReferralLogReasonRedeemReward:
		return localizer.Localize("ReferralLogReason_RedeemReward")
	case ReferralLogReasonConvertPoints:
		return localizer.Localize("ReferralLogReason_ConvertPoints")
	default:
		return ""
	}
}
