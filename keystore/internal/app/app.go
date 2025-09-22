package app

import (
	"bytes"
	"crypto/cipher"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/nextsurfer/ground/pkg/util"
	"github.com/nextsurfer/keystore/internal/app/httphandlers"
	"github.com/nextsurfer/keystore/internal/app/utils"
	"github.com/rs/cors"
	"github.com/syndtr/goleveldb/leveldb"
	"go.uber.org/zap"
	"golang.org/x/crypto/chacha20poly1305"
)

// HTTP app
type KeyStoreApp struct {
	Env        util.APPEnvType
	Name       string
	Host       string
	Port       int // http port
	SecretKey  []byte
	Nonce      []byte
	Aead       cipher.AEAD
	SystemKey  string
	KeyDB      *leveldb.DB
	DataDB     *leveldb.DB
	Connectors []net.IP
	Logger     *zap.Logger
}

func NewKeyStoreApp(name string, logger *zap.Logger, port int, host string, appEnv int) *KeyStoreApp {
	var err error
	env := util.EnvForInt(appEnv)

	// check CONNECTORS environment variable
	connectorsEnv := os.Getenv("CONNECTORS")
	if connectorsEnv == "" {
		logger.Fatal("must set env variable for 'CONNECTORS'")
	}
	var connectors []net.IP
	ipStrings := strings.Split(connectorsEnv, ",")
	for _, ipString := range ipStrings {
		if ip := net.ParseIP(ipString); ip != nil {
			connectors = append(connectors, ip)
		} else {
			logger.Fatal("'CONNECTORS' env variable has invalid ip string")
		}
	}

	// load leveldb key.db, data.db
	keydbEnv := strings.TrimSpace(os.Getenv("KEY_DB"))
	if keydbEnv == "" {
		logger.Fatal("must set env variable for 'KEY_DB'")
	}
	keyDB, err := leveldb.OpenFile(keydbEnv, nil)
	if err != nil {
		logger.Fatal("open key.db fail", zap.NamedError("appError", err))
	}
	datadbEnv := strings.TrimSpace(os.Getenv("DATA_DB"))
	if datadbEnv == "" {
		logger.Fatal("must set env variable for 'DATA_DB'")
	}
	dataDB, err := leveldb.OpenFile(datadbEnv, nil)
	if err != nil {
		logger.Fatal("open data.db fail", zap.NamedError("appError", err))
	}

	var secretKey []byte
	var aead cipher.AEAD
	var nonce []byte
	var systemKey string
	storedSystemKeyKey := []byte("pswd_systemKeyHash")

	// OLD_SYSTEM_KEY, SYSTEM_KEY
	oldSystemKeyEnv := strings.TrimSpace(os.Getenv("OLD_SYSTEM_KEY"))
	systemKeyEnv := strings.TrimSpace(os.Getenv("SYSTEM_KEY"))
	if oldSystemKeyEnv == "" && systemKeyEnv == "" {
		// validate system key
		if _, err := utils.GetStoredPasswordKey(keyDB, storedSystemKeyKey); err != nil {
			if err != leveldb.ErrNotFound {
				logger.Fatal("OLD_SYSTEM_KEY and SYSTEM_KEY are all empty", zap.NamedError("appError", err))
			}
		} else {
			logger.Fatal("OLD_SYSTEM_KEY and SYSTEM_KEY are all empty, but systemKeyHash has existed")
		}
		// this is the first setup
		secretKey, err = utils.NewSecretKey()
		if err != nil {
			logger.Fatal("OLD_SYSTEM_KEY and SYSTEM_KEY are all empty", zap.NamedError("appError", err))
		}
		nonce, aead, err = utils.NewNonceAndX(secretKey)
		if err != nil {
			logger.Fatal("OLD_SYSTEM_KEY and SYSTEM_KEY are all empty", zap.NamedError("appError", err))
		}
		systemKey, err = utils.SystemKey(secretKey, nonce)
		if err != nil {
			logger.Fatal("OLD_SYSTEM_KEY and SYSTEM_KEY are all empty", zap.NamedError("appError", err))
		}
		systemKeyHash, err := utils.Keccak256Hex([]byte(systemKey))
		if err != nil {
			logger.Fatal("OLD_SYSTEM_KEY and SYSTEM_KEY are all empty", zap.NamedError("appError", err))
		}
		if err := utils.PutStoredPasswordKey(keyDB, storedSystemKeyKey, utils.EncryptByX([]byte(systemKey), aead, nonce), systemKeyHash); err != nil {
			logger.Fatal("OLD_SYSTEM_KEY and SYSTEM_KEY are all empty", zap.NamedError("appError", err))
		}
	} else if systemKeyEnv != "" {
		// validate system key
		storedSystemKey, err := utils.GetStoredPasswordKey(keyDB, storedSystemKeyKey)
		if err != nil {
			logger.Fatal("SYSTEM_KEY is not empty", zap.NamedError("appError", err))
		}
		_, err = utils.GetStoredPasswordKey(keyDB, storedSystemKeyKey)
		if err != nil {
			logger.Fatal("SYSTEM_KEY is not empty", zap.NamedError("appError", err))
		}
		hashedSystemKeyEnv, err := utils.Keccak256Hex([]byte(systemKeyEnv))
		if err != nil {
			logger.Fatal("SYSTEM_KEY is not empty", zap.NamedError("appError", err))
		}
		if !bytes.Equal(hashedSystemKeyEnv, storedSystemKey.PasswordHash) {
			logger.Fatal("SYSTEM_KEY is not empty, but invalid")
		}
		systemKey = systemKeyEnv
		// parse system key
		secretKey, nonce, err = utils.ParseSystemKey(systemKeyEnv)
		if err != nil {
			logger.Fatal("OLD_SYSTEM_KEY is not empty", zap.NamedError("appError", err))
		}
		aead, err = chacha20poly1305.NewX(secretKey)
		if err != nil {
			logger.Fatal("SYSTEM_KEY is not empty", zap.NamedError("appError", err))
		}
	} else if oldSystemKeyEnv != "" {
		// validate old system key
		storedSystemKey, err := utils.GetStoredPasswordKey(keyDB, storedSystemKeyKey)
		if err != nil {
			logger.Fatal("OLD_SYSTEM_KEY is not empty", zap.NamedError("appError", err))
		}
		hashedOldSystemKeyEnv, err := utils.Keccak256Hex([]byte(oldSystemKeyEnv))
		if err != nil {
			logger.Fatal("OLD_SYSTEM_KEY is not empty", zap.NamedError("appError", err))
		}
		if !bytes.Equal(hashedOldSystemKeyEnv, storedSystemKey.PasswordHash) {
			logger.Fatal("OLD_SYSTEM_KEY is not empty, but invalid")
		}
		// parse old system key
		oldSecretKey, _, err := utils.ParseSystemKey(oldSystemKeyEnv)
		if err != nil {
			logger.Fatal("OLD_SYSTEM_KEY is not empty", zap.NamedError("appError", err))
		}
		oldAead, err := chacha20poly1305.NewX(oldSecretKey)
		if err != nil {
			logger.Fatal("OLD_SYSTEM_KEY is not empty", zap.NamedError("appError", err))
		}
		// new system key
		secretKey, err = utils.NewSecretKey()
		if err != nil {
			logger.Fatal("OLD_SYSTEM_KEY is not empty", zap.NamedError("appError", err))
		}
		nonce, aead, err = utils.NewNonceAndX(secretKey)
		if err != nil {
			logger.Fatal("OLD_SYSTEM_KEY is not empty", zap.NamedError("appError", err))
		}
		systemKey, err = utils.SystemKey(secretKey, nonce)
		if err != nil {
			logger.Fatal("OLD_SYSTEM_KEY is not empty", zap.NamedError("appError", err))
		}
		// copy bak_key.db from key.db
		if err := utils.CopyDir(keydbEnv, filepath.Dir(keydbEnv)+"/bak_key.db"); err != nil {
			logger.Fatal("OLD_SYSTEM_KEY is not empty", zap.NamedError("appError", err))
		}
		tmpKeyDBPath := filepath.Dir(keydbEnv) + "/tmp_key.db"
		tmpKeyDB, err := leveldb.OpenFile(tmpKeyDBPath, nil)
		if err != nil {
			logger.Fatal("OLD_SYSTEM_KEY is not empty", zap.NamedError("appError", err))
		}
		// re-encrypt data in key.db to tmp_key.db
		batch := new(leveldb.Batch)
		iter := keyDB.NewIterator(nil, nil)
		for iter.Next() {
			key := iter.Key()
			value := iter.Value()
			var storedKey utils.StoredKey
			if bytes.HasPrefix(key, []byte("pswd_")) {
				storedKey = &utils.StoredPasswordKey{}
			} else {
				storedKey = &utils.StoredPrivateKey{}
			}
			if err := json.Unmarshal(value, &storedKey); err != nil {
				logger.Fatal("OLD_SYSTEM_KEY is not empty", zap.NamedError("appError", fmt.Errorf("json unmarshal key id [%s] fail: %v", key, err)))
			}
			plaintext, err := utils.DecryptByX(storedKey.Key(), oldAead)
			if err != nil {
				logger.Fatal("OLD_SYSTEM_KEY is not empty", zap.NamedError("appError", fmt.Errorf("decrypt key id [%s] by old aead fail: %v", key, err)))
			}
			storedKey.SetKey(utils.EncryptByX(plaintext, aead, nonce))
			storedKeyBytes, err := json.Marshal(&storedKey)
			if err != nil {
				logger.Fatal("OLD_SYSTEM_KEY is not empty", zap.NamedError("appError", fmt.Errorf("json marshal key id [%s] fail: %v", key, err)))
			}
			batch.Put(key, storedKeyBytes)
		}
		iter.Release()
		if err = iter.Error(); err != nil {
			logger.Fatal("OLD_SYSTEM_KEY is not empty", zap.NamedError("appError", err))
		}
		if err := tmpKeyDB.Write(batch, nil); err != nil {
			logger.Fatal("OLD_SYSTEM_KEY is not empty", zap.NamedError("appError", err))
		}
		// remove key.db, and move tmp_key.db to key.db
		if err := keyDB.Close(); err != nil {
			logger.Fatal("OLD_SYSTEM_KEY is not empty", zap.NamedError("appError", err))
		}
		if err := tmpKeyDB.Close(); err != nil {
			logger.Fatal("OLD_SYSTEM_KEY is not empty", zap.NamedError("appError", err))
		}
		// remove key.db
		if err := os.RemoveAll(keydbEnv); err != nil {
			logger.Fatal("OLD_SYSTEM_KEY is not empty", zap.NamedError("appError", err))
		}
		// rename tmp_key.db to key.db
		if err := os.Rename(tmpKeyDBPath, keydbEnv); err != nil {
			logger.Fatal("OLD_SYSTEM_KEY is not empty", zap.NamedError("appError", err))
		}
		keyDB, err = leveldb.OpenFile(keydbEnv, nil)
		if err != nil {
			logger.Fatal("OLD_SYSTEM_KEY is not empty", zap.NamedError("appError", err))
		}
		// update systemKeyHash in key.db
		systemKeyHash, err := utils.Keccak256Hex([]byte(systemKey))
		if err != nil {
			logger.Fatal("OLD_SYSTEM_KEY is not empty", zap.NamedError("appError", err))
		}
		if err := utils.PutStoredPasswordKey(keyDB, storedSystemKeyKey, utils.EncryptByX([]byte(systemKey), aead, nonce), systemKeyHash); err != nil {
			logger.Fatal("OLD_SYSTEM_KEY is not empty", zap.NamedError("appError", err))
		}
	}
	logger.Info("environment variables",
		zap.String("ENV", util.LabelForEnv(env)),
		zap.String("CONNECTORS", connectorsEnv),
		zap.String("KEY_DB", keydbEnv),
		zap.String("DATA_DB", datadbEnv),
		zap.String("OLD_SYSTEM_KEY", oldSystemKeyEnv),
		zap.String("SYSTEM_KEY", systemKey),
	)

	return &KeyStoreApp{
		Env:        env,
		Name:       name,
		Host:       host,
		Port:       port,
		SecretKey:  secretKey,
		Nonce:      nonce,
		Aead:       aead,
		SystemKey:  systemKey,
		KeyDB:      keyDB,
		DataDB:     dataDB,
		Connectors: connectors,
		Logger:     logger,
	}
}

func (a *KeyStoreApp) Start() {
	http.HandleFunc("/keystore/createPassword/v1", httphandlers.CreatePassword(a.Logger, a.Connectors, a.KeyDB, a.Aead, a.Nonce))
	http.HandleFunc("/keystore/checkPassword/v1", httphandlers.CheckPassword(a.Logger, a.Connectors, a.KeyDB))
	http.HandleFunc("/keystore/getPassword/v1", httphandlers.GetPassword(a.Logger, a.Connectors, a.KeyDB, a.Aead))
	http.HandleFunc("/keystore/deletePassword/v1", httphandlers.DeletePassword(a.Logger, a.Connectors, a.KeyDB))
	http.HandleFunc("/keystore/createPrivateKey/v1", httphandlers.CreatePrivateKey(a.Logger, a.Connectors, a.KeyDB, a.Aead, a.Nonce))
	http.HandleFunc("/keystore/getPublicKey/v1", httphandlers.GetPublicKey(a.Logger, a.Connectors, a.KeyDB))
	http.HandleFunc("/keystore/deletePrivateKey/v1", httphandlers.DeletePrivateKey(a.Logger, a.Connectors, a.KeyDB))
	http.HandleFunc("/keystore/saveData/v1", httphandlers.SaveData(a.Logger, a.Connectors, a.KeyDB, a.DataDB, a.Aead))
	http.HandleFunc("/keystore/deleteData/v1", httphandlers.DeleteData(a.Logger, a.Connectors, a.KeyDB, a.DataDB))
	http.HandleFunc("/keystore/decryptData/v1", httphandlers.DecryptData(a.Logger, a.Connectors, a.KeyDB, a.Aead))
	http.HandleFunc("/keystore/getData/v1", httphandlers.GetData(a.Logger, a.Connectors, a.KeyDB, a.DataDB, a.Aead))
	http.HandleFunc("/keystore/checkKeyExisting/v1", httphandlers.CheckKeyExisting(a.Logger, a.Connectors, a.KeyDB))
	http.HandleFunc("/keystore/getLogs/v1", httphandlers.GetLogs(a.Logger, a.Connectors, a.KeyDB))
	http.HandleFunc("/keystore/getMonitorInfos/v1", httphandlers.GetMonitorInfos(a.Logger, a.Connectors))

	// cors
	handler := cors.Default().Handler(http.DefaultServeMux)

	// api server
	go func() {
		addr := fmt.Sprintf("%s:%v", a.Host, a.Port)
		a.Logger.Info(fmt.Sprintf("keystore api server is starting at %s ...", addr))
		server := &http.Server{Addr: addr, Handler: handler}
		if err := server.ListenAndServe(); err != nil {
			a.Logger.Fatal("server error", zap.NamedError("appError", err))
		}
	}()
}

func (a *KeyStoreApp) Stop() {
	if err := a.KeyDB.Close(); err != nil {
		a.Logger.Info("close keyDB fail", zap.NamedError("appError", err))
	}
	if err := a.DataDB.Close(); err != nil {
		a.Logger.Info("close dataDB fail", zap.NamedError("appError", err))
	}
}
