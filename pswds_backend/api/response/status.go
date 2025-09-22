package response

const (
	// prefix 4 digits are app codes, suffix 5 digits are business codes.

	// 0 represents success actions.
	StatusCodeOK = 0 // common action

	// xxxx4xxxx represents client wrong actions.
	StatusCodeBadRequest              = 108040000 // common action
	StatusCodeWrongParameters         = 108040010 // common wrong-parameters action
	StatusCodeEmptyParameters         = 108040011
	StatusCodeNoBackup                = 108040012
	StatusCodeDataPullAhead           = 108040013
	StatusCodeDataFallBehind          = 108040014
	StatusCodeNotSetSecurityQuestions = 108040015
	StatusCodeResourceExists          = 108040016
	StatusCodeRegisterEmailNotExists  = 108040020
	StatusCodeUnauthorized            = 108040100
	StatusCodeForbidden               = 108040300
	StatusCodeNotFound                = 108040400
	StatusCodeResourceLimit           = 108040050 // common resource-limit action

	// xxxx5xxxx represents server wrong actions.
	StatusCodeInternalServerError = 108050000 // common action
)

const (
	StatusMessageOk = "ok"

	StatusMessageBadRequest      = "bad request"
	StatusMessageEmptyParameters = "have empty parameters"
	StatusMessageWrongParameters = "have wrong parameters"

	StatusMessageInternalServerError = "internal server error"
)
