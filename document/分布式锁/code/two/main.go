package main

import (
	"fmt"
	RS "github.com/garyburd/redigo/redis"
	"github.com/gin-gonic/gin"
	uuid "github.com/satori/go.uuid"
	"log"
	"strconv"
	"sync"
	"two/redis"
)

var mutex sync.Mutex

func main() {
	r := gin.Default()
	gin.SetMode("release")
	redis := redis.InitRedis()

	r.GET("/hello", func(c *gin.Context) {
		mutex.Lock()
		defer mutex.Unlock()

		lockKey := "redisLock"
		clientID := uuid.NewV4().String()
		ok := redis.SetWitLock(lockKey, clientID, 10)
		if !ok {
			return
		}
		defer func() {
			// 删除自己所占用的锁， 看值是否一致，一致则删除，lua脚本实现
			script := "if redis.call('get',KEYS[1]) == ARGV[1] then return redis.call('del',KEYS[1]) else return 0  end"
			s := RS.NewScript(1, script)
			_, err := s.Do(redis.RS, lockKey, clientID)
			if err != nil {
				log.Fatal(err)
			}
		}()

		stack, _ := strconv.Atoi(redis.Get("num"))
		if stack > 0 {
			newStack := stack - 1
			res := redis.Set("num", newStack)
			if res {
				fmt.Println("库存修改完毕, 剩余库存：" + strconv.Itoa(newStack) + "-8060")
				return
			}
			fmt.Println("库存修改失败-8060")
			return
		}
		// 没有库存
		return
	})
	r.Run(":8060")
}
