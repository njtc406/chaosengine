package msg

import (
	"github.com/golang/protobuf/proto"
)

var MsgMap = map[int32]proto.Message{
	1:  &C2S_SignIn{},           // 登录
	2:  &C2S_DeleteAccount{},    // 注销账号
	10: &C2S_ServerList{},       // 请求服务器列表
	11: &C2S_AnnouncementList{}, // 请求公告列表
	12: &C2S_SelectServer{},     // 选择服务器
}
