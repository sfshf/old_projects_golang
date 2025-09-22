package consts

const ErrorCodeBase = 00200000

const ErrorCodeUnexpected = ErrorCodeBase + 10001
const ErrorCodeTimeout = ErrorCodeBase + 10002

// audio service
const ErrorCodeParameterWrong = ErrorCodeBase + 30003

// AWS
const ErrorCodeAWSInitFailed = ErrorCodeBase + 20001
const ErrorCodePollyError = ErrorCodeBase + 20002
const ErrorCodeS3UploadError = ErrorCodeBase + 20003
