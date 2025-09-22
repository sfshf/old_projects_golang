package services

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/polly"
	pollyTypes "github.com/aws/aws-sdk-go-v2/service/polly/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go"
	"github.com/aws/smithy-go/logging"
	gerror "github.com/nextsurfer/ground/pkg/err"
	"github.com/nextsurfer/ground/pkg/rpc"
	word_api "github.com/nextsurfer/word/api"
	"github.com/nextsurfer/word/api/response"
	"github.com/nextsurfer/word/pkg/consts"
	"go.uber.org/zap"
)

const BUCKET_NAME = "n1xt-audio"

// Maybe in the future, I will change the service provider.
const POLLY_BASE_PATH = "polly"

// WordService : text to audio
type TextToAudioService struct {
	logger      *zap.Logger
	s3Client    *s3.Client
	pollyClient *polly.Client
}

// NewTextToAudio
func NewTextToAudioService(logger *zap.Logger) *TextToAudioService {
	service := &TextToAudioService{
		logger: logger,
	}
	// aws
	cfg, awsErr := config.LoadDefaultConfig(context.TODO(), config.WithLogger(service))
	if awsErr != nil {
		logger.Panic("aws sdk LoadDefaultConfig failed: ", zap.NamedError("appError", awsErr))
	}
	// Create an Amazon S3 service client
	service.s3Client = s3.NewFromConfig(cfg)
	// polly service
	service.pollyClient = polly.NewFromConfig(cfg)
	return service
}

type audioOption struct {
	accent            consts.AudioAccent
	voice             consts.AudioVoice
	pollyVoiceID      pollyTypes.VoiceId
	pollyLanguageCode pollyTypes.LanguageCode
}

func parseOption(accent consts.AudioAccent, voice consts.AudioVoice) *audioOption {
	var voiceID pollyTypes.VoiceId
	var pollyLanguageCode pollyTypes.LanguageCode
	switch accent {
	case consts.UK:
		pollyLanguageCode = pollyTypes.LanguageCodeEnGb
		switch voice {
		case consts.UK_Amy:
			voiceID = pollyTypes.VoiceIdAmy
		case consts.UK_Brian:
			voiceID = pollyTypes.VoiceIdBrian
		default:
			return nil
		}
	case consts.US:
		pollyLanguageCode = pollyTypes.LanguageCodeEnUs
		switch voice {
		case consts.US_Joanna:
			voiceID = pollyTypes.VoiceIdJoanna
		case consts.US_Matthew:
			voiceID = pollyTypes.VoiceIdMatthew
		default:
			return nil
		}
	default:
		return nil
	}
	return &audioOption{
		accent:            accent,
		voice:             voice,
		pollyVoiceID:      voiceID,
		pollyLanguageCode: pollyLanguageCode,
	}
}

// trim sentence, remove blank or useless punctuation.
func trimSentence(s string) string {
	// TODO
	return strings.TrimSpace(s)
}

// get a hash of sentence. Use this hash as ID
func hashSentence(s string) string {
	data := []byte(s)
	return fmt.Sprintf("%x", sha256.Sum256(data))[:32]
}

// get path. Path = prefix + hash(ssml). prefix = options
func getPathOf(ssml string, option *audioOption) string {
	return strings.ToLower(fmt.Sprintf("%s/%s/%s.mp3", POLLY_BASE_PATH, string(option.voice), hashSentence(ssml)))
}

// url = cnd + path
func getAudioURL(path string) string {
	// https://d1efkjaw1zw0ez.cloudfront.net/f96c252a-68d9-4cda-819f-57851052a311.mp3
	return "https://d1efkjaw1zw0ez.cloudfront.net/" + path
}

// aws logger
func (s *TextToAudioService) Logf(classification logging.Classification, format string, v ...interface{}) {
	log := fmt.Sprintf(format, v...)
	if classification == logging.Warn {
		s.logger.Warn("AWS Warning : ", zap.String("log", log))
	} else if classification == logging.Debug {
		s.logger.Debug("AWS Debug : ", zap.String("log", log))
	}
}

// TODO https://docs.aws.amazon.com/polly/latest/dg/supportedtags.html special pronunciation.
func (s *TextToAudioService) GetAudio(ctx context.Context, rpcCtx *rpc.Context, apiKey, text, ssml, accent, voice string) (*word_api.TextToAudioResponse_Data, *gerror.AppError) {
	// 1. parse option
	var apiErr smithy.APIError
	audioOption := parseOption(consts.AudioAccent(accent), consts.AudioVoice(voice))
	if audioOption == nil {
		err := fmt.Errorf("GetAudio parameter wrong. accent: %s voice: %s", accent, voice)
		s.logger.Error("bad request", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("LoginErrMsg_WrongRequestParameters")).WithCode(response.StatusCodeWrongParameters)
	}
	if len(text) > 0 {
		ssml = "<speak>" + trimSentence(text) + "</speak>"
	}
	path := getPathOf(ssml, audioOption)
	// 2. check s3 or mysql
	input := &s3.HeadObjectInput{
		Bucket: aws.String(BUCKET_NAME),
		Key:    aws.String(path),
	}
	if _, err := s.s3Client.HeadObject(ctx, input); err != nil {
		if errors.As(err, &apiErr) {
			if apiErr.ErrorCode() != "NotFound" {
				s.logger.Error("internal error: ", zap.String("headError", apiErr.ErrorCode()), zap.String("error", apiErr.Error()))
				return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
			}
		} else {
			s.logger.Error("aws.s3.HeadObject error : ", zap.NamedError("appError", err))
			return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
		}
	} else {
		return &word_api.TextToAudioResponse_Data{
			AudioURL: getAudioURL(path),
		}, nil
	}
	// 3. call polly TODO no speed
	if err := s.callPolly(ctx, rpcCtx, audioOption, ssml, apiErr, path); err != nil {
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	// 5. compose response
	return &word_api.TextToAudioResponse_Data{
		AudioURL: getAudioURL(path),
	}, nil
}

func (s *TextToAudioService) callPolly(ctx context.Context, rpcCtx *rpc.Context, audioOption *audioOption, ssml string, apiErr smithy.APIError, path string) error {
	pollyResult, err := s.pollyClient.SynthesizeSpeech(
		ctx,
		&polly.SynthesizeSpeechInput{
			OutputFormat: pollyTypes.OutputFormatMp3,
			SampleRate:   aws.String("24000"),
			VoiceId:      audioOption.pollyVoiceID,
			Engine:       pollyTypes.EngineNeural,
			LanguageCode: audioOption.pollyLanguageCode,
			Text:         aws.String(ssml),
			TextType:     pollyTypes.TextTypeSsml,
		})
	if err != nil {
		s.logger.Error("pollyError.(awserr.Error);  failed", zap.String("Error", err.Error()), zap.String("ErrorCode", apiErr.ErrorCode()))
		return err
	}
	// 3.2 handle data
	data, err := ioutil.ReadAll(pollyResult.AudioStream)
	if err != nil {
		s.logger.Error("GetAudio ioutil.ReadAll failed", zap.String("Error", err.Error()))
		return err
	}
	pollyResult.AudioStream.Close()
	// 4. store to s3
	if _, err := s.s3Client.PutObject(
		ctx,
		&s3.PutObjectInput{
			Bucket: aws.String(BUCKET_NAME),
			Key:    aws.String(path),
			Body:   bytes.NewReader(data),
		}); err != nil {
		s.logger.Error("PutObject;  failed", zap.String("Error", err.Error()), zap.String("ErrorCode", apiErr.ErrorCode()))
		return err
	}
	return nil
}
