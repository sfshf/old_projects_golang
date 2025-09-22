package service

import (
	"context"
	"fmt"

	"github.com/go-acme/lego/certificate"
	"github.com/nextsurfer/ground/pkg/rpc"
	"github.com/nextsurfer/oracle/internal/common/acme"
	"github.com/nextsurfer/oracle/internal/common/connector"
	. "github.com/nextsurfer/oracle/internal/model"
	"go.uber.org/zap"
	"google.golang.org/grpc/metadata"
)

type AcmeResourceService struct {
	*ConsoleService
}

func NewAcmeResourceService(ctx context.Context, consoleService *ConsoleService) *AcmeResourceService {
	return &AcmeResourceService{
		ConsoleService: consoleService,
	}
}

type ListAcmeResourcesItem struct {
	Domain              string `json:"domain"`
	ExpectedRefreshedAt int64  `json:"expectedRefreshedAt"`
	CreatedAt           int64  `json:"createdAt"`
	UpdatedAt           int64  `json:"updatedAt"`
}

type ListAcmeResourcesData struct {
	List []ListAcmeResourcesItem `json:"list"`
}

type ListAcmeResourcesRequest struct {
	ApiKey string `json:"apiKey" validate:"required"`
}

func (s *AcmeResourceService) ListAcmeResources(ctx context.Context, request any) (any, error) {
	req := request.(*ListAcmeResourcesRequest)
	// validate api key
	if err := connector.ValidateApiKey(ctx, rpc.NewContext(metadata.NewIncomingContext(ctx, metadata.MD{}), s.LocalizeManager), s.AppID, req.ApiKey, connector.RoleRead); err != nil {
		s.Logger.Error("bad request", zap.NamedError("appError", err))
		return nil, err
	}
	resources, err := s.DaoManager.AcmeResourceDAO.GetAll(ctx)
	if err != nil {
		return nil, err
	}
	hostnames, err := s.DaoManager.HostManageDAO.GetAll(ctx)
	if err != nil {
		return nil, err
	}
	var list []ListAcmeResourcesItem
	for _, hostname := range hostnames {
		one := ListAcmeResourcesItem{
			Domain: hostname.Domain,
		}
		for _, resource := range resources {
			if hostname.Domain == resource.Domain {
				one.ExpectedRefreshedAt = resource.UpdatedAt.AddDate(0, 0, 90-15).UnixMilli()
				one.CreatedAt = resource.CreatedAt.UnixMilli()
				one.UpdatedAt = resource.UpdatedAt.UnixMilli()
				break
			}
		}
		list = append(list, one)
	}
	return &ListAcmeResourcesData{List: list}, nil
}

type RenewAcmeResourceRequest struct {
	ApiKey string `json:"apiKey" validate:"required"`
	Domain string `json:"domain" validate:"required"`
}

func (s *AcmeResourceService) RenewAcmeResource(ctx context.Context, request any) (any, error) {
	req := request.(*RenewAcmeResourceRequest)
	// validate api key
	if err := connector.ValidateApiKey(ctx, rpc.NewContext(metadata.NewIncomingContext(ctx, metadata.MD{}), s.LocalizeManager), s.AppID, req.ApiKey, connector.RoleWrite); err != nil {
		s.Logger.Error("bad request", zap.NamedError("appError", err))
		return nil, err
	}
	hostname, err := s.DaoManager.HostManageDAO.GetByDomain(ctx, req.Domain)
	if err != nil {
		return nil, err
	}
	if hostname == nil {
		return nil, fmt.Errorf("hostname not found by domain %s", req.Domain)
	}
	acmeResource, err := s.DaoManager.AcmeResourceDAO.GetByDomain(ctx, req.Domain)
	if err != nil {
		return nil, err
	}
	var resource *certificate.Resource
	http01Provider := acme.NewHttp01Provider(s.DaoManager) // !!! important, it has dao operations
	if acmeResource == nil {
		resource, err = acme.NewAcmeResource(req.Domain, http01Provider)
		if err != nil {
			return nil, fmt.Errorf("NewAcmeResource tls certificate of domain [%s] failed: %s", req.Domain, err)
		}
	} else {
		resource, err = acme.RenewAcmeResource(acmeResource, http01Provider)
		if err != nil {
			return nil, fmt.Errorf("RenewAcmeResource tls certificate of domain [%s] failed: %s", acmeResource.Domain, err)
		}
	}
	acmeResource, err = s.DaoManager.AcmeResourceDAO.GetByDomain(ctx, req.Domain)
	if err != nil {
		return nil, err
	}
	if acmeResource == nil {
		acmeResource = &AcmeResource{
			Domain:            req.Domain,
			CertURL:           resource.CertURL,
			CertStableURL:     resource.CertStableURL,
			PrivateKey:        string(resource.PrivateKey),
			Certificate:       string(resource.Certificate),
			IssuerCertificate: string(resource.IssuerCertificate),
			Csr:               string(resource.CSR),
		}
		// insert the record in db
		if err := s.DaoManager.AcmeResourceDAO.Create(ctx, acmeResource); err != nil {
			return nil, fmt.Errorf("create tls certificate of domain [%s] to db failed: %s", acmeResource.Domain, err)
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
			return nil, fmt.Errorf("update tls certificate of domain [%s] to db failed: %s", acmeResource.Domain, err)
		}
	}
	// notify gateways to RefreshCertificate
	return nil, s.notifyRefreshCertificate(ctx, acmeResource.Domain)
}
