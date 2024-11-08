// Package def
// @Title  title
// @Description  desc
// @Author  yr  2024/11/8
// @Update  yr  2024/11/8
package def

import "reflect"

type MethodInfo struct {
	method reflect.Method
	in     reflect.Type
	out    reflect.Type
}
