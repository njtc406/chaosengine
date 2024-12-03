// Package sdk
// @Title  功能名称
// @Description  功能描述
// @Author  yr  2024/5/16 下午3:32
// @Update  yr  2024/5/16 下午3:32
package sdk

import "github.com/njtc406/chaosengine/example/login/internal/service/login/inf"

var sdkHandler = map[string]inf.ISdk{}

func registerSDK(name string, sdk inf.ISdk) {
	sdkHandler[name] = sdk
}

func GetSDKHandler(name string) inf.ISdk {
	return sdkHandler[name]
}
