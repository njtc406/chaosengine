package config

import (
	"github.com/njtc406/chaosengine/engine/sysModule/httpmodule"
	"github.com/spf13/viper"
	"time"
)

type LoginServiceConf struct {
	LoginServerConf *httpmodule.Conf `binding:"required"`
}

func SetLoginServiceConfDefault(parser *viper.Viper) {
	parser.SetDefault("LoginServerConf", &httpmodule.Conf{
		Addr:              "0.0.0.0:80",
		ReadHeaderTimeout: time.Second * 10,
		IdleTimeout:       time.Second * 30,
		CachePath:         "",
		ResourceRootPath:  "",
		HttpDir:           "",
		StaticDir:         "",
		Auth:              false,
		Account:           nil,
	})
}

type GMServiceConf struct {
	GMServerConf *httpmodule.Conf `binding:"required"`
	GmAddr       string           `binding:""` // gm后台地址
}

func SetGMServiceConfDefault(parser *viper.Viper) {
	parser.SetDefault("GMServerConf", &httpmodule.Conf{
		Addr:              "0.0.0.0:81",
		ReadHeaderTimeout: time.Second * 10,
		IdleTimeout:       time.Second * 30,
		CachePath:         "",
		ResourceRootPath:  "",
		HttpDir:           "",
		StaticDir:         "",
		Auth:              false,
		Account:           nil,
	})
}
