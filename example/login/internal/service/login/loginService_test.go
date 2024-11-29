package login

import (
	"fmt"
	"github.com/njtc406/chaosengine/engine/utils/httplib"
	"math/rand"
	"net/http"
	"server/glob/dto"
	msg "server/msg/login"
	"testing"
)

func TestMsg_CS_SignIn(t *testing.T) {
	respData := dto.NewResponse()
	signResp := &msg.S2C_SignIn{}
	respData.SetData(signResp)
	defer dto.Release(respData)
	if err := httplib.Request(http.MethodPost, "tigergame.imwork.net:20009", "/api/v1/auth", &msg.C2S_SignIn{
		UserName: "111111",
		Type:     "guest",
	}, respData); err != nil {
		t.Error(err)
	}

	t.Logf("respData: %+v", respData)

	// 通过token获取服务器列表,公告列表,选择服务器
	serverListRespData := dto.NewResponse()
	if err := httplib.Request(http.MethodPost, "tigergame.imwork.net:20009", fmt.Sprintf("/api/v1/serverList?_=%d", rand.Int31()), &msg.C2S_ServerList{
		Token: signResp.Token,
	}, serverListRespData); err != nil {
		t.Error(err)
	}

	t.Logf("serverListRespData: %+v", serverListRespData)
	dto.Release(serverListRespData)

	announcementListRespData := dto.NewResponse()
	if err := httplib.Request(http.MethodPost, "tigergame.imwork.net:20009", fmt.Sprintf("/api/v1/announcementList?_=%d", rand.Int31()), &msg.C2S_AnnouncementList{
		Token: signResp.Token,
	}, announcementListRespData); err != nil {
		t.Error(err)
	}

	t.Logf("announcementListRespData: %+v", announcementListRespData)
	dto.Release(announcementListRespData)

	selectServerRespData := dto.NewResponse()
	if err := httplib.Request(http.MethodPost, "tigergame.imwork.net:20009", fmt.Sprintf("/api/v1/selectServer?_=%d", rand.Int31()), &msg.C2S_SelectServer{
		ServerId: 1,
		Token:    signResp.Token,
	}, selectServerRespData); err != nil {
		t.Error(err)
	}

	t.Logf("selectServerRespData: %+v", selectServerRespData)
	dto.Release(selectServerRespData)
}
