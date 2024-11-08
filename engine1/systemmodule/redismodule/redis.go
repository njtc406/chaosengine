// Package redismodule
// @Title  redis模块
// @Description  redis模块
// @Author  yr  2024/7/25 下午3:13
// @Update  yr  2024/7/25 下午3:13
package redismodule

import (
	"context"
	"encoding/json"
	"github.com/njtc406/chaosengine/engine1/errdef/errcode"
	"github.com/njtc406/chaosengine/engine1/service"
	"github.com/njtc406/chaosutil/chaoserrors"
	"github.com/redis/go-redis/v9"
	"time"
)

type RedisModule struct {
	service.Module

	client *redis.Client
}

type Callback func(tx *redis.Client, args ...interface{}) (interface{}, chaoserrors.CError)

func NewRedisModule() *RedisModule {
	return &RedisModule{}
}

func (rm *RedisModule) Init(conf *redis.Options) {
	rm.client = redis.NewClient(conf)
}

func (rm *RedisModule) OnRelease() {
	if rm.client != nil {
		rm.client.Close()
	}
}

func (rm *RedisModule) GetClient() *redis.Client {
	return rm.client
}

func (rm *RedisModule) SetString(key string, value interface{}, expire time.Duration) chaoserrors.CError {
	if err := rm.client.Set(context.Background(), key, value, expire).Err(); err != nil {
		return chaoserrors.NewErrCode(errcode.DBRedisError, "redis errdef", err)
	}
	return nil
}

func (rm *RedisModule) SetStringJson(key string, value interface{}, expire time.Duration) chaoserrors.CError {
	tmp, err := json.Marshal(value)
	if err != nil {
		return chaoserrors.NewErrCode(errcode.ParamsInvalid, "not in json format", err)
	}
	if err := rm.client.Set(context.Background(), key, string(tmp), expire).Err(); err != nil {
		return chaoserrors.NewErrCode(errcode.DBRedisError, "redis errdef", err)
	}
	return nil
}

func (rm *RedisModule) GetString(key string) (string, chaoserrors.CError) {
	val, err := rm.client.Get(context.Background(), key).Result()
	if err != nil {
		return "", chaoserrors.NewErrCode(errcode.DBRedisError, "redis errdef", err)
	}
	return val, nil
}

func (rm *RedisModule) GetStringJson(key string, value interface{}) chaoserrors.CError {
	val, err := rm.client.Get(context.Background(), key).Result()
	if err != nil {
		return chaoserrors.NewErrCode(errcode.DBRedisError, "redis errdef", err)
	}

	if err := json.Unmarshal([]byte(val), value); err != nil {
		return chaoserrors.NewErrCode(errcode.ParamsInvalid, "not in json format", err)
	}
	return nil
}

// HSetStruct 保存结构体(结构体需要带redis标签)
func (rm *RedisModule) HSetStruct(key string, value interface{}) chaoserrors.CError {
	if err := rm.client.HSet(context.Background(), key, value).Err(); err != nil {
		return chaoserrors.NewErrCode(errcode.DBRedisError, "redis errdef", err)
	}

	return nil
}

// HGetStruct 获取结构体(结构体需要带redis标签)
func (rm *RedisModule) HGetStruct(key string, value interface{}) chaoserrors.CError {
	if err := rm.client.HGetAll(context.Background(), key).Scan(value); err != nil {
		return chaoserrors.NewErrCode(errcode.DBRedisError, "redis errdef", err)
	}
	return nil
}

// ExecuteFun 执行一个函数
func (rm *RedisModule) ExecuteFun(f Callback, args ...interface{}) (interface{}, chaoserrors.CError) {
	return f(rm.client, args...)
}
