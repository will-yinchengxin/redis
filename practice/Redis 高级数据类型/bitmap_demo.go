package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

// BitmapDemo 演示 Bitmap 的各种使用场景
type BitmapDemo struct {
	client *redis.Client
	ctx    context.Context
}

// NewBitmapDemo 创建 Bitmap 演示实例
func NewBitmapDemo(client *redis.Client) *BitmapDemo {
	return &BitmapDemo{
		client: client,
		ctx:    context.Background(),
	}
}

// Example1_BasicUsage 基础使用示例
func (b *BitmapDemo) Example1_BasicUsage() {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("Bitmap 示例 1: 基础使用")
	fmt.Println(strings.Repeat("=", 60))

	key := "bitmap:demo:basic"
	b.client.Del(b.ctx, key)

	// 设置某些位为 1（表示用户签到）
	userIDs := []int64{1, 3, 5, 7, 10, 15, 20}
	fmt.Println("标记以下用户 ID 为活跃:")
	for _, userID := range userIDs {
		b.client.SetBit(b.ctx, key, userID, 1)
		fmt.Printf("  用户 %d\n", userID)
	}

	// 检查某个位是否为 1
	fmt.Println("\n检查用户状态:")
	checkUsers := []int64{1, 2, 5, 8, 10}
	for _, userID := range checkUsers {
		isActive, _ := b.client.GetBit(b.ctx, key, userID).Result()
		status := "未活跃"
		if isActive == 1 {
			status = "活跃"
		}
		fmt.Printf("  用户 %d: %s\n", userID, status)
	}

	// 统计有多少个 1（活跃用户数）
	count, _ := b.client.BitCount(b.ctx, key, nil).Result()
	fmt.Printf("\n总活跃用户数: %d\n", count)

	fmt.Println("\n✓ Bitmap 通过位操作实现高效的布尔值存储")
}

// Example2_DailySignIn 每日签到示例
func (b *BitmapDemo) Example2_DailySignIn() {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("Bitmap 示例 2: 每日签到功能")
	fmt.Println(strings.Repeat("=", 60))

	userID := int64(10086)
	today := time.Now()

	// 模拟用户在 1 月份的签到情况
	fmt.Printf("模拟用户 %d 在 1 月份的签到情况...\n", userID)

	// 用户在这些天签到了
	signInDays := []int{1, 2, 3, 5, 7, 10, 15, 20, 25, 28, 30}

	for _, day := range signInDays {
		date := time.Date(today.Year(), today.Month(), day, 0, 0, 0, 0, time.Local)
		key := fmt.Sprintf("signin:%d:%s", userID, date.Format("2006-01"))
		offset := int64(day - 1) // 第 1 天对应 offset 0
		b.client.SetBit(b.ctx, key, offset, 1)
	}

	// 查询用户 1 月份的签到情况
	key := fmt.Sprintf("signin:%d:%s", userID, today.Format("2006-01"))

	fmt.Println("\n签到日历:")
	for day := 1; day <= 31; day++ {
		offset := int64(day - 1)
		signed, _ := b.client.GetBit(b.ctx, key, offset).Result()
		if signed == 1 {
			fmt.Printf("%2d日: ✓ 已签到\n", day)
		}
	}

	// 统计本月签到天数
	signInCount, _ := b.client.BitCount(b.ctx, key, nil).Result()
	fmt.Printf("\n本月签到天数: %d 天\n", signInCount)

	// 检查今天是否已签到
	todayOffset := int64(today.Day() - 1)
	signedToday, _ := b.client.GetBit(b.ctx, key, todayOffset).Result()
	if signedToday == 1 {
		fmt.Println("今日状态: 已签到 ✓")
	} else {
		fmt.Println("今日状态: 未签到")
	}

	// 计算连续签到天数
	consecutiveDays := b.calculateConsecutiveDays(key, int(today.Day()))
	fmt.Printf("连续签到: %d 天\n", consecutiveDays)

	fmt.Println("\n✓ Bitmap 非常适合签到这种每日布尔值记录")
}

// Example3_UserActivity 用户活跃统计示例
func (b *BitmapDemo) Example3_UserActivity() {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("Bitmap 示例 3: 用户活跃度统计")
	fmt.Println(strings.Repeat("=", 60))

	// 模拟 7 天的用户活跃数据
	fmt.Println("模拟生成 7 天的用户活跃数据...")

	baseDate := time.Now().AddDate(0, 0, -6)
	days := []string{}

	for i := 0; i < 7; i++ {
		date := baseDate.AddDate(0, 0, i)
		dateStr := date.Format("2006-01-02")
		days = append(days, dateStr)
		key := fmt.Sprintf("bitmap:active:%s", dateStr)

		b.client.Del(b.ctx, key)

		// 每天随机 50-80 个用户活跃（用户 ID 范围 1-100）
		activeCount := 50 + i*5
		for uid := 1; uid <= activeCount; uid++ {
			b.client.SetBit(b.ctx, key, int64(uid), 1)
		}
	}

	// 统计每天的活跃用户数
	fmt.Println("\n每日活跃用户数(DAU):")
	for _, day := range days {
		key := fmt.Sprintf("bitmap:active:%s", day)
		count, _ := b.client.BitCount(b.ctx, key, nil).Result()
		fmt.Printf("%s: %d 人\n", day, count)
	}

	// 计算 7 天都活跃的用户（使用 AND 操作）
	fmt.Println("\n计算 7 天都活跃的用户(留存用户)...")
	destKey := "bitmap:active:7days:all"

	keys := make([]string, len(days))
	for i, day := range days {
		keys[i] = fmt.Sprintf("bitmap:active:%s", day)
	}

	// BitOpAnd: 所有位都为 1 的用户
	b.client.BitOpAnd(b.ctx, destKey, keys...)
	allActivCount, _ := b.client.BitCount(b.ctx, destKey, nil).Result()
	fmt.Printf("7 天都活跃的用户: %d 人\n", allActivCount)

	// 计算至少活跃 1 天的用户（使用 OR 操作）
	destKeyOr := "bitmap:active:7days:any"
	b.client.BitOpOr(b.ctx, destKeyOr, keys...)
	anyActiveCount, _ := b.client.BitCount(b.ctx, destKeyOr, nil).Result()
	fmt.Printf("至少活跃 1 天的用户: %d 人\n", anyActiveCount)

	// 计算留存率
	retention := float64(allActivCount) / float64(anyActiveCount) * 100
	fmt.Printf("7 日留存率: %.2f%%\n", retention)

	fmt.Println("\n✓ Bitmap 的位运算非常适合做用户集合分析")
}

// Example4_ABTesting A/B 测试示例
func (b *BitmapDemo) Example4_ABTesting() {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("Bitmap 示例 4: A/B 测试分组")
	fmt.Println(strings.Repeat("=", 60))

	keyA := "bitmap:test:groupA"
	keyB := "bitmap:test:groupB"

	b.client.Del(b.ctx, keyA, keyB)

	// 将用户分配到 A/B 组
	// A 组：用户 ID 1-50
	// B 组：用户 ID 51-100
	fmt.Println("分配用户到 A/B 测试组...")

	for uid := int64(1); uid <= 50; uid++ {
		b.client.SetBit(b.ctx, keyA, uid, 1)
	}

	for uid := int64(51); uid <= 100; uid++ {
		b.client.SetBit(b.ctx, keyB, uid, 1)
	}

	countA, _ := b.client.BitCount(b.ctx, keyA, nil).Result()
	countB, _ := b.client.BitCount(b.ctx, keyB, nil).Result()

	fmt.Printf("A 组用户数: %d\n", countA)
	fmt.Printf("B 组用户数: %d\n", countB)

	// 检查某个用户在哪个组
	fmt.Println("\n检查用户分组:")
	testUsers := []int64{10, 25, 60, 80}
	for _, uid := range testUsers {
		inA, _ := b.client.GetBit(b.ctx, keyA, uid).Result()
		inB, _ := b.client.GetBit(b.ctx, keyB, uid).Result()

		group := "未分组"
		if inA == 1 {
			group = "A 组"
		} else if inB == 1 {
			group = "B 组"
		}
		fmt.Printf("  用户 %d: %s\n", uid, group)
	}

	// 模拟转化数据
	keyAConverted := "bitmap:test:groupA:converted"
	keyBConverted := "bitmap:test:groupB:converted"

	b.client.Del(b.ctx, keyAConverted, keyBConverted)

	// A 组 20% 转化，B 组 25% 转化
	for uid := int64(1); uid <= 10; uid++ {
		b.client.SetBit(b.ctx, keyAConverted, uid, 1)
	}
	for uid := int64(51); uid <= 63; uid++ {
		b.client.SetBit(b.ctx, keyBConverted, uid, 1)
	}

	convertedA, _ := b.client.BitCount(b.ctx, keyAConverted, nil).Result()
	convertedB, _ := b.client.BitCount(b.ctx, keyBConverted, nil).Result()

	fmt.Println("\n转化数据:")
	fmt.Printf("A 组转化: %d/%d (%.1f%%)\n", convertedA, countA, float64(convertedA)/float64(countA)*100)
	fmt.Printf("B 组转化: %d/%d (%.1f%%)\n", convertedB, countB, float64(convertedB)/float64(countB)*100)

	fmt.Println("\n✓ Bitmap 可以高效管理 A/B 测试用户分组")
}

// Example5_UserPermissions 用户权限管理示例
func (b *BitmapDemo) Example5_UserPermissions() {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("Bitmap 示例 5: 用户权限管理")
	fmt.Println(strings.Repeat("=", 60))

	// 定义权限位
	permissions := map[string]int64{
		"读取": 0,
		"写入": 1,
		"删除": 2,
		"管理": 3,
		"审核": 4,
		"导出": 5,
		"分享": 6,
		"评论": 7,
	}

	userID := int64(1001)
	key := fmt.Sprintf("bitmap:permissions:user:%d", userID)
	b.client.Del(b.ctx, key)

	// 授予用户一些权限
	grantedPerms := []string{"读取", "写入", "评论"}
	fmt.Printf("授予用户 %d 以下权限:\n", userID)
	for _, perm := range grantedPerms {
		offset := permissions[perm]
		b.client.SetBit(b.ctx, key, offset, 1)
		fmt.Printf("  ✓ %s\n", perm)
	}

	// 检查用户权限
	fmt.Println("\n检查用户权限:")
	for perm, offset := range permissions {
		hasPermission, _ := b.client.GetBit(b.ctx, key, offset).Result()
		status := "✗ 无权限"
		if hasPermission == 1 {
			status = "✓ 有权限"
		}
		fmt.Printf("  %s: %s\n", perm, status)
	}

	// 撤销权限
	fmt.Println("\n撤销 '写入' 权限...")
	b.client.SetBit(b.ctx, key, permissions["写入"], 0)

	// 再次检查
	hasWrite, _ := b.client.GetBit(b.ctx, key, permissions["写入"]).Result()
	if hasWrite == 0 {
		fmt.Println("  ✓ 写入权限已撤销")
	}

	fmt.Println("\n✓ Bitmap 可以用位来表示不同的权限，非常高效")
}

// Example6_MemoryComparison 内存对比示例
func (b *BitmapDemo) Example6_MemoryComparison() {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("Bitmap 示例 6: 内存效率对比")
	fmt.Println(strings.Repeat("=", 60))

	numUsers := 100000
	fmt.Printf("存储 %d 个用户的活跃状态...\n\n", numUsers)

	// 方案 1: 使用 Set
	setKey := "compare:set:users"
	b.client.Del(b.ctx, setKey)

	startTime := time.Now()
	for uid := 1; uid <= numUsers; uid++ {
		b.client.SAdd(b.ctx, setKey, uid)
	}
	setDuration := time.Since(startTime)
	setMem := b.client.MemoryUsage(b.ctx, setKey).Val()
	setCount := b.client.SCard(b.ctx, setKey).Val()

	// 方案 2: 使用 Bitmap
	bitmapKey := "compare:bitmap:users"
	b.client.Del(b.ctx, bitmapKey)

	startTime = time.Now()
	for uid := 1; uid <= numUsers; uid++ {
		b.client.SetBit(b.ctx, bitmapKey, int64(uid), 1)
	}
	bitmapDuration := time.Since(startTime)
	bitmapMem := b.client.MemoryUsage(b.ctx, bitmapKey).Val()
	bitmapCount := b.client.BitCount(b.ctx, bitmapKey, nil).Val()

	// 方案 3: 使用 HyperLogLog
	hllKey := "compare:hll:users"
	b.client.Del(b.ctx, hllKey)

	startTime = time.Now()
	for uid := 1; uid <= numUsers; uid++ {
		b.client.PFAdd(b.ctx, hllKey, uid)
	}
	hllDuration := time.Since(startTime)
	hllMem := b.client.MemoryUsage(b.ctx, hllKey).Val()
	hllCount := b.client.PFCount(b.ctx, hllKey).Val()

	// 输出对比结果
	fmt.Println("【Set 方案】")
	fmt.Printf("  计数结果: %d (精确)\n", setCount)
	fmt.Printf("  内存占用: %d 字节 (%.2f MB)\n", setMem, float64(setMem)/1024/1024)
	fmt.Printf("  写入耗时: %v\n", setDuration)

	fmt.Println("\n【Bitmap 方案】")
	fmt.Printf("  计数结果: %d (精确)\n", bitmapCount)
	fmt.Printf("  内存占用: %d 字节 (%.2f KB)\n", bitmapMem, float64(bitmapMem)/1024)
	fmt.Printf("  写入耗时: %v\n", bitmapDuration)

	fmt.Println("\n【HyperLogLog 方案】")
	fmt.Printf("  计数结果: %d (估算, 误差 %d)\n", hllCount, abs64(hllCount-int64(numUsers)))
	fmt.Printf("  内存占用: %d 字节 (%.2f KB)\n", hllMem, float64(hllMem)/1024)
	fmt.Printf("  写入耗时: %v\n", hllDuration)

	fmt.Println("\n【对比总结】")
	fmt.Printf("  Bitmap 比 Set 节省内存: %.1f%%\n",
		(1-float64(bitmapMem)/float64(setMem))*100)
	fmt.Printf("  HyperLogLog 比 Bitmap 节省内存: %.1f%%\n",
		(1-float64(hllMem)/float64(bitmapMem))*100)

	fmt.Println("\n✓ 选择建议:")
	fmt.Println("  - 需要精确值 + 需要元素列表 → Set")
	fmt.Println("  - 需要精确值 + 用户ID连续 → Bitmap")
	fmt.Println("  - 只需计数 + 可接受误差 → HyperLogLog")
}

// Example7_OnlineUsers 在线用户统计示例
func (b *BitmapDemo) Example7_OnlineUsers() {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("Bitmap 示例 7: 实时在线用户统计")
	fmt.Println(strings.Repeat("=", 60))

	// 按小时统计在线用户
	now := time.Now()

	fmt.Println("模拟今日各小时在线用户...")
	for hour := 0; hour < 24; hour++ {
		key := fmt.Sprintf("bitmap:online:%s:hour:%02d",
			now.Format("2006-01-02"), hour)
		b.client.Del(b.ctx, key)

		// 模拟在线用户（白天多，晚上少）
		baseUsers := 1000
		if hour >= 9 && hour <= 22 {
			baseUsers = 5000
		}

		for uid := 1; uid <= baseUsers; uid++ {
			b.client.SetBit(b.ctx, key, int64(uid), 1)
		}

		// 设置过期时间（24 小时后自动删除）
		b.client.Expire(b.ctx, key, 24*time.Hour)
	}

	// 查看各时段在线人数
	fmt.Println("\n今日在线用户统计:")
	for hour := 0; hour < 24; hour++ {
		key := fmt.Sprintf("bitmap:online:%s:hour:%02d",
			now.Format("2006-01-02"), hour)
		count, _ := b.client.BitCount(b.ctx, key, nil).Result()

		bar := strings.Repeat("█", int(count/200))
		fmt.Printf("%02d:00 - %5d 人 %s\n", hour, count, bar)
	}

	// 计算今日活跃用户（任意时段在线）
	fmt.Println("\n计算今日总活跃用户...")
	keys := make([]string, 24)
	for hour := 0; hour < 24; hour++ {
		keys[hour] = fmt.Sprintf("bitmap:online:%s:hour:%02d",
			now.Format("2006-01-02"), hour)
	}

	destKey := fmt.Sprintf("bitmap:dau:%s", now.Format("2006-01-02"))
	b.client.BitOpOr(b.ctx, destKey, keys...)
	dau, _ := b.client.BitCount(b.ctx, destKey, nil).Result()

	fmt.Printf("今日活跃用户(DAU): %d 人\n", dau)

	fmt.Println("\n✓ Bitmap 可以高效统计分时段的在线用户")
}

// calculateConsecutiveDays 计算连续签到天数
func (b *BitmapDemo) calculateConsecutiveDays(key string, currentDay int) int {
	consecutive := 0
	for day := currentDay; day >= 1; day-- {
		offset := int64(day - 1)
		signed, _ := b.client.GetBit(b.ctx, key, offset).Result()
		if signed == 1 {
			consecutive++
		} else {
			break
		}
	}
	return consecutive
}

// RunAllExamples 运行所有示例
func (b *BitmapDemo) RunAllExamples() {
	fmt.Println("\n🚀 开始运行 Bitmap 所有示例...")

	b.Example1_BasicUsage()
	b.Example2_DailySignIn()
	b.Example3_UserActivity()
	b.Example4_ABTesting()
	b.Example5_UserPermissions()
	b.Example6_MemoryComparison()
	b.Example7_OnlineUsers()

	fmt.Println("\n✅ Bitmap 所有示例运行完成！")
}

// 辅助函数：计算绝对值（int64）
func abs64(n int64) int64 {
	if n < 0 {
		return -n
	}
	return n
}
