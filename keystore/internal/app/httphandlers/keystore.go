package httphandlers

import (
	"bufio"
	"bytes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/nextsurfer/keystore/internal/app/utils"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/syndtr/goleveldb/leveldb"
	"go.uber.org/zap"
)

type Response struct {
	Code         int         `json:"code"`
	Message      string      `json:"message"`
	DebugMessage string      `json:"debugMessage"`
	Data         interface{} `json:"data"`
}

func deferWriteResponse(w http.ResponseWriter, statusCode int, resp Response) {
	respData, err := json.Marshal(resp)
	if err != nil {
		statusCode = http.StatusInternalServerError
		resp.Code = 1
		resp.DebugMessage = err.Error()
	}
	w.WriteHeader(statusCode)
	w.Write(respData)
}

func mustMethodPost(r *http.Request, statusCode *int, resp *Response) bool {
	if r.Method != http.MethodPost {
		*statusCode = http.StatusMethodNotAllowed
		resp.Code = 1
		resp.DebugMessage = "invalid request method"
		return false
	}
	return true
}

func unmarshalRequestBody(r *http.Request, req interface{}) error {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(body, req); err != nil {
		return err
	}
	return nil
}

func getRealIP(r *http.Request) net.IP {
	var realIP string
	if realIP = r.Header.Get("x-forwarded-for"); realIP != "" {
		return net.ParseIP(realIP)
	}
	// ...
	host, _, _ := net.SplitHostPort(r.RemoteAddr)
	return net.ParseIP(host)
}

func validateClientIP(connectors []net.IP, r *http.Request) error {
	reqRealIP := getRealIP(r)
	var validIP bool
	for _, connector := range connectors {
		if connector.Equal(reqRealIP) {
			validIP = true
		}
	}
	if !validIP {
		return fmt.Errorf("the ip [%s] not in the whitelist", reqRealIP)
	}
	return nil
}

// apis --------------------------------------------------------------------------------------

func CreatePassword(logger *zap.Logger, connectors []net.IP, keyDB *leveldb.DB, aead cipher.AEAD, nonce []byte) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var (
			resp       Response
			statusCode = http.StatusOK
		)
		defer func() {
			deferWriteResponse(w, statusCode, resp)
		}()

		if !mustMethodPost(r, &statusCode, &resp) {
			logger.Error("invalid request method")
			return
		}

		// check connector
		if err := validateClientIP(connectors, r); err != nil {
			logger.Error("invalid client ip", zap.NamedError("appError", err))
			statusCode = http.StatusBadRequest
			resp.Code = 1
			resp.DebugMessage = fmt.Sprintf("validate client ip fail: %v", err)
			return
		}

		// ctx := r.Context()
		// parse request body
		var req struct {
			KeyID string `json:"keyID"`
		}
		if err := unmarshalRequestBody(r, &req); err != nil {
			logger.Error("unmarshal request body fail", zap.NamedError("appError", err))
			statusCode = http.StatusBadRequest
			resp.Code = 1
			resp.DebugMessage = fmt.Sprintf("failed to unmarshal request: %v", err)
			return
		}
		logger.Info("CreatePassword accessed", zap.String("keyID", req.KeyID))

		// validate parameters
		if !strings.HasPrefix(req.KeyID, "pswd_") {
			logger.Error("key id has no 'pswd_' prefix")
			statusCode = http.StatusBadRequest
			resp.Code = 1
			resp.DebugMessage = "key id has no 'pswd_' prefix"
			return
		}

		// check keyID
		keyID := []byte(req.KeyID)
		_, err := utils.GetStoredPasswordKey(keyDB, keyID)
		if err != nil {
			if err != leveldb.ErrNotFound {
				logger.Error("leveldb service error", zap.NamedError("appError", err))
				statusCode = http.StatusInternalServerError
				resp.Code = 1
				resp.DebugMessage = fmt.Sprintf("failed to get keyID [%s]: %v", req.KeyID, err)
				return
			}
		} else {
			logger.Error("keyID exists", zap.String("keyID", req.KeyID))
			statusCode = http.StatusBadRequest
			resp.Code = 1
			resp.DebugMessage = fmt.Sprintf("keyID [%s] exists", req.KeyID)
			return
		}

		// new a password
		password, err := utils.NewPassword()
		if err != nil {
			logger.Error("new password fail", zap.NamedError("appError", err))
			statusCode = http.StatusInternalServerError
			resp.Code = 1
			resp.DebugMessage = fmt.Sprintf("new password fail [keyID=%s]: %v", req.KeyID, err)
			return
		}
		// encrypt the password
		encryptedPassword := utils.EncryptByX([]byte(password), aead, nonce)
		// hash passwod
		passwordHash, err := utils.Keccak256Hex([]byte(password))
		if err != nil {
			logger.Error("hash password fail", zap.String("keyID", req.KeyID), zap.NamedError("appError", err))
			statusCode = http.StatusInternalServerError
			resp.Code = 1
			resp.DebugMessage = fmt.Sprintf("hash keyID [%s] password fail: %v", req.KeyID, err)
			return
		}

		// store the encrypted password
		if err := utils.PutStoredPasswordKey(keyDB, keyID, encryptedPassword, passwordHash); err != nil {
			logger.Error("failed to store keyID", zap.String("keyID", req.KeyID), zap.NamedError("appError", err))
			statusCode = http.StatusInternalServerError
			resp.Code = 1
			resp.DebugMessage = fmt.Sprintf("failed to store keyID [%s]: %v", req.KeyID, err)
			return
		}

		logger.Info("CreatePassword success")
		resp.Code = 0
		resp.Message = "ok"
		resp.Data = struct {
			Password string `json:"password"`
		}{
			Password: string(password),
		}
	}
}

func CheckPassword(logger *zap.Logger, connectors []net.IP, keyDB *leveldb.DB) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var (
			resp       Response
			statusCode = http.StatusOK
		)
		defer func() {
			deferWriteResponse(w, statusCode, resp)
		}()

		if !mustMethodPost(r, &statusCode, &resp) {
			logger.Error("invalid request method")
			return
		}

		// check connector
		if err := validateClientIP(connectors, r); err != nil {
			logger.Error("invalid client ip", zap.NamedError("appError", err))
			statusCode = http.StatusBadRequest
			resp.Code = 1
			resp.DebugMessage = fmt.Sprintf("validate client ip fail: %v", err)
			return
		}

		// ctx := r.Context()
		var req struct {
			KeyID        string `json:"keyID"`
			PasswordHash string `json:"passwordHash"`
		}
		if err := unmarshalRequestBody(r, &req); err != nil {
			logger.Error("unmarshal request body fail", zap.NamedError("appError", err))
			statusCode = http.StatusBadRequest
			resp.Code = 1
			resp.DebugMessage = fmt.Sprintf("failed to unmarshal request: %v", err)
			return
		}
		logger.Info("CheckPassword accessed", zap.String("keyID", req.KeyID))

		// validate parameters
		if !strings.HasPrefix(req.KeyID, "pswd_") {
			logger.Error("key id has no 'pswd_' prefix")
			statusCode = http.StatusBadRequest
			resp.Code = 1
			resp.DebugMessage = "key id has no 'pswd_' prefix"
			return
		}

		// get storedDataBytes from key.db
		keyID := []byte(req.KeyID)
		storedPasswordKey, err := utils.GetStoredPasswordKey(keyDB, keyID)
		if err != nil {
			logger.Error("get stored password key fail", zap.NamedError("appError", err))
			statusCode = http.StatusBadRequest
			resp.Code = 1
			resp.DebugMessage = fmt.Sprintf("get stored password key fail: %v", err)
			return
		}

		// validate the passwordHashs
		if !bytes.Equal(storedPasswordKey.PasswordHash, []byte(req.PasswordHash)) {
			logger.Error("invalid password hash", zap.String("keyID", req.KeyID), zap.String("passwordHash", req.PasswordHash))
			statusCode = http.StatusBadRequest
			resp.Code = 1
			resp.DebugMessage = fmt.Sprintf("invalid password hash [%s] with keyID [%s]", req.PasswordHash, req.KeyID)
			return
		}

		logger.Info("CheckPassword success")
		resp.Code = 0
		resp.Message = "ok"
		resp.Data = struct {
			Valid bool `json:"valid"`
		}{
			Valid: true,
		}
	})
}

func GetPassword(logger *zap.Logger, connectors []net.IP, keyDB *leveldb.DB, aead cipher.AEAD) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var (
			resp       Response
			statusCode = http.StatusOK
		)
		defer func() {
			deferWriteResponse(w, statusCode, resp)
		}()

		if !mustMethodPost(r, &statusCode, &resp) {
			logger.Error("invalid request method")
			return
		}

		// check connector
		if err := validateClientIP(connectors, r); err != nil {
			logger.Error("invalid client ip", zap.NamedError("appError", err))
			statusCode = http.StatusBadRequest
			resp.Code = 1
			resp.DebugMessage = fmt.Sprintf("validate client ip fail: %v", err)
			return
		}

		// ctx := r.Context()
		var req struct {
			KeyID string `json:"keyID"`
		}
		if err := unmarshalRequestBody(r, &req); err != nil {
			logger.Error("unmarshal request body fail", zap.NamedError("appError", err))
			statusCode = http.StatusBadRequest
			resp.Code = 1
			resp.DebugMessage = fmt.Sprintf("failed to unmarshal request: %v", err)
			return
		}
		logger.Info("GetPassword accessed", zap.String("keyID", req.KeyID))

		// validate parameters
		if !strings.HasPrefix(req.KeyID, "pswd_") {
			logger.Error("key id has no 'pswd_' prefix")
			statusCode = http.StatusBadRequest
			resp.Code = 1
			resp.DebugMessage = "key id has no 'pswd_' prefix"
			return
		}

		// get storedDataBytes from key.db
		keyID := []byte(req.KeyID)
		storedPasswordKey, err := utils.GetStoredPasswordKey(keyDB, keyID)
		if err != nil {
			logger.Error("get stored password key fail", zap.NamedError("appError", err))
			statusCode = http.StatusBadRequest
			resp.Code = 1
			resp.DebugMessage = fmt.Sprintf("get stored password key fail: %v", err)
			return
		}

		// decrypt the password
		decryptedPassword, err := utils.DecryptByX(storedPasswordKey.Password, aead)
		if err != nil {
			logger.Error("decrypt encrypted password fail", zap.String("keyID", req.KeyID))
			statusCode = http.StatusBadRequest
			resp.Code = 1
			resp.DebugMessage = fmt.Sprintf("decrypt encrypted password fail, keyID [%s]", req.KeyID)
			return
		}

		logger.Info("GetPassword success")
		resp.Code = 0
		resp.Message = "ok"
		resp.Data = struct {
			Password string `json:"password"`
		}{
			Password: string(decryptedPassword),
		}
	})
}

func DeletePassword(logger *zap.Logger, connectors []net.IP, keyDB *leveldb.DB) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var (
			resp       Response
			statusCode = http.StatusOK
		)
		defer func() {
			deferWriteResponse(w, statusCode, resp)
		}()

		if !mustMethodPost(r, &statusCode, &resp) {
			logger.Error("invalid request method")
			return
		}

		// check connector
		if err := validateClientIP(connectors, r); err != nil {
			logger.Error("invalid client ip", zap.NamedError("appError", err))
			statusCode = http.StatusBadRequest
			resp.Code = 1
			resp.DebugMessage = fmt.Sprintf("validate client ip fail: %v", err)
			return
		}

		// ctx := r.Context()
		// parse request body
		var req struct {
			KeyID string `json:"keyID"`
		}
		if err := unmarshalRequestBody(r, &req); err != nil {
			logger.Error("unmarshal request body fail", zap.NamedError("appError", err))
			statusCode = http.StatusBadRequest
			resp.Code = 1
			resp.DebugMessage = fmt.Sprintf("failed to unmarshal request: %v", err)
			return
		}
		logger.Info("DeletePassword accessed", zap.String("keyID", req.KeyID))

		// validate parameters
		if !strings.HasPrefix(req.KeyID, "pswd_") {
			logger.Error("key id has no 'pswd_' prefix")
			statusCode = http.StatusBadRequest
			resp.Code = 1
			resp.DebugMessage = "key id has no 'pswd_' prefix"
			return
		}

		// delete keyID
		if err := keyDB.Delete([]byte(req.KeyID), nil); err != nil {
			logger.Error("leveldb service error")
			statusCode = http.StatusInternalServerError
			resp.Code = 1
			resp.DebugMessage = fmt.Sprintf("failed to delete keyID [%s]: %v", req.KeyID, err)
			return
		}

		logger.Info("DeletePassword success")
		resp.Code = 0
		resp.Message = "ok"
	})
}

func CreatePrivateKey(logger *zap.Logger, connectors []net.IP, keyDB *leveldb.DB, aead cipher.AEAD, nonce []byte) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var (
			resp       Response
			statusCode = http.StatusOK
		)
		defer func() {
			deferWriteResponse(w, statusCode, resp)
		}()

		if !mustMethodPost(r, &statusCode, &resp) {
			logger.Error("invalid request method")
			return
		}

		// check connector
		if err := validateClientIP(connectors, r); err != nil {
			logger.Error("invalid client ip", zap.NamedError("appError", err))
			statusCode = http.StatusBadRequest
			resp.Code = 1
			resp.DebugMessage = fmt.Sprintf("validate client ip fail: %v", err)
			return
		}

		// ctx := r.Context()
		// parse request body
		var req struct {
			KeyID string `json:"keyID"`
		}
		if err := unmarshalRequestBody(r, &req); err != nil {
			logger.Error("unmarshal request body fail", zap.NamedError("appError", err))
			statusCode = http.StatusBadRequest
			resp.Code = 1
			resp.DebugMessage = fmt.Sprintf("failed to unmarshal request: %v", err)
			return
		}
		logger.Info("CreatePrivateKey accessed", zap.String("keyID", req.KeyID))

		// check keyID
		keyID := []byte(req.KeyID)
		_, err := utils.GetStoredPrivateKey(keyDB, keyID)
		if err != nil {
			if err != leveldb.ErrNotFound {
				logger.Error("leveldb service error", zap.NamedError("appError", err))
				statusCode = http.StatusInternalServerError
				resp.Code = 1
				resp.DebugMessage = fmt.Sprintf("failed to get keyID [%s]: %v", req.KeyID, err)
				return
			}
		} else {
			logger.Error("key exists", zap.String("keyID", req.KeyID))
			statusCode = http.StatusBadRequest
			resp.Code = 1
			resp.DebugMessage = fmt.Sprintf("keyID [%s] exists", req.KeyID)
			return
		}

		// generate private key
		privKey, err := utils.GeneratePrivateKey()
		if err != nil {
			logger.Error("generate private key fail", zap.NamedError("appError", err))
			statusCode = http.StatusInternalServerError
			resp.Code = 1
			resp.DebugMessage = fmt.Sprintf("generate private key fail: %v", err)
			return
		}

		// encrypt private key
		serializedPrivKey := utils.EncryptByX(privKey.Serialize(), aead, nonce)
		publicKey := privKey.PubKey().SerializeUncompressed()

		if err := utils.PutStoredPrivateKey(keyDB, []byte(req.KeyID), serializedPrivKey, publicKey); err != nil {
			logger.Error("failed to store keyID", zap.String("keyID", req.KeyID), zap.NamedError("appError", err))
			statusCode = http.StatusInternalServerError
			resp.Code = 1
			resp.DebugMessage = fmt.Sprintf("failed to store keyID [%s]: %v", req.KeyID, err)
			return
		}

		logger.Info("CreatePrivateKey success")
		resp.Code = 0
		resp.Message = "ok"
		resp.Data = struct {
			PublicKey string `json:"publicKey"`
		}{
			PublicKey: hex.EncodeToString(publicKey),
		}
	})
}

func GetPublicKey(logger *zap.Logger, connectors []net.IP, keyDB *leveldb.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var (
			resp       Response
			statusCode = http.StatusOK
		)
		defer func() {
			deferWriteResponse(w, statusCode, resp)
		}()

		if !mustMethodPost(r, &statusCode, &resp) {
			logger.Error("invalid request method")
			return
		}

		// check connector
		if err := validateClientIP(connectors, r); err != nil {
			logger.Error("invalid client ip", zap.NamedError("appError", err))
			statusCode = http.StatusBadRequest
			resp.Code = 1
			resp.DebugMessage = fmt.Sprintf("validate client ip fail: %v", err)
			return
		}

		// ctx := r.Context()
		var req struct {
			KeyID string `json:"keyID"`
		}
		if err := unmarshalRequestBody(r, &req); err != nil {
			logger.Error("unmarshal request body fail", zap.NamedError("appError", err))
			statusCode = http.StatusBadRequest
			resp.Code = 1
			resp.DebugMessage = fmt.Sprintf("failed to unmarshal request: %v", err)
			return
		}
		logger.Info("GetPublicKey accessed", zap.String("keyID", req.KeyID))

		// get private key from key.db
		storedPrivateKey, err := utils.GetStoredPrivateKey(keyDB, []byte(req.KeyID))
		if err != nil {
			logger.Error("get stored private key fail", zap.String("keyID", req.KeyID), zap.NamedError("appError", err))
			statusCode = http.StatusBadRequest
			resp.Code = 1
			resp.DebugMessage = fmt.Sprintf("get stored private key [keyID=%s] fail: %v", req.KeyID, err)
			return
		}

		logger.Info("GetPublicKey success")
		resp.Code = 0
		resp.Message = "ok"
		resp.Data = struct {
			PublicKey string `json:"publicKey"`
		}{
			PublicKey: hex.EncodeToString(storedPrivateKey.Public),
		}
	}
}

func DeletePrivateKey(logger *zap.Logger, connectors []net.IP, keyDB *leveldb.DB) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var (
			resp       Response
			statusCode = http.StatusOK
		)
		defer func() {
			deferWriteResponse(w, statusCode, resp)
		}()

		if !mustMethodPost(r, &statusCode, &resp) {
			logger.Error("invalid request method")
			return
		}

		// check connector
		if err := validateClientIP(connectors, r); err != nil {
			logger.Error("invalid client ip", zap.NamedError("appError", err))
			statusCode = http.StatusBadRequest
			resp.Code = 1
			resp.DebugMessage = fmt.Sprintf("validate client ip fail: %v", err)
			return
		}

		// ctx := r.Context()
		var req struct {
			KeyID string `json:"keyID"`
		}
		if err := unmarshalRequestBody(r, &req); err != nil {
			logger.Error("unmarshal request body fail", zap.NamedError("appError", err))
			statusCode = http.StatusBadRequest
			resp.Code = 1
			resp.DebugMessage = fmt.Sprintf("failed to unmarshal request: %v", err)
			return
		}
		logger.Info("DeletePrivateKey accessed", zap.String("keyID", req.KeyID))

		// delete private key from key.db
		if err := keyDB.Delete([]byte(req.KeyID), nil); err != nil {
			logger.Error("leveldb service error")
			statusCode = http.StatusInternalServerError
			resp.Code = 1
			resp.DebugMessage = fmt.Sprintf("failed to delete keyID [%s]: %v", req.KeyID, err)
			return
		}

		logger.Info("DeletePrivateKey success")
		resp.Code = 0
		resp.Message = "ok"
	})
}

type StoredData struct {
	KeyID string `json:"keyID"`
	Data  string `json:"data"`
}

func getStoredData(dataDB *leveldb.DB, dataID []byte) (*StoredData, error) {
	storedDataBytes, err := dataDB.Get(dataID, nil)
	if err != nil {
		if err != leveldb.ErrNotFound {
			return nil, fmt.Errorf("failed to get data [dataID=%s]: %v", dataID, err)
		} else {
			return nil, fmt.Errorf("data [dataID=%s] not exists", dataID)
		}
	}
	var res StoredData
	if err := json.Unmarshal(storedDataBytes, &res); err != nil {
		return nil, fmt.Errorf("unmarshal stored data bytes [dataID=%s] fail: %v", dataID, err)
	}
	return &res, nil
}

func SaveData(logger *zap.Logger, connectors []net.IP, keyDB, dataDB *leveldb.DB, aead cipher.AEAD) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var (
			resp       Response
			statusCode = http.StatusOK
		)
		defer func() {
			deferWriteResponse(w, statusCode, resp)
		}()

		if !mustMethodPost(r, &statusCode, &resp) {
			logger.Error("invalid request method")
			return
		}

		// check connector
		if err := validateClientIP(connectors, r); err != nil {
			logger.Error("invalid client ip", zap.NamedError("appError", err))
			statusCode = http.StatusBadRequest
			resp.Code = 1
			resp.DebugMessage = fmt.Sprintf("validate client ip fail: %v", err)
			return
		}

		// ctx := r.Context()
		var req struct {
			KeyID              string `json:"keyID"`
			DataID             string `json:"dataID"`
			ReplaceCurrentItem bool   `json:"replaceCurrentItem"`
			Data               string `json:"data"`
			PlaintextHash      string `json:"plaintextHash"`
		}
		if err := unmarshalRequestBody(r, &req); err != nil {
			logger.Error("unmarshal request body fail", zap.NamedError("appError", err))
			statusCode = http.StatusBadRequest
			resp.Code = 1
			resp.DebugMessage = fmt.Sprintf("failed to unmarshal request: %v", err)
			return
		}
		logger.Info("SaveData accessed", zap.String("keyID", req.KeyID), zap.String("dataID", req.DataID))

		// check data from data.db
		dataID := []byte(req.DataID)
		_, err := dataDB.Get(dataID, nil)
		if err != nil {
			if err != leveldb.ErrNotFound {
				logger.Error("leveldb service error")
				statusCode = http.StatusInternalServerError
				resp.Code = 1
				resp.DebugMessage = fmt.Sprintf("failed to get dataID [%s]: %v", req.DataID, err)
				return
			}
		} else {
			if !req.ReplaceCurrentItem {
				logger.Error("data exists, but replaceCurrentItem is false", zap.String("dataID", req.DataID))
				statusCode = http.StatusBadRequest
				resp.Code = 1
				resp.DebugMessage = fmt.Sprintf("dataID [%s] exists, but replaceCurrentItem is %t", req.DataID, req.ReplaceCurrentItem)
				return
			}
		}

		// get private key from key.db
		storedPrivateKey, err := utils.GetStoredPrivateKey(keyDB, []byte(req.KeyID))
		if err != nil {
			logger.Error("get stored private key fail", zap.String("keyID", req.KeyID), zap.String("dataID", req.DataID), zap.NamedError("appError", err))
			statusCode = http.StatusBadRequest
			resp.Code = 1
			resp.DebugMessage = fmt.Sprintf("get stored private key [keyID=%s] fail: %v", req.KeyID, err)
			return
		}
		// decrypt the key
		privKeyBytes, err := utils.DecryptByX(storedPrivateKey.Private, aead)
		if err != nil {
			logger.Error("decode private key fail", zap.String("keyID", req.KeyID), zap.String("dataID", req.DataID), zap.NamedError("appError", err))
			statusCode = http.StatusBadRequest
			resp.Code = 1
			resp.DebugMessage = fmt.Sprintf("decode private key [keyID=%s] [dataID=%s] fail: %v", req.KeyID, req.DataID, err)
			return
		}

		// decrypt the data
		data, err := base64.StdEncoding.DecodeString(req.Data)
		if err != nil {
			logger.Error("decode request base64 data fail", zap.String("dataID", req.DataID))
			statusCode = http.StatusBadRequest
			resp.Code = 1
			resp.DebugMessage = fmt.Sprintf("decode request base64 data [dataID=%s] fail: %v", req.DataID, err)
			return
		}
		plaintext, err := utils.DecryptByEcies(privKeyBytes, data)
		if err != nil {
			logger.Error("decrypt request data fail", zap.String("keyID", req.KeyID), zap.String("dataID", req.DataID), zap.NamedError("appError", err))
			statusCode = http.StatusBadRequest
			resp.Code = 1
			resp.DebugMessage = fmt.Sprintf("decrypt request data [keyID=%s] [dataID=%s] fail: %v", req.KeyID, req.DataID, err)
			return
		}
		// hash the decrypted data
		plaintextHash, err := utils.Keccak256Hex(plaintext)
		if err != nil {
			logger.Error("hash request data fail", zap.String("keyID", req.KeyID), zap.String("keyID", req.KeyID), zap.String("dataID", req.DataID), zap.NamedError("appError", err))
			statusCode = http.StatusInternalServerError
			resp.Code = 1
			resp.DebugMessage = fmt.Sprintf("hash request data [keyID=%s] [dataID=%s] fail: %v", req.KeyID, req.DataID, err)
			return
		}
		if !bytes.Equal(plaintextHash, []byte(req.PlaintextHash)) {
			logger.Error("invalid plaintext hash", zap.String("keyID", req.KeyID), zap.String("dataID", req.DataID))
			statusCode = http.StatusBadRequest
			resp.Code = 1
			resp.DebugMessage = fmt.Sprintf("invalid plaintext hash [keyID=%s] [dataID=%s]", req.KeyID, req.DataID)
			return
		}

		// store the data
		storedData := StoredData{
			KeyID: req.KeyID,
			Data:  req.Data,
		}
		storedDataBytes, err := json.Marshal(storedData)
		if err != nil {
			logger.Error("marshal storedData fail", zap.String("keyID", req.KeyID), zap.String("dataID", req.DataID), zap.NamedError("appError", err))
			statusCode = http.StatusInternalServerError
			resp.Code = 1
			resp.DebugMessage = fmt.Sprintf("marshal storedData fail [keyID=%s] [dataID=%s]: %v", req.KeyID, req.DataID, err)
			return
		}
		if err := dataDB.Put(dataID, storedDataBytes, nil); err != nil {
			logger.Error("store data fail", zap.String("keyID", req.KeyID), zap.String("dataID", req.DataID), zap.NamedError("appError", err))
			statusCode = http.StatusInternalServerError
			resp.Code = 1
			resp.DebugMessage = fmt.Sprintf("store data fail [keyID=%s] [dataID=%s]", req.KeyID, req.DataID)
			return
		}

		logger.Info("SaveData success")
		resp.Code = 0
		resp.Message = "ok"
	})
}

func DeleteData(logger *zap.Logger, connectors []net.IP, keyDB, dataDB *leveldb.DB) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var (
			resp       Response
			statusCode = http.StatusOK
		)
		defer func() {
			deferWriteResponse(w, statusCode, resp)
		}()

		if !mustMethodPost(r, &statusCode, &resp) {
			logger.Error("invalid request method")
			return
		}

		// check connector
		if err := validateClientIP(connectors, r); err != nil {
			logger.Error("invalid client ip", zap.NamedError("appError", err))
			statusCode = http.StatusBadRequest
			resp.Code = 1
			resp.DebugMessage = fmt.Sprintf("validate client ip fail: %v", err)
			return
		}

		// ctx := r.Context()
		var req struct {
			KeyID  string `json:"keyID"`
			DataID string `json:"dataID"`
		}
		if err := unmarshalRequestBody(r, &req); err != nil {
			logger.Error("unmarshal request body fail", zap.NamedError("appError", err))
			statusCode = http.StatusBadRequest
			resp.Code = 1
			resp.DebugMessage = fmt.Sprintf("failed to unmarshal request: %v", err)
			return
		}
		logger.Info("DeleteData accessed", zap.String("keyID", req.KeyID), zap.String("dataID", req.DataID))

		// get data from data.db
		dataID := []byte(req.DataID)
		storedData, err := getStoredData(dataDB, dataID)
		if err != nil {
			logger.Error("get stored data fail", zap.String("keyID", req.KeyID), zap.String("dataID", req.DataID), zap.NamedError("appError", err))
			statusCode = http.StatusBadRequest
			resp.Code = 1
			resp.DebugMessage = fmt.Sprintf("get stored data fail: %v\n", err)
			return
		}
		if req.KeyID != storedData.KeyID {
			logger.Error("invalid key id", zap.String("keyID", req.KeyID), zap.String("dataID", req.DataID))
			statusCode = http.StatusBadRequest
			resp.Code = 1
			resp.DebugMessage = fmt.Sprintf("invalid key id [keyID=%s] [dataID=%s]", req.KeyID, req.DataID)
			return
		}

		// delete the data
		if err := dataDB.Delete(dataID, nil); err != nil {
			logger.Error("failed to delete data", zap.String("keyID", req.KeyID), zap.String("dataID", req.DataID), zap.NamedError("appError", err))
			statusCode = http.StatusInternalServerError
			resp.Code = 1
			resp.DebugMessage = fmt.Sprintf("failed to delete data [dataID=%s]: %v", req.DataID, err)
			return
		}

		logger.Info("DeleteData success")
		resp.Code = 0
		resp.Message = "ok"
	})
}

func DecryptData(logger *zap.Logger, connectors []net.IP, keyDB *leveldb.DB, aead cipher.AEAD) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var (
			resp       Response
			statusCode = http.StatusOK
		)
		defer func() {
			deferWriteResponse(w, statusCode, resp)
		}()

		if !mustMethodPost(r, &statusCode, &resp) {
			logger.Error("invalid request method")
			return
		}

		// check connector
		if err := validateClientIP(connectors, r); err != nil {
			logger.Error("invalid client ip", zap.NamedError("appError", err))
			statusCode = http.StatusBadRequest
			resp.Code = 1
			resp.DebugMessage = fmt.Sprintf("validate client ip fail: %v", err)
			return
		}

		// ctx := r.Context()
		var req struct {
			KeyID string `json:"keyID"`
			Data  string `json:"data"`
		}
		if err := unmarshalRequestBody(r, &req); err != nil {
			logger.Error("unmarshal request body fail", zap.NamedError("appError", err))
			statusCode = http.StatusBadRequest
			resp.Code = 1
			resp.DebugMessage = fmt.Sprintf("failed to unmarshal request: %v", err)
			return
		}
		logger.Info("DecryptData accessed", zap.String("keyID", req.KeyID))

		// get private key from key.db
		storedPrivateKey, err := utils.GetStoredPrivateKey(keyDB, []byte(req.KeyID))
		if err != nil {
			logger.Error("get stored private key fail", zap.String("keyID", req.KeyID), zap.NamedError("appError", err))
			statusCode = http.StatusBadRequest
			resp.Code = 1
			resp.DebugMessage = fmt.Sprintf("get stored private key [keyID=%s] fail: %v", req.KeyID, err)
			return
		}
		// decrypt the key
		privKeyBytes, err := utils.DecryptByX(storedPrivateKey.Private, aead)
		if err != nil {
			logger.Error("decode private key fail", zap.String("keyID", req.KeyID), zap.NamedError("appError", err))
			statusCode = http.StatusBadRequest
			resp.Code = 1
			resp.DebugMessage = fmt.Sprintf("decode private key [keyID=%s] fail: %v", req.KeyID, err)
			return
		}

		// decrypt the data
		data, err := base64.StdEncoding.DecodeString(req.Data)
		if err != nil {
			logger.Error("decode request base64 data fail", zap.String("keyID", req.KeyID), zap.NamedError("appError", err))
			statusCode = http.StatusBadRequest
			resp.Code = 1
			resp.DebugMessage = fmt.Sprintf("decode request base64 data [keyID=%s] fail: %v", req.KeyID, err)
			return
		}
		plaintext, err := utils.DecryptByEcies(privKeyBytes, data)
		if err != nil {
			logger.Error("decrypt request data fail", zap.String("keyID", req.KeyID), zap.NamedError("appError", err))
			statusCode = http.StatusBadRequest
			resp.Code = 1
			resp.DebugMessage = fmt.Sprintf("decrypt request data [keyID=%s] fail: %v", req.KeyID, err)
			return
		}

		logger.Info("DecryptData success")
		resp.Code = 0
		resp.Message = "ok"
		resp.Data = struct {
			DecrypedData string `json:"decrypedData"`
		}{
			DecrypedData: base64.StdEncoding.EncodeToString(plaintext),
		}
	}
}

func GetData(logger *zap.Logger, connectors []net.IP, keyDB, dataDB *leveldb.DB, aead cipher.AEAD) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var (
			resp       Response
			statusCode = http.StatusOK
		)
		defer func() {
			deferWriteResponse(w, statusCode, resp)
		}()

		if !mustMethodPost(r, &statusCode, &resp) {
			logger.Error("invalid request method")
			return
		}

		// check connector
		if err := validateClientIP(connectors, r); err != nil {
			logger.Error("invalid client ip", zap.NamedError("appError", err))
			statusCode = http.StatusBadRequest
			resp.Code = 1
			resp.DebugMessage = fmt.Sprintf("validate client ip fail: %v", err)
			return
		}

		// ctx := r.Context()
		var req struct {
			KeyID  string `json:"keyID"`
			DataID string `json:"dataID"`
		}
		if err := unmarshalRequestBody(r, &req); err != nil {
			logger.Error("unmarshal request body fail", zap.NamedError("appError", err))
			statusCode = http.StatusBadRequest
			resp.Code = 1
			resp.DebugMessage = fmt.Sprintf("failed to unmarshal request: %v", err)
			return
		}
		logger.Info("GetData accessed", zap.String("keyID", req.KeyID), zap.String("dataID", req.DataID))

		// get private key from key.db
		storedPrivateKey, err := utils.GetStoredPrivateKey(keyDB, []byte(req.KeyID))
		if err != nil {
			logger.Error("get stored private key fail", zap.String("keyID", req.KeyID), zap.String("dataID", req.DataID), zap.NamedError("appError", err))
			statusCode = http.StatusBadRequest
			resp.Code = 1
			resp.DebugMessage = fmt.Sprintf("get stored private key [keyID=%s] fail: %v", req.KeyID, err)
			return
		}
		// decrypt the key
		privKeyBytes, err := utils.DecryptByX(storedPrivateKey.Private, aead)
		if err != nil {
			logger.Error("ecode private key fail", zap.String("keyID", req.KeyID), zap.String("dataID", req.DataID), zap.NamedError("appError", err))
			statusCode = http.StatusBadRequest
			resp.Code = 1
			resp.DebugMessage = fmt.Sprintf("decode private key [keyID=%s] fail: %v", req.KeyID, err)
			return
		}

		// get data from data.db
		storedData, err := getStoredData(dataDB, []byte(req.DataID))
		if err != nil {
			logger.Error("get stored data fail", zap.String("keyID", req.KeyID), zap.String("dataID", req.DataID), zap.NamedError("appError", err))
			statusCode = http.StatusBadRequest
			resp.Code = 1
			resp.DebugMessage = fmt.Sprintf("get stored data fail: %v\n", err)
			return
		}
		if req.KeyID != storedData.KeyID {
			logger.Error("invalid request key id", zap.String("keyID", req.KeyID), zap.String("dataID", req.DataID))
			statusCode = http.StatusBadRequest
			resp.Code = 1
			resp.DebugMessage = fmt.Sprintf("invalid request keyID [keyID=%s] [dataID=%s]", req.KeyID, req.DataID)
			return
		}

		// decrypt the data
		data, err := base64.StdEncoding.DecodeString(storedData.Data)
		if err != nil {
			logger.Error("decode request base64 data fail", zap.String("keyID", req.KeyID), zap.String("dataID", req.DataID), zap.NamedError("appError", err))
			statusCode = http.StatusBadRequest
			resp.Code = 1
			resp.DebugMessage = fmt.Sprintf("decode request base64 data [keyID=%s] [dataID=%s] fail: %v", req.KeyID, req.DataID, err)
			return
		}
		plaintext, err := utils.DecryptByEcies(privKeyBytes, data)
		if err != nil {
			logger.Error("decrypt request data fail", zap.String("keyID", req.KeyID), zap.String("dataID", req.DataID), zap.NamedError("appError", err))
			statusCode = http.StatusBadRequest
			resp.Code = 1
			resp.DebugMessage = fmt.Sprintf("decrypt request data [keyID=%s] [dataID=%s] fail: %v", req.KeyID, req.DataID, err)
			return
		}

		logger.Info("GetData success")
		resp.Code = 0
		resp.Message = "ok"
		resp.Data = struct {
			Data string `json:"data"`
		}{
			Data: base64.StdEncoding.EncodeToString(plaintext),
		}
	})
}

func CheckKeyExisting(logger *zap.Logger, connectors []net.IP, keyDB *leveldb.DB) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var (
			resp       Response
			statusCode = http.StatusOK
		)
		defer func() {
			deferWriteResponse(w, statusCode, resp)
		}()

		if !mustMethodPost(r, &statusCode, &resp) {
			logger.Error("invalid request method")
			return
		}

		// check connector
		if err := validateClientIP(connectors, r); err != nil {
			logger.Error("invalid client ip", zap.NamedError("appError", err))
			statusCode = http.StatusBadRequest
			resp.Code = 1
			resp.DebugMessage = fmt.Sprintf("validate client ip fail: %v", err)
			return
		}

		// ctx := r.Context()
		var req struct {
			KeyID string `json:"keyID"`
		}
		if err := unmarshalRequestBody(r, &req); err != nil {
			logger.Error("unmarshal request body fail", zap.NamedError("appError", err))
			statusCode = http.StatusBadRequest
			resp.Code = 1
			resp.DebugMessage = fmt.Sprintf("failed to unmarshal request: %v", err)
			return
		}
		logger.Info("CheckKeyExisting accessed", zap.String("keyID", req.KeyID))

		exist := true
		// get key from key.db
		if _, err := keyDB.Get([]byte(req.KeyID), nil); err != nil {
			if err != leveldb.ErrNotFound {
				logger.Error("failed to get key", zap.String("keyID", req.KeyID), zap.NamedError("appError", err))
				statusCode = http.StatusInternalServerError
				resp.Code = 1
				resp.DebugMessage = fmt.Sprintf("failed to get key [keyID=%s]: %v", req.KeyID, err)
				return
			} else {
				exist = false
			}
		}

		logger.Info("CheckKeyExisting success")
		resp.Code = 0
		resp.Message = "ok"
		resp.Data = struct {
			Existing bool `json:"existing"`
		}{
			Existing: exist,
		}
	})
}

func GetLogs(logger *zap.Logger, connectors []net.IP, keyDB *leveldb.DB) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var (
			resp       Response
			statusCode = http.StatusOK
		)
		defer func() {
			deferWriteResponse(w, statusCode, resp)
		}()

		if !mustMethodPost(r, &statusCode, &resp) {
			logger.Error("invalid request method")
			return
		}

		// check connector
		if err := validateClientIP(connectors, r); err != nil {
			logger.Error("invalid client ip", zap.NamedError("appError", err))
			statusCode = http.StatusBadRequest
			resp.Code = 1
			resp.DebugMessage = fmt.Sprintf("validate client ip fail: %v", err)
			return
		}

		// ctx := r.Context()
		var req struct {
			PageNumber int `json:"pageNumber"`
			PageSize   int `json:"pageSize"`
		}
		if err := unmarshalRequestBody(r, &req); err != nil {
			logger.Error("unmarshal request body fail", zap.NamedError("appError", err))
			statusCode = http.StatusBadRequest
			resp.Code = 1
			resp.DebugMessage = fmt.Sprintf("failed to unmarshal request: %v", err)
			return
		}
		logger.Info("GetLogs accessed")

		var total int
		// asc
		var list []string
		// parse log file
		logFile := "/tmp/keystore.log"
		f, err := os.Open(logFile)
		if err != nil {
			logger.Error("open keystore log file fail", zap.String("path", logFile), zap.NamedError("appError", err))
			statusCode = http.StatusInternalServerError
			resp.Code = 1
			resp.DebugMessage = fmt.Sprintf("open keystore log file fail: %v", err)
			return
		}
		br := bufio.NewReader(f)
		for {
			if _, err := br.ReadBytes('\n'); err != nil {
				break
			}
			total++
		}
		if err := f.Close(); err != nil {
			logger.Error("close keystore log file fail", zap.String("path", logFile), zap.NamedError("appError", err))
			statusCode = http.StatusInternalServerError
			resp.Code = 1
			resp.DebugMessage = fmt.Sprintf("close keystore log file fail: %v", err)
			return
		}
		// desc
		last := total - req.PageNumber*req.PageSize
		if last > 0 {
			first := total - (req.PageNumber+1)*req.PageSize
			if first < 0 {
				first = 0
			}
			f, err = os.Open(logFile)
			if err != nil {
				logger.Error("open keystore log file fail", zap.String("path", logFile), zap.NamedError("appError", err))
				statusCode = http.StatusInternalServerError
				resp.Code = 1
				resp.DebugMessage = fmt.Sprintf("open keystore log file fail: %v", err)
				return
			}
			br.Reset(f)
			total = 0
			for {
				line, err := br.ReadString('\n')
				if err != nil {
					break
				}
				if first <= total && total < last {
					list = append(list, line)
				}
				total++
			}
		}
		var res []string
		for i := len(list) - 1; i >= 0; i-- {
			res = append(res, list[i])
		}

		logger.Info("GetLogs success")
		resp.Code = 0
		resp.Message = "ok"
		resp.Data = struct {
			Total int      `json:"total"`
			List  []string `json:"list"`
		}{
			Total: total,
			List:  res,
		}
	})
}

type MonitorInfo struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

func GetMonitorInfos(logger *zap.Logger, connectors []net.IP) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var (
			resp       Response
			statusCode = http.StatusOK
		)
		defer func() {
			deferWriteResponse(w, statusCode, resp)
		}()

		if !mustMethodPost(r, &statusCode, &resp) {
			logger.Error("invalid request method")
			return
		}

		// check connector
		if err := validateClientIP(connectors, r); err != nil {
			logger.Error("invalid client ip", zap.NamedError("appError", err))
			statusCode = http.StatusBadRequest
			resp.Code = 1
			resp.DebugMessage = fmt.Sprintf("validate client ip fail: %v", err)
			return
		}

		logger.Info("GetMonitorInfos accessed")

		var infos []MonitorInfo

		keyDBPath := strings.TrimSpace(os.Getenv("KEY_DB"))
		out, err := exec.Command("du", "-hd", "1", keyDBPath).Output() // example output: 32K\t/var/lib/keystore/key.db\n
		if err != nil {
			logger.Error("get size of key.db fail", zap.String("keyDBPath", keyDBPath), zap.NamedError("appError", err))
			statusCode = http.StatusInternalServerError
			resp.Code = 1
			resp.DebugMessage = fmt.Sprintf("get size of key.db fail: %v", err)
			return
		}
		infos = append(infos, MonitorInfo{Name: "keyDBSize", Value: string(bytes.Split(out, []byte("\t"))[0])})

		dataDBPath := strings.TrimSpace(os.Getenv("DATA_DB"))
		out, err = exec.Command("du", "-hd", "1", dataDBPath).Output()
		if err != nil {
			logger.Error("get size of data.db fail", zap.String("dataDBPath", dataDBPath), zap.NamedError("appError", err))
			statusCode = http.StatusInternalServerError
			resp.Code = 1
			resp.DebugMessage = fmt.Sprintf("get size of data.db fail: %v", err)
			return
		}
		infos = append(infos, MonitorInfo{Name: "dataDBSize", Value: string(bytes.Split(out, []byte("\t"))[0])})

		hostPath := "/host"
		hostUsageStat, err := disk.Usage(hostPath)
		if err != nil {
			logger.Error("get disk usage of host fail", zap.String("hostPath", hostPath), zap.NamedError("appError", err))
			statusCode = http.StatusInternalServerError
			resp.Code = 1
			resp.DebugMessage = fmt.Sprintf("get disk usage of host fail: %v", err)
			return
		}
		infos = append(infos, MonitorInfo{Name: "totalSystemDisk", Value: fmt.Sprintf("%.2fG", float64(hostUsageStat.Total)/1024/1024/1024)})
		infos = append(infos, MonitorInfo{Name: "freeSystemDisk", Value: fmt.Sprintf("%.2fG", float64(hostUsageStat.Free)/1024/1024/1024)})
		infos = append(infos, MonitorInfo{Name: "systemDiskUsedPercent", Value: fmt.Sprintf("%.2f%%", hostUsageStat.UsedPercent)})

		logger.Info("GetMonitorInfos success")
		resp.Code = 0
		resp.Message = "ok"
		resp.Data = struct {
			Infos []MonitorInfo `json:"infos"`
		}{
			Infos: infos,
		}
	})
}
