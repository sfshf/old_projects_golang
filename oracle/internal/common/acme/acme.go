package acme

import (
	"context"
	"crypto"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"net"
	"strings"

	"github.com/go-acme/lego/certcrypto"
	"github.com/go-acme/lego/certificate"
	"github.com/go-acme/lego/challenge"
	"github.com/go-acme/lego/lego"
	"github.com/go-acme/lego/registration"
	"github.com/nextsurfer/oracle/internal/dao"
	"github.com/nextsurfer/oracle/internal/model"
)

// You'll need a user or account type that implements acme.User
type MyUser struct {
	Email        string
	Registration *registration.Resource
	key          crypto.PrivateKey
}

func (u *MyUser) GetEmail() string {
	return u.Email
}
func (u MyUser) GetRegistration() *registration.Resource {
	return u.Registration
}
func (u *MyUser) GetPrivateKey() crypto.PrivateKey {
	return u.key
}

func NewAcmeResource(domain string, http01Provider challenge.Provider) (*certificate.Resource, error) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}
	myUser := MyUser{
		Email: "luoxianmingg@gmail.com",
		key:   privateKey,
	}
	config := lego.NewConfig(&myUser)
	config.Certificate.KeyType = certcrypto.RSA2048
	client, err := lego.NewClient(config)
	if err != nil {
		return nil, err
	}
	err = client.Challenge.SetHTTP01Provider(http01Provider)
	if err != nil {
		return nil, err
	}
	reg, err := client.Registration.Register(registration.RegisterOptions{TermsOfServiceAgreed: true})
	if err != nil {
		return nil, err
	}
	myUser.Registration = reg
	request := certificate.ObtainRequest{
		Domains: []string{domain},
		Bundle:  false,
	}
	certificates, err := client.Certificate.Obtain(request)
	if err != nil {
		return nil, err
	}
	return certificates, nil
}

func ExtractDomains(cert *x509.Certificate) []string {
	var domains []string
	if cert.Subject.CommonName != "" {
		domains = append(domains, cert.Subject.CommonName)
	}

	// Check for SAN certificate
	for _, sanDomain := range cert.DNSNames {
		if sanDomain == cert.Subject.CommonName {
			continue
		}
		domains = append(domains, sanDomain)
	}

	commonNameIP := net.ParseIP(cert.Subject.CommonName)
	for _, sanIP := range cert.IPAddresses {
		if !commonNameIP.Equal(sanIP) {
			domains = append(domains, sanIP.String())
		}
	}

	return domains
}

func ParsePEMBundle(bundle []byte) ([]*x509.Certificate, error) {
	var certificates []*x509.Certificate
	var certDERBlock *pem.Block

	for {
		certDERBlock, bundle = pem.Decode(bundle)
		if certDERBlock == nil {
			break
		}

		if certDERBlock.Type == "CERTIFICATE" {
			cert, err := x509.ParseCertificate(certDERBlock.Bytes)
			if err != nil {
				return nil, err
			}
			certificates = append(certificates, cert)
		}
	}

	if len(certificates) == 0 {
		return nil, errors.New("no certificates were found while parsing the bundle")
	}

	return certificates, nil
}

func ParsePEMPrivateKey(key []byte) (crypto.PrivateKey, error) {
	keyBlockDER, _ := pem.Decode(key)
	if keyBlockDER == nil {
		return nil, errors.New("invalid PEM block")
	}

	if keyBlockDER.Type != "PRIVATE KEY" && !strings.HasSuffix(keyBlockDER.Type, " PRIVATE KEY") {
		return nil, fmt.Errorf("unknown PEM header %q", keyBlockDER.Type)
	}

	if key, err := x509.ParsePKCS1PrivateKey(keyBlockDER.Bytes); err == nil {
		return key, nil
	}

	if key, err := x509.ParsePKCS8PrivateKey(keyBlockDER.Bytes); err == nil {
		switch key := key.(type) {
		case *rsa.PrivateKey, *ecdsa.PrivateKey, ed25519.PrivateKey:
			return key, nil
		default:
			return nil, fmt.Errorf("found unknown private key type in PKCS#8 wrapping: %T", key)
		}
	}

	if key, err := x509.ParseECPrivateKey(keyBlockDER.Bytes); err == nil {
		return key, nil
	}

	return nil, errors.New("failed to parse private key")
}

func RenewAcmeResource(acmeResource *model.AcmeResource, http01Provider challenge.Provider) (*certificate.Resource, error) {
	certificates, err := ParsePEMBundle([]byte(acmeResource.Certificate))
	if err != nil {
		return nil, err
	}
	cert := certificates[0]
	certDomains := ExtractDomains(cert)
	privateKey, err := ParsePEMPrivateKey([]byte(acmeResource.PrivateKey))
	if err != nil {
		return nil, err
	}
	userPrivateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}
	myUser := MyUser{
		Email: "luoxianmingg@gmail.com",
		key:   userPrivateKey,
	}
	config := lego.NewConfig(&myUser)
	client, err := lego.NewClient(config)
	if err != nil {
		return nil, err
	}
	err = client.Challenge.SetHTTP01Provider(http01Provider)
	if err != nil {
		return nil, err
	}
	reg, err := client.Registration.Register(registration.RegisterOptions{TermsOfServiceAgreed: true})
	if err != nil {
		return nil, err
	}
	myUser.Registration = reg
	request := certificate.ObtainRequest{
		Domains:    certDomains,
		PrivateKey: privateKey,
		Bundle:     true,
	}
	return client.Certificate.Obtain(request)
}

// http provider server -----------------------------------------------------------------------------------------------

type Http01Provider struct {
	daoManager *dao.Manager
}

func NewHttp01Provider(daoManager *dao.Manager) *Http01Provider {
	return &Http01Provider{
		daoManager: daoManager,
	}
}

func (s *Http01Provider) GetAddress() string {
	return net.JoinHostPort("", "80")
}

func (s *Http01Provider) Present(domain, token, keyAuth string) error {
	ctx := context.Background()
	// store token, keyAuth to db
	acmeResource, err := s.daoManager.AcmeResourceDAO.GetByDomain(ctx, domain)
	if err != nil {
		return err
	}
	if acmeResource == nil {
		// insert a record of the domain
		if err := s.daoManager.AcmeResourceDAO.Create(ctx, &model.AcmeResource{
			Domain:  domain,
			Token:   token,
			KeyAuth: keyAuth,
		}); err != nil {
			return err
		}
	} else {
		// update the record of the domain
		acmeResource.Token = token
		acmeResource.KeyAuth = keyAuth
		if err := s.daoManager.AcmeResourceDAO.Update(ctx, acmeResource); err != nil {
			return err
		}
	}
	return nil
}

// CleanUp closes the HTTPS server.
func (s *Http01Provider) CleanUp(domain, token, keyAuth string) error {
	return nil
}
