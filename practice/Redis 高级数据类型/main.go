package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/redis/go-redis/v9"
)

func main() {
	// 连接 Redis
	client := redis.NewClient(&redis.Options{
		Addr:     "172.16.27.46:9736", // Redis 地址
		Password: "hasKITs!",          // 密码，如果没有则留空
		DB:       0,                   // 使用默认数据库
	})

	ctx := context.Background()

	// 测试连接
	_, err := client.Ping(ctx).Result()
	if err != nil {
		fmt.Printf("❌ 无法连接到 Redis: %v\n", err)
		fmt.Println("\n请确保:")
		fmt.Println("1. Redis 已安装并运行")
		fmt.Println("2. Redis 监听在 localhost:6379")
		fmt.Println("3. 如果有密码，请在代码中配置")
		fmt.Println("\n启动 Redis 的命令:")
		fmt.Println("  - macOS/Linux: redis-server")
		fmt.Println("  - Docker: docker run -d -p 6379:6379 redis")
		os.Exit(1)
	}

	fmt.Println("✅ 成功连接到 Redis!")
	fmt.Println(strings.Repeat("=", 60))

	// 显示菜单
	showMenu()

	// 读取用户选择
	var choice int
	fmt.Print("\n请输入选项 (1-6): ")
	_, err = fmt.Scanln(&choice)
	if err != nil {
		fmt.Println("❌ 输入错误")
		return
	}

	switch choice {
	case 1:
		// HyperLogLog 示例
		demo := NewHyperLogLogDemo(client)
		demo.RunAllExamples()

	case 2:
		// Geospatial 示例
		demo := NewGeospatialDemo(client)
		demo.RunAllExamples()

	case 3:
		// Bitmap 示例
		demo := NewBitmapDemo(client)
		demo.RunAllExamples()

	case 4:
		// 运行所有示例
		fmt.Println("\n" + strings.Repeat("=", 60))
		fmt.Println("🔥 开始运行所有示例...")
		fmt.Println(strings.Repeat("=", 60) + "\n")

		// 先运行 HyperLogLog
		hllDemo := NewHyperLogLogDemo(client)
		hllDemo.RunAllExamples()

		fmt.Println("\n" + strings.Repeat("=", 60))
		fmt.Println("⏸️  按 Enter 继续运行 Geospatial 示例...")
		fmt.Println(strings.Repeat("=", 60) + "\n")
		fmt.Scanln()

		// 再运行 Geospatial
		geoDemo := NewGeospatialDemo(client)
		geoDemo.RunAllExamples()

		fmt.Println("\n" + strings.Repeat("=", 60))
		fmt.Println("⏸️  按 Enter 继续运行 Bitmap 示例...")
		fmt.Println(strings.Repeat("=", 60) + "\n")
		fmt.Scanln()

		// 最后运行 Bitmap
		bitmapDemo := NewBitmapDemo(client)
		bitmapDemo.RunAllExamples()

	case 5:
		// 快速对比演示
		quickDemo(client)

	case 6:
		// 三种数据类型综合对比
		comprehensiveComparison(client)

	default:
		fmt.Println("❌ 无效的选项")
	}

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("👋 感谢使用！希望这些示例对你有帮助")
	fmt.Println(strings.Repeat("=", 60))
}

func showMenu() {
	fmt.Println("\n📚 Redis 高级数据类型学习系统")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("请选择要运行的示例:")
	fmt.Println()
	fmt.Println("1. HyperLogLog (唯一计数估计)")
	fmt.Println("   - 基础使用")
	fmt.Println("   - 网站 UV 统计")
	fmt.Println("   - 多天数据合并")
	fmt.Println("   - 性能和内存对比")
	fmt.Println("   - App DAU/MAU 统计")
	fmt.Println()
	fmt.Println("2. Geospatial (地理空间索引)")
	fmt.Println("   - 基础使用")
	fmt.Println("   - 查找附近位置")
	fmt.Println("   - 根据成员查找")
	fmt.Println("   - 网约车/外卖配送")
	fmt.Println("   - 动态更新位置")
	fmt.Println("   - 区域查询")
	fmt.Println("   - GeoHash 编码")
	fmt.Println("   - 删除位置")
	fmt.Println()
	fmt.Println("3. Bitmap (位图)")
	fmt.Println("   - 基础使用")
	fmt.Println("   - 每日签到")
	fmt.Println("   - 用户活跃度统计")
	fmt.Println("   - A/B 测试分组")
	fmt.Println("   - 用户权限管理")
	fmt.Println("   - 内存效率对比")
	fmt.Println("   - 在线用户统计")
	fmt.Println()
	fmt.Println("4. 运行所有示例（完整演示）")
	fmt.Println()
	fmt.Println("5. 快速对比演示（5 分钟速览）")
	fmt.Println()
	fmt.Println("6. 三种数据类型综合对比")
	fmt.Println()
	fmt.Println(strings.Repeat("=", 60))
}

// quickDemo 快速演示各数据类型的核心特性
func quickDemo(client *redis.Client) {
	ctx := context.Background()

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("⚡ 快速对比演示")
	fmt.Println(strings.Repeat("=", 60))

	// HyperLogLog 快速演示
	fmt.Println("\n【HyperLogLog - 内存效率对比】")
	fmt.Println(strings.Repeat("-", 60))

	hllKey := "quick:hll"
	setKey := "quick:set"
	client.Del(ctx, hllKey, setKey)

	// 添加 10000 个用户
	fmt.Println("添加 10,000 个唯一用户...")
	for i := 0; i < 10000; i++ {
		userID := fmt.Sprintf("user_%d", i)
		client.PFAdd(ctx, hllKey, userID)
		client.SAdd(ctx, setKey, userID)
	}

	// 对比结果
	hllCount, _ := client.PFCount(ctx, hllKey).Result()
	setCount, _ := client.SCard(ctx, setKey).Result()
	hllMem := client.MemoryUsage(ctx, hllKey).Val()
	setMem := client.MemoryUsage(ctx, setKey).Val()

	fmt.Printf("\nHyperLogLog: 计数 %d, 内存 %.2f KB\n", hllCount, float64(hllMem)/1024)
	fmt.Printf("Set:         计数 %d, 内存 %.2f KB\n", setCount, float64(setMem)/1024)
	fmt.Printf("内存节省:    %.1f%%\n", (1-float64(hllMem)/float64(setMem))*100)

	// Bitmap 快速演示
	fmt.Println("\n【Bitmap - 用户签到演示】")
	fmt.Println(strings.Repeat("-", 60))

	bitmapKey := "quick:bitmap:signin"
	client.Del(ctx, bitmapKey)

	// 模拟 31 天签到情况
	signInDays := []int{1, 2, 3, 5, 7, 10, 15, 20, 25, 28, 30}
	fmt.Println("用户签到日期: ", signInDays)

	for _, day := range signInDays {
		client.SetBit(ctx, bitmapKey, int64(day-1), 1)
	}

	signCount, _ := client.BitCount(ctx, bitmapKey, nil).Result()
	bitmapMem := client.MemoryUsage(ctx, bitmapKey).Val()

	fmt.Printf("签到天数: %d 天\n", signCount)
	fmt.Printf("内存占用: %d 字节\n", bitmapMem)

	// Geospatial 快速演示
	fmt.Println("\n【Geospatial - 附近搜索演示】")
	fmt.Println(strings.Repeat("-", 60))

	geoKey := "quick:geo"
	client.Del(ctx, geoKey)

	// 添加一些位置
	locations := map[string][2]float64{
		"星巴克 (王府井店)": {116.407526, 39.909186},
		"麦当劳 (天安门店)": {116.397428, 39.904989},
		"肯德基 (东单店)":   {116.410982, 39.908154},
		"全家便利店":        {116.395234, 39.915309},
		"711便利店":         {116.402345, 39.912456},
	}

	for name, coords := range locations {
		client.GeoAdd(ctx, geoKey, &redis.GeoLocation{
			Name:      name,
			Longitude: coords[0],
			Latitude:  coords[1],
		})
	}

	// 搜索附近 1 公里的店铺
	userLng, userLat := 116.400000, 39.910000
	fmt.Printf("用户位置: (%.6f, %.6f)\n", userLng, userLat)
	fmt.Println("搜索半径: 1 公里\n")

	results, _ := client.GeoRadius(ctx, geoKey, userLng, userLat, &redis.GeoRadiusQuery{
		Radius:   1,
		Unit:     "km",
		WithDist: true,
		Sort:     "ASC",
	}).Result()

	fmt.Println("附近的店铺:")
	for i, r := range results {
		fmt.Printf("%d. %s - %.2f 公里\n", i+1, r.Name, r.Dist)
	}

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("✅ 快速演示完成!")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("\n💡 提示:")
	fmt.Println("  - HyperLogLog 适合: 大规模去重计数（UV、DAU、MAU）")
	fmt.Println("  - Bitmap 适合: 连续 ID 的布尔值记录（签到、权限）")
	fmt.Println("  - Geospatial 适合: 位置服务（外卖、打车、找店铺）")
}

// comprehensiveComparison 三种数据类型综合对比
func comprehensiveComparison(client *redis.Client) {
	ctx := context.Background()

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("📊 三种数据类型综合对比")
	fmt.Println(strings.Repeat("=", 60))

	numUsers := 100000
	fmt.Printf("\n场景: 存储 %d 个用户的活跃状态\n\n", numUsers)

	// HyperLogLog
	hllKey := "compare:hll"
	client.Del(ctx, hllKey)
	for i := 0; i < numUsers; i++ {
		client.PFAdd(ctx, hllKey, i)
	}
	hllCount := client.PFCount(ctx, hllKey).Val()
	hllMem := client.MemoryUsage(ctx, hllKey).Val()

	// Bitmap（假设用户 ID 是连续的）
	bitmapKey := "compare:bitmap"
	client.Del(ctx, bitmapKey)
	for i := 0; i < numUsers; i++ {
		client.SetBit(ctx, bitmapKey, int64(i), 1)
	}
	bitmapCount := client.BitCount(ctx, bitmapKey, nil).Val()
	bitmapMem := client.MemoryUsage(ctx, bitmapKey).Val()

	// Set
	setKey := "compare:set"
	client.Del(ctx, setKey)
	for i := 0; i < numUsers; i++ {
		client.SAdd(ctx, setKey, i)
	}
	setCount := client.SCard(ctx, setKey).Val()
	setMem := client.MemoryUsage(ctx, setKey).Val()

	// 显示对比表格
	fmt.Println("┌" + strings.Repeat("─", 58) + "┐")
	fmt.Println("│ 数据类型       │ 计数结果  │ 精确度 │ 内存占用          │")
	fmt.Println("├" + strings.Repeat("─", 58) + "┤")
	fmt.Printf("│ Set            │ %7d   │ 100%%   │ %.2f MB         │\n", setCount, float64(setMem)/1024/1024)
	fmt.Printf("│ Bitmap         │ %7d   │ 100%%   │ %.2f KB         │\n", bitmapCount, float64(bitmapMem)/1024)
	fmt.Printf("│ HyperLogLog    │ %7d   │ ~99%%   │ %.2f KB         │\n", hllCount, float64(hllMem)/1024)
	fmt.Println("└" + strings.Repeat("─", 58) + "┘")

	fmt.Println("\n选择建议:")
	fmt.Println(strings.Repeat("-", 60))
	fmt.Println("1. Set (集合)")
	fmt.Println("   ✓ 需要知道具体元素")
	fmt.Println("   ✓ 需要 100% 精确计数")
	fmt.Println("   ✗ 内存消耗大")
	fmt.Println("   适合: 小规模数据（< 10 万）或需要元素列表")
	fmt.Println()
	fmt.Println("2. Bitmap (位图)")
	fmt.Println("   ✓ 需要 100% 精确计数")
	fmt.Println("   ✓ 用户 ID 连续（如自增 ID）")
	fmt.Println("   ✓ 内存效率高")
	fmt.Println("   ✗ ID 不连续时浪费空间")
	fmt.Println("   适合: 连续 ID 的布尔值记录（签到、在线状态、权限）")
	fmt.Println()
	fmt.Println("3. HyperLogLog (基数估计)")
	fmt.Println("   ✓ 内存占用极小（固定 12KB）")
	fmt.Println("   ✓ 适合海量数据")
	fmt.Println("   ✗ 约 1% 误差")
	fmt.Println("   ✗ 无法获取具体元素")
	fmt.Println("   适合: 大规模去重计数（UV、DAU、MAU）")

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("决策树:")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println()
	fmt.Println("需要知道具体元素？")
	fmt.Println("├─ 是 → 用 Set")
	fmt.Println("└─ 否 → 需要精确计数？")
	fmt.Println("    ├─ 是 → 用户 ID 连续？")
	fmt.Println("    │   ├─ 是 → 用 Bitmap")
	fmt.Println("    │   └─ 否 → 用 Set（如果数据量不大）")
	fmt.Println("    └─ 否（可接受 1% 误差）→ 用 HyperLogLog")
}
