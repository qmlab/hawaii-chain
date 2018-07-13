package config

import (
	config "github.com/micro/go-config"
	"github.com/micro/go-config/source/file"
)

var Usrcfg User
var InitialAccounts Accounts

type User struct {
	Address string `json:"address"`
	Type    int    `json:"type"`
}

type Account struct {
	Address string  `json:"address"`
	Val     float64 `json:"val"`
}

type Accounts []Account

func InitConfig(cfgfile string) {
	config.Load(file.NewSource(
		file.WithPath(cfgfile),
	))
	config.Get("user").Scan(&Usrcfg)
	config.Get("init").Scan(&InitialAccounts)
}
