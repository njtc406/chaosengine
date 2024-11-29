# 分布式服务框架

这是一个基于 Actor 模型的分布式服务框架，服务之间通过消息进行数据传递。框架特点：

- 服务默认单线程运行，支持多线程扩展。
- 结构灵活、模块化，便于开发和维护。

> 感谢 [github.com/duanhf2012/origin](https://github.com/duanhf2012/origin) 大佬提供的开源项目。本框架参考了其设计理念,做了一些适合自己项目的优化。

---

## 框架结构

本框架由以下三部分组成：

1. **Node（节点）**  
   每个节点对应一个独立进程，可以运行多个服务。

2. **Service（服务）**  
   每个服务是一个 Actor，以协程形式运行，支持多线程执行。服务由若干模块组成。

3. **Module（模块）**  
   模块是最小的业务单元，依附于服务运行，负责实现具体功能。

---

## 集群功能

1. **集群管理**  
   每个节点内置集群模块，用于发现和管理注册到集群中的服务。

2. **服务注册与发现**
    - 所有实现 `RPC` 开头接口的服务，都会自动注册到集群中。
    - 服务注册与发现通过 `etcd` 实现（还未做成接口,后续再说）。

3. **远程调用**
    - 基于 `rpcx` 实现跨节点远程调用。

### 当前实现

- 节点内和跨节点服务的调用功能。
- 基于 `etcd` 的服务发现机制。
- 配置可以远程,可以本地
---

### 可使用的环境变量
````
CHAOS_CONF_PATH = "./configs"                               // 本地配置文件路径(会优先使用环境变量)
CHAOS_ETCD_CONF_ENDPOINTS = "127.0.0.1:2379,127.0.0.1:2379" // etcd地址
CHAOS_ETCD_DIAL_TIMEOUT = 5s                                // 连接超时时间
CHAOS_ETCD_USERNAME = ""                                    // 用户名
CHAOS_ETCD_PASSWORD = ""                                    // 密码
CHAOS_ETCD_CONF_BASE_PATH = "/chaos/node/config/node.yaml"  // 配置文件路径
````
---

## 使用说明
熟悉origin的童鞋应该可以直接上手,这部分几乎没什么变化,origin的大佬已经做的很简化了
### 1. 启动一个节点

```go
package main
import "github.com/njtc406/chaosengine/engine/core/node"

var version = "1.0"         // 版本号
var configPath = "./configs/login" // 配置路径

func main() {
    // 可设置启动前钩子，例如日志初始化
    // node.SetStartHook(func() { ... })

    // 启动节点
    node.Start(version, configPath)
}
```

### 2. 增加一个服务

```go
package main
import (
    "github.com/njtc406/chaosengine/engine/core"
    "github.com/njtc406/chaosengine/engine/core/node"
    nodeConfig "github.com/njtc406/chaosengine/engine/core/node/config"
)

func init() {
    // 注册服务
    node.Setup(&MyService{})

    // 注册配置
    nodeConfig.RegisterConf(&nodeConfig.ConfInfo{
        ServiceName: "MyService",
        ConfName:    "myServiceConf",
        ConfPath:    "",
        ConfType:    "yaml",
        Cfg:         &config{},
    })
}

type config struct {
    a int
    b string
}

type MyService struct {
    core.Service
    data int
}

func (s *MyService) OnInit() error {
    return nil
}

func (s *MyService) OnStart() error {
    return nil
}

func (s *MyService) OnRelease() {}
```

3. 增加一个模块
```go
package main
import (
    "github.com/njtc406/chaosengine/engine/core"
)

type MyModule struct {
    core.Module
}

func (m *MyModule) OnInit() error {
    return nil
}

func (m *MyModule) OnRelease() {}

func (s *MyService) OnInit() error {
    module := &MyModule{}
    // 添加模块到服务
    moduleID, err := s.AddModule(module)
    if err != nil {
        return err
    }
    fmt.Printf("Module ID: %d\n", moduleID)
    return nil
}

func (s *MyService) OnRelease() {
    s.ReleaseAllChildModule()
}
```
