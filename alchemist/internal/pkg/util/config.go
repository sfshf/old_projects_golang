package util

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/klaytn/klaytn/crypto/sha3"
	"github.com/nextsurfer/alchemist/internal/pkg/dao"
	connector_grpc "github.com/nextsurfer/connector/pkg/grpc"
	"github.com/nextsurfer/ground/pkg/rpc"
)

type configInfo struct {
	mu         sync.Mutex
	AppConfigs []appConfig `json:"appConfigs,omitempty"`
}

type appConfig struct {
	ProductId                  string        `json:"productId,omitempty"`
	AppID                      string        `json:"appID,omitempty"`
	BindReferralCodeExpiration int           `json:"bindReferralCodeExpiration,omitempty"`
	IgnoreDeviceCheck          bool          `json:"ignoreDeviceCheck,omitempty"`
	DeviceCheck                DeviceCheck   `json:"deviceCheck,omitempty"`
	DiscountOffer              DiscountOffer `json:"discountOffer,omitempty"`
	PromoOfferKeyID            string        `json:"promoOfferKeyID,omitempty"`
	PromoOfferPrivKeyPem       string        `json:"promoOfferPrivKeyPem,omitempty"`
	RewardList                 []PromoReward `json:"rewardList,omitempty"`
}

type DeviceCheck struct {
	KeyID      string `json:"keyID,omitempty"`
	IssuerID   string `json:"issuerID,omitempty"`
	PrivKeyPem string `json:"privKeyPem,omitempty"`
}

type DiscountOffer struct {
	IDNewUser string `json:"idNewUser,omitempty"`
	ID10M     string `json:"id10M,omitempty"`
	ID8M      string `json:"id8M,omitempty"`
	ID6M      string `json:"id6M,omitempty"`
	ID4M      string `json:"id4M,omitempty"`
	ID2M      string `json:"id2M,omitempty"`
}

type PromoReward struct {
	ID             string `json:"id,omitempty"`
	Name           string `json:"name,omitempty"`
	Description    string `json:"description,omitempty"`
	OfferID        string `json:"offerID,omitempty"`
	Cost           int32  `json:"cost,omitempty"`
	Duration       string `json:"duration,omitempty"`
	DurationInDays int32  `json:"durationInDays,omitempty"`
}

var (
	_configInfo = &configInfo{}
)

func RefreshConfig(daoManager *dao.Manager) error {
	_configInfo.mu.Lock()
	defer _configInfo.mu.Unlock()
	configList, err := daoManager.AppConfigDAO.GetAll(context.Background())
	if err != nil {
		return err
	}
	var appConfigs []appConfig
	for _, item := range configList {
		var appConfig appConfig
		if err := json.NewDecoder(strings.NewReader(item.Config)).Decode(&appConfig); err != nil {
			return err
		}
		if appConfig.AppID != item.App {
			return fmt.Errorf("app config [id=%d], app id [%s] not equal to the appID [%s] in config", item.ID, item.App, appConfig.AppID)
		}
		appConfigs = append(appConfigs, appConfig)
	}
	_configInfo.AppConfigs = appConfigs
	return nil
}

func AppConfig(appID string) appConfig {
	for _, appConfig := range _configInfo.AppConfigs {
		if appConfig.AppID == appID {
			return appConfig
		}
	}
	return appConfig{}
}

func Keccak256(src []byte) ([]byte, error) {
	h := sha3.NewKeccak256()
	if _, err := h.Write(src); err != nil {
		return nil, err
	}
	return h.Sum(nil), nil
}

func Keccak256Hex(src []byte) ([]byte, error) {
	sum, err := Keccak256(src)
	if err != nil {
		return nil, err
	}
	dst := make([]byte, hex.EncodedLen(len(sum)))
	hex.Encode(dst, sum)
	return dst, nil
}

const (
	RoleWrite = "write"
	RoleRead  = "read"
)

var (
	ErrInvalidApikey = errors.New("invalid api key")
)

func ValidateApiKey(ctx context.Context, rpcCtx *rpc.Context, app, apikey, role string) error {
	passwordHash, err := Keccak256Hex([]byte(apikey))
	if err != nil {
		return err
	}
	exist, err := connector_grpc.ValidateApiKey(ctx, rpcCtx, app, string(passwordHash), role)
	if err != nil {
		return err
	}
	if !exist {
		return ErrInvalidApikey
	}
	return nil
}

func CheckConfigFormat(config string) (string, string, error) {
	var appConfig appConfig
	if err := json.NewDecoder(strings.NewReader(config)).Decode(&appConfig); err != nil {
		return "", config, err
	}
	marshaled, err := json.Marshal(appConfig)
	if err != nil {
		return "", config, err
	}
	return appConfig.AppID, string(marshaled), nil
}

var ErrorInvalidRewardId = errors.New("invalid reward id")

func Reward(app, id string) (*PromoReward, error) {
	for _, reward := range AppConfig(app).RewardList {
		if reward.ID == id {
			return &reward, nil
		}
	}
	return nil, ErrorInvalidRewardId
}
