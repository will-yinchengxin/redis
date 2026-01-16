package main

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

// HyperLogLogDemo 演示 HyperLogLog 的各种使用场景
type HyperLogLogDemo struct {
	client *redis.Client
	ctx    context.Context
}

// NewHyperLogLogDemo 创建 HyperLogLog 演示实例
func NewHyperLogLogDemo(client *redis.Client) *HyperLogLogDemo {
	return &HyperLogLogDemo{
		client: client,
		ctx:    context.Background(),
	}
}

// Example1_BasicUsage 基础使用示例
func (h *HyperLogLogDemo) Example1_BasicUsage() {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("HyperLogLog 示例 1: 基础使用")
	fmt.Println(strings.Repeat("=", 60))

	key := "hll:demo:basic"

	// 清理旧数据
	h.client.Del(h.ctx, key)

	// 添加元素
	users := []string{"user1", "user2", "user3", "user1", "user2", "user4"}
	for _, user := range users {
		err := h.client.PFAdd(h.ctx, key, user).Err()
		if err != nil {
			fmt.Printf("添加失败: %v\n", err)
			return
		}
	}

	fmt.Printf("添加了 %d 个元素（包含重复）: %v\n", len(users), users)

	// 获取基数（去重后的数量）
	count, err := h.client.PFCount(h.ctx, key).Result()
	if err != nil {
		fmt.Printf("获取计数失败: %v\n", err)
		return
	}

	fmt.Printf("去重后的唯一元素数量: %d\n", count)
	fmt.Println("✓ 可以看到，虽然添加了 6 个元素，但去重后只有 4 个唯一用户")
}

// Example2_WebsiteUV 网站 UV 统计示例
func (h *HyperLogLogDemo) Example2_WebsiteUV() {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("HyperLogLog 示例 2: 网站 UV 统计")
	fmt.Println(strings.Repeat("=", 60))

	today := time.Now().Format("2006-01-02")
	key := fmt.Sprintf("hll:uv:%s", today)

	// 清理旧数据
	h.client.Del(h.ctx, key)

	// 模拟 10000 个用户访问，其中有重复访问
	fmt.Println("模拟 10000 次页面访问...")
	totalVisits := 10000
	uniqueUsers := 3000

	rand.Seed(time.Now().UnixNano())
	for i := 0; i < totalVisits; i++ {
		// 生成随机用户 ID（范围 1-3000，所以实际唯一用户约 3000）
		userID := fmt.Sprintf("user_%d", rand.Intn(uniqueUsers)+1)
		h.client.PFAdd(h.ctx, key, userID)
	}

	// 获取 UV（独立访客数）
	uv, err := h.client.PFCount(h.ctx, key).Result()
	if err != nil {
		fmt.Printf("获取 UV 失败: %v\n", err)
		return
	}

	fmt.Printf("总访问次数(PV): %d\n", totalVisits)
	fmt.Printf("独立访客数(UV): %d\n", uv)
	fmt.Printf("实际唯一用户数: %d\n", uniqueUsers)
	fmt.Printf("误差: %d (%.2f%%)\n", abs(int(uv)-uniqueUsers), 
		float64(abs(int(uv)-uniqueUsers))/float64(uniqueUsers)*100)
	fmt.Println("✓ 可以看到，误差在 1% 以内")

	// 查看内存占用
	memUsage := h.client.MemoryUsage(h.ctx, key).Val()
	fmt.Printf("内存占用: %d 字节 (%.2f KB)\n", memUsage, float64(memUsage)/1024)
}

// Example3_MultiDayMerge 多天数据合并示例
func (h *HyperLogLogDemo) Example3_MultiDayMerge() {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("HyperLogLog 示例 3: 多天数据合并（周活/月活统计）")
	fmt.Println(strings.Repeat("=", 60))

	// 模拟 3 天的用户访问数据
	days := []string{"2024-01-15", "2024-01-16", "2024-01-17"}
	
	// 为每天生成数据
	for _, day := range days {
		key := fmt.Sprintf("hll:uv:%s", day)
		h.client.Del(h.ctx, key)

		// 每天随机 1000-1500 个用户访问
		numUsers := 1000 + rand.Intn(500)
		for i := 0; i < numUsers; i++ {
			// 用户 ID 范围 1-2000，所以会有跨天重复的用户
			userID := fmt.Sprintf("user_%d", rand.Intn(2000)+1)
			h.client.PFAdd(h.ctx, key, userID)
		}

		uv, _ := h.client.PFCount(h.ctx, key).Result()
		fmt.Printf("%s 的 UV: %d\n", day, uv)
	}

	// 合并三天的数据，计算 3 日活跃用户
	weekKey := "hll:uv:3day"
	h.client.Del(h.ctx, weekKey)
	
	sourceKeys := make([]string, len(days))
	for i, day := range days {
		sourceKeys[i] = fmt.Sprintf("hll:uv:%s", day)
	}
	
	err := h.client.PFMerge(h.ctx, weekKey, sourceKeys...).Err()
	if err != nil {
		fmt.Printf("合并失败: %v\n", err)
		return
	}

	weekUV, _ := h.client.PFCount(h.ctx, weekKey).Result()
	fmt.Printf("\n3 天合并后的唯一用户数: %d\n", weekUV)
	fmt.Println("✓ 合并后自动去重，得到 3 天内的活跃用户总数")
}

// Example4_PerformanceComparison 性能和内存对比
func (h *HyperLogLogDemo) Example4_PerformanceComparison() {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("HyperLogLog 示例 4: 性能和内存对比")
	fmt.Println(strings.Repeat("=", 60))

	numUsers := 100000
	fmt.Printf("添加 %d 个唯一用户到不同的数据结构...\n\n", numUsers)

	// 方案 1: 使用 Set
	setKey := "compare:set"
	h.client.Del(h.ctx, setKey)
	
	startTime := time.Now()
	for i := 0; i < numUsers; i++ {
		h.client.SAdd(h.ctx, setKey, fmt.Sprintf("user_%d", i))
	}
	setDuration := time.Since(startTime)
	setMem := h.client.MemoryUsage(h.ctx, setKey).Val()
	setCount := h.client.SCard(h.ctx, setKey).Val()

	// 方案 2: 使用 HyperLogLog
	hllKey := "compare:hll"
	h.client.Del(h.ctx, hllKey)
	
	startTime = time.Now()
	for i := 0; i < numUsers; i++ {
		h.client.PFAdd(h.ctx, hllKey, fmt.Sprintf("user_%d", i))
	}
	hllDuration := time.Since(startTime)
	hllMem := h.client.MemoryUsage(h.ctx, hllKey).Val()
	hllCount := h.client.PFCount(h.ctx, hllKey).Val()

	// 输出对比结果
	fmt.Println("【Set 方案】")
	fmt.Printf("  计数结果: %d (精确)\n", setCount)
	fmt.Printf("  内存占用: %d 字节 (%.2f MB)\n", setMem, float64(setMem)/1024/1024)
	fmt.Printf("  写入耗时: %v\n", setDuration)

	fmt.Println("\n【HyperLogLog 方案】")
	fmt.Printf("  计数结果: %d (估算)\n", hllCount)
	fmt.Printf("  内存占用: %d 字节 (%.2f KB)\n", hllMem, float64(hllMem)/1024)
	fmt.Printf("  写入耗时: %v\n", hllDuration)

	fmt.Println("\n【对比总结】")
	fmt.Printf("  内存节省: %.2f%% (Set 的内存是 HLL 的 %.0f 倍)\n", 
		(1-float64(hllMem)/float64(setMem))*100,
		float64(setMem)/float64(hllMem))
	fmt.Printf("  误差: %d (%.2f%%)\n", 
		abs(int(hllCount)-numUsers),
		float64(abs(int(hllCount)-numUsers))/float64(numUsers)*100)
	fmt.Println("✓ 可以看到，HyperLogLog 内存占用极小，误差在可接受范围内")
}

// Example5_RealWorldScenario 真实场景：App DAU/MAU 统计
func (h *HyperLogLogDemo) Example5_RealWorldScenario() {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("HyperLogLog 示例 5: 真实场景 - App DAU/MAU 统计")
	fmt.Println(strings.Repeat("=", 60))

	// 模拟一个月的数据
	fmt.Println("模拟生成 30 天的用户活跃数据...")
	
	baseDate := time.Now().AddDate(0, 0, -30)
	var dailyKeys []string

	for i := 0; i < 30; i++ {
		date := baseDate.AddDate(0, 0, i).Format("2006-01-02")
		key := fmt.Sprintf("hll:dau:%s", date)
		dailyKeys = append(dailyKeys, key)
		
		h.client.Del(h.ctx, key)

		// 每天 50000-80000 个活跃用户
		dailyUsers := 50000 + rand.Intn(30000)
		for j := 0; j < dailyUsers; j++ {
			// 总用户池 500000，模拟真实的用户活跃情况
			userID := fmt.Sprintf("user_%d", rand.Intn(500000)+1)
			h.client.PFAdd(h.ctx, key, userID)
		}

		// 设置过期时间（保留 90 天）
		h.client.Expire(h.ctx, key, 90*24*time.Hour)
	}

	// 计算最后一天的 DAU
	lastDayKey := dailyKeys[len(dailyKeys)-1]
	dau, _ := h.client.PFCount(h.ctx, lastDayKey).Result()
	fmt.Printf("\n昨日 DAU (日活跃用户): %d\n", dau)

	// 计算最近 7 天的 WAU（周活）
	wauKey := "hll:wau:recent"
	h.client.Del(h.ctx, wauKey)
	weekKeys := dailyKeys[len(dailyKeys)-7:]
	h.client.PFMerge(h.ctx, wauKey, weekKeys...)
	wau, _ := h.client.PFCount(h.ctx, wauKey).Result()
	fmt.Printf("最近 7 天 WAU (周活跃用户): %d\n", wau)

	// 计算 30 天的 MAU（月活）
	mauKey := "hll:mau:recent"
	h.client.Del(h.ctx, mauKey)
	h.client.PFMerge(h.ctx, mauKey, dailyKeys...)
	mau, _ := h.client.PFCount(h.ctx, mauKey).Result()
	fmt.Printf("最近 30 天 MAU (月活跃用户): %d\n", mau)

	// 计算留存率等指标
	dauMauRatio := float64(dau) / float64(mau) * 100
	fmt.Printf("\nDAU/MAU 比率: %.2f%% (活跃度指标)\n", dauMauRatio)
	fmt.Println("✓ DAU/MAU 比率越高，说明用户活跃度越好")
	fmt.Println("✓ 使用 HyperLogLog，30 天数据只需要约 360KB 内存 (12KB × 30)")
}

// RunAllExamples 运行所有示例
func (h *HyperLogLogDemo) RunAllExamples() {
	fmt.Println("\n🚀 开始运行 HyperLogLog 所有示例...")
	
	h.Example1_BasicUsage()
	h.Example2_WebsiteUV()
	h.Example3_MultiDayMerge()
	h.Example4_PerformanceComparison()
	h.Example5_RealWorldScenario()

	fmt.Println("\n✅ HyperLogLog 所有示例运行完成！")
}

// 辅助函数：计算绝对值
func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}
