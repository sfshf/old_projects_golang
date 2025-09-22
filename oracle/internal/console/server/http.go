package server

import (
	"context"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"strings"

	"github.com/nextsurfer/oracle/api/response"
	"github.com/nextsurfer/oracle/internal/common/simplehash"
	"github.com/nextsurfer/oracle/internal/common/simplehttp"
	"github.com/nextsurfer/oracle/internal/common/simpleproto"
	"github.com/nextsurfer/oracle/internal/console/service"
	. "github.com/nextsurfer/oracle/internal/model"
	"github.com/rs/cors"
)

func (s *ConsoleServer) registerRoutes() {
	// http mux
	var mux http.ServeMux
	mux.Handle("/", http.FileServer(http.Dir(s.WebPath)))
	mux.HandleFunc("/console/ping/v1", s.Ping)
	// application
	mux.Handle("/console/listApplications/v1", s.Handler(&service.ListApplicationsRequest{}, s.ConsoleService.ListApplications))
	mux.Handle("/console/deleteApplication/v1", s.Handler(&service.DeleteApplicationRequest{}, s.ConsoleService.DeleteApplication))
	// service
	mux.Handle("/console/listServices/v1", s.Handler(&service.ListServicesRequest{}, s.ConsoleService.ListServices))
	mux.Handle("/console/listServicePaths/v1", s.Handler(&service.ListServicePathsRequest{}, s.ConsoleService.ListServicePaths))
	mux.Handle("/console/deleteService/v1", s.Handler(&service.DeleteServiceRequest{}, s.ConsoleService.DeleteService))
	// gateway node
	mux.Handle("/console/listGatewayNodes/v1", s.Handler(&service.ListGatewayNodesRequest{}, s.ConsoleService.ListGatewayNodes))
	// acme resource
	mux.Handle("/console/listAcmeResources/v1", s.Handler(&service.ListAcmeResourcesRequest{}, s.ConsoleService.ListAcmeResources))
	mux.Handle("/console/renewAcmeResource/v1", s.Handler(&service.RenewAcmeResourceRequest{}, s.ConsoleService.RenewAcmeResource))
	// host manage
	mux.Handle("/console/createHostname/v1", s.Handler(&service.CreateHostnameRequest{}, s.ConsoleService.CreateHostname))
	mux.Handle("/console/updateHostname/v1", s.Handler(&service.UpdateHostnameRequest{}, s.ConsoleService.UpdateHostname))
	mux.Handle("/console/listHostnames/v1", s.Handler(&service.ListHostnamesRequest{}, s.ConsoleService.ListHostnames))
	mux.Handle("/console/deleteHostname/v1", s.Handler(&service.DeleteHostnameRequest{}, s.ConsoleService.DeleteHostname))
	// proto statistic
	mux.Handle("/console/listProtoStatistics/v1", s.Handler(&service.ListProtoStatisticsRequest{}, s.ConsoleService.ListProtoStatistics))
	// cron jobs
	mux.Handle("/console/listCronJobs/v1", s.Handler(&service.ListCronJobsRequest{}, s.ConsoleService.ListCronJobs))
	// alarm email
	mux.Handle("/console/listAlarmEmails/v1", s.Handler(&service.ListAlarmEmailsRequest{}, s.ConsoleService.ListAlarmEmails))
	mux.Handle("/console/addAlarmEmail/v1", s.Handler(&service.AddAlarmEmailRequest{}, s.ConsoleService.AddAlarmEmail))
	mux.Handle("/console/deleteAlarmEmail/v1", s.Handler(&service.DeleteAlarmEmailRequest{}, s.ConsoleService.DeleteAlarmEmail))
	mux.Handle("/console/updateAlarmEmail/v1", s.Handler(&service.UpdateAlarmEmailRequest{}, s.ConsoleService.UpdateAlarmEmail))
	// rate limit rule
	mux.Handle("/console/listRateLimitRules/v1", s.Handler(&service.ListRateLimitRulesRequest{}, s.ConsoleService.ListRateLimitRules))
	mux.Handle("/console/addRateLimitRule/v1", s.Handler(&service.AddRateLimitRuleRequest{}, s.ConsoleService.AddRateLimitRule))
	mux.Handle("/console/deleteRateLimitRule/v1", s.Handler(&service.DeleteRateLimitRuleRequest{}, s.ConsoleService.DeleteRateLimitRule))
	mux.Handle("/console/updateRateLimitRule/v1", s.Handler(&service.UpdateRateLimitRuleRequest{}, s.ConsoleService.UpdateRateLimitRule))
	// timeout statistic
	mux.Handle("/console/listTimeoutStatistics/v1", s.Handler(&service.ListTimeoutStatisticsRequest{}, s.ConsoleService.ListTimeoutStatistics))
	// global middlewares
	// cors
	handler := cors.Default().Handler(&mux)
	// http server
	s.HttpServer = &http.Server{Addr: fmt.Sprintf("%s:%v", "0.0.0.0", s.HttpPort), Handler: handler}
}

// check proto dependencies from local disk
func (s *ConsoleServer) checkPrerequisiteProtos() error {
	ctx := context.Background()
	fsys := os.DirFS("proto/api")
	if err := fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !strings.HasPrefix(path, "google") || d.IsDir() {
			return nil
		}
		service, err := s.DaoManager.ServiceDAO.GetByName(ctx, path, true /*omitProtoFile*/, true /*omitDeleted*/)
		if err != nil {
			return err
		}
		if service == nil {
			content, err := fs.ReadFile(fsys, path)
			if err != nil {
				return err
			}
			b64fdp, err := simpleproto.Base64FileDescriptorProto(path, nil)
			if err != nil {
				return err
			}
			if err := s.DaoManager.ServiceDAO.Create(ctx, &Service{
				Name:               path,
				ApplicationID:      0,                      // special sign
				PathPrefix:         "prerequisite/" + path, // unique key
				ProtoFile:          string(content),
				ProtoFileMd5:       simplehash.HexMd5ToString(content),
				FileDescriptorData: b64fdp,
			}); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}

func (s *ConsoleServer) checkAlarmEmails() error {
	ctx := context.Background()
	alarmEmail, err := s.DaoManager.AlarmEmailDAO.GetByAddress(ctx, "luoxianmingg@gmail.com")
	if err != nil {
		return err
	}
	if alarmEmail == nil {
		if err := s.DaoManager.AlarmEmailDAO.Create(ctx, &AlarmEmail{
			Address: "luoxianmingg@gmail.com",
		}); err != nil {
			return err
		}
	}
	alarmEmail, err = s.DaoManager.AlarmEmailDAO.GetByAddress(ctx, "gavin@n1xt.net")
	if err != nil {
		return err
	}
	if alarmEmail == nil {
		if err := s.DaoManager.AlarmEmailDAO.Create(ctx, &AlarmEmail{
			Address: "gavin@n1xt.net",
		}); err != nil {
			return err
		}
	}
	return nil
}

func (s *ConsoleServer) checkHostnames() error {
	ctx := context.Background()
	apiHostname := os.Getenv("GATEWAY_API_HOSTNAME")
	hostname, err := s.DaoManager.HostManageDAO.GetByDomain(ctx, apiHostname)
	if err != nil {
		return err
	}
	rawURL := fmt.Sprintf("https://%s:443", os.Getenv("HOST_IP"))
	if hostname == nil {
		if err := s.DaoManager.HostManageDAO.Create(ctx, &HostManage{
			Domain: apiHostname,
			RawURL: rawURL,
		}); err != nil {
			return err
		}
	}
	if hostname.RawURL != rawURL {
		if err := s.DaoManager.HostManageDAO.Update(ctx, &HostManage{ID: hostname.ID, RawURL: rawURL}); err != nil {
			return err
		}
	}
	return nil
}

// http handlers ----------------------------------------------------------------------------------------

func (s *ConsoleServer) Ping(w http.ResponseWriter, r *http.Request) {
	r = response.WithLocalizer(r, s.LocalizeManager)
	ctx := r.Context()
	var resp = &response.Response{Req: r}
	defer func() {
		response.DeferWriteResponse(ctx, w, resp)
	}()
	if !response.MustMethodPost(r, resp) {
		return
	}
	resp.Code = response.StatusCodeOK
	resp.Message = response.StatusMessageOk
	resp.Data = "pong"
}

func (s *ConsoleServer) Handler(req any, service func(ctx context.Context, req any) (any, error)) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r = response.WithLocalizer(r, s.LocalizeManager)
		ctx := r.Context()
		var resp = &response.Response{Req: r}
		defer func() {
			response.DeferWriteResponse(ctx, w, resp)
		}()
		if !response.MustMethodPost(r, resp) {
			return
		}
		if err := simplehttp.UnmarshalRequestBody(r, req); err != nil {
			resp.Code = response.StatusCodeWrongParameters
			resp.Message = response.LocalizeMessage(ctx, "ClientErrMsg_WrongParameters")
			resp.DebugMessage = fmt.Sprintf("failed to unmarshal request: %v\n", err)
			return
		}
		if err := s.Validator.Struct(req); err != nil {
			resp.Code = response.StatusCodeWrongParameters
			resp.Message = response.LocalizeMessage(ctx, "ClientErrMsg_WrongParameters")
			resp.DebugMessage = fmt.Sprintf("failed to validate request: %v\n", err)
			return
		}
		data, err := service(ctx, req)
		if err != nil {
			resp.Code = response.StatusCodeInternalServerError
			resp.Message = response.LocalizeMessage(ctx, "FatalErrMsg")
			resp.DebugMessage = fmt.Sprintf("internal error: %v\n", err)
			return
		}
		resp.Code = response.StatusCodeOK
		resp.Message = response.StatusMessageOk
		if data != nil {
			resp.Data = data
		}
	})
}
