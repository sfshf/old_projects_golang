package server

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	jsoniter "github.com/json-iterator/go"
	"github.com/juju/ratelimit"
	_ "github.com/mbobakov/grpc-consul-resolver"
	gutil "github.com/nextsurfer/ground/pkg/util"
	"github.com/nextsurfer/oracle/api/response"
	"github.com/nextsurfer/oracle/internal/common/simplehttp"
	"github.com/nextsurfer/oracle/internal/common/simplejson"
	"github.com/nextsurfer/oracle/internal/common/statistic"
	"github.com/nextsurfer/oracle/internal/dao"
	"github.com/nextsurfer/oracle/internal/gateway/service"
	. "github.com/nextsurfer/oracle/internal/model"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gorm.io/gorm"
)

// GatewayRateLimit only for gateway handlers and grpc2http handler
func (s *GatewayServer) GatewayRateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r = response.WithLocalizer(r, s.LocalizeManager) // localize manager
		ctx := r.Context()
		// first, check path rate limit
		if bucket, has := s.PathRateLimitRules[r.URL.Path]; has {
			if res := bucket.TakeAvailable(1); res == 0 {
				respBody, _ := json.Marshal(response.Response{
					Code:         response.StatusCodeTooManyRequests,
					Message:      response.LocalizeMessage(ctx, "ClientErrMsg_TooManyRequests"),
					DebugMessage: response.StatusMessageTooManyRequests,
				})
				w.Write(respBody)
				return
			}
			next.ServeHTTP(w, r)
			return
		}
		// second, check GatewayService rate limit
		if serviceCache := s.GatewayService.GrpcServiceCache(r); serviceCache != nil {
			if bucket, has := s.ServiceRateLimitRules[serviceCache.Name]; has {
				if res := bucket.TakeAvailable(1); res == 0 {
					respBody, _ := json.Marshal(response.Response{
						Code:         response.StatusCodeTooManyRequests,
						Message:      response.LocalizeMessage(ctx, "ClientErrMsg_TooManyRequests"),
						DebugMessage: response.StatusMessageTooManyRequests,
					})
					w.Write(respBody)
					return
				}
				next.ServeHTTP(w, r)
				return
			}
		}
		// third, other proxy
		next.ServeHTTP(w, r)
	})
}

func (s *GatewayServer) refreshRateLimitRules(ctx context.Context, serviceName string) error {
	GatewayService, err := s.DaoManager.ServiceDAO.GetByName(ctx, serviceName, false /*omitProtoFile*/, false /*omitDeleted*/) // fetch the GatewayService by name
	if err != nil {
		return err
	}
	if GatewayService == nil {
		return nil
	}
	paths := s.getServicePaths(GatewayService.ProtoFile)
	if GatewayService.DeletedAt > 0 { // delete rate limit
		delete(s.ServiceRateLimitRules, serviceName)
		for _, path := range paths {
			delete(s.PathRateLimitRules, path)
		}
		return nil
	}
	// first, fetch special rule named 'all'
	allRule, err := s.DaoManager.RateLimitRuleDAO.GetServiceRuleByName(ctx, "all")
	if err != nil {
		return err
	}
	if allRule != nil && allRule.Enabled {
		s.Logger.Info("refresh special rate limit rule -- all", zap.String("target", allRule.Target), zap.Int("capacity", int(allRule.Capacity)), zap.String("GatewayService", serviceName))
		// iterate all paths
		for _, path := range paths {
			s.PathRateLimitRules[path] = ratelimit.NewBucket(
				1*time.Second,
				allRule.Capacity,
			)
		}
	}
	// second, fetch GatewayService rule
	serviceRule, err := s.DaoManager.RateLimitRuleDAO.GetServiceRuleByName(ctx, serviceName)
	if err != nil {
		return err
	}
	if serviceRule != nil && serviceRule.Enabled {
		s.Logger.Info("refresh GatewayService rate limit rule", zap.String("target", serviceRule.Target), zap.Int("capacity", int(serviceRule.Capacity)), zap.String("GatewayService", serviceName))
		s.ServiceRateLimitRules[serviceRule.Target] = ratelimit.NewBucket(
			1*time.Second,
			serviceRule.Capacity,
		)
	}
	// third, iterate path rule
	for _, path := range paths {
		pathRule, err := s.DaoManager.RateLimitRuleDAO.GetPathRuleByName(ctx, path)
		if err != nil {
			return err
		}
		if pathRule != nil && pathRule.Enabled {
			s.Logger.Info("refresh path rate limit rule", zap.String("target", pathRule.Target), zap.Int("capacity", int(pathRule.Capacity)), zap.String("GatewayService", serviceName))
			s.PathRateLimitRules[pathRule.Target] = ratelimit.NewBucket(
				1*time.Second,
				pathRule.Capacity,
			)
		}
	}
	return nil
}

func (s *GatewayServer) isReverseProxy(r *http.Request) (*url.URL, bool) {
	rawURL := s.HostManage[r.Host]
	gatewayApiHostname := os.Getenv("GATEWAY_API_HOSTNAME")
	// status-backend
	if strings.HasPrefix(r.URL.Path, "/status") && (r.Host == gatewayApiHostname) {
		rawURL = "http://172.31.29.192:4010"
	}
	if rawURL == "" || r.Host == gatewayApiHostname {
		return nil, false
	}
	res, err := url.Parse(rawURL)
	if err != nil {
		return nil, false
	}
	return res, true
}

func (s *GatewayServer) ReverseProxy(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		proxyUrl, is := s.isReverseProxy(r)
		if is {
			(&httputil.ReverseProxy{Rewrite: func(r *httputil.ProxyRequest) {
				r.SetURL(proxyUrl)
				r.Out.Host = r.In.Host // if desired
				r.Out.Method = r.In.Method
			}}).ServeHTTP(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *GatewayServer) isUploadRequest(r *http.Request) bool {
	gatewayApiHostname := os.Getenv("GATEWAY_API_HOSTNAME")
	return r.Method == "POST" && r.URL.Path == "/upload" && r.Host == gatewayApiHostname
}

func (s *GatewayServer) uploadFile(ctx context.Context, r *http.Request) (string, error) {
	data, err := io.ReadAll(r.Body)
	if err != nil {
		return "", err
	}
	// hash 文件
	h := sha256.New()
	if _, err := h.Write(data); err != nil {
		return "", err
	}
	hashName := hex.EncodeToString(h.Sum(nil))
	// 先存本地缓存
	if err := os.MkdirAll("tmp/upload/", 0764); err != nil {
		return "", err
	}
	cacheFile := "tmp/upload/" + hashName
	if err := os.WriteFile(cacheFile, data, 0764); err != nil {
		return "", err
	}
	defer func() {
		// 删除缓存文件
		_ = os.Remove(cacheFile)
	}()
	keyPrefix := r.URL.Query().Get("keyPrefix")
	if keyPrefix == "" {
		keyPrefix = "upload"
	}
	fileName := r.URL.Query().Get("fileName")
	if fileName == "" {
		fileName = hashName
	} else {
		ext := filepath.Ext(fileName)
		fileName = strings.TrimSuffix(fileName, ext) + "-" + hashName + ext
	}
	// 根据 hashName 去 s3 查， 是否已经上传过
	_, err = s.S3Client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(os.Getenv("UPLOAD_BUCKET_NAME")),
		Key:    aws.String(keyPrefix + "/" + fileName),
	})
	if err == nil {
		return fileName, nil
	}
	if !strings.Contains(err.Error(), "404") && !strings.Contains(err.Error(), "NotFound") {
		return "", err
	}
	// 上传文件
	f, err := os.Open(cacheFile)
	if err != nil {
		return "", err
	}
	defer f.Close()
	_, err = s.S3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(os.Getenv("UPLOAD_BUCKET_NAME")),
		Key:    aws.String(keyPrefix + "/" + fileName),
		Body:   f,
	})
	if err != nil {
		return "", err
	}
	return fileName, nil
}

func (s *GatewayServer) UploadHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !s.isUploadRequest(r) {
			next.ServeHTTP(w, r)
			return
		}
		ctx := r.Context()
		var resp *response.Response
		// 限制上传大小为 5M 以内
		contentLength, err := strconv.Atoi(r.Header.Get("Content-Length"))
		if err != nil {
			resp = &response.Response{
				Code:         response.StatusCodeBadRequest,
				Message:      response.LocalizeMessage(ctx, "ClientErrMsg_BadRequest"),
				DebugMessage: fmt.Sprintf("bad request: %v", err),
			}
		} else {
			if contentLength > 200*1024*1024 {
				err = errors.New("http body content length too large")
				resp = &response.Response{
					Code:         response.StatusCodeBadRequest,
					Message:      response.LocalizeMessage(ctx, "ClientErrMsg_ContentLengthLimit"),
					DebugMessage: fmt.Sprintf("bad request: %v", err),
				}
			} else {
				// 先保存文件到缓存文件夹
				hashName, err := s.uploadFile(ctx, r)
				if err != nil {
					resp = &response.Response{
						Code:         response.StatusCodeInternalServerError,
						Message:      response.LocalizeMessage(ctx, "FatalErrMsg"),
						DebugMessage: fmt.Sprintf("internal error: %v", err),
					}
				} else {
					resp = &response.Response{
						Code:    response.StatusCodeOK,
						Message: response.LocalizeMessage(ctx, "ClientMsg_OK"),
						Data: struct {
							HashName string `json:"hashName"`
						}{
							HashName: hashName,
						},
					}
				}
			}
		}
		respBody, _ := json.Marshal(resp)
		w.Write(respBody)
	})
}

func (s *GatewayServer) logRequest(r *http.Request, logFields []zapcore.Field) {
	if r.Body != nil && s.Env != gutil.AppEnvPROD {
		// request body log in non prod environment
		reqBody, err := io.ReadAll(r.Body)
		if err != nil {
			s.Logger.Error("log handler error", zap.NamedError("appError", err))
		}
		if len(reqBody) > 0 {
			var reqBodyContent string
			reqBodyContent, err = simplejson.TruncateJsonTo200(reqBody)
			if err != nil {
				s.Logger.Error("log handler error", zap.NamedError("appError", err))
			}
			logFields = append(logFields, zap.String("requestBody", reqBodyContent))
			r.Body = io.NopCloser(bytes.NewBuffer(reqBody))
		}
	}
}

func (s *GatewayServer) logResponse(respBody []byte, logFields []zapcore.Field) {
	if s.Env != gutil.AppEnvPROD {
		respBodyContent, err := simplejson.TruncateJsonTo200(respBody)
		if err != nil {
			s.Logger.Error("log handler error", zap.NamedError("appError", err))
		}
		logFields = append(logFields, zap.String("responseBody", respBodyContent))
	}
}

func (s *GatewayServer) generateLogFields(r *http.Request) []zapcore.Field {
	logFields := []zapcore.Field{
		zap.String("ip", simplehttp.GetRealIP(r)),
		zap.String("ua", simplehttp.UserAgent(r)),
		zap.String("path", r.URL.Path),
		zap.String("method", r.Method),
		zap.String("requestID", s.GatewayService.GetOrSetRequestID(r)),
	}
	return logFields
}

func (s *GatewayServer) handleResponseBody(respBody []byte) []byte {
	// remove oracle field in the respBody
	var resp response.Response
	if respData := jsoniter.Get(respBody, "data"); respData != nil {
		resp.Data = respData.GetInterface()
	}
	if respCode := jsoniter.Get(respBody, "code"); respCode != nil {
		resp.Code = respCode.ToInt()
	}
	if respMessage := jsoniter.Get(respBody, "message"); respMessage != nil {
		resp.Message = respMessage.ToString()
	}
	if respDebugMessage := jsoniter.Get(respBody, "debugMessage"); respDebugMessage != nil {
		resp.DebugMessage = respDebugMessage.ToString()
	}
	res, _ := json.Marshal(resp)
	return res
}

func (s *GatewayServer) LogStatisticHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()                              // request timestamp
		r = response.WithLocalizer(r, s.LocalizeManager) // with localizer
		logFields := s.generateLogFields(r)
		s.logRequest(r, logFields) // log request body
		rw := httptest.NewRecorder()
		next.ServeHTTP(rw, r)
		var statisticInfo *statistic.StatisticInfo
		grpcServiceCache := s.GatewayService.GrpcServiceCache(r) // generate statistic info
		if grpcServiceCache != nil {
			_, symbol := s.GatewayService.GetSymbol(r, grpcServiceCache)
			if symbol != "" {
				statisticInfo = &statistic.StatisticInfo{
					Timestamp:       start,
					ResponseCode:    -1,
					Application:     grpcServiceCache.Application,
					Service:         grpcServiceCache.Name,
					Path:            r.URL.Path,
					ServiceDuration: 0,
				}
			}
		}
		respBody := rw.Body.Bytes() // response
		var codeValue int32
		if len(respBody) > 0 {
			s.logResponse(respBody, logFields) // log response body
			// grpc response fields, if has
			if code := jsoniter.Get(respBody, "code"); code != nil { // code field
				codeValue = code.ToInt32()
				logFields = append(logFields, zap.Int32("code", codeValue))
				if statisticInfo != nil {
					statisticInfo.ResponseCode = codeValue
				}
			}
			if message := jsoniter.Get(respBody, "message"); message != nil { // message field
				logFields = append(logFields, zap.String("message", message.ToString()))
			}
			if debugMessage := jsoniter.Get(respBody, "debugMessage"); debugMessage != nil { // debugMessage field
				logFields = append(logFields, zap.String("debugMessage", debugMessage.ToString()))
			}
			if oracle := jsoniter.Get(respBody, "oracle"); oracle != nil && statisticInfo != nil { // oracle field: json string
				if oracleString := oracle.ToString(); oracleString != "" {
					var oracleField statistic.OracleFieldType
					if err := json.Unmarshal([]byte(oracleString), &oracleField); err != nil {
						s.Logger.Error("unmarshal oracle field error", zap.NamedError("appError", err))
					} else {
						statisticInfo.ServiceDuration = oracleField.Duration
					}
				}
			}
			respBody = s.handleResponseBody(respBody)
		}
		for key, value := range rw.Header() { // response headers
			for _, v := range value {
				w.Header().Add(key, v)
			}
		}
		w.WriteHeader(rw.Code)
		w.Write(respBody)
		dur := time.Since(start)
		logFields = append(logFields, zap.String("duration", dur.String()))
		if codeValue == response.StatusCodeRequestTimeout { // warning log, if it is request timeout
			s.Logger.Warn("access log", logFields...)
		} else {
			s.Logger.Info("access log", logFields...)
		}
		if statisticInfo != nil {
			s.handleStatisticInfo(statisticInfo, dur)
		}
	})
}

func (s *GatewayServer) handleStatisticInfo(statisticInfo *statistic.StatisticInfo, dur time.Duration) {
	statisticInfo.Duration = int64(dur / time.Millisecond)
	beijing, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		return
	}
	date := statisticInfo.Timestamp.In(beijing)
	// upsert dateProtoStatistics
	dateFmt := date.Format("2006-01-02")
	// GatewayService lock
	s.GatewayService.Mu.Lock()
	dateProtoStatistics, has := s.GatewayService.ProtoStatistics[dateFmt]
	if !has {
		dateProtoStatistics = make(map[string]*statistic.ProtoStatisticHourly, 20)
	}
	// upsert pathProtoStatistic
	pathProtoStatistic, has := dateProtoStatistics[statisticInfo.Path]
	if !has {
		pathProtoStatistic = &statistic.ProtoStatisticHourly{
			Timestamp:   statisticInfo.Timestamp.UnixNano(),
			Application: statisticInfo.Application,
			Service:     statisticInfo.Service,
		}

	}
	if pathProtoStatistic.Path == "" { // path
		pathProtoStatistic.Path = statisticInfo.Path
	}
	pathProtoStatistic.Hit += 1          // hit
	if statisticInfo.ResponseCode == 0 { // success hit
		pathProtoStatistic.SuccessHit += 1
	}
	if !response.IsOracleGatewayErrorCode(statisticInfo.ResponseCode) { // proxy success hit
		pathProtoStatistic.ProxySuccessHit += 1
	}
	ms := statisticInfo.Duration // duration
	pathProtoStatistic.DurationTotal += ms
	if pathProtoStatistic.DurationMin == 0 || pathProtoStatistic.DurationMin > ms {
		pathProtoStatistic.DurationMin = ms
	}
	if pathProtoStatistic.DurationMax == 0 || pathProtoStatistic.DurationMax < ms {
		pathProtoStatistic.DurationMax = ms
	}
	ms = statisticInfo.ServiceDuration // service duration
	pathProtoStatistic.ServiceDurationTotal += ms
	if pathProtoStatistic.ServiceDurationMin == 0 || pathProtoStatistic.ServiceDurationMin > ms {
		pathProtoStatistic.ServiceDurationMin = ms
	}
	if pathProtoStatistic.ServiceDurationMax == 0 || pathProtoStatistic.ServiceDurationMax < ms {
		pathProtoStatistic.ServiceDurationMax = ms
	}
	dateProtoStatistics[statisticInfo.Path] = pathProtoStatistic
	s.GatewayService.ProtoStatistics[dateFmt] = dateProtoStatistics
	// GatewayService unlock
	s.GatewayService.Mu.Unlock()
}

func (s *GatewayServer) timeoutResponse(ctx context.Context, r *http.Request) *response.Response {
	date := time.Now()
	if err := s.DaoManager.DB.Transaction(func(tx *gorm.DB) error {
		daoManager := dao.NewManagerWithDB(tx)
		record, err := daoManager.TimeoutStatisticDAO.GetByDateAndPath(ctx, date, r.URL.Path)
		if err != nil {
			return err
		}
		if record == nil {
			// create a record
			serviceCache := s.GatewayService.GrpcServiceCache(r)
			if serviceCache == nil {
				return err
			} else {
				if err := daoManager.TimeoutStatisticDAO.Create(ctx, &TimeoutStatistic{
					Date:          date,
					ApplicationID: serviceCache.ApplicationID,
					ServiceID:     serviceCache.ID,
					Path:          r.URL.Path,
					Count:         1,
				}); err != nil {
					return err
				}
			}
		} else {
			// update the record
			record.Count++
			if err := daoManager.TimeoutStatisticDAO.UpdateCountByID(ctx, record.ID, record.Count); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return &response.Response{
			Code:         response.StatusCodeInternalServerError,
			Message:      response.LocalizeMessage(ctx, "FatalErrMsg"),
			DebugMessage: fmt.Sprintf("internal error: %v", err),
		}
	}
	return &response.Response{
		Code:         response.StatusCodeRequestTimeout,
		Message:      response.LocalizeMessage(ctx, "ClientErrMsg_RequestTimeout"),
		DebugMessage: fmt.Sprintf("bad request: %v", context.DeadlineExceeded),
	}
}

type respBodyChannelKey struct{}

func (s *GatewayServer) WithTimeout(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		respBodyChannel := make(chan []byte)
		oldCtx := r.Context()
		ctx, cancel := context.WithTimeout(context.WithValue(oldCtx, respBodyChannelKey{}, respBodyChannel), 30*time.Second)
		defer cancel()
		// response headers
		w.Header().Add("content-type", "application/json")
		r = r.WithContext(ctx)
		go next.ServeHTTP(w, r)
		var respBody []byte
		select {
		case respBody = <-respBodyChannel:
		case <-ctx.Done():
			respBody, _ = json.Marshal(s.timeoutResponse(oldCtx, r))
		}
		w.Write(respBody)
	})
}

func (s *GatewayServer) Http2GrpcProxy() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		// response body
		respBody, err := s.GatewayService.Http2GrpcProxy(ctx, w, r)
		if err != nil {
			var resp *response.Response
			if err == service.ErrUpstreamServiceNotFound || err == service.ErrUpstreamServiceMethodNotFound {
				resp = &response.Response{
					Code:         response.StatusCodeBadRequest,
					Message:      response.LocalizeMessage(ctx, "ClientErrMsg_BadRequest"),
					DebugMessage: fmt.Sprintf("bad request: %v", err),
				}
			} else if err == service.ErrGoRequestDelay {
				resp = &response.Response{
					Code:         response.StatusCodeGoRequestDelay,
					Message:      response.LocalizeMessage(ctx, "ClientErrMsg_BadRequest"),
					DebugMessage: fmt.Sprintf("bad request: %v", err),
				}
			} else if err == service.ErrGoRequestDeformedSecretKey {
				resp = &response.Response{
					Code:         response.StatusCodeGoRequestDeformedSecretKey,
					Message:      response.LocalizeMessage(ctx, "ClientErrMsg_BadRequest"),
					DebugMessage: fmt.Sprintf("bad request: %v", err),
				}
			} else {
				resp = &response.Response{
					Code:         response.StatusCodeInternalServerError,
					Message:      response.LocalizeMessage(ctx, "FatalErrMsg"),
					DebugMessage: fmt.Sprintf("internal error: %v", err),
				}
			}
			respBody, _ = json.Marshal(resp)
		}
		respBodyChannel := ctx.Value(respBodyChannelKey{})
		if respBodyChannel == nil {
			w.Write(respBody)
		} else {
			respBodyChannel.(chan []byte) <- respBody
		}
	})
}
