package config

import (
	"github.com/njtc406/chaosengine/engine/sysModule/mysqlmodule"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
	"time"
)

type DBService struct {
	MysqlConf    *mysqlmodule.Conf `binding:"required"` // mysql配置
	RedisConf    *redis.Options    `binding:"required"` // redis配置
	GoroutineNum int32             `binding:"required"` // goroutine数量
}

func DefaultDBService(parser *viper.Viper) {
	parser.SetDefault("MysqlConf", &mysqlmodule.Conf{
		UserName:           "",
		Passwd:             "",
		Net:                "tcp",
		Addr:               "0.0.0.0:3306",
		DBNamePrefix:       "chaos_",
		TimeZone:           "Local",
		Timeout:            time.Second * 10,
		ReadTimeout:        time.Second * 10,
		WriteTimeout:       time.Second * 10,
		SetConnMaxIdleTime: 0,
		SetConnMaxLifetime: 0,
		SetMaxIdleConns:    0,
		SetMaxOpenConns:    0,
	})
}