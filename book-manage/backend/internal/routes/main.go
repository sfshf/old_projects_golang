package routes

import (
	"crypto/tls"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang/glog"
	"github.com/nextsurfer/book-manage-api/internal/routes/middleware"
	"github.com/nextsurfer/book-manage-api/internal/tools"
)

// RequestType request type
type RequestType string

// RequestType get,post
const (
	RequestMethodPost RequestType = "POST"
	RequestMethodGet  RequestType = "GET"
)

// RouterOption config of a router
type RouterOption struct {
	requestMethod RequestType
	path          string
	handlerFunc   func(*gin.Context)
}

// RouterGroup config of a router group
type RouterGroup struct {
	path   string
	routes []*RouterOption
}

var router *gin.Engine
var routerGrpList []*RouterGroup

// var (
// 	router = gin.Default()
// 	// RouterGrpList route group list
// 	routerGrpList = make([]*RouterGroup, 0)
// )

// Run will start the server
func Run() {
	// gin.SetMode(gin.DebugMode)
	router = gin.New()

	// RouterGrpList route group list
	routerGrpList = make([]*RouterGroup, 0)

	router.Use(middleware.Download())

	router.Use(middleware.Cors())

	// router.Static("/download", tools.Config().DownloadPath)
	router.StaticFS("/download", http.Dir(tools.Config().DownloadPath))

	addRoutes()
	registerRoutes()

	errs := make(chan error)
	go func() {
		s := &http.Server{
			Addr:           tools.Config().ServerAddr,
			Handler:        router,
			MaxHeaderBytes: 1 << 20,
			TLSConfig:      &tls.Config{},
		}
		if err := s.ListenAndServe(); err != nil {
			errs <- err
		}
		glog.V(3).Info("Listening and serving HTTP on %s\n", tools.Config().ServerAddr)
	}()

	select {
	case err := <-errs:
		glog.Errorf("Could not start serving service due to (error: %s)", err.Error())
	}
}

func registerRoutes() {
	base := router.Group("/")

	for _, routerGrpCfg := range routerGrpList {
		rgObj := base.Group(routerGrpCfg.path)
		for _, r := range routerGrpCfg.routes {
			switch r.requestMethod {
			case RequestMethodGet:
				rgObj.GET(r.path, r.handlerFunc)
			case RequestMethodPost:
				rgObj.POST(r.path, r.handlerFunc)
			}
		}
	}
}
