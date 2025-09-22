package response

const (
	// prefix 4 digits are app codes, suffix 5 digits are business codes.

	// 0 represents success actions.
	StatusCodeOK = 0 // common action

	// xxxx4xxxx represents client wrong actions.
	StatusCodeBadRequest                = 105040000 // common action
	StatusCodeWrongParameters           = 105040010 // common wrong-parameters action
	StatusCodeEmptyParameters           = 105040011
	StatusCodeUnsupportedCryptocurrency = 105040012
	StatusCodeUnknownSpotPrice          = 105040013
	StatusCodeTooManyRequests           = 105042900

	// xxxx5xxxx represents server wrong actions.
	StatusCodeInternalServerError = 105050000 // common action
)

const (
	StatusMessageOk = "ok"

	StatusMessageBadRequest      = "bad request"
	StatusMessageEmptyParameters = "have empty parameters"
	StatusMessageWrongParameters = "have wrong parameters"

	StatusMessageInternalServerError = "internal server error"
)
