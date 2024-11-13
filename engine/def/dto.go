// Package def
// @Title  title
// @Description  desc
// @Author  yr  2024/11/13
// @Update  yr  2024/11/13
package def

import "fmt"

func GenNodeUid(nodeId int32, nodeType string) string {
	return fmt.Sprintf("%s_%d", nodeType, nodeId)
}
