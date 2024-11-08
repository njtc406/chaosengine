// Package utils
// @Title  请填写文件名称（需要改）
// @Description  请填写文件描述（需要改）
// @Author  yr  2024/9/3 下午4:13
// @Update  yr  2024/9/3 下午4:13
package utils

import (
	"github.com/njtc406/chaosengine/engine1/errdef/errcode"
	"github.com/njtc406/chaosutil/chaoserrors"
	"strings"
)

// SplitServiceMethod 拆分服务名和方法
func SplitServiceMethod(serviceMethod string) (string, string, chaoserrors.CError) {
	findIndex := strings.Index(serviceMethod, ".")
	if findIndex == -1 {
		return "", "", chaoserrors.NewErrCode(errcode.ServiceMethodError, "format error")
	}

	return serviceMethod[:findIndex], serviceMethod[findIndex+1:], nil
}
