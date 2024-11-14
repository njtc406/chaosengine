package config

import (
	"github.com/njtc406/chaosengine/engine/sysModule/mysqlmodule"
	"github.com/redis/go-redis/v9"
)

type DBService struct {
	MysqlConf    *mysqlmodule.Conf `binding:"required"` // mysql配置
	RedisConf    *redis.Options    `binding:"required"` // redis配置
	GoroutineNum int32             `binding:"required"` // goroutine数量
}
