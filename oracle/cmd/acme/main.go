package main

import (
	"log"
	"time"

	"github.com/go-acme/lego/challenge/http01"
	"github.com/nextsurfer/oracle/internal/common/acme"
	. "github.com/nextsurfer/oracle/internal/model"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	_oracleMysqlDsn := "root:waf12KFkwo22@tcp(172.31.91.246:3306)/oracle?charset=utf8&parseTime=true"
	_oracleGormDB, err := gorm.Open(mysql.Open(_oracleMysqlDsn), &gorm.Config{})
	if err != nil {
		log.Fatalln(err)
	}
	domain := "rp.test.n1xt.net"
	var result AcmeResource
	if err := _oracleGormDB.Table(TableNameAcmeResource).Where("domain = ?", domain).
		Where("deleted_at = 0").
		Limit(1).
		Find(&result).Error; err != nil {
		log.Fatalln(err)
	}
	if result.ID > 0 {
		resource, err := acme.RenewAcmeResource(
			&result,
			http01.NewProviderServer("", "80"),
		)
		if err != nil {
			log.Fatalln(err)
		}
		log.Printf("%#v\n", resource)
		result.CertURL = resource.CertURL
		result.CertStableURL = resource.CertStableURL
		result.PrivateKey = string(resource.PrivateKey)
		result.Certificate = string(resource.Certificate)
		result.IssuerCertificate = string(resource.IssuerCertificate)
		result.Csr = string(resource.CSR)
		result.UpdatedAt = time.Now()
		// update the record in db
		if err := _oracleGormDB.Table(TableNameAcmeResource).
			Where("id = ?", result.ID).
			Where("deleted_at = 0").
			Updates(result).Error; err != nil {
			log.Fatalln(err)
		}
	}
}
