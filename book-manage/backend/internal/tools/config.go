package tools

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/golang/glog"
)

type ConfigInfo struct {
	Debug        bool
	ServerAddr   string
	DownloadPath string
	BackupPath   string
	Mysql        string
	Accounts     []struct {
		User     string
		Password string
	}
	TTSAPIKey string
}

var (
	cnf *ConfigInfo
)

// InitConfig 初始化配置
func InitConfig(cfgFile string) {
	r, err := os.Open(cfgFile) //从本地读取配置文件
	if err != nil {
		glog.Fatalf("InitConfig open local file err %s", err.Error())
		return
	}

	decoder := json.NewDecoder(r)
	err = decoder.Decode(&cnf)
	if err != nil {
		glog.Fatalf(err.Error())
		return
	}
	fmt.Println("InitConfig success:")
	for _, account := range cnf.Accounts {
		fmt.Printf("account: %s, password: %s\n", account.User, account.Password)
	}
	glog.V(5).Infof("init config success")
}

type BookLevelConfig struct {
	Items []BookLevel
}

type BookLevel struct {
	BookID int64
	Level  string
}

var (
	DefaultBookLevelConfig *BookLevelConfig
)

func InitLevelToml(cfgFile string) {
	data, err := os.ReadFile(cfgFile)
	if err != nil {
		glog.Fatalf("InitLevelToml open local file err %s", err.Error())
		return
	}
	var config BookLevelConfig
	if err := toml.Unmarshal(data, &config); err != nil {
		glog.Fatalf("toml Unmarshal: %s", err.Error())
		return
	}
	DefaultBookLevelConfig = &config
	glog.V(5).Infof("InitLevelToml success")
}

func InBookLevelConfig(bookID int64) bool {
	for _, item := range DefaultBookLevelConfig.Items {
		if item.BookID == bookID {
			return true
		}
	}
	return false
}

func initMYSQLParams() {
	if err := InitMysql(cnf.Mysql); err != nil {
		panic(any("initial mysql error: " + err.Error()))
	}
}

func CheckPassword(password string) bool {
	for _, account := range cnf.Accounts {
		if account.Password == password {
			return true
		}
	}
	return false
}

func GetUserNameByPassword(password string) string {
	for _, account := range cnf.Accounts {
		if account.Password == password {
			return account.User
		}
	}
	return ""
}

func CheckAdminPassword(password string) bool {
	for _, account := range cnf.Accounts {
		if account.User == "admin" && account.Password == password {
			return true
		}
	}
	return false
}

func StaffList() []struct {
	User     string
	Password string
} {
	return Config().Accounts
}

func InitCommonTools() {
	initMYSQLParams()

	RegisterCustomValidation()
}

func Config() *ConfigInfo {
	if cnf == nil {
		panic(any("cnf == nil"))
	}

	return cnf
}
