// Package def
// @Title  公共结构定义
// @Description  公共结构定义
// @Author  yr  2024/7/19 上午10:42
// @Update  yr  2024/7/19 上午10:42
package def

// ServiceCommConf 服务公共配置
type ServiceCommConf struct {
	Id                  string `binding:""` // 服务id
	Name                string `binding:""` // 服务名称
	Goroutines          int32  `binding:""` // 协程数量
	OpenProfilerFlag    bool   `binding:""` // 是否开启性能分析器
	MaxEventChannelSize int    `binding:""` // 最大事件通道大小
}

func (s *ServiceCommConf) GetName() string {
	return s.Name
}

func (s *ServiceCommConf) SetName(name string) {
	s.Name = name
}

func (s *ServiceCommConf) GetID() string {
	return s.Id
}

func (s *ServiceCommConf) SetID(id string) {
	s.Id = id
}
