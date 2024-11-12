// Package repository
// @Title  title
// @Description  desc
// @Author  yr  2024/11/11
// @Update  yr  2024/11/11
package repository

// TODO 快速查询模块,现在仓库用的是map直接存储信息,后续可以优化为redis
// 直接使用map存储不太好扩展,一旦出现新东西就要改,只是说目前的设计已经覆盖了大多数场景,
// 但是依然不是很好用,所以我想要改成redis,直接将map的key换成redis的key,当出现多个时,直接scan或者key一些就都拿到了,效率应该也很高
// 但是使用redis的话又有几个问题,第一延迟,这里实际上是节点选择器使用的,如果这里慢,意味着整个调用链路都慢了,第二重连,如果redis挂了
// 会影响到整个服务调用链路, 还是在好好想想吧
