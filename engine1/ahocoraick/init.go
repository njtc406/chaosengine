// Package ahocorasick
// @Title
// @Description  ahocorasick
// @Author sly 2024/8/2
// @Created sly 2024/8/2
package ahocorasick

var (
	ACMatcher *Matcher
)

func Init(dict []string) {
	// 注册配置文件
	ACMatcher = NewStringMatcher(dict)
}
