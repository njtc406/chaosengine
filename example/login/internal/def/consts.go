// Package def
// @Title  title
// @Description  desc
// @Author  yr  2024/11/20
// @Update  yr  2024/11/20
package def

import (
	"net/http"
	"time"
)

const (
	LoginHttpModuleId uint32 = 1 // http模块id
)

const (
	DefaultHttpReqTimeout = time.Second
)

const (
	LoginTypeGuest = "guest"
)

const (
	DefaultHttpRespCode = http.StatusOK
	DefaultHttpRespMsg  = "ok"
)
