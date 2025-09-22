package response

const (
	// prefix 4 digits are app codes, suffix 5 digits are business codes.

	// 0 represents success actions.
	StatusCodeOK = 0 // common action

	// xxxx4xxxx represents client wrong actions.
	StatusCodeBadRequest                = 102040000 // common action
	StatusCodeWrongParameters           = 102040010 // common wrong-parameters action
	StatusCodeEmptyParameters           = 102040011
	StatusCodeUnauthorized              = 102040100 // common unauthorized action
	StatusCodeUserRegisteredOnOldDevice = 102040101
	StatusCodeUserHasNotRegistered      = 102040102
	StatusCodeUserHasRegisteredExceed   = 102040103

	// xxxx5xxxx represents server wrong actions.
	StatusCodeInternalServerError = 102050000 // common action
)

const (
	StatusMessageOk = "ok"

	StatusMessageBadRequest                = "bad request"
	StatusMessageEmptyParameters           = "have empty parameters"
	StatusMessageWrongParameters           = "have wrong parameters"
	StatusMessageUnauthorized              = "unauthorized"
	StatusMessageUserRegisteredOnOldDevice = "user registered on old device"
	StatusMessageUserHasNotRegistered      = "user has not registered"
	StatusMessageUserHasRegisteredExceed   = "user's registration duration exceeds"

	StatusMessageInternalServerError = "internal server error"
)
