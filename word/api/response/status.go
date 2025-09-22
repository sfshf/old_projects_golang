package response

const (
	// prefix 4 digits are app codes, suffix 5 digits are business codes.

	// 0 represents success actions.
	StatusCodeOK = 0 // common action

	// xxxx4xxxx represents client wrong actions.
	StatusCodeBadRequest      = 101040000 // common action
	StatusCodeWrongParameters = 101040010 // common wrong-parameters action
	StatusCodeEmptyParameters = 101040011
	StatusCodeUnauthorized    = 101040100 // common unauthorized action

	// xxxx5xxxx represents server wrong actions.
	StatusCodeInternalServerError = 101050000 // common action
)

const (
	StatusMessageOk = "ok"

	StatusMessageBadRequest      = "bad request"
	StatusMessageEmptyParameters = "have empty parameters"
	StatusMessageWrongParameters = "have wrong parameters"
	StatusMessageUnauthorized    = "unauthorized"

	StatusMessageInternalServerError = "internal server error"
)
