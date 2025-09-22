package response

const (
	// prefix 4 digits are app codes, suffix 5 digits are business codes.

	// 0 represents success actions.
	StatusCodeOK = 0 // common action

	// xxxx4xxxx represents client wrong actions.
	StatusCodeBadRequest                = 106040000 // common action
	StatusCodeWrongParameters           = 106040010 // common wrong-parameters action
	StatusCodeEmptyParameters           = 106040011
	StatusCodeUnsupportedCryptocurrency = 106040012
	StatusCodeNotFound                  = 106040400
	StatusCodeUnauthorized              = 106040100

	// xxxx5xxxx represents server wrong actions.
	StatusCodeInternalServerError = 106050000 // common action
)

const (
	StatusMessageOk = "ok"

	StatusMessageBadRequest      = "bad request"
	StatusMessageEmptyParameters = "have empty parameters"
	StatusMessageWrongParameters = "have wrong parameters"

	StatusMessageInternalServerError = "internal server error"
)
