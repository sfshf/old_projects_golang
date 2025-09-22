package response

const (
	// prefix 4 digits are app codes, suffix 5 digits are business codes.

	// 0 represents success actions.
	StatusCodeOK = 0 // common action

	// xxxx4xxxx represents client wrong actions.
	StatusCodeBadRequest          = 104040000 // common action
	StatusCodeWrongParameters     = 104040010 // common wrong-parameters action
	StatusCodeEmptyParameters     = 104040011
	StatusCodeApiKeyNameExists    = 104040012
	StatusCodeKeyIDExists         = 104040013
	StatusCodeKeyIDNotExists      = 104040014
	StatusCodeUnauthorized        = 104040100 // common unauthorized action
	StatusCodeInvalidApiKey       = 104040101
	StatusCodeUnmatchedPermission = 104040102

	// xxxx5xxxx represents server wrong actions.
	StatusCodeInternalServerError = 104050000 // common action
)

const (
	StatusMessageOk = "ok"

	StatusMessageBadRequest          = "bad request"
	StatusMessageWrongParameters     = "have wrong parameters"
	StatusMessageEmptyParameters     = "have empty parameters"
	StatusMessageApiKeyNameExists    = "api key name exists"
	StatusMessageKeyIDExists         = "key id exists"
	StatusMessageKeyIDNotExists      = "key id not exists"
	StatusMessageUnauthorized        = "unauthorized"
	StatusMessageInvalidApiKey       = "invalid api key"
	StatusMessageUnmatchedPermission = "unmatched permission"

	StatusMessageInternalServerError = "internal server error"
)
