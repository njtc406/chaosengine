// Package dbservice
// @Title  数据库服务
// @Description  数据库服务
// @Author  yr  2024/7/25 下午3:10
// @Update  yr  2024/7/25 下午3:10
package dbservice

import (
	"github.com/njtc406/chaosengine/engine/core"
	"github.com/njtc406/chaosengine/engine/node"
	nodeConfig "github.com/njtc406/chaosengine/engine/node/config"

	"github.com/njtc406/chaosengine/engine/sysModule/mysqlmodule"
	"github.com/njtc406/chaosengine/engine/sysModule/redismodule"
	"github.com/njtc406/chaosengine/engine/sysService/dbservice/config"
	"github.com/njtc406/chaosengine/engine/utils/log"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"runtime/debug"
	"time"
)

func init() {
	node.SetupBase(&DBService{})
	nodeConfig.RegisterConf(&nodeConfig.ConfInfo{
		ServiceName:   "DBService",
		ConfName:      "db",
		ConfType:      "yaml",
		ConfPath:      "",
		Cfg:           &config.DBService{},
		DefaultSetFun: config.DefaultDBService,
	})
}

type Callback func(rdb *redis.Client, mdb *gorm.DB, args ...interface{}) (interface{}, error)

type DBService struct {
	core.Service

	redisModule *redismodule.RedisModule
	mysqlModule *mysqlmodule.MysqlModule
	mongoModule int // 预留
}

func (db *DBService) getConf() *config.DBService {
	return db.GetServiceCfg().(*config.DBService)
}

func (db *DBService) OnInit() error {
	conf := db.getConf()
	db.redisModule = redismodule.NewRedisModule()
	db.redisModule.Init(conf.RedisConf)
	db.mysqlModule = mysqlmodule.NewMysqlModule()
	db.mysqlModule.InitConn(conf.MysqlConf, nodeConfig.Conf.NodeConf.ID)

	// 设置线程数
	db.SetGoRoutineNum(conf.GoroutineNum)

	db.AddModule(db.redisModule)
	db.AddModule(db.mysqlModule)
	return nil
}

func (db *DBService) OnStart() error {
	log.SysLogger.Infof("db服务启动完成...")
	return nil
}

func (db *DBService) OnRelease() {
	// 释放所有子模块
	db.ReleaseAllChildModule()
	log.SysLogger.Infof("db服务释放完成...")
}

func (db *DBService) APISetRedisString(key string, value interface{}, expire time.Duration) error {
	return db.redisModule.SetString(key, value, expire)
}

func (db *DBService) APIGetRedisString(key string) (string, error) {
	return db.redisModule.GetString(key)
}

func (db *DBService) APISetRedisStringJson(key string, value interface{}, expire time.Duration) error {
	return db.redisModule.SetStringJson(key, value, expire)
}

func (db *DBService) APIGetRedisStringJson(key string, value interface{}) error {
	return db.redisModule.GetStringJson(key, value)
}

// APIExecuteRedisFun 执行redis函数
func (db *DBService) APIExecuteRedisFun(f redismodule.Callback, args ...interface{}) (interface{}, error) {
	return db.redisModule.ExecuteFun(f, args...)
}

// APIExecuteMysqlFun 执行mysql函数
func (db *DBService) APIExecuteMysqlFun(f mysqlmodule.Callback, args ...interface{}) (interface{}, error) {
	return db.mysqlModule.ExecuteFun(f, args...)
}

func (db *DBService) APIExecuteMysqlTransaction(funcList ...mysqlmodule.TransactionCallback) error {
	return db.mysqlModule.ExecuteTransaction(funcList...)
}

// APIExecuteMixedFun 执行混合函数
func (db *DBService) APIExecuteMixedFun(f Callback, args ...interface{}) (interface{}, error) {
	defer func() {
		if r := recover(); r != nil {
			log.SysLogger.Errorf("mixed execute function panic: %v\ntrace:%s", r, debug.Stack())
		}
	}()
	return f(db.redisModule.GetClient(), db.mysqlModule.GetClient(), args...)
}
