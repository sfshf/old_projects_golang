package response

const (
	// prefix 4 digits are app codes, suffix 5 digits are business codes.

	// 0 represents success actions.
	StatusCodeOK = 0 // common action

	// xxxx4xxxx represents client wrong actions.
	StatusCodeBadRequest                 = 111140000 // common action
	StatusCodeWrongParameters            = 111140010 // common wrong-parameters action
	StatusCodeEmptyParameters            = 111140011
	StatusCodeUnauthorized               = 111140100
	StatusCodeUpstreamServiceNotFound    = 111140400
	StatusCodeMethodNotAllowed           = 111140500
	StatusCodeRequestTimeout             = 111140800 // request time out
	StatusCodeGoRequestDelay             = 111142600 // go request timestamp delay
	StatusCodeGoRequestDeformedSecretKey = 111142601
	StatusCodeTooManyRequests            = 111142900 // too many requests, used by rate limit mw

	// xxxx5xxxx represents server wrong actions.
	StatusCodeInternalServerError = 111150000 // common action
)

const (
	StatusMessageOk = "ok"

	StatusMessageTooManyRequests = "too many requests"
)

func IsOracleGatewayErrorCode(code int32) bool {
	return code == StatusCodeInternalServerError
}
