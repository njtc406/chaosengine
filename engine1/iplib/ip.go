// Package iplib
// @Title  获取服务器ip
// @Description  获取服务器ip
// @Author  yr  2024/6/13 下午6:22
// @Update  yr  2024/6/13 下午6:22
package iplib

import (
	"os/exec"
	"strings"
)

// GetPublicIP 获取公网ip
func GetPublicIP() (string, error) {
	cmd := exec.Command("curl", "ifconfig.me")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}
