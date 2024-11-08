// Package redismodule
// @Title  请填写文件名称（需要改）
// @Description  请填写文件描述（需要改）
// @Author  yr  2024/7/5 下午6:43
// @Update  yr  2024/7/5 下午6:43
package redismodule

import (
	"context"
	"errors"
	"fmt"
	"github.com/njtc406/chaosengine/engine1/errdef/errcode"
	"github.com/njtc406/chaosutil/chaoserrors"
	"github.com/redis/go-redis/v9"
)

type RedisBase struct{}

func (r *RedisBase) RedisKey(format string, args ...interface{}) string {
	return fmt.Sprintf(format, args...)
}

func (r *RedisBase) HGet(tx *redis.Tx, key string, field string) (bool, string, chaoserrors.CError) {
	if ret, err := tx.HGet(context.Background(), key, field).Result(); err != nil {
		if errors.Is(err, redis.Nil) {
			return false, "", nil
		}
		return false, "", chaoserrors.NewErrCode(errcode.DBRedisError, "redis查询错误", err)
	} else {
		return true, ret, nil
	}
}

func (r *RedisBase) HSet(tx *redis.Tx, key string, fields map[string]interface{}) chaoserrors.CError {
	if err := tx.HSet(context.Background(), key, fields).Err(); err != nil {
		return chaoserrors.NewErrCode(errcode.DBRedisError, "redis写入错误", err)
	}

	return nil
}

func (r *RedisBase) HGetAll(tx *redis.Client, key string, result interface{}) chaoserrors.CError {
	if err := tx.HGetAll(context.Background(), key).Scan(result); err != nil {
		return chaoserrors.NewErrCode(errcode.DBRedisError, "redis查询错误", err)
	}
	return nil
}
