package main

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	_ "net/http/pprof"
	"sync"
	"time"
)
import redigo "github.com/gomodule/redigo/redis"

//上锁脚本 k1 锁名 a1 生存时间 毫秒 a2 锁ID
const LOCK_SCRIPT = `if (redis.call('exists', KEYS[1]) == 0) then 
	redis.call('hincrby', KEYS[1], ARGV[2], 1); 
	redis.call('pexpire', KEYS[1], ARGV[1]); 
	return nil; 
	end; 
	if (redis.call('hexists', KEYS[1], ARGV[2]) == 1) then 
	redis.call('hincrby', KEYS[1], ARGV[2], 1); 
	redis.call('pexpire', KEYS[1], ARGV[1]); 
	return nil; 
	end; 
	return redis.call('pttl', KEYS[1]);`

//订阅事件前缀
const SUBSCRIBE_PRE = "msg_"

//续期时间
const RENEWAL_TIME = 10

//续期脚本 K1 锁名 a1 续期时间 a2 锁ID
const RENEWAL_SCRIPT = "if (redis.call('hexists', KEYS[1], ARGV[2]) == 1) then " +
	"redis.call('pexpire', KEYS[1], ARGV[1]); " +
	"return 1; " +
	"end; " +
	"return 0;"

//解锁脚本 k1 锁名 k2 发布事件名 a1 发布事件内容 a2 重入锁释放后上级锁的续期时间 a3 锁ID
const UNLOCK_SCRIPT = "if (redis.call('hexists', KEYS[1], ARGV[3]) == 0) then " +
	"return nil;" +
	"end; " +
	"local counter = redis.call('hincrby', KEYS[1], ARGV[3], -1); " +
	"if (counter > 0) then " +
	"redis.call('pexpire', KEYS[1], ARGV[2]); " +
	"return 0; " +
	"else " +
	"redis.call('del', KEYS[1]); " +
	"redis.call('publish', KEYS[2], ARGV[1]); " +
	"return 1; " +
	"end; " +
	"return nil;"

type RedisLock struct {
	LockName string
	Timeout int64
	TimeWait int64
	LockId string
}

var pool *redigo.Pool

var mut sync.Mutex

var group sync.WaitGroup

var t int


func main() {
	//test code 测试代码
}

//可以开启多个协程对j++
func test(name string,j *int){
	c := pool.Get()
	a := RedisLock{
		LockName: "zp",
		LockId: "",
		Timeout: 60000, //上锁后锁超时时间
		TimeWait: 50000,//等待锁释放时间，
	}
	a.Lock(c)
	for i := 0; i < 100000; i++ {
		*j++
	}
	a.Lock(c)
	for i := 0; i < 100000; i++ {
		*j++
	}
	//time.Sleep(time.Millisecond*1000)
	a.UnLock(c)
	for i := 0; i < 100000; i++ {
		*j++
	}
	a.UnLock(c)

	c.Close()
	t--
	group.Done()
}

// redis pool
func PoolInitRedis(server string, password string) *redigo.Pool {
	return &redigo.Pool{
		MaxIdle:     20, //空闲数
		IdleTimeout: 240 * time.Second,
		MaxActive:   200, //最大数
		Dial: func() (redigo.Conn, error) {
			c, err := redigo.Dial("tcp", server)
			if err != nil {
				return nil, err
			}
			if password != "" {
				if _, err := c.Do("AUTH", password); err != nil {
					c.Close()
					return nil, err
				}
			}
			return c, err
		},
		TestOnBorrow: func(c redigo.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
}

//返回true 表示获取到锁，返回false 如果timewait为0，则表示尝试获取一次锁失败 否则表示彻底获取锁失败
func (r *RedisLock) Lock(conn redigo.Conn) (bool, error) {
	if r.LockId == "" {
		err := r.GetUuid()
		if err != nil {
			return false,err
		}
	}
	if r.TimeWait == 0 {
		do, err := conn.Do("EVAL", LOCK_SCRIPT, 1, r.LockName, r.Timeout, r.LockId)
		if err != nil {
			return false, err
		}
		if do == nil {
			return true, nil
		}
		if r.TimeWait == 0 {
			return false, nil
		}
	}
	//当前时间
	currency := time.Now().UnixMilli()
	last := currency + r.TimeWait
	for time.Now().UnixMilli() - r.TimeWait <currency {
		time.Sleep(100*time.Microsecond)
		//首先尝试获取锁
		do, err := conn.Do("EVAL", LOCK_SCRIPT, 1, r.LockName, r.Timeout, r.LockId)
		if err != nil {
			return false, err
		}
		if do == nil {
			return true, nil
		}
		if r.TimeWait == 0 {
			return false, nil
		}
		//没有获取到锁，则进行订阅释放锁事件
		r.SubscribeUnlock(conn,last-time.Now().UnixMilli())
		//time.Sleep(100*time.Microsecond)
	}
	return false, errors.New("time out")
}

func (r *RedisLock) GetUuid() (error) {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return err
	}
	r.LockId = fmt.Sprintf("%x-%x-%x-%x-%x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
	return nil
}

//订阅锁释放事件 redis链接 锁名 订阅超时时间
func (r *RedisLock) SubscribeUnlock(conn redigo.Conn, timeout int64) (bool, error) {
	if timeout <= 0 {
		return false,nil
	}
	ctx, cancelFunc := context.WithTimeout(context.Background(), (time.Millisecond)*time.Duration(timeout))
	defer cancelFunc()
	done := make(chan error, 1)
	psc := redigo.PubSubConn{
		Conn: pool.Get(),
	}
	if err := psc.Subscribe(SUBSCRIBE_PRE + r.LockName); err != nil {
		return false, err
	}
	go func() {
		for {
			switch msg := psc.Receive().(type) {
			case error:
				fmt.Println(msg)
				done <- fmt.Errorf("redis pubsub receive err: %v", msg)
				return
			case redigo.Message:
				done <- nil
				return
			case redigo.Subscription:
				if msg.Count == 0 {
					// all channels are unsubscribed
					done <- errors.New("all channels are unsubscribed")
					return
				}
			}
		}
	}()

	// health check
	tick := time.NewTicker(time.Second)
	defer tick.Stop()
	for {
		select {
		case <-ctx.Done():
			if err := psc.Unsubscribe(); err != nil {
				return false,err
			}
			return false,errors.New("timeout,unsubscribe")
		case err := <-done:
			if err != nil {
				return false,err
			}else {
				return true,nil
			}
		case <-tick.C:
			if err := psc.Ping(""); err != nil {
				return false,errors.New("subscribe heart error")
			}
		default:
			//判断等待的锁是否还存在
			i, err := redigo.Int(conn.Do("exists", r.LockName))
			if err == nil&&i == 0 {
				return true,nil
			}
		}
	}
	return false, nil
}

func (r *RedisLock) UnLock(conn redigo.Conn) (bool,error) {
	//k1 锁名 k2 发布事件名 a1 发布事件内容 a2 重入锁释放后上级锁的续期时间 a3 锁ID
	do, err := conn.Do("EVAL", UNLOCK_SCRIPT, 2, r.LockName, SUBSCRIBE_PRE+r.LockName, "unlock", r.Timeout, r.LockId)
	if err != nil {
		return false,err
	}
	switch do.(type) {
	case int64 :
		return true,nil
	default:
		return false,errors.New("no you lock")
	}
	return false,nil
}

//检查是否还持有锁
func (r *RedisLock) LockState(conn redigo.Conn) (bool,error) {
	do, err := conn.Do("hget", r.LockName, r.LockId)
	if err != nil {
		return false,err
	}
	if do != nil {
		return true,nil
	}else{
		return false,nil
	}
}
