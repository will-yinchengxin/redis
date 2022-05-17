# Redis 变慢了，如何快速排查？
## 一. 设置慢查询日志
### 查看redis慢日志,redis提供了慢日志的统计功能
首先设置Redis的慢日志阈值，只有超过阈值的命令才会被记录，这里的单位是微妙，例如设置慢日志的阈值为5毫秒，同时设置只保留最近1000条慢日志记录：
````
# 命令查询超过5毫秒记录慢日志
CONFIG SET slowlog-log-slower-than 5000
# 只保留近1000条慢日志
CONFIG SET slowlog-max-len 1000
````
设置完成后,使用 `SLOWLOG get 5` 产看最近5条慢日志
## 二. 扫描大key
Redis也提供了扫描大key的方法：`redis-cli -h $host -p $port --bigkeys -i 0.01`
使用上述名利可以扫描出整个实例key大小的分布情况,它是以类型维度来展示的

当我们再线上使用大key扫描时,redis的QPS会快速突增,为了降低扫描过程对redis的影响,我们需要控制扫描的频率,使用-i控制即可
它表示扫描的时间间隔,这个命令其实就是scan名利,遍历所有key，然后针对不同类型的key
执行`strlen`、`llen`、`hlen`、`scard`、`zcard`来获取字符串的长度以及容器类型(list/dict/set/zset)的元素个数。
## 三. 解决集中过期
解决方案是，在集中过期时增加一个随机时间，把这些需要过期的key的时间打散即可。
伪代码: `redis.expireat(key, expire_time + random(300))`