# GoCache

## 特性

* 单机缓存和基于HTTP的分布式缓存
* 最近最少访问缓存策略
* 使用Go语言互斥锁防止缓存击穿
* 使用一致性哈希选择节点，实现负载均衡

## 代码主逻辑

```tex
                            是
接收 key --> 检查是否被缓存 -----> 返回缓存值 ⑴
                |  否                         是
                |-----> 是否应当从远程节点获取 -----> 与远程节点交互 --> 返回缓存值 ⑵
                            |  否
                            |-----> 调用`回调函数`，获取值并添加到缓存 --> 返回缓存值 ⑶
```

## 主要数据结构

```go
//查询缓存的主要结构
type Group struct {
	name      string // group名称
	getter    Getter // 本机找不到缓存的时候，执行该回调，例如客户端进行数据库读取然后返回值
	mainCache cache // 缓存结构
	peers     PeerPicker // 选择对应的服务器进行远程获取缓存
	// use singleflight.Group to make sure that
	// each key is only fetched once
	loader *singleflight.Group 
}
```

```go
//用于实现一致性哈希
type Map struct {
	hash     Hash // 哈希函数
	replicas int // 虚节点
	keys     []int // 一致性哈希
	hashMap  map[int]string // 虚拟节点对应的真实ip
}
```

```go
//对应缓存请求
type call struct {
	wg  sync.WaitGroup
	val interface{}
	err error
}

//用于解决缓存击穿
type Group struct {
	mu sync.Mutex       // protects m
	m  map[string]*call //每个group对应一个m，对于多个相同call请求只执行一个，返回这个call的值给客户
}
```
