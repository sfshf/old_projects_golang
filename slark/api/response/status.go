package response

const (
	// prefix 4 digits are app codes, suffix 5 digits are business codes.

	// 0 represents success actions.
	StatusCodeOK = 0 // common action

	// xxxx4xxxx represents client wrong actions.
	StatusCodeBadRequest                = 100040000 // common action
	StatusCodeWrongParameters           = 100040010 // common wrong-parameters action
	StatusCodeEmptyParameters           = 100040011
	StatusCodeInvalidUsernameOrPassword = 100040012
	StatusCodeInvalidUsername           = 100040013
	StatusCodeInvalidPassword           = 100040014
	StatusCodeRepeatedNickname          = 100040015
	StatusCodeDeformedEmail             = 100040016
	StatusCodeCaptchaExpired            = 100040017
	StatusCodeCaptchaWrong              = 100040018
	StatusCodeRegisterEmailExists       = 100040019
	StatusCodeRegisterEmailNotExists    = 100040020
	StatusCodeUnauthorized              = 100040100 // common unauthorized action
	StatusCodeLoginNotRegistered        = 100040101
	StatusCodeEmptySession              = 100040102
	StatusCodeLoginSessionExpired       = 100040103
	StatusCodeSecondaryPasswordExists   = 100040104

	// xxxx5xxxx represents server wrong actions.
	StatusCodeInternalServerError = 100050000 // common action
)

const (
	StatusMessageOk = "ok"

	StatusMessageBadRequest                = "bad request"
	StatusMessageEmptyParameters           = "have empty parameters"
	StatusMessageWrongParameters           = "have wrong parameters"
	StatusMessageInvalidUsernameOrPassword = "invalid username or password"
	StatusMessageInvalidUsername           = "invalid username"
	StatusMessageInvalidPassword           = "invalid password"
	StatusMessageRepeatedNickname          = "repeated nickname"
	StatusMessageDeformedEmail             = "deformed email"
	StatusMessageCaptchaExpired            = "captcha expired"
	StatusMessageCaptchaWrong              = "captcha wrong"
	StatusMessageRegisterEmailExists       = "register email exists"
	StatusMessageRegisterEmailNotExists    = "email not registered"
	StatusMessageUnauthorized              = "unauthorized"
	StatusMessageLoginNotRegistered        = "login not registered"
	StatusMessageEmptySession              = "empty session"
	StatusMessageLoginSessionExpired       = "login session expired"

	StatusMessageInternalServerError = "internal server error"
)
