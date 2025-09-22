package response

const (
	// prefix 4 digits are app codes, suffix 5 digits are business codes.

	// 0 represents success actions.
	StatusCodeOK = 0 // common action

	// xxxx4xxxx represents client wrong actions.
	StatusCodeBadRequest      = 107040000 // common action
	StatusCodeWrongParameters = 107040010 // common wrong-parameters action
	StatusCodeEmptyParameters = 107040011
	StatusCodeUnauthorized    = 107040100
	StatusCodeForbidden       = 107040300

	// xxxx5xxxx represents server wrong actions.
	StatusCodeInternalServerError = 107050000 // common action
)

const (
	StatusMessageOk = "ok"

	StatusMessageBadRequest      = "bad request"
	StatusMessageEmptyParameters = "have empty parameters"
	StatusMessageWrongParameters = "have wrong parameters"

	StatusMessageInternalServerError = "internal server error"
)
