package util

import (
	"context"
	"fmt"
	"sync"

	"github.com/nextsurfer/connector/internal/pkg/dao"
)

type AppKey struct {
	App        string `json:"app"`
	KeyID      string `json:"keyID"`
	Name       string `json:"name"`
	Permission string `json:"permission"`
	ApiKey     string `json:"apiKey"` // password hash
}

type configInfo struct {
	mu       sync.Mutex
	AppKeys  []AppKey `json:"appKeys,omitempty"`
	KeyStore string   `json:"keystore,omitempty"`
}

var (
	_configInfo = &configInfo{}
)

func ConfigInfo() *configInfo {
	return _configInfo
}

func InitConfig(keyStoreIp string) {
	_configInfo.KeyStore = keyStoreIp
}

func LoadAllAppKeys(daoManager *dao.Manager) error {
	_configInfo.mu.Lock()
	defer _configInfo.mu.Unlock()
	apiKeyList, err := daoManager.ApiKeyDAO.GetAll(context.Background())
	if err != nil {
		return err
	}
	var appKeys []AppKey
	for _, item := range apiKeyList {
		relationAppKey, err := daoManager.RelationAppKeyDAO.GetByKeyID(context.Background(), item.KeyID)
		if err != nil {
			return err
		}
		if relationAppKey == nil {
			return fmt.Errorf("api key [id=%d] has no relation_app_key record", item.ID)
		}
		apiKey := AppKey{
			App:        item.App,
			KeyID:      item.KeyID,
			Name:       item.Name,
			Permission: item.Permission,
			ApiKey:     relationAppKey.PasswordHash,
		}
		appKeys = append(appKeys, apiKey)
	}

	_configInfo.AppKeys = appKeys
	return nil
}

func AddAppKey(app, keyID, name, perm, apiKey string) {
	_configInfo.mu.Lock()
	defer _configInfo.mu.Unlock()
	_configInfo.AppKeys = append(_configInfo.AppKeys, AppKey{
		App:        app,
		KeyID:      keyID,
		Name:       name,
		Permission: perm,
		ApiKey:     apiKey,
	})
}

func RemoveAppKey(name string) {
	_configInfo.mu.Lock()
	defer _configInfo.mu.Unlock()
	appKeys := make([]AppKey, 0, len(_configInfo.AppKeys)-1)
	for _, item := range _configInfo.AppKeys {
		if item.Name != name {
			appKeys = append(appKeys, item)
		}
	}
	_configInfo.AppKeys = appKeys
}

func ValidateApiKey(pswd string) (AppKey, bool) {
	pswdHash, _ := Keccak256Hex([]byte(pswd))
	pswd = string(pswdHash)
	var maybeAdmin AppKey
	var valid bool
	fmt.Println("Validating API key:", pswd)
	for _, appKey := range _configInfo.AppKeys {
		fmt.Println("appKey:", appKey)
		if appKey.ApiKey == pswd {
			if !valid {
				maybeAdmin = appKey
				valid = true
			} else if appKey.App == AdminApp {
				maybeAdmin = appKey
			}
		}
	}
	return maybeAdmin, valid
}

const (
	PermWrite = "write"
	PermRead  = "read"

	AdminApp = "admin"
)

func CheckPerm(pswd, perm string) (AppKey, bool) {
	appKey, valid := ValidateApiKey(pswd)
	if !valid {
		return appKey, valid
	}
	return appKey, appKey.Permission == PermWrite || appKey.Permission == perm
}

func IsAdminApp(ak AppKey) bool {
	return ak.App == AdminApp
}
