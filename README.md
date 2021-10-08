# go-rabbit-pool
纯Go写的RabbitMQ连接池，做了基本的测试，支持`MaxLive`,`MaxIdle`,`MinAlive`设置，其他设置也可以通`options`修改

 
## 使用方法



### simple
```


var (
	minAlive = 4
	maxAlive = 16
	maxIdle  = 8
)

pool := NewPool(*url, MaxAlive(maxAlive), MaxIdle(maxIdle), MinAlive(minAlive))


// 获取连接实例，会被阻塞如果没有更多的连接可以获取 
// TODO 超时设置
con,err := pool.Acquire()


// 释放一个连接实例
err := pool.Release(con)

```


### idle 超时设置

pool 实例通过`options`进行设置，pool支持最大idle设置，会自动释放连接，保持`MaxIdle`连接数量