package service

import (
	"bytes"
	"context"
	"crypto/cipher"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/fullstorydev/grpcurl"
	"github.com/go-redis/redis/v8"
	consulApi "github.com/hashicorp/consul/api"
	"github.com/jhump/protoreflect/desc"
	jsoniter "github.com/json-iterator/go"
	gerror "github.com/nextsurfer/ground/pkg/err"
	"github.com/nextsurfer/ground/pkg/localize"
	"github.com/nextsurfer/ground/pkg/rpc"
	gutil "github.com/nextsurfer/ground/pkg/util"
	"github.com/nextsurfer/oracle/api/response"
	"github.com/nextsurfer/oracle/internal/common/random"
	"github.com/nextsurfer/oracle/internal/common/simplecrypto"
	"github.com/nextsurfer/oracle/internal/common/simplehttp"
	"github.com/nextsurfer/oracle/internal/common/simpleproto"
	"github.com/nextsurfer/oracle/internal/common/statistic"
	"github.com/nextsurfer/oracle/internal/dao"
	. "github.com/nextsurfer/oracle/internal/model"
	"go.uber.org/zap"
	"go.uber.org/zap/buffer"
	"golang.org/x/crypto/chacha20poly1305"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/descriptorpb"
)

type GatewayService struct {
	*CronService

	Name                 string
	AppID                string
	Env                  gutil.APPEnvType
	Logger               *zap.Logger
	RedisClient          *redis.Client
	DaoManager           *dao.Manager
	ServiceCaches        map[string]*simpleproto.ServiceCache // pathPrefix->simpleproto.ServiceCache
	ConsulClient         *consulApi.Client
	LocalizeManager      *localize.Manager
	PrerequisiteProtos   []*descriptorpb.FileDescriptorProto
	StatisticInfoChannel chan *statistic.StatisticInfo
	ProtoStatistics      map[string]map[string]*statistic.ProtoStatisticHourly // date->{path->ProtoStatisticHourly}
	Mu                   sync.Mutex
}

func NewGatewayService(ctx context.Context, gatewayName, appID string, env gutil.APPEnvType, logger *zap.Logger, DaoManager *dao.Manager, redisClient *redis.Client, LocalizeManager *localize.Manager) (*GatewayService, error) {
	s := &GatewayService{
		Name:                 gatewayName,
		AppID:                appID,
		Env:                  env,
		Logger:               logger,
		RedisClient:          redisClient,
		DaoManager:           DaoManager,
		LocalizeManager:      LocalizeManager,
		StatisticInfoChannel: make(chan *statistic.StatisticInfo, 20),
		ProtoStatistics:      make(map[string]map[string]*statistic.ProtoStatisticHourly, 2),
	}
	// consul client
	config := consulApi.DefaultConfig()
	client, err := consulApi.NewClient(config)
	if err != nil {
		return nil, err
	}
	s.ConsulClient = client
	if err := s.loadAllServices(ctx); err != nil {
		return nil, err
	}
	// cron service
	cronService, err := NewCronService(ctx, s)
	if err != nil {
		return nil, err
	}
	s.CronService = cronService
	return s, nil
}

// upstream service cache -----------------------------------------------------------------------------------------

// no support for concurrent calls
func (s *GatewayService) loadAllServices(ctx context.Context) error {
	// load all prerequisite protos
	PrerequisiteProtos, err := s.loadAllPrerequisiteProtos(ctx)
	if err != nil {
		return err
	}
	s.PrerequisiteProtos = PrerequisiteProtos
	// load all applications
	applications, err := s.DaoManager.ApplicationDAO.GetAll(ctx)
	if err != nil {
		return err
	}
	// load all service protos
	services, err := s.DaoManager.ServiceDAO.GetAllServices(ctx, true /*omitProtoFile*/, false /*omitFileDescriptor*/)
	if err != nil {
		return err
	}
	// generate service caches
	servicesLength := len(services)
	ServiceCaches := make(map[string]*simpleproto.ServiceCache, servicesLength)
	for _, service := range services {
		// generate service FileDescriptorProto
		fdp, err := simpleproto.FileDescriptorProtoFromBase64(service.FileDescriptorData)
		if err != nil {
			return err
		}
		// service FileDescriptorProto must be the last one
		fd, err := desc.CreateFileDescriptorFromSet(&descriptorpb.FileDescriptorSet{File: append(PrerequisiteProtos, fdp)})
		if err != nil {
			return err
		}
		var appName string
		for _, app := range applications {
			if service.ApplicationID == app.ID {
				appName = app.Name
				break
			}
		}
		sc := &simpleproto.ServiceCache{
			ID:             service.ID,
			Name:           service.Name,
			ApplicationID:  service.ApplicationID,
			Application:    appName,
			URL:            service.URL,
			PathPrefix:     service.PathPrefix,
			ProtoFileMd5:   service.ProtoFileMd5,
			CreatedAt:      service.CreatedAt.UnixMilli(),
			FileDescriptor: fd,
		}
		ServiceCaches[service.PathPrefix] = sc
	}
	// assign to gateway server
	s.ServiceCaches = ServiceCaches
	return nil
}

func (s *GatewayService) loadAllPrerequisiteProtos(ctx context.Context) ([]*descriptorpb.FileDescriptorProto, error) {
	prerequisites, err := s.DaoManager.ServiceDAO.GetAllPrerequisites(ctx)
	if err != nil {
		return nil, err
	}
	// generate prerequisite FileDescriptorProto
	var PrerequisiteProtos []*descriptorpb.FileDescriptorProto
	for _, prerequisite := range prerequisites {
		fdp, err := simpleproto.FileDescriptorProtoFromBase64(prerequisite.FileDescriptorData)
		if err != nil {
			return nil, err
		}
		PrerequisiteProtos = append(PrerequisiteProtos, fdp)
	}
	return PrerequisiteProtos, nil
}

func (s *GatewayService) fetchFileDescriptor(ctx context.Context, service *Service) (*desc.FileDescriptor, error) {
	// generate service FileDescriptorProto
	fdp, err := simpleproto.FileDescriptorProtoFromBase64(service.FileDescriptorData)
	if err != nil {
		return nil, err
	}
	// service FileDescriptorProto must be the last one
	fd, err := desc.CreateFileDescriptorFromSet(&descriptorpb.FileDescriptorSet{File: append(s.PrerequisiteProtos, fdp)})
	if err != nil {
		return nil, err
	}
	return fd, nil
}

func (s *GatewayService) newServiceCache(ctx context.Context, service *Service, fd *desc.FileDescriptor) error {
	applications, err := s.DaoManager.ApplicationDAO.GetAll(ctx) // load all applications
	if err != nil {
		return err
	}
	var appName string
	for _, app := range applications {
		if service.ApplicationID == app.ID {
			appName = app.Name
			break
		}
	}
	s.ServiceCaches[service.PathPrefix] = &simpleproto.ServiceCache{
		ID:             service.ID,
		Name:           service.Name,
		ApplicationID:  service.ApplicationID,
		Application:    appName,
		URL:            service.URL,
		PathPrefix:     service.PathPrefix,
		ProtoFileMd5:   service.ProtoFileMd5,
		CreatedAt:      service.CreatedAt.UnixMilli(),
		FileDescriptor: fd,
	}
	return nil
}

// no support for concurrent calls
func (s *GatewayService) RefreshService(ctx context.Context, rpcCtx *rpc.Context, name string) *gerror.AppError {
	service, err := s.DaoManager.ServiceDAO.GetByName(ctx, name, true /*omitProtoFile*/, false /*omitDeleted*/) // fetch the service by name
	if err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	if service == nil {
		return nil
	}
	if service.DeletedAt > 0 { // delete service cache
		delete(s.ServiceCaches, service.PathPrefix)
		return nil
	}

	fd, err := s.fetchFileDescriptor(ctx, service)
	if err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	for k, v := range s.ServiceCaches {
		if v.Name == name {
			if k == service.PathPrefix {
				s.ServiceCaches[k].ProtoFileMd5 = service.ProtoFileMd5
				s.ServiceCaches[k].FileDescriptor = fd
				return nil
			} else {
				serviceCache := s.ServiceCaches[k]
				serviceCache.PathPrefix = service.PathPrefix
				serviceCache.ProtoFileMd5 = service.ProtoFileMd5
				serviceCache.FileDescriptor = fd
				delete(s.ServiceCaches, k)
				s.ServiceCaches[service.PathPrefix] = serviceCache
				return nil
			}
		}
	}
	// has no cache before
	if err := s.newServiceCache(ctx, service, fd); err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	return nil
}

func (s *GatewayService) getPathPrefix(r *http.Request) string {
	path := r.URL.Path
	for s.ServiceCaches[path] == nil {
		if slashIndex := strings.LastIndex(path, "/"); slashIndex >= 0 {
			path = path[:slashIndex]
		} else {
			break
		}
	}
	return path
}

// a symbol is a fully qualified name of a service method
func (s *GatewayService) GetSymbol(r *http.Request, sc *simpleproto.ServiceCache) (*desc.MethodDescriptor, string) {
	fd := sc.FileDescriptor
	// service name
	sds := fd.GetServices()
	for _, sd := range sds {
		// method name
		mds := sd.GetMethods()
		for _, md := range mds {
			mop := md.GetMethodOptions()
			if strings.Contains(mop.String(), r.URL.Path) {
				return md, md.GetFullyQualifiedName()
			}
		}
	}
	return nil, ""
}

func genDialOpts(isTLS bool) []grpc.DialOption {
	opts := []grpc.DialOption{grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy": "round_robin"}`)}
	if isTLS {
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{})))
	} else {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}
	return opts
}

func (s *GatewayService) dialUpstreamGrpcServer(r *http.Request, serviceCache *simpleproto.ServiceCache) (cc *grpc.ClientConn, err error) {
	// first, check consul service
	if err := s.checkConsulService(serviceCache.Name); err != nil {
		return nil, err
	}
	consulAddr := os.Getenv("CONSUL_HTTP_ADDR")
	if consulAddr == "" {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	return grpc.DialContext(ctx, fmt.Sprintf("consul://%s/%s", consulAddr, serviceCache.Name), genDialOpts(r.URL.Scheme == "https")...)
}

func httpHeaderToMD(headers http.Header, additionalHeader map[string]string) []string {
	headers.Set("content-type", "application/grpc")
	headers.Del("connection")
	md := make([]string, 0, len(headers)+len(additionalHeader))
	//md := metadata.New(map[string]string{})
	for key, value := range headers {
		if strings.ToLower(key) == "user-agent" {
			for _, v := range value {
				md = append(md, fmt.Sprintf("%s: %s", key, v))
			}
			continue
		}
		for _, v := range value {
			md = append(md, fmt.Sprintf("%s: %s", key, v))
		}
	}
	for key, value := range additionalHeader {
		md = append(md, fmt.Sprintf("%s: %s", key, value))
	}
	return md
}

var (
	ErrUpstreamServiceNotFound       = errors.New("upstream service not found")
	ErrUpstreamServiceMethodNotFound = errors.New("upstream service method not found")
	ErrConsulServiceNotFound         = errors.New("consul service not found")
	ErrGoRequestDelay                = errors.New("go request delay")
	ErrGoRequestDeformedSecretKey    = errors.New("deformed secret key")
)

func (s *GatewayService) checkConsulService(serviceName string) error {
	services, _, err := s.ConsulClient.Catalog().Services(nil)
	if err != nil {
		return err
	}
	for srv := range services {
		if srv == serviceName {
			return nil
		}
	}
	return ErrConsulServiceNotFound
}

func (s *GatewayService) GetOrSetRequestID(r *http.Request) string {
	requestID := r.Header.Get("request-id")
	if requestID == "" {
		requestID = random.GenerateRequestId()
		r.Header.Add("request-id", requestID)
	}
	return requestID
}

func (s *GatewayService) GrpcServiceCache(r *http.Request) *simpleproto.ServiceCache {
	return s.ServiceCaches[s.getPathPrefix(r)]
}

type GoRequest struct {
	Path      string `json:"path" validate:"required"`
	Data      string `json:"data" validate:""`
	Timestamp int64  `json:"timestamp" validate:"required"`
}

type GoEncryptedData struct {
	EncryptedData interface{} `json:"encryptedData"`
}

func (s *GatewayService) handleGoRequest(r *http.Request) ([]byte, cipher.AEAD, error) {
	// 1. read request body
	reqBody, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, nil, err
	}
	var req GoRequest
	if err := json.Unmarshal(reqBody, &req); err != nil {
		return nil, nil, err
	}
	// 2. check timestamp
	ts := time.UnixMilli(req.Timestamp)
	if time.Since(ts) > 1*time.Minute {
		return nil, nil, ErrGoRequestDelay
	}
	// 3. decrypt base64-encoded, ChaCha20-Poly1305 encrypted data
	goKey, err := simplecrypto.Keccak256([]byte(fmt.Sprintf("%d9C9B913EB1B6254F4737CE947", req.Timestamp)))
	if err != nil {
		return nil, nil, err
	}
	aead, err := chacha20poly1305.NewX(goKey)
	if err != nil {
		return nil, nil, err
	}
	decryptedReq, err := simplecrypto.DecryptByX([]byte(req.Data), aead, simplecrypto.NonceZeroX())
	if err != nil {
		return nil, nil, ErrGoRequestDeformedSecretKey
	}
	// 4. update the request path and body
	r.URL.Path = req.Path
	r.Body = io.NopCloser(bytes.NewBuffer(decryptedReq))
	return goKey, aead, nil
}

func (s *GatewayService) handleGoResponse(respBody []byte, aead cipher.AEAD) ([]byte, error) {
	var resp response.Response
	var encryptedData []byte
	if respData := jsoniter.Get(respBody, "data"); respData != nil {
		encryptedData = simplecrypto.EncryptByX([]byte(respData.ToString()), aead, nil, simplecrypto.NonceZeroX())
	}
	resp.Data = GoEncryptedData{EncryptedData: string(encryptedData)}
	if respCode := jsoniter.Get(respBody, "code"); respCode != nil {
		resp.Code = respCode.ToInt()
	}
	if respMessage := jsoniter.Get(respBody, "message"); respMessage != nil {
		resp.Message = respMessage.ToString()
	}
	if respDebugMessage := jsoniter.Get(respBody, "debugMessage"); respDebugMessage != nil {
		resp.DebugMessage = respDebugMessage.ToString()
	}
	if respOracle := jsoniter.Get(respBody, "oracle"); respOracle != nil {
		resp.Oracle = respOracle.ToString()
	}
	return json.Marshal(resp)
}

func (s *GatewayService) Http2GrpcProxy(ctx context.Context, w http.ResponseWriter, r *http.Request) ([]byte, error) {
	var err error
	// update request instance, if path is /go
	var goKey []byte
	var aead cipher.AEAD
	if r.URL.Path == "/go" {
		goKey, aead, err = s.handleGoRequest(r)
		if err != nil {
			return nil, err
		}
	}
	// proxy to upstream grpc service
	serviceCache := s.GrpcServiceCache(r) // check service cache
	if serviceCache == nil {
		return nil, ErrUpstreamServiceNotFound
	}
	methodDescriptor, symbol := s.GetSymbol(r, serviceCache) // check service path
	if symbol == "" {
		return nil, ErrUpstreamServiceMethodNotFound
	}
	descSource, err := grpcurl.DescriptorSourceFromFileDescriptors(serviceCache.FileDescriptor) // generate desc source
	if err != nil {
		return nil, err
	}
	rf, formatter, err := grpcurl.RequestParserAndFormatter(grpcurl.FormatJSON, descSource, r.Body, grpcurl.FormatOptions{AllowUnknownFields: true})
	if err != nil {
		return nil, err
	}
	conn, err := s.dialUpstreamGrpcServer(r, serviceCache) // generate grpc client conn
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	// fetch and handle grpc response from upstream
	upstreamResponse := NewResponse()
	if err = grpcurl.InvokeRPC(ctx, descSource, conn, symbol,
		httpHeaderToMD(
			r.Header,
			map[string]string{
				"X-Forwarded-IP": simplehttp.GetRealIP(r),
				"X-Request-Id":   s.GetOrSetRequestID(r),
			},
		),
		&grpcurl.DefaultEventHandler{
			VerbosityLevel: 2,
			Out:            upstreamResponse,
			Formatter:      formatter,
		},
		rf.Next,
	); err != nil {
		return nil, err
	}
	for key, value := range upstreamResponse.header {
		w.Header().Add(key, value)
	}
	// handle upstream response body. emit default, if goKey is nil
	respBody, err := s.handleUpstreamResponseBody(upstreamResponse.Body(), methodDescriptor)
	if err != nil {
		return nil, err
	}
	if goKey == nil {
		return respBody, nil
	}
	return s.handleGoResponse(respBody, aead)
}

func (s *GatewayService) handleUpstreamResponseBody(body []byte, md *desc.MethodDescriptor) ([]byte, error) {
	if body == nil {
		return nil, nil
	}
	jstream := jsoniter.NewStream(jsoniter.ConfigCompatibleWithStandardLibrary, nil, len(body))
	respMD := md.GetOutputType()
	if err := s.handleResponseObject(body, jstream, respMD); err != nil {
		return nil, err
	}
	return jstream.Buffer(), nil
}

func (s *GatewayService) handleStruct(field *desc.FieldDescriptor, fieldValue jsoniter.Any, stream *jsoniter.Stream) error {
	switch ft := field.GetType(); ft { // when is struct
	case descriptorpb.FieldDescriptorProto_TYPE_MESSAGE:
		val := fieldValue.GetInterface() // when is nil
		if val == nil {
			stream.WriteVal(val)
			break
		}
		if field.GetMessageType().GetFullyQualifiedName() == "google.protobuf.Value" {
			fieldValue.WriteTo(stream)
			break
		}
		data, err := json.Marshal(val)
		if err != nil {
			return err
		}
		if err := s.handleResponseObject(data, stream, field.GetMessageType()); err != nil {
			return err
		}
	case descriptorpb.FieldDescriptorProto_TYPE_GROUP:
		return errors.New("FieldDescriptorProto_TYPE_GROUP not support")
	case descriptorpb.FieldDescriptorProto_TYPE_INT64, descriptorpb.FieldDescriptorProto_TYPE_SINT64, descriptorpb.FieldDescriptorProto_TYPE_SFIXED64:
		stream.WriteInt64(fieldValue.ToInt64())
	case descriptorpb.FieldDescriptorProto_TYPE_INT32, descriptorpb.FieldDescriptorProto_TYPE_SINT32, descriptorpb.FieldDescriptorProto_TYPE_SFIXED32:
		stream.WriteInt32(fieldValue.ToInt32())
	case descriptorpb.FieldDescriptorProto_TYPE_STRING:
		stream.WriteString(fieldValue.ToString())
	case descriptorpb.FieldDescriptorProto_TYPE_UINT64, descriptorpb.FieldDescriptorProto_TYPE_FIXED64:
		stream.WriteUint64(fieldValue.ToUint64())
	case descriptorpb.FieldDescriptorProto_TYPE_UINT32, descriptorpb.FieldDescriptorProto_TYPE_FIXED32:
		stream.WriteUint32(fieldValue.ToUint32())
	case descriptorpb.FieldDescriptorProto_TYPE_DOUBLE:
		stream.WriteFloat64(fieldValue.ToFloat64())
	case descriptorpb.FieldDescriptorProto_TYPE_FLOAT:
		stream.WriteFloat32(fieldValue.ToFloat32())
	case descriptorpb.FieldDescriptorProto_TYPE_BYTES:
		data, err := base64.StdEncoding.DecodeString(fieldValue.ToString())
		if err != nil {
			fieldValue.WriteTo(stream)
		} else {
			stream.WriteString(string(data))
		}
	case descriptorpb.FieldDescriptorProto_TYPE_ENUM:
		fieldValue.WriteTo(stream)
	case descriptorpb.FieldDescriptorProto_TYPE_BOOL:
		stream.WriteBool(fieldValue.ToBool())
	default:
		return fmt.Errorf("descriptor pb type [%s] not support", descriptorpb.FieldDescriptorProto_Type_name[int32(ft)])
	}
	return nil
}

func (s *GatewayService) handleResponseObject(body []byte, stream *jsoniter.Stream, msg *desc.MessageDescriptor) error {
	stream.WriteObjectStart()
	fields := msg.GetFields()
	// iterate fields of the struct
	for idx, field := range fields {
		fieldJsonName := field.GetJSONName()
		stream.WriteObjectField(fieldJsonName)
		fieldValue := jsoniter.Get(body, fieldJsonName)
		if field.IsRepeated() { // when is array
			if err := s.handleResponseArray(fieldValue, stream, field); err != nil {
				return err
			}
		} else {
			if err := s.handleStruct(field, fieldValue, stream); err != nil {
				return err
			}
		}
		if idx != len(fields)-1 {
			stream.Write([]byte{','})
		}
	}
	stream.WriteObjectEnd()
	return nil
}

func (s *GatewayService) handleResponseArray(array jsoniter.Any, stream *jsoniter.Stream, field *desc.FieldDescriptor) error {
	stream.WriteArrayStart()
	size := array.Size()
	for i := 0; i < size; i++ {
		elemValue := array.Get(i)
		val := elemValue.GetInterface()
		if val == nil {
			stream.WriteVal(val)
			break
		}
		switch ft := field.GetType(); ft {
		case descriptorpb.FieldDescriptorProto_TYPE_MESSAGE:
			// only support struct in array, not support array in array
			data, err := json.Marshal(val)
			if err != nil {
				return err
			}
			if err := s.handleResponseObject(data, stream, field.GetMessageType()); err != nil {
				return err
			}
		case descriptorpb.FieldDescriptorProto_TYPE_GROUP:
			return errors.New("FieldDescriptorProto_TYPE_GROUP in array not support")
		case descriptorpb.FieldDescriptorProto_TYPE_INT64, descriptorpb.FieldDescriptorProto_TYPE_SINT64, descriptorpb.FieldDescriptorProto_TYPE_SFIXED64:
			stream.WriteInt64(elemValue.ToInt64())
		case descriptorpb.FieldDescriptorProto_TYPE_INT32, descriptorpb.FieldDescriptorProto_TYPE_SINT32, descriptorpb.FieldDescriptorProto_TYPE_SFIXED32:
			stream.WriteInt32(elemValue.ToInt32())
		case descriptorpb.FieldDescriptorProto_TYPE_STRING:
			stream.WriteString(elemValue.ToString())
		case descriptorpb.FieldDescriptorProto_TYPE_UINT64, descriptorpb.FieldDescriptorProto_TYPE_FIXED64:
			stream.WriteUint64(elemValue.ToUint64())
		case descriptorpb.FieldDescriptorProto_TYPE_UINT32, descriptorpb.FieldDescriptorProto_TYPE_FIXED32:
			stream.WriteUint32(elemValue.ToUint32())
		case descriptorpb.FieldDescriptorProto_TYPE_DOUBLE:
			stream.WriteFloat64(elemValue.ToFloat64())
		case descriptorpb.FieldDescriptorProto_TYPE_FLOAT:
			stream.WriteFloat32(elemValue.ToFloat32())
		case descriptorpb.FieldDescriptorProto_TYPE_BYTES:
			elemValue.WriteTo(stream)
		case descriptorpb.FieldDescriptorProto_TYPE_ENUM:
			elemValue.WriteTo(stream)
		case descriptorpb.FieldDescriptorProto_TYPE_BOOL:
			stream.WriteBool(elemValue.ToBool())
		default:
			return fmt.Errorf("descriptor pb type [%s] in array not support", descriptorpb.FieldDescriptorProto_Type_name[int32(ft)])
		}
		if i != size-1 {
			stream.Write([]byte{','})
		}
	}
	stream.WriteArrayEnd()
	return nil
}

var (
	_responseHeaderPre  = "\nResponse headers received:\n"
	_responseContentPre = "\nResponse contents:\n"
	_responseTrailerPre = "\nResponse trailers received:\n"
)

func NewResponse() *Response {
	return &Response{
		header:    make(map[string]string),
		bodyWrite: false,
		body:      &buffer.Buffer{},
	}
}

type Response struct {
	header    map[string]string
	bodyWrite bool
	body      *buffer.Buffer
}

func (r *Response) Write(p []byte) (n int, err error) {
	str := string(p)
	if strings.HasPrefix(str, _responseHeaderPre) || strings.HasPrefix(str, _responseTrailerPre) {
		str = strings.Replace(str, _responseHeaderPre, "", 1)
		str = strings.Replace(str, _responseTrailerPre, "", 1)
		headers := strings.Split(str, "\n")
		if len(headers) == 2 && strings.HasPrefix(headers[1], "(empty)") {
			return len(p), nil
		}
		for _, header := range headers {
			if strings.TrimSpace(header) == "" {
				continue
			}

			values := strings.Split(header, ":")
			var v string
			if len(values) > 1 {
				v = values[1]
				r.header[values[0]] = v
			}
		}
	}
	if strings.HasPrefix(str, _responseContentPre) {
		r.bodyWrite = true
		return len(p), nil
	}
	if r.bodyWrite {
		r.body.Write(p)
		r.bodyWrite = false
	}
	return len(p), nil
}

func (r *Response) Body() []byte {
	return r.body.Bytes()
}

func (r *Response) Header() map[string]string {
	return r.header
}
