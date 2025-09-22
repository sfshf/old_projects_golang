package keystore

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/nextsurfer/connector/internal/pkg/dao"
	. "github.com/nextsurfer/connector/internal/pkg/model"
	"github.com/nextsurfer/connector/internal/pkg/simplehttp"
	"github.com/nextsurfer/connector/internal/pkg/util"
	"go.uber.org/zap"
)

type CheckKeyExistingRequest struct {
	KeyID string `json:"keyID"`
}

type CheckKeyExistingResponse struct {
	Code         int    `json:"code"`
	Message      string `json:"message"`
	DebugMessage string `json:"debugMessage"`
	Data         struct {
		Existing bool `json:"existing"`
	} `json:"data"`
}

func CheckKeyExisting(keyID string) (CheckKeyExistingResponse, error) {
	reqData := CheckKeyExistingRequest{KeyID: keyID}
	respData := CheckKeyExistingResponse{}
	resp, err := simplehttp.PostJsonRequest(fmt.Sprintf("http://%s/keystore/checkKeyExisting/v1", util.ConfigInfo().KeyStore), &reqData, nil, &respData)
	if err != nil {
		return respData, err
	}
	if resp.StatusCode != http.StatusOK {
		return respData, errors.New("http request status code not equal to 200")
	}
	if respData.Code != 0 {
		return respData, errors.New("response code not equal to 0")
	}
	return respData, nil
}

type CreatePasswordRequest struct {
	KeyID string `json:"keyID"`
}

type CreatePasswordResponse struct {
	Code         int    `json:"code"`
	Message      string `json:"message"`
	DebugMessage string `json:"debugMessage"`
	Data         struct {
		Password string `json:"password"`
	} `json:"data"`
}

func CreatePassword(keyID string) (CreatePasswordResponse, error) {
	reqData := CreatePasswordRequest{KeyID: keyID}
	respData := CreatePasswordResponse{}
	resp, err := simplehttp.PostJsonRequest(fmt.Sprintf("http://%s/keystore/createPassword/v1", util.ConfigInfo().KeyStore), &reqData, nil, &respData)
	if err != nil {
		return respData, err
	}
	if resp.StatusCode != http.StatusOK {
		return respData, errors.New("http request status code not equal to 200")
	}
	if respData.Code != 0 {
		return respData, errors.New("response code not equal to 0")
	}
	return respData, nil
}

type DeletePasswordRequest struct {
	KeyID string `json:"keyID"`
}

type DeletePasswordResponse struct {
	Code         int    `json:"code"`
	Message      string `json:"message"`
	DebugMessage string `json:"debugMessage"`
}

func DeletePassword(keyID string) (DeletePasswordResponse, error) {
	reqData := DeletePasswordRequest{KeyID: keyID}
	respData := DeletePasswordResponse{}
	resp, err := simplehttp.PostJsonRequest(fmt.Sprintf("http://%s/keystore/deletePassword/v1", util.ConfigInfo().KeyStore), &reqData, nil, &respData)
	if err != nil {
		return respData, err
	}
	if resp.StatusCode != http.StatusOK {
		return respData, errors.New("http request status code not equal to 200")
	}
	if respData.Code != 0 {
		return respData, errors.New("response code not equal to 0")
	}
	return respData, nil
}

type GetPasswordRequest struct {
	KeyID string `json:"keyID"`
}

type GetPasswordResponse struct {
	Code         int    `json:"code"`
	Message      string `json:"message"`
	DebugMessage string `json:"debugMessage"`
	Data         struct {
		Password string `json:"password"`
	} `json:"data"`
}

func GetPassword(keyID string) (GetPasswordResponse, error) {
	reqData := GetPasswordRequest{KeyID: keyID}
	respData := GetPasswordResponse{}
	resp, err := simplehttp.PostJsonRequest(fmt.Sprintf("http://%s/keystore/getPassword/v1", util.ConfigInfo().KeyStore), &reqData, nil, &respData)
	if err != nil {
		return respData, err
	}
	if resp.StatusCode != http.StatusOK {
		return respData, errors.New("http request status code not equal to 200")
	}
	if respData.Code != 0 {
		return respData, errors.New("response code not equal to 0")
	}
	return respData, nil
}

type CheckPasswordRequest struct {
	KeyID        string `json:"keyID"`
	PasswordHash string `json:"passwordHash"`
}

type CheckPasswordResponse struct {
	Code         int    `json:"code"`
	Message      string `json:"message"`
	DebugMessage string `json:"debugMessage"`
	Data         struct {
		Valid bool `json:"valid"`
	} `json:"data"`
}

func CheckPassword(keyID, passwordHash string) (CheckPasswordResponse, error) {
	reqData := CheckPasswordRequest{KeyID: keyID, PasswordHash: passwordHash}
	respData := CheckPasswordResponse{}
	resp, err := simplehttp.PostJsonRequest(fmt.Sprintf("http://%s/keystore/checkPassword/v1", util.ConfigInfo().KeyStore), &reqData, nil, &respData)
	if err != nil {
		return respData, err
	}
	if resp.StatusCode != http.StatusOK {
		return respData, errors.New("http request status code not equal to 200")
	}
	if respData.Code != 0 {
		return respData, errors.New("response code not equal to 0")
	}
	return respData, nil
}

type CreatePrivateKeyRequest struct {
	KeyID string `json:"keyID"`
}

type CreatePrivateKeyResponse struct {
	Code         int    `json:"code"`
	Message      string `json:"message"`
	DebugMessage string `json:"debugMessage"`
	Data         struct {
		PublicKey string `json:"publicKey"`
	} `json:"data"`
}

func CreatePrivateKey(keyID string) (CreatePrivateKeyResponse, error) {
	reqData := CreatePrivateKeyRequest{KeyID: keyID}
	respData := CreatePrivateKeyResponse{}
	resp, err := simplehttp.PostJsonRequest(fmt.Sprintf("http://%s/keystore/createPrivateKey/v1", util.ConfigInfo().KeyStore), &reqData, nil, &respData)
	if err != nil {
		return respData, err
	}
	if resp.StatusCode != http.StatusOK {
		return respData, errors.New("http request status code not equal to 200")
	}
	if respData.Code != 0 {
		return respData, errors.New("response code not equal to 0")
	}
	return respData, nil
}

type GetPublicKeyRequest struct {
	KeyID string `json:"keyID"`
}

type GetPublicKeyResponse struct {
	Code         int    `json:"code"`
	Message      string `json:"message"`
	DebugMessage string `json:"debugMessage"`
	Data         struct {
		PublicKey string `json:"publicKey"`
	} `json:"data"`
}

func GetPublicKey(keyID string) (GetPublicKeyResponse, error) {
	reqData := GetPublicKeyRequest{KeyID: keyID}
	respData := GetPublicKeyResponse{}
	resp, err := simplehttp.PostJsonRequest(fmt.Sprintf("http://%s/keystore/getPublicKey/v1", util.ConfigInfo().KeyStore), &reqData, nil, &respData)
	if err != nil {
		return respData, err
	}
	if resp.StatusCode != http.StatusOK {
		return respData, errors.New("http request status code not equal to 200")
	}
	if respData.Code != 0 {
		return respData, errors.New("response code not equal to 0")
	}
	return respData, nil
}

type DeletePrivateKeyRequest struct {
	KeyID string `json:"keyID"`
}

type DeletePrivateKeyResponse struct {
	Code         int    `json:"code"`
	Message      string `json:"message"`
	DebugMessage string `json:"debugMessage"`
}

func DeletePrivateKey(keyID string) (DeletePrivateKeyResponse, error) {
	reqData := DeletePrivateKeyRequest{KeyID: keyID}
	respData := DeletePrivateKeyResponse{}
	resp, err := simplehttp.PostJsonRequest(fmt.Sprintf("http://%s/keystore/deletePrivateKey/v1", util.ConfigInfo().KeyStore), &reqData, nil, &respData)
	if err != nil {
		return respData, err
	}
	if resp.StatusCode != http.StatusOK {
		return respData, errors.New("http request status code not equal to 200")
	}
	if respData.Code != 0 {
		return respData, errors.New("response code not equal to 0")
	}
	return respData, nil
}

type SaveDataRequest struct {
	KeyID              string `json:"keyID"`
	DataID             string `json:"dataID"`
	ReplaceCurrentItem bool   `json:"replaceCurrentItem"`
	Data               string `json:"data"`
	PlaintextHash      string `json:"plaintextHash"`
}

type SaveDataResponse struct {
	Code         int    `json:"code"`
	Message      string `json:"message"`
	DebugMessage string `json:"debugMessage"`
}

func SaveData(keyID, dataID, data, plaintextHash string, replace bool) (SaveDataResponse, error) {
	reqData := SaveDataRequest{
		KeyID:              keyID,
		DataID:             dataID,
		Data:               data,
		PlaintextHash:      plaintextHash,
		ReplaceCurrentItem: replace,
	}
	respData := SaveDataResponse{}
	resp, err := simplehttp.PostJsonRequest(fmt.Sprintf("http://%s/keystore/saveData/v1", util.ConfigInfo().KeyStore), &reqData, nil, &respData)
	if err != nil {
		return respData, err
	}
	if resp.StatusCode != http.StatusOK {
		return respData, errors.New("http request status code not equal to 200")
	}
	if respData.Code != 0 {
		return respData, errors.New("response code not equal to 0")
	}
	return respData, nil
}

type DeleteDataRequest struct {
	KeyID  string `json:"keyID"`
	DataID string `json:"dataID"`
}

type DeleteDataResponse struct {
	Code         int    `json:"code"`
	Message      string `json:"message"`
	DebugMessage string `json:"debugMessage"`
}

func DeleteData(keyID, dataID string) (DeleteDataResponse, error) {
	reqData := DeleteDataRequest{
		KeyID:  keyID,
		DataID: dataID,
	}
	respData := DeleteDataResponse{}
	resp, err := simplehttp.PostJsonRequest(fmt.Sprintf("http://%s/keystore/deleteData/v1", util.ConfigInfo().KeyStore), &reqData, nil, &respData)
	if err != nil {
		return respData, err
	}
	if resp.StatusCode != http.StatusOK {
		return respData, errors.New("http request status code not equal to 200")
	}
	if respData.Code != 0 {
		return respData, errors.New("response code not equal to 0")
	}
	return respData, nil
}

type DecryptDataRequest struct {
	KeyID string `json:"keyID"`
	Data  string `json:"data"`
}

type DecryptDataResponse struct {
	Code         int    `json:"code"`
	Message      string `json:"message"`
	DebugMessage string `json:"debugMessage"`
	Data         struct {
		DecrypedData string `json:"decrypedData"`
	} `json:"data"`
}

func DecryptData(keyID, data string) (DecryptDataResponse, error) {
	reqData := DecryptDataRequest{
		KeyID: keyID,
		Data:  data,
	}
	respData := DecryptDataResponse{}
	resp, err := simplehttp.PostJsonRequest(fmt.Sprintf("http://%s/keystore/decryptData/v1", util.ConfigInfo().KeyStore), &reqData, nil, &respData)
	if err != nil {
		return respData, err
	}
	if resp.StatusCode != http.StatusOK {
		return respData, errors.New("http request status code not equal to 200")
	}
	if respData.Code != 0 {
		return respData, errors.New("response code not equal to 0")
	}
	return respData, nil
}

type GetDataRequest struct {
	KeyID  string `json:"keyID"`
	DataID string `json:"dataID"`
}

type GetDataResponse struct {
	Code         int    `json:"code"`
	Message      string `json:"message"`
	DebugMessage string `json:"debugMessage"`
	Data         struct {
		Data string `json:"data"`
	} `json:"data"`
}

func GetData(keyID, dataID string) (GetDataResponse, error) {
	reqData := GetDataRequest{
		KeyID:  keyID,
		DataID: dataID,
	}
	respData := GetDataResponse{}
	resp, err := simplehttp.PostJsonRequest(fmt.Sprintf("http://%s/keystore/getData/v1", util.ConfigInfo().KeyStore), &reqData, nil, &respData)
	if err != nil {
		return respData, err
	}
	if resp.StatusCode != http.StatusOK {
		return respData, errors.New("http request status code not equal to 200")
	}
	if respData.Code != 0 {
		return respData, errors.New("response code not equal to 0")
	}
	return respData, nil
}

type GetLogsRequest struct {
	PageNumber int `json:"pageNumber"`
	PageSize   int `json:"pageSize"`
}

type GetLogsResponse struct {
	Code         int    `json:"code"`
	Message      string `json:"message"`
	DebugMessage string `json:"debugMessage"`
	Data         struct {
		Total int      `json:"total"`
		List  []string `json:"list"`
	} `json:"data"`
}

func GetLogs(pageNumber, pageSize int) (GetLogsResponse, error) {
	reqData := GetLogsRequest{
		PageNumber: pageNumber,
		PageSize:   pageSize,
	}
	respData := GetLogsResponse{}
	resp, err := simplehttp.PostJsonRequest(fmt.Sprintf("http://%s/keystore/getLogs/v1", util.ConfigInfo().KeyStore), &reqData, nil, &respData)
	if err != nil {
		return respData, err
	}
	if resp.StatusCode != http.StatusOK {
		return respData, errors.New("http request status code not equal to 200")
	}
	if respData.Code != 0 {
		return respData, errors.New("response code not equal to 0")
	}
	return respData, nil
}

type GetMonitorInfosRequest struct {
}

type GetMonitorInfosResponse struct {
	Code         int    `json:"code"`
	Message      string `json:"message"`
	DebugMessage string `json:"debugMessage"`
	Data         struct {
		Infos []struct {
			Name  string `json:"name"`
			Value string `json:"value"`
		} `json:"infos"`
	} `json:"data"`
}

func GetMonitorInfos() (GetMonitorInfosResponse, error) {
	reqData := GetMonitorInfosRequest{}
	respData := GetMonitorInfosResponse{}
	resp, err := simplehttp.PostJsonRequest(fmt.Sprintf("http://%s/keystore/getMonitorInfos/v1", util.ConfigInfo().KeyStore), &reqData, nil, &respData)
	if err != nil {
		return respData, err
	}
	if resp.StatusCode != http.StatusOK {
		return respData, errors.New("http request status code not equal to 200")
	}
	if respData.Code != 0 {
		return respData, errors.New("response code not equal to 0")
	}
	return respData, nil
}

const (
	ConnectorPasswordKeyID = "pswd_connector"
	AdminAppName           = "admin"
	AdminApiKeyName        = "riki"
)

func UpsertConnectorPassword(daoManager *dao.Manager, logger *zap.Logger) error {
	// check whether pswd_connector exists
	ctx := context.Background()
	appConfig, err := daoManager.AppConfigDAO.GetByApp(ctx, AdminAppName)
	if err != nil {
		return err
	}
	if appConfig == nil {
		if err := daoManager.AppConfigDAO.Create(ctx, &AppConfig{
			App:    AdminAppName,
			Config: `{"appName": "admin"}`,
		}); err != nil {
			return err
		}
	}
	var relationAppKey *RelationAppKey
	var newRelation bool
	relationAppKey, err = daoManager.RelationAppKeyDAO.GetByAppWithKeyID(ctx, AdminAppName, ConnectorPasswordKeyID)
	if err != nil {
		return err
	}
	if relationAppKey == nil {
		relationAppKey = &RelationAppKey{
			App:   AdminAppName,
			KeyID: ConnectorPasswordKeyID,
		}
		if err := daoManager.RelationAppKeyDAO.Create(ctx, relationAppKey); err != nil {
			return err
		}
	}
	if relationAppKey.PasswordHash == "" {
		newRelation = true
	}
	apiKey, err := daoManager.ApiKeyDAO.GetByName(ctx, AdminApiKeyName)
	if err != nil {
		return err
	}
	if apiKey == nil {
		if err := daoManager.ApiKeyDAO.Create(ctx, &APIKey{
			App:        AdminAppName,
			KeyID:      ConnectorPasswordKeyID,
			Name:       AdminApiKeyName,
			Permission: util.PermWrite,
		}); err != nil {
			return err
		}
	}
	checkRespData, err := CheckKeyExisting(ConnectorPasswordKeyID)
	if err != nil {
		return err
	}
	// check whether to update, if pswd_connector exists
	if checkRespData.Data.Existing {
		update, err := strconv.ParseBool(strings.TrimSpace(os.Getenv("UPDATE_CONNECTOR_PASSWORD")))
		if err != nil {
			return err
		}
		// get password, if no need to update
		if !update {
			getRespData, err := GetPassword(ConnectorPasswordKeyID)
			if err != nil {
				return err
			}
			if newRelation {
				passwordHash, err := util.Keccak256Hex([]byte(getRespData.Data.Password))
				if err != nil {
					return err
				}
				if err = daoManager.RelationAppKeyDAO.UpdatePasswordHashByID(ctx, relationAppKey.ID, string(passwordHash)); err != nil {
					return err
				}
			}
			logger.Info("connector password",
				zap.String("keyID", ConnectorPasswordKeyID),
				zap.String("password", getRespData.Data.Password),
			)
			return nil
		} else {
			newRelation = true
		}
		// delete password, if need to update
		deleteRespData, err := DeletePassword(ConnectorPasswordKeyID)
		if err != nil {
			return err
		}
		if deleteRespData.Code != 0 {
			return errors.New("delete connector password: response code not equal to 0")
		}
	}
	// create one
	createRespData, err := CreatePassword(ConnectorPasswordKeyID)
	if err != nil {
		return err
	}
	if createRespData.Code != 0 {
		return errors.New("create connector password: response code not equal to 0")
	}
	if newRelation {
		passwordHash, err := util.Keccak256Hex([]byte(createRespData.Data.Password))
		if err != nil {
			return err
		}
		if err = daoManager.RelationAppKeyDAO.UpdatePasswordHashByID(ctx, relationAppKey.ID, string(passwordHash)); err != nil {
			return err
		}
	}
	logger.Info("connector password",
		zap.String("keyID", ConnectorPasswordKeyID),
		zap.String("password", createRespData.Data.Password),
	)
	return nil
}
