package server

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/NYTimes/gziphandler"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go/logging"
	"github.com/go-playground/validator/v10"
	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
	"github.com/juju/ratelimit"
	_ "github.com/mbobakov/grpc-consul-resolver"
	"github.com/nextsurfer/ground/pkg/localize"
	"github.com/nextsurfer/ground/pkg/rpc"
	gutil "github.com/nextsurfer/ground/pkg/util"
	console_api "github.com/nextsurfer/oracle/api/console"
	gateway_api "github.com/nextsurfer/oracle/api/gateway"
	"github.com/nextsurfer/oracle/api/response"
	"github.com/nextsurfer/oracle/internal/common/acme"
	"github.com/nextsurfer/oracle/internal/dao"
	"github.com/nextsurfer/oracle/internal/gateway/service"
	. "github.com/nextsurfer/oracle/internal/model"
	console_grpc "github.com/nextsurfer/oracle/pkg/console/grpc"
	"github.com/rs/cors"
	"go.uber.org/zap"
	"google.golang.org/grpc/reflection"
)

type GatewayServer struct {
	gateway_api.UnimplementedGatewayServiceServer
	Name                  string
	Host                  string
	HttpPort              int
	TlsPort               int
	RpcPort               int
	Env                   gutil.APPEnvType
	Logger                *zap.Logger
	RedisClient           *redis.Client
	DaoManager            *dao.Manager
	LocalizeManager       *localize.Manager
	HttpServer            *http.Server
	TlsServer             *http.Server
	TlsCerts              map[string]tls.Certificate // domain->tls.Certificate
	HostManage            map[string]string          // domain->rawURL
	GatewayService        *service.GatewayService
	ServiceRateLimitRules map[string]*ratelimit.Bucket
	PathRateLimitRules    map[string]*ratelimit.Bucket
	RpcServer             *rpc.Server
	Validator             *validator.Validate
	S3Client              *s3.Client
}

func NewGatewayServer(ctx context.Context, appID string, Env gutil.APPEnvType, Logger *zap.Logger, DaoManager *dao.Manager, redisClient *redis.Client, Host string, HttpPort, TlsPort, RpcPort int, tomlPath string, validator *validator.Validate) (*GatewayServer, error) {
	s := &GatewayServer{
		Host:        Host,
		HttpPort:    HttpPort,
		TlsPort:     TlsPort,
		RpcPort:     RpcPort,
		Env:         Env,
		Logger:      Logger,
		RedisClient: redisClient,
		DaoManager:  DaoManager,
		Validator:   validator,
	}
	// important!!! gateway name
	gatewayName := fmt.Sprintf("gateway-%d", time.Now().UnixMilli())
	s.Name = gatewayName
	// gateway Host manage
	if err := s.loadHostManage(ctx); err != nil {
		return nil, err
	}
	// register http, tls, grpc servers
	s.registerServers(ctx)
	s.LocalizeManager = s.RpcServer.LocolizerManager
	gatewayService, err := service.NewGatewayService(ctx, gatewayName, appID, Env, Logger, DaoManager, redisClient, s.LocalizeManager)
	if err != nil {
		return nil, err
	}
	s.GatewayService = gatewayService
	// localize manager load message files
	if err := s.loadMessageFiles(tomlPath); err != nil {
		return nil, err
	}
	// register to console app
	if err := s.registerGatewayNode(ctx, s.Name); err != nil {
		return nil, err
	}
	// load rate limit rules
	if err := s.loadRateLimitRules(); err != nil {
		return nil, err
	}
	// aws
	cfg, awsErr := config.LoadDefaultConfig(ctx, config.WithLogger(s))
	if awsErr != nil {
	}
	// Create an Amazon S3 service client
	s.S3Client = s3.NewFromConfig(cfg)
	return s, nil
}

func (s *GatewayServer) Logf(classification logging.Classification, format string, v ...interface{}) {
	log := fmt.Sprintf(format, v...)
	if classification == logging.Warn {
		zap.L().Warn("AWS Warning : ", zap.String("log", log))
	} else if classification == logging.Debug {
		zap.L().Debug("AWS Debug : ", zap.String("log", log))
	}
}

func (s *GatewayServer) AcmeHttp01HandleFunc(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	token := vars["token"]
	// get acme resource record from db
	acmeResource, err := s.DaoManager.AcmeResourceDAO.GetByToken(ctx, token)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if acmeResource == nil {
		http.Error(w, "nil acme resource", http.StatusBadRequest)
		return
	}
	if strings.HasPrefix(r.Host, acmeResource.Domain) {
		w.Header().Add("Content-Type", "text/plain")
		_, err := w.Write([]byte(acmeResource.KeyAuth))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		s.Logger.Info(fmt.Sprintf("[%s] Served key authentication", acmeResource.Domain))
	} else {
		s.Logger.Warn(fmt.Sprintf("Received request for domain %s with method %s but the domain did not match any challenge. Please ensure your are passing the HOST header properly.", r.Host, r.Method))
		_, err := w.Write([]byte("TEST"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func (s *GatewayServer) loadHostManage(ctx context.Context) error {
	list, err := s.DaoManager.HostManageDAO.GetAll(ctx)
	if err != nil {
		return err
	}
	s.HostManage = make(map[string]string, len(list))
	for _, item := range list {
		s.HostManage[item.Domain] = item.RawURL
	}
	return nil
}

func (s *GatewayServer) registerHttpServer(ctx context.Context) error {
	httpMux := mux.NewRouter().StrictSlash(true)
	httpMux.HandleFunc("/.well-known/acme-challenge/{token}", s.AcmeHttp01HandleFunc).Methods(http.MethodGet)
	httpMux.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		proxyUrl, is := s.isReverseProxy(r)
		if is {
			(&httputil.ReverseProxy{Rewrite: func(r *httputil.ProxyRequest) {
				r.SetURL(proxyUrl)
				r.Out.Host = r.In.Host // if desired
				r.Out.Method = r.In.Method
			}}).ServeHTTP(w, r)
			return
		}
		http.Redirect(w, r, fmt.Sprintf("https://%s:443%s", os.Getenv("GATEWAY_HOST"), r.RequestURI), http.StatusMovedPermanently)
	}).Methods(http.MethodGet, http.MethodHead, http.MethodPost, http.MethodPut,
		http.MethodPatch, http.MethodDelete, http.MethodConnect, http.MethodOptions, http.MethodTrace)
	s.HttpServer = &http.Server{Addr: fmt.Sprintf("0.0.0.0:%d", s.HttpPort), Handler: httpMux}
	return nil
}

func (s *GatewayServer) registerTlsServer(ctx context.Context) error {
	hostnames, err := s.DaoManager.HostManageDAO.GetAll(ctx)
	if err != nil {
		return err
	}
	allowedOrigins := make([]string, 0, len(hostnames))
	gatewayInitCorsOrigins := os.Getenv("ORACLE_GATEWAY_INIT_CORS")
	allowedOrigins = append(
		allowedOrigins,
		strings.Split(gatewayInitCorsOrigins, ",")...,
	)
	for _, item := range hostnames {
		allowedOrigins = append(allowedOrigins, "https://"+item.Domain)
	}
	// mux
	var tlsMux http.ServeMux
	tlsMux.Handle("/", s.GatewayRateLimit(gziphandler.GzipHandler(s.ReverseProxy(s.LogStatisticHandler(s.UploadHandler(s.WithTimeout(s.Http2GrpcProxy())))))))
	// global middlewares
	// cors
	handler := cors.New(cors.Options{
		AllowedOrigins: allowedOrigins,
		AllowedMethods: []string{
			http.MethodHead,
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
		},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
		ExposedHeaders:   []string{"Set-Cookie"},
	}).Handler(&tlsMux)
	s.TlsCerts = make(map[string]tls.Certificate)
	s.TlsServer = &http.Server{
		Addr:    fmt.Sprintf("0.0.0.0:%d", s.TlsPort),
		Handler: handler,
		TLSConfig: &tls.Config{
			GetCertificate: func(info *tls.ClientHelloInfo) (*tls.Certificate, error) {
				cert, has := s.TlsCerts[info.ServerName]
				if !has {
					// first, fetch from db
					acmeResource, err := s.DaoManager.AcmeResourceDAO.GetByDomain(ctx, info.ServerName)
					if err != nil {
						return nil, err
					}
					if acmeResource == nil {
						return nil, fmt.Errorf("no acme resource of %s", info.ServerName)
					}
					if acmeResource.Certificate == "" || acmeResource.PrivateKey == "" {
						return nil, fmt.Errorf("no tls certificate of %s", info.ServerName)
					}
					// second, cache locally
					cert, err = tls.X509KeyPair([]byte(acmeResource.Certificate), []byte(acmeResource.PrivateKey))
					if err != nil {
						return nil, err
					}
					s.TlsCerts[info.ServerName] = cert
				}
				return &cert, nil
			},
		},
	}
	return nil
}

func (s *GatewayServer) registerServers(ctx context.Context) error {
	// 1. http server -- acme http01 handle
	if err := s.registerHttpServer(ctx); err != nil {
		return err
	}
	// 2. tls server
	if err := s.registerTlsServer(ctx); err != nil {
		return err
	}
	// 3. grpc server
	tracer := rpc.NewTracer(s.Name, s.Env)
	server, err := rpc.NewServer(s.Name, s.Env, s.Host, s.RpcPort, tracer)
	if err != nil {
		return err
	}
	grpcServer := server.GrpcServer()
	gateway_api.RegisterGatewayServiceServer(grpcServer, s)
	reflection.Register(grpcServer)
	s.RpcServer = server
	return nil
}

// load i18n toml files
func (s *GatewayServer) loadMessageFiles(tomlPath string) error {
	files, err := ioutil.ReadDir(tomlPath)
	if err != nil {
		return err
	}
	for _, f := range files {
		if strings.HasSuffix(f.Name(), ".toml") {
			fullName := fmt.Sprintf("%s/%s", strings.TrimRight(tomlPath, "/"), f.Name())
			s.LocalizeManager.Bundle.MustLoadMessageFile(fullName)
		}
	}
	return nil
}

// register the gateway node to console platform
func (s *GatewayServer) registerGatewayNode(ctx context.Context, gatewayName string) error {

	respData, err := console_grpc.RegisterGatewayNode(ctx, &console_api.RegisterGatewayNodeRequest{
		Name:    gatewayName,
		Ipv4:    os.Getenv("GATEWAY_HOST"), // gateway node local ipv4
		RpcPort: int32(s.RpcPort),
	})
	if err != nil {
		return err
	}
	if respData.Code != response.StatusCodeOK {
		return errors.New("register gateway node failed")
	}
	// async renew api host acme resource
	go func() {
		gatewayApiHostname := os.Getenv("GATEWAY_API_HOSTNAME")
		acmeResource, err := s.DaoManager.AcmeResourceDAO.GetByDomain(ctx, gatewayApiHostname)
		if err != nil {
			panic(err)
		}
		if acmeResource == nil || acmeResource.Certificate == "" || acmeResource.PrivateKey == "" {
			// renew an acme resource
			http01Provider := acme.NewHttp01Provider(s.DaoManager) // !!! important, it has dao operations
			resource, err := acme.NewAcmeResource(gatewayApiHostname, http01Provider)
			if err != nil {
				panic(fmt.Errorf("NewAcmeResource tls certificate of domain [%s] failed: %s", gatewayApiHostname, err))
			}
			acmeResource, err = s.DaoManager.AcmeResourceDAO.GetByDomain(ctx, gatewayApiHostname)
			if err != nil {
				panic(err)
			}
			if acmeResource == nil {
				acmeResource = &AcmeResource{
					Domain:            gatewayApiHostname,
					CertURL:           resource.CertURL,
					CertStableURL:     resource.CertStableURL,
					PrivateKey:        string(resource.PrivateKey),
					Certificate:       string(resource.Certificate),
					IssuerCertificate: string(resource.IssuerCertificate),
					Csr:               string(resource.CSR),
				}
				// insert the record in db
				if err := s.DaoManager.AcmeResourceDAO.Create(ctx, acmeResource); err != nil {
					panic(fmt.Errorf("create tls certificate of domain [%s] to db failed: %s", acmeResource.Domain, err))
				}
			} else {
				acmeResource.CertURL = resource.CertURL
				acmeResource.CertStableURL = resource.CertStableURL
				acmeResource.PrivateKey = string(resource.PrivateKey)
				acmeResource.Certificate = string(resource.Certificate)
				acmeResource.IssuerCertificate = string(resource.IssuerCertificate)
				acmeResource.Csr = string(resource.CSR)
				// update the record in db
				if err := s.DaoManager.AcmeResourceDAO.Update(ctx, acmeResource); err != nil {
					panic(fmt.Errorf("update tls certificate of domain [%s] to db failed: %s", acmeResource.Domain, err))
				}
			}
		}
	}()
	return nil
}

func (s *GatewayServer) getServicePaths(protoFile string) []string {
	re := regexp.MustCompile(`post\s*:\s*"(.+)"`)
	matrix := re.FindAllStringSubmatch(protoFile, -1)
	var res []string
	for _, slice := range matrix {
		res = append(res, slice[1])
	}
	return res
}

func (s *GatewayServer) loadRateLimitRules() error {
	ctx := context.Background()
	s.ServiceRateLimitRules = make(map[string]*ratelimit.Bucket)
	s.PathRateLimitRules = make(map[string]*ratelimit.Bucket)
	// first, fetch special rule named 'all'
	allRule, err := s.DaoManager.RateLimitRuleDAO.GetServiceRuleByName(ctx, "all")
	if err != nil {
		return err
	}
	if allRule != nil && allRule.Enabled {
		s.Logger.Info("load special rate limit rule -- all",
			zap.String("target", allRule.Target),
			zap.Int("capacity", int(allRule.Capacity)),
		)
		// iterate all services
		services, err := s.DaoManager.ServiceDAO.GetAllServices(ctx, false /*omitProtoFile*/, true /*omitFileDescriptor*/)
		if err != nil {
			return err
		}
		for _, service := range services {
			// iterate all paths
			paths := s.getServicePaths(service.ProtoFile)
			for _, path := range paths {
				s.PathRateLimitRules[path] = ratelimit.NewBucket(
					1*time.Second,
					allRule.Capacity,
				)
			}
		}
	}
	// second, iterate other rules
	rateLimitRules, err := s.DaoManager.RateLimitRuleDAO.GetAll(ctx)
	if err != nil {
		return err
	}
	for _, rateLimitRule := range rateLimitRules {
		if !rateLimitRule.Enabled {
			break
		}
		if rateLimitRule.Type == 1 { // service
			s.Logger.Info("load service rate limit rule",
				zap.String("target", rateLimitRule.Target),
				zap.Int("capacity", int(rateLimitRule.Capacity)),
			)
			if rateLimitRule.Target != "all" { // all is a special target, it is used for all path of all service
				s.ServiceRateLimitRules[rateLimitRule.Target] = ratelimit.NewBucket(
					1*time.Second,
					rateLimitRule.Capacity,
				)
			}
		} else if rateLimitRule.Type == 2 { // url path
			s.Logger.Info("load path rate limit rule",
				zap.String("target", rateLimitRule.Target),
				zap.Int("capacity", int(rateLimitRule.Capacity)),
			)
			s.PathRateLimitRules[rateLimitRule.Target] = ratelimit.NewBucket(
				1*time.Second,
				rateLimitRule.Capacity,
			)
		}
	}
	return nil
}

func (s *GatewayServer) Run(ctx context.Context) error {
	// 1. http server
	go func() {
		s.Logger.Info(fmt.Sprintf("Gateway http server is running at %s ...", s.HttpServer.Addr))
		if err := s.HttpServer.ListenAndServe(); err != nil {
			if err != http.ErrServerClosed {
				s.Logger.Fatal("Gateway http server error", zap.NamedError("appError", err))
			} else {
				s.Logger.Info("Gateway http server closed")
			}
		}
	}()

	// 2. tls server
	go func() {
		s.Logger.Info(fmt.Sprintf("Gateway tls server is running at %s ...", s.TlsServer.Addr))
		if err := s.TlsServer.ListenAndServeTLS("", ""); err != nil {
			if err != http.ErrServerClosed {
				s.Logger.Fatal("Gateway tls server error", zap.NamedError("appError", err))
			} else {
				s.Logger.Info("Gateway tls server closed")
			}
		}
	}()

	// 3. grpc server
	go func() {
		s.Logger.Info(fmt.Sprintf("Gateway app grpc server is running at %s:%d ...", s.Host, s.RpcPort))
		if err := s.RpcServer.Start(); err != nil {
			s.Logger.Fatal("Gateway app grpc server error", zap.NamedError("appError", err))
		}
	}()
	return nil
}

func (s *GatewayServer) Stop(ctx context.Context) error {
	s.Logger.Info("Gateway servers are stopping ...")
	err := s.RpcServer.Stop()
	if err := s.HttpServer.Shutdown(ctx); err != nil {
		return err
	}
	if err := s.TlsServer.Shutdown(ctx); err != nil {
		return err
	}
	return err
}
